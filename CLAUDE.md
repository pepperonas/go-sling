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

Binary flags: `--port 8420`, `--pin SECRET`, `--data-dir ./data`, `--retention 24`, `--config config.yaml`.

## Architecture

Single Go binary with embedded frontend (no build step, no npm). Two transfer modes:

1. **P2P (WebRTC DataChannel)** — Go server handles WebSocket signaling only; files flow browser-to-browser.
2. **Relay** — Files uploaded to server filesystem (`./data/`), served to other devices via REST API.

### Backend packages

- `internal/config/` — Config loading from YAML/env/flags with priority: flags > env > file > defaults
- `internal/server/` — HTTP server setup, auth middleware (PIN + rate-limiting + session cookies), routing
- `internal/api/` — REST handlers: file CRUD (`/api/files`, `/api/upload`, `/api/download/:id`), status (`/api/status`)
- `internal/ws/` — WebSocket hub with client tracking, WebRTC signaling relay (SDP + ICE candidates), heartbeat
- `internal/storage/` — File storage with metadata persistence (`.metadata.json`), auto-cleanup goroutine, on-the-fly tar.gz streaming for directories

### Frontend (web/)

Vanilla HTML/CSS/JS embedded via `go:embed`. No framework, no build step. Files:
- `js/app.js` — Main orchestrator, drop zone setup, polling
- `js/ws.js` — WebSocket client with auto-reconnect
- `js/webrtc.js` — RTCPeerConnection management, DataChannel setup
- `js/transfer.js` — Binary chunking protocol (64KB chunks, SHA-256 checksums)
- `js/ui.js` — DOM manipulation (all rendering uses safe DOM APIs, no innerHTML with untrusted data)
- `js/auth.js` — PIN login flow

### Key design constraints

- Must run on Raspberry Pi 3 (1GB RAM, ARMv7) with <50MB idle RAM
- Only 3 external Go deps: `gorilla/websocket`, `gopkg.in/yaml.v3`, `skip2/go-qrcode`
- LAN-only: WebRTC uses only host ICE candidates (no STUN/TURN)
- Frontend must work without any build tooling
