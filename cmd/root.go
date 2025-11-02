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

func init() {
	rootCmd.PersistentFlags().BoolP("version", "v", false, "Print the version and exit")
}

var rootCmd = &cobra.Command{
	Use:   "drift",
	Short: "Drift is a tool for detecting and preventing drift between your code and your documentation.",
	Version: fmt.Sprintf("drift version %s, commit %s, built at %s", version, commit, date),
	Run: func(cmd *cobra.Command, args []string) {
		if v, _ := cmd.Flags().GetBool("version"); v {
			fmt.Println(cmd.Version)
			return
		}
		fmt.Println("Hello from drift!")
	},
}

func Execute() error {
	return rootCmd.Execute()
}
