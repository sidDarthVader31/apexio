#!/usr/bin/env bash
# Phase 4 tests: OTLP ingest + sample client (unit + E2E).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

COMPOSE_FILE="${ROOT}/deploy/compose/docker-compose.yml"
COMPOSE=(docker compose -f "${COMPOSE_FILE}")
GATEWAY="${GATEWAY_URL:-http://127.0.0.1:18080}"

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

pass() { echo -e "${GREEN}PASS${NC}: $*"; }
fail() { echo -e "${RED}FAIL${NC}: $*"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $*"; }

REST_MSG="phase4-rest-$(date +%s)-$$"
OTLP_MSG="phase4-otlp-$(date +%s)-$$"

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

test_unit() {
  info "running unit tests"
  go test ./pkg/... ./cmd/gateway/... ./examples/sample-client/... -count=1 -timeout 120s
  pass "unit tests"
}

test_e2e() {
  require_cmd docker
  require_cmd curl

  info "building and starting compose stack"
  "${COMPOSE[@]}" up -d --build
  wait_http_ok "${GATEWAY}/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  pass "stack healthy"

  info "REST ingest"
  local code
  code="$(curl -s -o /tmp/apexio-phase4-rest.json -w '%{http_code}' \
    -X POST "${GATEWAY}/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": 41,
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"INFO\",
      \"message\": \"${REST_MSG}\",
      \"metadata\": {\"requestMethod\": \"GET\", \"requestPath\": \"/\", \"responseStatus\": 200, \"responseDuration\": 1},
      \"source\": {\"service\": \"phase4-rest\", \"host\": \"localhost\", \"environment\": \"dev\"}
    }")"
  [[ "${code}" == "201" ]] || fail "REST expected 201, got ${code}"
  wait_clickhouse_message "${REST_MSG}"
  pass "REST log in ClickHouse"

  info "OTLP ingest via sample client"
  go run ./examples/sample-client -gateway "${GATEWAY}" -mode otlp -message "${OTLP_MSG}" -service phase4-otlp
  wait_clickhouse_message "${OTLP_MSG}"
  pass "OTLP log in ClickHouse"

  info "metrics endpoint"
  curl -sf "${GATEWAY}/metrics" | grep -q 'apexio_ingest_requests_total' \
    || fail "metrics missing ingest counters"
  pass "gateway metrics exposed"
}

main() {
  require_cmd go
  test_unit
  test_e2e
  echo
  pass "Phase 4 OTLP tests all passed"
}

main "$@"
