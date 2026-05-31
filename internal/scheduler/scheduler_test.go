package scheduler

import (
	"context"
	"encoding/json"
	"image"
	"sync"
	"testing"
	"time"

	"github.com/benwiebe/udb-core/internal/config"
	"github.com/benwiebe/udb-core/internal/plugins"
	"github.com/benwiebe/udb-plugin-library/types"
)

// ── fake display ──────────────────────────────────────────────────────────────

type fakeDisplay struct {
	mu      sync.Mutex
	renders int
}

func (d *fakeDisplay) Render(_ image.Image) error {
	d.mu.Lock()
	d.renders++
	d.mu.Unlock()
	return nil
}

func (d *fakeDisplay) CloseDisplay() {}

func (d *fakeDisplay) count() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.renders
}

// ── shared blank image ────────────────────────────────────────────────────────

var blankImg image.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))

// ── fake static board ─────────────────────────────────────────────────────────

type fakeStaticBoard struct{}

func (b *fakeStaticBoard) GetId() string                                                                 { return "s" }
func (b *fakeStaticBoard) GetName() string                                                               { return "Static" }
func (b *fakeStaticBoard) GetSupportedDimensions() []types.BoardDimensions                              { return nil }
func (b *fakeStaticBoard) GetType() types.BoardType                                                      { return types.BoardTypeStatic }
func (b *fakeStaticBoard) GetDatasourceType() string                                                     { return "" }
func (b *fakeStaticBoard) Init(_ json.RawMessage, _ types.Datasource[any], _ types.BoardDimensions) error { return nil }
func (b *fakeStaticBoard) Render() image.Image                                                           { return blankImg }

// ── fake animated board ───────────────────────────────────────────────────────

type fakeAnimatedBoard struct {
	frames types.Animation
}

func (b *fakeAnimatedBoard) GetId() string                                                                 { return "a" }
func (b *fakeAnimatedBoard) GetName() string                                                               { return "Animated" }
func (b *fakeAnimatedBoard) GetSupportedDimensions() []types.BoardDimensions                              { return nil }
func (b *fakeAnimatedBoard) GetType() types.BoardType                                                      { return types.BoardTypeAnimated }
func (b *fakeAnimatedBoard) GetDatasourceType() string                                                     { return "" }
func (b *fakeAnimatedBoard) Init(_ json.RawMessage, _ types.Datasource[any], _ types.BoardDimensions) error { return nil }
func (b *fakeAnimatedBoard) Render() types.Animation                                                       { return b.frames }

// ── fake dynamic board ────────────────────────────────────────────────────────

type fakeDynamicBoard struct {
	frameDuration time.Duration
	renderCount   int
}

func (b *fakeDynamicBoard) GetId() string                                                                 { return "d" }
func (b *fakeDynamicBoard) GetName() string                                                               { return "Dynamic" }
func (b *fakeDynamicBoard) GetSupportedDimensions() []types.BoardDimensions                              { return nil }
func (b *fakeDynamicBoard) GetType() types.BoardType                                                      { return types.BoardTypeDynamic }
func (b *fakeDynamicBoard) GetDatasourceType() string                                                     { return "" }
func (b *fakeDynamicBoard) Init(_ json.RawMessage, _ types.Datasource[any], _ types.BoardDimensions) error { return nil }
func (b *fakeDynamicBoard) Render() types.AnimationFrame {
	b.renderCount++
	return types.AnimationFrame{Img: blankImg, Duration: b.frameDuration}
}

// ── fake datasource ───────────────────────────────────────────────────────────

type fakeDatasource struct {
	changed chan struct{}
}

func newFakeDatasource() *fakeDatasource {
	return &fakeDatasource{changed: make(chan struct{}, 1)}
}

func (d *fakeDatasource) GetId() string                 { return "ds" }
func (d *fakeDatasource) GetName() string               { return "DS" }
func (d *fakeDatasource) GetType() string               { return "" }
func (d *fakeDatasource) GetData() any                  { return nil }
func (d *fakeDatasource) Start(_ context.Context) error { return nil }
func (d *fakeDatasource) DataChanged() <-chan struct{}   { return d.changed }

// ── helpers ───────────────────────────────────────────────────────────────────

func makeEntry(board types.Board[any], durationSeconds int) plugins.BoardEntry {
	return plugins.BoardEntry{
		Board:  board,
		Config: config.BoardConfig{DurationSeconds: durationSeconds},
	}
}

func makeEntryWithDS(board types.Board[any], durationSeconds int, ds types.Datasource[any]) plugins.BoardEntry {
	return plugins.BoardEntry{
		Board:      board,
		Config:     config.BoardConfig{DurationSeconds: durationSeconds},
		Datasource: ds,
	}
}

