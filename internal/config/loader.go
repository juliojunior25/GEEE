package config

import (
	"fmt"
	"os"

	"github.com/yourusername/geee/internal/core"
	"github.com/yourusername/geee/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	// Name is the pipeline name
	Name string `yaml:"name"`

	// Description describes what this pipeline does
	Description string `yaml:"description"`

	// Timeout is the maximum execution time in seconds
	Timeout int `yaml:"timeout,omitempty"`

	// StopOnError determines if execution stops on first error
	StopOnError bool `yaml:"stop_on_error"`

	// Plugins contains the list of plugin execution steps
	Plugins []PluginConfig `yaml:"plugins"`
}

// PluginConfig represents a single plugin configuration in the pipeline
type PluginConfig struct {
	// ID is the plugin identifier
	ID string `yaml:"id"`

	// Config contains plugin-specific configuration
	Config map[string]interface{} `yaml:"config,omitempty"`

	// DependsOn lists the plugin IDs that must complete before this step
	DependsOn []string `yaml:"depends_on,omitempty"`

	// Optional indicates if this step's failure should not abort the pipeline
	Optional bool `yaml:"optional,omitempty"`

	// Timeout overrides the plugin's default timeout
	Timeout int `yaml:"timeout,omitempty"`
}

// LoadFromFile loads a configuration from a YAML file
func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.NewConfigLoadError(
			fmt.Sprintf("Failed to read config file: %s", path),
			"Ensure the file exists and you have read permissions",
			err,
		)
	}

	return LoadFromBytes(data)
}

// LoadFromBytes loads a configuration from YAML bytes
func LoadFromBytes(data []byte) (*Config, error) {
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.NewConfigLoadError(
			"Failed to parse YAML configuration",
			"Check YAML syntax and structure",
			err,
		)
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ValidateConfig validates a configuration
func ValidateConfig(config *Config) error {
	if config == nil {
		return errors.NewInvalidConfigError(
			"Configuration cannot be nil",
			"$",
			"Provide a valid configuration",
		)
	}

	if config.Name == "" {
		return errors.NewInvalidConfigError(
			"Pipeline name cannot be empty",
			"$.name",
			"Add a name field to your configuration",
		)
	}

	if len(config.Plugins) == 0 {
		return errors.NewInvalidConfigError(
			"Pipeline must have at least one plugin",
			"$.plugins",
			"Add at least one plugin to the plugins array",
		)
	}

	// Validate each plugin
	for i, plugin := range config.Plugins {
		if plugin.ID == "" {
			return errors.NewInvalidConfigError(
				"Plugin ID cannot be empty",
				fmt.Sprintf("$.plugins[%d].id", i),
				"Specify a valid plugin ID",
			)
		}
	}

	return nil
}

// ToExecutionPlan converts a Config to an ExecutionPlan
func (c *Config) ToExecutionPlan() *core.ExecutionPlan {
	steps := make([]core.ExecutionStep, len(c.Plugins))
	for i, plugin := range c.Plugins {
		steps[i] = core.ExecutionStep{
			PluginID:  plugin.ID,
			Config:    plugin.Config,
			DependsOn: plugin.DependsOn,
			Optional:  plugin.Optional,
			Timeout:   plugin.Timeout,
		}
	}

	return &core.ExecutionPlan{
		Name:        c.Name,
		Description: c.Description,
		Steps:       steps,
		Timeout:     c.Timeout,
		StopOnError: c.StopOnError,
	}
}
