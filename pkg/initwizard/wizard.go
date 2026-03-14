package initwizard

import (
	"fmt"
	"os"

	"github.com/driftee-ai/drift/pkg/config"
	"gopkg.in/yaml.v3"
)

func RunWizard(dir string, fastMode bool) error {
	// Step 1: Run Full-Screen App encompassing Provider Selection, Loading, and Review
	finalRules, usage, provider, err := runReviewApp(dir, fastMode)
	if err != nil {
		return err
	}

	if len(finalRules) == 0 || provider == "" {
		fmt.Println("Initialization cancelled. No configuration generated.")
		return nil
	}

	// Step 2: Save Config
	cfg := config.Config{
		Version:  1,
		Provider: provider,
		Rules:    finalRules,
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	commentedData := []byte("# .drift.yaml\n# This file defines the rules for checking drift between your code and documentation.\n\n" + string(data))
	if err := os.WriteFile(".drift.yaml", commentedData, 0644); err != nil {
		return fmt.Errorf("failed to write .drift.yaml: %w", err)
	}

	fmt.Println("\n✨ Success! Generated .drift.yaml configuration.")
	if usage.TotalTokens > 0 {
		fmt.Printf("📊 Built %d rules using %d LLM tokens.\n", len(finalRules), usage.TotalTokens)
	} else {
		fmt.Printf("📊 Built %d rules.\n", len(finalRules))
	}
	fmt.Println("🚀 You can now run 'drift check' to ensure your documentation is up to date.")
	return nil
}
