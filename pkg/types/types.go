package types

import (
	"context"
	"time"
)

// StructuredError provides detailed error information with suggestions
type StructuredError struct {
	// Code is the standardized error code
	Code string `json:"code"`

	// Message is the human-readable error message
	Message string `json:"message"`

	// FieldPath is the JSON path to the problematic field (if applicable)
	FieldPath string `json:"field_path,omitempty"`

	// Suggestion provides actionable guidance to fix the error
	Suggestion string `json:"suggestion,omitempty"`

	// PluginID identifies which plugin caused the error
	PluginID string `json:"plugin_id,omitempty"`
}

// Error implements the error interface for StructuredError
func (e *StructuredError) Error() string {
	return e.Message
}

// PluginResult represents the result of a plugin execution
type PluginResult struct {
	// Success indicates if the plugin execution was successful
	Success bool `json:"success"`

	// Data contains the output data from the plugin
	Data map[string]interface{} `json:"data,omitempty"`

	// Error contains error information if Success is false
	Error *StructuredError `json:"error,omitempty"`

	// Metadata contains execution metadata
	Metadata *ExecutionMetadata `json:"metadata,omitempty"`
}

// ExecutionMetadata contains metadata about plugin execution
type ExecutionMetadata struct {
	// PluginID is the unique identifier of the plugin
	PluginID string `json:"plugin_id"`

	// ExecutionID is the unique identifier for this execution
	ExecutionID string `json:"execution_id"`

	// StartTime is when the execution started
	StartTime time.Time `json:"start_time"`

	// EndTime is when the execution completed
	EndTime time.Time `json:"end_time"`

	// DurationMs is the execution duration in milliseconds
	DurationMs int64 `json:"duration_ms"`
}

// ExecutionContext provides context for plugin execution
type ExecutionContext struct {
	// Context is the Go context for cancellation and timeouts
	Context context.Context

	// ExecutionID is the unique identifier for this execution
	ExecutionID string

	// State is the current state/data being processed
	State map[string]interface{}

	// Config contains plugin-specific configuration
	Config map[string]interface{}

	// Logger provides structured logging capabilities
	Logger Logger

	// Metrics provides metrics collection capabilities
	Metrics MetricsCollector
}

// Logger provides structured logging interface
type Logger interface {
	Info(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	Debug(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
}

// MetricsCollector provides metrics collection interface
type MetricsCollector interface {
	// IncrementCounter increments a counter metric
	IncrementCounter(name string, labels map[string]string)

	// RecordDuration records a duration metric
	RecordDuration(name string, duration time.Duration, labels map[string]string)

	// SetGauge sets a gauge metric
	SetGauge(name string, value float64, labels map[string]string)
}

// PluginManifest describes plugin capabilities and requirements
type PluginManifest struct {
	// ID is the unique plugin identifier
	ID string `json:"id" yaml:"id"`

	// Name is the human-readable plugin name
	Name string `json:"name" yaml:"name"`

	// Version is the plugin version
	Version string `json:"version" yaml:"version"`

	// Description describes what the plugin does
	Description string `json:"description" yaml:"description"`

	// InputSchema defines expected input structure (JSON Schema)
	InputSchema map[string]interface{} `json:"input_schema,omitempty" yaml:"input_schema,omitempty"`

	// OutputSchema defines output structure (JSON Schema)
	OutputSchema map[string]interface{} `json:"output_schema,omitempty" yaml:"output_schema,omitempty"`

	// ConfigSchema defines configuration structure (JSON Schema)
	ConfigSchema map[string]interface{} `json:"config_schema,omitempty" yaml:"config_schema,omitempty"`

	// Optional indicates if plugin failure should not abort pipeline
	Optional bool `json:"optional" yaml:"optional"`

	// Timeout is the maximum execution time in seconds
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// MaxMemoryMB is the maximum memory limit in megabytes
	MaxMemoryMB int `json:"max_memory_mb,omitempty" yaml:"max_memory_mb,omitempty"`
}
