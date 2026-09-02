package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"archscan/internal/config"
)

type Result struct {
	RepoPath    string
	Language    string
	TotalFiles  int
	Violations  []Violation
	Suggestions []string
	Modules     []Module
	Stats       Stats
}
type Stats struct {
	FilesScanned     int
	DuplicateSets    int
	BrokenBoundaries int
	AntiPatterns     int
}
type Module struct {
	Name         string
	Path         string
	Imports      []string
	ExportedSyms []string
}
type Violation struct {
	Severity    Severity
	Category    Category
	Title       string
	Description string
	Files       []string
	Suggestion  string
}
type Severity string
const (
	Critical Severity = "CRITICAL"
	Warning  Severity = "WARNING"
	Info     Severity = "INFO"
)
type Category string
const (
	CatDuplication         Category = "Duplication"
	CatBoundaryViolation   Category = "BoundaryViolation"
	CatAntiPattern         Category = "AntiPattern"
	CatNamingInconsistency Category = "NamingInconsistency"
	CatDeadCode            Category = "DeadCode"
)
func (r *Result) HasCritical() bool {
	for _, v := range r.Violations {
		if v.Severity == Critical {
			return true
		}
	}
	return false
}
func Analyze(repoPath string, verbose bool, cfg *config.ArchscanConfig) (*Result, error) {
	result := &Result{
		RepoPath: repoPath,
	}
	lang, err := detectLanguage(repoPath)
	if err != nil {
		return nil, fmt.Errorf("language detection: %w", err)
	}
	result.Language = lang
	if verbose {
		fmt.Printf("   Detected language: %s\n", lang)
	}
	files, err := collectFiles(repoPath, lang, cfg.Exclude)
	if err != nil {
		return nil, fmt.Errorf("file collection: %w", err)
	}
	result.TotalFiles = len(files)
	if verbose {
		fmt.Printf("   Source files found: %d\n", len(files))
	}
	detectors := []Detector{
		&DuplicationDetector{},
		&AntiPatternDetector{},
		&NamingDetector{},
		&BoundaryDetector{},
	}
	for _, d := range detectors {
		violations, err := d.Detect(repoPath, files, lang, verbose, cfg)
		if err != nil {
			if verbose {
				fmt.Printf("   [warn] detector %T failed: %v\n", d, err)
			}
			continue
		}
		result.Violations = append(result.Violations, violations...)
	}
	result.Stats = buildStats(result.Violations)
	result.Suggestions = buildSuggestions(result)
	return result, nil
}
func detectLanguage(root string) (string, error) {
	counts := map[string]int{}
	extensions := map[string]string{
		".go":   "Go",
		".py":   "Python",
		".java": "Java",
		".ts":   "TypeScript",
		".js":   "JavaScript",
		".rs":   "Rust",
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && shouldSkipDir(info.Name()) {
			return filepath.SkipDir
		}
		ext := strings.ToLower(filepath.Ext(path))
		if lang, ok := extensions[ext]; ok {
			counts[lang]++
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if len(counts) == 0 {
		return "Unknown", nil
	}
	best, bestCount := "", 0
	for lang, count := range counts {
		if count > bestCount {
			best, bestCount = lang, count
		}
	}
	return best, nil
}
func collectFiles(root, lang string, exclude []string) ([]string, error) {
	extMap := map[string][]string{
		"Go":         {".go"},
		"Python":     {".py"},
		"Java":       {".java"},
		"TypeScript": {".ts", ".tsx"},
		"JavaScript": {".js", ".jsx"},
		"Rust":       {".rs"},
	}
	exts, ok := extMap[lang]
	if !ok {
		exts = []string{}
	}
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		if info.IsDir() {
			if shouldSkipDir(info.Name()) || isExcluded(rel, exclude) {
				return filepath.SkipDir
			}
			return nil
		}
		if isExcluded(rel, exclude) {
			return nil
		}
		fileExt := strings.ToLower(filepath.Ext(path))
		for _, ext := range exts {
			if fileExt == ext {
				files = append(files, path)
				break
			}
		}
		return nil
	})
	return files, err
}
func shouldSkipDir(name string) bool {
	skip := map[string]bool{
		".git": true, "node_modules": true, "vendor": true,
		".venv": true, "venv": true, "__pycache__": true,
		"dist": true, "build": true, "out": true, ".idea": true,
		".vscode": true, "target": true, ".cache": true,
		".gradle": true, ".mvn": true, "bin": true,
	}
	return skip[name]
}
func buildStats(violations []Violation) Stats {
	s := Stats{}
	for _, v := range violations {
		switch v.Category {
		case CatDuplication:
			s.DuplicateSets++
		case CatBoundaryViolation:
			s.BrokenBoundaries++
		case CatAntiPattern:
			s.AntiPatterns++
		}
		s.FilesScanned++
	}
	return s
}
func buildSuggestions(r *Result) []string {
	var s []string
	if r.Stats.DuplicateSets > 0 {
		s = append(s, fmt.Sprintf("Consolidate %d duplicate code sets into shared utilities", r.Stats.DuplicateSets))
	}
	if r.Stats.BrokenBoundaries > 0 {
		s = append(s, "Add architectural linting rules to CI to prevent layer violations")
	}
	if r.Stats.AntiPatterns > 0 {
		s = append(s, "Run `archscan --rules` to generate AI context files and prevent future pattern drift")
	}
	return s
}

func isExcluded(relPath string, exclude []string) bool {
	slashPath := filepath.ToSlash(relPath)
	for _, ex := range exclude {
		if strings.HasPrefix(ex, "**/") && strings.HasSuffix(ex, "/**") {
			match := strings.TrimSuffix(strings.TrimPrefix(ex, "**/"), "/**")
			if strings.Contains(slashPath, "/"+match+"/") || strings.HasPrefix(slashPath, match+"/") {
				return true
			}
		} else if strings.HasPrefix(ex, "**/") {
			match := strings.TrimPrefix(ex, "**/")
			matched, _ := filepath.Match(match, filepath.Base(slashPath))
			if matched {
				return true
			}
		} else if strings.HasSuffix(ex, "/") {
			match := strings.TrimSuffix(ex, "/")
			if strings.Contains(slashPath, "/"+match+"/") || strings.HasPrefix(slashPath, match+"/") || slashPath == match {
				return true
			}
		} else {
			matched, _ := filepath.Match(ex, filepath.Base(slashPath))
			if matched {
				return true
			}
		}
	}
	return false
}
