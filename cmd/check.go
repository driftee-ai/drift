package cmd

import (
	"fmt"
	"io"
	"log"
	"os"
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
		changedFiles, _ := cmd.Flags().GetStringSlice("changed-files")
		diffOnly, _ := cmd.Flags().GetBool("diff-only")
		diffFile, _ := cmd.Flags().GetString("diff")

		var diffContext string
		if diffOnly {
			if diffFile == "" {
				log.Fatal("When --diff-only is provided, you must also provide the diff file path using --diff, or use --diff - to read from stdin.")
			}

			if diffFile == "-" {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					log.Fatalf("Failed to read diff from stdin: %v", err)
				}
				diffContext = string(bytes)
			} else {
				bytes, err := os.ReadFile(diffFile)
				if err != nil {
					log.Fatalf("Failed to read diff file %s: %v", diffFile, err)
				}
				diffContext = string(bytes)
			}

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

		// Define the Response Schema structure
		// (Removed, moved to pkg/checker)

		chk := checker.New(docAssessor, cfg.Provider)
		results := chk.EvaluateRules(cmd.Context(), triggeredRules, diffOnly, diffContext)

		for _, result := range results {
			fmt.Printf("  - Rule: %s\n", result.Rule.Name)

			if result.Error != nil && strings.Contains(result.Error.Error(), "missing code or doc files") {
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
	checkCmd.Flags().StringSliceP("changed-files", "f", []string{}, "List of changed files to check for drift")
	checkCmd.Flags().Bool("diff-only", false, "Only fail if the detected drift was caused by the provided git diff")
	checkCmd.Flags().String("diff", "", "Path to a file containing the git diff, or '-' to read from stdin (required if --diff-only is used)")
}
