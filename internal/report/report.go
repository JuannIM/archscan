// Package report renders analysis results in multiple formats.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"archscan/internal/analyzer"

	"github.com/fatih/color"
)

// Report renders an analysis result.
type Report struct {
	result *analyzer.Result
	format string
}

// New creates a new report renderer.
func New(result *analyzer.Result, format string) *Report {
	return &Report{result: result, format: format}
}

// Render writes the formatted report to w.
func (r *Report) Render(w io.Writer) error {
	switch r.format {
	case "json":
		return r.renderJSON(w)
	case "markdown", "md":
		return r.renderMarkdown(w)
	default:
		return r.renderText(w)
	}
}

func (r *Report) renderText(w io.Writer) error {
	result := r.result

	bold := color.New(color.Bold)
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow, color.Bold)
	blue := color.New(color.FgBlue)
	green := color.New(color.FgGreen, color.Bold)
	dim := color.New(color.Faint)

	// Summary header
	fmt.Fprintf(w, "┌─────────────────────────────────────────────────────┐\n")
	fmt.Fprintf(w, "│  Scan Results — %s %-34s│\n", result.Language, "")
	fmt.Fprintf(w, "│  Files scanned: %-3d  Violations: %-19d│\n",
		result.TotalFiles, len(result.Violations))
	fmt.Fprintf(w, "└─────────────────────────────────────────────────────┘\n\n")

	if len(result.Violations) == 0 {
		green.Fprintln(w, "  ✓ No architectural drift detected. Clean codebase!")
		return nil
	}

	// Group by severity
	critical := filterBySeverity(result.Violations, analyzer.Critical)
	warnings := filterBySeverity(result.Violations, analyzer.Warning)
	infos := filterBySeverity(result.Violations, analyzer.Info)

	if len(critical) > 0 {
		red.Fprintf(w, "\n🔴 CRITICAL (%d)\n", len(critical))
		fmt.Fprintln(w, strings.Repeat("─", 54))
		for i, v := range critical {
			renderViolationText(w, i+1, v, bold, red, dim)
		}
	}

	if len(warnings) > 0 {
		yellow.Fprintf(w, "\n🟡 WARNINGS (%d)\n", len(warnings))
		fmt.Fprintln(w, strings.Repeat("─", 54))
		for i, v := range warnings {
			renderViolationText(w, i+1, v, bold, yellow, dim)
		}
	}

	if len(infos) > 0 {
		blue.Fprintf(w, "\n🔵 INFO (%d)\n", len(infos))
		fmt.Fprintln(w, strings.Repeat("─", 54))
		for i, v := range infos {
			renderViolationText(w, i+1, v, bold, blue, dim)
		}
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		fmt.Fprintf(w, "\n")
		bold.Fprintln(w, "💡 Recommendations")
		fmt.Fprintln(w, strings.Repeat("─", 54))
		for _, s := range result.Suggestions {
			fmt.Fprintf(w, "  • %s\n", s)
		}
	}

	// Score
	score := calculateScore(result)
	fmt.Fprintf(w, "\n")
	bold.Fprintf(w, "Architecture Health Score: ")
	scoreColor := green
	scoreLabel := "Excellent"
	if score < 40 {
		scoreColor = red
		scoreLabel = "Critical Drift"
	} else if score < 70 {
		scoreColor = yellow
		scoreLabel = "Degraded"
	} else if score < 85 {
		scoreColor = blue
		scoreLabel = "Fair"
	}
	scoreColor.Fprintf(w, "%d/100 (%s)\n\n", score, scoreLabel)

	return nil
}

func renderViolationText(w io.Writer, n int, v analyzer.Violation, bold, accent, dim *color.Color) {
	accent.Fprintf(w, "\n  %d. %s\n", n, v.Title)
	fmt.Fprintf(w, "     Category: %s\n", v.Category)

	// Wrap description
	for _, line := range strings.Split(v.Description, "\n") {
		if line != "" {
			dim.Fprintf(w, "     %s\n", line)
		}
	}

	if len(v.Files) > 0 {
		fmt.Fprintf(w, "     Files:\n")
		for _, f := range v.Files {
			fmt.Fprintf(w, "       → %s\n", f)
		}
	}

	bold.Fprintf(w, "     Fix: ")
	fmt.Fprintf(w, "%s\n", v.Suggestion)
}

func (r *Report) renderJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r.result)
}

func (r *Report) renderMarkdown(w io.Writer) error {
	result := r.result
	score := calculateScore(result)

	fmt.Fprintf(w, "# archscan Report\n\n")
	fmt.Fprintf(w, "**Language:** %s | **Files:** %d | **Violations:** %d | **Score:** %d/100\n\n",
		result.Language, result.TotalFiles, len(result.Violations), score)

	if len(result.Violations) == 0 {
		fmt.Fprintln(w, "> ✅ No architectural drift detected.")
		return nil
	}

	severities := []analyzer.Severity{analyzer.Critical, analyzer.Warning, analyzer.Info}
	icons := map[analyzer.Severity]string{
		analyzer.Critical: "🔴",
		analyzer.Warning:  "🟡",
		analyzer.Info:     "🔵",
	}

	for _, sev := range severities {
		violations := filterBySeverity(result.Violations, sev)
		if len(violations) == 0 {
			continue
		}
		fmt.Fprintf(w, "## %s %s (%d)\n\n", icons[sev], sev, len(violations))
		for _, v := range violations {
			fmt.Fprintf(w, "### %s\n", v.Title)
			fmt.Fprintf(w, "**Category:** %s\n\n", v.Category)
			fmt.Fprintf(w, "%s\n\n", v.Description)
			if len(v.Files) > 0 {
				fmt.Fprintln(w, "**Files:**")
				for _, f := range v.Files {
					fmt.Fprintf(w, "- `%s`\n", f)
				}
				fmt.Fprintln(w)
			}
			fmt.Fprintf(w, "> **Fix:** %s\n\n", v.Suggestion)
			fmt.Fprintln(w, "---")
		}
	}

	if len(result.Suggestions) > 0 {
		fmt.Fprintln(w, "## 💡 Recommendations")
		for _, s := range result.Suggestions {
			fmt.Fprintf(w, "- %s\n", s)
		}
	}

	return nil
}

func filterBySeverity(violations []analyzer.Violation, sev analyzer.Severity) []analyzer.Violation {
	var out []analyzer.Violation
	for _, v := range violations {
		if v.Severity == sev {
			out = append(out, v)
		}
	}
	return out
}

func calculateScore(r *analyzer.Result) int {
	if r.TotalFiles == 0 {
		return 100
	}
	score := 100
	for _, v := range r.Violations {
		switch v.Severity {
		case analyzer.Critical:
			score -= 15
		case analyzer.Warning:
			score -= 5
		case analyzer.Info:
			score -= 1
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}
