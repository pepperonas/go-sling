# go-sling

<div align="center">

<img src="web/assets/banner.png" alt="go-sling banner" width="100%">

<h3>Fast peer-to-peer file transfer over your local network</h3>

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://golang.org)
[![Release](https://img.shields.io/github/v/release/pepperonas/go-sling?style=flat-square&color=6366f1)](https://github.com/pepperonas/go-sling/releases)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows%20%7C%20Android-blue?style=flat-square)]()
[![ARM Support](https://img.shields.io/badge/ARM-Raspberry%20Pi%203%20%26%205-c51a4a?style=flat-square&logo=raspberrypi&logoColor=white)]()
[![WebRTC](https://img.shields.io/badge/WebRTC-P2P-333333?style=flat-square&logo=webrtc&logoColor=white)]()
[![Zero Dependencies](https://img.shields.io/badge/Runtime_Deps-None-brightgreen?style=flat-square)]()
[![Single Binary](https://img.shields.io/badge/Deploy-Single%20Binary-orange?style=flat-square)]()
[![No Build Tools](https://img.shields.io/badge/Frontend-No%20npm%2C%20No%20Webpack-yellow?style=flat-square)]()
[![Binary Size](https://img.shields.io/badge/Binary-~7MB-blueviolet?style=flat-square)]()
[![RAM Usage](https://img.shields.io/badge/RAM-~6MB-blueviolet?style=flat-square)]()
[![Android App](https://img.shields.io/badge/Android-APK%20Available-34A853?style=flat-square&logo=android&logoColor=white)](https://github.com/pepperonas/go-sling/releases)
[![Python Client](https://img.shields.io/badge/Python-CLI%20Client-3776AB?style=flat-square&logo=python&logoColor=white)]()
[![PRs Welcome](https://img.shields.io/badge/PRs-Welcome-brightgreen?style=flat-square)]()
[![Made with Love](https://img.shields.io/badge/Made%20with-❤️-red?style=flat-square)]()

<br>

**Send files and folders between any device on your LAN — directly via WebRTC or through a lightweight relay server. No cloud, no accounts, no file size limits. Includes auto-receive clients for macOS, Linux, Windows, and Android.**

<br>

```
┌──────────────┐         ┌──────────────────┐         ┌──────────────┐
│  Browser A   │◄──────► │  Go Server       │◄──────► │  Browser B   │
│  (Sender)    │  WS     │  (Raspi / PC)    │  WS     │  (Receiver)  │
│              │◄───────────────────────────────────► │              │
│              │    WebRTC DataChannel (P2P)          │              │
└──────────────┘                                      └──────────────┘
                                  ▲
                                  │ WS + HTTP
                          ┌───────┴────────┐
                          │ Headless Client │
                          │ (Python / Android) │
                          │  Auto-Receive   │
                          └────────────────┘
```

</div>

---

## Features

### Core
- **Peer-to-Peer transfers** — Files go directly between browsers via WebRTC DataChannel. The server only handles signaling. Zero server load during transfers.
- **Relay/Upload mode** — Fallback when only one device is online. Upload to the server for others to download later. Re-uploading a file with the same name overwrites the existing one.
- **Directory support** — Drag & drop entire folders. Directory structure is preserved 1:1 on all receiving ends (browser, Android app, Python CLI). Folders are displayed as a single entry (e.g., `📁 assets/ (4 files)`) instead of listing individual files. Hidden files (`.DS_Store`, `Thumbs.db`) are filtered automatically. Browser P2P bundles multiple files into a ZIP named after the source folder (e.g., `assets.zip`).
- **No file size limits** — Transfer multi-GB files. Chunked transfer (64KB) with flow control.

### Clients & Platforms
- **Browser UI** — Embedded SPA served from the Go binary. Works on any device with a modern browser.
- **macOS auto-receiver** — One-click installer with LaunchAgent. Starts on login, auto-downloads and extracts files in the background.
- **Android auto-receiver** — Native Kotlin app with foreground service. Auto-downloads to `Download/go-sling/`, extracts ZIPs, shows notifications.
- **Python CLI** — Cross-platform headless client (macOS/Linux/Windows). Connects as a background peer, auto-receives everything.

### Server
- **Single static binary** — ~7MB, no runtime dependencies. No database, no Redis, no Node.js.
- **Raspberry Pi optimized** — Runs on Pi 3 (ARMv7) and Pi 5 (ARM64) with ~6MB RAM.
- **Embedded web UI** — Frontend baked into the binary via `go:embed`. No separate web server needed.
- **QR code on startup** — Prints a scannable QR code to the terminal for easy mobile access.
- **Auto-cleanup** — Uploaded files are automatically deleted after a configurable retention period (default: 24h).

### UX
- **Smart device detection** — Each device gets a distinct icon and creative auto-generated name based on its type.
- **Dark & Light theme** — Modern, responsive UI. Theme preference persisted in localStorage.
- **PIN authentication** — Optional shared PIN. Rate-limited (5 attempts/min per IP). Persistent sessions survive server restarts. "Stay logged in" option.
- **Transfer progress** — Real-time progress bars with speed (MB/s) and ETA.
- **SHA-256 checksums** — File integrity verification on HTTPS connections.
- **LAN-only** — No STUN/TURN servers, no cloud relay. All traffic stays on your local network.

## Quick Start

### Download

Grab the latest release for your platform from the [Releases page](https://github.com/pepperonas/go-sling/releases):

| File | Platform | Description |
|------|----------|-------------|
| `go-sling-linux-arm64` | Raspberry Pi 5 | Server binary (ARM64) |
| `go-sling-linux-arm7` | Raspberry Pi 3 | Server binary (ARMv7) |
| `gosling-receiver-macos.zip` | macOS | Auto-receive installer with LaunchAgent |
| `gosling-receiver.apk` | Android | Auto-receive app (signed, 1.9MB) |

### Run the Server

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
  │           go-sling v1.0.2            │
  │       LAN File Transfer Server       │
  ├─────────────────────────────────────┤
  │  Local:   http://0.0.0.0:8420        │
  │  Network: http://192.168.1.42:8420   │
  └─────────────────────────────────────┘

  Scan to open:
  █████████████████████████
  █ ▄▄▄▄▄ █▀ ▄█▄█ ▄▄▄▄▄ █
  ...
```

Open the URL on any device in your network, enter your PIN, and start transferring.

### Build from Source

```bash
git clone https://github.com/pepperonas/go-sling.git
cd go-sling
make build        # Current platform
make build-raspi  # Raspberry Pi 3 (ARMv7)
make build-all    # All platforms
```

## How It Works

### Peer Discovery

When a device connects, it's automatically identified by type with a distinct icon and a creative auto-generated name:

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
| Headless Client | 🤖 | Lunar-Agent-x7f, Turbo-Catcher-a3b `[Auto]` |

All connected peers are shown in the sidebar with their icon, name, device type, and browser. Headless auto-receive clients are marked with an `[Auto]` tag.

### Transfer Modes

**P2P mode (browser ↔ browser):** Select a peer, drop your files, hit Send. A WebRTC DataChannel is established directly between the two browsers. Files are split into 64KB chunks and streamed peer-to-peer. The server only relays signaling messages. Files never touch the server disk.

**Relay mode (browser → server → download):** Switch to the "Files (Server)" tab to upload files to the server's filesystem. Other devices can browse and download them. Directories are served as streaming tar.gz archives. Useful when the receiver isn't online yet. Re-uploading a file with the same name replaces the existing one.

**Headless mode (browser → server → auto-receive client):** When the target peer is a headless client (Python CLI or Android app), the browser automatically uploads via the server relay. The server notifies the client via WebSocket, which then downloads and optionally extracts the files — all fully automatic, no manual interaction needed on the receiving end.

### Receiving Files

**In browser (P2P):**
- **Single file:** A "Save File" button appears — tap to download.
- **Multiple files:** Automatically bundled into a ZIP named after the source directory (e.g., `Photos.zip`) with a single "Save ZIP" button. Works reliably on Android, which blocks programmatic downloads.

**On headless clients (auto-receive):**
- Files are saved automatically to the configured output directory.
- Directory structure is preserved 1:1 end-to-end — sending `assets/` creates `assets/martin.jpg`, `assets/logo.png`, etc. on the receiver. The folder appears as a proper subdirectory (e.g., `Download/go-sling/assets/`), not as flat files.
- Hidden files (`.DS_Store`, `Thumbs.db`, `desktop.ini`) are filtered out before transfer.
- ZIP archives are auto-extracted and the archive file is removed.
- A notification is shown for each received file (Android).

## Raspberry Pi Setup

### Raspberry Pi 5 (ARM64) — Recommended

```bash
# On your dev machine
GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o go-sling .

# Copy to Pi
scp go-sling pi@raspberrypi:/home/pi/go-sling/

# SSH into Pi and install as systemd service
ssh pi@raspberrypi
chmod +x /home/pi/go-sling/go-sling
sudo bash scripts/install-service.sh
```

### Raspberry Pi 3 (ARMv7)

```bash
make build-raspi
scp bin/go-sling-linux-arm7 pi@raspberrypi:/home/pi/go-sling/
```

### Process Management

**Systemd (recommended for dedicated server):**
```bash
sudo systemctl status go-sling
sudo systemctl restart go-sling
sudo journalctl -u go-sling -f
```

**PM2 (if already using PM2 for other apps):**
```bash
pm2 start ./go-sling --name go-sling -- --pin changeme
pm2 save
```

## Headless Clients (Auto-Receive)

go-sling includes companion clients that run in the background and **automatically receive files** — no manual download needed. Files are saved to a configured directory, directory structure is preserved, and ZIP archives are auto-extracted.

### macOS — One-Click Installer

Download [`gosling-receiver-macos.zip`](https://github.com/pepperonas/go-sling/releases) and run:

```bash
unzip gosling-receiver-macos.zip
cd macos

# Edit config with your server address and PIN
nano config.json

# Install (sets up venv, LaunchAgent, starts receiving)
bash install.sh
```

**What `install.sh` does:**
1. Installs the Python client to `~/.gosling/`
2. Creates a virtual environment and installs `websockets`
3. Registers a **macOS LaunchAgent** (`~/Library/LaunchAgents/io.celox.gosling-receiver.plist`)
4. Starts receiving immediately

**The client auto-starts on every login and auto-restarts if it crashes** — no manual intervention needed after install.

```bash
# View live logs
tail -f ~/.gosling/gosling.log

# Stop
launchctl unload ~/Library/LaunchAgents/io.celox.gosling-receiver.plist

# Start
launchctl load ~/Library/LaunchAgents/io.celox.gosling-receiver.plist

# Uninstall completely (keeps received files)
bash ~/.gosling/uninstall.sh
```

**`config.json` format:**
```json
{
  "server": "192.168.1.42:8420",
  "pin": "your-pin",
  "output": "~/go-sling-received",
  "extract": true,
  "name": ""
}
```

Files are saved to `~/go-sling-received/` by default. Leave `name` empty for an auto-generated creative name.

### Android App

Download [`gosling-receiver.apk`](https://github.com/pepperonas/go-sling/releases) (1.9MB, signed) and install on your phone.

1. Open **go-sling**
2. Enter your server address (e.g., `192.168.1.42:8420`)
3. Enter your PIN
4. Tap **Start Receiving**

Settings are saved automatically — next time you open the app, everything is pre-filled.

**The app provides:**
- Foreground service that stays running even when the app is closed
- Auto-download of files to `Download/go-sling/` on the device
- Auto-extract of ZIP archives with directory structure preserved
- Notification for each received file
- Auto-reconnect on connection loss
- Live log view in the app

**Build from source:**
```bash
cd clients/android
# Copy keystore (see KEYSTORE.md for setup)
./gradlew assembleRelease
# Output: app/build/outputs/apk/release/app-release.apk
```

### Python CLI (Manual Setup)

For advanced users or non-macOS platforms (Linux, Windows):

```bash
cd clients/python
pip install -r requirements.txt

# Start receiving
python gosling-client.py --server 192.168.1.42:8420 --pin 3001

# Custom output directory and name
python gosling-client.py -s 192.168.1.42:8420 -p 3001 -o ~/received -n my-laptop

# Don't auto-extract ZIPs
python gosling-client.py -s 192.168.1.42:8420 -p 3001 --no-extract
```

| Flag | Default | Description |
|------|---------|-------------|
| `--server`, `-s` | *(required)* | Server address (host:port) |
| `--pin`, `-p` | *(none)* | Authentication PIN |
| `--output`, `-o` | `~/go-sling-received` | Download directory |
| `--name`, `-n` | *(auto-generated)* | Custom peer name |
| `--no-extract` | `false` | Don't auto-extract ZIP files |

**Requirements:** Python 3.8+, `websockets` (`pip install websockets`)

For autostart on Linux, create a systemd user service:
```bash
mkdir -p ~/.config/systemd/user
cat > ~/.config/systemd/user/gosling-receiver.service << EOF
[Unit]
Description=go-sling receiver
After=network-online.target

[Service]
ExecStart=/path/to/venv/bin/python /path/to/gosling-client.py -s HOST:8420 -p PIN
Restart=always

[Install]
WantedBy=default.target
EOF
systemctl --user enable --now gosling-receiver
```

### How Headless Transfer Works

```
Browser (Sender)              Server              Headless Client
  │                            │                        │
  │  select headless peer      │                        │
  │  drop files                │                        │
  │                            │                        │
  │─── POST /api/send-to/id ─►│                        │
  │    (multipart upload)      │── store files          │
  │                            │                        │
  │    200 OK ◄────────────────│                        │
  │                            │─── WS "file-ready" ──►│
  │                            │    {id, name, size}    │
  │                            │                        │
  │                            │◄── GET /api/download ──│
  │                            │──── file stream ──────►│
  │                            │                        │── save to disk
  │                            │                        │── extract if ZIP
  │                            │                        │── done ✓
```

The browser detects headless peers automatically (marked with `[Auto]`) and uses server relay instead of WebRTC. Directory structure is preserved end-to-end.

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

### Signaling Flow (P2P)

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
| `FILE_META` | `0x01` | File metadata (name, size, relative path, total file count) |
| `FILE_CHUNK` | `0x02` | Binary chunk — `[type(1)][fileIndex(4)][chunkIndex(4)][data]` |
| `FILE_DONE` | `0x03` | File complete + SHA-256 checksum |
| `TRANSFER_DONE` | `0x04` | All files transferred |
| `ACK` | `0x05` | Acknowledge chunks (flow control) |
| `ERROR` | `0x06` | Error message |
| `PAUSE` | `0x07` | Pause transfer |
| `RESUME` | `0x08` | Resume transfer |
| `CANCEL` | `0x09` | Cancel transfer |

### REST API

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| `POST` | `/api/auth` | No | Authenticate with PIN, returns session cookie |
| `GET` | `/api/auth/status` | No | Check if auth is required |
| `GET` | `/api/files` | Yes | List uploaded files with metadata |
| `POST` | `/api/upload` | Yes | Upload files (multipart/form-data) |
| `POST` | `/api/send-to/{peerId}` | Yes | Upload files to a headless peer via relay |
| `GET` | `/api/download/{id}` | Yes | Download file or directory (auto tar.gz) |
| `DELETE` | `/api/files/{id}` | Yes | Delete a file |
| `GET` | `/api/status` | Yes | Server info: version, uptime, peers, storage, memory |

### WebSocket

| Path | Description |
|------|-------------|
| `/ws` | Signaling + peer discovery + headless notifications |

**Message types:** `join`, `welcome`, `peer-list`, `offer`, `answer`, `ice-candidate`, `transfer-request`, `transfer-accept`, `transfer-reject`, `file-ready`

### Project Structure

```
go-sling/
├── main.go                     # Entry point, CLI flags, config loading
├── Makefile                    # Cross-compilation targets
├── internal/
│   ├── config/config.go        # Config from YAML / env / flags
│   ├── server/
│   │   ├── server.go           # HTTP server, middleware, QR code banner
│   │   ├── auth.go             # PIN auth, sessions persisted to disk, rate limiting
│   │   └── routes.go           # Route registration
│   ├── api/
│   │   ├── files.go            # Upload, download, list, delete, send-to handlers
│   │   └── status.go           # Server status & runtime metrics
│   ├── ws/
│   │   ├── hub.go              # WebSocket hub, peer tracking, headless notifications
│   │   ├── client.go           # WebSocket client, heartbeat, name generation
│   │   └── signaling.go        # SDP/ICE relay, transfer negotiation
│   └── storage/
│       ├── store.go            # File storage, dedup/overwrite, tar.gz streaming
│       └── cleanup.go          # Auto-cleanup goroutine
├── web/                        # Embedded via go:embed
│   ├── index.html              # Single-page app
│   ├── css/style.css           # Dark/light themes, responsive
│   ├── js/
│   │   ├── app.js              # Orchestrator, drop zones, headless relay detection
│   │   ├── auth.js             # PIN login flow
│   │   ├── ws.js               # WebSocket client, auto-reconnect
│   │   ├── webrtc.js           # RTCPeerConnection, DataChannel management
│   │   ├── transfer.js         # Chunked transfer protocol, progress tracking
│   │   ├── zip.js              # Client-side ZIP builder (STORE, CRC-32)
│   │   ├── ui.js               # Safe DOM rendering, toasts, modals
│   │   └── utils.js            # Device detection, formatting, icons, name generation
│   └── assets/
│       ├── banner.png
│       └── favicon.svg
├── dist/
│   ├── macos/                  # macOS one-click installer
│   │   ├── install.sh          # Sets up venv, LaunchAgent, starts service
│   │   ├── uninstall.sh        # Removes everything cleanly
│   │   ├── config.json         # Server address, PIN, output dir
│   │   └── gosling-client.py   # Python headless receiver
│   └── gosling-receiver.apk    # Signed Android APK
├── clients/
│   ├── python/
│   │   ├── gosling-client.py   # Cross-platform headless receiver
│   │   └── requirements.txt    # websockets
│   └── android/                # Android Studio / Gradle project
│       ├── KEYSTORE.md         # Keystore setup instructions
│       └── app/src/main/
│           ├── java/io/celox/gosling/
│           │   ├── MainActivity.kt      # Settings UI, service control
│           │   └── ReceiverService.kt   # WS connection, auto-download, extract
│           └── res/                     # Layouts, styles, icons
└── scripts/
    ├── install-service.sh      # Systemd service installer for Raspberry Pi
    └── generate-cert.sh        # Self-signed TLS certificate generator
```

## Security

- **LAN-only by design** — No STUN/TURN servers. Only host ICE candidates are used, keeping all traffic strictly on the local network.
- **Optional PIN auth** — Rate-limited to 5 failed attempts per minute per IP. Sessions stored as HTTP-only cookies and persisted to disk (survive restarts). "Stay logged in" option for permanent sessions.
- **No data leaves your network** — P2P transfers go directly between browsers. Relay uploads stay on the server's filesystem.
- **SHA-256 integrity** — File checksums verified on HTTPS connections (crypto.subtle requires secure context).
- **Path traversal protection** — All file paths are sanitized and validated against directory escape attacks.
- **No external requests** — The binary makes zero outgoing network connections. Everything runs locally.
- **Zip slip protection** — Android client validates all ZIP entry paths against directory traversal before extraction.

## Performance

| Metric | Value |
|--------|-------|
| Server binary size | ~7 MB (stripped) |
| Android APK size | ~1.9 MB (signed, minified) |
| Idle RAM (server) | ~6 MB |
| Startup time | < 1 second |
| Max concurrent peers | Tested with 10+ |
| P2P transfer speed | Limited only by LAN bandwidth |
| Relay transfer speed | Limited by server disk I/O |

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Server | Go 1.22+ (stdlib `net/http`, `go:embed`) |
| WebSocket | [gorilla/websocket](https://github.com/gorilla/websocket) |
| Config | [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) |
| QR Code | [skip2/go-qrcode](https://github.com/skip2/go-qrcode) |
| Frontend | Vanilla HTML / CSS / JS — no npm, no build step |
| P2P | WebRTC DataChannel (browser-native) |
| ZIP | Custom client-side ZIP builder (STORE method, CRC-32) |
| Python Client | `websockets` (async) or `websocket-client` (sync) |
| Android App | Kotlin, Material 3, [Java-WebSocket](https://github.com/TooTallNate/Java-WebSocket) |

## License

[MIT](LICENSE) — Martin Pfeffer / [celox.io](https://celox.io)
