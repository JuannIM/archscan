package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"archscan/internal/analyzer"
	"archscan/internal/config"
	"archscan/internal/report"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Watch a directory and re-scan on file changes",
	Long: `Watch a repository for changes and automatically re-run the architectural
scan whenever a source file is modified. Perfect for live feedback while coding.

  archscan watch /path/to/repo`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			return fmt.Errorf("invalid path: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", absPath)
		}

		cyan := color.New(color.FgCyan, color.Bold)
		dim := color.New(color.Faint)

		cyan.Printf("\n👁  Watching %s for changes...\n", absPath)
		dim.Println("   Press Ctrl+C to stop.")

		// Run initial scan
		runScan(absPath)

		// Set up filesystem watcher
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("could not create watcher: %w", err)
		}
		defer watcher.Close()

		// Watch all subdirectories recursively
		err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				name := info.Name()
				if name == ".git" || name == "node_modules" || name == "vendor" ||
					name == ".venv" || name == "dist" || name == "build" || name == ".cache" {
					return filepath.SkipDir
				}
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("could not set up directory watch: %w", err)
		}

		// Debounce timer — wait 500ms after last event before re-scanning
		var debounce *time.Timer
		sourceExts := map[string]bool{
			".go": true, ".py": true, ".ts": true,
			".tsx": true, ".js": true, ".jsx": true, ".java": true, ".rs": true,
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				// Only react to Write events on source files
				if event.Op&fsnotify.Write == 0 {
					continue
				}
				ext := strings.ToLower(filepath.Ext(event.Name))
				if !sourceExts[ext] {
					continue
				}
				// Reset debounce timer
				if debounce != nil {
					debounce.Stop()
				}
				changedFile := event.Name
				debounce = time.AfterFunc(500*time.Millisecond, func() {
					dim.Printf("\n   ↺  Change detected: %s\n", filepath.Base(changedFile))
					runScan(absPath)
				})

			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				fmt.Fprintf(os.Stderr, "   [watcher error] %v\n", err)
			}
		}
	},
}

// runScan clears the terminal and performs a fresh scan of the repo.
func runScan(repoPath string) {
	// Clear terminal
	fmt.Print("\033[H\033[2J")

	cyan := color.New(color.FgCyan, color.Bold)
	dim := color.New(color.Faint)

	now := time.Now().Format("15:04:05")
	cyan.Printf("\n⚡ archscan — Architectural Drift Detector")
	dim.Printf("  [last scan: %s]\n", now)
	fmt.Printf("   Scanning: %s\n\n", repoPath)

	cfg := config.Load(repoPath)
	result, err := analyzer.Analyze(repoPath, false, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "   analysis failed: %v\n", err)
		return
	}

	r := report.New(result, "text")
	if err := r.Render(os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "   report error: %v\n", err)
	}

	cyan.Println("\n👁  Watching for changes... (Ctrl+C to stop)")
}

func init() {
	rootCmd.AddCommand(watchCmd)
}
