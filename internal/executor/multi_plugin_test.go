package executor

import (
	"context"
	"io"
	"testing"

	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/internal/registry"
	"github.com/yourusername/geee/plugins"
)

func TestExecute_MultiplePluginsSameID(t *testing.T) {
	// Setup
	reg := registry.NewDefaultRegistry()
	if err := plugins.RegisterAll(reg); err != nil {
		t.Fatalf("Failed to register plugins: %v", err)
	}

	logger := observability.NewJSONLogger(io.Discard, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := NewDefaultExecutor(reg, logger, metrics)

	// Create plan with 2 json-transformer plugins
	// The second step depends on the first to ensure sequential execution
	plan := &core.ExecutionPlan{
		Name: "test-multi-plugin",
		Steps: []core.ExecutionStep{
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "step1",
							"to":   "step1_out",
						},
					},
				},
			},
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "step1_out",
							"to":   "final_output",
						},
					},
				},
				DependsOn: []string{"json-transformer"}, // Wait for first transformer to complete
			},
		},
	}

	// Execute
	input := map[string]interface{}{
		"step1": "original_value",
	}

	result, err := exec.Execute(context.Background(), plan, input)
	if err != nil {
		t.Fatalf("Execution failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected successful execution")
	}

	t.Logf("Result data: %+v", result.Data)

	// Check that both transformations happened
	if result.Data["final_output"] != "original_value" {
		t.Errorf("Expected final_output='original_value', got %v", result.Data["final_output"])
	}

	// Original field should be gone
	if _, exists := result.Data["step1"]; exists {
		t.Error("Expected step1 field to be removed, but it's still present")
	}

	// Intermediate field should be gone
	if _, exists := result.Data["step1_out"]; exists {
		t.Error("Expected step1_out field to be removed, but it's still present")
	}
}
