package integration

import (
	"testing"

	"github.com/yourusername/geee/internal/config"
)

func TestCLI_Run_Integration(t *testing.T) {
	// Skip this test for now - it requires plugins to be registered
	// This will be tested in Phase 4 after plugins are implemented
	t.Skip("Skipping integration test - requires registered plugins (Phase 4)")
}

func TestConfigLoader_Integration(t *testing.T) {
	// Test loading the default config
	configPath := "../../configs/default.yaml"

	cfg, err := config.LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}

	if cfg.Name != "default-pipeline" {
		t.Errorf("Expected name 'default-pipeline', got '%s'", cfg.Name)
	}

	if len(cfg.Plugins) == 0 {
		t.Error("Expected at least one plugin in default config")
	}

	// Verify it can be converted to execution plan
	plan := cfg.ToExecutionPlan()
	if plan == nil {
		t.Error("Expected non-nil execution plan")
	}

	if plan.Name != cfg.Name {
		t.Errorf("Expected plan name '%s', got '%s'", cfg.Name, plan.Name)
	}
}
