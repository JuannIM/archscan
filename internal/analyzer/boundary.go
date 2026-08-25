package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// BoundaryDetector finds violations of architectural layer boundaries.
// Example: UI/presentation layer directly importing from data/database layer.
type BoundaryDetector struct{}

// layerPattern maps directory/package name patterns to architectural layers.
var layerPatterns = []struct {
	Layer   string
	Pattern *regexp.Regexp
}{
	{"presentation", regexp.MustCompile(`(?i)(handler|controller|view|ui|api|route|endpoint|http|rest|graphql|grpc)`)},
	{"application",  regexp.MustCompile(`(?i)(service|usecase|use_case|application|app|command|query)`)},
	{"domain",       regexp.MustCompile(`(?i)(domain|model|entity|aggregate|repository|repo)`)},
	{"infrastructure", regexp.MustCompile(`(?i)(infra|infrastructure|db|database|storage|cache|queue|email|notification|persistence|adapter)`)},
}

// Forbidden cross-layer imports: layer X should not import from layer Y directly.
// Enforces: presentation -> application -> domain <- infrastructure
var forbiddenCross = []struct {
	From string
	To   string
}{
	{"presentation", "infrastructure"},
	{"presentation", "domain"},
	{"application",  "infrastructure"},
}

func (d *BoundaryDetector) Detect(root string, files []string, lang string, verbose bool) ([]Violation, error) {
	var violations []Violation

	for _, f := range files {
		fromLayer := classifyFile(f)
		if fromLayer == "" {
			continue
		}

		imports, err := extractImports(f, lang)
		if err != nil {
			continue
		}

		for _, imp := range imports {
			toLayer := classifyImport(imp)
			if toLayer == "" {
				continue
			}

			if isForbidden(fromLayer, toLayer) {
				short := shortenPath(root, f)
				violations = append(violations, Violation{
					Severity: Critical,
					Category: CatBoundaryViolation,
					Title:    fmt.Sprintf("Layer violation: %s → %s", fromLayer, toLayer),
					Description: fmt.Sprintf(
						"File `%s` (layer: %s) directly imports from the `%s` layer.\n"+
							"This breaks Clean Architecture / Hexagonal Architecture boundaries.\n"+
							"AI tools frequently introduce this pattern because they optimize for local context.",
						short, fromLayer, toLayer,
					),
					Files:      []string{short},
					Suggestion: fmt.Sprintf("Introduce an interface/port in the %s layer and inject it via dependency injection.", fromLayer),
				})
			}
		}
	}

	return violations, nil
}

func classifyFile(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	for _, part := range parts {
		for _, lp := range layerPatterns {
			if lp.Pattern.MatchString(part) {
				return lp.Layer
			}
		}
	}
	return ""
}

func classifyImport(imp string) string {
	for _, lp := range layerPatterns {
		if lp.Pattern.MatchString(imp) {
			return lp.Layer
		}
	}
	return ""
}

func isForbidden(from, to string) bool {
	for _, fc := range forbiddenCross {
		if fc.From == from && fc.To == to {
			return true
		}
	}
	return false
}

// extractImports returns import paths from a source file (language-aware).
func extractImports(filePath, lang string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var imports []string
	scanner := bufio.NewScanner(f)

	switch lang {
	case "Go":
		inImportBlock := false
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Detect import block open: "import (" (possibly with trailing whitespace)
			if line == "import (" {
				inImportBlock = true
				continue
			}
			if inImportBlock && line == ")" {
				inImportBlock = false
				continue
			}
			if inImportBlock {
				// Skip blank lines and comments inside import block
				if line == "" || strings.HasPrefix(line, "//") {
					continue
				}
				// Handle aliased imports: `alias "pkg/path"` → take last field, strip quotes
				fields := strings.Fields(line)
				if len(fields) == 0 {
					continue
				}
				imp := strings.Trim(fields[len(fields)-1], `"`)
				if imp != "" {
					imports = append(imports, imp)
				}
			} else if strings.HasPrefix(line, `import "`) {
				imp := strings.TrimPrefix(line, `import "`)
				imp = strings.TrimSuffix(imp, `"`)
				imports = append(imports, imp)
			}
		}

	case "Python":
		reFrom := regexp.MustCompile(`^from\s+([\w.]+)\s+import`)
		reImport := regexp.MustCompile(`^import\s+([\w.,\s]+)`)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if m := reFrom.FindStringSubmatch(line); len(m) > 1 {
				imports = append(imports, m[1])
			} else if m := reImport.FindStringSubmatch(line); len(m) > 1 {
				for _, pkg := range strings.Split(m[1], ",") {
					imports = append(imports, strings.TrimSpace(pkg))
				}
			}
		}

	case "Java":
		reImport := regexp.MustCompile(`^import\s+([\w.]+);`)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if m := reImport.FindStringSubmatch(line); len(m) > 1 {
				imports = append(imports, m[1])
			}
		}

	case "TypeScript", "JavaScript":
		reFrom := regexp.MustCompile(`from\s+['"]([^'"]+)['"]`)
		reRequire := regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
		for scanner.Scan() {
			line := scanner.Text()
			if m := reFrom.FindStringSubmatch(line); len(m) > 1 {
				imports = append(imports, m[1])
			}
			if m := reRequire.FindStringSubmatch(line); len(m) > 1 {
				imports = append(imports, m[1])
			}
		}
	}

	return imports, nil
}
