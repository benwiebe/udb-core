// StubDisplay is a no-op display implementation for local development and testing.
// It is always compiled in (no build tag) so it can be used on any platform,
// including Linux/Raspberry Pi when you want to run without hardware attached.

package display

import (
	"fmt"
	"image"

	"github.com/benwiebe/udb-core/internal/config"
)

func init() {
	register("stub", newStubDisplay)
}

type StubDisplay struct{}

func newStubDisplay(cfg config.DisplayConfig) (Display, error) {
	fmt.Printf("Stub display initialized (%dx%d)\n", cfg.Width, cfg.Height)
	return StubDisplay{}, nil
}

func (d StubDisplay) Render(img image.Image) error {
	fmt.Printf("Stub display render: %dx%d image\n", img.Bounds().Dx(), img.Bounds().Dy())
	return nil
}

func (d StubDisplay) CloseDisplay() {
	fmt.Println("Stub display closed")
}
