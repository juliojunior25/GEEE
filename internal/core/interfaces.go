package core

import (
	"context"
	"fmt"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
)

// Plugin defines the interface that all plugins must implement
type Plugin interface {
	// Manifest returns the plugin's manifest describing its capabilities
	Manifest() types.PluginManifest

	// Execute runs the plugin with the given execution context
	// It receives the current state and returns the updated state or an error
	Execute(ctx *types.ExecutionContext) (*types.PluginResult, error)

	// Cleanup is called after execution (success or failure) to release resources
	// This is guaranteed to be called even if Execute panics
	Cleanup(ctx context.Context) error

	// Validate validates the plugin configuration and input schema
	// This is called before Execute to ensure the plugin can run successfully
	Validate(ctx *types.ExecutionContext) error
}

// PluginRegistry manages plugin registration and retrieval
type PluginRegistry interface {
	// Register adds a plugin to the registry
	// Returns an error if a plugin with the same ID already exists
	Register(plugin Plugin) error

	// Get retrieves a plugin by its ID
	// Returns an error if the plugin is not found
	Get(pluginID string) (Plugin, error)

	// List returns all registered plugin IDs
	List() []string

	// Has checks if a plugin with the given ID exists
	Has(pluginID string) bool

	// Unregister removes a plugin from the registry
	// Returns an error if the plugin is not found
	Unregister(pluginID string) error
}

// Executor orchestrates plugin execution according to a plan
type Executor interface {
	// Execute runs a pipeline of plugins according to the execution plan
	// It manages state flow between plugins and handles errors appropriately
	Execute(ctx context.Context, plan *ExecutionPlan, initialState map[string]interface{}) (*types.PluginResult, error)

	// ExecutePlugin runs a single plugin with the given context
	// This is useful for testing individual plugins
	ExecutePlugin(ctx context.Context, pluginID string, state map[string]interface{}, config map[string]interface{}) (*types.PluginResult, error)
}

// ExecutionPlan defines the sequence of plugins to execute
type ExecutionPlan struct {
	// Name is a human-readable name for this plan
	Name string

	// Description describes what this plan does
	Description string

	// Steps defines the ordered list of plugin executions
	Steps []ExecutionStep

	// Timeout is the maximum execution time for the entire plan (in seconds)
	Timeout int

	// StopOnError determines if execution should stop on first error
	// If false, optional plugin errors are logged but execution continues
	StopOnError bool
}

// ExecutionStep defines a single step in the execution plan
type ExecutionStep struct {
	// PluginID is the ID of the plugin to execute
	PluginID string

	// Config contains plugin-specific configuration
	Config map[string]interface{}

	// DependsOn lists the plugin IDs that must complete before this step
	// Empty means this step can run immediately
	DependsOn []string

	// Optional indicates if this step's failure should not abort the pipeline
	Optional bool

	// Timeout overrides the plugin's default timeout (in seconds)
	Timeout int
}

// ValidateExecutionPlan validates an execution plan
func ValidateExecutionPlan(plan *ExecutionPlan, registry PluginRegistry) error {
	if plan == nil {
		return errors.NewValidationError(
			"Execution plan cannot be nil",
			"$",
			"Provide a valid execution plan",
		)
	}

	if len(plan.Steps) == 0 {
		return errors.NewValidationError(
			"Execution plan must have at least one step",
			"$.steps",
			"Add at least one plugin execution step",
		)
	}

	// Validate each step
	seenPlugins := make(map[string]bool)
	for i, step := range plan.Steps {
		if step.PluginID == "" {
			return errors.NewValidationError(
				"Step plugin_id cannot be empty",
				fmt.Sprintf("$.steps[%d].plugin_id", i),
				"Specify a valid plugin ID",
			)
		}

		// Check if plugin exists in registry
		if !registry.Has(step.PluginID) {
			return errors.NewPluginNotFoundError(step.PluginID)
		}

		// Validate dependencies
		for _, depID := range step.DependsOn {
			if !seenPlugins[depID] {
				return errors.NewValidationError(
					fmt.Sprintf("Step depends on plugin '%s' which hasn't been executed yet", depID),
					fmt.Sprintf("$.steps[%d].depends_on", i),
					"Ensure dependencies are listed in execution order",
				)
			}
		}

		seenPlugins[step.PluginID] = true
	}

	return nil
}
