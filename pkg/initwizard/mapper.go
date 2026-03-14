package initwizard

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/driftee-ai/drift/pkg/config"
	"github.com/driftee-ai/drift/pkg/llm"
)

type MappingResponse struct {
	Rules []config.Rule `json:"rules"`
}

type Mapper struct {
	client llm.Client
}

func NewMapper(client llm.Client) *Mapper {
	return &Mapper{client: client}
}

// MapFiles takes a map of file paths to their contents (or just empty strings if fast mode)
// and returns suggested rules.
func (m *Mapper) MapFiles(ctx context.Context, files map[string]string, fastMode bool) ([]config.Rule, error) {
	var promptBuilder strings.Builder

	promptBuilder.WriteString("You are a configuration builder for a tool called 'drift' that checks if code is in sync with documentation.\n")
	promptBuilder.WriteString("Your task is to group the provided files into logical 'rules'.\n")
	promptBuilder.WriteString("Each rule should map specific code paths to specific documentation paths that document them.\n")
	promptBuilder.WriteString("Use glob patterns (e.g., 'src/api/**/*.go') to describe the paths.\n")
	promptBuilder.WriteString("Give each rule a descriptive name.\n")

	if fastMode {
		promptBuilder.WriteString("\nHere are the files in the repository:\n")
		for path := range files {
			promptBuilder.WriteString("- ")
			promptBuilder.WriteString(path)
			promptBuilder.WriteString("\n")
		}
	} else {
		promptBuilder.WriteString("\nHere are the files and their contents:\n")
		for path, content := range files {
			promptBuilder.WriteString("\n--- File: ")
			promptBuilder.WriteString(path)
			promptBuilder.WriteString(" ---\n")
			// Truncating very large files could be done here, but assuming current context windows are large enough.
			promptBuilder.WriteString(content)
			promptBuilder.WriteString("\n--- End of file ---\n")
		}
	}

	responseJSON, _, err := m.client.GenerateJSON(ctx, promptBuilder.String(), MappingResponse{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate mappings: %w", err)
	}

	var response MappingResponse
	if err := json.Unmarshal([]byte(responseJSON), &response); err != nil {
		return nil, fmt.Errorf("failed to parse mapping response: %w", err)
	}

	return response.Rules, nil
}
