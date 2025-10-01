package config

import plugintypes "github.com/benwiebe/udb-plugin-library/types"

type RootConfig struct {
	Plugins []PluginsConfig `json:"plugins"`
}

type PluginsConfig struct {
	ID     string                   `json:"id"`
	Config plugintypes.PluginConfig `json:"config"`
}
