#!/usr/bin/env bash
# Phase 3 vertical-slice tests: unit + E2E (REST → Redpanda → ClickHouse).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

COMPOSE_FILE="${ROOT}/deploy/compose/docker-compose.yml"
COMPOSE=(docker compose -f "${COMPOSE_FILE}")

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

pass() { echo -e "${GREEN}PASS${NC}: $*"; }
fail() { echo -e "${RED}FAIL${NC}: $*"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $*"; }

SMOKE_MSG="phase3-smoke-$(date +%s)-$$"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

test_unit() {
  info "running unit tests"
  go test ./pkg/... ./cmd/... -count=1 -timeout 120s
  pass "unit tests"
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

test_e2e() {
  require_cmd docker
  require_cmd curl

  info "building and starting compose stack"
  "${COMPOSE[@]}" up -d --build

  wait_http_ok "http://127.0.0.1:18080/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  pass "gateway and writer healthy"

  # Invalid payload must NOT return 201
  local bad_code
  bad_code="$(curl -s -o /tmp/apexio-bad.json -w '%{http_code}' \
    -X POST "http://127.0.0.1:18080/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d '{"id":1,"timestamp":1,"logLevel":"INFO","message":"x","source":{"service":""}}')"
  [[ "${bad_code}" == "400" ]] || fail "expected 400 for invalid ingest, got ${bad_code} body=$(cat /tmp/apexio-bad.json)"
  pass "invalid ingest returns 400 (not 201)"

  local ts
  ts="$(date +%s)000"
  local payload
  payload="$(cat <<EOF
{
  "id": 3003,
  "timestamp": ${ts},
  "logLevel": "INFO",
  "message": "${SMOKE_MSG}",
  "metadata": {
    "requestId": "phase3-req",
    "clientIp": "127.0.0.1",
    "userAgent": "phase3-test",
    "requestMethod": "POST",
    "requestPath": "/api/v1/log",
    "responseStatus": 201,
    "responseDuration": 12.5,
    "extra": {"traceId": "phase3"}
  },
  "source": {
    "host": "localhost",
    "service": "phase3-test",
    "environment": "dev"
  }
}
EOF
)"

  local code
  code="$(curl -s -o /tmp/apexio-ok.json -w '%{http_code}' \
    -X POST "http://127.0.0.1:18080/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "${payload}")"
  [[ "${code}" == "201" ]] || fail "expected 201, got ${code} body=$(cat /tmp/apexio-ok.json)"
  pass "ingest returned 201"

  info "waiting for row in ClickHouse"
  local found=0
  local i count
  for ((i = 1; i <= 45; i++)); do
    count="$(docker exec apexio-clickhouse clickhouse-client --query \
      "SELECT count() FROM apexio.logs WHERE message = '${SMOKE_MSG}'" 2>/dev/null || echo 0)"
    if [[ "${count}" -ge 1 ]]; then
      found=1
      break
    fi
    sleep 2
  done
  [[ "${found}" == "1" ]] || fail "message never appeared in ClickHouse: ${SMOKE_MSG}"
  pass "log visible in ClickHouse"

  # Named volumes survive a bounce (stop/start, not recreate -v)
  info "bouncing stack to verify volume durability"
  "${COMPOSE[@]}" stop gateway writer redpanda clickhouse
  "${COMPOSE[@]}" start clickhouse redpanda
  # wait for clickhouse before writer
  local i
  for ((i = 1; i <= 60; i++)); do
    if docker exec apexio-clickhouse clickhouse-client --query "SELECT 1" >/dev/null 2>&1; then
      break
    fi
    sleep 2
  done
  "${COMPOSE[@]}" start gateway writer
  wait_http_ok "http://127.0.0.1:18080/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  count="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM apexio.logs WHERE message = '${SMOKE_MSG}'")"
  [[ "${count}" -ge 1 ]] || fail "ClickHouse data lost after bounce"
  pass "ClickHouse data survived bounce (named volumes)"
}

main() {
  require_cmd go
  test_unit
  test_e2e
  echo
  pass "Phase 3 vertical-slice tests all passed"
}

main "$@"
