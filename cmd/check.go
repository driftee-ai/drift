package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/files"
	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/driftee-ai/drift/pkg/rules"
	"github.com/google/generative-ai-go/genai"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks for drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		configFile, _ := cmd.Flags().GetString("config")
		changedFiles, _ := cmd.Flags().GetStringSlice("changed-files")

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
		type AssessmentResult struct {
			IsInSync bool   `json:"is_in_sync"`
			Reason   string `json:"reason"`
		}

		for _, rule := range triggeredRules {
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
			codeStr := ""
			for path, content := range codeContents {
				totalSize += len(content)
				codeStr += fmt.Sprintf("\n--- Code file: %s ---\n%s\n", path, content)
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

			if len(codeFiles) == 0 && len(docFiles) == 0 {
				fmt.Printf("    Result: Skipped (no code or doc files found)\n")
				continue
			}
			if len(codeFiles) == 0 || len(docFiles) == 0 {
				fmt.Printf("    Result: Out of Sync (missing code or doc files)\n")
				allInSync = false
				continue
			}

			// Assess the drift
			prompt := fmt.Sprintf(`You are a senior software engineer reviewing documentation for a codebase.
Your task is to determine if the documentation is in sync with the code.

Here is the documentation:
---
%s
---

And here is the code:
%s
`, docContent, codeStr)

			var schema interface{}
			if cfg.Provider == "gemini" {
				schema = &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"is_in_sync": {
							Type:        genai.TypeBoolean,
							Description: "True if the documentation accurately matches the code signature and description, false otherwise.",
						},
						"reason": {
							Type:        genai.TypeString,
							Description: "A short sentence explaining why they are or are not in sync.",
						},
					},
					Required: []string{"is_in_sync", "reason"},
				}
			} else {
				schema = AssessmentResult{}
			}

			jsonRes, _, err := docAssessor.GenerateJSON(cmd.Context(), prompt, schema)
			if err != nil {
				log.Printf("Error assessing drift for rule '%s': %v", rule.Name, err)
				allInSync = false // Consider assessment error as out of sync
				continue
			}

			// Clean raw string output since some models inject Markdown blocks
			var result AssessmentResult

			// We manually strip off standard markdown code block tags if they exist.
			cleanJson := strings.TrimPrefix(strings.TrimSpace(jsonRes), "```json")
			cleanJson = strings.TrimPrefix(cleanJson, "```")
			cleanJson = strings.TrimSuffix(cleanJson, "```")
			cleanJson = strings.TrimSpace(cleanJson)

			if err := json.Unmarshal([]byte(cleanJson), &result); err != nil {
				log.Printf("Error parsing JSON response for rule '%s': %v\nRaw Response: %s", rule.Name, err, jsonRes)
				allInSync = false
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
	checkCmd.Flags().StringSliceP("changed-files", "f", []string{}, "List of changed files to check for drift")
}
