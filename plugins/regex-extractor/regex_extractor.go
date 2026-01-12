package regexextractor

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// Pattern represents a regex extraction pattern
type Pattern struct {
	Name    string `json:"name" yaml:"name"`
	Pattern string `json:"pattern" yaml:"pattern"`
}

// Config holds the plugin configuration
type Config struct {
	Patterns   []Pattern `json:"patterns" yaml:"patterns"`
	InputField string    `json:"input_field" yaml:"input_field"`
}

// Plugin implements regex-based data extraction
type Plugin struct {
	*base.BasePlugin
}

// New creates a new regex-extractor plugin instance
func New() *Plugin {
	manifest := types.PluginManifest{
		ID:          "regex-extractor",
		Name:        "Regex Extractor",
		Version:     "1.0.0",
		Description: "Extracts data from text using regular expression patterns",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		OutputSchema: map[string]interface{}{
			"type": "object",
		},
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"patterns": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"name":    map[string]interface{}{"type": "string"},
							"pattern": map[string]interface{}{"type": "string"},
						},
						"required": []string{"name", "pattern"},
					},
				},
				"input_field": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{"patterns"},
		},
	}

	return &Plugin{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

// Execute performs the regex extraction
func (p *Plugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	// Validate execution context
	if err := p.BasePlugin.Validate(ctx); err != nil {
		return nil, err
	}

	// Parse config
	var config Config
	configBytes, err := json.Marshal(ctx.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := json.Unmarshal(configBytes, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if len(config.Patterns) == 0 {
		return nil, fmt.Errorf("at least one pattern is required")
	}

	// Default input field to "text"
	if config.InputField == "" {
		config.InputField = "text"
	}

	// Get the input text
	textValue, exists := ctx.State[config.InputField]
	if !exists {
		return nil, fmt.Errorf("input field '%s' not found", config.InputField)
	}

	text, ok := textValue.(string)
	if !ok {
		return nil, fmt.Errorf("input field '%s' must be a string, got %T", config.InputField, textValue)
	}

	// Create output map with all original fields
	output := make(map[string]interface{})
	for k, v := range ctx.State {
		output[k] = v
	}

	// Extract patterns
	for _, pattern := range config.Patterns {
		regex, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex pattern '%s': %w", pattern.Name, err)
		}

		match := regex.FindString(text)
		if match != "" {
			output[pattern.Name] = match
		}
	}

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}
