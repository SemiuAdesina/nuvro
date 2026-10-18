#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${root}/backend"
[[ "$("${SQLC:-sqlc}" version)" == "v1.31.1" ]] || { echo "sqlc v1.31.1 required"; exit 1; }
if [[ "${1:-}" != check ]]; then
  exec "${SQLC:-sqlc}" generate
fi
work="$(mktemp -d)"
trap 'rm -rf "${work}"' EXIT
cp sqlc.yaml "${work}/sqlc.yaml"
cp -R db "${work}/db"
"${SQLC:-sqlc}" generate -f "${work}/sqlc.yaml"
diff -ru internal/adapter/postgres/sqlc "${work}/internal/adapter/postgres/sqlc"
