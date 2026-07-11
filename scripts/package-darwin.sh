#!/usr/bin/env bash
# Assembles JobTracker.app from a built binary and packages it as a .dmg for release.
# Usage: package-darwin.sh <binary-path> <version>   (version like v1.2.3)
# macOS-only: uses sips, iconutil, codesign, hdiutil.
set -euo pipefail

if [ $# -ne 2 ]; then
  echo "Usage: $0 <binary-path> <version>" >&2
  exit 1
fi

BINARY="$1"
VERSION="${2#v}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 1. Build icon.icns from the 512px source PNG.
ICONSET="JobTracker.iconset"
rm -rf "$ICONSET"
mkdir -p "$ICONSET"
SRC_ICON="$ROOT/resources/images/icons/icon-512.png"

sips -z 16 16 "$SRC_ICON" --out "$ICONSET/icon_16x16.png" >/dev/null
sips -z 32 32 "$SRC_ICON" --out "$ICONSET/icon_16x16@2x.png" >/dev/null
sips -z 32 32 "$SRC_ICON" --out "$ICONSET/icon_32x32.png" >/dev/null
sips -z 64 64 "$SRC_ICON" --out "$ICONSET/icon_32x32@2x.png" >/dev/null
sips -z 128 128 "$SRC_ICON" --out "$ICONSET/icon_128x128.png" >/dev/null
sips -z 256 256 "$SRC_ICON" --out "$ICONSET/icon_128x128@2x.png" >/dev/null
sips -z 256 256 "$SRC_ICON" --out "$ICONSET/icon_256x256.png" >/dev/null
sips -z 512 512 "$SRC_ICON" --out "$ICONSET/icon_256x256@2x.png" >/dev/null
sips -z 512 512 "$SRC_ICON" --out "$ICONSET/icon_512x512.png" >/dev/null

iconutil -c icns "$ICONSET" -o icon.icns
rm -rf "$ICONSET"

# 2. Assemble the bundle.
rm -rf JobTracker.app
mkdir -p JobTracker.app/Contents/MacOS JobTracker.app/Contents/Resources
install -m755 "$BINARY" JobTracker.app/Contents/MacOS/jobtracker
mv icon.icns JobTracker.app/Contents/Resources/icon.icns
sed "s|@VERSION@|$VERSION|" "$ROOT/resources/darwin/Info.plist" > JobTracker.app/Contents/Info.plist

# 3. Ad-hoc sign — required for Apple Silicon to launch the binary at all.
codesign --force --deep -s - JobTracker.app

# 4. Build a .dmg with an /Applications symlink for drag-to-install.
DMG_STAGE="$(mktemp -d)"
cp -R JobTracker.app "$DMG_STAGE/"
cp "$ROOT/LICENSE" "$DMG_STAGE/LICENSE"
ln -s /Applications "$DMG_STAGE/Applications"

DMG_NAME="JobTracker-darwin-arm64.dmg"
rm -f "$DMG_NAME"
hdiutil create \
  -volname "JobTracker" \
  -srcfolder "$DMG_STAGE" \
  -fs HFS+ \
  -format UDZO \
  -ov \
  "$DMG_NAME"
rm -rf "$DMG_STAGE"

echo "Created $(pwd)/${DMG_NAME}"
