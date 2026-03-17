package cmd

import (
	"fmt"
	"os"

	"github.com/driftee-ai/drift/pkg/initwizard"
	"github.com/spf13/cobra"
)

var initFastMode bool
var initDir string
var initNonInteractive bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Interactive wizard to bootstrap your .drift.yaml configuration.",
	Run: func(cmd *cobra.Command, args []string) {
		err := initwizard.RunWizard(initDir, initFastMode, initNonInteractive)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initFastMode, "fast", false, "Use filename-only analysis for faster (but potentially less accurate) rule discovery")
	initCmd.Flags().BoolVarP(&initNonInteractive, "non-interactive", "y", false, "Bypass the interactive wizard and automatically generate and save mapping rules")
	initCmd.Flags().StringVarP(&initDir, "dir", "d", ".", "Directory to analyze")
}
