//go:build !linux

// This file is a stub implementation of the Hub75 display for use on non-Linux platforms
// (e.g. local development on macOS). It satisfies the same interface as hub75display.go
// but does not depend on any hardware libraries. All display operations are no-ops.

package display

import (
	"fmt"
	"image"

	"github.com/benwiebe/udb-core/internal/config"
)

type Hub75Display struct{}

func InitializeDisplay(displayConfig config.DisplayConfig) Hub75Display {
	fmt.Printf("Stub display initialized (%dx%d)\n", displayConfig.Width, displayConfig.Height)
	return Hub75Display{}
}

func (disp Hub75Display) Render(img image.Image) error {
	fmt.Printf("Stub display render: %dx%d image\n", img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

func (disp Hub75Display) CloseDisplay() {
	fmt.Println("Stub display closed")
}
