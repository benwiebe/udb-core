package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/benwiebe/udb-core/internal/config"
	"github.com/benwiebe/udb-plugin-library/types"
)

// ── fake board ────────────────────────────────────────────────────────────────

type fakeBoard struct {
	id         string
	dsType     string
	initErr    error
	initCalled bool
	initDims   types.BoardDimensions
}

func (b *fakeBoard) GetId() string                          { return b.id }
func (b *fakeBoard) GetName() string                        { return b.id }
func (b *fakeBoard) GetSupportedDimensions() []types.BoardDimensions { return nil }
func (b *fakeBoard) GetType() types.BoardType               { return types.BoardTypeStatic }
func (b *fakeBoard) GetDatasourceType() string              { return b.dsType }
func (b *fakeBoard) Init(_ json.RawMessage, _ types.Datasource[any], dims types.BoardDimensions) error {
	b.initCalled = true
	b.initDims = dims
	return b.initErr
}

// ── fake datasource ───────────────────────────────────────────────────────────

type fakeDatasource struct {
	id       string
	dsType   string
	startErr error
	started  bool
}

func (d *fakeDatasource) GetId() string                 { return d.id }
func (d *fakeDatasource) GetName() string               { return d.id }
func (d *fakeDatasource) GetType() string               { return d.dsType }
func (d *fakeDatasource) GetData() any                  { return nil }
func (d *fakeDatasource) Start(_ context.Context) error { d.started = true; return d.startErr }
func (d *fakeDatasource) DataChanged() <-chan struct{}   { return nil }

// ── WireDatasources ───────────────────────────────────────────────────────────

func TestWireDatasources_Explicit(t *testing.T) {
	ds := &fakeDatasource{id: "ds1", dsType: "type-a"}
	board := &fakeBoard{id: "b1"}
	boards := []BoardEntry{{Board: board, Config: config.BoardConfig{Datasource: "ds1"}}}

	WireDatasources(boards, map[string]types.Datasource[any]{"ds1": ds})

	if boards[0].Datasource != ds {
		t.Fatal("expected explicit datasource to be wired")
	}
}

func TestWireDatasources_ExplicitNotFound(t *testing.T) {
	board := &fakeBoard{id: "b1"}
	boards := []BoardEntry{{Board: board, Config: config.BoardConfig{Datasource: "missing"}}}

	WireDatasources(boards, map[string]types.Datasource[any]{})

	if boards[0].Datasource != nil {
		t.Fatal("expected nil datasource when explicit ref is missing")
	}
}

func TestWireDatasources_AutoMatch_OneMatch(t *testing.T) {
	ds := &fakeDatasource{id: "ds1", dsType: "type-a"}
	board := &fakeBoard{id: "b1", dsType: "type-a"}
	boards := []BoardEntry{{Board: board, Config: config.BoardConfig{}}}

	WireDatasources(boards, map[string]types.Datasource[any]{"ds1": ds})

	if boards[0].Datasource != ds {
		t.Fatal("expected auto-matched datasource to be wired")
	}
}

func TestWireDatasources_AutoMatch_NoMatch(t *testing.T) {
	board := &fakeBoard{id: "b1", dsType: "type-a"}
	boards := []BoardEntry{{Board: board, Config: config.BoardConfig{}}}

	WireDatasources(boards, map[string]types.Datasource[any]{})

	if boards[0].Datasource != nil {
		t.Fatal("expected nil datasource when no auto-match found")
	}
}

func TestWireDatasources_AutoMatch_MultipleMatches(t *testing.T) {
	ds1 := &fakeDatasource{id: "ds1", dsType: "type-a"}
	ds2 := &fakeDatasource{id: "ds2", dsType: "type-a"}
	board := &fakeBoard{id: "b1", dsType: "type-a"}
	boards := []BoardEntry{{Board: board, Config: config.BoardConfig{}}}

	WireDatasources(boards, map[string]types.Datasource[any]{"ds1": ds1, "ds2": ds2})

	if boards[0].Datasource != nil {
		t.Fatal("expected nil datasource when multiple auto-matches are ambiguous")
	}
}

func TestWireDatasources_NoDatasourceNeeded(t *testing.T) {
	board := &fakeBoard{id: "b1", dsType: ""}
	boards := []BoardEntry{{Board: board, Config: config.BoardConfig{}}}

	WireDatasources(boards, map[string]types.Datasource[any]{})

	if boards[0].Datasource != nil {
		t.Fatal("expected nil datasource for board that declares no datasource type")
	}
}

// ── InitBoards ────────────────────────────────────────────────────────────────

func TestInitBoards_AllSucceed(t *testing.T) {
	b1 := &fakeBoard{id: "b1"}
	b2 := &fakeBoard{id: "b2"}
	boards := []BoardEntry{{Board: b1}, {Board: b2}}
	dims := types.BoardDimensions{Width: 64, Height: 32}

	result := InitBoards(boards, dims)

	if len(result) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(result))
	}
	if !b1.initCalled || !b2.initCalled {
		t.Fatal("expected Init to be called on all boards")
	}
	if b1.initDims != dims || b2.initDims != dims {
		t.Fatal("expected display dimensions passed to Init")
	}
}

func TestInitBoards_FailedInitDropped(t *testing.T) {
	good := &fakeBoard{id: "good"}
	bad := &fakeBoard{id: "bad", initErr: errors.New("init failed")}
	boards := []BoardEntry{{Board: good}, {Board: bad}}

	result := InitBoards(boards, types.BoardDimensions{Width: 64, Height: 32})

	if len(result) != 1 {
		t.Fatalf("expected 1 board after bad init dropped, got %d", len(result))
	}
	if result[0].Board != good {
		t.Fatal("expected the successfully-initialized board in result")
	}
}

// ── StartDatasources ──────────────────────────────────────────────────────────

func TestStartDatasources_AllStart(t *testing.T) {
	ds1 := &fakeDatasource{id: "ds1"}
	ds2 := &fakeDatasource{id: "ds2"}
	dsMap := map[string]types.Datasource[any]{"ds1": ds1, "ds2": ds2}

	result := StartDatasources(context.Background(), dsMap)

	if len(result) != 2 {
		t.Fatalf("expected 2 started datasources, got %d", len(result))
	}
	if !ds1.started || !ds2.started {
		t.Fatal("expected Start to be called on all datasources")
	}
}

func TestStartDatasources_FailedStartDropped(t *testing.T) {
	good := &fakeDatasource{id: "good"}
	bad := &fakeDatasource{id: "bad", startErr: errors.New("start failed")}
	dsMap := map[string]types.Datasource[any]{"good": good, "bad": bad}

	result := StartDatasources(context.Background(), dsMap)

	if len(result) != 1 {
		t.Fatalf("expected 1 datasource after failed start dropped, got %d", len(result))
	}
	if _, ok := result["good"]; !ok {
		t.Fatal("expected the successfully-started datasource in result")
	}
}
