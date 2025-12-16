package llm

import (
	"testing"

	"github.com/google/generative-ai-go/genai"
	"github.com/stretchr/testify/assert"
)

func TestToGenaiSchema(t *testing.T) {
	tests := []struct {
		name     string
		input    *Schema
		expected *genai.Schema
	}{
		{
			name:     "Nil Schema",
			input:    nil,
			expected: nil,
		},
		{
			name: "Simple Object",
			input: &Schema{
				Type:        TypeObject,
				Description: "A simple object",
				Properties: map[string]*Schema{
					"name": {Type: TypeString},
					"age":  {Type: TypeInteger},
				},
				Required: []string{"name"},
			},
			expected: &genai.Schema{
				Type:        genai.TypeObject,
				Description: "A simple object",
				Properties: map[string]*genai.Schema{
					"name": {Type: genai.TypeString},
					"age":  {Type: genai.TypeInteger},
				},
				Required: []string{"name"},
			},
		},
		{
			name: "Array of Strings",
			input: &Schema{
				Type: TypeArray,
				Items: &Schema{
					Type: TypeString,
				},
			},
			expected: &genai.Schema{
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeString,
				},
			},
		},
		{
			name: "Nested Object",
			input: &Schema{
				Type: TypeObject,
				Properties: map[string]*Schema{
					"meta": {
						Type: TypeObject,
						Properties: map[string]*Schema{
							"active": {Type: TypeBoolean},
						},
					},
				},
			},
			expected: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"meta": {
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"active": {Type: genai.TypeBoolean},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toGenaiSchema(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
