package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "drift",
	Short: "Drift is a tool for detecting and preventing drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from drift!")
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the version, commit, and build date",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("drift version %s, commit %s, built at %s\n", version, commit, date)
	},
}

func Execute() error {
	rootCmd.AddCommand(versionCmd)
	return rootCmd.Execute()
}
