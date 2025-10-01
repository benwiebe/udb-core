package config

import (
	"testing"

	"github.com/benwiebe/udb-core/internal/config"
)

func TestLoadDefaultConfig(t *testing.T) {
	// Create new config loader
	loader := config.NewDefaultConfigLoader()
	if loader == nil {
		t.Fatal("Failed to create config loader")
	}

	// Load the config
	err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config is accessible
	config := loader.GetConfig()
	if config == nil {
		t.Fatal("Config is nil after loading")
	}

	// Assert config object has the expected content
	if len(config.Plugins) != 1 {
		t.Fatalf("Expected exactly one plugin in config, got %d", len(config.Plugins))
	}
}

func TestLoadCustomConfig(t *testing.T) {
	// Create new config loader
	loader := config.NewConfigLoaderWithPath("./config2.json")
	if loader == nil {
		t.Fatal("Failed to create config loader")
	}

	// Load the config
	err := loader.Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config is accessible
	config := loader.GetConfig()
	if config == nil {
		t.Fatal("Config is nil after loading")
	}

	// Assert config object has the expected content
	if len(config.Plugins) != 2 {
		t.Fatalf("Expected exactly two plugins in config, got %d", len(config.Plugins))
	}
}
