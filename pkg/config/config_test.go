package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driftee-ai/drift/pkg/config"
)

func TestCreateScaffold(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "drift_test_config")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	scaffoldPath := filepath.Join(tmpDir, ".drift.yaml")

	err = config.CreateScaffold(scaffoldPath)
	if err != nil {
		t.Fatalf("CreateScaffold failed: %v", err)
	}

	content, err := os.ReadFile(scaffoldPath)
	if err != nil {
		t.Fatalf("Failed to read scaffold file: %v", err)
	}

	expectedContent := `version: 1
provider: gemini
rules:
    - name: Example API Documentation
      code:
        - src/api/**/*.go
      docs:
        - docs/api/**/*.md
`
	actualContent := string(content)
	actualContent = removeComments(actualContent)
	actualContent = strings.TrimSpace(actualContent)
	expectedContent = strings.TrimSpace(expectedContent)

	if actualContent != expectedContent {
		t.Errorf("Generated scaffold content mismatch.\nExpected:\n%s\nActual:\n%s", expectedContent, actualContent)
	}
}

func TestLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "drift_test_load")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, ".drift.yaml")

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

	loadedConfig, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

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

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := config.Load("nonexistent.yaml")
	if err == nil {
		t.Error("Expected an error when loading a non-existent file, got nil")
	}
}

func TestLoad_MalformedYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "drift_test_malformed")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	configPath := filepath.Join(tmpDir, ".drift.yaml")
	testConfig := `
version: 1
rules:
  - name: Test Rule
  invalid_yaml: [
`
	if err := os.WriteFile(configPath, []byte(testConfig), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	_, err = config.Load(configPath)
	if err == nil {
		t.Error("Expected an error when loading malformed YAML, got nil")
	}
}

func TestLoad_FallbackFiles(t *testing.T) {
	// Fallbacks are checked if passed path is strictly ".drift.yaml" and it doesn't exist.
	// Since os.ReadFile reads from CWD when passed a relative path, we need to create
	// the fallback file in the current working directory of the test temporarily.

	fallbackName := "drift.yml"
	testConfig := `
version: 2
rules:
  - name: Fallback Rule
`
	err := os.WriteFile(fallbackName, []byte(testConfig), 0644)
	if err != nil {
		t.Fatalf("Failed to create fallback file: %v", err)
	}
	defer os.Remove(fallbackName)

	loadedConfig, err := config.Load(".drift.yaml")
	if err != nil {
		t.Fatalf("Failed to load fallback config: %v", err)
	}

	if loadedConfig.Version != 2 {
		t.Errorf("Expected version 2 from fallback, got %d", loadedConfig.Version)
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
