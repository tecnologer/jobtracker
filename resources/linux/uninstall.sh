#!/usr/bin/env bash
# Removes what install.sh installed. User data is left untouched.
set -euo pipefail

rm -f "$HOME/.local/bin/jobtracker-desktop"
rm -f "$HOME/.local/share/icons/hicolor/512x512/apps/jobtracker.png"
rm -f "$HOME/.local/share/applications/JobTracker.desktop"

update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true
gtk-update-icon-cache -q -t "$HOME/.local/share/icons/hicolor" 2>/dev/null || true

echo "JobTracker uninstalled. Your data in $HOME/.local/share/jobtracker/ was left untouched."
