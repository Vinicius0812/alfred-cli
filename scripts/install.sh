#!/usr/bin/env sh
set -eu

REPO="${ALFRED_REPO:-Vinicius0812/alfred-cli}"
VERSION="${ALFRED_VERSION:-latest}"
INSTALL_DIR="${ALFRED_INSTALL_DIR:-$HOME/.local/bin}"

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

validate_repo() {
  owner="${REPO%%/*}"
  name="${REPO#*/}"

  if [ -z "$owner" ] || [ -z "$name" ] || [ "$name" = "$REPO" ]; then
    echo "invalid repository: expected owner/repository" >&2
    exit 1
  fi
  case "$name" in
    */*) echo "invalid repository: expected owner/repository" >&2; exit 1 ;;
  esac
  case "$owner$name" in
    *[!A-Za-z0-9._-]*) echo "invalid repository: expected owner/repository" >&2; exit 1 ;;
  esac
}

validate_version() {
  if ! printf '%s\n' "$VERSION" | grep -Eq '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'; then
    echo "invalid Alfred version: expected vMAJOR.MINOR.PATCH" >&2
    exit 1
  fi
  case "$VERSION" in
    *'
'*) echo "invalid Alfred version: expected vMAJOR.MINOR.PATCH" >&2; exit 1 ;;
  esac
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

calculate_sha256() {
  file="$1"
  if command_exists sha256sum; then
    sha256sum "$file" | awk '{print tolower($1)}'
  elif command_exists shasum; then
    shasum -a 256 "$file" | awk '{print tolower($1)}'
  elif command_exists openssl; then
    openssl dgst -sha256 "$file" | awk '{print tolower($NF)}'
  else
    echo "sha256sum, shasum, or openssl is required to verify the download" >&2
    exit 1
  fi
}

expected_sha256() {
  checksum_file="$1"
  filename="$2"
  awk -v name="$filename" '$2 == name || $2 == "*" name { print tolower($1) }' "$checksum_file"
}

validate_repo

if [ "$VERSION" = "latest" ]; then
  VERSION="$(latest_version)"
fi

validate_version

command_exists tar || { echo "tar is required" >&2; exit 1; }
command_exists install || { echo "install is required" >&2; exit 1; }

os="$(detect_os)"
arch="$(detect_arch)"
archive="alfred_${VERSION}_${os}_${arch}.tar.gz"
release_base_url="https://github.com/$REPO/releases/download/$VERSION"
archive_url="$release_base_url/$archive"
checksum_url="$release_base_url/checksums.txt"
tmp_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

mkdir -p "$INSTALL_DIR"
download "$archive_url" "$tmp_dir/$archive"
download "$checksum_url" "$tmp_dir/checksums.txt"

expected="$(expected_sha256 "$tmp_dir/checksums.txt" "$archive")"
case "$expected" in
  *'
'*) echo "could not find exactly one valid checksum for $archive" >&2; exit 1 ;;
esac
if ! printf '%s\n' "$expected" | grep -Eq '^[a-f0-9]{64}$'; then
  echo "could not find exactly one valid checksum for $archive" >&2
  exit 1
fi

actual="$(calculate_sha256 "$tmp_dir/$archive")"
if [ "$actual" != "$expected" ]; then
  echo "checksum verification failed for $archive" >&2
  exit 1
fi

echo "Verified SHA-256 checksum for $archive"
tar -xzf "$tmp_dir/$archive" -C "$tmp_dir"

if [ ! -f "$tmp_dir/alfred" ]; then
  echo "the downloaded archive does not contain alfred" >&2
  exit 1
fi

install -m 0755 "$tmp_dir/alfred" "$INSTALL_DIR/alfred"

echo "Alfred installed at $INSTALL_DIR/alfred"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) echo "Add $INSTALL_DIR to your PATH to run 'alfred' from anywhere." ;;
esac
