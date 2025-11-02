package cmd

import (
	"fmt"
	"log"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks for drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load(".drift.yaml")
		if err != nil {
			log.Fatalf("Failed to load .drift.yaml: %v", err)
		}

		fmt.Printf("Loaded %d rules from .drift.yaml\n", len(cfg.Rules))
		for _, rule := range cfg.Rules {
			fmt.Printf("  - Rule: %s\n", rule.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
