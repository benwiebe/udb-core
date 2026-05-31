package display

import (
	"image"
	"image/color"
	"testing"

	"github.com/benwiebe/udb-core/internal/config"
)

// ── NewDisplay registry tests ─────────────────────────────────────────────────

func TestNewDisplay_UnknownType(t *testing.T) {
	_, err := NewDisplay(config.DisplayConfig{Type: "nonexistent", Width: 4, Height: 4})
	if err == nil {
		t.Fatal("expected error for unknown display type, got nil")
	}
}

func TestNewDisplay_Stub(t *testing.T) {
	d, err := NewDisplay(config.DisplayConfig{Type: "stub", Width: 4, Height: 4})
	if err != nil {
		t.Fatalf("unexpected error creating stub display: %v", err)
	}
	if d == nil {
		t.Fatal("expected non-nil display")
	}
	d.CloseDisplay()
}

// ── StubDisplay tests ─────────────────────────────────────────────────────────

func TestStubDisplay_Render(t *testing.T) {
	d := StubDisplay{}
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	if err := d.Render(img); err != nil {
		t.Fatalf("unexpected error from stub Render: %v", err)
	}
}

// ── HttpDisplay tests ─────────────────────────────────────────────────────────

func newTestHttpDisplay(scale int) *HttpDisplay {
	return &HttpDisplay{
		clients: make(map[chan []byte]struct{}),
		scale:   scale,
	}
}

func TestHttpDisplay_NoClients_Render(t *testing.T) {
	d := newTestHttpDisplay(1)
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	if err := d.Render(img); err != nil {
		t.Fatalf("Render with no clients returned error: %v", err)
	}
}

func TestHttpDisplay_ClientReceivesFrame(t *testing.T) {
	d := newTestHttpDisplay(1)
	ch := make(chan []byte, 1)
	d.clients[ch] = struct{}{}

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	if err := d.Render(img); err != nil {
		t.Fatal(err)
	}
	if len(ch) != 1 {
		t.Fatalf("expected 1 frame in client channel, got %d", len(ch))
	}
}

func TestHttpDisplay_DrainStaleFrame(t *testing.T) {
	// Verify that a slow client (channel still full) gets the latest frame,
	// not a queued stale one. After two renders without consuming, channel
	// must still hold exactly one frame.
	d := newTestHttpDisplay(1)
	ch := make(chan []byte, 1)
	d.clients[ch] = struct{}{}

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))

	if err := d.Render(img); err != nil {
		t.Fatal(err)
	}
	if len(ch) != 1 {
		t.Fatalf("expected 1 frame after first render, got %d", len(ch))
	}

	// Don't consume — simulate a slow client.
	if err := d.Render(img); err != nil {
		t.Fatal(err)
	}
	if len(ch) != 1 {
		t.Fatalf("expected 1 frame after drain+render, got %d", len(ch))
	}
}

func TestHttpDisplay_MultipleClients(t *testing.T) {
	d := newTestHttpDisplay(1)
	ch1 := make(chan []byte, 1)
	ch2 := make(chan []byte, 1)
	d.clients[ch1] = struct{}{}
	d.clients[ch2] = struct{}{}

	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	if err := d.Render(img); err != nil {
		t.Fatal(err)
	}
	if len(ch1) != 1 || len(ch2) != 1 {
		t.Fatalf("expected both clients to receive a frame; ch1=%d ch2=%d", len(ch1), len(ch2))
	}
}

func TestHttpDisplay_Scale(t *testing.T) {
	d := newTestHttpDisplay(4)
	ch := make(chan []byte, 1)
	d.clients[ch] = struct{}{}

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	// Render a non-black pixel so the JPEG isn't all zeros.
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	if err := d.Render(img); err != nil {
		t.Fatalf("Render with scale returned error: %v", err)
	}
	if len(ch) != 1 {
		t.Fatalf("expected 1 frame after scaled render, got %d", len(ch))
	}
}
