package llmanalyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins/base"
)

// LLMAnalyzer plugin analyzes text using llama.cpp server with dynamic field extraction
type LLMAnalyzer struct {
	*base.BasePlugin
	clientOnce sync.Once
	client     *http.Client
}

// Config holds the configuration for the LLM analyzer
type Config struct {
	APIEndpoint     string   `json:"api_endpoint"`
	FieldsToExtract []string `json:"fields_to_extract"`
	PromptTemplate  string   `json:"prompt_template"`
	Temperature     float64  `json:"temperature"`
	MaxTokens       int      `json:"max_tokens"`
}

// LlamaRequest represents a request to llama.cpp server
type LlamaRequest struct {
	Prompt      string   `json:"prompt"`
	Temperature float64  `json:"temperature"`
	NPredict    int      `json:"n_predict"`
	Stop        []string `json:"stop"`
}

// LlamaResponse represents a response from llama.cpp server
type LlamaResponse struct {
	Content string `json:"content"`
}

// New creates a new LLMAnalyzer plugin
func New() *LLMAnalyzer {
	manifest := types.PluginManifest{
		ID:          "llm-analyzer",
		Name:        "LLM Analyzer",
		Version:     "1.0.0",
		Description: "Analyzes text using llama.cpp server with dynamic field extraction",
		Optional:    false,

		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"transcript": map[string]interface{}{
					"type":        "string",
					"description": "Text to analyze",
				},
			},
			"required": []interface{}{"transcript"},
		},

		OutputSchema: map[string]interface{}{
			"type":        "object",
			"description": "Dynamic analysis results based on fields_to_extract config",
		},

		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"api_endpoint": map[string]interface{}{
					"type":        "string",
					"description": "llama.cpp server endpoint",
				},
				"fields_to_extract": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "List of fields to extract (e.g., ['summary', 'entities', 'locations'])",
				},
				"prompt_template": map[string]interface{}{
					"type":        "string",
					"description": "Go template for prompt. Use {{.Transcript}} and {{.Fields}}",
				},
				"temperature": map[string]interface{}{
					"type":        "number",
					"description": "LLM temperature (0.0-1.0)",
				},
				"max_tokens": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum tokens to generate",
				},
			},
			"required": []interface{}{"fields_to_extract"},
		},

		Timeout:     60, // 1 minute
		MaxMemoryMB: 50,
	}

	return &LLMAnalyzer{
		BasePlugin: base.NewBasePlugin(manifest),
	}
}

func (p *LLMAnalyzer) getClient() *http.Client {
	p.clientOnce.Do(func() {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.MaxIdleConns = 100
		transport.IdleConnTimeout = 90 * time.Second
		p.client = &http.Client{
			Transport: transport,
		}
	})

	return p.client
}

// Execute analyzes the transcript using LLM
func (p *LLMAnalyzer) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	// Validate execution context
	if err := p.Validate(ctx); err != nil {
		return nil, err
	}

	// Parse config with defaults
	var config Config
	if ctx.Config != nil {
		configBytes, _ := json.Marshal(ctx.Config)
		json.Unmarshal(configBytes, &config)
	}

	// Set defaults
	if config.APIEndpoint == "" {
		config.APIEndpoint = "http://localhost:8080/completion"
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 500
	}

	// Validate fields_to_extract
	if len(config.FieldsToExtract) == 0 {
		return nil, errors.NewValidationError(
			"fields_to_extract is required and cannot be empty",
			"$.fields_to_extract",
			"Specify at least one field to extract (e.g., ['summary', 'entities'])",
		)
	}

	// Extract transcript from state
	transcript, ok := ctx.State["transcript"].(string)
	if !ok {
		return nil, errors.NewValidationError(
			"transcript must be a string",
			"$.transcript",
			"Ensure transcript field is a string",
		)
	}

	if len(transcript) > 10000000 { // 10MB limit for safety
		return nil, fmt.Errorf("transcript too large (max 10MB)")
	}

	if strings.TrimSpace(transcript) == "" {
		return nil, errors.NewValidationError(
			"transcript cannot be empty",
			"$.transcript",
			"Provide a non-empty transcript to analyze",
		)
	}

	// Build prompt
	prompt, err := p.buildPrompt(transcript, config)
	if err != nil {
		return nil, fmt.Errorf("failed to build prompt: %w", err)
	}

	// Call LLM API
	analysis, err := p.callLLM(ctx, prompt, config)
	if err != nil {
		return nil, err
	}

	// Validate that requested fields exist
	missingFields := []string{}
	for _, field := range config.FieldsToExtract {
		if _, exists := analysis[field]; !exists {
			missingFields = append(missingFields, field)
		}
	}

	if len(missingFields) > 0 {
		ctx.Logger.Warn("Some fields missing from LLM response", map[string]interface{}{
			"missing_fields": missingFields,
		})
	}

	// Preserve state and add analysis fields
	output := make(map[string]interface{}, len(ctx.State)+len(analysis))
	for k, v := range ctx.State {
		output[k] = v
	}

	// Add analysis fields to output
	for k, v := range analysis {
		output[k] = v
	}

	ctx.Logger.Info("LLM analysis completed", map[string]interface{}{
		"execution_id":     ctx.ExecutionID,
		"fields_extracted": len(analysis),
		"fields_requested": len(config.FieldsToExtract),
	})

	return &types.PluginResult{
		Success: true,
		Data:    output,
	}, nil
}

