#!/bin/zsh
# User-local macOS Installer for Acronomicon (No sudo required)
set -e
DIR="$(cd "$(dirname "$0")" && pwd)"
TARGET_BIN_DIR="$HOME/.local/bin"
TARGET_APP_DIR="$HOME/Applications"

echo ""
echo "=== Installing Acronomicon (macOS Universal) ==="
mkdir -p "$TARGET_BIN_DIR" "$TARGET_APP_DIR"

cp -f "$DIR/acronomicon" "$TARGET_BIN_DIR/acronomicon"
chmod +x "$TARGET_BIN_DIR/acronomicon"
xattr -dr com.apple.quarantine "$TARGET_BIN_DIR/acronomicon" 2>/dev/null || true

if [ -d "$DIR/Acronomicon.app" ]; then
    rm -rf "$TARGET_APP_DIR/Acronomicon.app"
    cp -R "$DIR/Acronomicon.app" "$TARGET_APP_DIR/Acronomicon.app"
    xattr -dr com.apple.quarantine "$TARGET_APP_DIR/Acronomicon.app" 2>/dev/null || true
fi

# Ensure ~/.local/bin is in ~/.zshrc
if ! grep -q '.local/bin' "$HOME/.zshrc" 2>/dev/null; then
    echo '' >> "$HOME/.zshrc"
    echo '# Added by Acronomicon installer' >> "$HOME/.zshrc"
    echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$HOME/.zshrc"
    echo "[OK] Added ~/.local/bin to ~/.zshrc"
fi

echo "[OK] Installed CLI binary to: $TARGET_BIN_DIR/acronomicon"
echo "[OK] Installed macOS App to:  $TARGET_APP_DIR/Acronomicon.app"
echo ""
echo "Launching Acronomicon..."
exec "$TARGET_BIN_DIR/acronomicon"
