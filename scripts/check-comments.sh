#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/sources.sh"

violations=0

inspect() {
  local file="$1"
  local relative="$2"

  awk -v path="${relative}" '
    function report(reason) {
      printf "FAIL %s:%d %s\n\t%s\n", path, FNR, reason, trimmed
      failures++
    }

    {
      line = $0
      trimmed = line
      sub(/^[ \t]+/, "", trimmed)
      sub(/[ \t]+$/, "", trimmed)

      if (block) {
        if (trimmed ~ /\*\//) { block = 0 }
        report("block comment")
        next
      }

      if (trimmed ~ /^\/\*/) {
        if (trimmed !~ /\*\//) { block = 1 }
        report("block comment")
        next
      }

      if (trimmed ~ /(^|[^A-Za-z])(TODO|FIXME|HACK|XXX)([^A-Za-z]|$)/ && trimmed ~ /\/\//) {
        report("forbidden marker")
        next
      }

      if (trimmed ~ /^\/\//) {
        if (trimmed ~ /^\/\/(go|line|export|extern|sys|nolint|swift-tools-version|lint):/) { next }
        if (trimmed ~ /^\/\/[ ]?Code generated .* DO NOT EDIT\.$/) { next }
        if (trimmed ~ /^\/\/[ ]?swift-tools-version/) { next }
        report("explanatory comment")
        next
      }

      stripped = line
      gsub(/"[^"]*"/, "\"\"", stripped)
      gsub(/`[^`]*`/, "``", stripped)
      gsub(/'"'"'[^'"'"']*'"'"'/, "", stripped)

      if (stripped ~ /\/\//) {
        report("trailing comment")
        next
      }
      if (stripped ~ /\/\*/) {
        report("inline block comment")
        next
      }
    }

    END { exit failures > 0 ? 1 : 0 }
  ' "${file}"
}

while IFS= read -r file; do
  [[ -z "${file}" ]] && continue
  nuvro_is_generated "${file}" && continue

  relative=$(nuvro_relative_path "${file}")
  if ! inspect "${file}" "${relative}"; then
    violations=$(( violations + 1 ))
  fi
done < <(nuvro_source_files go swift kt kts)

if (( violations > 0 )); then
  printf 'check-comments: FAIL, %d file(s) contain forbidden comments\n' "${violations}" >&2
  exit 1
fi

printf 'check-comments: PASS, no explanatory comments in handwritten runtime source\n'
