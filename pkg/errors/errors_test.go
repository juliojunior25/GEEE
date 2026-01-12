package errors

import (
	"errors"
	"testing"

	"github.com/yourusername/geee/pkg/types"
)

func TestStructuredError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *StructuredError
		expected string
	}{
		{
			name: "error with plugin ID",
			err: &StructuredError{
				StructuredError: &types.StructuredError{
					Code:       ErrCodeValidation,
					Message:    "Field is required",
					Suggestion: "Add the field",
					PluginID:   "test-plugin",
				},
			},
			expected: "[VALIDATION_FAILED] Field is required: Add the field (plugin: test-plugin)",
		},
		{
			name: "error without plugin ID",
			err: &StructuredError{
				StructuredError: &types.StructuredError{
					Code:       ErrCodeConfigLoad,
					Message:    "Failed to load config",
					Suggestion: "Check file path",
				},
			},
			expected: "[CONFIG_LOAD_FAILED] Failed to load config: Check file path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStructuredError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &StructuredError{
		StructuredError: &types.StructuredError{
			Code: ErrCodeExecution,
		},
		Cause: cause,
	}

	if got := err.Unwrap(); got != cause {
		t.Errorf("Unwrap() = %v, want %v", got, cause)
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("test message", "$.field", "test suggestion")

	if err.Code != ErrCodeValidation {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeValidation)
	}
	if err.Message != "test message" {
		t.Errorf("Message = %v, want %v", err.Message, "test message")
	}
	if err.FieldPath != "$.field" {
		t.Errorf("FieldPath = %v, want %v", err.FieldPath, "$.field")
	}
	if err.Suggestion != "test suggestion" {
		t.Errorf("Suggestion = %v, want %v", err.Suggestion, "test suggestion")
	}
}

func TestNewSchemaValidationError(t *testing.T) {
	err := NewSchemaValidationError("schema error", "$.input", "fix schema", "my-plugin")

	if err.Code != ErrCodeSchemaValidation {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeSchemaValidation)
	}
	if err.PluginID != "my-plugin" {
		t.Errorf("PluginID = %v, want %v", err.PluginID, "my-plugin")
	}
}

func TestNewExecutionError(t *testing.T) {
	err := NewExecutionError("my-plugin", "execution failed", "retry operation")

	if err.Code != ErrCodeExecution {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeExecution)
	}
	if err.PluginID != "my-plugin" {
		t.Errorf("PluginID = %v, want %v", err.PluginID, "my-plugin")
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("slow-plugin", 30)

	if err.Code != ErrCodeTimeout {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeTimeout)
	}
	if err.PluginID != "slow-plugin" {
		t.Errorf("PluginID = %v, want %v", err.PluginID, "slow-plugin")
	}
}

func TestNewMemoryLimitError(t *testing.T) {
	err := NewMemoryLimitError("memory-hog", 100)

	if err.Code != ErrCodeMemoryLimit {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeMemoryLimit)
	}
	if err.PluginID != "memory-hog" {
		t.Errorf("PluginID = %v, want %v", err.PluginID, "memory-hog")
	}
}

func TestNewPluginNotFoundError(t *testing.T) {
	err := NewPluginNotFoundError("missing-plugin")

	if err.Code != ErrCodePluginNotFound {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodePluginNotFound)
	}
	if err.PluginID != "missing-plugin" {
		t.Errorf("PluginID = %v, want %v", err.PluginID, "missing-plugin")
	}
}

func TestNewConfigLoadError(t *testing.T) {
	cause := errors.New("file not found")
	err := NewConfigLoadError("config load failed", "check path", cause)

	if err.Code != ErrCodeConfigLoad {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeConfigLoad)
	}
	if err.Cause != cause {
		t.Errorf("Cause = %v, want %v", err.Cause, cause)
	}
}

func TestNewInvalidConfigError(t *testing.T) {
	err := NewInvalidConfigError("invalid config", "$.plugins", "add plugins")

	if err.Code != ErrCodeInvalidConfig {
		t.Errorf("Code = %v, want %v", err.Code, ErrCodeInvalidConfig)
	}
	if err.FieldPath != "$.plugins" {
		t.Errorf("FieldPath = %v, want %v", err.FieldPath, "$.plugins")
	}
}
