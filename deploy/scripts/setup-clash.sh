#!/usr/bin/env bash
set -euo pipefail

MIHOMO_VERSION="v1.19.0"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/mihomo"
ARCH=$(uname -m)

case "$ARCH" in
    x86_64) BINARY="mihomo-linux-amd64" ;;
    aarch64) BINARY="mihomo-linux-arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

echo "=== Installing mihomo ${MIHOMO_VERSION} for ${ARCH} ==="

if command -v mihomo &>/dev/null; then
    echo "mihomo already installed: $(mihomo -v 2>/dev/null || echo 'unknown version')"
    echo "Reinstalling..."
fi

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading mihomo..."
curl -L -o "${TMPDIR}/mihomo.gz" \
    "https://github.com/MetaCubeX/mihomo/releases/download/${MIHOMO_VERSION}/${BINARY}-${MIHOMO_VERSION}.gz"

echo "Installing to ${INSTALL_DIR}..."
gunzip "${TMPDIR}/mihomo.gz"
chmod +x "${TMPDIR}/mihomo"
sudo cp "${TMPDIR}/mihomo" "${INSTALL_DIR}/mihomo"

echo "Setting up config directory..."
sudo mkdir -p "${CONFIG_DIR}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CLASH_CONFIG="${SCRIPT_DIR}/../clash/config.yaml"

if [ -f "$CLASH_CONFIG" ]; then
    if [ ! -f "${CONFIG_DIR}/config.yaml" ]; then
        sudo cp "$CLASH_CONFIG" "${CONFIG_DIR}/config.yaml"
        echo "Config copied to ${CONFIG_DIR}/config.yaml"
        echo "IMPORTANT: Edit ${CONFIG_DIR}/config.yaml to set your proxy server and password"
    else
        echo "Config already exists at ${CONFIG_DIR}/config.yaml, not overwriting"
    fi
else
    echo "WARNING: ${CLASH_CONFIG} not found"
    echo "Please manually create ${CONFIG_DIR}/config.yaml"
fi

echo "Creating systemd service..."
sudo tee /etc/systemd/system/mihomo.service > /dev/null <<'EOF'
[Unit]
Description=Mihomo (Clash Meta)
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/mihomo -d /etc/mihomo
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable mihomo

if [ -f "${CONFIG_DIR}/config.yaml" ]; then
    sudo systemctl start mihomo
    sleep 2
    if systemctl is-active --quiet mihomo; then
        echo "=== mihomo started successfully ==="
        echo "Listening on: 127.0.0.1:7890"
    else
        echo "ERROR: mihomo failed to start"
        sudo journalctl -u mihomo -n 20 --no-pager
        exit 1
    fi
else
    echo "=== mihomo installed but not started ==="
    echo "Edit ${CONFIG_DIR}/config.yaml then run: sudo systemctl start mihomo"
fi

echo ""
echo "=== Verification ==="
echo "mihomo version: $(mihomo -v 2>/dev/null || echo 'not found')"
echo "Config: ${CONFIG_DIR}/config.yaml"
echo "Service: $(systemctl is-active mihomo 2>/dev/null || echo 'not running')"
