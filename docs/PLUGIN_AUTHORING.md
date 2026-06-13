# Plugin Authoring Guide

A UDB plugin is a normal Go package that provides boards, datasources, or both. It is compiled directly into the UDB binary — there are no `.so` files, no special build modes, and no Linux requirement for development.

## Prerequisites

- Go 1.21 or later
- [`udb-plugin-library`](https://github.com/benwiebe/udb-plugin-library) as a dependency

## Concepts

**Board** — knows how to render a frame (or sequence of frames) as an `image.Image`. It receives data from a datasource.

**Datasource** — knows how to fetch and cache data from an external source. It exposes that data via `GetData()`.

**Plugin** — a container that packages one or more boards and/or datasources, and exposes them to the core via a standard interface.

## Setting Up a Plugin Project

```bash
mkdir my-udb-plugin
cd my-udb-plugin
go mod init github.com/yourname/my-udb-plugin
go get github.com/benwiebe/udb-plugin-library
```

## Plugin Entry Point

Plugins register themselves by calling `udb_plugin_library.Register()` from an `init()` function. The blank import of your package in `plugin_imports.go` triggers this automatically at startup.

```go
package my_udb_plugin

import (
    library "github.com/benwiebe/udb-plugin-library"
    "github.com/benwiebe/udb-plugin-library/types"
)

func init() {
    library.Register(&MyPlugin{})
}

type MyPlugin struct{}

func (p *MyPlugin) GetId() string                          { return "my-plugin" }
func (p *MyPlugin) GetName() string                        { return "My Plugin" }
func (p *MyPlugin) GetPluginType() types.PluginType        { return types.PluginTypeBoards }
func (p *MyPlugin) Configure(cfg types.PluginConfig) error { return nil }
```

`GetId()` must match the `id` field in the user's `config.json` plugin entry (if the plugin needs configuration). If it needs no configuration, no `config.json` entry is required at all.

### Plugin Types

| Constant | Use When |
|----------|----------|
| `types.PluginTypeBoards` | Your plugin provides only boards |
| `types.PluginTypeDatasource` | Your plugin provides only datasources |
| `types.PluginTypeCombined` | Your plugin provides both |

To expose boards, implement `UdbBoardPlugin`:

```go
func (p *MyPlugin) GetBoardMap() map[string]types.Board[any] {
    return map[string]types.Board[any]{
        "my-board": NewMyBoard(),
    }
}

func (p *MyPlugin) GetAllBoards() []types.Board[any] {
    boards := []types.Board[any]{}
    for _, b := range p.GetBoardMap() {
        boards = append(boards, b)
    }
    return boards
}
```

To expose datasources, implement `UdbDatasourcePlugin`:

```go
func (p *MyPlugin) GetDatasourceMap() map[string]types.Datasource[any] {
    return map[string]types.Datasource[any]{
        "my-datasource": &MyDatasource{},
    }
}

func (p *MyPlugin) GetAllDatasources() []types.Datasource[any] {
    datasources := []types.Datasource[any]{}
    for _, d := range p.GetDatasourceMap() {
        datasources = append(datasources, d)
    }
    return datasources
}
```

## Implementing a Board

Every board implements the base `Board[T]` interface plus one of the three render interfaces depending on how it animates.

### Board Interface

```go
type Board[T any] interface {
    GetId() string
    GetName() string
    GetSupportedDimensions() []BoardDimensions
    GetType() BoardType
    GetDatasourceType() string
    Init(config json.RawMessage, datasource Datasource[T], dimensions BoardDimensions) error
}
```

- **`GetSupportedDimensions()`** — return an empty slice to indicate the board handles any size. Otherwise list the panel dimensions you natively support (e.g. `{Width: 64, Height: 32}`). UDB uses this for future auto-scaling.
- **`GetDatasourceType()`** — return a type string matching your datasource's `GetType()`, or `""` if the board needs no datasource. See [Type String Convention](#type-string-convention).
- **`Init()`** — parse your board config from `config` (JSON), store the datasource for use in `Render()`, and use `dimensions` to pre-compute any layout values (font sizes, image buffers, drawing coordinates) that depend on the display size. The display dimensions never change after `Init()` returns. Return a non-nil error to decline initialization (the board will be skipped with a warning).

### Board Types

#### Static Board

Renders once and holds the image for the configured duration.

```go
type MyStaticBoard struct {
    cachedImage image.Image
}

func (b *MyStaticBoard) GetType() types.BoardType { return types.BoardTypeStatic }

func (b *MyStaticBoard) Init(cfg json.RawMessage, _ types.Datasource[any], dims types.BoardDimensions) error {
    img := image.NewRGBA(image.Rect(0, 0, dims.Width, dims.Height))
    // draw to img using config values...
    b.cachedImage = img
    return nil
}

func (b *MyStaticBoard) Render() image.Image {
    return b.cachedImage
}
```

#### Animated Board

Returns a full pre-baked sequence of frames. UDB cycles through them, repeating until the board's duration elapses.

```go
func (b *MyAnimatedBoard) GetType() types.BoardType { return types.BoardTypeAnimated }

func (b *MyAnimatedBoard) Init(cfg json.RawMessage, _ types.Datasource[any], dims types.BoardDimensions) error {
    // pre-build all frames using dims
    return nil
}

func (b *MyAnimatedBoard) Render() types.Animation {
    // types.Animation is []AnimationFrame
    return []types.AnimationFrame{
        {Img: b.frame1, Duration: 100 * time.Millisecond},
        {Img: b.frame2, Duration: 100 * time.Millisecond},
    }
}
```

#### Dynamic Board

Called repeatedly on the render loop. Each call returns one frame; the frame's `Duration` controls how long to wait before calling again. Use this for live-updating data like scores or clocks.

```go
func (b *MyDynamicBoard) GetType() types.BoardType { return types.BoardTypeDynamic }

func (b *MyDynamicBoard) Init(cfg json.RawMessage, ds types.Datasource[any], dims types.BoardDimensions) error {
    b.datasource = ds
    b.width = dims.Width
    b.height = dims.Height
    // pre-compute layout (e.g. font size, drawing coordinates) using dims
    return nil
}

func (b *MyDynamicBoard) Render() types.AnimationFrame {
    data := b.datasource.GetData()
    img := image.NewRGBA(image.Rect(0, 0, b.width, b.height))
    // draw current state using data and pre-computed layout...
    return types.AnimationFrame{
        Img:      img,
        Duration: time.Second, // call Render() again after 1 second
    }
}
```

### Full Static Board Example

```go
package boards

import (
    "encoding/json"
    "image"
    "image/color"
    "image/draw"

    "github.com/benwiebe/udb-plugin-library/types"
)

type SingleColourBoard struct {
    id          string
    colour      color.Color
    cachedImage image.Image
}

func NewSingleColourBoard(id string) *SingleColourBoard {
    return &SingleColourBoard{id: id, colour: color.White}
}

func (b *SingleColourBoard) GetId() string   { return b.id }
func (b *SingleColourBoard) GetName() string { return "Single Colour" }
func (b *SingleColourBoard) GetSupportedDimensions() []types.BoardDimensions { return nil }
func (b *SingleColourBoard) GetType() types.BoardType  { return types.BoardTypeStatic }
func (b *SingleColourBoard) GetDatasourceType() string { return "" }

func (b *SingleColourBoard) Init(cfg json.RawMessage, _ types.Datasource[any], dims types.BoardDimensions) error {
    if len(cfg) > 0 {
        var c struct{ Colour string `json:"colour"` }
        if err := json.Unmarshal(cfg, &c); err != nil {
            return err
        }
        if c.Colour != "" {
            b.colour = hexToColor(c.Colour)
        }
    }
    img := image.NewRGBA(image.Rect(0, 0, dims.Width, dims.Height))
    draw.Draw(img, img.Bounds(), &image.Uniform{b.colour}, image.Point{}, draw.Src)
    b.cachedImage = img
    return nil
}

func (b *SingleColourBoard) Render() image.Image {
    return b.cachedImage
}
```

## Implementing a Datasource

```go
type Datasource[T any] interface {
    GetId() string
    GetName() string
    GetType() string
    GetData() T
    Start(ctx context.Context) error
    DataChanged() <-chan struct{}
}
```

- **`GetType()`** — uniquely identifies what data this datasource provides. Must match the `GetDatasourceType()` of any board that uses it. See [Type String Convention](#type-string-convention).
- **`GetData()`** — returns the current cached data. Must be non-blocking. Never do network I/O here.
- **`Start(ctx)`** — called once at startup before any board is initialized. Launch your background fetch goroutine here. Use `ctx` for cancellation — when the app shuts down (SIGINT/SIGTERM), the context is cancelled and your goroutine should exit. Return a non-nil error to signal that the datasource cannot start; it will be removed from the map and not wired to any board.
- **`DataChanged()`** — return a channel that your datasource sends to whenever new data arrives and an immediate re-render is warranted. The scheduler selects on this channel alongside the frame timer, so a signal causes the board to re-render without waiting for the next tick. Return `nil` if you don't need push notifications — a nil channel blocks forever in a `select`, which is the correct no-op.

### Datasource with Background Refresh

```go
type WeatherDatasource struct {
    mu      sync.RWMutex
    data    WeatherData
    changed chan struct{}
}

func NewWeatherDatasource() *WeatherDatasource {
    return &WeatherDatasource{changed: make(chan struct{}, 1)}
}

func (d *WeatherDatasource) GetType() string { return "yourname/my-plugin/weather" }

func (d *WeatherDatasource) GetData() WeatherData {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.data
}

func (d *WeatherDatasource) Start(ctx context.Context) error {
    go func() {
        ticker := time.NewTicker(5 * time.Minute)
        defer ticker.Stop()
        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                data, err := fetchWeatherFromAPI()
                if err != nil {
                    continue
                }
                d.mu.Lock()
                d.data = data
                d.mu.Unlock()
                // Non-blocking send: if the channel is already full, a re-render
                // is already pending and we don't need to queue another.
                select {
                case d.changed <- struct{}{}:
                default:
                }
            }
        }
    }()
    return nil
}

func (d *WeatherDatasource) DataChanged() <-chan struct{} { return d.changed }
```

`GetData()` always returns immediately with the most recently cached value — it is called on the render path and must never block. The background goroutine started in `Start()` handles all I/O and exits cleanly when `ctx` is cancelled.

### Datasource without Push Notifications

If your datasource's data is always fresh on demand (e.g. reading from an in-process clock), implement no-op versions:

```go
func (d *MyDatasource) Start(_ context.Context) error { return nil }
func (d *MyDatasource) DataChanged() <-chan struct{}   { return nil }
```

## Type String Convention

Type strings connect boards to datasources. The convention is:

```
"author/plugin-name/type-name"
```

For example:
- `"benwiebe/nhl-plugin/game-data"`
- `"yourname/weather-plugin/current-conditions"`

This namespacing prevents collisions when multiple plugins are loaded. Use your GitHub username or organisation name as the author component.

A board that does not need a datasource returns `""` from `GetDatasourceType()`.

## Publishing Your Plugin

Publish your plugin as a normal Go module on GitHub (or any Go module proxy). Users include it by adding a blank import to their UDB binary:

```go
// In plugin_imports.go
import (
    _ "github.com/yourname/my-udb-plugin"
)
```

When using **udb-builder**, users simply select your plugin by module path and the builder handles the import and compilation.

## Config.json Wiring

If your plugin needs configuration (API keys, settings), users add an entry to `plugins` in `config.json`. The `id` must match your plugin's `GetId()`:

```json
{
  "plugins": [
    { "id": "my-plugin", "config": { "api_key": "..." } }
  ],
  "boards": [
    {
      "plugin": "my-plugin",
      "boardId": "my-board",
      "durationSeconds": 15,
      "config": { "colour": "#FF6600" }
    }
  ]
}
```

If your plugin needs no configuration, the `plugins` entry can be omitted entirely.

If your board needs a datasource, declare it and reference it:

```json
{
  "datasources": [
    {
      "id": "my-data",
      "plugin": "my-plugin",
      "datasourceId": "my-datasource"
    }
  ],
  "boards": [
    {
      "plugin": "my-plugin",
      "boardId": "my-board",
      "durationSeconds": 30,
      "datasource": "my-data"
    }
  ]
}
```

## Checklist

- [ ] Normal Go package (not `package main`); exports nothing special
- [ ] `init()` calls `udb_plugin_library.Register(&MyPlugin{})`
- [ ] `GetId()` returns a stable, unique string (matches `id` in `config.json` if configuration is needed)
- [ ] Board `GetDatasourceType()` matches datasource `GetType()` exactly (or returns `""`)
- [ ] `Board.Init()` accepts `dimensions BoardDimensions` and pre-computes all layout values that depend on display size
- [ ] `Board.Render()` takes no parameters — uses values pre-computed in `Init()`
- [ ] `Datasource.Start(ctx)` launches any background goroutine and returns immediately; goroutine exits when `ctx` is cancelled
- [ ] `Datasource.DataChanged()` returns a channel (buffered, size 1) if push notifications are needed, or `nil` otherwise
- [ ] `GetData()` never blocks — background goroutine handles fetching
- [ ] `Render()` only constructs images from cached data — no I/O on the render path
