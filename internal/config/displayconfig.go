package config

type DisplayConfig struct {
	Height     int `json:"height"`
	Width      int `json:"width"`
	Brightness int `json:"brightness,omitempty"`
}
