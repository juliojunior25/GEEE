package templaterenderer

import (
	"context"
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
	if manifest.ID != "template-renderer" {
		t.Errorf("expected ID 'template-renderer', got %s", manifest.ID)
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %s", manifest.Version)
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		input    map[string]interface{}
		expected string
		wantErr  bool
	}{
		{
			name:   "simple template",
			config: map[string]interface{}{},
			input: map[string]interface{}{
				"template": "Hello {{.Name}}!",
				"data": map[string]interface{}{
					"Name": "World",
				},
			},
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:   "template with multiple fields",
			config: map[string]interface{}{},
			input: map[string]interface{}{
				"template": "{{.Greeting}} {{.Name}}, you are {{.Age}} years old",
				"data": map[string]interface{}{
					"Greeting": "Hello",
					"Name":     "Alice",
					"Age":      30,
				},
			},
			expected: "Hello Alice, you are 30 years old",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin := New()

			execCtx := &types.ExecutionContext{
				Context:     context.Background(),
				ExecutionID: "test-123",
				State:       tt.input,
				Config:      tt.config,
				Logger:      observability.NewDefaultLogger(false),
				Metrics:     observability.NewNoOpMetricsCollector(),
			}

			result, err := plugin.Execute(execCtx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("expected non-nil result")
				}

				if !result.Success {
					t.Error("expected successful result")
				}

				outputValue, exists := result.Data["result"]
				if !exists {
					t.Error("expected result field in output")
					return
				}

				output, ok := outputValue.(string)
				if !ok {
					t.Errorf("expected output to be string, got %T", outputValue)
					return
				}

				if output != tt.expected {
					t.Errorf("expected output %q, got %q", tt.expected, output)
				}
			}
		})
	}
}
