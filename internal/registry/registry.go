package registry

import (
	"fmt"
	"sync"

	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/errors"
)

// DefaultRegistry is the default plugin registry implementation
type DefaultRegistry struct {
	plugins map[string]core.Plugin
	mu      sync.RWMutex
}

// NewDefaultRegistry creates a new default plugin registry
func NewDefaultRegistry() *DefaultRegistry {
	return &DefaultRegistry{
		plugins: make(map[string]core.Plugin),
	}
}

// Register adds a plugin to the registry
// Returns an error if a plugin with the same ID already exists
func (r *DefaultRegistry) Register(plugin core.Plugin) error {
	if plugin == nil {
		return errors.NewValidationError(
			"Cannot register nil plugin",
			"$",
			"Provide a valid plugin instance",
		)
	}

	manifest := plugin.Manifest()
	if manifest.ID == "" {
		return errors.NewValidationError(
			"Plugin manifest must have a non-empty ID",
			"$.manifest.id",
			"Set a unique ID in the plugin manifest",
		)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[manifest.ID]; exists {
		return errors.NewPluginError(
			manifest.ID,
			fmt.Sprintf("Plugin with ID '%s' is already registered", manifest.ID),
			"Use a unique plugin ID or unregister the existing plugin first",
		)
	}

	r.plugins[manifest.ID] = plugin
	return nil
}

// Get retrieves a plugin by its ID
// Returns an error if the plugin is not found
func (r *DefaultRegistry) Get(pluginID string) (core.Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[pluginID]
	if !exists {
		return nil, errors.NewPluginNotFoundError(pluginID)
	}

	return plugin, nil
}

// List returns all registered plugin IDs
func (r *DefaultRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		ids = append(ids, id)
	}

	return ids
}

// Has checks if a plugin with the given ID exists
func (r *DefaultRegistry) Has(pluginID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.plugins[pluginID]
	return exists
}

// Unregister removes a plugin from the registry
// Returns an error if the plugin is not found
func (r *DefaultRegistry) Unregister(pluginID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[pluginID]; !exists {
		return errors.NewPluginNotFoundError(pluginID)
	}

	delete(r.plugins, pluginID)
	return nil
}
