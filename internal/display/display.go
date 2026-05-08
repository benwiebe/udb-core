package display

import (
	"fmt"
	"image"

	"github.com/benwiebe/udb-core/internal/config"
)

// Display is the interface satisfied by all display implementations.
// The real hardware implementation is in hub75display.go (linux only).
// A no-op stub for local development is in stubdisplay.go.
// A Fyne window implementation would live in fynedisplay.go.
type Display interface {
	// Render pushes an image to the display.
	Render(img image.Image) error
	// CloseDisplay releases display resources.
	CloseDisplay()
}

// defaultType is the display type used when none is specified in config.
// It is "stub" by default; hub75display.go overrides it to "hub75" on Linux.
var defaultType = "stub"

var registry = map[string]func(config.DisplayConfig) (Display, error){}

func register(name string, f func(config.DisplayConfig) (Display, error)) {
	registry[name] = f
}

// NewDisplay instantiates the display implementation named by cfg.Type.
// If cfg.Type is empty the platform default is used ("hub75" on Linux, "stub" elsewhere).
// Returns an error if the requested type is unknown or not compiled in.
func NewDisplay(cfg config.DisplayConfig) (Display, error) {
	t := cfg.Type
	if t == "" {
		t = defaultType
	}
	f, ok := registry[t]
	if !ok {
		return nil, fmt.Errorf("unknown display type %q (not compiled in for this platform?)", t)
	}
	return f(cfg)
}
