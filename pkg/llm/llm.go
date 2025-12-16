package llm

import "context"

// Generator is the interface for LLM providers.
type Generator interface {
	// Generate generates text content based on the prompt.
	Generate(ctx context.Context, prompt string) (string, error)
	// GenerateJSON generates a JSON response based on the prompt and unmarshals it into the provided result interface.
	// schema: optional schema definition to guide the LLM.
	// result: pointer to the struct to unmarshal into.
	GenerateJSON(ctx context.Context, prompt string, schema *Schema, result any) error
}

// SchemaType represents the type of a schema element.
type SchemaType string

const (
	TypeObject  SchemaType = "OBJECT"
	TypeArray   SchemaType = "ARRAY"
	TypeString  SchemaType = "STRING"
	TypeBoolean SchemaType = "BOOLEAN"
	TypeInteger SchemaType = "INTEGER"
	TypeNumber  SchemaType = "NUMBER"
)

// Schema defines the structure of the expected JSON response.
// It is a provider-agnostic representation that can be converted to provider-specific schemas.
type Schema struct {
	Type        SchemaType
	Properties  map[string]*Schema
	Items       *Schema  // For arrays
	Required    []string // List of required property names
	Description string   // Description of the field (optional)
}
