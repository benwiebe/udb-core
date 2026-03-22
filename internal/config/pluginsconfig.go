package config

import plugintypes "github.com/benwiebe/udb-plugin-library/types"

// PluginsConfig defines the array of plugins which will be loaded.
type PluginsConfig = []PluginConfig

// PluginConfig defines the configuration for a single plugin. The plugin ID is required.
// The path is optional, and if not specified, the plugin will be loaded from the default
// location.
//
// ID is the plugin-defined ID and must match the plugin's containing folder.
//
// Path is the absolute or relative path to the plugin. This is optional.
//
// Config is the plugin-specific configuration.
type PluginConfig struct {
	ID     string                   `json:"id"`
	Path   string                   `json:"path,omitempty"`
	Config plugintypes.PluginConfig `json:"config,omitempty"`
}
