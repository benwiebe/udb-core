package config

import "encoding/json"

// DatasourcesConfig defines the array of datasources available for boards to use.
type DatasourcesConfig = []DatasourceConfig

// DatasourceConfig defines a named datasource instance. The user-defined ID can be
// referenced in board configs to explicitly assign a datasource to a board.
//
// Id is the user-defined identifier for this datasource instance. It must be unique
// across all datasources in the config.
//
// Plugin is the ID of the plugin that provides this datasource.
//
// DatasourceId is the ID of the specific datasource within the plugin.
//
// Config is a JSON object containing the configuration for the datasource. This varies
// depending on the datasource type.
type DatasourceConfig struct {
	Id           string          `json:"id"`
	Plugin       string          `json:"plugin"`
	DatasourceId string          `json:"datasourceId"`
	Config       json.RawMessage `json:"config,omitempty"`
}
