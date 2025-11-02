package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new .drift.yaml configuration file.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("drift init called")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
