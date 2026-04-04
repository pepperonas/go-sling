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
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white)]()
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

- **Peer-to-Peer transfers** — Files go directly between browsers via WebRTC DataChannel. The server only handles signaling.
- **Relay/Upload mode** — Fallback when only one device is online. Upload files to the server for others to download.
- **Directory support** — Drag & drop entire folders. Directory structure is preserved.
- **No file size limits** — Transfer multi-GB files without issues. Chunked transfer with flow control.
- **Zero runtime dependencies** — Single static binary. No database, no Redis, no Node.js.
- **Raspberry Pi optimized** — Runs on a Pi 3 with <50MB RAM. Cross-compiled for ARMv7.
- **Embedded web UI** — Frontend is baked into the binary via `go:embed`. No separate web server needed.
- **PIN authentication** — Optional shared PIN for access control. Rate-limited to prevent brute force.
- **Dark & Light theme** — Modern, responsive UI that works on desktop, tablet, and phone.
- **QR code on startup** — Prints a scannable QR code to the terminal for easy mobile access.
- **Auto-cleanup** — Uploaded files are automatically deleted after a configurable retention period.
- **Transfer progress** — Real-time progress bars with speed and ETA for both P2P and relay transfers.
- **SHA-256 checksums** — File integrity verification on every transfer.
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

Open the URL printed in the terminal (or scan the QR code) on any device in your network.

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

# All platforms
make build-all
```

Binaries are output to `bin/`.

## Raspberry Pi Setup

### Quick Install

```bash
# Copy binary to Pi
scp bin/go-sling-linux-arm7 pi@raspberrypi:/home/pi/go-sling/

# SSH into Pi
ssh pi@raspberrypi

# Run
cd /home/pi/go-sling
chmod +x go-sling-linux-arm7
./go-sling-linux-arm7 --pin changeme
```

### Systemd Service (Auto-Start)

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

### Transfer Modes

**P2P mode (default):** WebRTC DataChannel between two browsers. The Go server only relays signaling messages (SDP offers/answers, ICE candidates) over WebSocket. Files never touch the server.

**Relay mode (fallback):** Files are uploaded to the server's filesystem via chunked multipart upload. Other devices browse and download them through the web UI. Useful when only one device is online.

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
| `FILE_META` | `0x01` | File metadata (name, size, path) |
| `FILE_CHUNK` | `0x02` | Binary file chunk (64KB) |
| `FILE_DONE` | `0x03` | File complete + SHA-256 checksum |
| `TRANSFER_DONE` | `0x04` | All files transferred |
| `ACK` | `0x05` | Acknowledge chunks |
| `ERROR` | `0x06` | Error message |
| `CANCEL` | `0x09` | Cancel transfer |

### Project Structure

```
go-sling/
├── main.go                 # Entry point, CLI flags, config
├── internal/
│   ├── server/             # HTTP server, auth, routing
│   ├── api/                # REST API handlers (files, status)
│   ├── ws/                 # WebSocket hub, signaling
│   ├── storage/            # File storage, auto-cleanup
│   └── config/             # Configuration loading
├── web/                    # Embedded frontend (go:embed)
│   ├── index.html
│   ├── css/style.css
│   └── js/                 # Vanilla JS modules
└── scripts/                # Deployment scripts
```

## API Reference

### REST Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/api/auth` | Authenticate with PIN |
| `GET` | `/api/auth/status` | Check if auth is required |
| `GET` | `/api/files` | List uploaded files |
| `POST` | `/api/upload` | Upload files (multipart) |
| `GET` | `/api/download/:id` | Download file/directory |
| `DELETE` | `/api/files/:id` | Delete a file |
| `GET` | `/api/status` | Server status & metrics |

### WebSocket

| Path | Description |
|------|-------------|
| `/ws` | Signaling server for WebRTC |

## Security

- **LAN-only by design** — No STUN/TURN servers. Only host ICE candidates are used, keeping all traffic on the local network.
- **Optional PIN auth** — Rate-limited (5 attempts/minute per IP). Sessions stored as HTTP-only cookies.
- **No data leaves your network** — P2P transfers go directly between browsers. Relay uploads stay on the server's filesystem.
- **SHA-256 integrity** — Every file transfer is verified with a checksum.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Backend | Go (stdlib `net/http`) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Config | [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) |
| QR Code | [skip2/go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | Vanilla HTML/CSS/JS |
| P2P | WebRTC DataChannel (browser-native) |

## License

[MIT](LICENSE) — Martin Pfeffer / [celox.io](https://celox.io)
