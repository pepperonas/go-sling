# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
make build          # Build for current platform → bin/go-sling
make run            # Build and run
make build-raspi    # Cross-compile for Raspberry Pi 3 (linux/arm7)
make build-all      # All platforms (raspi, linux-amd64, mac-arm64, windows-amd64)
make test           # Run all tests
make clean          # Remove bin/
```

Cross-compile for Raspberry Pi 5 (ARM64):
```bash
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w -X main.version=1.0.2" -o bin/go-sling-linux-arm64 .
```

Binary flags: `--port 8420`, `--pin SECRET`, `--data-dir ./data`, `--retention 24`, `--config config.yaml`.

## Deployment

go-sling runs on **raspi5** (192.168.178.105) as a systemd service:
```bash
# Deploy
ssh raspi5 'sudo systemctl stop go-sling'
scp bin/go-sling-linux-arm64 raspi5:/home/pi/go-sling/go-sling
ssh raspi5 'sudo systemctl start go-sling'

# Logs
ssh raspi5 'sudo journalctl -u go-sling -f'
```

PIN: configured in `/etc/systemd/system/go-sling.service` on raspi5.

Android APK build:
```bash
cd clients/android
JAVA_HOME=$(/usr/libexec/java_home) ./gradlew assembleRelease
# Output: app/build/outputs/apk/release/app-release.apk
# Install: ~/Library/Android/sdk/platform-tools/adb install -r dist/gosling-receiver.apk
```

Keystore: `/Users/martin/My Drive/dev/keystore/go-sling-keystore/` (release.jks + keystore.properties).

## Architecture

Single Go binary with embedded frontend (no build step, no npm). Three transfer modes:

1. **P2P (WebRTC DataChannel)** — Go server handles WebSocket signaling only; files flow browser-to-browser.
2. **Relay** — Files uploaded to server filesystem (`./data/`), served to other devices via REST API. Same-name files are overwritten.
3. **Headless relay** — Browser sends to `POST /api/send-to/{peerId}`, server stores files and notifies headless client via WS `file-ready` message. Client downloads via `GET /api/download/{id}`.

### Backend packages

- `internal/config/` — Config loading from YAML/env/flags with priority: flags > env > file > defaults
- `internal/server/` — HTTP server, auth middleware (PIN + rate-limiting + persistent session cookies on disk), routing, QR code banner
- `internal/api/` — REST handlers: file CRUD, status, `SendTo` handler for headless peer relay with explicit path preservation
- `internal/ws/` — WebSocket hub with client tracking (including headless flag), WebRTC signaling relay, peer notifications
- `internal/storage/` — File storage with metadata persistence (`.metadata.json`), same-name overwrite, auto-cleanup goroutine, on-the-fly tar.gz streaming, path-aware file creation

### Frontend (web/)

Vanilla HTML/CSS/JS embedded via `go:embed`. No framework, no build step. Files:
- `js/app.js` — Main orchestrator, drop zones, headless peer detection (auto-relay instead of WebRTC), hidden file filtering (.DS_Store etc.)
- `js/ws.js` — WebSocket client with auto-reconnect
- `js/webrtc.js` — RTCPeerConnection management, DataChannel setup
- `js/transfer.js` — Binary chunking protocol (64KB chunks), checksums skipped on HTTP (crypto.subtle requires HTTPS)
- `js/zip.js` — Client-side ZIP builder (STORE method, CRC-32, no dependencies)
- `js/ui.js` — Safe DOM rendering, folder-grouped staged view, peer list with device icons and [Auto] tags
- `js/utils.js` — Device detection (MacBook/Android/iPad/etc.), creative name generation per device type
- `js/auth.js` — PIN login flow

### Headless clients (clients/)

- `clients/python/gosling-client.py` — Cross-platform auto-receiver. Connects via WS as headless peer, downloads files on `file-ready` notification, auto-extracts ZIPs. Uses `websockets` (async).
- `clients/android/` — Kotlin app with foreground service. Same WS-based flow. Saves to `Download/go-sling/`, auto-extracts, shows notifications.
- `dist/macos/` — One-click installer: `install.sh` creates venv, installs deps, sets up LaunchAgent for autostart on login.

### Key design constraints

- Must run on Raspberry Pi 3 (1GB RAM, ARMv7) and Pi 5 (ARM64) with <50MB idle RAM
- Only 3 external Go deps: `gorilla/websocket`, `gopkg.in/yaml.v3`, `skip2/go-qrcode`
- LAN-only: WebRTC uses only host ICE candidates (no STUN/TURN)
- Frontend must work without any build tooling
- `crypto.subtle` unavailable on HTTP — checksums gracefully skipped
- Hidden files (.DS_Store, Thumbs.db) filtered before transfer
- Directory structure preserved end-to-end via explicit paths JSON in relay uploads
