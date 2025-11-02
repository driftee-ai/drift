package cmd

import (
	"fmt"
	"log"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/files"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks for drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config") // Get the config file path from the flag

		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("Failed to load config file %s: %v", configFile, err)
		}

		fmt.Printf("Loaded %d rules from %s\n", len(cfg.Rules), configFile) // Updated print statement
		for _, rule := range cfg.Rules {
			fmt.Printf("  - Rule: %s\n", rule.Name)

			// Find and read code files
			codeFiles, err := files.FindFiles(rule.Code)
			if err != nil {
				log.Printf("Error finding code files for rule '%s': %v", rule.Name, err)
				continue
			}
			codeContent, err := files.ReadAndConcatenate(codeFiles)
			if err != nil {
				log.Printf("Error reading code content for rule '%s': %v", rule.Name, err)
				continue
			}
			fmt.Printf("    Found %d code files, total size: %d bytes\n", len(codeFiles), len(codeContent))

			// Find and read docs files
			docFiles, err := files.FindFiles(rule.Docs)
			if err != nil {
				log.Printf("Error finding doc files for rule '%s': %v", rule.Name, err)
				continue
			}
			docContent, err := files.ReadAndConcatenate(docFiles)
			if err != nil {
				log.Printf("Error reading doc content for rule '%s': %v", rule.Name, err)
				continue
			}
			fmt.Printf("    Found %d doc files, total size: %d bytes\n", len(docFiles), len(docContent))
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
	checkCmd.Flags().StringP("config", "c", ".drift.yaml", "Path to the drift configuration file") // Added flag
}
