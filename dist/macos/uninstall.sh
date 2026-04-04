#!/bin/bash
PLIST_NAME="io.celox.gosling-receiver"
PLIST_PATH="$HOME/Library/LaunchAgents/${PLIST_NAME}.plist"
INSTALL_DIR="$HOME/.gosling"

echo ""
echo "  🔗 go-sling receiver uninstaller"
echo ""

if [ -f "$PLIST_PATH" ]; then
    launchctl unload "$PLIST_PATH" 2>/dev/null || true
    rm -f "$PLIST_PATH"
    echo "  ✓ LaunchAgent removed"
fi

if [ -d "$INSTALL_DIR" ]; then
    rm -rf "$INSTALL_DIR"
    echo "  ✓ Installation removed ($INSTALL_DIR)"
fi

echo "  ✓ Uninstalled. Received files were kept."
echo ""
