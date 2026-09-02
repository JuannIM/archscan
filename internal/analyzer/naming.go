package analyzer

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"archscan/internal/config"
)

// NamingDetector finds inconsistent naming conventions across a codebase.
// AI tools often mix conventions (camelCase vs snake_case, ALL_CAPS vs CamelCase, etc.)
type NamingDetector struct{}

func (d *NamingDetector) Detect(root string, files []string, lang string, verbose bool, cfg *config.ArchscanConfig) ([]Violation, error) {
	if len(files) < 5 {
		return nil, nil
	}
	switch lang {
	case "Go":
		return detectGoNaming(root, files)
	case "Python":
		return detectPythonNaming(root, files)
	default:
		return nil, nil
	}
}

func detectGoNaming(root string, files []string) ([]Violation, error) {
	allCapsCount := 0
	camelCount := 0

	allCapsRe := regexp.MustCompile(`const\s+[A-Z_]{3,}\s*=`)
	camelConstRe := regexp.MustCompile(`const\s+[A-Z][a-z]+[A-Z]`)

	for _, f := range files {
		content, err := safeReadFile(f, 4096)
		if err != nil {
			continue
		}
		if allCapsRe.MatchString(content) {
			allCapsCount++
		}
		if camelConstRe.MatchString(content) {
			camelCount++
		}
	}

	var violations []Violation
	if allCapsCount > 0 && camelCount > 0 {
		violations = append(violations, Violation{
			Severity: Warning,
			Category: CatNamingInconsistency,
			Title:    "Mixed constant naming conventions (ALL_CAPS vs CamelCase)",
			Description: fmt.Sprintf(
				"Found %d files using ALL_CAPS constants and %d using CamelCase.\n"+
					"Go convention is CamelCase/camelCase. ALL_CAPS is a C/Python pattern.\n"+
					"AI tools frequently mix these styles when context is ambiguous.",
				allCapsCount, camelCount,
			),
			Files:      []string{"multiple files"},
			Suggestion: "Standardize on Go convention: ExportedConst and unexportedConst.",
		})
	}
	return violations, nil
}

func detectPythonNaming(root string, files []string) ([]Violation, error) {
	camelFuncRe := regexp.MustCompile(`(?m)^def\s+([a-z][a-zA-Z0-9]*[A-Z][a-zA-Z0-9]*)\s*\(`)
	snakeFuncRe := regexp.MustCompile(`(?m)^def\s+([a-z][a-z0-9_]+)\s*\(`)

	camelCount, snakeCount := 0, 0
	var camelExamples, snakeExamples []string

	for _, f := range files {
		content, err := safeReadFile(f, 8192)
		if err != nil {
			continue
		}
		if mm := camelFuncRe.FindAllStringSubmatch(content, -1); mm != nil {
			camelCount += len(mm)
			for _, m := range mm {
				if len(camelExamples) < 3 {
					camelExamples = append(camelExamples, m[1])
				}
			}
		}
		if mm := snakeFuncRe.FindAllStringSubmatch(content, -1); mm != nil {
			snakeCount += len(mm)
			for _, m := range mm {
				if len(snakeExamples) < 3 {
					snakeExamples = append(snakeExamples, m[1])
				}
			}
		}
	}

	var violations []Violation
	if camelCount > 0 && snakeCount > 2 {
		violations = append(violations, Violation{
			Severity: Warning,
			Category: CatNamingInconsistency,
			Title:    fmt.Sprintf("Mixed function naming: camelCase (%d) vs snake_case (%d)", camelCount, snakeCount),
			Description: fmt.Sprintf(
				"Python convention (PEP 8) requires snake_case for functions.\n"+
					"camelCase found: %s\nsnake_case found: %s\n"+
					"AI tools mix these conventions when trained on multi-language corpora.",
				strings.Join(camelExamples, ", "),
				strings.Join(snakeExamples, ", "),
			),
			Files:      []string{"multiple files"},
			Suggestion: "Rename all function definitions to snake_case per PEP 8.",
		})
	}
	return violations, nil
}

// safeReadFile reads up to maxBytes of a file as a string.
func safeReadFile(path string, maxBytes int) (string, error) {
	_ = filepath.Base(path) // ensure import used
	data, err := readFileSafe(path, maxBytes)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
