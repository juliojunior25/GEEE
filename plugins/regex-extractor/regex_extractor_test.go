package regexextractor

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
	if manifest.ID != "regex-extractor" {
		t.Errorf("expected ID 'regex-extractor', got %s", manifest.ID)
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
		expected map[string]string
		wantErr  bool
	}{
		{
			name: "extract email",
			config: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"name":    "email",
						"pattern": `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
					},
				},
				"input_field": "text",
			},
			input: map[string]interface{}{
				"text": "Contact us at foo@bar.com for more info",
			},
			expected: map[string]string{
				"email": "foo@bar.com",
			},
			wantErr: false,
		},
		{
			name: "extract multiple patterns",
			config: map[string]interface{}{
				"patterns": []interface{}{
					map[string]interface{}{
						"name":    "email",
						"pattern": `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
					},
					map[string]interface{}{
						"name":    "phone",
						"pattern": `\d{3,}`,
					},
				},
				"input_field": "text",
			},
			input: map[string]interface{}{
				"text": "email: foo@bar.com, phone: 123456",
			},
			expected: map[string]string{
				"email": "foo@bar.com",
				"phone": "123456",
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
			}
		})
	}
}
