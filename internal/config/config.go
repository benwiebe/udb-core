package config

import plugintypes "github.com/benwiebe/udb-plugin-library/types"

type RootConfig struct {
	Display DisplayConfig   `json:"display"`
	Plugins []PluginsConfig `json:"plugins"`
}

type PluginsConfig struct {
	ID     string                   `json:"id"`
	Path   string                   `json:"path,omitempty"`
	Config plugintypes.PluginConfig `json:"config"`
}
