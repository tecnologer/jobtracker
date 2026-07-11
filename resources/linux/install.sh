#!/usr/bin/env bash
# Installs JobTracker for the current user only (no sudo). Run from the
# extracted release tarball: ./install.sh
set -euo pipefail

cd "$(dirname "$0")"

install -Dm755 jobtracker-desktop "$HOME/.local/bin/jobtracker-desktop"
install -Dm644 jobtracker.png "$HOME/.local/share/icons/hicolor/512x512/apps/jobtracker.png"
mkdir -p "$HOME/.local/share/applications"
sed "s|^Exec=.*|Exec=$HOME/.local/bin/jobtracker-desktop|" JobTracker.desktop > "$HOME/.local/share/applications/JobTracker.desktop"

update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
gtk-update-icon-cache -q -t "$HOME/.local/share/icons/hicolor" 2>/dev/null || true

echo "JobTracker installed. Launch it from your app menu, or run: $HOME/.local/bin/jobtracker-desktop"
echo "Your data lives in $HOME/.local/share/jobtracker/jobs.db"
