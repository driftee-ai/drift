package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/driftee-ai/drift/pkg/assessor"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/files"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks for drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")

		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("failed to load config file %s: %v", configFile, err)
		}

		docAssessor, err := assessor.New(cfg.Provider)
		if err != nil {
			log.Fatalf("failed to create assessor: %v", err)
		}

		fmt.Printf("Loaded %d rules from %s (provider: %s)\n", len(cfg.Rules), configFile, cfg.Provider)
		allInSync := true
		for _, rule := range cfg.Rules {
			fmt.Printf("  - Rule: %s\n", rule.Name)

			// Find and read code files
			codeFiles, err := files.FindFiles(rule.Code)
			if err != nil {
				log.Printf("Error finding code files for rule '%s': %v", rule.Name, err)
				allInSync = false
				continue
			}
			codeContents, err := files.ReadFiles(codeFiles)
			if err != nil {
				log.Printf("Error reading code content for rule '%s': %v", rule.Name, err)
				allInSync = false
				continue
			}
			totalSize := 0
			for _, content := range codeContents {
				totalSize += len(content)
			}
			fmt.Printf("    Found %d code files, total size: %d bytes\n", len(codeFiles), totalSize)

			// Find and read docs files
			docFiles, err := files.FindFiles(rule.Docs)
			if err != nil {
				log.Printf("Error finding doc files for rule '%s': %v", rule.Name, err)
				allInSync = false
				continue
			}
			docContent, err := files.ReadAndConcatenate(docFiles)
			if err != nil {
				log.Printf("Error reading doc content for rule '%s': %v", rule.Name, err)
				allInSync = false
				continue
			}
			fmt.Printf("    Found %d doc files, total size: %d bytes\n", len(docFiles), len(docContent))

			// Assess the drift
			result, err := docAssessor.Assess(docContent, codeContents)
			if err != nil {
				log.Printf("Error assessing drift for rule '%s': %v", rule.Name, err)
				allInSync = false // Consider assessment error as out of sync
				continue
			}

			if result.IsInSync {
				fmt.Printf("    Result: In Sync\n")
			} else {
				fmt.Printf("    Result: Out of Sync (%s)\n", result.Reason)
				allInSync = false // Set flag to false
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
}