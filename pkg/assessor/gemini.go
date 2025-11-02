package assessor

import (
	"context"
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

	model := client.GenerativeModel("gemini-2.5-flash") // Or another suitable model

	return &GeminiAssessor{client: model}, nil
}

// Assess uses the Gemini API to assess drift between code and documentation.
func (a *GeminiAssessor) Assess(docContent string, codeContent string) (*AssessmentResult, error) {
	ctx := context.Background()

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
Please answer with "yes" or "no", followed by a brief explanation.
For example:
"yes, the documentation accurately reflects the code."
"no, the documentation is missing the 'is_active' parameter in the updateUser function."
`, docContent, codeContent)

	resp, err := a.client.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// A very basic way to parse the response.
	// This will need to be made more robust.
	if len(resp.Candidates) > 0 {
		content := resp.Candidates[0].Content
		if len(content.Parts) > 0 {
			response_text := fmt.Sprintf("%v", content.Parts[0])
			if len(response_text) > 2 && response_text[:2] == "no" {
				return &AssessmentResult{
					IsInSync: false,
					Reason:   response_text,
				}, nil
			}
			return &AssessmentResult{
				IsInSync: true,
				Reason:   response_text,
			}, nil
		}
	}

	return &AssessmentResult{
		IsInSync: false,
		Reason:   "Failed to parse response from Gemini API.",
	}, nil
}
