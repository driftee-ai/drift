//go:build eval

package evals_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/google/generative-ai-go/genai"
	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Name     string `yaml:"name"`
	Files    []File `yaml:"files"`
	Expected struct {
		HasDrift            bool   `yaml:"has_drift"`
		DriftReasonContains string `yaml:"drift_reason_contains"`
	} `yaml:"expected"`
}

type File struct {
	Path    string `yaml:"path"`
	Content string `yaml:"content"`
}

func TestEvals(t *testing.T) {
	testDataDir := "testdata"
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	docAssessor, err := llm.New("gemini")
	if err != nil {
		t.Fatalf("failed to create llm client: %v", err)
	}

	var totalPromptTokens, totalCompletionTokens int
	var truePositives, falsePositives, trueNegatives, falseNegatives int

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		filePath := filepath.Join(testDataDir, entry.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read test case %s: %v", filePath, err)
		}

		var tc TestCase
		if err := yaml.Unmarshal(data, &tc); err != nil {
			t.Fatalf("Failed to parse test case %s: %v", filePath, err)
		}

		t.Run(tc.Name, func(t *testing.T) {
			codeStr := ""
			docStr := ""

			for _, file := range tc.Files {
				if strings.HasSuffix(file.Path, ".md") || strings.HasSuffix(file.Path, ".mdx") {
					docStr += file.Content + "\n"
				} else {
					codeStr += fmt.Sprintf("\n--- Code file: %s ---\n%s\n", file.Path, file.Content)
				}
			}

			// Duplicating the core drift logic prompt to isolate testing
			prompt := fmt.Sprintf(`You are a senior software engineer reviewing documentation for a codebase.
Your task is to determine if the documentation is in sync with the code.

Here is the documentation:
---
%s
---

And here is the code:
%s
`, docStr, codeStr)

			schema := &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"is_in_sync": {
						Type:        genai.TypeBoolean,
						Description: "True if the documentation accurately matches the code signature and description, false otherwise.",
					},
					"reason": {
						Type:        genai.TypeString,
						Description: "A short sentence explaining why they are or are not in sync.",
					},
				},
				Required: []string{"is_in_sync", "reason"},
			}

			jsonRes, usage, err := docAssessor.GenerateJSON(context.Background(), prompt, schema)
			if err != nil {
				t.Fatalf("LLM call failed: %v", err)
			}

			totalPromptTokens += usage.PromptTokens
			totalCompletionTokens += usage.CompletionTokens

			type AssessmentResult struct {
				IsInSync bool   `json:"is_in_sync"`
				Reason   string `json:"reason"`
			}
			
			cleanJson := strings.TrimPrefix(strings.TrimSpace(jsonRes), "```json")
			cleanJson = strings.TrimPrefix(cleanJson, "```")
			cleanJson = strings.TrimSuffix(cleanJson, "```")
			cleanJson = strings.TrimSpace(cleanJson)

			var result AssessmentResult
			if err := json.Unmarshal([]byte(cleanJson), &result); err != nil {
				t.Fatalf("Failed to parse json response: %v\nRaw: %s", err, cleanJson)
			}

			actualHasDrift := !result.IsInSync
			
			// Scoring logic
			if actualHasDrift && tc.Expected.HasDrift {
				truePositives++
			} else if actualHasDrift && !tc.Expected.HasDrift {
				falsePositives++
				t.Errorf("False Positive: Expected In_Sync but got Drift Detected. Reason: %s", result.Reason)
			} else if !actualHasDrift && tc.Expected.HasDrift {
				falseNegatives++
				t.Errorf("False Negative: Expected Drift Detected but got In_Sync. Reason: %s", result.Reason)
			} else {
				trueNegatives++
			}

			if actualHasDrift && tc.Expected.DriftReasonContains != "" {
				if !strings.Contains(strings.ToLower(result.Reason), strings.ToLower(tc.Expected.DriftReasonContains)) {
					t.Errorf("Expected reason to contain %q, but got %q", tc.Expected.DriftReasonContains, result.Reason)
				}
			}
		})
	}

	t.Cleanup(func() {
		precision := 0.0
		if truePositives+falsePositives > 0 {
			precision = float64(truePositives) / float64(truePositives+falsePositives)
		}
		recall := 0.0
		if truePositives+falseNegatives > 0 {
			recall = float64(truePositives) / float64(truePositives+falseNegatives)
		}
		f1 := 0.0
		if precision+recall > 0 {
			f1 = 2 * (precision * recall) / (precision + recall)
		}

		fmt.Printf("\n=== Eval Summary ===\n")
		fmt.Printf("Total Test Cases: %d\n", truePositives+falsePositives+trueNegatives+falseNegatives)
		fmt.Printf("Token Usage:      %d Prompt / %d Completion\n", totalPromptTokens, totalCompletionTokens)
		fmt.Printf("True Positives:   %d\n", truePositives)
		fmt.Printf("False Positives:  %d\n", falsePositives)
		fmt.Printf("True Negatives:   %d\n", trueNegatives)
		fmt.Printf("False Negatives:  %d\n", falseNegatives)
		fmt.Printf("Precision:        %.4f\n", precision)
		fmt.Printf("Recall:           %.4f\n", recall)
		fmt.Printf("F1 Score:         %.4f\n", f1)
		fmt.Printf("====================\n")
	})
}
