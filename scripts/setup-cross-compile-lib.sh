#!/usr/bin/env bash
# setup-cross-compile-lib.sh — cross-compile rpi-rgb-led-matrix for a target
# triple using Zig as the C/C++ cross-compiler.
#
# Usage: scripts/setup-cross-compile-lib.sh <zig-target-triple>
# Example: scripts/setup-cross-compile-lib.sh aarch64-linux-gnu
#
# The compiled library lands in build/libs/rpi-rgb-led-matrix-<target>/
# so it does not conflict with the native library built by setup_matrix_library.sh.
#
# Install Zig: brew install zig

set -euo pipefail

TARGET="${1:?Usage: $0 <zig-target-triple>}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
LIB_DIR="$REPO_ROOT/build/libs/rpi-rgb-led-matrix-$TARGET"

if ! command -v zig &>/dev/null; then
    echo "Error: zig is not installed. Install with: brew install zig" >&2
    exit 1
fi

if [ -d "$LIB_DIR/.git" ]; then
    echo "rpi-rgb-led-matrix already cloned for $TARGET, skipping"
else
    echo "Cloning rpi-rgb-led-matrix for $TARGET..."
    mkdir -p "$(dirname "$LIB_DIR")"
    git clone --depth=1 https://github.com/hzeller/rpi-rgb-led-matrix.git "$LIB_DIR"
fi

# Apply the same GPIO slowdown patch as the native build.
sed -i.bak 's/.*RGB_SLOWDOWN_GPIO.*/DEFINE+=-DRGB_SLOWDOWN_GPIO=2/' "$LIB_DIR/lib/Makefile" \
    && rm -f "$LIB_DIR/lib/Makefile.bak"

echo "Cross-compiling rpi-rgb-led-matrix for $TARGET..."
make -C "$LIB_DIR/lib" \
    CC="zig cc -target $TARGET" \
    CXX="zig c++ -target $TARGET" \
    AR="zig ar" \
    CPU_ARCH_FLAGS="" \
    -j

echo "Done: $LIB_DIR/lib/librgbmatrix.a"
