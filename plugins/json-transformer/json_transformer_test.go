package jsontransformer

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
	if manifest.ID != "json-transformer" {
		t.Errorf("expected ID 'json-transformer', got %s", manifest.ID)
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
		expected map[string]interface{}
		wantErr  bool
	}{
		{
			name: "simple field rename",
			config: map[string]interface{}{
				"mappings": []interface{}{
					map[string]interface{}{"from": "a", "to": "x"},
				},
			},
			input: map[string]interface{}{
				"a": 1,
			},
			expected: map[string]interface{}{
				"x": 1,
			},
			wantErr: false,
		},
		{
			name: "multiple field renames",
			config: map[string]interface{}{
				"mappings": []interface{}{
					map[string]interface{}{"from": "a", "to": "x"},
					map[string]interface{}{"from": "b", "to": "y"},
				},
			},
			input: map[string]interface{}{
				"a": 1,
				"b": "text",
				"c": true,
			},
			expected: map[string]interface{}{
				"x": 1,
				"y": "text",
				"c": true,
			},
			wantErr: false,
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

				for k, expectedVal := range tt.expected {
					actualVal, exists := result.Data[k]
					if !exists {
						t.Errorf("expected field %s not found in output", k)
						continue
					}
					if actualVal != expectedVal {
						t.Errorf("field %s: expected %v, got %v", k, expectedVal, actualVal)
					}
				}

				for k := range result.Data {
					if _, exists := tt.expected[k]; !exists {
						t.Errorf("unexpected field %s in output", k)
					}
				}
			}
		})
	}
}
