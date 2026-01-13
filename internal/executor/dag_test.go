package executor

import (
	"testing"

	"github.com/yourusername/geee/internal/core"
)

func TestNewDAG(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2"},
	}

	dag := NewDAG(steps)
	if dag == nil {
		t.Fatal("Expected non-nil DAG")
	}

	if len(dag.steps) != 2 {
		t.Fatalf("Expected 2 steps, got %d", len(dag.steps))
	}
}

func TestGetReadySteps_NoDependencies(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2"},
		{PluginID: "step3"},
	}

	dag := NewDAG(steps)
	ready := dag.GetReadySteps()

	// All steps should be ready since none have dependencies
	if len(ready) != 3 {
		t.Fatalf("Expected 3 ready steps, got %d", len(ready))
	}
}

func TestGetReadySteps_WithDependencies(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2", DependsOn: []string{"step1"}},
		{PluginID: "step3", DependsOn: []string{"step2"}},
	}

	dag := NewDAG(steps)

	// First call: only step1 should be ready
	ready := dag.GetReadySteps()
	if len(ready) != 1 {
		t.Fatalf("Expected 1 ready step, got %d", len(ready))
	}
	if ready[0].PluginID != "step1" {
		t.Errorf("Expected step1, got %s", ready[0].PluginID)
	}

	// Mark step1 as completed
	dag.MarkCompleted(ready[0])

	// Second call: only step2 should be ready
	ready = dag.GetReadySteps()
	if len(ready) != 1 {
		t.Fatalf("Expected 1 ready step, got %d", len(ready))
	}
	if ready[0].PluginID != "step2" {
		t.Errorf("Expected step2, got %s", ready[0].PluginID)
	}

	// Mark step2 as completed
	dag.MarkCompleted(ready[0])

	// Third call: only step3 should be ready
	ready = dag.GetReadySteps()
	if len(ready) != 1 {
		t.Fatalf("Expected 1 ready step, got %d", len(ready))
	}
	if ready[0].PluginID != "step3" {
		t.Errorf("Expected step3, got %s", ready[0].PluginID)
	}

	// Mark step3 as completed
	dag.MarkCompleted(ready[0])

	// Fourth call: no more steps
	ready = dag.GetReadySteps()
	if len(ready) != 0 {
		t.Fatalf("Expected 0 ready steps, got %d", len(ready))
	}
}

func TestGetReadySteps_ParallelExecution(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2"},
		{PluginID: "step3", DependsOn: []string{"step1", "step2"}},
	}

	dag := NewDAG(steps)

	// First call: step1 and step2 should be ready
	ready := dag.GetReadySteps()
	if len(ready) != 2 {
		t.Fatalf("Expected 2 ready steps, got %d", len(ready))
	}

	readyIDs := make(map[string]*core.ExecutionStep)
	for _, step := range ready {
		readyIDs[step.PluginID] = step
	}

	if readyIDs["step1"] == nil || readyIDs["step2"] == nil {
		t.Error("Expected step1 and step2 to be ready")
	}

	// Mark step1 as completed (step2 still in progress)
	dag.MarkCompleted(readyIDs["step1"])

	// step3 should not be ready yet (step2 not completed)
	ready = dag.GetReadySteps()
	if len(ready) != 0 {
		t.Fatalf("Expected 0 ready steps, got %d", len(ready))
	}

	// Mark step2 as completed
	dag.MarkCompleted(readyIDs["step2"])

	// Now step3 should be ready
	ready = dag.GetReadySteps()
	if len(ready) != 1 {
		t.Fatalf("Expected 1 ready step, got %d", len(ready))
	}
	if ready[0].PluginID != "step3" {
		t.Errorf("Expected step3, got %s", ready[0].PluginID)
	}
}

func TestMarkCompleted(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
	}

	dag := NewDAG(steps)

	if dag.completed[0] {
		t.Fatal("step1 should not be completed initially")
	}

	ready := dag.GetReadySteps()
	if len(ready) != 1 {
		t.Fatal("Expected 1 ready step")
	}

	dag.MarkCompleted(ready[0])

	if !dag.completed[0] {
		t.Fatal("step1 should be marked as completed")
	}
}

func TestIsComplete(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2"},
	}

	dag := NewDAG(steps)

	if dag.IsComplete() {
		t.Fatal("DAG should not be complete initially")
	}

	ready := dag.GetReadySteps() // Mark as in progress
	if len(ready) != 2 {
		t.Fatal("Expected 2 ready steps")
	}
	dag.MarkCompleted(ready[0])

	if dag.IsComplete() {
		t.Fatal("DAG should not be complete with only 1 step completed")
	}

	dag.MarkCompleted(ready[1])

	if !dag.IsComplete() {
		t.Fatal("DAG should be complete with all steps completed")
	}
}

func TestGetCompleted(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2"},
	}

	dag := NewDAG(steps)
	ready := dag.GetReadySteps() // Mark as in progress
	if len(ready) != 2 {
		t.Fatal("Expected 2 ready steps")
	}
	dag.MarkCompleted(ready[0])

	completed := dag.GetCompleted()
	if len(completed) != 1 {
		t.Fatalf("Expected 1 completed step, got %d", len(completed))
	}
	if completed[0] != "step1" && completed[0] != "step2" {
		t.Errorf("Expected step1 or step2, got %s", completed[0])
	}
}

func TestGetInProgress(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2"},
	}

	dag := NewDAG(steps)
	ready := dag.GetReadySteps() // Mark as in progress

	inProgress := dag.GetInProgress()
	if len(inProgress) != 2 {
		t.Fatalf("Expected 2 in-progress steps, got %d", len(inProgress))
	}

	dag.MarkCompleted(ready[0])

	inProgress = dag.GetInProgress()
	if len(inProgress) != 1 {
		t.Fatalf("Expected 1 in-progress step after completion, got %d", len(inProgress))
	}
}

func TestGetPending(t *testing.T) {
	steps := []core.ExecutionStep{
		{PluginID: "step1"},
		{PluginID: "step2", DependsOn: []string{"step1"}},
	}

	dag := NewDAG(steps)

	pending := dag.GetPending()
	if len(pending) != 2 {
		t.Fatalf("Expected 2 pending steps initially, got %d", len(pending))
	}

	dag.GetReadySteps() // Mark step1 as in progress

	pending = dag.GetPending()
	if len(pending) != 1 {
		t.Fatalf("Expected 1 pending step, got %d", len(pending))
	}
	if pending[0] != "step2" {
		t.Errorf("Expected step2, got %s", pending[0])
	}
}
