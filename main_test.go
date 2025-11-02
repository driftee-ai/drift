package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

const (
	testBinaryName = "drift_test"
)

// TestMain is the entry point for the tests.
// It builds the test binary and cleans it up after the tests are done.
func TestMain(m *testing.M) {
	// Build the test binary
	cmd := exec.Command("go", "build", "-o", testBinaryName, ".")
	if err := cmd.Run(); err != nil {
		panic("Failed to build test binary: " + err.Error())
	}

	// Run the tests
	exitCode := m.Run()

	// Clean up the test binary
	os.Remove(testBinaryName)

	os.Exit(exitCode)
}

func TestCheckCommand_DummyProvider(t *testing.T) {
	// Run the check command with the dummy provider config
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.dummy.yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(output))
	}

	// Assert on the output
	expectedOutput := "Result: In Sync (This is a dummy assessment.)"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}

func TestCheckCommand_GeminiProvider_NoApiKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	// Run the check command with the gemini provider config, without the API key
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.test.yaml")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("check command should have failed, but it didn't.\nOutput:\n%s", string(output))
	}

	// Assert on the output
	expectedOutput := "GEMINI_API_KEY environment variable not set"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}

func TestCheckCommand_GeminiProvider_WithApiKey(t *testing.T) {
	t.Log("Running live Gemini API test: WithApiKey") // Added log

	// Skip this test if GEMINI_API_KEY is not set
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("GEMINI_API_KEY not set, skipping live Gemini API test")
	}

	// Run the check command with the gemini provider config and API key
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.test.yaml")
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	output, err := cmd.CombinedOutput()
	t.Logf("Command output:\n%s", string(output)) // Added log

	if err != nil {
		t.Fatalf("check command failed: %v", err)
	}

	// Assert on the output (assuming test data is in sync)
	expectedOutput := "Result: In Sync"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}

func TestCheckCommand_CatchesMissingParam(t *testing.T) {
	t.Log("Running live Gemini API test: CatchesMissingParam") // Added log

	// Skip this test if GEMINI_API_KEY is not set
	if os.Getenv("GEMINI_API_KEY") == "" {
		t.Skip("GEMINI_API_KEY not set, skipping live Gemini API test")
	}

	// Run the check command with the config file for the drift example
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/e2e/true_positives/missing_param_in_docs/.drift.yaml")
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	output, err := cmd.CombinedOutput()
	t.Logf("Command output:\n%s", string(output)) // Added log

	if err == nil {
		t.Fatalf("check command should have failed, but it didn't.")
	}

	// Assert on the output
	expectedOutput := "Result: In Sync"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}
