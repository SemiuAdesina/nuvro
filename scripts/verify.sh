#!/usr/bin/env bash
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
BACKEND="${REPO_ROOT}/backend"
IOS_PACKAGE="${REPO_ROOT}/clients/ios/Packages/NuvroKit"
ANDROID="${REPO_ROOT}/clients/android"

failures=0
skipped=0

run_gate() {
  local name="$1"
  shift
  printf '\n=== %s ===\n' "${name}"
  if "$@"; then
    printf 'PASS %s\n' "${name}"
  else
    printf 'FAIL %s\n' "${name}" >&2
    failures=$(( failures + 1 ))
  fi
}

skip_gate() {
  printf '\n=== %s ===\nSKIPPED %s: %s\n' "$1" "$1" "$2"
  skipped=$(( skipped + 1 ))
}

run_gate 'loc' "${SCRIPT_DIR}/check-loc.sh"
run_gate 'comments' "${SCRIPT_DIR}/check-comments.sh"
run_gate 'go-deps' "${SCRIPT_DIR}/check-go-deps.sh"

if command -v go >/dev/null 2>&1; then
  run_gate 'api-validate' make -C "${REPO_ROOT}" api-validate
  run_gate 'api-check' make -C "${REPO_ROOT}" api-check
  if command -v "${SQLC:-sqlc}" >/dev/null 2>&1; then
    run_gate 'sql-check' make -C "${REPO_ROOT}" sql-check
  else
    skip_gate 'sql-check' 'sqlc v1.31.1 is not installed'
  fi
  run_gate 'gofmt' bash -c "cd '${BACKEND}' && unformatted=\$(gofmt -l .) && [ -z \"\${unformatted}\" ] || { echo \"\${unformatted}\"; exit 1; }"
  run_gate 'go-vet' bash -c "cd '${BACKEND}' && go vet ./..."
  run_gate 'go-test' bash -c "cd '${BACKEND}' && go test -timeout 120s ./..."
  run_gate 'go-test-race' bash -c "cd '${BACKEND}' && go test -race -timeout 300s ./..."
else
  skip_gate 'gofmt' 'the Go toolchain is not installed'
  skip_gate 'go-vet' 'the Go toolchain is not installed'
  skip_gate 'go-test' 'the Go toolchain is not installed'
  skip_gate 'go-test-race' 'the Go toolchain is not installed'
fi

if command -v golangci-lint >/dev/null 2>&1; then
  run_gate 'golangci-lint' bash -c "cd '${BACKEND}' && golangci-lint run"
else
  skip_gate 'golangci-lint' 'golangci-lint is not installed'
fi

if command -v govulncheck >/dev/null 2>&1; then
  run_gate 'govulncheck' bash -c "cd '${BACKEND}' && govulncheck ./..."
else
  skip_gate 'govulncheck' 'govulncheck is not installed'
fi

run_gate 'ios-contract' "${REPO_ROOT}/scripts/ios-contract-sync.sh" check
if [[ -f "${IOS_PACKAGE}/Package.swift" ]] && command -v swift >/dev/null 2>&1; then
  run_gate 'swift-test' bash -c "cd '${IOS_PACKAGE}' && swift test"
elif [[ ! -f "${IOS_PACKAGE}/Package.swift" ]]; then
  skip_gate 'swift-test' 'the NuvroKit package does not exist yet'
else
  skip_gate 'swift-test' 'the Swift toolchain is not installed'
fi

if [[ -f "${ANDROID}/settings.gradle.kts" ]]; then
  run_gate 'android-test' bash -c "cd '${ANDROID}' && ./gradlew testDebugUnitTest"
else
  skip_gate 'android-test' 'the Android Gradle project is not initialised yet, see clients/android/README.md'
fi

printf '\n=== summary ===\n'
printf 'failed gates:  %d\n' "${failures}"
printf 'skipped gates: %d\n' "${skipped}"

if (( failures > 0 )); then
  printf 'verify: FAIL\n' >&2
  exit 1
fi

printf 'verify: PASS, with %d gate(s) skipped and reported above\n' "${skipped}"
