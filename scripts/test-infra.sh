#!/usr/bin/env bash
# Compose infrastructure: Redpanda, ClickHouse schema, Grafana datasource.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"

wait_for_healthy() {
  local service="$1"
  local attempts="${2:-60}"
  local i
  for ((i = 1; i <= attempts; i++)); do
    local status
    status="$("${COMPOSE[@]}" ps --format json "${service}" 2>/dev/null | head -n1 || true)"
    if [[ -n "${status}" ]] && echo "${status}" | grep -q '"Health":"healthy"'; then
      return 0
    fi
    # Fallback: older compose without Health field — check running + probe
    if [[ "${service}" == "clickhouse" ]]; then
      if docker exec apexio-clickhouse clickhouse-client --query "SELECT 1" >/dev/null 2>&1; then
        return 0
      fi
    fi
    if [[ "${service}" == "grafana" ]]; then
      if curl -sf "http://127.0.0.1:3000/api/health" >/dev/null 2>&1; then
        return 0
      fi
    fi
    if [[ "${service}" == "redpanda" ]]; then
      if docker exec apexio-redpanda rpk cluster health 2>/dev/null | grep -qi 'Healthy.*true\|clean'; then
        return 0
      fi
    fi
    sleep 2
  done
  fail "${service} did not become healthy within $((attempts * 2))s"
}

test_compose_file() {
  [[ -f "${COMPOSE_FILE}" ]] || fail "missing ${COMPOSE_FILE}"
  "${COMPOSE[@]}" config -q
  pass "compose file validates"
}

test_stack_up() {
  info "starting compose stack (if not already running)"
  "${COMPOSE[@]}" up -d
  wait_for_healthy redpanda 60
  wait_for_healthy clickhouse 60
  wait_for_healthy grafana 90
  pass "all services healthy"
}

test_redpanda() {
  local out
  out="$(docker exec apexio-redpanda rpk cluster health 2>&1)" || fail "rpk cluster health failed: ${out}"
  echo "${out}" | grep -qiE 'Healthy[[:space:]]*:[[:space:]]*true|Healthy.*true' \
    || fail "Redpanda cluster not healthy: ${out}"
  # Kafka API reachable on host-advertised port
  if command -v nc >/dev/null 2>&1; then
    nc -z 127.0.0.1 19092 || fail "Redpanda Kafka port 19092 not accepting connections"
  fi
  pass "Redpanda healthy and Kafka port open"
}

apply_clickhouse_schema_if_needed() {
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

test_clickhouse_schema() {
  local db_exists table_exists cols
  apply_clickhouse_schema_if_needed

  db_exists="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM system.databases WHERE name = 'apexio'")"
  [[ "${db_exists}" == "1" ]] || fail "database apexio missing (got ${db_exists})"

  table_exists="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM system.tables WHERE database = 'apexio' AND name = 'logs'")"
  [[ "${table_exists}" == "1" ]] || fail "table apexio.logs missing (got ${table_exists})"

  cols="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT name FROM system.columns WHERE database = 'apexio' AND table = 'logs' ORDER BY name")"

  local required=(
    attrs
    client_ip
    environment
    host
    id
    ingested_at
    log_level
    message
    request_id
    request_method
    request_path
    response_duration_ms
    response_status
    service
    timestamp
    user_agent
  )
  local c
  for c in "${required[@]}"; do
    echo "${cols}" | grep -qx "${c}" || fail "column missing on apexio.logs: ${c}"
  done

  # Round-trip insert/select to prove table is writable
  docker exec apexio-clickhouse clickhouse-client --query "
    INSERT INTO apexio.logs
      (timestamp, id, log_level, message, service, host, environment,
       request_id, client_ip, user_agent, request_method, request_path,
       response_status, response_duration_ms, attrs)
    VALUES
      (now64(3), 1, 'INFO', 'infra-smoke', 'apexio-test', 'localhost', 'dev',
       'req-1', '127.0.0.1', 'smoke', 'GET', '/health',
       200, 1.5, map('traceId', 'infra'))
  "
  local count
  count="$(docker exec apexio-clickhouse clickhouse-client --query \
    "SELECT count() FROM apexio.logs WHERE message = 'infra-smoke'")"
  [[ "${count}" -ge 1 ]] || fail "smoke insert not readable (count=${count})"

  pass "ClickHouse schema present and writable"
}

test_grafana() {
  local health datasources ds_name ds_type
  health="$(curl -sf "http://127.0.0.1:3000/api/health")" \
    || fail "Grafana /api/health not reachable"
  echo "${health}" | grep -q 'ok\|database' || fail "unexpected Grafana health: ${health}"

  # Default admin:admin from compose
  datasources="$(curl -sf -u admin:admin "http://127.0.0.1:3000/api/datasources")" \
    || fail "could not list Grafana datasources (plugin may still be installing)"

  echo "${datasources}" | grep -q 'apexio_clickhouse' \
    || fail "provisioned datasource uid apexio_clickhouse not found: ${datasources}"

  ds_type="$(curl -sf -u admin:admin "http://127.0.0.1:3000/api/datasources/uid/apexio_clickhouse" \
    | sed -n 's/.*"type":"\([^"]*\)".*/\1/p' | head -n1)"
  [[ "${ds_type}" == "grafana-clickhouse-datasource" ]] \
    || fail "datasource type expected grafana-clickhouse-datasource, got '${ds_type}'"

  pass "Grafana healthy with ClickHouse datasource provisioned"
}

main() {
  require_cmd docker
  require_cmd curl
  register_compose_cleanup
  info "infra tests (root=${ROOT})"
  test_compose_file
  test_stack_up
  test_redpanda
  test_clickhouse_schema
  test_grafana
  echo
  pass "infra tests passed"
}

main "$@"
