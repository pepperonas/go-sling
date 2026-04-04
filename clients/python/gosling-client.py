#!/usr/bin/env python3
"""
go-sling headless client — auto-receives files from go-sling server.

Connects as a headless peer via WebSocket. When a browser sends files
to this client, they are automatically downloaded and ZIP files are extracted.

Usage:
    python gosling-client.py --server 192.168.178.103:8420 --pin 3001 --output ~/Downloads/go-sling
"""

import argparse
import json
import os
import platform
import signal
import sys
import time
import zipfile
from pathlib import Path
from urllib.parse import urljoin
from urllib.request import Request, urlopen
from urllib.error import URLError

try:
    import websockets
    import asyncio
    HAS_WEBSOCKETS = True
except ImportError:
    HAS_WEBSOCKETS = False

# Also support running without websockets via a simple sync WS client
try:
    from websocket import WebSocketApp
    HAS_WEBSOCKET_CLIENT = True
except ImportError:
    HAS_WEBSOCKET_CLIENT = False


def detect_device():
    system = platform.system()
    machine = platform.machine()
    if system == "Darwin":
        return "macOS"
    elif system == "Windows":
        return "Windows"
    elif system == "Linux":
        if "arm" in machine or "aarch" in machine:
            return "Linux ARM"
        return "Linux"
    return system


def generate_name():
    import random
    import string
    device = detect_device()
    adjectives = {
        "macOS": ["Lunar", "Stellar", "Cosmic", "Nebula", "Astro"],
        "Windows": ["Pixel", "Cyber", "Neon", "Prism", "Volt"],
        "Linux": ["Kernel", "Root", "Daemon", "Cron", "Shell"],
        "Linux ARM": ["Berry", "Tiny", "Micro", "Nano", "Pico"],
    }
    nouns = {
        "macOS": ["Agent", "Daemon", "Catcher", "Receiver", "Vault"],
        "Windows": ["Agent", "Daemon", "Catcher", "Receiver", "Vault"],
        "Linux": ["Agent", "Daemon", "Catcher", "Receiver", "Vault"],
        "Linux ARM": ["Agent", "Daemon", "Catcher", "Receiver", "Vault"],
    }
    adj = random.choice(adjectives.get(device, ["Auto"]))
    noun = random.choice(nouns.get(device, ["Client"]))
    suffix = ''.join(random.choices(string.ascii_lowercase + string.digits, k=3))
    return f"{adj}-{noun}-{suffix}"


