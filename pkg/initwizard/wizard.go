package initwizard

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
	"gopkg.in/yaml.v3"
)

func RunWizard(fastMode bool) error {
	var provider string

	// Step 1: Provider Selection
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an LLM Provider for Auto-Discovery").
				Description("This will analyze your codebase and propose documentation mapping rules.").
				Options(
					huh.NewOption("Google Gemini (Default)", "gemini"),
					huh.NewOption("OpenAI", "openai"),
					huh.NewOption("Anthropic", "anthropic"),
				).
				Value(&provider),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	client, err := llm.New(provider)
	if err != nil {
		return fmt.Errorf("failed to initialize LLM client: %w", err)
	}

	mapper := NewMapper(client)
	var files map[string]string
	var rules []config.Rule
	var scanErr, mapErr error

	// Step 2 & 3: Scan and Map
	action := func() {
		files, scanErr = ScanProject(fastMode)
		if scanErr != nil {
			return
		}
		rules, mapErr = mapper.MapFiles(context.Background(), files, fastMode)
	}

	title := "Analyzing repository and discovering rules..."
	if fastMode {
		title = "Analyzing repository paths (FAST mode) and discovering rules..."
	}

	err = spinner.New().
		Title(title).
		Action(action).
		Run()

	if err != nil {
		return fmt.Errorf("wizard failed: %w", err)
	}
	if scanErr != nil {
		return fmt.Errorf("failed scanning project: %w", scanErr)
	}
	if mapErr != nil {
		return fmt.Errorf("failed generating mappings from LLM: %w", mapErr)
	}

	if len(rules) == 0 {
		fmt.Println("No rules could be automatically discovered. Bootstrapping with example rule.")
		rules = []config.Rule{
			{
				Name: "Example API Documentation",
				Code: []string{"src/api/**/*.go"},
				Docs: []string{"docs/api/**/*.md"},
			},
		}
	}

	// Step 4: Full-Screen Interactive Rule Review
	finalRules, err := runReviewApp(rules)
	if err != nil {
		return err
	}

	if len(finalRules) == 0 {
		fmt.Println("Initialization cancelled or no rules selected. No configuration generated.")
		return nil
	}

	// Step 5: Save Config
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
	fmt.Println("You can now run 'drift check' to ensure your documentation is up to date.")
	return nil
}
