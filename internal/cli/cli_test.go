package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/types"
)

// mockPlugin is a simple plugin for testing
type mockPlugin struct {
	id      string
	execErr error
}

func (m *mockPlugin) Manifest() types.PluginManifest {
	return types.PluginManifest{
		ID:          m.id,
		Name:        "Mock Plugin",
		Version:     "1.0.0",
		Description: "A mock plugin for testing",
	}
}

func (m *mockPlugin) Execute(ctx *types.ExecutionContext) (*types.PluginResult, error) {
	if m.execErr != nil {
		return nil, m.execErr
	}

	return &types.PluginResult{
		Success: true,
		Data:    ctx.State,
	}, nil
}

func (m *mockPlugin) Cleanup(ctx context.Context) error {
	return nil
}

func (m *mockPlugin) Validate(ctx *types.ExecutionContext) error {
	return nil
}

func TestCLI_LoadInput_JSON(t *testing.T) {
	c := NewCLI(false)

	// Create temporary JSON file
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.json")
	inputData := `{"key": "value", "number": 42}`

	if err := os.WriteFile(inputPath, []byte(inputData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load input
	data, err := c.loadInput(inputPath)
	if err != nil {
		t.Fatalf("Failed to load input: %v", err)
	}

	if data["key"] != "value" {
		t.Errorf("Expected key='value', got '%v'", data["key"])
	}

	if data["number"].(float64) != 42 {
		t.Errorf("Expected number=42, got '%v'", data["number"])
	}
}

func TestCLI_LoadInput_YAML(t *testing.T) {
	c := NewCLI(false)

	// Create temporary YAML file
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.yaml")
	inputData := `key: value
number: 42`

	if err := os.WriteFile(inputPath, []byte(inputData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load input
	data, err := c.loadInput(inputPath)
	if err != nil {
		t.Fatalf("Failed to load input: %v", err)
	}

	if data["key"] != "value" {
		t.Errorf("Expected key='value', got '%v'", data["key"])
	}

	if data["number"].(int) != 42 {
		t.Errorf("Expected number=42, got '%v'", data["number"])
	}
}

func TestCLI_LoadInput_Empty(t *testing.T) {
	c := NewCLI(false)

	// Load with empty path
	data, err := c.loadInput("")
	if err != nil {
		t.Fatalf("Failed to load empty input: %v", err)
	}

	if len(data) != 0 {
		t.Errorf("Expected empty map, got %d elements", len(data))
	}
}

func TestCLI_LoadInput_InvalidJSON(t *testing.T) {
	c := NewCLI(false)

	// Create temporary invalid JSON file
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "invalid.json")
	inputData := `{invalid json}`

	if err := os.WriteFile(inputPath, []byte(inputData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load input - should fail
	_, err := c.loadInput(inputPath)
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCLI_LoadInput_UnsupportedFormat(t *testing.T) {
	c := NewCLI(false)

	// Create temporary file with unsupported format
	tmpDir := t.TempDir()
	inputPath := filepath.Join(tmpDir, "input.txt")
	inputData := `some text`

	if err := os.WriteFile(inputPath, []byte(inputData), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Load input - should fail
	_, err := c.loadInput(inputPath)
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
}

func TestCLI_SaveOutput_JSON(t *testing.T) {
	c := NewCLI(false)

	// Create result
	result := &types.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	// Save to temporary file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.json")

	if err := c.saveOutput(outputPath, result); err != nil {
		t.Fatalf("Failed to save output: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(data) == 0 {
		t.Error("Output file is empty")
	}
}

func TestCLI_SaveOutput_YAML(t *testing.T) {
	c := NewCLI(false)

	// Create result
	result := &types.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	// Save to temporary file
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.yaml")

	if err := c.saveOutput(outputPath, result); err != nil {
		t.Fatalf("Failed to save output: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file was not created")
	}
}

func TestCLI_SaveOutput_UnsupportedFormat(t *testing.T) {
	c := NewCLI(false)

	// Create result
	result := &types.PluginResult{
		Success: true,
		Data: map[string]interface{}{
			"key": "value",
		},
	}

	// Try to save with unsupported format
	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "output.txt")

	err := c.saveOutput(outputPath, result)
	if err == nil {
		t.Error("Expected error for unsupported format, got nil")
	}
}

func TestCLI_ListPlugins_Empty(t *testing.T) {
	c := NewCLI(false)

	// Should not error even with no plugins
	if err := c.ListPlugins(); err != nil {
		t.Errorf("ListPlugins failed: %v", err)
	}
}

func TestCLI_ListPlugins_WithPlugins(t *testing.T) {
	c := NewCLI(false)

	// Register a mock plugin
	plugin := &mockPlugin{id: "test-plugin"}
	if err := c.registry.Register(plugin); err != nil {
		t.Fatalf("Failed to register plugin: %v", err)
	}

	// List plugins
	if err := c.ListPlugins(); err != nil {
		t.Errorf("ListPlugins failed: %v", err)
	}
}

func TestNewCLI(t *testing.T) {
	// Test non-verbose
	c := NewCLI(false)
	if c == nil {
		t.Error("Expected non-nil CLI")
	}

	if c.registry == nil {
		t.Error("Expected non-nil registry")
	}

	if c.executor == nil {
		t.Error("Expected non-nil executor")
	}

	if c.logger == nil {
		t.Error("Expected non-nil logger")
	}

	// Test verbose
	c = NewCLI(true)
	if c == nil {
		t.Error("Expected non-nil CLI")
	}
}

// Ensure mockPlugin implements core.Plugin
var _ core.Plugin = (*mockPlugin)(nil)
