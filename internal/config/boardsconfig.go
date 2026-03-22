package config

import "encoding/json"

// BoardsConfig defines the array of boards which will be displayed in sequence.
type BoardsConfig = []BoardConfig

// BoardConfig defines the configuration for a single board. The plugin ID and board ID
// are required to identify the exact board to display.
//
// Plugin represents the ID of the plugin containing the desired board.
//
// BoardId represents the ID of the board within the plugin.
//
// DurationSeconds is the number of seconds to display the board for. If not specified,
// the board will be displayed indefinitely.
//
// Config is a JSON object containing the configuration for the board. This varies
// depending on the type of board.
//
// Datasource is the user-defined ID of a datasource from the top-level datasources config.
// If omitted, UDB will attempt to automatically match a compatible datasource from the
// same plugin.
type BoardConfig struct {
	Plugin          string          `json:"plugin"`
	BoardId         string          `json:"boardId"`
	DurationSeconds int             `json:"durationSeconds,omitempty"`
	Config          json.RawMessage `json:"config,omitempty"`
	Datasource      string          `json:"datasource,omitempty"`
}
