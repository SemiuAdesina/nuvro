#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/sources.sh"

MODULE="github.com/SemiuAdesina/nuvro/backend"
INTERNAL="${REPO_ROOT}/backend/internal"

FORBIDDEN_PACKAGE_NAMES='^(utils|util|helpers|helper|common|misc|shared|stuff|base|general|manager|service|services)$'

violations=0

layer_of() {
  case "$1" in
    internal/domain/*)    printf 'domain' ;;
    internal/engine/*)    printf 'engine' ;;
    internal/usecase/*)   printf 'usecase' ;;
    internal/transport/*) printf 'transport' ;;
    internal/adapter/*)   printf 'adapter' ;;
    internal/worker/*)    printf 'worker' ;;
    internal/app|internal/app/*) printf 'app' ;;
    *)                    printf 'other' ;;
  esac
}

forbidden_targets_for() {
  case "$1" in
    domain)    printf 'adapter transport usecase engine app' ;;
    engine)    printf 'adapter transport usecase app' ;;
    usecase)   printf 'adapter transport app' ;;
    transport) printf 'adapter app' ;;
    adapter)   printf 'transport usecase app' ;;
    *)         printf '' ;;
  esac
}

imports_of() {
  awk '
    /^import \(/ { inside = 1; next }
    inside && /^\)/ { inside = 0; next }
    inside || /^import / {
      if (match($0, /"[^"]+"/)) {
        path = substr($0, RSTART + 1, RLENGTH - 2)
        print path
      }
    }
  ' "$1"
}

while IFS= read -r file; do
  [[ -z "${file}" ]] && continue
  case "${file}" in "${REPO_ROOT}/backend/"*) ;; *) continue ;; esac
  nuvro_is_generated "${file}" && continue

  relative=$(nuvro_relative_path "${file}")
  package_path="${relative#backend/}"
  layer=$(layer_of "${package_path}")
  forbidden=$(forbidden_targets_for "${layer}")
  [[ -z "${forbidden}" ]] && continue

  while IFS= read -r imported; do
    [[ "${imported}" == "${MODULE}/"* ]] || continue
    target_path="${imported#"${MODULE}"/}"
    target_layer=$(layer_of "${target_path}")
    case "${target_path}" in internal/app|internal/app/*) target_layer='app' ;; esac

    for banned in ${forbidden}; do
      if [[ "${target_layer}" == "${banned}" ]]; then
        printf 'FAIL %s\n\t%s imports %s: %s\n' "${relative}" "${layer}" "${banned}" "${imported}"
        violations=$(( violations + 1 ))
      fi
    done
  done < <(imports_of "${file}")
done < <(nuvro_source_files go)

while IFS= read -r directory; do
  [[ -z "${directory}" ]] && continue
  name=$(basename "${directory}")
  if [[ "${name}" =~ ${FORBIDDEN_PACKAGE_NAMES} ]]; then
    printf 'FAIL %s\n\tjunk drawer package name: %s\n' "$(nuvro_relative_path "${directory}")" "${name}"
    violations=$(( violations + 1 ))
  fi
done < <(find "${INTERNAL}" -type d 2>/dev/null | LC_ALL=C sort)

if (( violations > 0 )); then
  printf 'check-go-deps: FAIL, %d boundary violation(s)\n' "${violations}" >&2
  exit 1
fi

printf 'check-go-deps: PASS, dependency direction and package naming hold\n'
