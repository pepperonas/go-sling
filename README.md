# go-sling

<div align="center">

<img src="web/assets/banner.png" alt="go-sling banner" width="100%">

<h3>Fast peer-to-peer file transfer over your local network</h3>

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-blue?style=flat-square)]()
[![ARM Support](https://img.shields.io/badge/ARM-Raspberry%20Pi-c51a4a?style=flat-square&logo=raspberrypi&logoColor=white)]()
[![WebRTC](https://img.shields.io/badge/WebRTC-P2P-333333?style=flat-square&logo=webrtc&logoColor=white)]()
[![Zero Dependencies](https://img.shields.io/badge/Runtime_Deps-None-brightgreen?style=flat-square)]()
[![Single Binary](https://img.shields.io/badge/Deploy-Single%20Binary-orange?style=flat-square)]()
[![No Build Tools](https://img.shields.io/badge/Frontend-No%20npm%2C%20No%20Webpack-yellow?style=flat-square)]()
[![Binary Size](https://img.shields.io/badge/Binary-~7MB-blueviolet?style=flat-square)]()
[![RAM Usage](https://img.shields.io/badge/RAM-~6MB-blueviolet?style=flat-square)]()
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen?style=flat-square)]()
[![Made with Love](https://img.shields.io/badge/Made%20with-❤️-red?style=flat-square)]()

<br>

**Send files and folders between devices on your LAN — directly via WebRTC or through a lightweight relay server. No cloud, no accounts, no file size limits.**

<br>

```
┌──────────────┐         ┌──────────────────┐         ┌──────────────┐
│  Browser A   │◄──────► │  Go Server       │◄──────► │  Browser B   │
│  (Sender)    │  WS     │  (Raspi / PC)    │  WS     │  (Receiver)  │
│              │◄───────────────────────────────────► │              │
│              │    WebRTC DataChannel (P2P)          │              │
└──────────────┘                                      └──────────────┘
```

</div>

---

## Features

- **Peer-to-Peer transfers** — Files go directly between browsers via WebRTC DataChannel. The server only handles signaling. Zero server load during transfers.
- **Relay/Upload mode** — Fallback when only one device is online. Upload files to the server for others to download later.
- **Directory support** — Drag & drop entire folders. Directory structure is preserved. Multiple files are automatically bundled into a ZIP on the receiving end.
- **Smart device detection** — Each connected device is identified by type (MacBook, Android Phone, iPad, Windows Desktop, etc.) with distinct icons and creative auto-generated names like "Lunar-Glider-x7f" or "Turbo-Droid-a3b" for easy identification.
- **No file size limits** — Transfer multi-GB files without issues. Chunked transfer (64KB) with flow control.
- **Zero runtime dependencies** — Single static binary (~7MB). No database, no Redis, no Node.js.
- **Raspberry Pi optimized** — Runs on a Pi 3 with ~6MB RAM. Cross-compiled for ARMv7.
- **Embedded web UI** — Frontend is baked into the binary via `go:embed`. No separate web server needed.
- **PIN authentication** — Optional shared PIN for access control. Rate-limited (5 attempts/min per IP). Persistent sessions survive server restarts.
- **Dark & Light theme** — Modern, responsive UI that works on desktop, tablet, and phone. Theme preference persisted in localStorage.
- **QR code on startup** — Prints a scannable QR code to the terminal for easy mobile access.
- **Auto-cleanup** — Uploaded files are automatically deleted after a configurable retention period (default: 24h).
- **Transfer progress** — Real-time progress bars with speed (MB/s) and ETA for both P2P and relay transfers.
- **SHA-256 checksums** — File integrity verification on HTTPS connections.
- **LAN-only** — No STUN/TURN servers, no cloud relay. All traffic stays on your local network.

## Quick Start

### Download Binary

Download the latest release for your platform from the [Releases](https://github.com/pepperonas/go-sling/releases) page.

### Run

```bash
# No auth (trusted LAN)
./go-sling

# With PIN
./go-sling --pin mysecretpin

# Custom port
./go-sling --port 9000
```

On startup, go-sling prints the local URL and a QR code:

```
  ┌─────────────────────────────────────┐
  │           go-sling v1.0.0            │
  │       LAN File Transfer Server       │
  ├─────────────────────────────────────┤
  │  Local:   http://0.0.0.0:8420        │
  │  Network: http://192.168.1.42:8420   │
  └─────────────────────────────────────┘

  Scan to open:
  █████████████████████████
  █ ▄▄▄▄▄ █▀ ▄█▄█ ▄▄▄▄▄ █
  █ █   █ ██▀█ ▀█ █   █ █
  ...
```

Open the URL on any device in your network, enter your PIN, and start transferring.

### Build from Source

```bash
git clone https://github.com/pepperonas/go-sling.git
cd go-sling
make build
./bin/go-sling
```

### Cross-Compile

```bash
# Raspberry Pi 3 (ARMv7)
make build-raspi

# Linux x86_64
make build-linux

# macOS Apple Silicon
make build-mac

# Windows x86_64
make build-windows

# All platforms at once
make build-all
```

Binaries are output to `bin/`.

## How It Works

### Peer Discovery

When a device opens go-sling in the browser, it connects via WebSocket and announces itself. The server detects the device type from the User-Agent and assigns a creative name:

| Device | Icon | Example Names |
|--------|------|---------------|
| MacBook | 💻 | Lunar-Book-x7f, Stellar-Wing-a3b, Cosmic-Rider-9e2 |
| Mac Desktop / iMac | 🖥️ | Thunder-Tower-f1c, Storm-Hub-4d8, Forge-Core-b7a |
| Android Phone | 📱 | Turbo-Droid-e5f, Flash-Spark-2c1, Blitz-Pulse-8a3 |
| Android Tablet | 📲 | Atlas-Pad-d4e, Titan-Grid-7b9, Nova-Pane-1f6 |
| iPhone | 📱 | Zippy-Pocket-c3d, Swift-Dart-6a2, Nimble-Comet-f8e |
| iPad | 📲 | Mighty-Canvas-a1b, Grand-Shield-5c7, Bright-Slate-e9d |
| Windows Laptop | 💻 | Pixel-Flip-b2c, Cyber-Node-8f1, Neon-Link-3d7 |
| Windows Desktop | 🖥️ | Granite-Rig-f4a, Steel-Desk-1e9, Iron-Mill-7c3 |
| Linux | 🐧 | Kernel-Box-d6e, Root-Node-2a8, Daemon-Stack-9f1 |
| Raspberry Pi | 🍓 | Berry-Pi-a4b, Tiny-Chip-8c2, Micro-Dot-3e7 |

All connected peers are shown in the sidebar with their icon, name, device type, and browser — making it easy to pick the right target for your transfer.

### Transfer Modes

**P2P mode (default):** Select a peer, drop your files, hit Send. A WebRTC DataChannel is established directly between the two browsers. Files are split into 64KB chunks and streamed peer-to-peer. The server only relays the initial signaling (SDP offer/answer + ICE candidates). Files never touch the server.

**Relay mode (fallback):** Switch to the "Files (Server)" tab to upload files to the server's filesystem. Other devices can browse and download them. Directories are automatically served as streaming tar.gz archives. Useful when the receiver isn't online yet.

### Receiving Files

When a P2P transfer completes on the receiving device:
- **Single file:** A "Save File" button appears — tap to download.
- **Multiple files:** All files are automatically bundled into a ZIP (named after the source directory, e.g., `Photos.zip`) with a single "Save ZIP" button.

This works reliably on all platforms including Android, which blocks programmatic downloads.

## Raspberry Pi Setup

### Quick Install

```bash
# Build on your dev machine
make build-raspi

# Copy binary to Pi
scp bin/go-sling-linux-arm7 pi@raspberrypi:/home/pi/go-sling/

# SSH into Pi and run
ssh pi@raspberrypi
cd /home/pi/go-sling
chmod +x go-sling-linux-arm7
./go-sling-linux-arm7 --pin changeme
```

### PM2 (Recommended)

```bash
ssh pi@raspberrypi
cd /home/pi/go-sling
pm2 start ./go-sling-linux-arm7 --name go-sling -- --pin changeme
pm2 save
```

### Systemd Service

```bash
# On the Raspberry Pi
sudo bash scripts/install-service.sh
```

This creates a systemd service that starts go-sling on boot. Edit the PIN in `/etc/systemd/system/go-sling.service`.

```bash
# Manage the service
sudo systemctl status go-sling
sudo systemctl restart go-sling
sudo journalctl -u go-sling -f
```

## Configuration

go-sling can be configured via CLI flags, environment variables, or a `config.yaml` file.

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `8420` | Server port |
| `--pin` | *(none)* | Authentication PIN |
| `--data-dir` | `./data` | File storage directory |
| `--retention` | `24` | File retention in hours |
| `--config` | `config.yaml` | Config file path |
| `--version` | | Show version and exit |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `PORT` | Server port |
| `PIN` | Authentication PIN |
| `DATA_DIR` | File storage directory |

### Config File

A `config.yaml` is auto-generated on first run:

```yaml
server:
  port: 8420
  host: "0.0.0.0"
auth:
  pin: ""
storage:
  data_dir: "./data"
  max_upload_size: 10737418240  # 10GB
  retention_hours: 24
  auto_cleanup: true
transfer:
  chunk_size: 65536   # 64KB
  ack_interval: 16    # ACK every 16 chunks
ui:
  default_theme: "dark"
  app_name: "go-sling"
```

Priority: CLI flags > environment variables > config file > defaults.

## Architecture

### Signaling Flow

```
Sender                    Server                    Receiver
  │─── join ──────────────►│                           │
  │                        │◄────── join ──────────────│
  │◄── peer-list ─────────│──────► peer-list ─────────►│
  │─── offer (to B) ─────►│──────► offer ─────────────►│
  │◄── answer ────────────│◄────── answer (to A) ─────│
  │─── ice-candidate ────►│──────► ice-candidate ─────►│
  │◄═══════ DataChannel (direct P2P) ═══════════════►│
```

### DataChannel Protocol

Binary protocol for efficient file transfer:

| Type | Code | Description |
|------|------|-------------|
| `FILE_META` | `0x01` | File metadata (name, size, path, total files) |
| `FILE_CHUNK` | `0x02` | Binary file chunk — header: `[type(1)][fileIndex(4)][chunkIndex(4)][data]` |
| `FILE_DONE` | `0x03` | File complete + SHA-256 checksum |
| `TRANSFER_DONE` | `0x04` | All files transferred |
| `ACK` | `0x05` | Acknowledge chunks (flow control) |
| `ERROR` | `0x06` | Error message |
| `PAUSE` | `0x07` | Pause transfer |
| `RESUME` | `0x08` | Resume transfer |
| `CANCEL` | `0x09` | Cancel transfer |

### Project Structure

```
go-sling/
├── main.go                 # Entry point, CLI flags, config loading
├── Makefile                # Cross-compilation targets
├── internal/
│   ├── config/config.go    # Config from YAML/env/flags
│   ├── server/
│   │   ├── server.go       # HTTP server, middleware, banner/QR
│   │   ├── auth.go         # PIN auth, sessions (persisted to disk), rate limiting
│   │   └── routes.go       # Route registration
│   ├── api/
│   │   ├── files.go        # Upload, download, list, delete handlers
│   │   └── status.go       # Server status & metrics
│   ├── ws/
│   │   ├── hub.go          # WebSocket hub, peer tracking, broadcast
│   │   ├── client.go       # WebSocket client, heartbeat, name generation
│   │   └── signaling.go    # SDP/ICE relay, transfer negotiation
│   └── storage/
│       ├── store.go        # File storage, metadata, tar.gz streaming
│       └── cleanup.go      # Auto-cleanup goroutine
├── web/                    # Embedded via go:embed
│   ├── index.html          # Single-page app
│   ├── css/style.css       # Dark/light themes, responsive
│   ├── js/
│   │   ├── app.js          # Main orchestrator, drop zones, polling
│   │   ├── auth.js         # PIN login flow
│   │   ├── ws.js           # WebSocket client, auto-reconnect
│   │   ├── webrtc.js       # RTCPeerConnection, DataChannel
│   │   ├── transfer.js     # Chunked transfer protocol, progress
│   │   ├── zip.js          # Client-side ZIP builder (no dependencies)
│   │   ├── ui.js           # Safe DOM rendering, toasts, modals
│   │   └── utils.js        # Device detection, formatting, icons
│   └── assets/
│       ├── banner.png
│       └── favicon.svg
└── scripts/
    ├── install-service.sh  # Systemd service installer
    └── generate-cert.sh    # Self-signed TLS certificate generator
```

## API Reference

### REST Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/auth` | No | Authenticate with PIN, returns session cookie |
| `GET` | `/api/auth/status` | No | Check if auth is required |
| `GET` | `/api/files` | Yes | List uploaded files with metadata |
| `POST` | `/api/upload` | Yes | Upload files (multipart/form-data) |
| `GET` | `/api/download/:id` | Yes | Download file or directory (tar.gz) |
| `DELETE` | `/api/files/:id` | Yes | Delete a file |
| `GET` | `/api/status` | Yes | Server info: version, uptime, peers, storage, memory |

### WebSocket

| Path | Auth | Description |
|------|------|-------------|
| `/ws` | No | Signaling server for WebRTC peer discovery and connection setup |

**Message types:** `join`, `peer-list`, `welcome`, `offer`, `answer`, `ice-candidate`, `transfer-request`, `transfer-accept`, `transfer-reject`

## Security

- **LAN-only by design** — No STUN/TURN servers. Only host ICE candidates are used, keeping all traffic strictly on the local network.
- **Optional PIN auth** — Rate-limited to 5 failed attempts per minute per IP. Sessions stored as HTTP-only cookies and persisted to disk (survive restarts). "Stay logged in" option for permanent sessions.
- **No data leaves your network** — P2P transfers go directly between browsers. Relay uploads stay on the server's filesystem.
- **SHA-256 integrity** — File checksums verified on HTTPS connections (crypto.subtle requires secure context).
- **Path traversal protection** — All file paths are sanitized and validated against directory escape attacks.
- **No external requests** — The binary makes zero outgoing network connections. Everything runs locally.

## Performance

| Metric | Value |
|--------|-------|
| Binary size | ~7 MB (stripped, ARM) |
| Idle RAM | ~6 MB |
| Startup time | <1 second |
| Max concurrent peers | Tested with 10+ |
| Transfer speed (P2P) | Limited only by LAN bandwidth |
| Transfer speed (Relay) | Limited by server disk I/O |

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go (stdlib `net/http`, `go:embed`) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Config | [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) |
| QR Code | [skip2/go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | Vanilla HTML/CSS/JS — no npm, no build step |
| P2P | WebRTC DataChannel (browser-native) |
| ZIP | Custom client-side ZIP builder (STORE method, CRC-32) |

## License

[MIT](LICENSE) — Martin Pfeffer / [celox.io](https://celox.io)
