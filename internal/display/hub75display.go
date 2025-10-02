package display

import (
	"github.com/tfk1410/go-rpi-rgb-led-matrix"
)

type Hub75Display struct {
	config  rgbmatrix.HardwareConfig
	matrix  rgbmatrix.Matrix
	toolkit *rgbmatrix.ToolKit
}

func InitializeDisplayWithConfig(config rgbmatrix.HardwareConfig) Hub75Display {
	m, _ := rgbmatrix.NewRGBLedMatrix(&config)
	dispObj := Hub75Display{
		config:  config,
		matrix:  m,
		toolkit: rgbmatrix.NewToolKit(m),
	}
	return dispObj
}

func InitializeDisplayBySize(height int, width int) Hub75Display {
	config := rgbmatrix.DefaultConfig
	config.Cols = width
	config.Rows = height
	return InitializeDisplayWithConfig(config)
}

func (disp Hub75Display) NewCanvas() *rgbmatrix.Canvas {
	return rgbmatrix.NewCanvas(disp.matrix)
}

func (disp Hub75Display) CloseDisplay() {
	disp.matrix.Close()
}
