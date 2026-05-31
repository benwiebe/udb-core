// StubDisplay is a no-op display implementation for local development and testing.
// It is always compiled in (no build tag) so it can be used on any platform,
// including Linux/Raspberry Pi when you want to run without hardware attached.

package display

import (
	"image"
	"log/slog"

	"github.com/benwiebe/udb-core/internal/config"
)

func init() {
	register("stub", newStubDisplay)
}

type StubDisplay struct{}

func newStubDisplay(cfg config.DisplayConfig) (Display, error) {
	slog.Info("stub display initialized", "width", cfg.Width, "height", cfg.Height)
	return StubDisplay{}, nil
}

func (d StubDisplay) Render(img image.Image) error {
	slog.Debug("stub display render", "width", img.Bounds().Dx(), "height", img.Bounds().Dy())
	return nil
}

func (d StubDisplay) CloseDisplay() {
	slog.Info("stub display closed")
}
