//go:build integration

package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestChangedFiles_NoFlag(t *testing.T) {
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.filter-test.yaml")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(output))
	}

	// Expect both rules to be checked
	if !strings.Contains(string(output), "Rule-Go") {
		t.Errorf("Expected output to contain 'Rule-Go', but got:\n%s", string(output))
	}
	if !strings.Contains(string(output), "Rule-MD") {
		t.Errorf("Expected output to contain 'Rule-MD', but got:\n%s", string(output))
	}
	// Check that the filtering message is NOT present
	if strings.Contains(string(output), "rules were triggered") {
		t.Errorf("Expected output NOT to contain 'rules were triggered', but it did")
	}
}

func TestChangedFiles_TriggerOneRule(t *testing.T) {
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.filter-test.yaml", "--changed-files", "testdata/src/api/user.go")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(output))
	}

	// Expect only Rule-Go to be checked
	if !strings.Contains(string(output), "Rule-Go") {
		t.Errorf("Expected output to contain 'Rule-Go', but got:\n%s", string(output))
	}
	if strings.Contains(string(output), "Rule-MD") {
		t.Errorf("Expected output NOT to contain 'Rule-MD', but got:\n%s", string(output))
	}
	// Check that the filtering message IS present
	if !strings.Contains(string(output), "1 rules were triggered") {
		t.Errorf("Expected output to contain '1 rules were triggered', but it did not")
	}
}

func TestChangedFiles_TriggerOtherRule(t *testing.T) {
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.filter-test.yaml", "--changed-files", "README.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(output))
	}

	// Expect only Rule-MD to be checked
	if strings.Contains(string(output), "Rule-Go") {
		t.Errorf("Expected output NOT to contain 'Rule-Go', but got:\n%s", string(output))
	}
	if !strings.Contains(string(output), "Rule-MD") {
		t.Errorf("Expected output to contain 'Rule-MD', but got:\n%s", string(output))
	}
	if !strings.Contains(string(output), "1 rules were triggered") {
		t.Errorf("Expected output to contain '1 rules were triggered', but it did not")
	}
}

func TestChangedFiles_TriggerNoRules(t *testing.T) {
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.filter-test.yaml", "--changed-files", "Makefile")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(output))
	}

	// Expect no rules to be checked
	if strings.Contains(string(output), "Rule-Go") {
		t.Errorf("Expected output NOT to contain 'Rule-Go', but got:\n%s", string(output))
	}
	if strings.Contains(string(output), "Rule-MD") {
		t.Errorf("Expected output NOT to contain 'Rule-MD', but got:\n%s", string(output))
	}
	if !strings.Contains(string(output), "0 rules were triggered") {
		t.Errorf("Expected output to contain '0 rules were triggered', but it did not")
	}
}

func TestChangedFiles_TriggerBothRules(t *testing.T) {
	cmd := exec.Command("./"+testBinaryName, "check", "--config", "testdata/.drift.filter-test.yaml", "--changed-files", "testdata/src/api/user.go,README.md")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("check command failed: %v\nOutput:\n%s", err, string(output))
	}

	// Expect both rules to be checked
	if !strings.Contains(string(output), "Rule-Go") {
		t.Errorf("Expected output to contain 'Rule-Go', but got:\n%s", string(output))
	}
	if !strings.Contains(string(output), "Rule-MD") {
		t.Errorf("Expected output to contain 'Rule-MD', but got:\n%s", string(output))
	}
	if !strings.Contains(string(output), "2 rules were triggered") {
		t.Errorf("Expected output to contain '2 rules were triggered', but it did not")
	}
}
