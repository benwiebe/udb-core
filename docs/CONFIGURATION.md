# Configuration Reference

UDB is configured via `config.json` in the working directory. You can specify an alternate path by setting a custom config loader path in your startup script.

## Top-Level Structure

```json
{
  "display":     { ... },
  "plugins":     [ ... ],
  "datasources": [ ... ],
  "boards":      [ ... ]
}
```

---

## `display`

Controls the physical (or virtual) display.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `type` | string | No | `hub75` on Linux, `stub` elsewhere | Display backend to use. See [Display Types](#display-types). |
| `width` | int | Yes | — | Panel width in pixels |
| `height` | int | Yes | — | Panel height in pixels |
| `brightness` | int | No | Library default (~100) | Brightness 0–100. Hardware display only. |
| `gpio_mapping` | string | No | `"regular"` | GPIO wiring layout. See [GPIO Mappings](#gpio-mappings). |
| `scale` | int | No | `1` | Pixel scale multiplier for the `http` display. A value of `4` makes each LED pixel 4×4 in the browser preview. |

### Display Types

| Value | Platform | Description |
|-------|----------|-------------|
| `hub75` | Linux only | Real Hub75 LED matrix hardware |
| `http` | Any | MJPEG stream at `http://localhost:8080`. Use `scale` to get a better feel for the blocky LED aesthetic. |
| `stub` | Any | No-op; logs render calls to stdout |

### GPIO Mappings

These correspond directly to the `HardwareMapping` values in [hzeller/rpi-rgb-led-matrix](https://github.com/hzeller/rpi-rgb-led-matrix):

| Value | Description |
|-------|-------------|
| `regular` | Direct GPIO wiring (default) |
| `adafruit-hat` | Adafruit RGB Matrix Bonnet / HAT |
| `adafruit-hat-pwm` | Adafruit RGB Matrix Bonnet with PWM (better brightness control) |
| `regular-pi1` | Original Pi 1 pin layout |
| `classic` | Older wiring variant |
| `classic-pi1` | Older wiring, Pi 1 |

### Example

```json
"display": {
  "type": "hub75",
  "width": 64,
  "height": 32,
  "brightness": 80,
  "gpio_mapping": "adafruit-hat-pwm"
}
```

Development example with pixel scale:

```json
"display": {
  "type": "http",
  "width": 64,
  "height": 32,
  "scale": 8
}
```

---

## `plugins`

An array of plugins to load. Each plugin is a compiled `.so` file.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Unique identifier for this plugin. Used to reference the plugin from boards and datasources. Must match `GetId()` in the plugin. |
| `path` | string | No | Path to the `.so` file. If omitted, defaults to `./plugins/{id}/{id}.so`. |
| `config` | object | No | Arbitrary JSON passed to the plugin's `Configure()` method at load time. |

### Example

```json
"plugins": [
  {
    "id": "nhl-plugin",
    "path": "./plugins/nhl-plugin.so",
    "config": { "apiKey": "..." }
  },
  {
    "id": "clock-plugin"
  }
]
```

---

## `datasources`

Named datasource instances. Boards reference these by the `id` you give them here.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Your name for this datasource instance. Referenced by boards via the `datasource` field. |
| `plugin` | string | Yes | The `id` of the plugin that provides this datasource. |
| `datasourceId` | string | Yes | The datasource identifier within the plugin (from `GetDatasourceMap()`). |
| `config` | object | No | Arbitrary JSON passed to the datasource during `Init()`. |

### Example

```json
"datasources": [
  {
    "id": "wpg-jets",
    "plugin": "nhl-plugin",
    "datasourceId": "nhl-live-scores",
    "config": { "team": "WPG" }
  }
]
```

---

## `boards`

The ordered list of boards to display. UDB cycles through them in sequence.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `plugin` | string | Yes | The `id` of the plugin that provides this board. |
| `boardId` | string | Yes | The board identifier within the plugin (from `GetBoardMap()`). |
| `durationSeconds` | int | No | How long to show this board before advancing to the next. `0` or omitted behaves differently per board type — see below. |
| `config` | object | No | Arbitrary JSON passed to the board during `Init()`. |
| `datasource` | string | No | The `id` of the datasource to wire to this board. If omitted, the core will attempt to auto-match by type. |

### durationSeconds = 0 behavior by board type

| Board type | Behavior when `durationSeconds` is 0 or omitted |
|------------|--------------------------------------------------|
| `static` | Renders once then immediately advances to the next board. The image is shown for only one display cycle before the loop moves on. |
| `animated` | Plays through all frames exactly once, then advances to the next board. |
| `dynamic` | Loops indefinitely, calling `Render()` on each frame tick, until the process is stopped. Only useful if it's the only board in the list. |

For `static` and `animated` boards, set `durationSeconds` to control how long they hold before advancing.

### Example

```json
"boards": [
  {
    "plugin": "nhl-plugin",
    "boardId": "scoreboard",
    "durationSeconds": 30,
    "datasource": "wpg-jets"
  },
  {
    "plugin": "clock-plugin",
    "boardId": "clock",
    "durationSeconds": 10
  }
]
```

---

## Datasource Auto-Matching

If a board's `datasource` field is omitted, UDB compares the board's required datasource type (from `board.GetDatasourceType()`) against all loaded datasources' types (from `datasource.GetType()`). If exactly one match is found, it is used automatically. If zero or multiple matches are found, a warning is logged — provide an explicit `datasource` field to resolve ambiguity.

---

## Complete Example

```json
{
  "display": {
    "type": "hub75",
    "width": 64,
    "height": 32,
    "brightness": 80,
    "gpio_mapping": "adafruit-hat-pwm"
  },
  "plugins": [
    {
      "id": "nhl-plugin",
      "path": "./plugins/nhl-plugin.so"
    },
    {
      "id": "clock-plugin"
    }
  ],
  "datasources": [
    {
      "id": "wpg-jets",
      "plugin": "nhl-plugin",
      "datasourceId": "nhl-live-scores",
      "config": { "team": "WPG" }
    }
  ],
  "boards": [
    {
      "plugin": "nhl-plugin",
      "boardId": "scoreboard",
      "durationSeconds": 30,
      "datasource": "wpg-jets"
    },
    {
      "plugin": "clock-plugin",
      "boardId": "clock",
      "durationSeconds": 10
    }
  ]
}
```
