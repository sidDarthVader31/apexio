#!/usr/bin/env bash
# Phase 6 tests: auth middleware, writer batching, broker docs, API key E2E.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

COMPOSE_FILE="${ROOT}/deploy/compose/docker-compose.yml"
COMPOSE=(docker compose -f "${COMPOSE_FILE}")
GATEWAY="${GATEWAY_URL:-http://127.0.0.1:18080}"
WRITER_METRICS="${WRITER_METRICS_URL:-http://127.0.0.1:8081/metrics}"
API_KEY="${GATEWAY_API_KEY:-phase6-test-key}"
API_HEADER="${GATEWAY_API_KEY_HEADER:-X-API-Key}"

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

pass() { echo -e "${GREEN}PASS${NC}: $*"; }
fail() { echo -e "${RED}FAIL${NC}: $*"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $*"; }

REST_MSG="phase6-auth-$(date +%s)-$$"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "required command not found: $1"
}

test_unit() {
  require_cmd go
  go test ./pkg/auth/... ./cmd/writer/... ./cmd/gateway/...
  pass "unit tests (auth, writer batching, gateway)"
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

post_log() {
  local with_key="$1"
  local id="${RANDOM}${RANDOM}"
  if [[ "${with_key}" == "yes" ]]; then
    curl -s -o /tmp/apexio-phase6-post.json -w '%{http_code}' \
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
        \"requestPath\": \"/phase6\",
        \"responseStatus\": 200,
        \"responseDuration\": 12.5
      },
      \"source\": {
        \"service\": \"phase6-hardening\",
        \"host\": \"localhost\",
        \"environment\": \"dev\"
      }
    }"
    return
  fi
  curl -s -o /tmp/apexio-phase6-post.json -w '%{http_code}' \
    -X POST "${GATEWAY}/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": ${id},
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"INFO\",
      \"message\": \"${REST_MSG}\",
      \"metadata\": {
        \"requestMethod\": \"GET\",
        \"requestPath\": \"/phase6\",
        \"responseStatus\": 200,
        \"responseDuration\": 12.5
      },
      \"source\": {
        \"service\": \"phase6-hardening\",
        \"host\": \"localhost\",
        \"environment\": \"dev\"
      }
    }"
}

test_api_key_auth() {
  local code
  code="$(post_log no)"
  [[ "${code}" == "401" ]] || fail "expected 401 without API key, got ${code}: $(cat /tmp/apexio-phase6-post.json)"
  pass "ingest rejected without API key (401)"

  code="$(post_log yes)"
  [[ "${code}" == "201" ]] || fail "expected 201 with API key, got ${code}: $(cat /tmp/apexio-phase6-post.json)"
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
  pass "Phase 6 hardening tests all passed"
  info "Copy .env.example to .env and set GATEWAY_API_KEY to enable auth locally"
}

main "$@"
