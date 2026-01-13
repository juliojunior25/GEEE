package executor

import (
	"sync"

	"github.com/yourusername/geee/internal/core"
)

// DAG represents a directed acyclic graph of plugin execution steps
type DAG struct {
	steps             []*core.ExecutionStep
	stepToIdx         map[*core.ExecutionStep]int // Maps step pointer to its index
	completed         map[int]bool                // Maps step index to completion status
	inProgress        map[int]bool                // Maps step index to in-progress status
	completedPluginID map[string]bool             // Tracks which plugin IDs have completed (for dependency resolution)
	mu                sync.RWMutex
}

// NewDAG creates a new DAG from execution steps
func NewDAG(steps []core.ExecutionStep) *DAG {
	stepsCopy := make([]*core.ExecutionStep, len(steps))
	stepToIdx := make(map[*core.ExecutionStep]int)
	for i := range steps {
		stepsCopy[i] = &steps[i]
		stepToIdx[stepsCopy[i]] = i
	}

	return &DAG{
		steps:             stepsCopy,
		stepToIdx:         stepToIdx,
		completed:         make(map[int]bool),
		inProgress:        make(map[int]bool),
		completedPluginID: make(map[string]bool),
	}
}

// GetReadySteps returns steps that are ready to execute
// A step is ready if all its dependencies have been completed
func (d *DAG) GetReadySteps() []*core.ExecutionStep {
	d.mu.Lock()
	defer d.mu.Unlock()

	ready := make([]*core.ExecutionStep, 0)

	for i, step := range d.steps {
		// Skip if already completed or in progress
		if d.completed[i] || d.inProgress[i] {
			continue
		}

		// Check if all dependencies are completed
		if d.areDependenciesCompleted(step) {
			ready = append(ready, step)
			d.inProgress[i] = true
		}
	}

	return ready
}

// MarkCompleted marks a step as completed
func (d *DAG) MarkCompleted(step *core.ExecutionStep) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if idx, ok := d.stepToIdx[step]; ok {
		d.completed[idx] = true
		delete(d.inProgress, idx)
		d.completedPluginID[step.PluginID] = true
	}
}

// areDependenciesCompleted checks if all dependencies of a step are completed
func (d *DAG) areDependenciesCompleted(step *core.ExecutionStep) bool {
	for _, depID := range step.DependsOn {
		if !d.completedPluginID[depID] {
			return false
		}
	}
	return true
}

// IsComplete returns true if all steps have been completed
func (d *DAG) IsComplete() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return len(d.completed) == len(d.steps)
}

// GetCompleted returns the list of completed plugin IDs
func (d *DAG) GetCompleted() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	completed := make([]string, 0, len(d.completed))
	for idx := range d.completed {
		completed = append(completed, d.steps[idx].PluginID)
	}
	return completed
}

// GetInProgress returns the list of in-progress plugin IDs
func (d *DAG) GetInProgress() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	inProgress := make([]string, 0, len(d.inProgress))
	for idx := range d.inProgress {
		inProgress = append(inProgress, d.steps[idx].PluginID)
	}
	return inProgress
}

// GetPending returns the list of pending plugin IDs
func (d *DAG) GetPending() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	pending := make([]string, 0)
	for i, step := range d.steps {
		if !d.completed[i] && !d.inProgress[i] {
			pending = append(pending, step.PluginID)
		}
	}
	return pending
}
