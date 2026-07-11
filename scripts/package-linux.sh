#!/usr/bin/env bash
# Packages a built jobtracker-desktop binary into a release tarball.
# Usage: package-linux.sh <binary-path> <arch>
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "Usage: $0 <binary-path> <arch>" >&2
  exit 1
fi

BINARY="$1"
ARCH="$2"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

STAGE_NAME="jobtracker-desktop-linux-${ARCH}"
STAGE_DIR="$(mktemp -d)/${STAGE_NAME}"
mkdir -p "$STAGE_DIR"

install -Dm755 "$BINARY" "$STAGE_DIR/jobtracker-desktop"
install -Dm644 "$ROOT/resources/images/icons/icon-512.png" "$STAGE_DIR/jobtracker.png"
install -Dm644 "$ROOT/resources/linux/JobTracker.desktop" "$STAGE_DIR/JobTracker.desktop"
install -Dm755 "$ROOT/resources/linux/install.sh" "$STAGE_DIR/install.sh"
install -Dm755 "$ROOT/resources/linux/uninstall.sh" "$STAGE_DIR/uninstall.sh"
install -Dm644 "$ROOT/LICENSE" "$STAGE_DIR/LICENSE"

cat > "$STAGE_DIR/README.txt" <<'EOF'
JobTracker desktop — job application tracker (Wails v2 app).

Requires: libwebkit2gtk-4.1 (and GTK3) at runtime.
  Debian/Ubuntu: apt install libwebkit2gtk-4.1-0
  Fedora:        dnf install webkit2gtk4.1
  Arch:          pacman -S webkit2gtk-4.1

Run ./install.sh (installs to ~/.local, no root); ./uninstall.sh to remove.
EOF

tar -czf "${STAGE_NAME}.tar.gz" -C "$(dirname "$STAGE_DIR")" "$STAGE_NAME"
rm -rf "$(dirname "$STAGE_DIR")"

echo "Created $(pwd)/${STAGE_NAME}.tar.gz"
