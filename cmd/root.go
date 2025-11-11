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
	Version: fmt.Sprintf("drift version %s, commit %s, built at %s", version, commit, date),
}

func Execute() error {
	return rootCmd.Execute()
}
