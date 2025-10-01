package config

import (
	"encoding/json"
	"os"
)

const defaultConfigPath = "./config.json"

type ConfigLoader struct {
	configPath string
	config     *RootConfig
}

func NewConfigLoaderWithPath(pathOverride string) *ConfigLoader {
	return &ConfigLoader{
		configPath: pathOverride,
	}
}

func NewDefaultConfigLoader() *ConfigLoader {
	return NewConfigLoaderWithPath(defaultConfigPath)
}

func (loader *ConfigLoader) Load() error {
	// Open the config file
	file, err := os.Open(loader.configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read the file contents
	decoder := json.NewDecoder(file)
	var config RootConfig
	if err := decoder.Decode(&config); err != nil {
		return err
	}

	// Store the loaded config in the loader
	loader.config = &config
	return nil
}

// GetConfig returns the loaded configuration
func (loader *ConfigLoader) GetConfig() *RootConfig {
	return loader.config
}
