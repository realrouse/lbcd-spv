#!/usr/bin/env bash
# Download a compact-filter-capable LBC full node binary for harness use.
#
# Default: official lbryio/lbcd GitHub release (published builds exist today).
# Override with LBCD_BIN, LBCD_VERSION, LBCD_RELEASE_URL, or LBCD_REPO.
# Never clones a git org into the wallet; this script only fetches a tarball.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="${LBCD_DEST:-$ROOT/bin}"
VERSION="${LBCD_VERSION:-v0.22.119}"
REPO="${LBCD_REPO:-lbryio/lbcd}"

if [[ -n "${LBCD_BIN:-}" && -x "${LBCD_BIN}" ]]; then
  echo "Using existing LBCD_BIN=${LBCD_BIN}"
  exit 0
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  aarch64|arm64) arch=arm64 ;;
  *)
    echo "Unsupported arch: $arch" >&2
    exit 1
    ;;
esac
case "$os" in
  linux|darwin) ;;
  mingw*|msys*|cygwin*) os=windows ;;
  *)
    echo "Unsupported OS: $os" >&2
    exit 1
    ;;
esac

# Strip leading v for the asset name used by goreleaser.
ver_num="${VERSION#v}"
asset="lbcd_${ver_num}_${os}_${arch}.tar.gz"
url="${LBCD_RELEASE_URL:-https://github.com/${REPO}/releases/download/${VERSION}/${asset}}"

mkdir -p "$DEST"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

echo "Fetching $url"
if ! curl -fsSL "$url" -o "$tmp/$asset"; then
  echo "Download failed. Set LBCD_BIN to a local daemon, or LBCD_RELEASE_URL to another tarball." >&2
  echo "Foundation builds: set LBCD_REPO=LBRYFoundation/lbcd once that repo publishes releases." >&2
  exit 1
fi

tar -xzf "$tmp/$asset" -C "$tmp"
# Tarball layout varies; find the lbcd binary.
found="$(find "$tmp" -type f -name lbcd -perm -111 | head -1)"
if [[ -z "$found" ]]; then
  # windows
  found="$(find "$tmp" -type f -name 'lbcd.exe' | head -1)"
fi
if [[ -z "$found" ]]; then
  echo "No lbcd binary in archive" >&2
  tar -tzf "$tmp/$asset" | head
  exit 1
fi

cp "$found" "$DEST/lbcd"
chmod +x "$DEST/lbcd"

ctl="$(find "$tmp" -type f \( -name lbcctl -o -name lbcctl.exe \) | head -1)"
if [[ -n "$ctl" ]]; then
  cp "$ctl" "$DEST/lbcctl"
  chmod +x "$DEST/lbcctl"
fi

echo "Installed:"
ls -l "$DEST/lbcd" "$DEST/lbcctl" 2>/dev/null || ls -l "$DEST/lbcd"
echo "Point harness at it with: export LBCD_BIN=$DEST/lbcd"
