package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestRootCmd(t *testing.T) {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{})
	Execute()

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			// Don't call t.Fatalf here, just send an empty string
			outC <- ""
			return
		}
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	if !strings.Contains(out, "Hello from drift!") {
		t.Errorf("Expected 'Hello from drift!' but got '%s'", out)
	}
}
