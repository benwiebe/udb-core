//go:build linux

// This file contains the real Hub75 LED matrix display implementation using the
// go-rpi-rgb-led-matrix library. It is only compiled on Linux (i.e. Raspberry Pi).
// For local development on other platforms, see stubdisplay.go.

package display

import (
	"image"
	"image/draw"

	"github.com/benwiebe/udb-core/internal/config"
	rgbmatrix "github.com/tfk1410/go-rpi-rgb-led-matrix"
)

func init() {
	defaultType = "hub75"
	register("hub75", newHub75Display)
}

type Hub75Display struct {
	config  rgbmatrix.HardwareConfig
	matrix  rgbmatrix.Matrix
	toolkit *rgbmatrix.ToolKit
}

func newHub75Display(displayConfig config.DisplayConfig) (Display, error) {
	hwConfig := rgbmatrix.DefaultConfig
	hwConfig.Cols = displayConfig.Width
	hwConfig.Rows = displayConfig.Height
	if displayConfig.Brightness > 0 {
		hwConfig.Brightness = displayConfig.Brightness
	}
	if displayConfig.GpioMapping != "" {
		hwConfig.HardwareMapping = displayConfig.GpioMapping
	}
	hwConfig.DisableHardwarePulsing = displayConfig.DisableHardwarePulsing
	m, err := rgbmatrix.NewRGBLedMatrix(&hwConfig)
	if err != nil {
		return nil, err
	}
	return Hub75Display{
		config:  hwConfig,
		matrix:  m,
		toolkit: rgbmatrix.NewToolKit(m),
	}, nil
}

func (disp Hub75Display) Render(img image.Image) error {
	canvas := rgbmatrix.NewCanvas(disp.matrix)
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)
	return canvas.Render()
}

func (disp Hub75Display) CloseDisplay() {
	disp.matrix.Close()
}
