package config

import plugintypes "github.com/benwiebe/udb-plugin-library/types"

// PluginsConfig defines the array of plugins to configure.
type PluginsConfig = []PluginConfig

// PluginConfig defines the configuration for a single plugin.
//
// ID must match the value returned by the plugin's GetId() method.
//
// Config is the plugin-specific configuration (API keys, settings, etc.).
// Omit for plugins that require no configuration.
type PluginConfig struct {
	ID     string                   `json:"id"`
	Config plugintypes.PluginConfig `json:"config,omitempty"`
}
