#!/usr/bin/env bash
# Grafana dashboard provisioning backed by ClickHouse.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

DASHBOARD_FILE="${ROOT}/deploy/grafana/provisioning/dashboards/json/apexio-logs.json"
GATEWAY="${GATEWAY_URL:-http://127.0.0.1:18080}"
GRAFANA="${GRAFANA_URL:-http://127.0.0.1:3000}"
GRAFANA_AUTH="admin:admin"
SEED_PREFIX="grafana-seed-$(date +%s)-$$"

test_dashboard_file() {
  [[ -f "${DASHBOARD_FILE}" ]] || fail "missing dashboard ${DASHBOARD_FILE}"
  python3 - <<'PY' "${DASHBOARD_FILE}"
import json, sys
path = sys.argv[1]
with open(path) as f:
    d = json.load(f)
assert d.get("uid") == "apexio-logs", d.get("uid")
assert d.get("title") == "Apexio Logs"
panels = [p for p in d.get("panels", []) if p.get("type") != "row"]
assert len(panels) >= 12, f"expected >=12 data panels, got {len(panels)}"
titles = {p["title"] for p in panels}
expected = {
    "Total Logs",
    "Errors",
    "Error Rate %",
    "p95 Latency",
    "Active Services",
    "Log Volume by Level",
    "Error Count Over Time",
    "Top Services by Volume",
    "HTTP Status (when present)",
    "Latency Percentiles (p50 / p95 / p99)",
    "Slowest Paths",
    "Recent Errors",
    "Log Viewer",
}
missing = expected - titles
assert not missing, f"missing panels: {missing}"
for p in panels:
    uid = p.get("datasource", {}).get("uid")
    assert uid == "apexio_clickhouse", p.get("title")
    sql = p["targets"][0].get("rawSql", "")
    assert "apexio.logs" in sql, p.get("title")
vars_ = {v["name"] for v in d.get("templating", {}).get("list", [])}
for name in ("service", "environment", "log_level", "host", "search"):
    assert name in vars_, f"missing variable {name}"
assert d.get("time", {}).get("from") == "now-1h", d.get("time")
viewer = next(p for p in panels if p["title"] == "Log Viewer")
assert "formatDateTime" in viewer["targets"][0]["rawSql"], "Log Viewer should stringify timestamps"
assert "match(service" in viewer["targets"][0]["rawSql"], "Log Viewer should use variable match filters"
print("ok")
PY
  pass "dashboard JSON valid (panels, variables, ClickHouse datasource)"
}

post_log() {
  local message="$1"
  local level="$2"
  local status="$3"
  local duration="$4"
  local service="${5:-grafana-dashboard}"
  local path="${6:-/grafana-test}"
  local id="${RANDOM}${RANDOM}"
  local code
  code="$(curl -s -o /tmp/apexio-grafana-post.json -w '%{http_code}' \
    -X POST "${GATEWAY}/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": ${id},
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"${level}\",
      \"message\": \"${message}\",
      \"metadata\": {
        \"requestId\": \"req-${id}\",
        \"requestMethod\": \"GET\",
        \"requestPath\": \"${path}\",
        \"responseStatus\": ${status},
        \"responseDuration\": ${duration}
      },
      \"source\": {
        \"service\": \"${service}\",
        \"host\": \"localhost\",
        \"environment\": \"dev\"
      }
    }")"
  [[ "${code}" == "201" ]] || fail "seed ingest failed (${code}): $(cat /tmp/apexio-grafana-post.json)"
}

seed_clickhouse_data() {
  info "seeding sample logs for dashboard panels"
  post_log "${SEED_PREFIX}-info-200" "INFO" 200 45.5 "api-gateway" "/health"
  post_log "${SEED_PREFIX}-info-201" "INFO" 201 120.0 "user-service" "/api/users"
  post_log "${SEED_PREFIX}-warn-404" "WARN" 404 80.2 "api-gateway" "/api/missing"
  post_log "${SEED_PREFIX}-error-500" "ERROR" 500 250.7 "user-service" "/api/users"
  post_log "${SEED_PREFIX}-fatal-502" "FATAL" 502 900.1 "payments" "/api/charge"
  pass "seeded 5 logs via gateway"
}

