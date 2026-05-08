# UDB Architecture

## Overview

UDB is structured around three concerns: **configuration**, **plugins** (boards + datasources), and **display**. The core runtime wires them together and drives a display loop.

```
config.json
    │
    ▼
ConfigLoader ──► RootConfig
                     │
          ┌──────────┼──────────┐
          ▼          ▼          ▼
    DisplayConfig  PluginsConfig  BoardsConfig / DatasourcesConfig
          │          │                    │
          ▼          ▼                    ▼
      Display    LoadPlugins()     LoadDatasources()
      (hub75 /               └──► StartDatasources()
       http /                └──► LoadBoards()
       stub)                 └──► WireDatasources()
          │                  └──► InitBoards()
          │                          │
          └──────────────────────────┘
                        │
                        ▼
                  Scheduler.Run()
                  (round-robin loop)
```

## Components

### Config

`internal/config` parses `config.json` into a `RootConfig` struct. The loader is straightforward — no merging, no environment variable substitution, no defaults beyond what Go's zero values provide. The config file is the single source of truth.

### Plugin Loading

`internal/plugins/pluginloader.go` uses Go's `plugin` package to load `.so` files at runtime. Each `.so` must export a `Plugin` symbol of type `UdbPlugin`. The loader calls `Configure()` on each plugin immediately after loading, passing the plugin-level JSON config block.

Plugins live at `./plugins/{id}/{id}.so` by default, or at a path specified explicitly in config.

### Datasource Initialization

`LoadDatasources()` walks the `datasources` config block, looks up the corresponding plugin, retrieves the named datasource instance, and returns a map keyed by the user-defined datasource `id`. These IDs are how boards reference their data source in config.

### Board Setup and Wiring

`LoadBoards()` retrieves board instances from their respective plugins. `WireDatasources()` then attaches a datasource to each board using one of two strategies:

1. **Explicit**: the board config specifies a `datasource` ID — the previously-loaded datasource with that ID is used directly.
2. **Auto-match**: if no datasource ID is given, the core compares `board.GetDatasourceType()` against `datasource.GetType()` across all loaded datasources and picks the one that matches. A warning is logged if zero or multiple matches are found.

`StartDatasources()` calls `datasource.Start(ctx)` on each loaded datasource before wiring. Datasources that return an error from `Start()` are removed from the map so they cannot be wired to boards.

`InitBoards()` calls `board.Init(config, datasource, dimensions)` on each board, passing the display dimensions alongside the config and datasource. Boards receive dimensions here so they can pre-compute layout (font sizes, image buffers, etc.) once at startup rather than on every render call. Boards that fail to initialize are dropped with a warning.

### Scheduler

`internal/scheduler/scheduler.go` runs a simple round-robin loop over the initialized boards. The behavior per board type:

| Board Type | Render Strategy |
|------------|----------------|
| **Static** | `Render()` called once; result held on display for `durationSeconds` |
| **Animated** | `Render()` returns a full `[]AnimationFrame`; frames are cycled in order until duration elapses (or indefinitely if `durationSeconds` is 0) |
| **Dynamic** | `Render()` is called repeatedly in a tight loop; each call returns one `AnimationFrame` whose `.Duration` controls the sleep between frames; continues until duration elapses |

For dynamic boards, the scheduler also listens on the datasource's `DataChanged()` channel. If the datasource signals a change, the board is re-rendered immediately rather than waiting for the frame timer to expire. Datasources that return `nil` from `DataChanged()` opt out of this — a nil channel blocks forever in a `select`, which is the correct no-op.

The scheduler respects context cancellation — when `SIGINT` or `SIGTERM` is received, the context is cancelled and the loop exits cleanly, allowing `Display.CloseDisplay()` to release hardware resources.

### Display

`internal/display` defines a two-method interface:

```go
type Display interface {
    Render(img image.Image) error
    CloseDisplay()
}
```

Implementations are registered at init time via build tags and a registry map, so the correct implementation is selected based on the `type` field in config (defaulting to `hub75` on Linux, `stub` elsewhere).

| Type | Platform | Description |
|------|----------|-------------|
| `hub75` | Linux only | Real hardware via `tfk1410/go-rpi-rgb-led-matrix` CGo binding |
| `http` | Any | MJPEG stream at `http://localhost:8080`; useful for dev |
| `stub` | Any | No-op; logs render calls to stdout |

## Plugin Contract

Plugins are the extension point. The full interface contract is defined in `udb-plugin-library`. See [PLUGIN_AUTHORING.md](PLUGIN_AUTHORING.md) for a guide to writing one.

The key invariant: **boards and datasources are matched by type string**. A board declares what datasource type it needs via `GetDatasourceType()`; a datasource declares what it provides via `GetType()`. The convention for these strings is `"author/plugin-name/type-name"` — namespacing prevents collisions between plugins from different authors.

## Data Flow at Runtime

```
startup
    │
    ▼
datasource.Start(ctx)          ← datasource launches its background fetch goroutine here;
    │                            goroutine exits when ctx is cancelled (SIGINT / SIGTERM)
    │
    ├── (periodic) fetch from network / file / etc. → cache internally
    │
    └── (optional) signal datasource.DataChanged()  ← triggers immediate re-render
            │
            ▼
board.Init(config, datasource, dimensions)  ← board pre-computes layout (font sizes,
    │                                          image buffers) using the display dimensions
    ▼
── render loop ──────────────────────────────────────────────────────
    │
    ▼
datasource.GetData()  ◄── called by board inside Render(); must return cached value instantly
    │
    ▼
board.Render()             ← no dimensions parameter; uses values pre-computed in Init()
    │  constructs image.Image from cached data
    ▼
display.Render(img)
    │  pushes pixels to hardware / HTTP stream / stdout
    ▼
LED panel / browser / log
```

Datasources are responsible for their own background fetch loop, started in `Start()`. `GetData()` must return a cached value immediately — it is called on the render path and must never block on I/O.
