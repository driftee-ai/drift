package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/driftee-ai/drift/pkg/tui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new .drift.yaml configuration file interactively.",
	Run: func(cmd *cobra.Command, args []string) {
		// Create a new Bubble Tea program
		p := tea.NewProgram(tui.NewModel(), tea.WithAltScreen())

		// Run the TUI program
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running init wizard: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
