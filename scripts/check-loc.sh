#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/sources.sh"

MAXIMUM=400
REVIEW=300

violations=0
review_count=0
largest_file=""
largest_lines=0

while IFS= read -r file; do
  [[ -z "${file}" ]] && continue
  nuvro_is_generated "${file}" && continue

  lines=$(wc -l < "${file}" | tr -d ' ')
  relative=$(nuvro_relative_path "${file}")

  if (( lines > largest_lines )); then
    largest_lines=${lines}
    largest_file="${relative}"
  fi

  if (( lines > MAXIMUM )); then
    printf 'FAIL %5d %s\n' "${lines}" "${relative}"
    violations=$(( violations + 1 ))
  elif (( lines > REVIEW )); then
    printf 'REVIEW %3d %s\n' "${lines}" "${relative}"
    review_count=$(( review_count + 1 ))
  fi
done < <(nuvro_source_files go swift kt kts sql sh yaml yml)

printf 'largest handwritten source file: %s (%d lines)\n' "${largest_file:-none}" "${largest_lines}"
printf 'files between %d and %d lines: %d\n' "$(( REVIEW + 1 ))" "${MAXIMUM}" "${review_count}"

if (( violations > 0 )); then
  printf 'check-loc: FAIL, %d file(s) exceed %d lines\n' "${violations}" "${MAXIMUM}" >&2
  exit 1
fi

printf 'check-loc: PASS, no file exceeds %d lines\n' "${MAXIMUM}"