class GoslingClient:
    def __init__(self, server, pin, output_dir, extract=True, name=None):
        self.server = server.rstrip("/")
        self.pin = pin
        self.output_dir = Path(output_dir).expanduser().resolve()
        self.extract = extract
        self.name = name or generate_name()
        self.session_cookie = None
        self.running = True
        self.base_url = f"http://{self.server}"

        self.output_dir.mkdir(parents=True, exist_ok=True)

    def log(self, msg):
        ts = time.strftime("%H:%M:%S")
        print(f"[{ts}] {msg}", flush=True)

    def authenticate(self):
        """Authenticate with the server and get a session cookie."""
        url = f"{self.base_url}/api/auth/status"
        try:
            req = Request(url)
            resp = urlopen(req, timeout=5)
            data = json.loads(resp.read())
            if not data.get("required"):
                self.log("No authentication required")
                return True
        except Exception as e:
            self.log(f"Failed to check auth status: {e}")
            return False

        if not self.pin:
            self.log("ERROR: Server requires PIN but none provided (use --pin)")
            return False

        url = f"{self.base_url}/api/auth"
        payload = json.dumps({"pin": self.pin, "remember": True}).encode()
        req = Request(url, data=payload, headers={"Content-Type": "application/json"})
        try:
            resp = urlopen(req, timeout=5)
            # Extract session cookie
            cookie_header = resp.headers.get("Set-Cookie", "")
            for part in cookie_header.split(";"):
                part = part.strip()
                if part.startswith("gosling_session="):
                    self.session_cookie = part
                    break
            self.log("Authenticated successfully")
            return True
        except URLError as e:
            self.log(f"Authentication failed: {e}")
            return False

    def download_file(self, file_id, file_name):
        """Download a file from the server."""
        url = f"{self.base_url}/api/download/{file_id}"
        headers = {}
        if self.session_cookie:
            headers["Cookie"] = self.session_cookie

        req = Request(url, headers=headers)
        try:
            resp = urlopen(req, timeout=300)
            file_path = self.output_dir / file_name
            file_path.parent.mkdir(parents=True, exist_ok=True)

            size = 0
            with open(file_path, "wb") as f:
                while True:
                    chunk = resp.read(65536)
                    if not chunk:
                        break
                    f.write(chunk)
                    size += len(chunk)

            self.log(f"Downloaded: {file_name} ({self._fmt_size(size)})")

            # Auto-extract ZIP files
            if self.extract and file_name.lower().endswith(".zip"):
                self._extract_zip(file_path)

            return file_path
        except Exception as e:
            self.log(f"Download failed for {file_name}: {e}")
            return None

    def _extract_zip(self, zip_path):
        """Extract a ZIP file and remove the archive."""
        try:
            extract_dir = zip_path.parent / zip_path.stem
            with zipfile.ZipFile(zip_path, "r") as zf:
                zf.extractall(extract_dir)
            os.remove(zip_path)
            self.log(f"Extracted: {zip_path.name} → {extract_dir.name}/")
        except zipfile.BadZipFile:
            self.log(f"Not a valid ZIP: {zip_path.name}, keeping as-is")
        except Exception as e:
            self.log(f"Extract failed: {e}")

    def _fmt_size(self, size):
        for unit in ["B", "KB", "MB", "GB"]:
            if size < 1024:
                return f"{size:.1f} {unit}"
            size /= 1024
        return f"{size:.1f} TB"

    def run_sync(self):
        """Run using websocket-client (synchronous)."""
        import websocket

        ws_url = f"ws://{self.server}/ws"

        def on_open(ws):
            self.log(f"Connected to {self.server} as '{self.name}'")
            join_msg = json.dumps({
                "type": "join",
                "payload": {
                    "name": self.name,
                    "os": detect_device(),
                    "browser": "Python CLI",
                    "headless": True,
                }
            })
            ws.send(join_msg)

        def on_message(ws, message):
            try:
                msg = json.loads(message)
                self._handle_message(msg)
            except json.JSONDecodeError:
                pass

        def on_error(ws, error):
            self.log(f"WebSocket error: {error}")

        def on_close(ws, code, reason):
            self.log("Disconnected, reconnecting in 5s...")

        while self.running:
            try:
                cookie = self.session_cookie if self.session_cookie else None
                ws = websocket.WebSocketApp(
                    ws_url,
                    on_open=on_open,
                    on_message=on_message,
                    on_error=on_error,
                    on_close=on_close,
                    cookie=cookie,
                )
                ws.run_forever(ping_interval=30, ping_timeout=10)
            except Exception as e:
                self.log(f"Connection error: {e}")

            if self.running:
                time.sleep(5)

    async def run_async(self):
        """Run using websockets (async)."""
        ws_url = f"ws://{self.server}/ws"
        extra_headers = {}
        if self.session_cookie:
            extra_headers["Cookie"] = self.session_cookie

        while self.running:
            try:
                async with websockets.connect(ws_url, extra_headers=extra_headers) as ws:
                    self.log(f"Connected to {self.server} as '{self.name}'")

                    join_msg = json.dumps({
                        "type": "join",
                        "payload": {
                            "name": self.name,
                            "os": detect_device(),
                            "browser": "Python CLI",
                            "headless": True,
                        }
                    })
                    await ws.send(join_msg)

                    async for message in ws:
                        try:
                            msg = json.loads(message)
                            self._handle_message(msg)
                        except json.JSONDecodeError:
                            pass

            except Exception as e:
                self.log(f"Connection error: {e}")

            if self.running:
                self.log("Reconnecting in 5s...")
                await asyncio.sleep(5)

    def _handle_message(self, msg):
        msg_type = msg.get("type")

        if msg_type == "welcome":
            peer_id = msg.get("payload", {}).get("id", "?")
            self.log(f"Registered as peer {peer_id}")

        elif msg_type == "peer-list":
            peers = msg.get("peers", [])
            other = [p for p in peers if not p.get("headless") or p.get("name") != self.name]
            self.log(f"Peers online: {len(peers)} ({len(other)} browsers)")

        elif msg_type == "file-ready":
            payload = msg.get("payload", {})
            file_id = payload.get("id")
            file_name = payload.get("name", "unknown")
            file_size = payload.get("size", 0)
            self.log(f"Incoming: {file_name} ({self._fmt_size(file_size)})")
            self.download_file(file_id, file_name)

    def run(self):
        self.log(f"go-sling client v1.0.0")
        self.log(f"Output: {self.output_dir}")

        if not self.authenticate():
            sys.exit(1)

        if HAS_WEBSOCKETS:
            self.log("Using websockets (async)")
            asyncio.run(self.run_async())
        elif HAS_WEBSOCKET_CLIENT:
            self.log("Using websocket-client (sync)")
            self.run_sync()
        else:
            self.log("ERROR: No WebSocket library found. Install one:")
            self.log("  pip install websockets")
            self.log("  pip install websocket-client")
            sys.exit(1)


def main():
    parser = argparse.ArgumentParser(
        description="go-sling headless client — auto-receive files from your LAN",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  %(prog)s --server 192.168.1.42:8420
  %(prog)s --server 192.168.1.42:8420 --pin 3001 --output ~/received
  %(prog)s --server 192.168.1.42:8420 --no-extract --name my-laptop
        """,
    )
    parser.add_argument("--server", "-s", required=True, help="go-sling server address (host:port)")
    parser.add_argument("--pin", "-p", default="", help="Authentication PIN")
    parser.add_argument("--output", "-o", default="~/go-sling-received", help="Download directory (default: ~/go-sling-received)")
    parser.add_argument("--no-extract", action="store_true", help="Don't auto-extract ZIP files")
    parser.add_argument("--name", "-n", default=None, help="Custom peer name")

    args = parser.parse_args()

    client = GoslingClient(
        server=args.server,
        pin=args.pin,
        output_dir=args.output,
        extract=not args.no_extract,
        name=args.name,
    )

    def handle_signal(sig, frame):
        client.running = False
        client.log("Shutting down...")
        sys.exit(0)

    signal.signal(signal.SIGINT, handle_signal)
    signal.signal(signal.SIGTERM, handle_signal)

    client.run()


if __name__ == "__main__":
    main()
