package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  int    `yaml:"version"`
	Provider string `yaml:"provider"`
	Rules    []Rule `yaml:"rules"`
}

type Rule struct {
	Name string   `yaml:"name"`
	Code []string `yaml:"code"`
	Docs []string `yaml:"docs"`
}

// Load finds and unmarshals a .drift.yaml file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// CreateScaffold creates a blank, commented .drift.yaml for drift init
func CreateScaffold(path string) error {
	exampleConfig := Config{
		Version:  1,
		Provider: "gemini",
		Rules: []Rule{
			{
				Name: "Example API Documentation",
				Code: []string{"src/api/**/*.go"},
				Docs: []string{"docs/api/**/*.md"},
			},
		},
	}

	data, err := yaml.Marshal(exampleConfig)
	if err != nil {
		return err
	}

	// Add comments to the YAML
	commentedData := []byte("# .drift.yaml\n# This file defines the rules for checking drift between your code and documentation.\n\n" + string(data))

	return os.WriteFile(path, commentedData, 0644)
}
