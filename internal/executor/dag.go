package executor

import (
	"sync"

	"github.com/yourusername/geee/internal/core"
)

// DAG represents a directed acyclic graph of plugin execution steps
type DAG struct {
	steps      []*core.ExecutionStep
	completed  map[string]bool
	inProgress map[string]bool
	mu         sync.RWMutex
}

// NewDAG creates a new DAG from execution steps
func NewDAG(steps []core.ExecutionStep) *DAG {
	stepsCopy := make([]*core.ExecutionStep, len(steps))
	for i := range steps {
		stepsCopy[i] = &steps[i]
	}

	return &DAG{
		steps:      stepsCopy,
		completed:  make(map[string]bool),
		inProgress: make(map[string]bool),
	}
}

// GetReadySteps returns steps that are ready to execute
// A step is ready if all its dependencies have been completed
func (d *DAG) GetReadySteps() []*core.ExecutionStep {
	d.mu.Lock()
	defer d.mu.Unlock()

	ready := make([]*core.ExecutionStep, 0)

	for _, step := range d.steps {
		// Skip if already completed or in progress
		if d.completed[step.PluginID] || d.inProgress[step.PluginID] {
			continue
		}

		// Check if all dependencies are completed
		if d.areDependenciesCompleted(step) {
			ready = append(ready, step)
			d.inProgress[step.PluginID] = true
		}
	}

	return ready
}

// MarkCompleted marks a step as completed
func (d *DAG) MarkCompleted(pluginID string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.completed[pluginID] = true
	delete(d.inProgress, pluginID)
}

// areDependenciesCompleted checks if all dependencies of a step are completed
func (d *DAG) areDependenciesCompleted(step *core.ExecutionStep) bool {
	for _, depID := range step.DependsOn {
		if !d.completed[depID] {
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
	for id := range d.completed {
		completed = append(completed, id)
	}
	return completed
}

// GetInProgress returns the list of in-progress plugin IDs
func (d *DAG) GetInProgress() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	inProgress := make([]string, 0, len(d.inProgress))
	for id := range d.inProgress {
		inProgress = append(inProgress, id)
	}
	return inProgress
}

// GetPending returns the list of pending plugin IDs
func (d *DAG) GetPending() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	pending := make([]string, 0)
	for _, step := range d.steps {
		if !d.completed[step.PluginID] && !d.inProgress[step.PluginID] {
			pending = append(pending, step.PluginID)
		}
	}
	return pending
}
