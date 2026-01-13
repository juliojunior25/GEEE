package httpfetcher

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/pkg/types"
)

func TestNew(t *testing.T) {
	plugin := New()
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}

	manifest := plugin.Manifest()
	if manifest.ID != "http-fetcher" {
		t.Errorf("expected ID 'http-fetcher', got %s", manifest.ID)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", manifest.Version)
	}
}

func TestExecute(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	plugin := New()

	config := map[string]interface{}{
		"url_field":      "url",
		"response_field": "response",
	}

	input := map[string]interface{}{
		"url": server.URL,
	}

	execCtx := &types.ExecutionContext{
		Context:     context.Background(),
		ExecutionID: "test-123",
		State:       input,
		Config:      config,
		Logger:      observability.NewDefaultLogger(false),
		Metrics:     observability.NewNoOpMetricsCollector(),
	}

	result, err := plugin.Execute(execCtx)
	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if !result.Success {
		t.Error("expected successful result")
	}

	// Verify response is present
	responseValue, exists := result.Data["response"]
	if !exists {
		t.Error("expected response field in output")
		return
	}

	// Verify response structure
	response, ok := responseValue.(Response)
	if !ok {
		t.Errorf("expected response to be Response type, got %T", responseValue)
		return
	}

	if response.Status != http.StatusOK {
		t.Errorf("expected status 200, got %d", response.Status)
	}

	if response.Body != `{"message": "success"}` {
		t.Errorf("unexpected response body: %s", response.Body)
	}
}
