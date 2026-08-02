#!/usr/bin/env bash
# End-to-end pipeline smoke: REST ingest → Redpanda → ClickHouse (Docker Compose).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

SMOKE_MSG="pipeline-smoke-$(date +%s)-$$"

test_e2e() {
  require_cmd docker
  require_cmd curl
  register_compose_cleanup

  info "building and starting compose stack"
  "${COMPOSE[@]}" up -d --build

  wait_http_ok "http://127.0.0.1:18080/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  pass "gateway and writer healthy"

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
    "requestId": "pipeline-req",
    "clientIp": "127.0.0.1",
    "userAgent": "pipeline-test",
    "requestMethod": "POST",
    "requestPath": "/api/v1/log",
    "responseStatus": 201,
    "responseDuration": 12.5,
    "extra": {"traceId": "pipeline"}
  },
  "source": {
    "host": "localhost",
    "service": "pipeline-test",
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
  wait_clickhouse_message "${SMOKE_MSG}"
  pass "log visible in ClickHouse"

  info "bouncing stack to verify volume durability"
  "${COMPOSE[@]}" stop gateway writer redpanda clickhouse
  "${COMPOSE[@]}" start clickhouse redpanda
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
  local count
  count="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM apexio.logs WHERE message = '${SMOKE_MSG}'")"
  [[ "${count}" -ge 1 ]] || fail "ClickHouse data lost after bounce"
  pass "ClickHouse data survived bounce (named volumes)"
}

main() {
  test_e2e
  echo
  pass "pipeline tests passed"
}

main "$@"