wait_seed_in_clickhouse() {
  local i count
  for ((i = 1; i <= 45; i++)); do
    count="$(docker exec apexio-clickhouse clickhouse-client --query \
      "SELECT count() FROM apexio.logs WHERE message LIKE '${SEED_PREFIX}%'" 2>/dev/null || echo 0)"
    if [[ "${count}" -ge 5 ]]; then
      return 0
    fi
    sleep 2
  done
  fail "seed logs not visible in ClickHouse"
}

test_clickhouse_panel_queries() {
  local errors volume statuses services
  errors="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM apexio.logs WHERE message LIKE '${SEED_PREFIX}%' AND log_level IN ('ERROR','FATAL')")"
  [[ "${errors}" -ge 2 ]] || fail "error panel query expected >=2, got ${errors}"

  volume="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM apexio.logs WHERE message LIKE '${SEED_PREFIX}%'")"
  [[ "${volume}" -ge 5 ]] || fail "volume query expected >=5, got ${volume}"

  statuses="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count(DISTINCT response_status) FROM apexio.logs WHERE message LIKE '${SEED_PREFIX}%' AND response_status > 0")"
  [[ "${statuses}" -ge 3 ]] || fail "status distribution expected >=3 codes, got ${statuses}"

  services="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT uniqExact(service) FROM apexio.logs WHERE message LIKE '${SEED_PREFIX}%'")"
  [[ "${services}" -ge 2 ]] || fail "expected >=2 services, got ${services}"

  pass "ClickHouse panel queries return expected seed data"
}

test_grafana_provisioning() {
  wait_http_ok "${GRAFANA}/api/health" 90
  # Plugin install + file provisioning can take a short while after container start.
  local i dash=""
  for ((i = 1; i <= 45; i++)); do
    dash="$(curl -sf -u "${GRAFANA_AUTH}" "${GRAFANA}/api/dashboards/uid/apexio-logs" 2>/dev/null || true)"
    if [[ -n "${dash}" ]] && echo "${dash}" | grep -q 'Apexio Logs'; then
      break
    fi
    sleep 2
  done
  local ds
  ds="$(curl -sf -u "${GRAFANA_AUTH}" "${GRAFANA}/api/datasources/uid/apexio_clickhouse")" \
    || fail "ClickHouse datasource not provisioned"
  echo "${ds}" | grep -q 'grafana-clickhouse-datasource' \
    || fail "unexpected datasource: ${ds}"
  pass "ClickHouse datasource provisioned"

  [[ -n "${dash}" ]] || dash="$(curl -sf -u "${GRAFANA_AUTH}" "${GRAFANA}/api/dashboards/uid/apexio-logs" 2>/dev/null || true)"
  [[ -n "${dash}" ]] || fail "dashboard apexio-logs not provisioned"
  echo "${dash}" | grep -q 'Apexio Logs' || fail "dashboard title missing"
  echo "${dash}" | grep -q 'Log Viewer' || fail "Log Viewer panel missing in provisioned dashboard"
  echo "${dash}" | grep -q 'Error Rate %' || fail "Error Rate % panel missing"
  echo "${dash}" | grep -q 'apexio_clickhouse' || fail "dashboard not wired to ClickHouse datasource"
  echo "${dash}" | grep -q '"name":"service"' || fail "service variable missing"
  echo "${dash}" | grep -q '"name":"search"' || fail "search variable missing"
  pass "Grafana dashboard provisioned (uid=apexio-logs)"
}

test_e2e() {
  require_cmd docker
  require_cmd curl
  require_cmd python3
  register_compose_cleanup

  info "starting stack"
  "${COMPOSE[@]}" up -d --build
  # Ensure Grafana picks up dashboard JSON from the mounted provisioning folder.
  "${COMPOSE[@]}" up -d --force-recreate grafana
  wait_http_ok "${GATEWAY}/healthz" 90
  wait_http_ok "http://127.0.0.1:8081/healthz" 90
  pass "stack healthy"

  seed_clickhouse_data
  wait_seed_in_clickhouse
  test_clickhouse_panel_queries
  test_grafana_provisioning
}

main() {
  test_dashboard_file
  test_e2e
  echo
  pass "Grafana tests passed"
  info "Open ${GRAFANA}/d/apexio-logs/apexio-logs (admin/admin)"
}

main "$@"
