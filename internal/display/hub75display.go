package display

import "github.com/tfk1410/go-rpi-rgb-led-matrix"

type Hub75Display struct {
	config rgbmatrix.HardwareConfig
	matrix rgbmatrix.Matrix
}

func initializeDisplayWithConfig(config *rgbmatrix.HardwareConfig) Hub75Display {
	m, _ := rgbmatrix.NewRGBLedMatrix(config)
	dispObj := Hub75Display{
		config: *config,
		matrix: m,
	}
	return dispObj
}

func initializeDisplayBySize(height int, width int) Hub75Display {
	config := rgbmatrix.DefaultConfig
	config.Cols = width
	config.Rows = height
	return initializeDisplayWithConfig(&config)
}

func closeDisplay(disp *Hub75Display) {
	disp.matrix.Close()
}
