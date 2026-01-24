package jsontransformer

import (
	"encoding/json"
	"fmt"

	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// Mapping represents a field transformation rule
type Mapping struct {
	From string `json:"from" yaml:"from"`
	To   string `json:"to" yaml:"to"`
}

// Config holds the plugin configuration
type Config struct {
	Mappings []Mapping `json:"mappings" yaml:"mappings"`
}

// Plugin implements field transformation and renaming
type Plugin struct {
	*base.BasePlugin
}

// New creates a new json-transformer plugin instance
func New() *Plugin {
	manifest := types.PluginManifest{
		ID:          "json-transformer",
		Name:        "JSON Transformer",
		Version:     "1.0.0",
		Description: "Transforms and renames JSON fields based on mapping rules",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		OutputSchema: map[string]interface{}{
			"type": "object",
		},
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"mappings": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"from": map[string]interface{}{"type": "string"},
							"to":   map[string]interface{}{"type": "string"},
						},
						"required": []string{"from", "to"},
					},
				},
			},
			"required": []string{"mappings"},
		},
	}

	return &Plugin{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

// Execute performs the field transformation
func (p *Plugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	// Validate execution context
	if err := p.Validate(ctx); err != nil {
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

	if len(config.Mappings) == 0 {
		return nil, fmt.Errorf("at least one mapping is required")
	}

	// Create output map
	output := make(map[string]interface{}, len(ctx.State)+len(config.Mappings))

	// Copy all fields first
	for k, v := range ctx.State {
		output[k] = v
	}

	// Apply mappings
	for _, mapping := range config.Mappings {
		if value, exists := ctx.State[mapping.From]; exists {
			output[mapping.To] = value
			// Remove old field if it's different from the new one
			if mapping.From != mapping.To {
				delete(output, mapping.From)
			}
		}
	}

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}
