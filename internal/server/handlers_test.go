package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yourusername/geee/internal/executor"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/internal/registry"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins"
)

// setupTestHandlers creates a test Handlers instance with registered plugins
func setupTestHandlers(t *testing.T) *Handlers {
	t.Helper()

	pluginRegistry := registry.NewDefaultRegistry()
	if err := plugins.RegisterAll(pluginRegistry); err != nil {
		t.Fatalf("Failed to register plugins: %v", err)
	}

	logger := observability.NewJSONLogger(os.Stdout, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := executor.NewDefaultExecutor(pluginRegistry, logger, metrics)

	return NewHandlers(pluginRegistry, exec, logger)
}

func TestHandleHealth(t *testing.T) {
	handlers := setupTestHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handlers.HandleHealth(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var healthResp HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if healthResp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", healthResp.Status)
	}

	if len(healthResp.Plugins) == 0 {
		t.Error("Expected plugins in health response, got none")
	}
}

func TestHandleReady(t *testing.T) {
	handlers := setupTestHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()

	handlers.HandleReady(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var readyResp ReadyResponse
	if err := json.NewDecoder(resp.Body).Decode(&readyResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !readyResp.Ready {
		t.Errorf("Expected ready=true, got false")
	}

	if readyResp.Since.IsZero() {
		t.Error("Expected since timestamp, got zero value")
	}
}

func TestHandlePlugins(t *testing.T) {
	handlers := setupTestHandlers(t)

	req := httptest.NewRequest(http.MethodGet, "/plugins", nil)
	w := httptest.NewRecorder()

	handlers.HandlePlugins(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var pluginsResp PluginsResponse
	if err := json.NewDecoder(resp.Body).Decode(&pluginsResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if pluginsResp.Count == 0 {
		t.Error("Expected plugins in response, got none")
	}

	if len(pluginsResp.Plugins) != pluginsResp.Count {
		t.Errorf("Expected count to match plugins length, got %d != %d", pluginsResp.Count, len(pluginsResp.Plugins))
	}

	// Verify plugin info structure
	for _, plugin := range pluginsResp.Plugins {
		if plugin.ID == "" {
			t.Error("Expected plugin ID, got empty string")
		}
		if plugin.Name == "" {
			t.Error("Expected plugin name, got empty string")
		}
	}
}

func TestHandleRun_MissingConfig(t *testing.T) {
	handlers := setupTestHandlers(t)

	reqBody := RunRequest{
		Config: "",
		Input:  map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRun(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var runResp RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp.Success {
		t.Error("Expected success=false for missing config")
	}

	if runResp.Error == nil {
		t.Error("Expected error in response")
	}

	if runResp.Error.Code != "MISSING_CONFIG" {
		t.Errorf("Expected error code MISSING_CONFIG, got %s", runResp.Error.Code)
	}
}

func TestHandleRun_InvalidConfig(t *testing.T) {
	handlers := setupTestHandlers(t)

	reqBody := RunRequest{
		Config: "invalid: [yaml",
		Input:  map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRun(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var runResp RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp.Success {
		t.Error("Expected success=false for invalid config")
	}

	if runResp.Error == nil {
		t.Error("Expected error in response")
	}
}

func TestHandleRun_Success(t *testing.T) {
	handlers := setupTestHandlers(t)

	// Create a valid inline YAML config
	configYAML := `
name: test-pipeline
description: Test pipeline
plugins:
  - id: json-transformer
    config:
      mappings:
        - from: input
          to: output
`

	reqBody := RunRequest{
		Config: configYAML,
		Input: map[string]interface{}{
			"input": "test-value",
		},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRun(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var runResp RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !runResp.Success {
		t.Errorf("Expected success=true, got false. Error: %v", runResp.Error)
	}

	if runResp.ExecutionID == "" {
		t.Error("Expected execution ID in response")
	}

	if runResp.DurationMS < 0 {
		t.Errorf("Expected duration >= 0, got %d", runResp.DurationMS)
	}

	if runResp.Data == nil {
		t.Error("Expected data in response")
	}
}

func TestHandleRun_InvalidJSON(t *testing.T) {
	handlers := setupTestHandlers(t)

	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRun(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var runResp RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp.Success {
		t.Error("Expected success=false for invalid JSON")
	}
}

// mockPlugin is a mock plugin for testing
type mockPlugin struct {
	id          string
	shouldError bool
}

func (m *mockPlugin) Manifest() types.PluginManifest {
	return types.PluginManifest{
		ID:          m.id,
		Name:        "Mock Plugin",
		Description: "A mock plugin for testing",
		Version:     "1.0.0",
	}
}

func (m *mockPlugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	if m.shouldError {
		return nil, &types.StructuredError{
			Code:       "MOCK_ERROR",
			Message:    "Mock plugin error",
			FieldPath:  "$.input",
			Suggestion: "Fix the mock error",
		}
	}

	ctx.State["output"] = "mock-output"

	return &types.PluginResult{
		Success: true,
		Data:    ctx.State,
	}, nil
}

func (m *mockPlugin) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockPlugin) Validate(ctx *types.ExecutionContext) error {
	return nil
}

func TestHandleRun_PluginError(t *testing.T) {
	pluginRegistry := registry.NewDefaultRegistry()
	logger := observability.NewJSONLogger(os.Stdout, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := executor.NewDefaultExecutor(pluginRegistry, logger, metrics)

	// Register a mock plugin that returns an error
	mockPlug := &mockPlugin{id: "mock-error-plugin", shouldError: true}
	if err := pluginRegistry.Register(mockPlug); err != nil {
		t.Fatalf("Failed to register mock plugin: %v", err)
	}

	handlers := NewHandlers(pluginRegistry, exec, logger)

	configYAML := `
name: test-pipeline
description: Test pipeline with error
stop_on_error: true
plugins:
  - id: mock-error-plugin
    optional: false
`

	reqBody := RunRequest{
		Config: configYAML,
		Input:  map[string]interface{}{},
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.HandleRun(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}

	var runResp RunResponse
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp.Success {
		t.Error("Expected success=false for plugin error")
	}

	if runResp.Error == nil {
		t.Fatal("Expected error in response")
	}

	if runResp.Error.Code != "MOCK_ERROR" {
		t.Errorf("Expected error code MOCK_ERROR, got %s", runResp.Error.Code)
	}
}

func TestNewHandlers(t *testing.T) {
	pluginRegistry := registry.NewDefaultRegistry()
	logger := observability.NewJSONLogger(os.Stdout, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := executor.NewDefaultExecutor(pluginRegistry, logger, metrics)

	handlers := NewHandlers(pluginRegistry, exec, logger)

	if handlers == nil {
		t.Fatal("Expected handlers, got nil")
	}

	if handlers.registry == nil {
		t.Error("Expected registry to be set")
	}

	if handlers.executor == nil {
		t.Error("Expected executor to be set")
	}

	if handlers.logger == nil {
		t.Error("Expected logger to be set")
	}

	if !handlers.ready {
		t.Error("Expected handlers to be ready")
	}

	if handlers.readyAt.IsZero() {
		t.Error("Expected readyAt to be set")
	}
}
