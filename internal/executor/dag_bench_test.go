package executor

import (
	"testing"

	"github.com/yourusername/geee/internal/core"
)

func BenchmarkDAG_GetReadySteps_NoDependencies(b *testing.B) {
	steps := make([]core.ExecutionStep, 10)
	for i := 0; i < 10; i++ {
		steps[i] = core.ExecutionStep{
			PluginID: "test-plugin",
			Config:   make(map[string]interface{}),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dag := NewDAG(steps)
		ready := dag.GetReadySteps()
		if len(ready) != 10 {
			b.Fatalf("Expected 10 ready steps, got %d", len(ready))
		}
	}
}

func BenchmarkDAG_GetReadySteps_WithDependencies(b *testing.B) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2", DependsOn: []string{"step1"}},
		{PluginID: "step3", DependsOn: []string{"step2"}},
		{PluginID: "step4", DependsOn: []string{"step3"}},
		{PluginID: "step5", DependsOn: []string{"step4"}},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dag := NewDAG(steps)

		// Simulate execution
		for !dag.IsComplete() {
			ready := dag.GetReadySteps()
			for _, step := range ready {
				dag.MarkCompleted(step)
			}
		}
	}
}

func BenchmarkDAG_MarkCompleted(b *testing.B) {
	steps := []core.ExecutionStep{
		{PluginID: "test-plugin"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		dag := NewDAG(steps)
		ready := dag.GetReadySteps()
		b.StartTimer()

		dag.MarkCompleted(ready[0])
	}
}

func BenchmarkDAG_IsComplete(b *testing.B) {
	steps := make([]core.ExecutionStep, 100)
	for i := 0; i < 100; i++ {
		steps[i] = core.ExecutionStep{
			PluginID: "test-plugin",
		}
	}

	dag := NewDAG(steps)
	ready := dag.GetReadySteps()
	for _, step := range ready {
		dag.MarkCompleted(step)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dag.IsComplete()
	}
}
