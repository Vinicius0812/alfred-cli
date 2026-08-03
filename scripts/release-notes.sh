#!/usr/bin/env bash
set -euo pipefail

version="${1:-}"
changelog="${2:-CHANGELOG.md}"
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

bash "$script_dir/validate-version.sh" "$version"

if [[ ! -f "$changelog" ]]; then
  echo "changelog not found: $changelog" >&2
  exit 1
fi

notes="$(awk -v heading="## $version" '
  $0 == heading { found = 1; next }
  found && /^## / { exit }
  found { print }
  END { if (!found) exit 2 }
' "$changelog")" || {
  echo "release section ## $version not found in $changelog" >&2
  exit 1
}

if ! printf '%s\n' "$notes" | grep -q '[^[:space:]]'; then
  echo "release section ## $version is empty in $changelog" >&2
  exit 1
fi

printf '%s\n' "$notes"
