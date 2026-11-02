#!/usr/bin/env bash
# Ambiente descartavel, sem portas publicadas e sem credenciais de outros ambientes.
set -euo pipefail
cd "$(dirname "$0")/.."
export PERF_PROFILE=${1:-smoke}
case "$PERF_PROFILE" in smoke|load) ;; *) echo 'Uso: bash scripts/test-performance.sh [smoke|load]' >&2; exit 2 ;; esac
export PERF_DB_PASSWORD PERF_JWT_SECRET PERF_RESULTS_DIR PERF_UID PERF_GID
PERF_DB_PASSWORD=$(openssl rand -hex 24)
PERF_JWT_SECRET=$(openssl rand -hex 32)
mkdir -p reports/performance
PERF_RESULTS_DIR=$(cd reports/performance && pwd)
PERF_UID=$(id -u)
PERF_GID=$(id -g)
compose=(docker compose -p "oficina-perf-$(date +%s)-$$" -f tests/performance/compose.yaml)
cleanup() {
  local result=$?
  trap - EXIT
  for service in api postgres migrate; do
    "${compose[@]}" logs --no-color "$service" > "$PERF_RESULTS_DIR/$service.log" 2>&1 || true
  done
  "${compose[@]}" down --volumes --remove-orphans || result=1
  exit "$result"
}
# Remove resumos anteriores para nao publicar evidencias antigas em caso de erro no setup.
rm -f "$PERF_RESULTS_DIR/summary.json" "$PERF_RESULTS_DIR/summary.md"
trap cleanup EXIT
"${compose[@]}" up --build --wait --wait-timeout 180 api
"${compose[@]}" run --rm --no-deps k6
