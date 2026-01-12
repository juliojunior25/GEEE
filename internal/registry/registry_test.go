package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/types"
)

// mockPlugin is a mock plugin for testing
type mockPlugin struct {
	id string
}

func (m *mockPlugin) Manifest() types.PluginManifest {
	return types.PluginManifest{
		ID:      m.id,
		Name:    "Mock Plugin",
		Version: "1.0.0",
	}
}

func (m *mockPlugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	return &types.PluginResult{Success: true}, nil
}

func (m *mockPlugin) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockPlugin) Validate(ctx *types.ExecutionContext) error {
	return nil
}

func TestNewDefaultRegistry(t *testing.T) {
	registry := NewDefaultRegistry()
	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}
	if registry.plugins == nil {
		t.Fatal("Expected plugins map to be initialized")
	}
	if len(registry.plugins) != 0 {
		t.Fatalf("Expected empty registry, got %d plugins", len(registry.plugins))
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		plugin    core.Plugin
		wantError bool
	}{
		{
			name:      "valid plugin",
			plugin:    &mockPlugin{id: "test-plugin"},
			wantError: false,
		},
		{
			name:      "nil plugin",
			plugin:    nil,
			wantError: true,
		},
		{
			name:      "plugin with empty ID",
			plugin:    &mockPlugin{id: ""},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewDefaultRegistry()
			err := registry.Register(tt.plugin)
			if (err != nil) != tt.wantError {
				t.Errorf("Register() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestRegisterDuplicate(t *testing.T) {
	registry := NewDefaultRegistry()
	plugin := &mockPlugin{id: "test-plugin"}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("First Register() failed: %v", err)
	}

	err = registry.Register(plugin)
	if err == nil {
		t.Fatal("Expected error when registering duplicate plugin")
	}
}

func TestGet(t *testing.T) {
	registry := NewDefaultRegistry()
	plugin := &mockPlugin{id: "test-plugin"}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	retrieved, err := registry.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}
	if retrieved != plugin {
		t.Fatal("Retrieved plugin does not match registered plugin")
	}
}

func TestGetNotFound(t *testing.T) {
	registry := NewDefaultRegistry()

	_, err := registry.Get("nonexistent")
	if err == nil {
		t.Fatal("Expected error when getting nonexistent plugin")
	}
}

func TestList(t *testing.T) {
	registry := NewDefaultRegistry()

	plugins := []core.Plugin{
		&mockPlugin{id: "plugin1"},
		&mockPlugin{id: "plugin2"},
		&mockPlugin{id: "plugin3"},
	}

	for _, p := range plugins {
		if err := registry.Register(p); err != nil {
			t.Fatalf("Register() failed: %v", err)
		}
	}

	list := registry.List()
	if len(list) != 3 {
		t.Fatalf("Expected 3 plugins, got %d", len(list))
	}

	// Check that all plugin IDs are present
	idMap := make(map[string]bool)
	for _, id := range list {
		idMap[id] = true
	}

	for _, p := range plugins {
		id := p.Manifest().ID
		if !idMap[id] {
			t.Errorf("Plugin ID %s not found in list", id)
		}
	}
}

func TestListEmpty(t *testing.T) {
	registry := NewDefaultRegistry()

	list := registry.List()
	if len(list) != 0 {
		t.Fatalf("Expected empty list, got %d items", len(list))
	}
}

func TestHas(t *testing.T) {
	registry := NewDefaultRegistry()
	plugin := &mockPlugin{id: "test-plugin"}

	if registry.Has("test-plugin") {
		t.Fatal("Has() returned true for nonexistent plugin")
	}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	if !registry.Has("test-plugin") {
		t.Fatal("Has() returned false for registered plugin")
	}
}

func TestUnregister(t *testing.T) {
	registry := NewDefaultRegistry()
	plugin := &mockPlugin{id: "test-plugin"}

	err := registry.Register(plugin)
	if err != nil {
		t.Fatalf("Register() failed: %v", err)
	}

	err = registry.Unregister("test-plugin")
	if err != nil {
		t.Fatalf("Unregister() failed: %v", err)
	}

	if registry.Has("test-plugin") {
		t.Fatal("Plugin still exists after unregister")
	}
}

func TestUnregisterNotFound(t *testing.T) {
	registry := NewDefaultRegistry()

	err := registry.Unregister("nonexistent")
	if err == nil {
		t.Fatal("Expected error when unregistering nonexistent plugin")
	}
}

func TestConcurrentAccess(t *testing.T) {
	registry := NewDefaultRegistry()

	done := make(chan bool)

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		go func(id int) {
			plugin := &mockPlugin{id: fmt.Sprintf("plugin-%d", id)}
			registry.Register(plugin)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			registry.List()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
