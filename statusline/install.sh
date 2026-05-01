#!/usr/bin/env bash
set -e

REPO="coredgeio/claude-tools"
BINARY="claudeline"
INSTALL_DIR="${HOME}/.claude"
SETTINGS="${INSTALL_DIR}/settings.json"

# detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac
case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS. For Windows use install.ps1"; exit 1 ;;
esac

ASSET="${BINARY}-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/statusline-latest/${ASSET}"

echo "Downloading ${ASSET}..."
mkdir -p "$INSTALL_DIR"
if command -v curl &>/dev/null; then
  curl -fsSL "$URL" -o "${INSTALL_DIR}/${BINARY}"
elif command -v wget &>/dev/null; then
  wget -qO "${INSTALL_DIR}/${BINARY}" "$URL"
else
  echo "curl or wget required"; exit 1
fi
chmod +x "${INSTALL_DIR}/${BINARY}"

# patch settings.json
COMMAND="${INSTALL_DIR}/${BINARY}"
if command -v jq &>/dev/null; then
  if [ -f "$SETTINGS" ]; then
    tmp=$(mktemp)
    jq --arg cmd "$COMMAND" '.statusLine.command = $cmd' "$SETTINGS" > "$tmp"
    mv "$tmp" "$SETTINGS"
  else
    echo "{\"statusLine\":{\"command\":\"${COMMAND}\"}}" > "$SETTINGS"
  fi
else
  echo "Warning: jq not found. Add this to ${SETTINGS} manually:"
  echo "  \"statusLine\": { \"command\": \"${COMMAND}\" }"
fi

echo "Installed to ${INSTALL_DIR}/${BINARY}"
echo "Restart Claude Code to apply."
