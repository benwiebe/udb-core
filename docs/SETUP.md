# Setup Guide

## Hardware Requirements

| Component | Requirement |
|-----------|-------------|
| Raspberry Pi | Pi 4 (2 GB minimum; 4 GB recommended). Pi 3B+ works for simple boards. |
| LED Matrix | Hub75-compatible RGB panel (32×64, 64×64, 64×128, etc.) |
| Power supply | 5V, sufficient amperage for your panel size (typically 4–10A) |
| HAT (optional) | [Adafruit RGB Matrix Bonnet](https://www.adafruit.com/product/3211) or HAT simplifies wiring |
| OS | Raspberry Pi OS (64-bit recommended), Bookworm or later |

The Pi 5 also works and offers headroom for heavy plugins, but is not required.

## Development (No Hardware)

You can develop and test boards on any machine — macOS, Linux, or Windows (without plugin loading, which requires Linux) — using the `http` display type, which streams an MJPEG preview to a browser.

Set your display config to:

```json
"display": {
  "type": "http",
  "width": 64,
  "height": 32,
  "scale": 8
}
```

Then open `http://localhost:8080` in a browser after starting UDB. The `scale` field multiplies each LED pixel to better simulate the blocky appearance of real hardware.

Note: Go's `plugin` package requires Linux for loading `.so` files at runtime. Plugin development and testing requires Linux (or a Linux VM/container).

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git
- On Linux: GCC (for CGo when targeting `hub75` display)

```bash
# Install Go (if not already installed)
# See https://go.dev/dl/ for the latest release

# Clone the repo
git clone https://github.com/benwiebe/udb-core
cd udb-core
```

### Build the LED Matrix Library (Linux / Raspberry Pi only)

The `hub75` display backend links against [hzeller/rpi-rgb-led-matrix](https://github.com/hzeller/rpi-rgb-led-matrix), a C++ library that must be compiled locally.

```bash
./scripts/setup_matrix_library.sh
```

This clones and builds the library and places the static archive where the Go build can find it. Run it once after cloning.

### Build UDB

```bash
go build -o udb .
```

On macOS (no hardware), this builds without the hub75 backend — only `http` and `stub` display types are available.

## Installing Plugins

Plugins are compiled `.so` files. Place them in:

```
./plugins/{plugin-id}/{plugin-id}.so
```

For example:

```
./plugins/nhl-plugin/nhl-plugin.so
./plugins/clock-plugin/clock-plugin.so
```

Or specify an explicit `path` in the plugin config block. See [CONFIGURATION.md](CONFIGURATION.md) for details.

## Wiring

### Using an Adafruit RGB Matrix Bonnet (recommended)

The Bonnet routes all required GPIO pins to the Hub75 connector and handles power distribution. Slot it onto the Pi's GPIO header, connect your panel's data cable to the Bonnet's output connector, and power the panel separately via the Bonnet's barrel jack or screw terminals.

Set `"gpio_mapping": "adafruit-hat-pwm"` in your display config for best brightness control.

### Direct GPIO Wiring

If wiring directly to the Pi's GPIO header, follow the hzeller library's [wiring guide](https://github.com/hzeller/rpi-rgb-led-matrix/blob/master/wiring.md). Use `"gpio_mapping": "regular"` (or omit the field — it's the default).

## Running UDB

```bash
./udb
```

UDB looks for `config.json` in the current directory. Stop with `Ctrl+C`; the display is cleanly shut down on `SIGINT` or `SIGTERM`.

## Running as a Service

To start UDB automatically on boot, create a systemd service:

```ini
# /etc/systemd/system/udb.service
[Unit]
Description=Universal Display Board
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/udb
ExecStart=/home/pi/udb/udb
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Enable and start it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable udb
sudo systemctl start udb
```

View logs:

```bash
journalctl -u udb -f
```

## Optional: CPU Core Isolation

The hzeller library uses busy-polling on a CPU core to refresh the panel. Isolating core 3 for this thread prevents OS scheduling interference, reducing flicker under load:

Add `isolcpus=3` to `/boot/firmware/cmdline.txt` (on Bookworm; `/boot/cmdline.txt` on older releases). The file is a single line — append the parameter to the end:

```
... rootwait isolcpus=3
```

Reboot for it to take effect. This is optional but worth doing for a stable deployment.

## Audio Note

The hzeller library uses the same hardware subsystem as the Pi's onboard audio (`snd_bcm2835`). They cannot coexist. If you need audio output (e.g. for goal horn sounds from a plugin), use a USB audio device.
