package base

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yourusername/geee/pkg/errors"
	"github.com/yourusername/geee/pkg/types"
)

// BasePlugin provides common functionality for all plugins
// Plugins should embed this struct and override the ExecuteImpl method
type BasePlugin struct {
	manifest types.PluginManifest

	// cleanupFuncs stores cleanup functions to be called during Cleanup
	cleanupFuncs []func(ctx context.Context) error
	cleanupMu    sync.Mutex

	// resourceTracker tracks resource usage
	resourceTracker *ResourceTracker
}

// NewBasePlugin creates a new base plugin with the given manifest
func NewBasePlugin(manifest types.PluginManifest) *BasePlugin {
	return &BasePlugin{
		manifest:        manifest,
		cleanupFuncs:    make([]func(ctx context.Context) error, 0),
		resourceTracker: NewResourceTracker(manifest.MaxMemoryMB, manifest.Timeout),
	}
}

// Manifest returns the plugin's manifest
func (b *BasePlugin) Manifest() types.PluginManifest {
	return b.manifest
}

// RegisterCleanup adds a cleanup function to be called during Cleanup
// This allows plugins to register resources that need cleanup
func (b *BasePlugin) RegisterCleanup(fn func(ctx context.Context) error) {
	b.cleanupMu.Lock()
	defer b.cleanupMu.Unlock()
	b.cleanupFuncs = append(b.cleanupFuncs, fn)
}

// Cleanup calls all registered cleanup functions
// This is guaranteed to be called even if Execute panics
func (b *BasePlugin) Cleanup(ctx context.Context) error {
	b.cleanupMu.Lock()
	funcs := make([]func(ctx context.Context) error, len(b.cleanupFuncs))
	copy(funcs, b.cleanupFuncs)
	b.cleanupMu.Unlock()

	var errs []error
	for i := len(funcs) - 1; i >= 0; i-- {
		if err := funcs[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// Validate validates the execution context against the plugin's manifest
func (b *BasePlugin) Validate(ctx *types.ExecutionContext) error {
	if ctx == nil {
		return errors.NewValidationError(
			"Execution context cannot be nil",
			"$",
			"Provide a valid execution context",
		)
	}

	// Validate config schema if defined
	if len(b.manifest.ConfigSchema) > 0 {
		if err := ValidateAgainstSchema(ctx.Config, b.manifest.ConfigSchema); err != nil {
			return errors.NewValidationError(
				fmt.Sprintf("Config validation failed: %v", err),
				"$.config",
				"Ensure config matches the plugin's config schema",
			)
		}
	}

	// Validate input schema if defined
	if len(b.manifest.InputSchema) > 0 {
		if err := ValidateAgainstSchema(ctx.State, b.manifest.InputSchema); err != nil {
			return errors.NewValidationError(
				fmt.Sprintf("Input validation failed: %v", err),
				"$.input",
				"Ensure input matches the plugin's input schema",
			)
		}
	}

	return nil
}

// ExecuteWithTimeout executes a function with timeout and resource tracking
func (b *BasePlugin) ExecuteWithTimeout(ctx *types.ExecutionContext, fn func() (*types.PluginResult, error)) (*types.PluginResult, error) {
	// Start resource tracking
	b.resourceTracker.Start()
	defer b.resourceTracker.Stop()

	// Create a context with timeout
	timeout := b.getTimeout(ctx)
	execCtx, cancel := context.WithTimeout(ctx.Context, timeout)
	defer cancel()

	// Execute in a goroutine to handle timeout
	resultChan := make(chan *types.PluginResult, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := fn()
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-execCtx.Done():
		return nil, errors.NewExecutionError(
			b.manifest.ID,
			fmt.Sprintf("Plugin execution exceeded timeout of %v", timeout),
			"Increase the plugin timeout or optimize the plugin logic",
		)
	}
}

// getTimeout returns the timeout for this execution
func (b *BasePlugin) getTimeout(ctx *types.ExecutionContext) time.Duration {
	// Check if timeout is specified in config
	if timeout, ok := ctx.Config["timeout"].(int); ok && timeout > 0 {
		return time.Duration(timeout) * time.Second
	}

	// Use manifest timeout
	if b.manifest.Timeout > 0 {
		return time.Duration(b.manifest.Timeout) * time.Second
	}

	// Default timeout
	return 30 * time.Second
}

// CheckResourceLimits checks if resource limits are exceeded
func (b *BasePlugin) CheckResourceLimits() error {
	return b.resourceTracker.CheckLimits()
}
