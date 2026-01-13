package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/geee/internal/executor"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/internal/registry"
	"github.com/yourusername/geee/internal/server"
	"github.com/yourusername/geee/plugins"
)

// setupTestServer creates a test HTTP server with all plugins registered
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Setup components
	pluginRegistry := registry.NewDefaultRegistry()
	if err := plugins.RegisterAll(pluginRegistry); err != nil {
		t.Fatalf("Failed to register plugins: %v", err)
	}

	logger := observability.NewJSONLogger(io.Discard, false)
	metrics := observability.NewPrometheusMetricsCollector()
	exec := executor.NewDefaultExecutor(pluginRegistry, logger, metrics)

	// Create server
	config := server.DefaultServerConfig()
	srv := server.NewServer(config, pluginRegistry, exec, logger, metrics)

	// Create test server
	return httptest.NewServer(srv.Handler())
}

func TestServer_HealthEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to call /health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var healthResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&healthResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if healthResp["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", healthResp["status"])
	}

	// Check plugins in health response
	plugins, ok := healthResp["plugins"].(map[string]interface{})
	if !ok || len(plugins) == 0 {
		t.Error("Expected plugins in health response")
	}
}

func TestServer_ReadyEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/ready")
	if err != nil {
		t.Fatalf("Failed to call /ready: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var readyResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&readyResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if readyResp["ready"] != true {
		t.Errorf("Expected ready=true, got %v", readyResp["ready"])
	}
}

func TestServer_PluginsEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/plugins")
	if err != nil {
		t.Fatalf("Failed to call /plugins: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var pluginsResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&pluginsResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	count, ok := pluginsResp["count"].(float64)
	if !ok || count == 0 {
		t.Error("Expected non-zero plugin count")
	}

	pluginsList, ok := pluginsResp["plugins"].([]interface{})
	if !ok || len(pluginsList) == 0 {
		t.Error("Expected plugins list")
	}
}

func TestServer_RunEndpoint_Success(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Create a valid pipeline config
	config := map[string]interface{}{
		"config": `
name: integration-test
description: Integration test pipeline
plugins:
  - id: json-transformer
    config:
      mappings:
        - from: input_field
          to: output_field
`,
		"input": map[string]interface{}{
			"input_field": "test_value",
		},
	}

	body, _ := json.Marshal(config)
	resp, err := http.Post(ts.URL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to call /run: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var runResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp["success"] != true {
		t.Errorf("Expected success=true, got %v. Error: %v", runResp["success"], runResp["error"])
	}

	// Check execution_id is present
	if runResp["execution_id"] == "" {
		t.Error("Expected execution_id in response")
	}

	// Check data is present
	data, ok := runResp["data"].(map[string]interface{})
	if !ok {
		t.Error("Expected data in response")
	}

	// Verify transformation happened
	if data["output_field"] != "test_value" {
		t.Errorf("Expected output_field='test_value', got %v", data["output_field"])
	}
}

func TestServer_RunEndpoint_InvalidConfig(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	config := map[string]interface{}{
		"config": "invalid yaml: [",
		"input":  map[string]interface{}{},
	}

	body, _ := json.Marshal(config)
	resp, err := http.Post(ts.URL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to call /run: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var runResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp["success"] != false {
		t.Error("Expected success=false for invalid config")
	}
}

func TestServer_RunEndpoint_MissingConfig(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	config := map[string]interface{}{
		"config": "",
		"input":  map[string]interface{}{},
	}

	body, _ := json.Marshal(config)
	resp, err := http.Post(ts.URL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to call /run: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var runResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp["success"] != false {
		t.Error("Expected success=false for missing config")
	}

	errorData, ok := runResp["error"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected error in response")
	}

	if errorData["code"] != "MISSING_CONFIG" {
		t.Errorf("Expected error code MISSING_CONFIG, got %v", errorData["code"])
	}
}

func TestServer_RunEndpoint_MultiplePlugins(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	config := map[string]interface{}{
		"config": `
name: multi-plugin-test
description: Test with multiple plugins
plugins:
  - id: json-transformer
    config:
      mappings:
        - from: step1
          to: step1_out
  - id: json-transformer
    config:
      mappings:
        - from: step1_out
          to: final_output
    depends_on:
      - json-transformer
`,
		"input": map[string]interface{}{
			"step1": "original_value",
		},
	}

	body, _ := json.Marshal(config)
	resp, err := http.Post(ts.URL+"/run", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to call /run: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		t.Errorf("Expected status 200, got %d. Body: %s", resp.StatusCode, string(bodyBytes))
	}

	var runResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&runResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if runResp["success"] != true {
		t.Errorf("Expected success=true, got %v", runResp["success"])
	}

	data, ok := runResp["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data in response")
	}

	if data["final_output"] != "original_value" {
		t.Errorf("Expected final_output='original_value', got %v", data["final_output"])
	}
}

func TestServer_CORS(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	req, _ := http.NewRequest("OPTIONS", ts.URL+"/health", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to call OPTIONS: %v", err)
	}
	defer resp.Body.Close()

	// Check CORS headers are present
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin == "" {
		t.Error("Expected Access-Control-Allow-Origin header")
	}
}

func TestServer_ConcurrentRequests(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	// Send multiple concurrent requests
	concurrency := 10
	done := make(chan bool, concurrency)

	config := map[string]interface{}{
		"config": `
name: concurrent-test
plugins:
  - id: json-transformer
    config:
      mappings:
        - from: input
          to: output
`,
		"input": map[string]interface{}{
			"input": "test",
		},
	}

	body, _ := json.Marshal(config)

	for i := 0; i < concurrency; i++ {
		go func() {
			resp, err := http.Post(ts.URL+"/run", "application/json", bytes.NewReader(body))
			if err != nil {
				t.Errorf("Request failed: %v", err)
				done <- false
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
				done <- false
				return
			}

			done <- true
		}()
	}

	// Wait for all requests to complete
	timeout := time.After(5 * time.Second)
	for i := 0; i < concurrency; i++ {
		select {
		case <-done:
			// Request completed
		case <-timeout:
			t.Fatal("Timeout waiting for concurrent requests")
		}
	}
}

func TestServer_MetricsEndpoint(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatalf("Failed to call /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check content type is text/plain (Prometheus format)
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/plain; version=0.0.4; charset=utf-8" {
		t.Logf("Content-Type: %s", contentType)
		// Don't fail, just log - Prometheus handler sets its own content type
	}

	// Read body and check for metrics
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	body := string(bodyBytes)
	if len(body) == 0 {
		t.Error("Expected non-empty metrics response")
	}

	// Check for some expected metrics (they should exist even if zero)
	expectedMetrics := []string{
		"geee_active_executions",
		"geee_http_request_total",
	}

	for _, metric := range expectedMetrics {
		if !bytes.Contains(bodyBytes, []byte(metric)) {
			t.Errorf("Expected metric %s in response", metric)
		}
	}
}
