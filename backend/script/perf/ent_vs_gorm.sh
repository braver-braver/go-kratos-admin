#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/../.. && pwd)"
SPECS_DIR="${ROOT}/specs/001-ent-to-gorm"
OUT="${SPECS_DIR}/contracts/perf_results.md"

DSN="${DSN:-host=timescaledb.pg15-timescale.orb.local port=5432 user=postgres password=Abcd1234 dbname=postgres sslmode=disable}"
SERVER_URL_ENT="${SERVER_URL_ENT:-http://localhost:8000}"
SERVER_URL_GORM="${SERVER_URL_GORM:-http://localhost:8001}"
TARGET_PATH="${TARGET_PATH:-/admin/v1/users?page=1&pageSize=20}"
CONCURRENCY="${CONCURRENCY:-10}"
REQUESTS="${REQUESTS:-100}"

run_case() {
  local name="$1" url="$2"
  local log="${OUT%.md}_${name}.log"
  echo "## ${name} @ ${url}" | tee -a "${OUT}"
  if command -v hey >/dev/null 2>&1; then
    hey -z 10s -c "${CONCURRENCY}" -q 1 -disable-keepalive "${url}${TARGET_PATH}" | tee "${log}"
  elif command -v ab >/dev/null 2>&1; then
    ab -n "${REQUESTS}" -c "${CONCURRENCY}" "${url}${TARGET_PATH}" | tee "${log}"
  else
    # Fallback: rough timing with curl
    for i in $(seq 1 5); do
      /usr/bin/time -p curl -s -o /dev/null "${url}${TARGET_PATH}"
    done | tee "${log}"
  fi
  echo "" | tee -a "${OUT}"
}

mkdir -p "$(dirname "${OUT}")"
{
  echo "# Perf Results: ent vs gorm"
  echo "- DSN: ${DSN}"
  echo "- Endpoint: ${TARGET_PATH}"
  echo "- Concurrency: ${CONCURRENCY}"
  echo "- Requests: ${REQUESTS}"
  echo "- Ent base URL: ${SERVER_URL_ENT}"
  echo "- Gorm base URL: ${SERVER_URL_GORM}"
  echo ""
} > "${OUT}"

# Run ent (baseline) then gorm (USE_GORM=true)
run_case "ent" "${SERVER_URL_ENT}"
run_case "gorm" "${SERVER_URL_GORM}"

echo "Perf results written to ${OUT}"
