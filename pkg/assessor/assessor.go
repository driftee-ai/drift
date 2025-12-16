package assessor

import (
	"context"
	"fmt"

	"github.com/driftee-ai/drift/pkg/llm"
)

// AssessmentResult holds the result of a drift assessment.
type AssessmentResult struct {
	IsInSync bool   `json:"is_in_sync"`
	Reason   string `json:"reason"`
}

// DocAssessor is the interface for assessing drift between code and documentation.
type DocAssessor interface {
	Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error)
}

// DriftAssessor implements DocAssessor using an LLM backend.
type DriftAssessor struct {
	generator llm.Generator
}

// NewDriftAssessor creates a new DriftAssessor with the provided generator.
func NewDriftAssessor(generator llm.Generator) *DriftAssessor {
	return &DriftAssessor{generator: generator}
}

// Assess assesses drift between code and documentation.
func (a *DriftAssessor) Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error) {
	codeStr := ""
	for path, content := range codeContents {
		codeStr += fmt.Sprintf("File: %s\n---\n%s\n---\n", path, content)
	}

	prompt := fmt.Sprintf(`
You are a senior software engineer reviewing documentation for a codebase.
Your task is to determine if the documentation is in sync with the code.

Here is the documentation:
---
%s
---

And here is the code:
---
%s
---

Is the documentation in sync with the code?
Please provide your answer in JSON format, with a boolean "is_in_sync" field and a "reason" field.
The reason should be a short explanation of why the documentation is not in sync with the code. If they are in sync, the reason should be an empty string.
`, docContent, codeStr)

	schema := &llm.Schema{
		Type: llm.TypeObject,
		Properties: map[string]*llm.Schema{
			"is_in_sync": {Type: llm.TypeBoolean},
			"reason":     {Type: llm.TypeString},
		},
		Required: []string{"is_in_sync", "reason"},
	}

	var result AssessmentResult
	if err := a.generator.GenerateJSON(context.Background(), prompt, schema, &result); err != nil {
		return nil, fmt.Errorf("assessment failed: %w", err)
	}

	return &result, nil
}
