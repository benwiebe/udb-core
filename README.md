# UDB — Universal Display Board

UDB is a plugin-based runtime for driving LED matrix displays (Hub75 RGB panels on Raspberry Pi). It decouples data sources from display rendering: boards define how to render, datasources define how to fetch data, and neither needs to know about the other.

## Features

- **Plugin-based** — boards and datasources are independently loadable `.so` files
- **Decoupled** — swap a broken data source without touching the display code, and vice versa
- **Fast** — written in Go
- **Extensible** — build plugins without touching the core
- **Dev-friendly** — HTTP MJPEG preview display for developing without hardware attached

## Quick Start

```bash
# 1. Build the LED matrix C library (Linux/Raspberry Pi only)
./scripts/setup_matrix_library.sh

# 2. Build udb-core
go build -o udb .

# 3. Place your plugin .so files in ./plugins/{id}/{id}.so
#    or specify a path in config.json

# 4. Configure your setup (see docs/CONFIGURATION.md)
cp config.json.example config.json
#    On macOS or without hardware, set display.type to "http" in config.json

# 5. Run
./udb
```

For hardware setup and Raspberry Pi configuration, see [docs/SETUP.md](docs/SETUP.md).

## Configuration

UDB is configured via `config.json` in the working directory. A minimal example:

```json
{
  "display": {
    "height": 32,
    "width": 64
  },
  "plugins": [
    { "id": "my-plugin", "path": "./plugins/my-plugin.so" }
  ],
  "boards": [
    {
      "plugin": "my-plugin",
      "boardId": "my-board",
      "durationSeconds": 30
    }
  ]
}
```

Full configuration reference: [docs/CONFIGURATION.md](docs/CONFIGURATION.md)

## Documentation

| Doc | Description |
|-----|-------------|
| [docs/SETUP.md](docs/SETUP.md) | Hardware requirements, Raspberry Pi setup, running as a service |
| [docs/CONFIGURATION.md](docs/CONFIGURATION.md) | Full `config.json` reference |
| [docs/PLUGIN_AUTHORING.md](docs/PLUGIN_AUTHORING.md) | How to write your own boards and datasources |
| [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) | How UDB works internally |
| [docs/ROADMAP.md](docs/ROADMAP.md) | Planned post-MVP enhancements |

## Repository Structure

| Repo | Purpose |
|------|---------|
| [`udb-core`](https://github.com/benwiebe/udb-core) | This repo — the runtime |
| [`udb-plugin-library`](https://github.com/benwiebe/udb-plugin-library) | Interface contracts every plugin must implement |

## Requirements

- **Hardware**: Raspberry Pi 4 (2 GB RAM minimum; 4 GB recommended) with a Hub75 LED matrix panel
- **OS**: Raspberry Pi OS (64-bit recommended)
- **Go**: 1.21 or later (for building from source)
- **Development** (no hardware): macOS or Linux with Go installed; use the `http` display type to preview output in a browser

## License

GPL-3.0 - See [LICENSE](LICENSE) for full license text.
