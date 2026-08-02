# Shared helpers for Apexio component test scripts.
# shellcheck shell=bash

ROOT="$(cd "$(dirname "${BASH_SOURCE[1]:-${BASH_SOURCE[0]}}")/.." && pwd)"
COMPOSE_FILE="${ROOT}/deploy/compose/docker-compose.yml"
COMPOSE=(docker compose -f "${COMPOSE_FILE}")

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

_APEXIO_CLEANUP_FUNCS=()
_APEXIO_CLEANUP_TRAP_SET=0

pass() { echo -e "${GREEN}PASS${NC}: $*"; }
fail() { echo -e "${RED}FAIL${NC}: $*"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $*"; }

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

_apexio_run_cleanup() {
  local fn
  for ((i = ${#_APEXIO_CLEANUP_FUNCS[@]} - 1; i >= 0; i--)); do
    "${_APEXIO_CLEANUP_FUNCS[i]}" || true
  done
}

# Register a teardown function; runs in reverse order on script exit (success or fail).
register_cleanup() {
  _APEXIO_CLEANUP_FUNCS+=("$1")
  if [[ "${_APEXIO_CLEANUP_TRAP_SET}" == "0" ]]; then
    trap _apexio_run_cleanup EXIT
    _APEXIO_CLEANUP_TRAP_SET=1
  fi
}

compose_teardown() {
  info "stopping compose stack and volumes"
  "${COMPOSE[@]}" down -v --remove-orphans 2>/dev/null || true
}

register_compose_cleanup() {
  register_cleanup compose_teardown
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

wait_clickhouse_message_compose() {
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
  fail "message not found in ClickHouse (compose): ${msg}"
}

wait_clickhouse_message_k8s() {
  local msg="$1"
  local attempts="${2:-60}"
  local i count
  require_cmd kubectl
  for ((i = 1; i <= attempts; i++)); do
    count="$(kubectl -n apexio exec clickhouse-0 -- clickhouse-client --query \
      "SELECT count() FROM apexio.logs WHERE message = '${msg}'" 2>/dev/null || echo 0)"
    if [[ "${count}" -ge 1 ]]; then
      return 0
    fi
    sleep 2
  done
  fail "message not found in ClickHouse (k8s): ${msg}"
}

# Backwards-compatible alias for compose tests.
wait_clickhouse_message() {
  wait_clickhouse_message_compose "$@"
}

ensure_compose_clickhouse_schema() {
  local table_exists
  table_exists="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM system.tables WHERE database = 'apexio' AND name = 'logs'" 2>/dev/null || echo 0)"
  if [[ "${table_exists}" == "1" ]]; then
    return 0
  fi
  info "apexio.logs missing; applying deploy/clickhouse/init/01_schema.sql"
  docker exec -i apexio-clickhouse clickhouse-client --multiquery \
    < "${ROOT}/deploy/clickhouse/init/01_schema.sql"
}

ensure_k8s_clickhouse_schema() {
  require_cmd kubectl
  local table_exists
  table_exists="$(kubectl -n apexio exec clickhouse-0 -- clickhouse-client --query \
    "SELECT count() FROM system.tables WHERE database = 'apexio' AND name = 'logs'" 2>/dev/null || echo 0)"
  if [[ "${table_exists}" == "1" ]]; then
    return 0
  fi
  info "apexio.logs missing in cluster; applying schema"
  kubectl -n apexio exec -i clickhouse-0 -- clickhouse-client --multiquery \
    < "${ROOT}/deploy/clickhouse/init/01_schema.sql"
}
