package cmd

import (
	"fmt"
	"strings"

	"archscan/internal/license"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var activateCmd = &cobra.Command{
	Use:   "activate",
	Short: "Activate archscan Pro with your license key",
	Long: `Activate archscan Pro by providing your email and license key.

Your license key is emailed to you after purchase at:
  https://polar.sh/archscan

Example:
  archscan activate --email you@example.com --key ARCHSCAN-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX`,
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		key, _ := cmd.Flags().GetString("key")

		if email == "" || key == "" {
			return fmt.Errorf("both --email and --key are required")
		}

		bold := color.New(color.Bold)
		green := color.New(color.FgGreen, color.Bold)
		red := color.New(color.FgRed, color.Bold)

		fmt.Println()
		bold.Println("  Validating license key...")

		lic, err := license.ValidateKey(key, email, license.PlanPro)
		if err != nil {
			fmt.Println()
			red.Println("  ✗ Invalid license key.")
			fmt.Println("  Check that your email and key are correct.")
			fmt.Println("  Support: archscan@proton.me")
			fmt.Println()
			return nil
		}

		if err := lic.Save(); err != nil {
			return fmt.Errorf("failed to save license: %w", err)
		}

		fmt.Println()
		green.Println("  ✓ archscan Pro activated!")
		fmt.Printf("  Email:   %s\n", lic.Email)
		fmt.Printf("  Plan:    %s\n", strings.ToUpper(string(lic.Plan)))
		fmt.Printf("  Expires: %s (%d days remaining)\n",
			lic.Expiry.Format("2006-01-02"), lic.DaysRemaining())
		fmt.Println()
		return nil
	},
}

var deactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Remove the stored Pro license",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := license.Remove(); err != nil {
			return fmt.Errorf("could not remove license: %w", err)
		}
		color.New(color.FgYellow).Println("  License removed. archscan is now running in Free mode.")
		return nil
	},
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "Show current license status",
	Run: func(cmd *cobra.Command, args []string) {
		lic := license.Load()
		bold := color.New(color.Bold)
		green := color.New(color.FgGreen, color.Bold)
		yellow := color.New(color.FgYellow, color.Bold)

		fmt.Println()
		if lic.IsPro() {
			green.Println("  ✓ archscan Pro — Active")
			fmt.Printf("  Email:   %s\n", lic.Email)
			fmt.Printf("  Expires: %s (%d days remaining)\n",
				lic.Expiry.Format("2006-01-02"), lic.DaysRemaining())
		} else {
			yellow.Println("  archscan Free")
			bold.Println("\n  Upgrade to Pro to unlock:")
			fmt.Println("    • Boundary violation detection (architectural layers)")
			fmt.Println("    • Generate .cursorrules + CLAUDE.md for your AI tools")
			fmt.Println("    • JSON & Markdown output (for CI/CD)")
			fmt.Println("    • Unlimited file scanning")
			fmt.Println()
			fmt.Println("  → https://polar.sh/archscan  ($9/mo or $79/yr)")
		}
		fmt.Println()
	},
}

// generateKeyCmd is a hidden command for the seller to generate keys.
var generateKeyCmd = &cobra.Command{
	Use:    "gen-key",
	Hidden: true, // Not shown in help
	Short:  "Generate a Pro license key (seller only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		days, _ := cmd.Flags().GetInt("days")

		if email == "" {
			return fmt.Errorf("--email is required")
		}
		if days == 0 {
			days = 365
		}

		key, err := license.GenerateKey(email, license.PlanPro, days)
		if err != nil {
			return err
		}

		fmt.Printf("\nLicense Key for %s (%d days):\n\n  %s\n\n", email, days, key)
		return nil
	},
}

func init() {
	activateCmd.Flags().String("email", "", "Email used for purchase")
	activateCmd.Flags().String("key", "", "License key (ARCHSCAN-XXXX-XXXX-XXXX-XXXX)")
	activateCmd.MarkFlagRequired("email")
	activateCmd.MarkFlagRequired("key")

	generateKeyCmd.Flags().String("email", "", "Customer email")
	generateKeyCmd.Flags().Int("days", 365, "License validity in days")

	rootCmd.AddCommand(activateCmd)
	rootCmd.AddCommand(deactivateCmd)
	rootCmd.AddCommand(licenseCmd)
	rootCmd.AddCommand(generateKeyCmd)
}
