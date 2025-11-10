package assessor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiAssessor uses the Gemini API to assess drift.
type GeminiAssessor struct {
	client *genai.GenerativeModel
}

// NewGeminiAssessor creates a new GeminiAssessor.
// It reads the Gemini API key from the GEMINI_API_KEY environment variable.
func NewGeminiAssessor() (*GeminiAssessor, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-flash")

	return &GeminiAssessor{client: model}, nil
}

// Assess uses the Gemini API to assess drift between code and documentation.
func (a *GeminiAssessor) Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error) {
	ctx := context.Background()

	// Define the response schema
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"is_in_sync": {Type: genai.TypeBoolean},
			"reason":     {Type: genai.TypeString},
		},
		Required: []string{"is_in_sync", "reason"},
	}

	// Set the response mime type and schema
	a.client.ResponseMIMEType = "application/json"
	a.client.ResponseSchema = schema

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
`, docContent, codeStr)

	resp, err := a.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Candidates) > 0 {
		content := resp.Candidates[0].Content
		if len(content.Parts) > 0 {
			// The response should be a JSON string, unmarshal it
			var result AssessmentResult
			if err := json.Unmarshal([]byte(fmt.Sprintf("%s", content.Parts[0])), &result); err != nil {
				return nil, fmt.Errorf("failed to unmarshal response: %w", err)
			}
			return &result, nil
		}
	}

	return &AssessmentResult{
		IsInSync: false,
		Reason:   "Failed to parse response from Gemini API.",
	}, nil
}
