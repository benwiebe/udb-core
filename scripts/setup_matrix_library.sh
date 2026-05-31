#!/bin/sh

REPO_ROOT=$(git -C "$(dirname "$0")" rev-parse --show-toplevel)
LIB_DIR="$REPO_ROOT/build/libs/rpi-rgb-led-matrix"

if [ -d "$LIB_DIR/.git" ]; then
    echo "rpi-rgb-led-matrix already cloned, skipping"
else
    mkdir -p "$LIB_DIR"
    git clone https://github.com/hzeller/rpi-rgb-led-matrix.git "$LIB_DIR"
fi

sed -i.bak 's/.*RGB_SLOWDOWN_GPIO.*/DEFINE+=-DRGB_SLOWDOWN_GPIO=2/' "$LIB_DIR/lib/Makefile" && rm -f "$LIB_DIR/lib/Makefile.bak"
make -C "$LIB_DIR/lib" -j

CGO_CFLAGS_VAL="-I$LIB_DIR/include"
CGO_LDFLAGS_VAL="-L$LIB_DIR/lib -lrgbmatrix -lstdc++ -lm"

if [ -n "$GITHUB_ENV" ]; then
    echo "CGO_CFLAGS=$CGO_CFLAGS_VAL" >> "$GITHUB_ENV"
    echo "CGO_LDFLAGS=$CGO_LDFLAGS_VAL" >> "$GITHUB_ENV"
else
    echo ""
    echo "Library built. To build udb, set:"
    echo "  export CGO_CFLAGS=\"$CGO_CFLAGS_VAL\""
    echo "  export CGO_LDFLAGS=\"$CGO_LDFLAGS_VAL\""
fi
