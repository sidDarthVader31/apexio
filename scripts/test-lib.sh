# Shared helpers for Apexio component test scripts.
# shellcheck shell=bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}")/.." && pwd)"
COMPOSE_FILE="${ROOT}/deploy/compose/docker-compose.yml"
COMPOSE=(docker compose -f "${COMPOSE_FILE}")

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

pass() { echo -e "${GREEN}PASS${NC}: $*"; }
fail() { echo -e "${RED}FAIL${NC}: $*"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $*"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

wait_http_ok() {
  local url="$1"
  local attempts="${2:-60}"
  local i
  for ((i = 1; i <= attempts; i++)); do
    if curl -sf "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 2
  done
  fail "timed out waiting for ${url}"
}

wait_clickhouse_message() {
  local msg="$1"
  local attempts="${2:-45}"
  local i count
  for ((i = 1; i <= attempts; i++)); do
    count="$(docker exec apexio-clickhouse clickhouse-client --query \
      "SELECT count() FROM apexio.logs WHERE message = '${msg}'" 2>/dev/null || echo 0)"
    if [[ "${count}" -ge 1 ]]; then
      return 0
    fi
    sleep 2
  done
  fail "message not found in ClickHouse: ${msg}"
}
