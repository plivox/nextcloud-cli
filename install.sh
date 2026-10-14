#!/bin/sh
# Install the latest nextcloud-cli release (macOS and Linux).
#
#   curl -fsSL https://raw.githubusercontent.com/plivox/nextcloud-cli/main/install.sh | sh
#
# Environment variables:
#   VERSION      release tag to install, e.g. v0.1.0 (default: latest)
#   INSTALL_DIR  where to put the binary (default: /usr/local/bin)

set -eu

REPO="plivox/nextcloud-cli"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

os=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$os" in
  darwin | linux) ;;
  *) echo "Unsupported OS: $os (download a release manually)" >&2; exit 1 ;;
esac

arch=$(uname -m)
case "$arch" in
  x86_64 | amd64) arch=amd64 ;;
  arm64 | aarch64) arch=arm64 ;;
  *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  base="https://github.com/$REPO/releases/latest/download"
else
  base="https://github.com/$REPO/releases/download/$VERSION"
fi

archive="nextcloud-cli_${os}_${arch}.tar.gz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

echo "Downloading $archive ($VERSION)..."
curl -fsSL -o "$tmp/$archive" "$base/$archive"
curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt"

expected=$(grep " $archive\$" "$tmp/checksums.txt" | cut -d ' ' -f 1)
if command -v sha256sum >/dev/null 2>&1; then
  actual=$(sha256sum "$tmp/$archive" | cut -d ' ' -f 1)
else
  actual=$(shasum -a 256 "$tmp/$archive" | cut -d ' ' -f 1)
fi

if [ -z "$expected" ] || [ "$expected" != "$actual" ]; then
  echo "Checksum mismatch for $archive" >&2
  exit 1
fi

tar -xzf "$tmp/$archive" -C "$tmp" nextcloud

if [ -w "$INSTALL_DIR" ]; then
  mv "$tmp/nextcloud" "$INSTALL_DIR/nextcloud"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$tmp/nextcloud" "$INSTALL_DIR/nextcloud"
fi

echo "Installed $("$INSTALL_DIR/nextcloud" --version) to $INSTALL_DIR/nextcloud"
