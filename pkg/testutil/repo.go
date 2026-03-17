package testutil

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CheckoutCommit clones (or fetches) a repository, checks out a specific commit,
// and returns the absolute path to the checked-out repository.
func CheckoutCommit(repoURL, commitSHA, cacheDir string) (repoDir string, err error) {
	if repoURL == "" || commitSHA == "" {
		return "", fmt.Errorf("repository URL and commit SHA are required")
	}

	repoName := filepath.Base(repoURL)
	repoName = strings.TrimSuffix(repoName, ".git")

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	evalDir := filepath.Join(cacheDir, repoName)

	// 1. Clone or Fetch the repository
	if _, err := os.Stat(evalDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repoURL, evalDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to clone repo %s: %v\nOutput: %s", repoURL, err, output)
		}
	} else {
		cmd := exec.Command("git", "fetch", "origin")
		cmd.Dir = evalDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", fmt.Errorf("failed to fetch repo %s: %v\nOutput: %s", repoURL, err, output)
		}
	}

	// 2. Checkout the specific commit
	checkoutCmd := exec.Command("git", "checkout", commitSHA)
	checkoutCmd.Dir = evalDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("failed to checkout %s: %v\nOutput: %s", commitSHA, err, output)
	}

	// Get absolute path
	absEvalDir, err := filepath.Abs(evalDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for %s: %w", evalDir, err)
	}

	return absEvalDir, nil
}

// CheckoutAndDiff clones (or fetches) a repository, checks out a specific commit,
// and returns the unified diff and the list of changed files for that commit relative
// to its parent. Returns the absolute path to the checked-out repository.
func CheckoutAndDiff(repoURL, commitSHA, cacheDir string) (repoDir string, diffContext string, changedFiles []string, err error) {
	if repoURL == "" || commitSHA == "" {
		return "", "", nil, fmt.Errorf("repository URL and commit SHA are required")
	}

	repoName := filepath.Base(repoURL)
	repoName = strings.TrimSuffix(repoName, ".git")

	// Ensure cache directory exists
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", "", nil, fmt.Errorf("failed to create cache directory %s: %w", cacheDir, err)
	}

	evalDir := filepath.Join(cacheDir, repoName)

	// 1. Clone or Fetch the repository
	if _, err := os.Stat(evalDir); os.IsNotExist(err) {
		cmd := exec.Command("git", "clone", repoURL, evalDir)
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", "", nil, fmt.Errorf("failed to clone repo %s: %v\nOutput: %s", repoURL, err, output)
		}
	} else {
		cmd := exec.Command("git", "fetch", "origin")
		cmd.Dir = evalDir
		if output, err := cmd.CombinedOutput(); err != nil {
			return "", "", nil, fmt.Errorf("failed to fetch repo %s: %v\nOutput: %s", repoURL, err, output)
		}
	}

	// 2. Checkout the specific commit
	checkoutCmd := exec.Command("git", "checkout", commitSHA)
	checkoutCmd.Dir = evalDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return "", "", nil, fmt.Errorf("failed to checkout %s: %v\nOutput: %s", commitSHA, err, output)
	}

	// 3. Calculate the PR Diff and identify changed files
	diffCmd := exec.Command("git", "diff", "HEAD~1..HEAD")
	diffCmd.Dir = evalDir
	diffBytes, err := diffCmd.Output()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to calculate diff: %w", err)
	}
	diffContext = string(diffBytes)

	nameOnlyCmd := exec.Command("git", "diff", "--name-only", "HEAD~1..HEAD")
	nameOnlyCmd.Dir = evalDir
	nameOnlyBytes, err := nameOnlyCmd.Output()
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to calculate changed files: %w", err)
	}

	changedFilesRaw := strings.Split(strings.TrimSpace(string(nameOnlyBytes)), "\n")
	for _, f := range changedFilesRaw {
		f = strings.TrimSpace(f)
		if f != "" {
			changedFiles = append(changedFiles, f)
		}
	}

	// Get absolute path
	absEvalDir, err := filepath.Abs(evalDir)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to get absolute path for %s: %w", evalDir, err)
	}

	return absEvalDir, diffContext, changedFiles, nil
}
