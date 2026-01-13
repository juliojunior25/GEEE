package templaterenderer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// Config holds the plugin configuration
type Config struct {
	TemplateField string `json:"template_field" yaml:"template_field"`
	DataField     string `json:"data_field" yaml:"data_field"`
	OutputField   string `json:"output_field" yaml:"output_field"`
}

// Plugin implements Go template rendering functionality
type Plugin struct {
	*base.BasePlugin
}

// New creates a new template-renderer plugin instance
func New() *Plugin {
	manifest := types.PluginManifest{
		ID:          "template-renderer",
		Name:        "Template Renderer",
		Version:     "1.0.0",
		Description: "Renders Go templates with provided data",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
		OutputSchema: map[string]interface{}{
			"type": "object",
		},
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"template_field": map[string]interface{}{
					"type": "string",
				},
				"data_field": map[string]interface{}{
					"type": "string",
				},
				"output_field": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}

	return &Plugin{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

// Execute performs the template rendering
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

	// Set defaults
	if config.TemplateField == "" {
		config.TemplateField = "template"
	}
	if config.DataField == "" {
		config.DataField = "data"
	}
	if config.OutputField == "" {
		config.OutputField = "result"
	}

	// Get the template string
	templateValue, exists := ctx.State[config.TemplateField]
	if !exists {
		return nil, fmt.Errorf("template field '%s' not found", config.TemplateField)
	}

	templateStr, ok := templateValue.(string)
	if !ok {
		return nil, fmt.Errorf("template field '%s' must be a string, got %T", config.TemplateField, templateValue)
	}

	// Get the data
	data, exists := ctx.State[config.DataField]
	if !exists {
		return nil, fmt.Errorf("data field '%s' not found", config.DataField)
	}

	// Parse template
	tmpl, err := template.New("main").Parse(templateStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Create output map with all original fields
	output := make(map[string]interface{})
	for k, v := range ctx.State {
		output[k] = v
	}
	output[config.OutputField] = buf.String()

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}
