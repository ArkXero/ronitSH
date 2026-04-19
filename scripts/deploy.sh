#!/bin/bash
# Deploy ronit-sh to the Ubuntu server.
# Run via: make deploy (which builds first)
set -euo pipefail

SERVER="arkx@192.168.0.249"
REMOTE_DIR="/opt/termfolio/bin"
SERVICE="ronit-sh"

if [[ ! -f dist/ronit-sh ]]; then
  echo "ERROR: dist/ronit-sh not found. Run 'make build' first."
  exit 1
fi

echo "About to deploy to $SERVER"
echo "  ronit-sh    -> $REMOTE_DIR/ronit-sh"
echo "  ronit-sh-ctl -> $REMOTE_DIR/ronit-sh-ctl"
echo ""
read -rp "Continue? (y/n) " -n 1 REPLY
echo ""

if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

echo "Copying binaries..."
scp dist/ronit-sh dist/ronit-sh-ctl "$SERVER:$REMOTE_DIR/"

echo "Restarting service..."
ssh "$SERVER" "sudo systemctl restart $SERVICE && sudo systemctl status $SERVICE --no-pager"

echo "Deployed."
