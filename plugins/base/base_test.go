package base

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/geee/pkg/types"
)

func TestNewBasePlugin(t *testing.T) {
	manifest := types.PluginManifest{
		ID:      "test-plugin",
		Name:    "Test Plugin",
		Version: "1.0.0",
	}

	plugin := NewBasePlugin(manifest)
	if plugin == nil {
		t.Fatal("Expected non-nil plugin")
	}

	if plugin.manifest.ID != "test-plugin" {
		t.Errorf("Expected ID 'test-plugin', got '%s'", plugin.manifest.ID)
	}
}

func TestManifest(t *testing.T) {
	manifest := types.PluginManifest{
		ID:      "test-plugin",
		Name:    "Test Plugin",
		Version: "1.0.0",
	}

	plugin := NewBasePlugin(manifest)
	retrievedManifest := plugin.Manifest()

	if retrievedManifest.ID != manifest.ID {
		t.Errorf("Expected ID '%s', got '%s'", manifest.ID, retrievedManifest.ID)
	}
}

func TestRegisterCleanup(t *testing.T) {
	manifest := types.PluginManifest{
		ID: "test-plugin",
	}

	plugin := NewBasePlugin(manifest)

	called := false
	cleanupFn := func(ctx context.Context) error {
		called = true
		return nil
	}

	plugin.RegisterCleanup(cleanupFn)

	err := plugin.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if !called {
		t.Fatal("Cleanup function was not called")
	}
}

func TestCleanupMultipleFunctions(t *testing.T) {
	manifest := types.PluginManifest{
		ID: "test-plugin",
	}

	plugin := NewBasePlugin(manifest)

	callOrder := make([]int, 0)

	plugin.RegisterCleanup(func(ctx context.Context) error {
		callOrder = append(callOrder, 1)
		return nil
	})

	plugin.RegisterCleanup(func(ctx context.Context) error {
		callOrder = append(callOrder, 2)
		return nil
	})

	plugin.RegisterCleanup(func(ctx context.Context) error {
		callOrder = append(callOrder, 3)
		return nil
	})

	err := plugin.Cleanup(context.Background())
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	if len(callOrder) != 3 {
		t.Fatalf("Expected 3 cleanup calls, got %d", len(callOrder))
	}

	// Cleanup should be called in reverse order (LIFO)
	if callOrder[0] != 3 || callOrder[1] != 2 || callOrder[2] != 1 {
		t.Errorf("Expected cleanup order [3, 2, 1], got %v", callOrder)
	}
}

func TestValidateNilContext(t *testing.T) {
	manifest := types.PluginManifest{
		ID: "test-plugin",
	}

	plugin := NewBasePlugin(manifest)

	err := plugin.Validate(nil)
	if err == nil {
		t.Fatal("Expected error when validating nil context")
	}
}

func TestValidateWithConfigSchema(t *testing.T) {
	manifest := types.PluginManifest{
		ID: "test-plugin",
		ConfigSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"timeout": map[string]interface{}{
					"type": "number",
				},
			},
			"required": []interface{}{"timeout"},
		},
	}

	plugin := NewBasePlugin(manifest)

	tests := []struct {
		name      string
		config    map[string]interface{}
		wantError bool
	}{
		{
			name:      "valid config",
			config:    map[string]interface{}{"timeout": 30},
			wantError: false,
		},
		{
			name:      "missing required field",
			config:    map[string]interface{}{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &types.ExecutionContext{
				Context: context.Background(),
				Config:  tt.config,
				State:   map[string]interface{}{},
			}

			err := plugin.Validate(ctx)
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	tests := []struct {
		name            string
		manifestTimeout int
		configTimeout   interface{}
		expected        time.Duration
	}{
		{
			name:            "use manifest timeout",
			manifestTimeout: 60,
			configTimeout:   nil,
			expected:        60 * time.Second,
		},
		{
			name:            "use config timeout",
			manifestTimeout: 60,
			configTimeout:   10,
			expected:        10 * time.Second,
		},
		{
			name:            "use default timeout",
			manifestTimeout: 0,
			configTimeout:   nil,
			expected:        30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := types.PluginManifest{
				ID:      "test-plugin",
				Timeout: tt.manifestTimeout,
			}

			plugin := NewBasePlugin(manifest)

			config := map[string]interface{}{}
			if tt.configTimeout != nil {
				config["timeout"] = tt.configTimeout
			}

			ctx := &types.ExecutionContext{
				Context: context.Background(),
				Config:  config,
			}

			timeout := plugin.getTimeout(ctx)
			if timeout != tt.expected {
				t.Errorf("Expected timeout %v, got %v", tt.expected, timeout)
			}
		})
	}
}
