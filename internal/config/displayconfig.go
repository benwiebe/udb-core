package config

import rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"

type DisplayConfig struct {
	Height     int `json:"height"`
	Width      int `json:"width"`
	Brightness int `json:"brightness,omitempty"`
}

func (displayConfig DisplayConfig) ConvertToHardwareConfig() rgbmatrix.HardwareConfig {
	hwConfig := rgbmatrix.DefaultConfig
	hwConfig.Cols = displayConfig.Width
	hwConfig.Rows = displayConfig.Height
	if displayConfig.Brightness > 0 {
		hwConfig.Brightness = displayConfig.Brightness
	}
	return hwConfig
}
