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
      (hub75 /   (registry)   └──► StartDatasources()
       http /                 └──► LoadBoards()
       stub)                  └──► WireDatasources()
          │                   └──► InitBoards()
          │                            │
          └────────────────────────────┘
                        │
                        ▼
                  Scheduler.Run()
                  (round-robin loop)
```

## Components

### Config

`internal/config` parses `config.json` into a `RootConfig` struct. The loader is straightforward — no merging, no environment variable substitution, no defaults beyond what Go's zero values provide. The config file is the single source of truth.

### Plugin System

UDB uses a **statically-linked plugin model**. Plugins are normal Go packages compiled into the binary at build time. There are no `.so` files, no runtime dynamic loading, and no ABI compatibility requirements.

#### How plugins register

Plugins call `udb_plugin_library.Register()` from their `init()` function:

```go
// In the plugin package
func init() {
    udb_plugin_library.Register(&MyPlugin{})
}
```

The binary includes plugins via blank imports in `plugin_imports.go`:

```go
import (
    _ "github.com/benwiebe/udb-plugin-nhl"
    _ "github.com/benwiebe/udb-plugin-weather"
)
```

A blank import triggers the package's `init()`, which calls `Register()`. By the time `main()` runs, the registry contains all included plugins.

#### How the core loads them

`LoadPlugins()` in `internal/plugins/pluginloader.go` calls `udb_plugin_library.Registered()` to get all registered plugins. It then matches each plugin against the `plugins` config block by `GetId()`, calls `Configure()` with the matching config (or `nil` for plugins that need none), and builds the `PluginData` map used by the rest of startup.

The `plugins` config block exists solely for per-plugin credentials and settings (API keys, etc.). Plugins that need no configuration can be omitted from it entirely — they are still available as long as they are imported in `plugin_imports.go`.

#### Distribution and custom builds

End users download a pre-built binary from GitHub Releases that includes a curated set of official plugins. Users who want a different set of plugins use **udb-builder** (a separate CLI tool) to compile a binary with their chosen plugins on their own machine. The Pi never compiles anything. See [ROADMAP.md](ROADMAP.md) for the udb-builder plan.

The `plugin_imports.go` file is managed by udb-builder. Developers can also edit it manually.

### Datasource Initialization

`LoadDatasources()` walks the `datasources` config block, looks up the corresponding plugin by ID from the `PluginData` map, retrieves the named datasource instance, and returns a map keyed by the user-defined datasource `id`. These IDs are how boards reference their data source in config.

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
init() functions run (triggered by blank imports in plugin_imports.go)
    │  each plugin calls udb_plugin_library.Register()
    ▼
LoadPlugins() iterates Registered(), calls Configure() on each
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

## Design Decisions

**Go generics (`[T any]`)** — used throughout the plugin library to keep board/datasource pairs type-safe without requiring plugins to use `interface{}` or reflection. The actual data type `T` only needs to match between a paired board and datasource; the core treats everything as `any`.

**Static linking over dynamic `.so` loading** — Go's `plugin` package requires plugins to be compiled with the exact same Go version and toolchain as the host binary, is Linux-only, and provides no real isolation. Since builds were already coupled in practice, we lean into that: everything is one build. ABI mismatch, toolchain pinning, and cross-compilation complexity all disappear. The tradeoff is that adding a new plugin requires a rebuild, which udb-builder handles transparently.

**`plugin_imports.go` as the plugin manifest** — blank imports in a single file make the set of included plugins explicit and diffable. udb-builder generates this file; developers can also edit it manually. The file is committed so the binary's plugin set is reproducible from source.

**`go.work` workspace** — `udb-core`, `udb-plugin-library`, and plugins under active development are linked via a Go workspace, avoiding the need to push and tag releases during development. Production builds (via udb-builder) use tagged releases via `go get`.

**Alert expiry** — when the scheduler adds alert board support, alert events must carry a timestamp. A datasource reconnecting after a drop should not re-fire stale events; the core discards any alert older than a configurable threshold before displaying it.

**Conditional rotation state** — the datasource state used for scheduler conditions (`GetState() string`) is intentionally a lightweight string enum, not the full data object from `GetData()`. The scheduler should not need to understand plugin-specific data types to evaluate routing conditions.
