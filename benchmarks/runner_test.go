package benchmarks_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/driftee-ai/drift/benchmarks/competitors"
	"github.com/driftee-ai/drift/pkg/testutil"
	"gopkg.in/yaml.v3"
)

type BenchmarkCase struct {
	Name       string               `yaml:"name"`
	Repository string               `yaml:"repository"`
	CommitSHA  string               `yaml:"commit_sha"`
	Expected   BenchmarkExpectation `yaml:"expected"`
}

type BenchmarkExpectation struct {
	HasDrift bool `yaml:"has_drift"`
}

// ScriptCompetitor implements the Competitor interface for a bash script
type ScriptCompetitor struct {
	scriptPath string
	name       string
}

func (s *ScriptCompetitor) Name() string {
	return s.name
}

func (s *ScriptCompetitor) Evaluate(repoDir string, diffFile string) (competitors.EvaluationResult, error) {
	start := time.Now()

	cmd := exec.Command(s.scriptPath, repoDir, diffFile)
	output, err := cmd.CombinedOutput()

	duration := time.Since(start).Milliseconds()

	result := competitors.EvaluationResult{
		ExecutionTimeMillis: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				result.HasDrift = true
				return result, nil
			}
		}
		return result, fmt.Errorf("script failed with err %v\nOutput: %s", err, output)
	}

	result.HasDrift = false
	return result, nil
}

func getActiveCompetitors(t *testing.T) []competitors.Competitor {
	var activeCompetitors []competitors.Competitor
	competitorsDir := "competitors"

	entries, err := os.ReadDir(competitorsDir)
	if err != nil {
		t.Logf("Could not read competitors directory: %v", err)
		return activeCompetitors
	}

	for _, d := range entries {
		if d.IsDir() {
			scriptPath := filepath.Join(competitorsDir, d.Name(), "run.sh")
			if _, err := os.Stat(scriptPath); err == nil {
				// Get absolute path to the script
				absScriptPath, err := filepath.Abs(scriptPath)
				if err != nil {
					t.Fatalf("Failed to get absolute path for script %s: %v", scriptPath, err)
				}
				activeCompetitors = append(activeCompetitors, &ScriptCompetitor{
					scriptPath: absScriptPath,
					name:       strings.ToTitle(d.Name()), // Capitalize folder name
				})
			}
		}
	}

	return activeCompetitors
}

func TestBenchmarks(t *testing.T) {
	datasetDir := "dataset"

	// Parse all benchmark definitions
	var benchmarkFiles []string
	err := filepath.WalkDir(datasetDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && (strings.HasSuffix(d.Name(), ".yaml") || strings.HasSuffix(d.Name(), ".yml")) {
			benchmarkFiles = append(benchmarkFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Logf("No dataset directory found or failed to walk: %v. Creating empty dir.", err)
		os.MkdirAll(datasetDir, 0755)
		t.Skip("No benchmark datasets available yet.")
		return
	}

	if len(benchmarkFiles) == 0 {
		t.Skip("No benchmark datasets available yet in " + datasetDir)
		return
	}

	activeCompetitors := getActiveCompetitors(t)
	if len(activeCompetitors) == 0 {
		t.Skip("No competitors discovered in competitors/ directory.")
		return
	}

	// Scorecard tracking
	type Scorecard struct {
		TruePositives  int
		FalsePositives int
		TrueNegatives  int
		FalseNegatives int
		TotalTime      int64
	}
	scores := make(map[string]*Scorecard)
	for _, comp := range activeCompetitors {
		scores[comp.Name()] = &Scorecard{}
	}

	for _, filePath := range benchmarkFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Fatalf("Failed to read benchmark case %s: %v", filePath, err)
		}

		var bc BenchmarkCase
		if err := yaml.Unmarshal(data, &bc); err != nil {
			t.Fatalf("Failed to parse benchmark case %s: %v", filePath, err)
		}

		t.Run(bc.Name, func(t *testing.T) {
			t.Logf("Checking out %s at %s", bc.Repository, bc.CommitSHA)

			// 1. Checkout repository & calculate diff
			repoDir, diffContext, _, err := testutil.CheckoutAndDiff(bc.Repository, bc.CommitSHA, ".benchmark_repos")
			if err != nil {
				t.Fatalf("Failed to checkout repository: %v", err)
			}

			// Write diff to a temporary file
			tmpDiffFile, err := os.CreateTemp("", "drift_benchmark_diff_*.patch")
			if err != nil {
				t.Fatalf("Failed to create temporary diff file: %v", err)
			}
			defer os.Remove(tmpDiffFile.Name())

			if _, err := tmpDiffFile.WriteString(diffContext); err != nil {
				t.Fatalf("Failed to write to temporary diff file: %v", err)
			}
			tmpDiffFile.Close()

			// 3. Evaluate against all competitors
			for _, comp := range activeCompetitors {
				result, err := comp.Evaluate(repoDir, tmpDiffFile.Name())
				if err != nil {
					t.Errorf("[%s] Failed to evaluate: %v", comp.Name(), err)
					continue
				}

				score := scores[comp.Name()]
				score.TotalTime += result.ExecutionTimeMillis

				if result.HasDrift && bc.Expected.HasDrift {
					score.TruePositives++
				} else if result.HasDrift && !bc.Expected.HasDrift {
					score.FalsePositives++
				} else if !result.HasDrift && bc.Expected.HasDrift {
					score.FalseNegatives++
				} else {
					score.TrueNegatives++
				}
			}
		})
	}

	// Print Unified Scorecard
	t.Cleanup(func() {
		fmt.Printf("\n===================================\n")
		fmt.Printf("🏆 COMPETITIVE BENCHMARK SCORECARD 🏆\n")
		fmt.Printf("===================================\n\n")

		fmt.Printf("%-25s | %-10s | %-10s | %-10s | %-12s\n", "Competitor", "Precision", "Recall", "F1 Score", "Total Time")
		fmt.Printf("--------------------------|------------|------------|------------|-------------\n")

		for _, comp := range activeCompetitors {
			s := scores[comp.Name()]

			precision := 0.0
			if s.TruePositives+s.FalsePositives > 0 {
				precision = float64(s.TruePositives) / float64(s.TruePositives+s.FalsePositives)
			}
			recall := 0.0
			if s.TruePositives+s.FalseNegatives > 0 {
				recall = float64(s.TruePositives) / float64(s.TruePositives+s.FalseNegatives)
			}
			f1 := 0.0
			if precision+recall > 0 {
				f1 = 2 * (precision * recall) / (precision + recall)
			}

			fmt.Printf("%-25s | %.4f     | %.4f     | %.4f     | %d ms\n",
				comp.Name(), precision, recall, f1, s.TotalTime)
		}
		fmt.Printf("\n===================================\n")
	})
}
