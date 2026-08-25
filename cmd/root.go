package cmd

import (
	"fmt"
	"os"
	"strings"

	"archscan/internal/analyzer"
	"archscan/internal/license"
	"archscan/internal/report"
	"archscan/internal/rules"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const freeTierFileLimit = 200

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
context rules to prevent future violations.

  Free tier:  basic detection, up to 200 files
  Pro  tier:  unlimited files, boundary detection, --rules, --format json/markdown

  Get Pro: https://polar.sh/archscan`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		if _, err := os.Stat(repoPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", repoPath)
		}

		// Load license (also checks ARCHSCAN_LICENSE_KEY env var)
		lic := loadLicense()

		cyan := color.New(color.FgCyan, color.Bold)
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen, color.Bold)
		yellow := color.New(color.FgYellow)
		dim := color.New(color.Faint)

		cyan.Println("\n⚡ archscan — Architectural Drift Detector")
		fmt.Printf("   Scanning: %s\n", repoPath)

		if lic.IsPro() {
			dim.Printf("   License:  Pro (%s, %d days left)\n\n", lic.Email, lic.DaysRemaining())
		} else {
			yellow.Printf("   License:  Free (max %d files) · polar.sh/archscan for Pro\n\n", freeTierFileLimit)
		}

		// --- Pro feature gates ---

		// --format json/markdown requires Pro
		if outputFormat != "text" && !lic.IsPro() {
			return proRequired("--format " + outputFormat)
		}

		// --rules requires Pro
		if generateRules && !lic.IsPro() {
			return proRequired("--rules")
		}

		// Run analysis (pass license so analyzer can gate boundary detector)
		result, err := analyzer.Analyze(repoPath, verbose, lic.IsPro(), freeTierFileLimit)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		// Render report
		r := report.New(result, outputFormat)
		if err := r.Render(os.Stdout); err != nil {
			return fmt.Errorf("report rendering failed: %w", err)
		}

		// Generate AI rule files (Pro)
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

// loadLicense checks env var first, then disk.
func loadLicense() *license.License {
	if key := os.Getenv("ARCHSCAN_LICENSE_KEY"); key != "" {
		// Key format stored in env: "email:ARCHSCAN-XXXX..."
		parts := strings.SplitN(key, ":", 2)
		if len(parts) == 2 {
			lic, err := license.ValidateKey(parts[1], parts[0], license.PlanPro)
			if err == nil {
				return lic
			}
		}
	}
	return license.Load()
}

func proRequired(feature string) error {
	yellow := color.New(color.FgYellow, color.Bold)
	fmt.Println()
	yellow.Printf("  🔒 %s is a Pro feature.\n\n", feature)
	fmt.Println("  Upgrade at: https://polar.sh/archscan  ($9/mo · $79/yr)")
	fmt.Println("  Then activate with:")
	fmt.Println("    archscan activate --email you@example.com --key ARCHSCAN-...")
	fmt.Println()
	os.Exit(2)
	return nil
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&outputFormat, "format", "f", "text",
		"Output format: text | json | markdown  [Pro]")
	rootCmd.Flags().BoolVarP(&generateRules, "rules", "r", false,
		"Generate .cursorrules and CLAUDE.md for AI tools  [Pro]")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false,
		"Verbose output")
}
