#!/bin/bash
set -e

APP_NAME="go-sling"
INSTALL_DIR="/home/pi/${APP_NAME}"
BINARY="${INSTALL_DIR}/${APP_NAME}-linux-arm7"
SERVICE_FILE="/etc/systemd/system/${APP_NAME}.service"

echo "Installing ${APP_NAME} systemd service..."

# Create install directory
sudo mkdir -p "${INSTALL_DIR}"

# Copy binary if it exists locally
if [ -f "bin/${APP_NAME}-linux-arm7" ]; then
    sudo cp "bin/${APP_NAME}-linux-arm7" "${BINARY}"
    sudo chmod +x "${BINARY}"
    echo "Binary copied to ${BINARY}"
fi

# Create systemd service
sudo tee "${SERVICE_FILE}" > /dev/null << EOF
[Unit]
Description=go-sling - LAN File Transfer
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=pi
WorkingDirectory=${INSTALL_DIR}
ExecStart=${BINARY}
Restart=always
RestartSec=5
Environment=PIN=changeme

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable "${APP_NAME}"
sudo systemctl start "${APP_NAME}"

echo ""
echo "Service installed and started!"
echo "  Status:  sudo systemctl status ${APP_NAME}"
echo "  Logs:    sudo journalctl -u ${APP_NAME} -f"
echo "  Stop:    sudo systemctl stop ${APP_NAME}"
echo "  Restart: sudo systemctl restart ${APP_NAME}"
