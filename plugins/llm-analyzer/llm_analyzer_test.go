package llmanalyzer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/pkg/types"
)

func TestNew(t *testing.T) {
	plugin := New()
	assert.NotNil(t, plugin)
	assert.Equal(t, "llm-analyzer", plugin.Manifest().ID)
	assert.Equal(t, "1.0.0", plugin.Manifest().Version)
	assert.Equal(t, 60, plugin.Manifest().Timeout)
}

func TestLLMAnalyzer_ManifestSchemas(t *testing.T) {
	plugin := New()
	manifest := plugin.Manifest()

	// Verify input schema
	assert.NotNil(t, manifest.InputSchema)
	inputProps := manifest.InputSchema["properties"].(map[string]interface{})
	assert.Contains(t, inputProps, "transcript")

	// Verify output schema (dynamic)
	assert.NotNil(t, manifest.OutputSchema)

	// Verify config schema
	assert.NotNil(t, manifest.ConfigSchema)
	configProps := manifest.ConfigSchema["properties"].(map[string]interface{})
	assert.Contains(t, configProps, "api_endpoint")
	assert.Contains(t, configProps, "fields_to_extract")
	assert.Contains(t, configProps, "prompt_template")
	assert.Contains(t, configProps, "temperature")
	assert.Contains(t, configProps, "max_tokens")
}

func TestLLMAnalyzer_Execute_NoFieldsToExtract(t *testing.T) {
	plugin := New()

	config := map[string]interface{}{}
	input := map[string]interface{}{
		"transcript": "Test transcript",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "fields_to_extract")
}

func TestLLMAnalyzer_Execute_NoTranscript(t *testing.T) {
	plugin := New()

	config := map[string]interface{}{
		"fields_to_extract": []string{"summary"},
	}
	input := map[string]interface{}{}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestLLMAnalyzer_Execute_EmptyTranscript(t *testing.T) {
	plugin := New()

	config := map[string]interface{}{
		"fields_to_extract": []string{"summary"},
	}
	input := map[string]interface{}{
		"transcript": "   ", // Empty/whitespace
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "transcript cannot be empty")
}

func TestLLMAnalyzer_Execute_Success(t *testing.T) {
	// Create mock llama.cpp server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse request
		var llamaReq LlamaRequest
		err := json.NewDecoder(r.Body).Decode(&llamaReq)
		assert.NoError(t, err)

		// Verify prompt contains expected fields
		assert.Contains(t, llamaReq.Prompt, "summary")
		assert.Contains(t, llamaReq.Prompt, "entities")
		assert.Contains(t, llamaReq.Prompt, "Test transcript")

		// Mock response with requested fields
		response := LlamaResponse{
			Content: `{
				"summary": "This is a test summary",
				"entities": ["Entity1", "Entity2"]
			}`,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	plugin := New()

	config := map[string]interface{}{
		"api_endpoint":      server.URL,
		"fields_to_extract": []string{"summary", "entities"},
		"temperature":       0.5,
		// max_tokens omitted - will use default (500)
	}

	input := map[string]interface{}{
		"transcript": "Test transcript about coding",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	// Verify extracted fields
	assert.Contains(t, result.Data, "summary")
	assert.Contains(t, result.Data, "entities")
	assert.Equal(t, "This is a test summary", result.Data["summary"])

	// Verify state preserved
	assert.Contains(t, result.Data, "transcript")

	// Cleanup
	err = plugin.Cleanup(context.Background())
	assert.NoError(t, err)
}

func TestLLMAnalyzer_Execute_WithMarkdownJSON(t *testing.T) {
	// Create mock server that returns JSON wrapped in markdown
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LlamaResponse{
			Content: "```json\n{\"summary\": \"Markdown wrapped summary\"}\n```",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	plugin := New()

	config := map[string]interface{}{
		"api_endpoint":      server.URL,
		"fields_to_extract": []string{"summary"},
	}

	input := map[string]interface{}{
		"transcript": "Test",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	require.NoError(t, err)
	assert.Equal(t, "Markdown wrapped summary", result.Data["summary"])
}

func TestBuildDefaultPrompt(t *testing.T) {
	transcript := "This is a test transcript."
	fields := []string{"summary", "entities", "sentiment"}

	prompt := buildDefaultPrompt(transcript, fields)

	// Verify prompt contains transcript
	assert.Contains(t, prompt, transcript)

	// Verify prompt contains all fields
	for _, field := range fields {
		assert.Contains(t, prompt, field)
	}

	// Verify prompt asks for JSON
	assert.Contains(t, prompt, "JSON")
}

func TestExtractJSON_MarkdownWrapped(t *testing.T) {
	text := "```json\n{\"key\": \"value\"}\n```"
	result := extractJSON(text)
	assert.Equal(t, `{"key": "value"}`, result)
}

func TestExtractJSON_PlainMarkdown(t *testing.T) {
	text := "```\n{\"key\": \"value\"}\n```"
	result := extractJSON(text)
	assert.Equal(t, `{"key": "value"}`, result)
}

func TestExtractJSON_PlainJSON(t *testing.T) {
	text := "{\"key\": \"value\"}"
	result := extractJSON(text)
	assert.Equal(t, `{"key": "value"}`, result)
}

func TestExtractJSON_EmbeddedJSON(t *testing.T) {
	text := "Here is the result: {\"key\": \"value\"} end."
	result := extractJSON(text)
	assert.Equal(t, `{"key": "value"}`, result)
}

func TestExtractJSON_NoJSON(t *testing.T) {
	text := "No JSON here"
	result := extractJSON(text)
	assert.Equal(t, "No JSON here", result)
}

func TestLLMAnalyzer_Execute_APIError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}))
	defer server.Close()

	plugin := New()

	config := map[string]interface{}{
		"api_endpoint":      server.URL,
		"fields_to_extract": []string{"summary"},
	}

	input := map[string]interface{}{
		"transcript": "Test",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "LLM API error")
}

func TestLLMAnalyzer_Execute_InvalidJSON(t *testing.T) {
	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := LlamaResponse{
			Content: "This is not JSON at all",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	plugin := New()

	config := map[string]interface{}{
		"api_endpoint":      server.URL,
		"fields_to_extract": []string{"summary"},
	}

	input := map[string]interface{}{
		"transcript": "Test",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	require.NoError(t, err) // Should not error, fallback to raw_analysis
	assert.Contains(t, result.Data, "raw_analysis")
}

func TestLLMAnalyzer_Execute_CustomPromptTemplate(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var llamaReq LlamaRequest
		json.NewDecoder(r.Body).Decode(&llamaReq)

		// Verify custom template was used
		assert.Contains(t, llamaReq.Prompt, "CUSTOM TEMPLATE")
		assert.Contains(t, llamaReq.Prompt, "Test transcript")

		response := LlamaResponse{
			Content: `{"summary": "test"}`,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	plugin := New()

	config := map[string]interface{}{
		"api_endpoint":      server.URL,
		"fields_to_extract": []string{"summary"},
		"prompt_template":   "CUSTOM TEMPLATE: {{.Transcript}}",
	}

	input := map[string]interface{}{
		"transcript": "Test transcript",
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: uuid.New().String(),
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	require.NoError(t, err)
	assert.NotNil(t, result)
}
