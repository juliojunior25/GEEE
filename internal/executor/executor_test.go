package executor

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/internal/registry"
	"github.com/yourusername/geee/pkg/types"
)

// mockLogger implements types.Logger for testing
type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields map[string]interface{})  {}
func (m *mockLogger) Error(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Debug(msg string, fields map[string]interface{}) {}
func (m *mockLogger) Warn(msg string, fields map[string]interface{})  {}

// mockMetrics implements types.MetricsCollector for testing
type mockMetrics struct{}

func (m *mockMetrics) IncrementCounter(name string, labels map[string]string) {}
func (m *mockMetrics) RecordDuration(name string, duration time.Duration, labels map[string]string) {
}
func (m *mockMetrics) SetGauge(name string, value float64, labels map[string]string) {}

// mockPlugin implements core.Plugin for testing
type mockPlugin struct {
	id       string
	executed bool
	result   *types.PluginResult
	err      error
}

func (m *mockPlugin) Manifest() types.PluginManifest {
	return types.PluginManifest{
		ID:      m.id,
		Name:    "Mock Plugin",
		Version: "1.0.0",
	}
}

func (m *mockPlugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	m.executed = true
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &types.PluginResult{
		Success: true,
		Data:    map[string]interface{}{"output": "test"},
	}, nil
}

func (m *mockPlugin) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockPlugin) Validate(ctx *types.ExecutionContext) error {
	return nil
}

func TestNewDefaultExecutor(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	executor := NewDefaultExecutor(reg, logger, metrics)
	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}
}

func TestExecutePlugin(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	plugin := &mockPlugin{id: "test-plugin"}
	reg.Register(plugin)

	executor := NewDefaultExecutor(reg, logger, metrics)

	result, err := executor.ExecutePlugin(
		context.Background(),
		"test-plugin",
		map[string]interface{}{"input": "test"},
		map[string]interface{}{},
	)

	if err != nil {
		t.Fatalf("ExecutePlugin failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected successful execution")
	}

	if !plugin.executed {
		t.Fatal("Plugin was not executed")
	}
}

func TestExecutePluginNotFound(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	executor := NewDefaultExecutor(reg, logger, metrics)

	_, err := executor.ExecutePlugin(
		context.Background(),
		"nonexistent",
		map[string]interface{}{},
		map[string]interface{}{},
	)

	if err == nil {
		t.Fatal("Expected error when executing nonexistent plugin")
	}
}

func TestExecute_SimplePipeline(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	plugin1 := &mockPlugin{
		id: "plugin1",
		result: &types.PluginResult{
			Success: true,
			Data:    map[string]interface{}{"step1": "done"},
		},
	}
	plugin2 := &mockPlugin{
		id: "plugin2",
		result: &types.PluginResult{
			Success: true,
			Data:    map[string]interface{}{"step2": "done"},
		},
	}

	reg.Register(plugin1)
	reg.Register(plugin2)

	executor := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name:        "test-pipeline",
		Description: "Test pipeline",
		Steps: []core.ExecutionStep{
			{PluginID: "plugin1", Config: map[string]interface{}{}},
			{PluginID: "plugin2", Config: map[string]interface{}{}},
		},
	}

	result, err := executor.Execute(context.Background(), plan, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected successful execution")
	}

	if !plugin1.executed || !plugin2.executed {
		t.Fatal("Not all plugins were executed")
	}
}

func TestExecute_WithDependencies(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	plugin1 := &mockPlugin{id: "plugin1"}
	plugin2 := &mockPlugin{id: "plugin2"}

	reg.Register(plugin1)
	reg.Register(plugin2)

	executor := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name: "test-pipeline",
		Steps: []core.ExecutionStep{
			{PluginID: "plugin1", Config: map[string]interface{}{}},
			{
				PluginID:  "plugin2",
				Config:    map[string]interface{}{},
				DependsOn: []string{"plugin1"},
			},
		},
	}

	result, err := executor.Execute(context.Background(), plan, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected successful execution")
	}
}

func TestExecute_OptionalPlugin(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	plugin1 := &mockPlugin{
		id:  "plugin1",
		err: nil,
	}
	plugin2 := &mockPlugin{
		id:  "plugin2",
		err: fmt.Errorf("test error"),
	}
	plugin3 := &mockPlugin{id: "plugin3"}

	reg.Register(plugin1)
	reg.Register(plugin2)
	reg.Register(plugin3)

	executor := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name:        "test-pipeline",
		StopOnError: true,
		Steps: []core.ExecutionStep{
			{PluginID: "plugin1", Config: map[string]interface{}{}},
			{
				PluginID: "plugin2",
				Config:   map[string]interface{}{},
				Optional: true, // This plugin can fail
			},
			{PluginID: "plugin3", Config: map[string]interface{}{}},
		},
	}

	result, err := executor.Execute(context.Background(), plan, map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Fatal("Expected successful execution despite optional plugin failure")
	}

	// All plugins should have been executed
	if !plugin1.executed || !plugin2.executed || !plugin3.executed {
		t.Fatal("Not all plugins were executed")
	}
}

func TestExecute_InvalidPlan(t *testing.T) {
	reg := registry.NewDefaultRegistry()
	logger := &mockLogger{}
	metrics := &mockMetrics{}

	executor := NewDefaultExecutor(reg, logger, metrics)

	plan := &core.ExecutionPlan{
		Name:  "test-pipeline",
		Steps: []core.ExecutionStep{},
	}

	_, err := executor.Execute(context.Background(), plan, map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected error with empty plan")
	}
}
