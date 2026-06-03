# Roadmap

This document covers planned post-MVP enhancements. The MVP scope is a simple round-robin display loop on a fixed display size — enough to validate the full plugin contract end-to-end and build useful real-world boards.

---

## Multi-Size Display Support & Auto-Scaling

Displays come in many sizes (32×64, 64×64, 64×128, chained panels, etc.). The planned strategy:

- **Plugin-handled**: if the active display matches one of the board's `GetSupportedDimensions()` entries, call `Render()` and use the result directly — the board drew at native resolution.
- **Core auto-scaling fallback**: if the display size is not in the board's supported list, UDB core renders at the nearest supported size and scales the output image to fit. Default: nearest-neighbor (correct for pixel-art LED panels). Bilinear available as an opt-in.
- Boards authored for a "canonical" size (e.g. 64×32) get reasonable auto-scale on other panels for free.

**Config addition needed:** `display.scaling_mode` — `"nearest"` | `"bilinear"` | `"none"` (default `"nearest"`).

---

## Advanced Scheduling

The MVP uses round-robin rotation. Post-MVP, the scheduler becomes significantly smarter.

### Time-Based Scheduling

Boards (or rotation groups) can declare operating hours:

```json
{
  "board_id": "clock",
  "schedule": { "active_hours": "07:00-23:00" }
}
```

Display-off support:
```json
{
  "display": { "off_hours": "23:00-07:00" }
}
```

### Conditional Rotation Groups

The config defines multiple **rotations** (ordered lists of boards), each with an activation condition. A condition is evaluated by the datasource — it returns a named state string. The scheduler picks the highest-priority rotation whose condition is currently true, falling back to the default rotation.

```json
{
  "rotations": [
    {
      "id": "default",
      "boards": ["clock", "weather", "flights"],
      "condition": null
    },
    {
      "id": "nhl-intermission",
      "boards": ["clock", "weather", "nhl-stats"],
      "condition": { "datasource_id": "nhl-live", "state": "intermission" }
    },
    {
      "id": "nhl-live-period",
      "boards": ["nhl-scoreboard"],
      "condition": { "datasource_id": "nhl-live", "state": "in_period" },
      "priority": 10
    }
  ]
}
```

This requires **state reporting** in the datasource interface: an optional `GetState() string` method the scheduler can call to evaluate conditions. This is lightweight — a status string, not the full data payload — and distinct from `GetData()`.

---

## Alert Boards

Alert boards are not part of any rotation. They are triggered by an event from a datasource and displayed immediately, interrupting whatever is currently showing. After the alert duration, normal rotation resumes.

**Use cases:**
- NHL goal scored → red light celebration board
- Severe weather warning → urgent alert
- A flight of interest lands → pop-up notification

**Design:**
- Datasources signal alert events via a channel registered during `Start()`
- The core scheduler listens on an alert channel; when an event arrives, it preempts the current board
- Multiple simultaneous alerts are queued, not dropped
- Alerts have an optional **expiry**: if the alert timestamp is older than a threshold (e.g. 30 seconds), the core discards it — prevents stale alerts replaying after a reconnect

```json
{
  "alert_boards": [
    {
      "plugin_id": "nhl-plugin",
      "board_id": "goal-light",
      "duration": "8s",
      "datasource_id": "nhl-live"
    }
  ]
}
```

---

## Plugin-Defined Rotations ("Playlists")

Plugins can ship pre-built rotation suggestions — opinionated, curated orderings of their own boards for common use cases.

A plugin optionally implements `GetRotations()` returning named `PluginRotation` definitions:

```go
type PluginRotation struct {
    ID          string
    Name        string
    Description string
    Boards      []PluginRotationBoard
}

type PluginRotationBoard struct {
    BoardID   string
    Duration  time.Duration
    Condition string // only include if datasource state matches
}
```

Users reference a plugin rotation in config:
```json
{
  "rotations": [
    {
      "id": "nhl-gameday",
      "plugin_rotation": { "plugin_id": "nhl-plugin", "rotation_id": "nhl-gameday" },
      "condition": { "datasource_id": "nhl-live", "state": "game_day" },
      "priority": 10
    }
  ]
}
```

**Constraints:**
- Plugin rotations can only reference boards and datasources from their own plugin
- The user's config always has final say
- Missing boards/datasources in a rotation trigger a warning and are skipped gracefully

---

## Board Builder (Web Tool)

A web-based tool (separate repo, e.g. `udb-builder`) for creating a UDB setup without hand-editing JSON:

1. User selects display hardware (panel size, chained panels)
2. User browses a **plugin marketplace** or pastes a GitHub URL
3. For each plugin, the builder shows available boards/datasources and prompts for config values (sourced from a plugin-provided schema)
4. User arranges boards into a rotation, sets durations, optionally configures schedules and alert boards
5. Builder outputs a `config.json` or a pre-built disk image for a fresh Raspberry Pi

**Plugin marketplace:** a central registry (GitHub-hosted JSON index) listing plugins with metadata. Plugins self-submit by opening a PR to the registry repo.

---

## Other Enhancements

| Feature | Description |
|---------|-------------|
| Hot-reload config | Watch `config.json` for changes and reload without restart |
| Plugin hot-reload | Reload a plugin `.so` without restarting the whole app |
| Multiple displays | Drive more than one LED panel simultaneously |
| Startup splash | Show a boot animation while datasources initialize |
| Animation caching | Cache pre-rendered `AnimatedBoard` frames in RAM |
| Multi-zone display | Split a large panel into regions, each showing a different board |
| Plugin sandboxing | Run plugins in a separate process with an RPC boundary for stability isolation |
| Plugin versioning | Compatibility checks between plugin-library version and plugin build |
| gRPC/HTTP datasources | Built-in adapter so datasources can run on a separate machine |
| Additional hardware | Support e-ink, OLED, etc. via the `Display` interface |
| Animation streaming | `AnimatedBoard.Render()` returns frames one at a time rather than the full slice, capping peak memory for long animations |
| Canvas pooling | Reuse canvas allocations across frames rather than allocating per-render |
