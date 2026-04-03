#!/bin/bash
set -e

CERT_DIR="certs"
mkdir -p "${CERT_DIR}"

echo "Generating self-signed TLS certificate..."

openssl req -x509 -newkey rsa:4096 -sha256 -days 3650 \
    -nodes -keyout "${CERT_DIR}/key.pem" -out "${CERT_DIR}/cert.pem" \
    -subj "/CN=go-sling" \
    -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

echo ""
echo "Certificate generated:"
echo "  Cert: ${CERT_DIR}/cert.pem"
echo "  Key:  ${CERT_DIR}/key.pem"
