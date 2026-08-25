#!/bin/sh
# archscan installer — https://github.com/archscan/archscan
set -e

REPO="archscan/archscan"
INSTALL_DIR="$HOME/.local/bin"
BIN="archscan"

# Detect OS and arch
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

echo "⚡ Installing archscan..."
echo "   Platform: ${OS}/${ARCH}"

# Get latest release tag
VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
  | grep '"tag_name"' | sed 's/.*"tag_name": "\(.*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "Error: could not determine latest version."
  exit 1
fi

echo "   Version: $VERSION"

# Build download URL
FILENAME="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"

# Download and extract
TMP=$(mktemp -d)
trap "rm -rf $TMP" EXIT

echo "   Downloading $FILENAME..."
curl -fsSL "$URL" -o "$TMP/$FILENAME"
tar -xzf "$TMP/$FILENAME" -C "$TMP"

# Install binary
mkdir -p "$INSTALL_DIR"
mv "$TMP/$BIN" "$INSTALL_DIR/$BIN"
chmod +x "$INSTALL_DIR/$BIN"

echo ""
echo "✓ archscan installed to $INSTALL_DIR/$BIN"

# Check PATH
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  echo ""
  echo "  Add to your PATH by adding this to ~/.bashrc or ~/.zshrc:"
  echo "    export PATH=\"\$PATH:$INSTALL_DIR\""
fi

echo ""
echo "  Run: archscan --help"
echo "  Pro: archscan activate --email you@example.com --key ARCHSCAN-..."
echo "  Buy: https://polar.sh/archscan"
echo ""
