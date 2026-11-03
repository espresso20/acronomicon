#!/bin/zsh
# Acronomicon Uninstaller for macOS
echo "Removing Acronomicon..."
rm -f "$HOME/.local/bin/acronomicon"
rm -rf "$HOME/Applications/Acronomicon.app"
if [ -w "/usr/local/bin/acronomicon" ]; then
    rm -f "/usr/local/bin/acronomicon"
fi
if [ -L "/opt/homebrew/bin/acronomicon" ]; then
    rm -f "/opt/homebrew/bin/acronomicon"
fi
if [ -w "/Applications/Acronomicon.app" ]; then
    rm -rf "/Applications/Acronomicon.app"
fi
echo "[OK] Acronomicon removed."
