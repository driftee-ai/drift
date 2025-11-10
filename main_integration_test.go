//go:build integration

package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCheckCommand_GeminiProvider_WithApiKey(t *testing.T) {
	t.Log("Running live Gemini API test: WithApiKey") // Added log

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

	// Run the check command with the config file for the drift example
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/e2e/true_positives/missing_param_in_docs/.drift.yaml")
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	output, err := cmd.CombinedOutput()
	t.Logf("Command output:\n%s", string(output)) // Added log

	if err == nil {
		t.Fatalf("check command should have failed, but it didn't.")
	}

	// Assert on the output
	expectedOutput := "Result: Out of Sync"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}

func TestCheckCommand_InSync(t *testing.T) {
	// Run the check command with the config file for the in-sync example
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/e2e/true_negatives/in_sync_example/.drift.yaml")
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	output, err := cmd.CombinedOutput()
	t.Logf("Command output:\n%s", string(output))

	if err != nil {
		t.Fatalf("check command failed: %v", err)
	}

	// Assert on the output
	expectedOutput := "Result: In Sync"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}

func TestCheckCommand_CosmeticDiff(t *testing.T) {
	// Run the check command with the config file for the cosmetic diff example
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/e2e/false_positives/cosmetic_diff_example/.drift.yaml")
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	output, err := cmd.CombinedOutput()
	t.Logf("Command output:\n%s", string(output))

	if err != nil {
		t.Fatalf("check command failed: %v", err)
	}

	// Assert on the output
	expectedOutput := "Result: In Sync"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}

func TestCheckCommand_SubtleDrift(t *testing.T) {
	// Run the check command with the config file for the subtle drift example
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/e2e/false_negatives/subtle_drift_example/.drift.yaml")
	cmd.Env = append(os.Environ(), "GEMINI_API_KEY="+os.Getenv("GEMINI_API_KEY"))

	output, err := cmd.CombinedOutput()
	t.Logf("Command output:\n%s", string(output))

	if err == nil {
		t.Fatalf("check command should have failed, but it didn't.")
	}

	// Assert on the output
	expectedOutput := "Result: Out of Sync"
	if !strings.Contains(string(output), expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, string(output))
	}
}
