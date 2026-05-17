#!/usr/bin/env bash
set -euo pipefail

# Generate THIRD_PARTY_LICENSES.txt for release redistribution.
# Requires: go install github.com/google/go-licenses/v2@v2.0.1

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root"

out="${1:-THIRD_PARTY_LICENSES.txt}"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

go install github.com/google/go-licenses/v2@v2.0.1
go-licenses save ./... --save_path="$tmp/licenses" --force

{
  echo "Third-party licenses for connectivity"
  echo "Generated: $(date -u +'%Y-%m-%dT%H:%M:%SZ')"
  echo
  find "$tmp/licenses" -type f | sort | while read -r f; do
    echo "================================================================================"
    echo "${f#"$tmp/licenses"/}"
    echo "================================================================================"
    cat "$f"
    echo
  done
} >"$out"

echo "Wrote $out"
