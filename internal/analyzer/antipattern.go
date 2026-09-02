package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"archscan/internal/config"
)

// AntiPatternDetector finds known bad patterns that AI tools commonly introduce.
type AntiPatternDetector struct{}

// antiPattern defines a detectable bad pattern.
type antiPattern struct {
	ID          string
	Title       string
	Description string
	Suggestion  string
	Severity    Severity
	Languages   []string
	Pattern     *regexp.Regexp
}

var knownAntiPatterns = []antiPattern{
	{
		ID:          "silent-error",
		Title:       "Silent error suppression",
		Description: "Error is captured but silently ignored (assigned to `_` or empty catch). AI tools frequently generate this pattern to make code compile without addressing error handling.",
		Suggestion:  "Handle the error explicitly: log it, wrap it, or propagate it to the caller.",
		Severity:    Critical,
		Languages:   []string{"Go"},
		Pattern:     regexp.MustCompile(`\b_\s*=\s*\w+\.\w+\(`),
	},
	{
		ID:          "bare-panic",
		Title:       "Bare panic in non-main code",
		Description: "Use of `panic()` outside of main/init for normal error handling. AI tools use panic as a shortcut instead of returning errors.",
		Suggestion:  "Return an error value instead of panicking in library/service code.",
		Severity:    Warning,
		Languages:   []string{"Go"},
		Pattern:     regexp.MustCompile(`\bpanic\(`),
	},
	{
		ID:          "broad-except",
		Title:       "Broad exception catch",
		Description: "Catching all exceptions with bare `except:` or `except Exception:` masks bugs and makes debugging extremely difficult. Common AI pattern.",
		Suggestion:  "Catch specific exception types relevant to the operation.",
		Severity:    Warning,
		Languages:   []string{"Python"},
		Pattern:     regexp.MustCompile(`except(\s+Exception)?\s*:`),
	},
	{
		ID:          "print-debug",
		Title:       "Debug print statements in production code",
		Description: "Raw print()/fmt.Println() used for debugging instead of a structured logger. AI often leaves debug output.",
		Suggestion:  "Replace with structured logging (log.Info, logger.debug, etc.).",
		Severity:    Info,
		Languages:   []string{"Go", "Python", "Java"},
		Pattern:     regexp.MustCompile(`\b(fmt\.Println|fmt\.Printf|print\(|System\.out\.print)`),
	},
	{
		ID:          "hardcoded-secret",
		Title:       "Potential hardcoded credential or secret",
		Description: "A string literal that looks like a secret, API key, or password. AI tools sometimes hallucinate fake but plausibly structured secrets into code.",
		Suggestion:  "Use environment variables or a secrets manager. Never hardcode credentials.",
		Severity:    Critical,
		Languages:   []string{"Go", "Python", "Java", "TypeScript", "JavaScript"},
		Pattern:     regexp.MustCompile(`(?i)(api_?key|secret|password|token|passwd)\s*[:=]\s*["'][A-Za-z0-9+/=_\-]{10,}["']`),
	},
	{
		ID:          "todo-fixme-critical",
		Title:       "Unresolved TODO/FIXME/HACK comment",
		Description: "Technical debt marker left in code. AI generates placeholder TODOs that never get resolved, accumulating drift.",
		Suggestion:  "Create a tracked issue for this item and remove the comment.",
		Severity:    Info,
		Languages:   []string{"Go", "Python", "Java", "TypeScript", "JavaScript"},
		Pattern:     regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX)\b`),
	},
	{
		ID:          "god-function",
		Title:       "Excessively long function (potential God Function)",
		Description: "A function exceeding 80 lines. AI tends to generate monolithic functions that handle too many responsibilities.",
		Suggestion:  "Break down into smaller, single-responsibility functions.",
		Severity:    Warning,
		Languages:   []string{"Go", "Python", "Java", "TypeScript", "JavaScript"},
		Pattern:     nil, // handled specially
	},
}

func (d *AntiPatternDetector) Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig) ([]Violation, error) {
	var violations []Violation

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		text := string(content)
		short := shortenPath(root, f)

		for _, ap := range knownAntiPatterns {
			// Language filter
			if !containsLang(ap.Languages, lang) {
				continue
			}

			if ap.ID == "god-function" {
				if v := detectGodFunction(f, short, text, lang); v != nil {
					violations = append(violations, *v)
				}
				continue
			}

			matches := ap.Pattern.FindAllStringIndex(text, -1)
			if len(matches) == 0 {
				continue
			}

			// Find line numbers for each match
			var lineRefs []string
			for _, m := range matches {
				line := countLines(text[:m[0]])
				lineRefs = append(lineRefs, fmt.Sprintf("%s:%d", short, line))
				if len(lineRefs) >= 5 {
					lineRefs = append(lineRefs, "...")
					break
				}
			}

			violations = append(violations, Violation{
				Severity:    ap.Severity,
				Category:    CatAntiPattern,
				Title:       ap.Title,
				Description: ap.Description,
				Files:       lineRefs,
				Suggestion:  ap.Suggestion,
			})
		}
	}

	return violations, nil
}

func detectGodFunction(filePath, short, content, lang string) *Violation {
	const maxLines = 80

	scanner := bufio.NewScanner(strings.NewReader(content))
	funcStart := -1
	funcName := ""
	lineNum := 0
	depth := 0

	reFuncGo := regexp.MustCompile(`^func\s+(\w+)`)
	reFuncPy := regexp.MustCompile(`^def\s+(\w+)`)
	reFuncJava := regexp.MustCompile(`\b(public|private|protected|static)\b.*\w+\s+(\w+)\s*\(`)

	var re *regexp.Regexp
	switch lang {
	case "Go":
		re = reFuncGo
	case "Python":
		re = reFuncPy
	case "Java":
		re = reFuncJava
	default:
		return nil
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if funcStart == -1 {
			if m := re.FindStringSubmatch(strings.TrimSpace(line)); len(m) > 0 {
				funcStart = lineNum
				funcName = m[len(m)-1]
				depth = 0
			}
		}

		if funcStart != -1 {
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if lang == "Python" {
				// For Python, track by indentation change
				if lineNum > funcStart+1 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") && strings.TrimSpace(line) != "" {
					length := lineNum - funcStart
					if length > maxLines {
						return &Violation{
							Severity: Warning,
							Category: CatAntiPattern,
							Title:    fmt.Sprintf("God Function: `%s` (%d lines)", funcName, length),
							Description: fmt.Sprintf(
								"Function `%s` in %s spans %d lines (max recommended: %d).\nAI tools often generate monolithic functions handling too many responsibilities.",
								funcName, short, length, maxLines,
							),
							Files:      []string{fmt.Sprintf("%s:%d", short, funcStart)},
							Suggestion: "Break down into smaller, single-responsibility functions.",
						}
					}
					funcStart = -1
				}
			} else {
				if depth <= 0 && lineNum > funcStart {
					length := lineNum - funcStart
					if length > maxLines {
						return &Violation{
							Severity: Warning,
							Category: CatAntiPattern,
							Title:    fmt.Sprintf("God Function: `%s` (%d lines)", funcName, length),
							Description: fmt.Sprintf(
								"Function `%s` in %s spans %d lines (max recommended: %d).\nAI tools often generate monolithic functions handling too many responsibilities.",
								funcName, short, length, maxLines,
							),
							Files:      []string{fmt.Sprintf("%s:%d", short, funcStart)},
							Suggestion: "Break down into smaller, single-responsibility functions.",
						}
					}
					funcStart = -1
				}
			}
		}
	}
	return nil
}

func containsLang(langs []string, lang string) bool {
	for _, l := range langs {
		if l == lang {
			return true
		}
	}
	return false
}

func countLines(s string) int {
	return strings.Count(s, "\n") + 1
}
