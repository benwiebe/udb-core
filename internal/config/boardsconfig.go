package config

import "encoding/json"

// BoardsConfig defines the array of boards which will be displayed in sequence.
type BoardsConfig = []BoardConfig

// BoardConfig defines the configuration for a single board. The plugin ID and board ID
// are required to identify the exact board to display.
//
// Plugin represents the ID of the plugin containing the desired board.
//
// ID represents the ID of the board within the plugin.
//
// DurationSeconds is the number of seconds to display the board for. If not specified,
// the board will be displayed indefinitely.
//
// Config is a JSON object containing the configuration for the board. This varies
// depending on the type of board.
//
// Datasource is the datasource to use for this board. If not specified, UDB will attempt
// to automatically match a datasource of the right type to the board.
type BoardConfig struct {
	Plugin          string          `json:"plugin"`
	ID              string          `json:"id"`
	DurationSeconds int             `json:"durationSeconds,omitempty"`
	Config          json.RawMessage `json:"config,omitempty"`
	Datasource      BoardDatasource `json:"datasource,omitempty"`
}

// BoardDatasource defines the configuration for a datasource to use for a board instance.
// Multiple boards from the same plugin, or even of the same type, can be configured to use
// the same or different datasources. The plugin ID and datasource ID are required to identify
// the exact datasource to use.
type BoardDatasource struct {
	// Plugin represents the ID of the plugin containing the desired datasource.
	Plugin string `json:"plugin"`
	// ID represents the ID of the datasource within the plugin.
	ID string `json:"id"`
	// Config is a JSON object containing the configuration for the datasource. This varies
	// depending on the type of datasource.
	Config json.RawMessage `json:"config,omitempty"`
}
