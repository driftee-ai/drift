package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/driftee-ai/drift/pkg/checker"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/driftee-ai/drift/pkg/rules"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks for drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		base, _ := cmd.Flags().GetString("base")
		diffOnly, _ := cmd.Flags().GetBool("diff-only")

		var changedFiles []string
		var diffContext string

		if base == "" {
			// Auto-detect default branch
			branches := []string{"origin/main", "main", "origin/master", "master"}
			for _, b := range branches {
				cmdStr := exec.Command("git", "rev-parse", "--verify", b)
				if err := cmdStr.Run(); err == nil {
					base = b
					fmt.Printf("Auto-detected default branch: %s\n", base)
					break
				}
			}
			if base == "" {
				log.Fatal("Could not auto-detect a default branch (tried origin/main, main, origin/master, master). Please provide one explicitly with --base.")
			}
		}

		// Get changed files
		nameOnlyCmd := exec.Command("git", "diff", "--name-only", fmt.Sprintf("%s...HEAD", base))
		nameOnlyBytes, err := nameOnlyCmd.Output()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				log.Fatalf("failed to calculate changed files against base %s. Git output: %s", base, string(exitErr.Stderr))
			}
			log.Fatalf("failed to calculate changed files against base %s: %v", base, err)
		}

		changedFilesRaw := strings.Split(strings.TrimSpace(string(nameOnlyBytes)), "\n")
		for _, f := range changedFilesRaw {
			f = strings.TrimSpace(f)
			if f != "" {
				changedFiles = append(changedFiles, f)
			}
		}

		if diffOnly {
			diffCmd := exec.Command("git", "diff", fmt.Sprintf("%s...HEAD", base))
			diffBytes, err := diffCmd.Output()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					log.Fatalf("failed to calculate diff against base %s. Git output: %s", base, string(exitErr.Stderr))
				}
				log.Fatalf("failed to calculate diff against base %s: %v", base, err)
			}
			diffContext = string(diffBytes)

			if strings.TrimSpace(diffContext) == "" {
				log.Println("Warning: --diff-only provided but the diff context is empty.")
			}
		}

		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("failed to load config file %s: %v", configFile, err)
		}

		docAssessor, err := llm.New(cfg.Provider)
		if err != nil {
			log.Fatalf("failed to create llm client: %v", err)
		}

		triggeredRules, err := rules.FilterTriggeredRules(cfg.Rules, changedFiles)
		if err != nil {
			log.Fatalf("failed to filter rules based on changed files: %v", err)
		}

		fmt.Printf("Loaded %d rules from %s (provider: %s)\n", len(cfg.Rules), configFile, cfg.Provider)
		if len(changedFiles) > 0 {
			fmt.Printf("Filtering rules based on %d changed files. %d rules were triggered.\n", len(changedFiles), len(triggeredRules))
		}
		allInSync := true

		chk := checker.New(docAssessor, cfg.Provider)
		results := chk.EvaluateRules(cmd.Context(), triggeredRules, diffOnly, diffContext)

		for _, result := range results {
			fmt.Printf("  - Rule: %s\n", result.Rule.Name)

			if errors.Is(result.Error, checker.ErrMissingFiles) {
				fmt.Printf("    Result: Out of Sync (missing code or doc files)\n")
				allInSync = false
				continue
			}

			if result.Skipped {
				fmt.Printf("    Result: Skipped (no code or doc files found)\n")
				continue
			}

			if result.Error != nil {
				fmt.Printf("    Error: %v\n", result.Error)
				allInSync = false
				continue
			}

			fmt.Printf("    Found %d code files, total size: %d bytes\n", result.CodeFilesCount, result.CodeTotalBytes)
			fmt.Printf("    Found %d doc files, total size: %d bytes\n", result.DocsFilesCount, result.DocsTotalBytes)

			if result.IsInSync {
				if result.IgnoredDueToDiff {
					fmt.Printf("    Result: Drift detected, but ignoring since it was not caused by the diff and we are in --diff-only mode. (Reason: %s)\n", result.Reason)
				} else {
					fmt.Printf("    Result: In Sync\n")
				}
			} else {
				fmt.Printf("    Result: Out of Sync (%s)\n", result.Reason)
				allInSync = false
			}
		}

		if !allInSync {
			fmt.Println("Drift detected.")
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringP("config", "c", ".drift.yaml", "Path to the drift configuration file")
	checkCmd.Flags().String("base", "", "Base branch or commit to compare against for changed files and diffs. If omitted, attempts to auto-detect main or master.")
	checkCmd.Flags().Bool("diff-only", false, "Only fail if the detected drift was caused by the diff between the base and HEAD")
}
