package observability

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, false)

	logger.Info("test message", map[string]interface{}{
		"key": "value",
	})

	var entry logEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", entry.Level)
	}

	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}

	if entry.Fields["key"] != "value" {
		t.Errorf("Expected field 'key' to be 'value', got '%v'", entry.Fields["key"])
	}
}

func TestJSONLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, false)

	logger.Error("error occurred", map[string]interface{}{
		"error": "something went wrong",
	})

	var entry logEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Level != "error" {
		t.Errorf("Expected level 'error', got '%s'", entry.Level)
	}

	if entry.Message != "error occurred" {
		t.Errorf("Expected message 'error occurred', got '%s'", entry.Message)
	}
}

func TestJSONLogger_Debug_Verbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, true) // verbose mode

	logger.Debug("debug message", nil)

	var entry logEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Level != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", entry.Level)
	}
}

func TestJSONLogger_Debug_NotVerbose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, false) // non-verbose mode

	logger.Debug("debug message", nil)

	// Debug should not be logged in non-verbose mode
	if buf.Len() > 0 {
		t.Error("Expected no output in non-verbose mode for debug messages")
	}
}

func TestJSONLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, false)

	logger.Warn("warning message", map[string]interface{}{
		"warning": "be careful",
	})

	var entry logEntry
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	if entry.Level != "warn" {
		t.Errorf("Expected level 'warn', got '%s'", entry.Level)
	}
}

func TestJSONLogger_MultipleMessages(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewJSONLogger(buf, false)

	logger.Info("message 1", nil)
	logger.Error("message 2", nil)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 log lines, got %d", len(lines))
	}
}

func TestNewDefaultLogger(t *testing.T) {
	logger := NewDefaultLogger(true)
	if logger == nil {
		t.Error("Expected non-nil logger")
	}

	if !logger.verbose {
		t.Error("Expected verbose to be true")
	}
}
