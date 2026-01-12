package errors

import (
	"fmt"

	"github.com/yourusername/geee/pkg/types"
)

// ErrorCode represents standardized error codes
const (
	// ErrCodeValidation indicates validation failure
	ErrCodeValidation = "VALIDATION_FAILED"

	// ErrCodeSchemaValidation indicates schema validation failure
	ErrCodeSchemaValidation = "SCHEMA_VALIDATION_FAILED"

	// ErrCodeExecution indicates plugin execution failure
	ErrCodeExecution = "EXECUTION_FAILED"

	// ErrCodeTimeout indicates execution timeout
	ErrCodeTimeout = "TIMEOUT"

	// ErrCodeMemoryLimit indicates memory limit exceeded
	ErrCodeMemoryLimit = "MEMORY_LIMIT_EXCEEDED"

	// ErrCodePluginNotFound indicates plugin not found in registry
	ErrCodePluginNotFound = "PLUGIN_NOT_FOUND"

	// ErrCodeConfigLoad indicates configuration loading failure
	ErrCodeConfigLoad = "CONFIG_LOAD_FAILED"

	// ErrCodeInvalidConfig indicates invalid configuration
	ErrCodeInvalidConfig = "INVALID_CONFIG"
)

// StructuredError wraps types.StructuredError with error interface methods
type StructuredError struct {
	*types.StructuredError
	Cause error
}

// Error implements the error interface
func (e *StructuredError) Error() string {
	if e.PluginID != "" {
		return fmt.Sprintf("[%s] %s: %s (plugin: %s)", e.Code, e.Message, e.Suggestion, e.PluginID)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Suggestion)
}

// Unwrap returns the underlying cause
func (e *StructuredError) Unwrap() error {
	return e.Cause
}

// NewValidationError creates a validation error
func NewValidationError(message, fieldPath, suggestion string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeValidation,
			Message:    message,
			FieldPath:  fieldPath,
			Suggestion: suggestion,
		},
	}
}

// NewSchemaValidationError creates a schema validation error
func NewSchemaValidationError(message, fieldPath, suggestion, pluginID string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeSchemaValidation,
			Message:    message,
			FieldPath:  fieldPath,
			Suggestion: suggestion,
			PluginID:   pluginID,
		},
	}
}

// NewExecutionError creates an execution error
func NewExecutionError(pluginID, message, suggestion string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeExecution,
			Message:    message,
			Suggestion: suggestion,
			PluginID:   pluginID,
		},
	}
}

// NewPluginError creates a plugin-specific error
func NewPluginError(pluginID, message, suggestion string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeExecution,
			Message:    message,
			Suggestion: suggestion,
			PluginID:   pluginID,
		},
	}
}

// NewResourceLimitError creates a resource limit error
func NewResourceLimitError(resource, current, limit string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeMemoryLimit,
			Message:    fmt.Sprintf("Resource limit exceeded for %s: current=%s, limit=%s", resource, current, limit),
			Suggestion: fmt.Sprintf("Increase the %s limit in plugin configuration", resource),
		},
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(pluginID string, timeoutSeconds int) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeTimeout,
			Message:    fmt.Sprintf("Plugin execution exceeded timeout of %d seconds", timeoutSeconds),
			Suggestion: "Increase timeout in plugin configuration or optimize plugin performance",
			PluginID:   pluginID,
		},
	}
}

// NewMemoryLimitError creates a memory limit error
func NewMemoryLimitError(pluginID string, limitMB int) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeMemoryLimit,
			Message:    fmt.Sprintf("Plugin execution exceeded memory limit of %d MB", limitMB),
			Suggestion: "Increase memory limit in plugin configuration or optimize plugin memory usage",
			PluginID:   pluginID,
		},
	}
}

// NewPluginNotFoundError creates a plugin not found error
func NewPluginNotFoundError(pluginID string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodePluginNotFound,
			Message:    fmt.Sprintf("Plugin '%s' not found in registry", pluginID),
			Suggestion: "Check plugin ID spelling or ensure plugin is registered",
			PluginID:   pluginID,
		},
	}
}

// NewConfigLoadError creates a config load error
func NewConfigLoadError(message, suggestion string, cause error) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeConfigLoad,
			Message:    message,
			Suggestion: suggestion,
		},
		Cause: cause,
	}
}

// NewInvalidConfigError creates an invalid config error
func NewInvalidConfigError(message, fieldPath, suggestion string) *StructuredError {
	return &StructuredError{
		StructuredError: &types.StructuredError{
			Code:       ErrCodeInvalidConfig,
			Message:    message,
			FieldPath:  fieldPath,
			Suggestion: suggestion,
		},
	}
}
