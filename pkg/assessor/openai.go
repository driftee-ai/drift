package assessor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

// OpenAIAssessor is a doc assessor that uses the OpenAI API.
type OpenAIAssessor struct {
	client *openai.Client
}

// NewOpenAIAssessor creates a new OpenAIAssessor.
func NewOpenAIAssessor() (*OpenAIAssessor, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	client := openai.NewClient(apiKey)
	return &OpenAIAssessor{client: client}, nil
}

// Assess assesses the documentation against the code using the OpenAI API.
func (a *OpenAIAssessor) Assess(docContent string, codeContents map[string]string) (*AssessmentResult, error) {
	// Create the prompt
	prompt := "The user wants to check if the documentation is in sync with the code. Please analyze the following documentation and code and determine if they are in sync. The output should be a JSON object with the following structure: {\"is_in_sync\": boolean, \"reason\": string}. The reason should be a short explanation of why the documentation is not in sync with the code. If they are in sync, the reason should be an empty string."
	prompt += "\n\n--- Documentation ---\n" + docContent
	for path, content := range codeContents {
		prompt += "\n\n--- Code file: " + path + " ---\n" + content
	}

	// Create the request
	req := openai.ChatCompletionRequest{
		Model: openai.GPT3Dot5Turbo,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	// Make the API call
	resp, err := a.client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %w", err)
	}

	// Parse the response
	var result AssessmentResult
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal assessment result: %w", err)
	}

	return &result, nil
}
