package assessor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// OpenAIAssessor uses the OpenAI API to assess drift.
type OpenAIAssessor struct {
	apiKey string
	client *http.Client
}

// NewOpenAIAssessor creates a new OpenAIAssessor.
// It reads the OpenAI API key from the OPENAI_API_KEY environment variable.
func NewOpenAIAssessor() (*OpenAIAssessor, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	return &OpenAIAssessor{
		apiKey: apiKey,
		client: &http.Client{},
	}, nil
}

// Assess uses the OpenAI API to assess drift between code and documentation.
func (a *OpenAIAssessor) Assess(docContent string, codeContent string) (*AssessmentResult, error) {
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
Your response should be only the JSON object, without any other text or formatting.
`, docContent, codeContent)

	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant that provides responses in JSON format."},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request to OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API returned non-200 status code: %d", resp.StatusCode)
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response from OpenAI API: %w", err)
	}

	if len(openAIResp.Choices) > 0 {
		content := openAIResp.Choices[0].Message.Content
		var result AssessmentResult
		if err := json.Unmarshal([]byte(content), &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response from OpenAI: %w", err)
		}
		return &result, nil
	}

	return &AssessmentResult{
		IsInSync: false,
		Reason:   "Failed to parse response from OpenAI API.",
	}, nil
}
