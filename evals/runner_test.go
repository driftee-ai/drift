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
	Diff     string `yaml:"diff,omitempty"`
	Expected struct {
		HasDrift            bool   `yaml:"has_drift"`
		DriftReasonContains string `yaml:"drift_reason_contains,omitempty"`
		DiffCausedDrift     *bool  `yaml:"diff_caused_drift,omitempty"`
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
	var diffTruePositives, diffFalsePositives, diffTrueNegatives, diffFalseNegatives int

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

			if tc.Diff != "" {
				prompt += fmt.Sprintf("\nHere is the git diff of the recent changes:\n---\n%s\n---\nBased on the provided diff, did the changes introduced in this diff cause the documentation drift?", tc.Diff)
			}

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

			if tc.Diff != "" {
				schema.Properties["is_drift_caused_by_diff"] = &genai.Schema{
					Type:        genai.TypeBoolean,
					Description: "True if the documentation drift was caused by the provided git diff, false if the drift is preexisting or unrelated.",
				}
				schema.Required = append(schema.Required, "is_drift_caused_by_diff")
			}

			jsonRes, usage, err := docAssessor.GenerateJSON(context.Background(), prompt, schema)
			if err != nil {
				t.Fatalf("LLM call failed: %v", err)
			}

			totalPromptTokens += usage.PromptTokens
			totalCompletionTokens += usage.CompletionTokens

			type AssessmentResult struct {
				IsInSync            bool   `json:"is_in_sync"`
				Reason              string `json:"reason"`
				IsDriftCausedByDiff *bool  `json:"is_drift_caused_by_diff,omitempty"`
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

			if tc.Expected.DiffCausedDrift != nil && result.IsDriftCausedByDiff != nil {
				actualDiffCausedDrift := *result.IsDriftCausedByDiff
				expectedDiffCausedDrift := *tc.Expected.DiffCausedDrift
				
				if actualDiffCausedDrift && expectedDiffCausedDrift {
					diffTruePositives++
				} else if actualDiffCausedDrift && !expectedDiffCausedDrift {
					diffFalsePositives++
					t.Errorf("Diff False Positive: Expected diff to NOT cause drift, but LLM said it did. Reason: %s", result.Reason)
				} else if !actualDiffCausedDrift && expectedDiffCausedDrift {
					diffFalseNegatives++
					t.Errorf("Diff False Negative: Expected diff to cause drift, but LLM said it didn't. Reason: %s", result.Reason)
				} else {
					diffTrueNegatives++
				}
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

		if diffTruePositives+diffFalsePositives+diffTrueNegatives+diffFalseNegatives > 0 {
			diffPrecision := 0.0
			if diffTruePositives+diffFalsePositives > 0 {
				diffPrecision = float64(diffTruePositives) / float64(diffTruePositives+diffFalsePositives)
			}
			diffRecall := 0.0
			if diffTruePositives+diffFalseNegatives > 0 {
				diffRecall = float64(diffTruePositives) / float64(diffTruePositives+diffFalseNegatives)
			}
			diffF1 := 0.0
			if diffPrecision+diffRecall > 0 {
				diffF1 = 2 * (diffPrecision * diffRecall) / (diffPrecision + diffRecall)
			}
			fmt.Printf("\n--- Diff Attribution Metrics ---\n")
			fmt.Printf("True Positives:   %d\n", diffTruePositives)
			fmt.Printf("False Positives:  %d\n", diffFalsePositives)
			fmt.Printf("True Negatives:   %d\n", diffTrueNegatives)
			fmt.Printf("False Negatives:  %d\n", diffFalseNegatives)
			fmt.Printf("Diff Precision:   %.4f\n", diffPrecision)
			fmt.Printf("Diff Recall:      %.4f\n", diffRecall)
			fmt.Printf("Diff F1 Score:    %.4f\n", diffF1)
		}
		
		fmt.Printf("====================\n")
	})
}
