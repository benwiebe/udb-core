//go:build linux

// This file contains the real Hub75 LED matrix display implementation using the
// go-rpi-rgb-led-matrix library. It is only compiled on Linux (i.e. Raspberry Pi).
// For local development on other platforms, see hub75display_stub.go.

package display

import (
	"image"
	"image/draw"

	"github.com/benwiebe/udb-core/internal/config"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

type Hub75Display struct {
	config  rgbmatrix.HardwareConfig
	matrix  rgbmatrix.Matrix
	toolkit *rgbmatrix.ToolKit
}

func InitializeDisplay(displayConfig config.DisplayConfig) Hub75Display {
	hwConfig := rgbmatrix.DefaultConfig
	hwConfig.Cols = displayConfig.Width
	hwConfig.Rows = displayConfig.Height
	if displayConfig.Brightness > 0 {
		hwConfig.Brightness = displayConfig.Brightness
	}
	m, _ := rgbmatrix.NewRGBLedMatrix(&hwConfig)
	return Hub75Display{
		config:  hwConfig,
		matrix:  m,
		toolkit: rgbmatrix.NewToolKit(m),
	}
}

func (disp Hub75Display) Render(img image.Image) error {
	canvas := rgbmatrix.NewCanvas(disp.matrix)
	defer canvas.Close()
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)
	return canvas.Render()
}

func (disp Hub75Display) CloseDisplay() {
	disp.matrix.Close()
}
