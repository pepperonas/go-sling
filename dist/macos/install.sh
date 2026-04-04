#!/bin/bash
set -e

APP_NAME="gosling-receiver"
INSTALL_DIR="$HOME/.gosling"
PLIST_NAME="io.celox.gosling-receiver"
PLIST_PATH="$HOME/Library/LaunchAgents/${PLIST_NAME}.plist"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo ""
echo "  🔗 go-sling receiver installer"
echo "  ─────────────────────────────────"
echo ""

# Read config
if [ ! -f "$SCRIPT_DIR/config.json" ]; then
    echo "  ✗ config.json not found"
    exit 1
fi

SERVER=$(python3 -c "import json; print(json.load(open('$SCRIPT_DIR/config.json'))['server'])")
PIN=$(python3 -c "import json; print(json.load(open('$SCRIPT_DIR/config.json'))['pin'])")
OUTPUT=$(python3 -c "import json; print(json.load(open('$SCRIPT_DIR/config.json'))['output'])")
EXTRACT=$(python3 -c "import json; print('--no-extract' if not json.load(open('$SCRIPT_DIR/config.json'))['extract'] else '')")
CUSTOM_NAME=$(python3 -c "import json; n=json.load(open('$SCRIPT_DIR/config.json'))['name']; print(f'--name {n}' if n else '')")

echo "  Server:  $SERVER"
echo "  PIN:     ${PIN:0:1}***"
echo "  Output:  $OUTPUT"
echo ""

# Stop existing service
if launchctl list | grep -q "$PLIST_NAME" 2>/dev/null; then
    echo "  → Stopping existing service..."
    launchctl unload "$PLIST_PATH" 2>/dev/null || true
fi

# Create install directory
echo "  → Installing to $INSTALL_DIR..."
mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR/gosling-client.py" "$INSTALL_DIR/"
cp "$SCRIPT_DIR/config.json" "$INSTALL_DIR/"

# Create virtual environment
echo "  → Setting up Python environment..."
if [ ! -d "$INSTALL_DIR/venv" ]; then
    python3 -m venv "$INSTALL_DIR/venv"
fi
"$INSTALL_DIR/venv/bin/pip" install -q websockets 2>/dev/null

# Create output directory
mkdir -p "$(eval echo "$OUTPUT")"

# Build arguments
ARGS="--server $SERVER"
[ -n "$PIN" ] && ARGS="$ARGS --pin $PIN"
[ -n "$OUTPUT" ] && ARGS="$ARGS --output $OUTPUT"
[ -n "$EXTRACT" ] && ARGS="$ARGS $EXTRACT"
[ -n "$CUSTOM_NAME" ] && ARGS="$ARGS $CUSTOM_NAME"

# Create LaunchAgent for autostart
echo "  → Creating LaunchAgent (autostart on login)..."
mkdir -p "$HOME/Library/LaunchAgents"
cat > "$PLIST_PATH" << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${PLIST_NAME}</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_DIR}/venv/bin/python3</string>
        <string>${INSTALL_DIR}/gosling-client.py</string>
$(for arg in $ARGS; do echo "        <string>$arg</string>"; done)
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>${INSTALL_DIR}/gosling.log</string>
    <key>StandardErrorPath</key>
    <string>${INSTALL_DIR}/gosling.log</string>
    <key>WorkingDirectory</key>
    <string>${INSTALL_DIR}</string>
</dict>
</plist>
PLIST

# Start service
echo "  → Starting service..."
launchctl load "$PLIST_PATH"

echo ""
echo "  ✓ Installed and running!"
echo ""
echo "  Files will be saved to: $OUTPUT"
echo "  Logs: $INSTALL_DIR/gosling.log"
echo ""
echo "  Commands:"
echo "    View logs:     tail -f $INSTALL_DIR/gosling.log"
echo "    Stop:          launchctl unload $PLIST_PATH"
echo "    Start:         launchctl load $PLIST_PATH"
echo "    Uninstall:     bash $INSTALL_DIR/uninstall.sh"
echo ""
