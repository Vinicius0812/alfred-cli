#!/usr/bin/env sh
set -eu

REPO="${ALFRED_REPO:-Vinicius0812/alfred-cli}"
VERSION="${ALFRED_VERSION:-latest}"
INSTALL_DIR="${ALFRED_INSTALL_DIR:-$HOME/.local/bin}"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

detect_os() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux) printf '%s' "linux" ;;
    darwin) printf '%s' "darwin" ;;
    *) echo "unsupported operating system: $os" >&2; exit 1 ;;
  esac
}

detect_arch() {
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) printf '%s' "amd64" ;;
    arm64|aarch64) printf '%s' "arm64" ;;
    *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
  esac
}

latest_version() {
  if command_exists curl; then
    curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1
  elif command_exists wget; then
    wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n 1
  else
    echo "curl or wget is required" >&2
    exit 1
  fi
}

download() {
  url="$1"
  output="$2"
  if command_exists curl; then
    curl -fL "$url" -o "$output"
  elif command_exists wget; then
    wget -O "$output" "$url"
  else
    echo "curl or wget is required" >&2
    exit 1
  fi
}

if [ "$VERSION" = "latest" ]; then
  VERSION="$(latest_version)"
fi

if [ -z "$VERSION" ]; then
  echo "could not determine Alfred version" >&2
  exit 1
fi

os="$(detect_os)"
arch="$(detect_arch)"
archive="alfred_${VERSION}_${os}_${arch}.tar.gz"
url="https://github.com/$REPO/releases/download/$VERSION/$archive"
tmp_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

mkdir -p "$INSTALL_DIR"
download "$url" "$tmp_dir/$archive"
tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"
install "$tmp_dir/alfred" "$INSTALL_DIR/alfred"

echo "Alfred installed at $INSTALL_DIR/alfred"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Add $INSTALL_DIR to your PATH to run 'alfred' from anywhere." ;;
esac
