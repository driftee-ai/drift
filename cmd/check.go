package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Checks for drift between your code and your documentation.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("drift check called")
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}
