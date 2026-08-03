#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
test_root="$(mktemp -d)"
mock_bin="$test_root/mock-bin"
fixture_dir="$test_root/fixture"
fixture_archive="$test_root/fixture.tar.gz"
install_dir="$test_root/installed"
rejected_dir="$test_root/rejected"
archive_name="alfred_v9.8.7_linux_amd64.tar.gz"

cleanup() {
  rm -rf "$test_root"
}
trap cleanup EXIT INT TERM

mkdir -p "$mock_bin" "$fixture_dir"
printf '#!/usr/bin/env sh\nprintf "safe fixture\\n"\n' > "$fixture_dir/alfred"
chmod +x "$fixture_dir/alfred"
tar -C "$fixture_dir" -czf "$fixture_archive" alfred
fixture_hash="$(sha256sum "$fixture_archive" | awk '{print $1}')"

cat > "$mock_bin/curl" <<'MOCK'
#!/usr/bin/env sh
set -eu

url=""
output=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    -o)
      output="$2"
      shift 2
      ;;
    -*)
      shift
      ;;
    *)
      url="$1"
      shift
      ;;
  esac
done

case "$url" in
  */checksums.txt)
    if [ "${ALFRED_TEST_BAD_CHECKSUM:-0}" = "1" ]; then
      hash="0000000000000000000000000000000000000000000000000000000000000000"
    else
      hash="$ALFRED_TEST_FIXTURE_HASH"
    fi
    printf '%s  %s\n' "$hash" "$ALFRED_TEST_ARCHIVE_NAME" > "$output"
    ;;
  */"$ALFRED_TEST_ARCHIVE_NAME")
    cp "$ALFRED_TEST_FIXTURE_ARCHIVE" "$output"
    ;;
  *)
    echo "unexpected test download: $url" >&2
    exit 1
    ;;
esac
MOCK
chmod +x "$mock_bin/curl"

cat > "$mock_bin/uname" <<'MOCK'
#!/usr/bin/env sh
case "${1:-}" in
  -s) printf '%s\n' "Linux" ;;
  -m) printf '%s\n' "x86_64" ;;
  *) printf '%s\n' "Linux" ;;
esac
MOCK
chmod +x "$mock_bin/uname"

PATH="$mock_bin:$PATH" \
ALFRED_VERSION="v9.8.7" \
ALFRED_REPO="example/alfred" \
ALFRED_INSTALL_DIR="$install_dir" \
ALFRED_TEST_FIXTURE_ARCHIVE="$fixture_archive" \
ALFRED_TEST_FIXTURE_HASH="$fixture_hash" \
ALFRED_TEST_ARCHIVE_NAME="$archive_name" \
sh "$script_dir/install.sh"

if [[ ! -f "$install_dir/alfred" ]]; then
  echo "installer did not copy the verified binary" >&2
  exit 1
fi

if ! grep -Fq "safe fixture" "$install_dir/alfred"; then
  echo "installed binary does not match the verified fixture" >&2
  exit 1
fi

if PATH="$mock_bin:$PATH" \
  ALFRED_VERSION="v9.8.7" \
  ALFRED_REPO="example/alfred" \
  ALFRED_INSTALL_DIR="$rejected_dir" \
  ALFRED_TEST_FIXTURE_ARCHIVE="$fixture_archive" \
  ALFRED_TEST_FIXTURE_HASH="$fixture_hash" \
  ALFRED_TEST_ARCHIVE_NAME="$archive_name" \
  ALFRED_TEST_BAD_CHECKSUM="1" \
  sh "$script_dir/install.sh" >/dev/null 2>&1; then
  echo "installer accepted an archive with an invalid checksum" >&2
  exit 1
fi

if [[ -f "$rejected_dir/alfred" ]]; then
  echo "installer copied a binary after checksum verification failed" >&2
  exit 1
fi

echo "shell installer tests passed"
