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

	"github.com/driftee-ai/drift/pkg/checker"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/driftee-ai/drift/pkg/rules"
	"github.com/driftee-ai/drift/pkg/testutil"
	"github.com/google/generative-ai-go/genai"
	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Name       string `yaml:"name"`
	Type       string `yaml:"type,omitempty"` // "remote_repo" or empty for inline
	Repository string `yaml:"repository,omitempty"`
	BaseSHA    string `yaml:"base_sha,omitempty"`
	CommitSHA  string `yaml:"commit_sha,omitempty"`
	Files      []File `yaml:"files"`
	Diff       string `yaml:"diff,omitempty"`
	Expected   struct {
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
	var testFiles []string
	err := filepath.WalkDir(testDataDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(d.Name(), ".yaml") || strings.HasSuffix(d.Name(), ".yml")) {
			if d.Name() != ".drift.yaml" {
				testFiles = append(testFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk testdata directory: %v", err)
	}

	docAssessor, err := llm.New("gemini")
	if err != nil {
		t.Fatalf("failed to create llm client: %v", err)
	}

	var totalPromptTokens, totalCompletionTokens int
	var truePositives, falsePositives, trueNegatives, falseNegatives int
	var diffTruePositives, diffFalsePositives, diffTrueNegatives, diffFalseNegatives int

	for _, filePath := range testFiles {
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

			if tc.Type == "remote_repo" {
				evalRemoteRepo(t, tc, filePath, docAssessor, &truePositives, &falsePositives, &trueNegatives, &falseNegatives, &diffTruePositives, &diffFalsePositives, &diffTrueNegatives, &diffFalseNegatives)
				return // we handle parsing and evaluation in evalRemoteRepo
			}

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
				t.Logf("False Positive: Expected In_Sync but got Drift Detected. Reason: %s", result.Reason)
			} else if !actualHasDrift && tc.Expected.HasDrift {
				falseNegatives++
				t.Logf("False Negative: Expected Drift Detected but got In_Sync. Reason: %s", result.Reason)
			} else {
				trueNegatives++
			}

			if tc.Expected.DiffCausedDrift != nil {
				if result.IsDriftCausedByDiff == nil {
					t.Logf("Expected IsDriftCausedByDiff to be present, but it was nil")
				} else {
					actualDiffCausedDrift := *result.IsDriftCausedByDiff
					expectedDiffCausedDrift := *tc.Expected.DiffCausedDrift

					if actualDiffCausedDrift && expectedDiffCausedDrift {
						diffTruePositives++
					} else if actualDiffCausedDrift && !expectedDiffCausedDrift {
						diffFalsePositives++
						t.Logf("Diff False Positive: Expected diff to NOT cause drift, but LLM said it did. Reason: %s", result.Reason)
					} else if !actualDiffCausedDrift && expectedDiffCausedDrift {
						diffFalseNegatives++
						t.Logf("Diff False Negative: Expected diff to cause drift, but LLM said it didn't. Reason: %s", result.Reason)
					} else {
						diffTrueNegatives++
					}
				}
			}

			if actualHasDrift && tc.Expected.DriftReasonContains != "" {
				if !strings.Contains(strings.ToLower(result.Reason), strings.ToLower(tc.Expected.DriftReasonContains)) {
					t.Logf("Expected reason to contain %q, but got %q", tc.Expected.DriftReasonContains, result.Reason)
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

		// Assert Thresholds
		if f1 < 0.95 {
			t.Errorf("Eval F1 Score (%.4f) fell below the 0.95 threshold", f1)
		}
	})
}

// evalRemoteRepo handles the complex fetching, extraction, and evaluation of a real-world repository
func evalRemoteRepo(
	t *testing.T,
	tc TestCase,
	tcFilePath string,
	docAssessor llm.Client,
	truePositives, falsePositives, trueNegatives, falseNegatives *int,
	diffTruePositives, diffFalsePositives, diffTrueNegatives, diffFalseNegatives *int,
) {
	if tc.Repository == "" || tc.CommitSHA == "" {
		t.Fatalf("remote_repo test cases require 'repository' and 'commit_sha'")
	}

	evalDir, diffContext, changedFiles, err := testutil.CheckoutAndDiff(tc.Repository, tc.BaseSHA, tc.CommitSHA, ".eval_repos")
	if err != nil {
		t.Fatalf("Failed to checkout and calculate diff: %v", err)
	}

	// 3. Inject the .drift.yaml configuration from the test directory
	testDir := filepath.Dir(tcFilePath)
	sourceConfigPath := filepath.Join(testDir, ".drift.yaml")
	if _, err := os.Stat(sourceConfigPath); os.IsNotExist(err) {
		t.Fatalf("Expected configuration file %s not found. Every remote_repo test must have an alongside .drift.yaml", sourceConfigPath)
	}

	configData, err := os.ReadFile(sourceConfigPath)
	if err != nil {
		t.Fatalf("Failed to read test configuration %s: %v", sourceConfigPath, err)
	}

	configPath := filepath.Join(evalDir, ".drift.yaml")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to inject .drift.yaml into eval repository: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config file %s: %v", configPath, err)
	}

	// 5. Trigger rules
	triggeredRules, err := rules.FilterTriggeredRules(cfg.Rules, changedFiles)
	if err != nil {
		t.Fatalf("failed to filter rules based on changed files: %v", err)
	}

	if len(triggeredRules) == 0 {
		t.Fatalf("Test setup failure: the commit %s did not trigger any drift rules in the native configuration", tc.CommitSHA)
	}

	// 6. Evaluate via pkg/checker
	t.Logf("Running checker using %s...", configPath)
	chk := checker.New(docAssessor, cfg.Provider)
	diffOnly := true // Always test real commits in diffOnly mode to ensure exact attribution
	results := chk.EvaluateRules(context.Background(), triggeredRules, evalDir, diffOnly, diffContext)

	// We'll consider it "drift" if ANY rule failed and was not ignored due to diff
	actualHasDrift := false
	var reasons []string
	var diffCausedCount int

	for _, result := range results {
		if result.Error != nil {
			t.Fatalf("Checker error on rule %s: %v", result.Rule.Name, result.Error)
		}
		if result.Skipped {
			continue
		}

		if !result.IsInSync {
			actualHasDrift = true
			reasons = append(reasons, result.Reason)
			diffCausedCount++
		} else if result.IgnoredDueToDiff {
			actualHasDrift = true // Technically there was drift
			reasons = append(reasons, result.Reason)
			// diffCausedCount implies the diff didn't cause it, so we leave it 0
		}
	}

	combinedReason := strings.Join(reasons, " | ")
	actualDiffCausedDrift := diffCausedCount > 0

	// 7. Scoring Logic
	if actualHasDrift && tc.Expected.HasDrift {
		*truePositives++
	} else if actualHasDrift && !tc.Expected.HasDrift {
		*falsePositives++
		t.Logf("False Positive: Expected In_Sync but got Drift Detected. Reason: %s", combinedReason)
	} else if !actualHasDrift && tc.Expected.HasDrift {
		*falseNegatives++
		t.Logf("False Negative: Expected Drift Detected but got In_Sync. Reason: %s", combinedReason)
	} else {
		*trueNegatives++
	}

	if tc.Expected.DiffCausedDrift != nil {
		expectedDiffCausedDrift := *tc.Expected.DiffCausedDrift

		if actualDiffCausedDrift && expectedDiffCausedDrift {
			*diffTruePositives++
		} else if actualDiffCausedDrift && !expectedDiffCausedDrift {
			*diffFalsePositives++
			t.Logf("Diff False Positive: Expected diff to NOT cause drift, but LLM said it did. Reason: %s", combinedReason)
		} else if !actualDiffCausedDrift && expectedDiffCausedDrift {
			*diffFalseNegatives++
			t.Logf("Diff False Negative: Expected diff to cause drift, but LLM said it didn't. Reason: %s", combinedReason)
		} else {
			*diffTrueNegatives++
		}
	}

	if actualHasDrift && tc.Expected.DriftReasonContains != "" {
		if !strings.Contains(strings.ToLower(combinedReason), strings.ToLower(tc.Expected.DriftReasonContains)) {
			t.Logf("Expected reason to contain %q, but got %q", tc.Expected.DriftReasonContains, combinedReason)
		}
	}
}
