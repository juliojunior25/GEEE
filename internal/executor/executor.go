package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/types"
)

// DefaultExecutor implements the Executor interface
type DefaultExecutor struct {
	registry core.PluginRegistry
	logger   types.Logger
	metrics  types.MetricsCollector
}

// NewDefaultExecutor creates a new default executor
func NewDefaultExecutor(registry core.PluginRegistry, logger types.Logger, metrics types.MetricsCollector) *DefaultExecutor {
	return &DefaultExecutor{
		registry: registry,
		logger:   logger,
		metrics:  metrics,
	}
}

// Execute runs a pipeline of plugins according to the execution plan
func (e *DefaultExecutor) Execute(ctx context.Context, plan *core.ExecutionPlan, initialState map[string]interface{}) (*types.PluginResult, error) {
	// Validate execution plan
	if err := core.ValidateExecutionPlan(plan, e.registry); err != nil {
		return nil, err
	}

	// Create execution context
	execID := uuid.New().String()
	startTime := time.Now()

	e.logger.Info("Starting pipeline execution", map[string]interface{}{
		"execution_id": execID,
		"plan_name":    plan.Name,
		"steps":        len(plan.Steps),
	})

	// Setup context with timeout
	execCtx := ctx
	if plan.Timeout > 0 {
		var cancel context.CancelFunc
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(plan.Timeout)*time.Second)
		defer cancel()
	}

	// Track active executions
	e.metrics.SetGauge("active_executions", float64(1), map[string]string{})
	defer e.metrics.SetGauge("active_executions", float64(0), map[string]string{})

	// Build dependency graph
	dag := NewDAG(plan.Steps)

	// Execute DAG
	state := initialState
	var finalResult *types.PluginResult

	for {
		// Get ready steps (steps with all dependencies completed)
		readySteps := dag.GetReadySteps()
		if len(readySteps) == 0 {
			break
		}

		// Execute ready steps in parallel
		results := e.executeStepsInParallel(execCtx, execID, readySteps, state)

		// Process results
		for i, result := range results {
			step := readySteps[i]

			if result.err != nil {
				// Handle error
				if !step.Optional && plan.StopOnError {
					e.logger.Error("Pipeline execution failed", map[string]interface{}{
						"execution_id": execID,
						"plugin_id":    step.PluginID,
						"error":        result.err.Error(),
					})

					return nil, result.err
				}

				// Log error but continue
				e.logger.Warn("Optional plugin failed", map[string]interface{}{
					"execution_id": execID,
					"plugin_id":    step.PluginID,
					"error":        result.err.Error(),
				})

				e.metrics.IncrementCounter("geee_plugin_errors_total", map[string]string{
					"plugin_id": step.PluginID,
				})
			} else {
				// Replace state with result data (plugin returns complete state)
				if result.result != nil && result.result.Data != nil {
					state = result.result.Data
					finalResult = result.result
				}

				e.logger.Info("Plugin execution completed", map[string]interface{}{
					"execution_id": execID,
					"plugin_id":    step.PluginID,
					"duration_ms":  result.durationMs,
				})

				e.metrics.RecordDuration("geee_plugin_execution_duration_seconds", time.Duration(result.durationMs)*time.Millisecond, map[string]string{
					"plugin_id": step.PluginID,
				})
			}

			// Mark step as completed
			dag.MarkCompleted(step)
		}
	}

	// Record pipeline duration
	duration := time.Since(startTime)
	e.metrics.RecordDuration("geee_execution_duration_seconds", duration, map[string]string{
		"plan_name": plan.Name,
	})

	e.logger.Info("Pipeline execution completed", map[string]interface{}{
		"execution_id": execID,
		"duration_ms":  duration.Milliseconds(),
	})

	if finalResult == nil {
		finalResult = &types.PluginResult{
			Success: true,
			Data:    state,
		}
	}

	// Always set metadata
	finalResult.Metadata = &types.ExecutionMetadata{
		ExecutionID: execID,
		StartTime:   startTime,
		EndTime:     time.Now(),
		DurationMs:  duration.Milliseconds(),
	}

	return finalResult, nil
}

// ExecutePlugin runs a single plugin with the given context
func (e *DefaultExecutor) ExecutePlugin(ctx context.Context, pluginID string, state map[string]interface{}, config map[string]interface{}) (*types.PluginResult, error) {
	plugin, err := e.registry.Get(pluginID)
	if err != nil {
		return nil, err
	}

	execID := uuid.New().String()
	execCtx := &types.ExecutionContext{
		Context:     ctx,
		ExecutionID: execID,
		State:       state,
		Config:      config,
		Logger:      e.logger,
		Metrics:     e.metrics,
	}

	return e.executePluginWithCleanup(execCtx, plugin)
}

// stepResult holds the result of a step execution
type stepResult struct {
	result     *types.PluginResult
	err        error
	durationMs int64
}

// executeStepsInParallel executes multiple steps in parallel
func (e *DefaultExecutor) executeStepsInParallel(ctx context.Context, execID string, steps []*core.ExecutionStep, state map[string]interface{}) []stepResult {
	results := make([]stepResult, len(steps))
	var wg sync.WaitGroup

	for i, step := range steps {
		wg.Add(1)
		go func(idx int, s *core.ExecutionStep) {
			defer wg.Done()

			plugin, err := e.registry.Get(s.PluginID)
			if err != nil {
				results[idx] = stepResult{err: err}
				return
			}

			// Create execution context
			execCtx := &types.ExecutionContext{
				Context:     ctx,
				ExecutionID: execID,
				State:       state,
				Config:      s.Config,
				Logger:      e.logger,
				Metrics:     e.metrics,
			}

			// Apply step timeout if specified
			if s.Timeout > 0 {
				var cancel context.CancelFunc
				execCtx.Context, cancel = context.WithTimeout(ctx, time.Duration(s.Timeout)*time.Second)
				defer cancel()
			}

			startTime := time.Now()
			result, err := e.executePluginWithCleanup(execCtx, plugin)
			durationMs := time.Since(startTime).Milliseconds()

			results[idx] = stepResult{
				result:     result,
				err:        err,
				durationMs: durationMs,
			}
		}(i, step)
	}

	wg.Wait()
	return results
}

// executePluginWithCleanup executes a plugin and ensures cleanup is called
func (e *DefaultExecutor) executePluginWithCleanup(execCtx *types.ExecutionContext, plugin core.Plugin) (result *types.PluginResult, err error) {
	// Ensure cleanup is called even if Execute panics
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("plugin panicked: %v", r)
		}

		if cleanupErr := plugin.Cleanup(execCtx.Context); cleanupErr != nil {
			e.logger.Warn("Plugin cleanup failed", map[string]interface{}{
				"plugin_id": plugin.Manifest().ID,
				"error":     cleanupErr.Error(),
			})
		}
	}()

	// Validate plugin
	if err := plugin.Validate(execCtx); err != nil {
		return nil, err
	}

	// Execute plugin
	result, err = plugin.Execute(execCtx)
	return result, err
}
