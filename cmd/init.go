package cmd

import (
	"fmt"
	"log"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new .drift.yaml configuration file.",
	Run: func(cmd *cobra.Command, args []string) {
		err := config.CreateScaffold(".drift.yaml")
		if err != nil {
			log.Fatalf("Failed to create .drift.yaml: %v", err)
		}
		fmt.Println(".drift.yaml created successfully.")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}