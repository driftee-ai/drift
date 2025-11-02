package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drift",
	Short: "Drift is a tool for detecting and preventing drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from drift!")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
