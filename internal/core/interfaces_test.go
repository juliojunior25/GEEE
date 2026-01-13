package core

import (
	"context"
	"testing"

	"github.com/yourusername/geee/pkg/types"
)

// Mock implementations for testing

type mockPlugin struct {
	manifest types.PluginManifest
}

func (m *mockPlugin) Manifest() types.PluginManifest {
	return m.manifest
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

type mockRegistry struct {
	plugins map[string]Plugin
}

func newMockRegistry() *mockRegistry {
	return &mockRegistry{
		plugins: make(map[string]Plugin),
	}
}

func (r *mockRegistry) Register(plugin Plugin) error {
	r.plugins[plugin.Manifest().ID] = plugin
	return nil
}

func (r *mockRegistry) Get(pluginID string) (Plugin, error) {
	plugin, ok := r.plugins[pluginID]
	if !ok {
		return nil, nil
	}
	return plugin, nil
}

func (r *mockRegistry) List() []string {
	ids := make([]string, 0, len(r.plugins))
	for id := range r.plugins {
		ids = append(ids, id)
	}
	return ids
}

func (r *mockRegistry) Has(pluginID string) bool {
	_, ok := r.plugins[pluginID]
	return ok
}

func (r *mockRegistry) Unregister(pluginID string) error {
	delete(r.plugins, pluginID)
	return nil
}

func TestValidateExecutionPlan(t *testing.T) {
	registry := newMockRegistry()

	// Register test plugins
	plugin1 := &mockPlugin{
		manifest: types.PluginManifest{ID: "plugin1", Name: "Plugin 1"},
	}
	plugin2 := &mockPlugin{
		manifest: types.PluginManifest{ID: "plugin2", Name: "Plugin 2"},
	}

	_ = registry.Register(plugin1)
	_ = registry.Register(plugin2)

	tests := []struct {
		name    string
		plan    *ExecutionPlan
		wantErr bool
	}{
		{
			name: "valid plan",
			plan: &ExecutionPlan{
				Name: "test",
				Steps: []ExecutionStep{
					{PluginID: "plugin1"},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil plan",
			plan:    nil,
			wantErr: true,
		},
		{
			name: "empty steps",
			plan: &ExecutionPlan{
				Name:  "test",
				Steps: []ExecutionStep{},
			},
			wantErr: true,
		},
		{
			name: "empty plugin ID",
			plan: &ExecutionPlan{
				Name: "test",
				Steps: []ExecutionStep{
					{PluginID: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "plugin not found",
			plan: &ExecutionPlan{
				Name: "test",
				Steps: []ExecutionStep{
					{PluginID: "nonexistent"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid dependencies",
			plan: &ExecutionPlan{
				Name: "test",
				Steps: []ExecutionStep{
					{PluginID: "plugin1"},
					{PluginID: "plugin2", DependsOn: []string{"plugin1"}},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid dependencies - forward reference",
			plan: &ExecutionPlan{
				Name: "test",
				Steps: []ExecutionStep{
					{PluginID: "plugin1", DependsOn: []string{"plugin2"}},
					{PluginID: "plugin2"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateExecutionPlan(tt.plan, registry)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateExecutionPlan() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutionStep(t *testing.T) {
	step := ExecutionStep{
		PluginID:  "test-plugin",
		Config:    map[string]interface{}{"key": "value"},
		DependsOn: []string{"dep1", "dep2"},
		Optional:  true,
		Timeout:   30,
	}

	if step.PluginID != "test-plugin" {
		t.Errorf("PluginID = %v, want %v", step.PluginID, "test-plugin")
	}
	if !step.Optional {
		t.Errorf("Optional = %v, want %v", step.Optional, true)
	}
	if step.Timeout != 30 {
		t.Errorf("Timeout = %v, want %v", step.Timeout, 30)
	}
	if len(step.DependsOn) != 2 {
		t.Errorf("DependsOn length = %v, want %v", len(step.DependsOn), 2)
	}
}

func TestExecutionPlan(t *testing.T) {
	plan := &ExecutionPlan{
		Name:        "test-plan",
		Description: "Test description",
		Steps: []ExecutionStep{
			{PluginID: "plugin1"},
			{PluginID: "plugin2"},
		},
		Timeout:     60,
		StopOnError: true,
	}

	if plan.Name != "test-plan" {
		t.Errorf("Name = %v, want %v", plan.Name, "test-plan")
	}
	if plan.Description != "Test description" {
		t.Errorf("Description = %v, want %v", plan.Description, "Test description")
	}
	if len(plan.Steps) != 2 {
		t.Errorf("Steps length = %v, want %v", len(plan.Steps), 2)
	}
	if plan.Timeout != 60 {
		t.Errorf("Timeout = %v, want %v", plan.Timeout, 60)
	}
	if !plan.StopOnError {
		t.Errorf("StopOnError = %v, want %v", plan.StopOnError, true)
	}
}
