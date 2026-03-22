package display

import "image"

// Display is the interface satisfied by all display implementations.
// The real hardware implementation is in hub75display.go (linux only).
// A no-op stub for local development is in hub75display_stub.go.
type Display interface {
	// Render pushes an image to the display.
	Render(img image.Image) error
	// CloseDisplay releases display resources.
	CloseDisplay()
}
