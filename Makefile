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

.PHONY: all build test clean deps tidy setup-lib fmt vet

all: tidy deps build

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
	rm -f $(OUTPUT)
	rm -rf ./build/
