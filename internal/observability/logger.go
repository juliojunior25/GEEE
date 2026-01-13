package observability

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/yourusername/geee/pkg/types"
)

// LogLevel represents the severity of a log message
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// JSONLogger implements the types.Logger interface with JSON formatting
type JSONLogger struct {
	writer  io.Writer
	level   LogLevel
	verbose bool
}

// NewJSONLogger creates a new JSON logger
func NewJSONLogger(writer io.Writer, verbose bool) *JSONLogger {
	level := LogLevelInfo
	if verbose {
		level = LogLevelDebug
	}

	return &JSONLogger{
		writer:  writer,
		level:   level,
		verbose: verbose,
	}
}

// NewDefaultLogger creates a logger that writes to stdout
func NewDefaultLogger(verbose bool) *JSONLogger {
	return NewJSONLogger(os.Stdout, verbose)
}

// logEntry represents a structured log entry
type logEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// log writes a log entry if the level is appropriate
func (l *JSONLogger) log(level LogLevel, msg string, fields map[string]interface{}) {
	// Filter by log level
	if !l.shouldLog(level) {
		return
	}

	entry := logEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   msg,
		Fields:    fields,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple format if JSON marshaling fails
		_, _ = fmt.Fprintf(l.writer, "[%s] %s: %s\n", time.Now().Format(time.RFC3339), level, msg)
		return
	}

	_, _ = fmt.Fprintln(l.writer, string(data))
}

// shouldLog determines if a message at the given level should be logged
func (l *JSONLogger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
	}

	return levels[level] >= levels[l.level]
}

// Info logs an info message
func (l *JSONLogger) Info(msg string, fields map[string]interface{}) {
	l.log(LogLevelInfo, msg, fields)
}

// Error logs an error message
func (l *JSONLogger) Error(msg string, fields map[string]interface{}) {
	l.log(LogLevelError, msg, fields)
}

// Debug logs a debug message
func (l *JSONLogger) Debug(msg string, fields map[string]interface{}) {
	l.log(LogLevelDebug, msg, fields)
}

// Warn logs a warning message
func (l *JSONLogger) Warn(msg string, fields map[string]interface{}) {
	l.log(LogLevelWarn, msg, fields)
}

// Ensure JSONLogger implements types.Logger
var _ types.Logger = (*JSONLogger)(nil)
