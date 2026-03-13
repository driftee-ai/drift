package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/files"
	"github.com/driftee-ai/drift/pkg/llm"
	"github.com/google/generative-ai-go/genai"
)

// AssessmentResult holds the structured response from the LLM.
type AssessmentResult struct {
	IsInSync            bool   `json:"is_in_sync"`
	Reason              string `json:"reason"`
	IsDriftCausedByDiff *bool  `json:"is_drift_caused_by_diff,omitempty"`
}

// RuleResult holds the outcome of evaluating a single rule.
type RuleResult struct {
	Rule                config.Rule
	CodeFilesCount      int
	DocsFilesCount      int
	CodeTotalBytes      int
	DocsTotalBytes      int
	Error               error
	Skipped             bool
	IsInSync            bool
	Reason              string
	IsDriftCausedByDiff *bool
	IgnoredDueToDiff    bool
}

// Checker encapsulates the logic for evaluating rules against an LLM client.
type Checker struct {
	client   llm.Client
	provider string
}

// New creates a new Checker instance.
func New(client llm.Client, provider string) *Checker {
	return &Checker{
		client:   client,
		provider: provider,
	}
}

// EvaluateRules takes a list of triggered rules, reads the associated files,
// queries the LLM, and returns the results.
func (c *Checker) EvaluateRules(ctx context.Context, rules []config.Rule, diffOnly bool, diffContext string) []RuleResult {
	var results []RuleResult

	for _, rule := range rules {
		result := RuleResult{Rule: rule}

		// Find and read code files
		codeFiles, err := files.FindFiles(rule.Code)
		if err != nil {
			result.Error = fmt.Errorf("error finding code files: %w", err)
			results = append(results, result)
			continue
		}
		codeContents, err := files.ReadFiles(codeFiles)
		if err != nil {
			result.Error = fmt.Errorf("error reading code content: %w", err)
			results = append(results, result)
			continue
		}

		codeStr := ""
		totalCodeSize := 0
		for path, content := range codeContents {
			totalCodeSize += len(content)
			codeStr += fmt.Sprintf("\n--- Code file: %s ---\n%s\n", path, content)
		}
		result.CodeFilesCount = len(codeFiles)
		result.CodeTotalBytes = totalCodeSize

		// Find and read docs files
		docFiles, err := files.FindFiles(rule.Docs)
		if err != nil {
			result.Error = fmt.Errorf("error finding doc files: %w", err)
			results = append(results, result)
			continue
		}
		docContent, err := files.ReadAndConcatenate(docFiles)
		if err != nil {
			result.Error = fmt.Errorf("error reading doc content: %w", err)
			results = append(results, result)
			continue
		}
		result.DocsFilesCount = len(docFiles)
		result.DocsTotalBytes = len(docContent)

		if len(codeFiles) == 0 && len(docFiles) == 0 {
			result.Skipped = true
			results = append(results, result)
			continue
		}
		if len(codeFiles) == 0 || len(docFiles) == 0 {
			result.Error = fmt.Errorf("missing code or doc files")
			result.IsInSync = false
			results = append(results, result)
			continue
		}

		// Assess the drift
		prompt := fmt.Sprintf(`You are a senior software engineer reviewing documentation for a codebase.
Your task is to determine if the documentation is in sync with the code.

Here is the documentation:
---
%s
---

And here is the code:
%s
`, docContent, codeStr)

		if diffOnly && diffContext != "" {
			prompt += fmt.Sprintf("\nHere is the git diff of the recent changes:\n---\n%s\n---\nBased on the provided diff, did the changes introduced in this diff cause the documentation drift?", diffContext)
		}

		var schema interface{}
		if c.provider == "gemini" {
			geminiSchema := &genai.Schema{
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

			if diffOnly && diffContext != "" {
				geminiSchema.Properties["is_drift_caused_by_diff"] = &genai.Schema{
					Type:        genai.TypeBoolean,
					Description: "True if the documentation drift was caused by the provided git diff, false if the drift is preexisting or unrelated.",
				}
				geminiSchema.Required = append(geminiSchema.Required, "is_drift_caused_by_diff")
			}
			schema = geminiSchema
		} else {
			schema = AssessmentResult{}
		}

		jsonRes, _, err := c.client.GenerateJSON(ctx, prompt, schema)
		if err != nil {
			result.Error = fmt.Errorf("error assessing drift from LLM: %w", err)
			result.IsInSync = false
			results = append(results, result)
			continue
		}

		// Clean raw string output since some models inject Markdown blocks
		cleanJson := strings.TrimPrefix(strings.TrimSpace(jsonRes), "```json")
		cleanJson = strings.TrimPrefix(cleanJson, "```")
		cleanJson = strings.TrimSuffix(cleanJson, "```")
		cleanJson = strings.TrimSpace(cleanJson)

		var assessment AssessmentResult
		if err := json.Unmarshal([]byte(cleanJson), &assessment); err != nil {
			result.Error = fmt.Errorf("error parsing JSON response: %w\nRaw Response: %s", err, jsonRes)
			result.IsInSync = false
			results = append(results, result)
			continue
		}

		result.IsInSync = assessment.IsInSync
		result.Reason = assessment.Reason
		result.IsDriftCausedByDiff = assessment.IsDriftCausedByDiff

		if !result.IsInSync && diffOnly && diffContext != "" && result.IsDriftCausedByDiff != nil && !*result.IsDriftCausedByDiff {
			result.IgnoredDueToDiff = true
			result.IsInSync = true // Overwrite to true since we are ignoring it
		}

		results = append(results, result)
	}

	return results
}
