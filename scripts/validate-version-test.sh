#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
validator="$script_dir/validate-version.sh"

valid_versions=(
  "v0.1.0"
  "v1.0.0"
  "v10.24.300"
)

invalid_versions=(
  ""
  "0.1.0"
  "v01.2.3"
  "v1.02.3"
  "v1.2.03"
  "v1.2"
  "v1.2.3-rc.1"
  $'v1.2.3\nunsafe'
)

for version in "${valid_versions[@]}"; do
  bash "$validator" "$version"
done

for version in "${invalid_versions[@]}"; do
  if bash "$validator" "$version" >/dev/null 2>&1; then
    printf 'expected version %q to be rejected\n' "$version" >&2
    exit 1
  fi
done

printf 'version validation tests passed\n'
