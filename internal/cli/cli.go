package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yourusername/geee/internal/config"
	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/internal/executor"
	"github.com/yourusername/geee/internal/observability"
	"github.com/yourusername/geee/internal/registry"
	"github.com/yourusername/geee/pkg/types"
	"github.com/yourusername/geee/plugins"
	"gopkg.in/yaml.v3"
)

// CLI represents the command-line interface
type CLI struct {
	registry core.PluginRegistry
	executor core.Executor
	logger   types.Logger
	metrics  types.MetricsCollector
}

// NewCLI creates a new CLI instance
func NewCLI(verbose bool) *CLI {
	logger := observability.NewDefaultLogger(verbose)
	metrics := observability.NewNoOpMetricsCollector()
	reg := registry.NewDefaultRegistry()

	// Register all available plugins
	if err := plugins.RegisterAll(reg); err != nil {
		logger.Error("Failed to register plugins", map[string]interface{}{
			"error": err.Error(),
		})
	}

	exec := executor.NewDefaultExecutor(reg, logger, metrics)

	return &CLI{
		registry: reg,
		executor: exec,
		logger:   logger,
		metrics:  metrics,
	}
}

// Run executes a pipeline with the given configuration
func (c *CLI) Run(configPath, inputPath, outputPath string, verbose bool) error {
	// Load configuration
	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	c.logger.Info("Loaded configuration", map[string]interface{}{
		"config_path": configPath,
		"plan_name":   cfg.Name,
		"plugins":     len(cfg.Plugins),
	})

	// Load input data
	inputData, err := c.loadInput(inputPath)
	if err != nil {
		return fmt.Errorf("failed to load input: %w", err)
	}

	c.logger.Debug("Loaded input data", map[string]interface{}{
		"input_path": inputPath,
		"data_size":  len(inputData),
	})

	// Convert config to execution plan
	plan := cfg.ToExecutionPlan()

	// Execute pipeline
	c.logger.Info("Starting pipeline execution", map[string]interface{}{
		"plan_name": plan.Name,
	})

	result, err := c.executor.Execute(context.Background(), plan, inputData)
	if err != nil {
		c.logger.Error("Pipeline execution failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("pipeline execution failed: %w", err)
	}

	c.logger.Info("Pipeline execution completed successfully", map[string]interface{}{
		"duration_ms": result.Metadata.DurationMs,
	})

	// Save output
	if err := c.saveOutput(outputPath, result); err != nil {
		return fmt.Errorf("failed to save output: %w", err)
	}

	c.logger.Info("Output saved successfully", map[string]interface{}{
		"output_path": outputPath,
	})

	return nil
}

// ListPlugins lists all registered plugins
func (c *CLI) ListPlugins() error {
	pluginIDs := c.registry.List()

	if len(pluginIDs) == 0 {
		fmt.Println("No plugins registered")
		return nil
	}

	fmt.Printf("Registered plugins (%d):\n", len(pluginIDs))
	for _, id := range pluginIDs {
		plugin, err := c.registry.Get(id)
		if err != nil {
			continue
		}

		manifest := plugin.Manifest()
		fmt.Printf("  - %s (v%s): %s\n", manifest.ID, manifest.Version, manifest.Description)
	}

	return nil
}

// loadInput loads input data from a file
func (c *CLI) loadInput(path string) (map[string]interface{}, error) {
	if path == "" {
		// Return empty state if no input specified
		return make(map[string]interface{}), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read input file: %w", err)
	}

	var input map[string]interface{}

	// Determine format by extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("failed to parse JSON input: %w", err)
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &input); err != nil {
			return nil, fmt.Errorf("failed to parse YAML input: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported input format: %s (supported: .json, .yaml, .yml)", ext)
	}

	return input, nil
}

// saveOutput saves the result to a file
func (c *CLI) saveOutput(path string, result *types.PluginResult) error {
	if path == "" {
		// Print to stdout if no output path specified
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		fmt.Println(string(data))
		return nil
	}

	// Determine format by extension
	ext := strings.ToLower(filepath.Ext(path))
	var data []byte
	var err error

	switch ext {
	case ".json":
		data, err = json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON output: %w", err)
		}
	case ".yaml", ".yml":
		data, err = yaml.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML output: %w", err)
		}
	default:
		return fmt.Errorf("unsupported output format: %s (supported: .json, .yaml, .yml)", ext)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
