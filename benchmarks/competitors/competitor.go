package competitors

// EvaluationResult represents the outcome of a competitor evaluating a single diff.
type EvaluationResult struct {
	HasDrift            bool
	ExecutionTimeMillis int64
}

// Competitor defines the interface that any drift-detection approach must implement
// to be benchmarked against the dataset.
type Competitor interface {
	// Name returns the display name of the competitor (e.g., "Drift Core", "Naive Regex")
	Name() string

	// Evaluate is called for every commit in the benchmark dataset.
	// It receives the absolute path to the repository, and the absolute path to a file
	// containing the raw unified diff of the PR.
	Evaluate(repoDir string, diffFile string) (EvaluationResult, error)
}
