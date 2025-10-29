#!/bin/bash
# Nerve Agent Installation Script

set -e

# Default values
SERVER_URL="${SERVER_URL:-http://localhost:8080}"
TOKEN="${TOKEN:-}"

# Parse arguments
while [[ "$#" -gt 0 ]]; do
  case $1 in
    --token) TOKEN="$2"; shift ;;
    --server) SERVER_URL="$2"; shift ;;
    --help) echo "Usage: $0 --token=<token> --server=<url>"; exit 0 ;;
  esac
  shift
done

if [ -z "$TOKEN" ]; then
  echo "Error: --token is required"
  echo "Usage: $0 --token=<your-token> [--server=<server-url>]"
  exit 1
fi

echo "========================================="
echo "  Nerve Agent Installation"
echo "========================================="
echo "Server: $SERVER_URL"
echo ""

# Check if already installed
if systemctl is-active --quiet nerve-agent; then
  echo "Nerve Agent is already running. Stopping..."
  systemctl stop nerve-agent
fi

# Download agent binary
echo "Downloading Nerve Agent..."
curl -fsSL "${SERVER_URL}/download?token=${TOKEN}" -o /usr/local/bin/nerve-agent
chmod +x /usr/local/bin/nerve-agent

# Verify download
if [ ! -f /usr/local/bin/nerve-agent ]; then
  echo "Error: Failed to download agent binary"
  exit 1
fi

# Create systemd service
echo "Creating systemd service..."
cat > /etc/systemd/system/nerve-agent.service <<EOF
[Unit]
Description=Nerve Agent - Distributed Infrastructure Intelligence
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/nerve-agent --server=${SERVER_URL} --token=${TOKEN}
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/nerve-agent

[Install]
WantedBy=multi-user.target
EOF

# Create data directory
mkdir -p /var/lib/nerve-agent

# Reload systemd and enable service
echo "Starting Nerve Agent..."
systemctl daemon-reload
systemctl enable nerve-agent
systemctl start nerve-agent

# Wait for service to start
sleep 2

# Check status
if systemctl is-active --quiet nerve-agent; then
  echo ""
  echo "✓ Nerve Agent installed and running successfully!"
  echo ""
  echo "Useful commands:"
  echo "  systemctl status nerve-agent  # Check status"
  echo "  journalctl -u nerve-agent -f  # View logs"
  echo "  systemctl stop nerve-agent    # Stop agent"
else
  echo "Error: Failed to start Nerve Agent"
  systemctl status nerve-agent || true
  exit 1
fi