// ── static board tests ────────────────────────────────────────────────────────

func TestStaticBoard_RendersOnce_NoDuration(t *testing.T) {
	d := &fakeDisplay{}
	runBoard(context.Background(), d, makeEntry(&fakeStaticBoard{}, 0))
	if d.count() != 1 {
		t.Fatalf("expected 1 render, got %d", d.count())
	}
}

func TestStaticBoard_ContextCancels(t *testing.T) {
	d := &fakeDisplay{}
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		runBoard(ctx, d, makeEntry(&fakeStaticBoard{}, 60))
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runBoard did not exit after context cancellation")
	}
	if d.count() != 1 {
		t.Fatalf("expected 1 render before cancel, got %d", d.count())
	}
}

// ── animated board tests ──────────────────────────────────────────────────────

func TestAnimatedBoard_RendersAllFrames_NoDuration(t *testing.T) {
	d := &fakeDisplay{}
	board := &fakeAnimatedBoard{
		frames: types.Animation{
			{Img: blankImg, Duration: 1 * time.Millisecond},
			{Img: blankImg, Duration: 1 * time.Millisecond},
			{Img: blankImg, Duration: 1 * time.Millisecond},
		},
	}
	runBoard(context.Background(), d, makeEntry(board, 0))
	if d.count() != 3 {
		t.Fatalf("expected 3 renders (one per frame), got %d", d.count())
	}
}

func TestAnimatedBoard_DeadlineCap(t *testing.T) {
	// Frame duration (10s) >> board duration (1s). The deadline cap must
	// clamp the sleep so the board exits near the 1s mark, not at 10s.
	d := &fakeDisplay{}
	board := &fakeAnimatedBoard{
		frames: types.Animation{
			{Img: blankImg, Duration: 10 * time.Second},
		},
	}
	start := time.Now()
	runBoard(context.Background(), d, makeEntry(board, 1))
	elapsed := time.Since(start)

	if d.count() < 1 {
		t.Fatal("expected at least 1 render")
	}
	if elapsed > 3*time.Second {
		t.Fatalf("expected board to exit near the 1s deadline, took %v", elapsed)
	}
}

// ── dynamic board tests ───────────────────────────────────────────────────────

func TestDynamicBoard_RendersRepeatedly(t *testing.T) {
	d := &fakeDisplay{}
	board := &fakeDynamicBoard{frameDuration: 5 * time.Millisecond}
	runBoard(context.Background(), d, makeEntry(board, 1))
	if d.count() < 10 {
		t.Fatalf("expected many renders in 1s, got %d", d.count())
	}
}

func TestDynamicBoard_DeadlineCap(t *testing.T) {
	// Frame duration (10s) >> board duration (1s). Deadline cap must kick in.
	d := &fakeDisplay{}
	board := &fakeDynamicBoard{frameDuration: 10 * time.Second}
	start := time.Now()
	runBoard(context.Background(), d, makeEntry(board, 1))
	elapsed := time.Since(start)

	if board.renderCount < 1 {
		t.Fatal("expected at least 1 render")
	}
	if elapsed > 3*time.Second {
		t.Fatalf("expected board to exit near the 1s deadline, took %v", elapsed)
	}
}

func TestDynamicBoard_DataChangedTriggersReRender(t *testing.T) {
	// Frame duration is very long, so without the DataChanged signal only one
	// render would occur. The signal should cause an immediate second render.
	d := &fakeDisplay{}
	board := &fakeDynamicBoard{frameDuration: 10 * time.Second}
	ds := newFakeDatasource()

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		ds.changed <- struct{}{}
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	runBoard(ctx, d, makeEntryWithDS(board, 60, ds))

	if board.renderCount < 2 {
		t.Fatalf("expected at least 2 renders (initial + DataChanged), got %d", board.renderCount)
	}
}

func TestDynamicBoard_ContextCancels(t *testing.T) {
	d := &fakeDisplay{}
	board := &fakeDynamicBoard{frameDuration: 10 * time.Second}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runBoard(ctx, d, makeEntry(board, 0)) // duration=0 → infinite
		close(done)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("runBoard did not exit after context cancellation")
	}
}

// ── Run() tests ───────────────────────────────────────────────────────────────

func TestRun_CyclesThroughAllBoards(t *testing.T) {
	d := &fakeDisplay{}
	boards := []plugins.BoardEntry{
		makeEntry(&fakeStaticBoard{}, 0),
		makeEntry(&fakeStaticBoard{}, 0),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	Run(ctx, d, boards)

	if d.count() < 2 {
		t.Fatalf("expected both boards rendered, got %d total renders", d.count())
	}
}
