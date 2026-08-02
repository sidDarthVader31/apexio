#!/usr/bin/env bash
# Auth middleware, writer metrics, broker docs, and API-key E2E (Docker Compose).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

GATEWAY="${GATEWAY_URL:-http://127.0.0.1:18080}"
WRITER_METRICS="${WRITER_METRICS_URL:-http://127.0.0.1:8081/metrics}"
API_KEY="${GATEWAY_API_KEY:-auth-test-key}"
API_HEADER="${GATEWAY_API_KEY_HEADER:-X-API-Key}"
REST_MSG="auth-smoke-$(date +%s)-$$"

test_unit() {
  require_cmd go
  go test ./pkg/auth/... ./cmd/writer/... ./cmd/gateway/...
  pass "auth and service unit tests"
}

test_env_example() {
  [[ -f "${ROOT}/.env.example" ]] || fail "missing .env.example"
  grep -q 'GATEWAY_API_KEY' "${ROOT}/.env.example" || fail ".env.example missing GATEWAY_API_KEY"
  grep -q 'WRITER_BATCH_SIZE' "${ROOT}/.env.example" || fail ".env.example missing WRITER_BATCH_SIZE"
  pass ".env.example present with gateway/writer settings"
}

test_broker_docs() {
  local readme="${ROOT}/pkg/broker/README.md"
  [[ -f "${readme}" ]] || fail "missing ${readme}"
  grep -q 'Delivery policy' "${readme}" || fail "broker README missing delivery policy"
  grep -q 'Bring your own broker' "${readme}" || fail "broker README missing BYO broker section"
  pass "broker delivery policy documented"
}

post_log() {
  local with_key="$1"
  local id="${RANDOM}${RANDOM}"
  if [[ "${with_key}" == "yes" ]]; then
    curl -s -o /tmp/apexio-auth-post.json -w '%{http_code}' \
      -X POST "${GATEWAY}/api/v1/log" \
      -H 'Content-Type: application/json' \
      -H "${API_HEADER}: ${API_KEY}" \
      -d "{
      \"id\": ${id},
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"INFO\",
      \"message\": \"${REST_MSG}\",
      \"metadata\": {
        \"requestMethod\": \"GET\",
        \"requestPath\": \"/auth-test\",
        \"responseStatus\": 200,
        \"responseDuration\": 12.5
      },
      \"source\": {
        \"service\": \"auth-test\",
        \"host\": \"localhost\",
        \"environment\": \"dev\"
      }
    }"
    return
  fi
  curl -s -o /tmp/apexio-auth-post.json -w '%{http_code}' \
    -X POST "${GATEWAY}/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": ${id},
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"INFO\",
      \"message\": \"${REST_MSG}\",
      \"metadata\": {
        \"requestMethod\": \"GET\",
        \"requestPath\": \"/auth-test\",
        \"responseStatus\": 200,
        \"responseDuration\": 12.5
      },
      \"source\": {
        \"service\": \"auth-test\",
        \"host\": \"localhost\",
        \"environment\": \"dev\"
      }
    }"
}

test_api_key_auth() {
  local code
  code="$(post_log no)"
  [[ "${code}" == "401" ]] || fail "expected 401 without API key, got ${code}: $(cat /tmp/apexio-auth-post.json)"
  pass "ingest rejected without API key (401)"

  code="$(post_log yes)"
  [[ "${code}" == "201" ]] || fail "expected 201 with API key, got ${code}: $(cat /tmp/apexio-auth-post.json)"
  pass "ingest accepted with API key (201)"
}

wait_writer_metrics() {
  local i body
  for ((i = 1; i <= 45; i++)); do
    body="$(curl -sf "${WRITER_METRICS}" 2>/dev/null || true)"
    if echo "${body}" | grep -q 'apexio_writer_events_written_total' && \
       echo "${body}" | grep -q 'apexio_writer_batches_flushed_total'; then
      if echo "${body}" | awk '/apexio_writer_events_written_total / { if ($2+0 > 0) found=1 } END { exit !found }'; then
        return 0
      fi
    fi
    sleep 2
  done
  fail "writer /metrics did not show events written"
}

test_e2e() {
  require_cmd docker
  require_cmd curl
  register_compose_cleanup

  info "starting stack with GATEWAY_API_KEY enabled"
  GATEWAY_API_KEY="${API_KEY}" GATEWAY_API_KEY_HEADER="${API_HEADER}" \
    "${COMPOSE[@]}" up -d --build
  wait_http_ok "${GATEWAY}/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  pass "stack healthy with API key auth enabled"

  test_api_key_auth
  wait_writer_metrics
  pass "writer exposes pipeline self-metrics"
}

main() {
  test_unit
  test_env_example
  test_broker_docs
  test_e2e
  echo
  pass "auth tests passed"
  info "Copy .env.example to .env and set GATEWAY_API_KEY to enable auth locally"
}

main "$@"
