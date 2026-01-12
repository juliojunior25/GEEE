package server

import "time"

// RunRequest represents a request to execute a pipeline
type RunRequest struct {
	Config string                 `json:"config"` // Path to config file or inline YAML
	Input  map[string]interface{} `json:"input"`  // Input data for the pipeline
}

// RunResponse represents the response from a pipeline execution
type RunResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Error       *ErrorResponse         `json:"error,omitempty"`
	ExecutionID string                 `json:"execution_id"`
	DurationMS  int64                  `json:"duration_ms"`
}

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	FieldPath  string `json:"field_path,omitempty"`
	Suggestion string `json:"suggestion,omitempty"`
	PluginID   string `json:"plugin_id,omitempty"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status  string            `json:"status"`
	Plugins map[string]string `json:"plugins,omitempty"`
}

// ReadyResponse represents the readiness probe response
type ReadyResponse struct {
	Ready   bool      `json:"ready"`
	Message string    `json:"message,omitempty"`
	Since   time.Time `json:"since,omitempty"`
}

// PluginInfo represents plugin information
type PluginInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Inputs      []string `json:"inputs"`
	Outputs     []string `json:"outputs"`
}

// PluginsResponse represents the plugins list response
type PluginsResponse struct {
	Plugins []PluginInfo `json:"plugins"`
	Count   int          `json:"count"`
}
