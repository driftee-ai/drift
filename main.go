package main

import (
	"os"

	"github.com/driftee-ai/drift/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
