package initwizard

import (
	"context"
	"fmt"
	"os"
	"strings"

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

	// Step 4: Interactive Rule Review
	finalRules, err := reviewRules(rules)
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

func reviewRules(rules []config.Rule) ([]config.Rule, error) {
	var finalRules []config.Rule
	fmt.Println("\nReviewing Proposed Rules:")
	fmt.Println("The AI has analyzed your repo and suggested the following rules.")

	for _, rule := range rules {
		var action string

		codeGlob := strings.Join(rule.Code, ", ")
		docGlob := strings.Join(rule.Docs, ", ")

		if codeGlob == "" {
			codeGlob = "**/*.go"
		}
		if docGlob == "" {
			docGlob = "**/*.md"
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(fmt.Sprintf("Discovered Rule: %s", rule.Name)).
					Description(fmt.Sprintf("Code: %s\nDocs: %s", codeGlob, docGlob)).
					Options(
						huh.NewOption("✅ Accept as is", "accept"),
						huh.NewOption("✏️  Edit rule", "edit"),
						huh.NewOption("❌ Discard", "discard"),
					).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil {
			return nil, err
		}

		switch action {
		case "accept":
			finalRules = append(finalRules, rule)
		case "discard":
			continue
		case "edit":
			editForm := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Rule Name").
						Value(&rule.Name),
					huh.NewInput().
						Title("Code File Glob (comma separated)").
						Value(&codeGlob),
					huh.NewInput().
						Title("Documentation File Glob (comma separated)").
						Value(&docGlob),
				),
			)
			if err := editForm.Run(); err != nil {
				return nil, err
			}

			// Simple split and trim
			rule.Code = []string{}
			for _, p := range strings.Split(codeGlob, ",") {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					rule.Code = append(rule.Code, trimmed)
				}
			}

			rule.Docs = []string{}
			for _, p := range strings.Split(docGlob, ",") {
				if trimmed := strings.TrimSpace(p); trimmed != "" {
					rule.Docs = append(rule.Docs, trimmed)
				}
			}
			finalRules = append(finalRules, rule)
		}
	}

	return finalRules, nil
}
