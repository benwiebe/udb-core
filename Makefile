OUTPUT    := udb
REPO_ROOT := $(shell git rev-parse --show-toplevel)
LIB_DIR   := $(REPO_ROOT)/build/libs/rpi-rgb-led-matrix
UNAME_S   := $(shell uname -s)

# CGo flags are only needed on Linux where the Hub75 display code is compiled.
# On other platforms the hub75display.go build tag is inactive and these are unused.
ifeq ($(UNAME_S),Linux)
export CGO_CFLAGS  := -I$(LIB_DIR)/include
export CGO_LDFLAGS := -L$(LIB_DIR)/lib -lrgbmatrix -lstdc++ -lm
endif

# Space-separated list of plugin module paths (with optional @version) to include.
# When set, plugin_imports.go is regenerated before building.
# When unset, whatever is already in plugin_imports.go is used.
#
# Examples:
#   make build PLUGINS="github.com/benwiebe/udb-plugin-nhl"
#   make build PLUGINS="github.com/benwiebe/udb-plugin-nhl@v1.2.0 github.com/benwiebe/udb-plugin-weather"
PLUGINS ?=

# ── Cross-compilation targets ─────────────────────────────────────────────────
# Requires Zig: brew install zig
#
# Pi 3 / Pi 4 / Pi 5 / Pi Zero 2 W  →  make build-pi   (arm64, 64-bit OS)
# Pi Zero / Pi Zero W                →  make build-pi-zero  (armv6, 32-bit OS)
#
# PLUGINS= works the same as for the native build target.

ZIG_TARGET_ARM64 := aarch64-linux-gnu
ZIG_TARGET_ARMV6 := armv6-linux-gnueabihf

LIB_DIR_ARM64 := $(REPO_ROOT)/build/libs/rpi-rgb-led-matrix-$(ZIG_TARGET_ARM64)
LIB_DIR_ARMV6 := $(REPO_ROOT)/build/libs/rpi-rgb-led-matrix-$(ZIG_TARGET_ARMV6)

.PHONY: all build build-pi build-pi-arm64 build-pi-zero \
        test clean deps tidy setup-lib \
        setup-cross-lib-arm64 setup-cross-lib-armv6 \
        fmt vet

all: tidy deps build

# ── Native build ──────────────────────────────────────────────────────────────

build: setup-lib
ifneq ($(PLUGINS),)
	./scripts/build-with-plugins.sh $(PLUGINS)
endif
	go build -v -o $(OUTPUT) .

test: setup-lib
	go test -v ./...

# Clone and compile the rpi-rgb-led-matrix C library (Linux only).
# The script is idempotent: skips cloning if the repo is already present.
setup-lib:
ifeq ($(UNAME_S),Linux)
	./scripts/setup_matrix_library.sh
else
	@echo "Skipping LED matrix library setup (not Linux)"
endif

# ── Cross-compile for Raspberry Pi ───────────────────────────────────────────

# Pi 3, Pi 4, Pi 5, Pi Zero 2 W — 64-bit arm64 OS (most common modern target)
build-pi: build-pi-arm64

build-pi-arm64: setup-cross-lib-arm64
ifneq ($(PLUGINS),)
	./scripts/build-with-plugins.sh $(PLUGINS)
endif
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 \
	CC="zig cc -target $(ZIG_TARGET_ARM64)" \
	CXX="zig c++ -target $(ZIG_TARGET_ARM64)" \
	CGO_CFLAGS="-I$(LIB_DIR_ARM64)/include" \
	CGO_LDFLAGS="-L$(LIB_DIR_ARM64)/lib -lrgbmatrix -lstdc++ -lm" \
	go build -v -o $(OUTPUT)-linux-arm64 .

# Pi Zero / Pi Zero W — 32-bit ARMv6 OS
build-pi-zero: setup-cross-lib-armv6
ifneq ($(PLUGINS),)
	./scripts/build-with-plugins.sh $(PLUGINS)
endif
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=1 \
	CC="zig cc -target $(ZIG_TARGET_ARMV6)" \
	CXX="zig c++ -target $(ZIG_TARGET_ARMV6)" \
	CGO_CFLAGS="-I$(LIB_DIR_ARMV6)/include" \
	CGO_LDFLAGS="-L$(LIB_DIR_ARMV6)/lib -lrgbmatrix -lstdc++ -lm" \
	go build -v -o $(OUTPUT)-linux-armv6 .

setup-cross-lib-arm64:
	./scripts/setup-cross-compile-lib.sh $(ZIG_TARGET_ARM64)

setup-cross-lib-armv6:
	./scripts/setup-cross-compile-lib.sh $(ZIG_TARGET_ARMV6)

# ── Utilities ─────────────────────────────────────────────────────────────────

deps:
	go mod download
	go mod verify

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

clean:
	rm -f $(OUTPUT) $(OUTPUT)-linux-arm64 $(OUTPUT)-linux-armv6
	rm -rf ./build/
