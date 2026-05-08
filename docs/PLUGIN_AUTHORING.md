# Plugin Authoring Guide

A UDB plugin is a Go shared library (`.so` file) that provides boards, datasources, or both. This guide walks through building one from scratch.

## Prerequisites

- Go 1.21 or later
- Linux (Go's `plugin` package only supports runtime loading on Linux)
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

Your project must be `package main` and must export a `Plugin` variable. This is how UDB's runtime finds your plugin entry point.

## Plugin Entry Point

```go
package main

import (
    library "github.com/benwiebe/udb-plugin-library"
    "github.com/benwiebe/udb-plugin-library/types"
)

// Plugin is the exported symbol UDB looks for when loading your .so.
var Plugin library.UdbPlugin = &MyPlugin{}

type MyPlugin struct{}

func (p *MyPlugin) GetId() string                      { return "my-plugin" }
func (p *MyPlugin) GetName() string                    { return "My Plugin" }
func (p *MyPlugin) GetPluginType() types.PluginType    { return types.PluginTypeBoards }
func (p *MyPlugin) Configure(cfg types.PluginConfig) error { return nil }
```

`GetId()` must match the `id` field in the user's `config.json` plugin entry.

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
    Init(config json.RawMessage, datasource Datasource[T]) error
}
```

- **`GetSupportedDimensions()`** — return an empty slice to indicate the board handles any size. Otherwise list the panel dimensions you natively support (e.g. `{Width: 64, Height: 32}`). UDB uses this for future auto-scaling.
- **`GetDatasourceType()`** — return a type string matching your datasource's `GetType()`, or `""` if the board needs no datasource. See [Type String Convention](#type-string-convention).
- **`Init()`** — parse your board config from `config` (JSON) and store the datasource for use in `Render()`. Return a non-nil error to decline initialization (the board will be skipped with a warning).

### Board Types

#### Static Board

Renders once and holds the image for the configured duration.

```go
type MyStaticBoard struct {
    colour color.Color
}

func (b *MyStaticBoard) GetType() types.BoardType { return types.BoardTypeStatic }

func (b *MyStaticBoard) Render(dims types.BoardDimensions) image.Image {
    img := image.NewRGBA(image.Rect(0, 0, dims.Width, dims.Height))
    // draw to img...
    return img
}
```

#### Animated Board

Returns a full pre-baked sequence of frames. UDB cycles through them, repeating until the board's duration elapses.

```go
func (b *MyAnimatedBoard) GetType() types.BoardType { return types.BoardTypeAnimated }

func (b *MyAnimatedBoard) Render(dims types.BoardDimensions) types.Animation {
    // types.Animation is []AnimationFrame
    return []types.AnimationFrame{
        {Img: frame1, Duration: 100 * time.Millisecond},
        {Img: frame2, Duration: 100 * time.Millisecond},
    }
}
```

#### Dynamic Board

Called repeatedly on the render loop. Each call returns one frame; the frame's `Duration` controls how long to wait before calling again. Use this for live-updating data like scores or clocks.

```go
func (b *MyDynamicBoard) GetType() types.BoardType { return types.BoardTypeDynamic }

func (b *MyDynamicBoard) Render(dims types.BoardDimensions) types.AnimationFrame {
    data := b.datasource.GetData()
    img := image.NewRGBA(image.Rect(0, 0, dims.Width, dims.Height))
    // draw current state using data...
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
    id     string
    colour color.Color
}

func NewSingleColourBoard(id string) *SingleColourBoard {
    return &SingleColourBoard{id: id, colour: color.White}
}

func (b *SingleColourBoard) GetId() string   { return b.id }
func (b *SingleColourBoard) GetName() string { return "Single Colour" }
func (b *SingleColourBoard) GetSupportedDimensions() []types.BoardDimensions { return nil }
func (b *SingleColourBoard) GetType() types.BoardType        { return types.BoardTypeStatic }
func (b *SingleColourBoard) GetDatasourceType() string       { return "" }

func (b *SingleColourBoard) Init(cfg json.RawMessage, _ types.Datasource[any]) error {
    if len(cfg) > 0 {
        var c struct{ Colour string `json:"colour"` }
        if err := json.Unmarshal(cfg, &c); err != nil {
            return err
        }
        if c.Colour != "" {
            b.colour = hexToColor(c.Colour)
        }
    }
    return nil
}

func (b *SingleColourBoard) Render(dims types.BoardDimensions) image.Image {
    img := image.NewRGBA(image.Rect(0, 0, dims.Width, dims.Height))
    draw.Draw(img, img.Bounds(), &image.Uniform{b.colour}, image.Point{}, draw.Src)
    return img
}
```

## Implementing a Datasource

```go
type Datasource[T any] interface {
    GetId() string
    GetName() string
    GetType() string
    GetData() T
}
```

- **`GetType()`** — uniquely identifies what data this datasource provides. Must match the `GetDatasourceType()` of any board that uses it. See [Type String Convention](#type-string-convention).
- **`GetData()`** — returns the current cached data. Must be non-blocking. Never do network I/O here.

### Datasource with Background Refresh

```go
type WeatherDatasource struct {
    mu   sync.RWMutex
    data WeatherData
}

// GetType identifies the data contract for board matching.
func (d *WeatherDatasource) GetType() string { return "yourname/my-plugin/weather" }

func (d *WeatherDatasource) GetData() WeatherData {
    d.mu.RLock()
    defer d.mu.RUnlock()
    return d.data
}

// StartFetching launches the background refresh loop.
// Call this from the plugin's Configure() method.
func (d *WeatherDatasource) StartFetching(interval time.Duration) {
    go func() {
        for range time.Tick(interval) {
            data, err := fetchWeatherFromAPI()
            if err != nil {
                continue
            }
            d.mu.Lock()
            d.data = data
            d.mu.Unlock()
        }
    }()
}
```

The datasource manages its own goroutine. `GetData()` always returns immediately with the most recently cached value — it is called on every render frame and must never block.

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

## Building the Plugin

```bash
go build -buildmode=plugin -o my-plugin.so .
```

The output `.so` must be built with the same Go version as `udb-core`. Place the `.so` at `./plugins/my-plugin/my-plugin.so` relative to the UDB working directory, or specify a `path` in config.

> **Important**: Plugins must be compiled on Linux. macOS has partial support for `plugin` mode but runtime loading is unreliable. Windows is not supported.

## Config.json Wiring

Once your plugin is built and placed, wire it up in `config.json`:

```json
{
  "plugins": [
    { "id": "my-plugin", "path": "./plugins/my-plugin.so" }
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

If your board needs a datasource, declare it and reference it:

```json
{
  "datasources": [
    {
      "id": "my-data",
      "plugin": "my-plugin",
      "datasourceId": "my-datasource",
      "config": { "refreshInterval": "60s" }
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

- [ ] `package main` with an exported `Plugin` variable of type `UdbPlugin`
- [ ] `GetId()` returns the same string used as `id` in `config.json`
- [ ] Board `GetDatasourceType()` matches datasource `GetType()` exactly (or returns `""`)
- [ ] `GetData()` never blocks — background goroutine handles fetching
- [ ] `Render()` only constructs images from cached data — no I/O on the render path
- [ ] Built with `go build -buildmode=plugin` on Linux, same Go version as `udb-core`
