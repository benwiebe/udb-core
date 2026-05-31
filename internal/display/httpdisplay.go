// HttpDisplay serves a live MJPEG stream over HTTP so any browser can view the
// display output in real time without refreshing. Useful for local development
// on machines without LED hardware attached.

package display

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"net/http"
	"sync"

	"github.com/benwiebe/udb-core/internal/config"
	xdraw "golang.org/x/image/draw"
)

func init() {
	register("http", newHttpDisplay)
}

type HttpDisplay struct {
	mu      sync.Mutex
	clients map[chan []byte]struct{}
	scale   int
}

func newHttpDisplay(cfg config.DisplayConfig) (Display, error) {
	scale := cfg.Scale
	if scale <= 0 {
		scale = 1
	}
	d := &HttpDisplay{
		clients: make(map[chan []byte]struct{}),
		scale:   scale,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.streamHandler)
	go http.ListenAndServe(":8080", mux)

	fmt.Println("HTTP display streaming at http://localhost:8080")
	return d, nil
}

func (d *HttpDisplay) Render(img image.Image) error {
	if d.scale > 1 {
		b := img.Bounds()
		scaled := image.NewRGBA(image.Rect(0, 0, b.Dx()*d.scale, b.Dy()*d.scale))
		xdraw.NearestNeighbor.Scale(scaled, scaled.Bounds(), img, b, draw.Src, nil)
		img = scaled
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return err
	}
	frame := buf.Bytes()

	d.mu.Lock()
	defer d.mu.Unlock()
	for ch := range d.clients {
		// Drain any unconsumed frame so the latest always wins.
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- frame:
		default:
		}
	}
	return nil
}

func (d *HttpDisplay) streamHandler(w http.ResponseWriter, r *http.Request) {
	ch := make(chan []byte, 1)
	d.mu.Lock()
	d.clients[ch] = struct{}{}
	d.mu.Unlock()
	defer func() {
		d.mu.Lock()
		delete(d.clients, ch)
		d.mu.Unlock()
	}()

	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache")

	for {
		select {
		case <-r.Context().Done():
			return
		case frame := <-ch:
			fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n", len(frame))
			w.Write(frame)
			fmt.Fprint(w, "\r\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

func (d *HttpDisplay) CloseDisplay() {}
