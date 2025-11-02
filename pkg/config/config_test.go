package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driftee-ai/drift/pkg/config"
)

func TestCreateScaffold(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "drift_test_config")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir) // Clean up the temporary directory

	// Define the path for the scaffold file
	scaffoldPath := filepath.Join(tmpDir, ".drift.yaml")

	// Call the function to create the scaffold
	err = config.CreateScaffold(scaffoldPath)
	if err != nil {
		t.Fatalf("CreateScaffold failed: %v", err)
	}

	// Read the content of the created file
	content, err := os.ReadFile(scaffoldPath)
	if err != nil {
		t.Fatalf("Failed to read scaffold file: %v", err)
	}

	// Define the expected content (without comments for easier comparison)
	expectedContent := `version: 1
rules:
    - name: Example API Documentation
      code:
        - src/api/**/*.go
      docs:
        - docs/api/**/*.md
`
	// Remove comments and leading/trailing whitespace for comparison
	actualContent := string(content)
	actualContent = removeComments(actualContent)
	actualContent = strings.TrimSpace(actualContent)
	expectedContent = strings.TrimSpace(expectedContent)


	if actualContent != expectedContent {
		t.Errorf("Generated scaffold content mismatch.\nExpected:\n%s\nActual:\n%s", expectedContent, actualContent)
	}
}

// Helper function to remove comments from the YAML string
func removeComments(s string) string {
	lines := strings.Split(s, "\n")
	var cleanLines []string
	for _, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "#") {
			cleanLines = append(cleanLines, line)
		}
	}
	return strings.Join(cleanLines, "\n")
}
