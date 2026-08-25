package analyzer

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"strings"
)

// DuplicationDetector finds duplicate function-level blocks across files.
// Uses function-level extraction (not sliding window) for O(n) performance.
type DuplicationDetector struct{}

func (d *DuplicationDetector) Detect(root string, files []string, lang string, verbose bool) ([]Violation, error) {
	// Map: normalized-function-hash -> list of (file:line, preview)
	type occurrence struct {
		location string
		preview  string
	}
	seen := map[string][]occurrence{}

	for _, f := range files {
		funcs, err := extractFunctions(f, lang)
		if err != nil {
			continue
		}
		for _, fn := range funcs {
			if fn.lineCount < 8 {
				continue // Too short to be meaningful
			}
			h := hashContent(fn.normalized)
			loc := fmt.Sprintf("%s:%d", shortenPath(root, f), fn.startLine)
			seen[h] = append(seen[h], occurrence{location: loc, preview: fn.preview})
		}
	}

	var violations []Violation
	reported := map[string]bool{}

	for hash, occs := range seen {
		if len(occs) < 2 || reported[hash] {
			continue
		}
		reported[hash] = true

		fileList := make([]string, 0, len(occs))
		for _, o := range occs {
			fileList = append(fileList, o.location)
		}

		severity := Warning
		if len(occs) >= 3 {
			severity = Critical
		}

		violations = append(violations, Violation{
			Severity: severity,
			Category: CatDuplication,
			Title:    fmt.Sprintf("Duplicate function body found in %d locations", len(occs)),
			Description: fmt.Sprintf(
				"An identical function (normalized) appears in multiple files — classic AI drift pattern.\n\nPreview:\n%s",
				occs[0].preview,
			),
			Files:      fileList,
			Suggestion: "Extract into a shared utility/helper. Search before generating.",
		})
	}

	return violations, nil
}

type funcBlock struct {
	normalized string
	preview    string
	startLine  int
	lineCount  int
}

// extractFunctions extracts function bodies from a file using brace-counting (Go/Java/TS)
// or indentation (Python).
func extractFunctions(filePath, lang string) ([]funcBlock, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<20) // 1MB buffer
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	switch lang {
	case "Go", "Java", "TypeScript", "JavaScript":
		return extractBraceFunctions(lines)
	case "Python":
		return extractPythonFunctions(lines)
	default:
		return nil, nil
	}
}

func extractBraceFunctions(lines []string) ([]funcBlock, error) {
	var blocks []funcBlock
	inFunc := false
	depth := 0
	start := 0
	var body []string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inFunc {
			// Detect function start (has opening brace, looks like a function)
			if isFuncStart(trimmed) && strings.Contains(line, "{") {
				inFunc = true
				start = i + 1
				depth = strings.Count(line, "{") - strings.Count(line, "}")
				body = []string{normalizeLine(trimmed)}
				continue
			}
		}

		if inFunc {
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			body = append(body, normalizeLine(trimmed))

			if depth <= 0 {
				// Function ended
				if len(body) >= 8 {
					normalized := strings.Join(body, "\n")
					blocks = append(blocks, funcBlock{
						normalized: normalized,
						preview:    buildPreview(body, 4),
						startLine:  start,
						lineCount:  len(body),
					})
				}
				inFunc = false
				body = nil
			}
		}
	}
	return blocks, nil
}

func extractPythonFunctions(lines []string) ([]funcBlock, error) {
	var blocks []funcBlock
	inFunc := false
	start := 0
	var body []string
	var defIndent string

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if inFunc {
				body = append(body, "")
			}
			continue
		}

		indent := leadingWhitespace(line)

		// Check if the current function should end
		if inFunc {
			// A function ends if we encounter a non-comment line with <= indentation than the 'def' line.
			// We ignore closing brackets/parens which sometimes align with the 'def' line.
			if len(indent) <= len(defIndent) && !strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, ")") && !strings.HasPrefix(trimmed, "]") && !strings.HasPrefix(trimmed, "}") && !strings.HasPrefix(trimmed, "@") {
				if len(body) >= 8 {
					blocks = append(blocks, funcBlock{
						normalized: strings.Join(body, "\n"),
						preview:    buildPreview(body, 4),
						startLine:  start,
						lineCount:  len(body),
					})
				}
				inFunc = false
				body = nil
			}
		}

		// Detect a new function
		if !inFunc && strings.HasPrefix(trimmed, "def ") {
			inFunc = true
			start = i + 1
			body = []string{normalizeLine(trimmed)}
			defIndent = indent
			continue
		}

		if inFunc {
			body = append(body, normalizeLine(trimmed))
		}
	}

	// Catch the last function if the file ends
	if inFunc && len(body) >= 8 {
		blocks = append(blocks, funcBlock{
			normalized: strings.Join(body, "\n"),
			preview:    buildPreview(body, 4),
			startLine:  start,
			lineCount:  len(body),
		})
	}

	return blocks, nil
}

func isFuncStart(line string) bool {
	prefixes := []string{"func ", "def ", "function ", "public ", "private ", "protected ", "static ", "async "}
	for _, p := range prefixes {
		if strings.HasPrefix(line, p) {
			return true
		}
	}
	return false
}

func normalizeLine(line string) string {
	// Strip single-line comments
	for _, commentPrefix := range []string{"//", "#"} {
		if idx := strings.Index(line, commentPrefix); idx > 0 {
			line = line[:idx]
		}
	}
	line = strings.TrimSpace(line)
	// Collapse whitespace
	return strings.Join(strings.Fields(line), " ")
}

func buildPreview(lines []string, max int) string {
	var nonEmpty []string
	for _, l := range lines {
		if l != "" {
			nonEmpty = append(nonEmpty, l)
		}
		if len(nonEmpty) >= max {
			break
		}
	}
	return "    " + strings.Join(nonEmpty, "\n    ") + "\n    ..."
}

func leadingWhitespace(s string) string {
	return s[:len(s)-len(strings.TrimLeft(s, " \t"))]
}

func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:12])
}
