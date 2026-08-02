#!/usr/bin/env bash
# Gateway OTLP ingest and sample client (Docker Compose).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

GATEWAY="${GATEWAY_URL:-http://127.0.0.1:18080}"
REST_MSG="otlp-rest-$(date +%s)-$$"
OTLP_MSG="otlp-http-$(date +%s)-$$"

test_e2e() {
  require_cmd docker
  require_cmd curl
  require_cmd go

  info "building and starting compose stack"
  "${COMPOSE[@]}" up -d --build
  wait_http_ok "${GATEWAY}/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  pass "stack healthy"

  info "REST ingest"
  local code
  code="$(curl -s -o /tmp/apexio-otlp-rest.json -w '%{http_code}' \
    -X POST "${GATEWAY}/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": 41,
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"INFO\",
      \"message\": \"${REST_MSG}\",
      \"metadata\": {\"requestMethod\": \"GET\", \"requestPath\": \"/\", \"responseStatus\": 200, \"responseDuration\": 1},
      \"source\": {\"service\": \"otlp-rest\", \"host\": \"localhost\", \"environment\": \"dev\"}
    }")"
  [[ "${code}" == "201" ]] || fail "REST expected 201, got ${code}"
  wait_clickhouse_message "${REST_MSG}"
  pass "REST log in ClickHouse"

  info "OTLP ingest via sample client"
  go run ./examples/sample-client -gateway "${GATEWAY}" -mode otlp -message "${OTLP_MSG}" -service otlp-demo
  wait_clickhouse_message "${OTLP_MSG}"
  pass "OTLP log in ClickHouse"

  info "metrics endpoint"
  curl -sf "${GATEWAY}/metrics" | grep -q 'apexio_ingest_requests_total' \
    || fail "metrics missing ingest counters"
  pass "gateway metrics exposed"
}

main() {
  test_e2e
  echo
  pass "OTLP tests passed"
}

main "$@"
