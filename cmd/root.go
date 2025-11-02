package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "drift",
	Short: "Drift is a tool for detecting and preventing drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from drift!")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