// buildPrompt creates a prompt from template or default
func (p *LLMAnalyzer) buildPrompt(transcript string, config Config) (string, error) {
	if config.PromptTemplate != "" {
		// Use custom template
		tmpl, err := template.New("prompt").Parse(config.PromptTemplate)
		if err != nil {
			return "", fmt.Errorf("invalid prompt template: %w", err)
		}

		var buf bytes.Buffer
		data := map[string]interface{}{
			"Transcript": transcript,
			"Fields":     config.FieldsToExtract,
		}

		if err := tmpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("failed to execute template: %w", err)
		}

		return buf.String(), nil
	}

	// Use default dynamic prompt
	return buildDefaultPrompt(transcript, config.FieldsToExtract), nil
}

// callLLM calls llama.cpp server and parses response
func (p *LLMAnalyzer) callLLM(ctx *types.ExecutionContext, prompt string, config Config) (map[string]interface{}, error) {
	client := p.getClient()

	// Build request
	llamaReq := LlamaRequest{
		Prompt:      prompt,
		Temperature: config.Temperature,
		NPredict:    config.MaxTokens,
		Stop:        []string{"</analysis>", "\n\n\n"},
	}

	reqBody, err := json.Marshal(llamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	reqCtx, cancel := context.WithTimeout(ctx.Context, 50*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST", config.APIEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	ctx.Logger.Info("Calling LLM API", map[string]interface{}{
		"endpoint":      config.APIEndpoint,
		"fields":        config.FieldsToExtract,
		"prompt_length": len(prompt),
	})

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LLM API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var llamaResp LlamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&llamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Extract JSON from response
	jsonContent := extractJSON(llamaResp.Content)

	var analysis map[string]interface{}
	if err := json.Unmarshal([]byte(jsonContent), &analysis); err != nil {
		ctx.Logger.Warn("Failed to parse LLM output as JSON, returning raw text", map[string]interface{}{
			"error": err.Error(),
		})

		// Fallback: create analysis object with raw text
		analysis = map[string]interface{}{
			"raw_analysis": llamaResp.Content,
		}
	}

	return analysis, nil
}

// Helper functions

// buildDefaultPrompt creates a dynamic prompt based on fields to extract
func buildDefaultPrompt(transcript string, fields []string) string {
	var prompt strings.Builder

	prompt.WriteString("Analyze the following transcript and extract the requested information.\n\n")
	prompt.WriteString("Transcript:\n")
	prompt.WriteString(transcript)
	prompt.WriteString("\n\n")
	prompt.WriteString("Extract the following fields and return as JSON:\n")

	for _, field := range fields {
		prompt.WriteString(fmt.Sprintf("- %s\n", field))
	}

	prompt.WriteString("\nReturn ONLY valid JSON with these exact field names. Example:\n")
	prompt.WriteString("{\n")
	for i, field := range fields {
		if i > 0 {
			prompt.WriteString(",\n")
		}
		prompt.WriteString(fmt.Sprintf("  \"%s\": \"...\"", field))
	}
	prompt.WriteString("\n}\n\nJSON output:")

	return prompt.String()
}

// extractJSON attempts to extract JSON from text that might contain markdown or other wrapping
func extractJSON(text string) string {
	// Try to find JSON between ```json and ``` markers
	if start := strings.Index(text, "```json"); start != -1 {
		start += 7
		if end := strings.Index(text[start:], "```"); end != -1 {
			return strings.TrimSpace(text[start : start+end])
		}
	}

	// Try to find JSON between ``` and ``` (without json specifier)
	if start := strings.Index(text, "```"); start != -1 {
		start += 3
		if end := strings.Index(text[start:], "```"); end != -1 {
			content := strings.TrimSpace(text[start : start+end])
			// Check if it looks like JSON
			if strings.HasPrefix(content, "{") {
				return content
			}
		}
	}

	// Try to find JSON between { and }
	if start := strings.Index(text, "{"); start != -1 {
		if end := strings.LastIndex(text, "}"); end != -1 && end > start {
			return text[start : end+1]
		}
	}

	// Return original text if no JSON found
	return text
}
