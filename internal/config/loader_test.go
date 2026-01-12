package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromBytes(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "valid config",
			yaml: `
name: test-pipeline
description: Test pipeline
plugins:
  - id: plugin1
    config:
      key: value
`,
			wantErr: false,
		},
		{
			name: "missing name",
			yaml: `
description: Test pipeline
plugins:
  - id: plugin1
`,
			wantErr: true,
		},
		{
			name: "empty plugins",
			yaml: `
name: test-pipeline
plugins: []
`,
			wantErr: true,
		},
		{
			name: "invalid yaml",
			yaml: `
name: test
  bad: indentation
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadFromBytes([]byte(tt.yaml))
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromBytes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		filename string
		content  string
		wantErr  bool
	}{
		{
			name:     "valid file",
			filename: "valid.yaml",
			content: `
name: test-pipeline
description: Test
plugins:
  - id: plugin1
`,
			wantErr: false,
		},
		{
			name:     "invalid file",
			filename: "invalid.yaml",
			content: `
name: test
plugins: []
`,
			wantErr: true,
		},
		{
			name:     "non-existent file",
			filename: "does-not-exist.yaml",
			content:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.content != "" {
				path = filepath.Join(tmpDir, tt.filename)
				if err := os.WriteFile(path, []byte(tt.content), 0644); err != nil {
					t.Fatalf("Failed to write test file: %v", err)
				}
			} else {
				path = filepath.Join(tmpDir, tt.filename)
			}

			_, err := LoadFromFile(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Name:        "test",
				Description: "test pipeline",
				Plugins: []PluginConfig{
					{ID: "plugin1"},
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "empty name",
			config: &Config{
				Name: "",
				Plugins: []PluginConfig{
					{ID: "plugin1"},
				},
			},
			wantErr: true,
		},
		{
			name: "no plugins",
			config: &Config{
				Name:    "test",
				Plugins: []PluginConfig{},
			},
			wantErr: true,
		},
		{
			name: "plugin with empty ID",
			config: &Config{
				Name: "test",
				Plugins: []PluginConfig{
					{ID: ""},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ToExecutionPlan(t *testing.T) {
	config := &Config{
		Name:        "test-pipeline",
		Description: "Test description",
		Timeout:     60,
		StopOnError: true,
		Plugins: []PluginConfig{
			{
				ID: "plugin1",
				Config: map[string]interface{}{
					"key": "value",
				},
				Optional: false,
				Timeout:  10,
			},
			{
				ID:        "plugin2",
				DependsOn: []string{"plugin1"},
				Optional:  true,
			},
		},
	}

	plan := config.ToExecutionPlan()

	if plan.Name != config.Name {
		t.Errorf("Name = %v, want %v", plan.Name, config.Name)
	}
	if plan.Description != config.Description {
		t.Errorf("Description = %v, want %v", plan.Description, config.Description)
	}
	if plan.Timeout != config.Timeout {
		t.Errorf("Timeout = %v, want %v", plan.Timeout, config.Timeout)
	}
	if plan.StopOnError != config.StopOnError {
		t.Errorf("StopOnError = %v, want %v", plan.StopOnError, config.StopOnError)
	}
	if len(plan.Steps) != len(config.Plugins) {
		t.Errorf("Steps count = %v, want %v", len(plan.Steps), len(config.Plugins))
	}

	// Validate first step
	if plan.Steps[0].PluginID != "plugin1" {
		t.Errorf("Steps[0].PluginID = %v, want %v", plan.Steps[0].PluginID, "plugin1")
	}
	if plan.Steps[0].Optional != false {
		t.Errorf("Steps[0].Optional = %v, want %v", plan.Steps[0].Optional, false)
	}
	if plan.Steps[0].Timeout != 10 {
		t.Errorf("Steps[0].Timeout = %v, want %v", plan.Steps[0].Timeout, 10)
	}

	// Validate second step
	if plan.Steps[1].PluginID != "plugin2" {
		t.Errorf("Steps[1].PluginID = %v, want %v", plan.Steps[1].PluginID, "plugin2")
	}
	if len(plan.Steps[1].DependsOn) != 1 || plan.Steps[1].DependsOn[0] != "plugin1" {
		t.Errorf("Steps[1].DependsOn = %v, want [plugin1]", plan.Steps[1].DependsOn)
	}
}
