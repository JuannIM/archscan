package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// DeadCodeDetector finds functions that are defined but never referenced anywhere else.
type DeadCodeDetector struct{}

func (d *DeadCodeDetector) Detect(root string, files []string, lang string, verbose bool) ([]Violation, error) {
	idFreq := make(map[string]int)

	type funcDef struct {
		name string
		file string
		line int
	}
	var definitions []funcDef

	idReg := regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*`)

	var funcReg *regexp.Regexp
	switch lang {
	case "Python":
		funcReg = regexp.MustCompile(`^\s*def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	case "Go":
		funcReg = regexp.MustCompile(`^\s*func\s+(?:\([^)]+\)\s+)?([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	case "Java":
		funcReg = regexp.MustCompile(`^\s*(?:public\s+|private\s+|protected\s+|static\s+)*[\w<>\[\]]+\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	default:
		funcReg = regexp.MustCompile(`^\s*(?:function|async\s+function)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`)
	}

	for _, f := range files {
		file, err := os.Open(f)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		buf := make([]byte, 1024*1024)
		scanner.Buffer(buf, 1024*1024)

		lineNum := 1
		for scanner.Scan() {
			text := scanner.Text()

			// Find definitions
			matches := funcReg.FindStringSubmatch(text)
			if len(matches) > 1 {
				definitions = append(definitions, funcDef{
					name: matches[1],
					file: f,
					line: lineNum,
				})
			}

			// Count all identifiers
			ids := idReg.FindAllString(text, -1)
			for _, id := range ids {
				idFreq[id]++
			}

			lineNum++
		}
		file.Close()
	}

	var violations []Violation
	count := 0

	for _, def := range definitions {
		if count >= 20 {
			break
		}

		name := def.name
		if name == "main" || name == "init" || name == "String" || name == "Error" {
			continue
		}
		
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, "test") ||
			strings.HasPrefix(name, "Benchmark") ||
			strings.HasPrefix(name, "Example") {
			continue
		}

		// If frequency is exactly 1, it is never called or referenced
		if idFreq[name] == 1 {
			violations = append(violations, Violation{
				Severity: Warning,
				Category: CatDeadCode,
				Title:    fmt.Sprintf("Potentially dead code: %s", name),
				Description: fmt.Sprintf(
					"The function '%s' is defined but its name does not appear anywhere else in the codebase.",
					name,
				),
				Files:      []string{fmt.Sprintf("%s:%d", shortenPath(root, def.file), def.line)},
				Suggestion: "Remove the function if it is no longer used.",
			})
			count++
		}
	}

	return violations, nil
}
