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
provider: gemini
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

func TestLoad(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "drift_test_load")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Define the path for the test config file
	configPath := filepath.Join(tmpDir, ".drift.yaml")

	// Create a test config file
	testConfig := `
version: 1
rules:
  - name: Test Rule
    code:
      - "src/test.go"
    docs:
      - "docs/test.md"
`
	err = os.WriteFile(configPath, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Call the Load function
	loadedConfig, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Assert the loaded config
	if loadedConfig.Version != 1 {
		t.Errorf("Expected version 1, got %d", loadedConfig.Version)
	}
	if len(loadedConfig.Rules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(loadedConfig.Rules))
	}
	rule := loadedConfig.Rules[0]
	if rule.Name != "Test Rule" {
		t.Errorf("Expected rule name 'Test Rule', got '%s'", rule.Name)
	}
	if len(rule.Code) != 1 || rule.Code[0] != "src/test.go" {
		t.Errorf("Expected code 'src/test.go', got %v", rule.Code)
	}
	if len(rule.Docs) != 1 || rule.Docs[0] != "docs/test.md" {
		t.Errorf("Expected docs 'docs/test.md', got %v", rule.Docs)
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
