package config

// RootConfig defines the shape of the config file.
//
// Display defines the configuration for the display.
//
// Plugins defines the configuration for the plugins to load.
//
// Datasources defines named datasource instances available for boards to reference.
//
// Boards defines the configuration for the boards sequence to display.
type RootConfig struct {
	Display     DisplayConfig     `json:"display"`
	Plugins     PluginsConfig     `json:"plugins"`
	Datasources DatasourcesConfig `json:"datasources"`
	Boards      BoardsConfig      `json:"boards"`
}
