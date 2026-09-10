#!/bin/sh
set -eu

REPO="autokeren/kerenscope"
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux*) OS="linux" ;;
  Darwin*) OS="darwin" ;;
  *) echo "✗ Unsupported OS: $OS (this installer covers Linux and macOS; on Windows use install.ps1)" ; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "✗ Unsupported architecture: $ARCH" ; exit 1 ;;
esac

BASE="https://github.com/${REPO}/releases/latest/download"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

echo "→ Downloading kerenscope for ${OS}-${ARCH}…"
curl -fsSL "${BASE}/kerenscope-${OS}-${ARCH}.tar.gz" -o "${TMP}/kerenscope.tar.gz"

BIN_DIR="${HOME}/.local/bin"
mkdir -p "$BIN_DIR"

echo "→ Installing to ${BIN_DIR}…"
tar -xzf "${TMP}/kerenscope.tar.gz" -C "$BIN_DIR"
mv "${BIN_DIR}/kerenscope-${OS}-${ARCH}" "${BIN_DIR}/keren"
chmod +x "${BIN_DIR}/keren"

case ":$PATH:" in
  *":${BIN_DIR}:"*) ;;
  *)
    echo ""
    echo "⚠  $BIN_DIR is not in your PATH. Add it with:"
    echo "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc"
    ;;
esac

echo ""
"${BIN_DIR}/keren" --version
echo "✓ Installed: keren — try: keren company BBCA"
