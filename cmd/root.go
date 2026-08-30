package cmd

import (
	"fmt"
	"os"

	"archscan/internal/analyzer"
	"archscan/internal/report"
	"archscan/internal/rules"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	outputFormat  string
	generateRules bool
	verbose       bool
)

var rootCmd = &cobra.Command{
	Use:   "archscan [path]",
	Short: "Architectural drift detector for AI-assisted repositories",
	Long: `archscan analyzes your codebase for architectural drift caused by
AI coding tools (Cursor, Copilot, Claude Code, etc.) and generates
context rules to prevent future violations.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", repoPath)
		}

		cyan := color.New(color.FgCyan, color.Bold)
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen, color.Bold)

		cyan.Println("\n⚡ archscan — Architectural Drift Detector")
		fmt.Printf("   Scanning: %s\n\n", repoPath)

		// Run analysis (no limits)
		result, err := analyzer.Analyze(repoPath, verbose)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		// Render report
		r := report.New(result, outputFormat)
		if err := r.Render(os.Stdout); err != nil {
			return fmt.Errorf("report rendering failed: %w", err)
		}

		// Generate AI rule files
		if generateRules {
			bold.Println("\n📝 Generating AI context rules...")
			gen := rules.NewGenerator(result)
			files, err := gen.Generate(repoPath)
			if err != nil {
				return fmt.Errorf("rules generation failed: %w", err)
			}
			for _, f := range files {
				green.Printf("   ✓ Generated: %s\n", f)
			}
		}

		// Non-zero exit on critical violations (useful for CI gates)
		if result.HasCritical() {
			os.Exit(1)
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "text",
		"Output format: text | json | markdown")
	rootCmd.Flags().BoolVarP(&generateRules, "rules", "r", false,
		"Generate .cursorrules and CLAUDE.md for AI tools")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose output")
}
