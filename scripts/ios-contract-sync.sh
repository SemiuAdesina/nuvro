#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
target="${root}/clients/ios/Packages/NuvroKit/Sources/NuvroAPI/openapi.json"
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
(cd "${root}/backend" && go run ./cmd/contract -output "${work}/openapi.json")
if [[ "${1:-}" == check ]]; then
  if [[ ! -f "${target}" ]]; then
    echo "ios-contract: ${target#"${root}/"} is missing; run scripts/ios-contract-sync.sh" >&2
    exit 1
  fi
  diff -u "${target}" "${work}/openapi.json"
else
  cp "${work}/openapi.json" "${target}"
fi
