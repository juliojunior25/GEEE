package executor

import (
	"testing"

	"github.com/yourusername/geee/internal/core"
)

func TestDAG_DuplicatePluginIDs(t *testing.T) {
	// Create steps with duplicate plugin IDs
	steps := []core.ExecutionStep{
		{PluginID: "json-transformer"},
		{PluginID: "json-transformer"},
	}

	dag := NewDAG(steps)

	// First call: both steps should be ready (no dependencies)
	ready := dag.GetReadySteps()
	if len(ready) != 2 {
		t.Fatalf("Expected 2 ready steps, got %d", len(ready))
	}

	t.Logf("Ready steps: %d", len(ready))
	for i, step := range ready {
		t.Logf("  Step %d: PluginID=%s, ptr=%p", i, step.PluginID, step)
	}

	// Mark first step as completed
	dag.MarkCompleted(ready[0])

	// Check that it's marked as completed
	if !dag.completed[0] {
		t.Error("Step 0 should be marked as completed")
	}

	// Second step should still be in progress
	if !dag.inProgress[1] {
		t.Error("Step 1 should still be in progress")
	}

	// Mark second step as completed
	dag.MarkCompleted(ready[1])

	// Check that both are completed
	if !dag.completed[0] || !dag.completed[1] {
		t.Error("Both steps should be marked as completed")
	}

	// DAG should be complete
	if !dag.IsComplete() {
		t.Error("DAG should be complete")
	}
}
