#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}/backend"
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
go run ./cmd/contract -output "${work}/openapi.json"
go tool oapi-codegen -config oapi-codegen.yaml -o "${work}/server.gen.go" "${work}/openapi.json"
target=internal/transport/http/api/server.gen.go
if [[ "${1:-}" == check ]]; then
  diff -u "${target}" "${work}/server.gen.go"
else
  mkdir -p "$(dirname "${target}")"
  cp "${work}/server.gen.go" "${target}"
fi
