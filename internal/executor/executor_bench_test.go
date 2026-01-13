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

func BenchmarkExecutor_SimplePipeline(b *testing.B) {
	// Setup
	reg := registry.NewDefaultRegistry()
	if err := plugins.RegisterAll(reg); err != nil {
		b.Fatalf("Failed to register plugins: %v", err)
	}

	logger := observability.NewJSONLogger(io.Discard, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name: "benchmark",
		Steps: []core.ExecutionStep{
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "input",
							"to":   "output",
						},
					},
				},
			},
		},
	}

	input := map[string]interface{}{
		"input": "test_value",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exec.Execute(ctx, plan, input)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

func BenchmarkExecutor_MultiplePlugins(b *testing.B) {
	// Setup
	reg := registry.NewDefaultRegistry()
	if err := plugins.RegisterAll(reg); err != nil {
		b.Fatalf("Failed to register plugins: %v", err)
	}

	logger := observability.NewJSONLogger(io.Discard, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name: "benchmark",
		Steps: []core.ExecutionStep{
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "step1",
							"to":   "step2",
						},
					},
				},
			},
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "step2",
							"to":   "step3",
						},
					},
				},
				DependsOn: []string{"json-transformer"},
			},
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "step3",
							"to":   "output",
						},
					},
				},
				DependsOn: []string{"json-transformer"},
			},
		},
	}

	input := map[string]interface{}{
		"step1": "test_value",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exec.Execute(ctx, plan, input)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}

func BenchmarkExecutor_ParallelPlugins(b *testing.B) {
	// Setup
	reg := registry.NewDefaultRegistry()
	if err := plugins.RegisterAll(reg); err != nil {
		b.Fatalf("Failed to register plugins: %v", err)
	}

	logger := observability.NewJSONLogger(io.Discard, false)
	metrics := observability.NewNoOpMetricsCollector()
	exec := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name: "benchmark",
		Steps: []core.ExecutionStep{
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "input1",
							"to":   "output1",
						},
					},
				},
			},
			{
				PluginID: "json-transformer",
				Config: map[string]interface{}{
					"mappings": []interface{}{
						map[string]interface{}{
							"from": "input2",
							"to":   "output2",
						},
					},
				},
			},
		},
	}

	input := map[string]interface{}{
		"input1": "value1",
		"input2": "value2",
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := exec.Execute(ctx, plan, input)
		if err != nil {
			b.Fatalf("Execution failed: %v", err)
		}
	}
}
