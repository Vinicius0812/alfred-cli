#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
test_root="$(mktemp -d)"
changelog="$test_root/CHANGELOG.md"

cleanup() {
  rm -rf "$test_root"
}
trap cleanup EXIT INT TERM

cat > "$changelog" <<'CHANGELOG'
# Changelog

## Unreleased

- Work in progress.

## v1.2.3

- Secure release notes.

## v1.2.2

- Previous release.
CHANGELOG

notes="$(bash "$script_dir/release-notes.sh" "v1.2.3" "$changelog")"
if [[ "$notes" != *"Secure release notes."* ]] || [[ "$notes" == *"Previous release."* ]]; then
  echo "release notes extraction returned the wrong section" >&2
  exit 1
fi

if bash "$script_dir/release-notes.sh" "v9.9.9" "$changelog" >/dev/null 2>&1; then
  echo "missing release section was accepted" >&2
  exit 1
fi

echo "release notes tests passed"
