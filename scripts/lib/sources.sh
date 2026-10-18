#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

EXCLUDED_PATH_PATTERN='/(\.git|\.build|build|DerivedData|\.gradle|vendor|node_modules|Pods|third-party|third_party|Generated|generated|sqlc)/'

nuvro_source_files() {
  local extensions=("$@")
  local pattern=""
  local extension

  for extension in "${extensions[@]}"; do
    if [[ -n "${pattern}" ]]; then
      pattern="${pattern}|"
    fi
    pattern="${pattern}${extension}"
  done

  find "${REPO_ROOT}" -type f \
    | grep -E "\.(${pattern})$" \
    | grep -Ev "${EXCLUDED_PATH_PATTERN}" \
    | grep -Ev '/(go\.sum|Package\.resolved|gradlew)$' \
    | LC_ALL=C sort
}

nuvro_relative_path() {
  printf '%s\n' "${1#"${REPO_ROOT}/"}"
}

nuvro_is_generated() {
  head -n 5 "$1" 2>/dev/null | grep -qE '^(//|--|#) Code generated .* DO NOT EDIT\.$'
}
