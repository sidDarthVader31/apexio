#!/usr/bin/env bash
set -euo pipefail
GATEWAY="${GATEWAY_URL:-http://127.0.0.1:18080}"
PATH="/usr/local/bin:/usr/bin:/bin:/opt/homebrew/bin:$PATH"

seed() {
  local msg="$1" level="$2" http_status="$3" dur="$4" svc="$5" rpath="$6" host="${7:-api-1}" env="${8:-dev}"
  local id="${RANDOM}${RANDOM}"
  curl -sS -o /dev/null -w "%{http_code}\n" -X POST "${GATEWAY}/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": ${id},
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"${level}\",
      \"message\": \"${msg}\",
      \"metadata\": {
        \"requestId\": \"req-${id}\",
        \"clientIp\": \"10.0.0.$((RANDOM % 50))\",
        \"userAgent\": \"apexio-seed/1.0\",
        \"requestMethod\": \"GET\",
        \"requestPath\": \"${rpath}\",
        \"responseStatus\": ${http_status},
        \"responseDuration\": ${dur}
      },
      \"source\": {\"service\": \"${svc}\", \"host\": \"${host}\", \"environment\": \"${env}\"}
    }"
}

for i in $(seq 1 15); do
  seed "checkout completed order=${i}" INFO 200 $((20 + RANDOM % 80)) checkout /api/orders api-1 production
  seed "user profile loaded id=${i}" INFO 200 $((30 + RANDOM % 120)) user-service /api/users api-2 production
  seed "payment authorized txn=${i}" INFO 201 $((40 + RANDOM % 200)) payments /api/charge api-1 staging
done
for i in $(seq 1 5); do
  seed "slow catalog search q=${i}" WARN 200 $((400 + RANDOM % 400)) catalog /api/search api-2 production
  seed "cart item missing id=${i}" WARN 404 $((50 + RANDOM % 50)) checkout /api/cart api-1 production
done
for i in $(seq 1 4); do
  seed "upstream timeout on payments i=${i}" ERROR 500 $((200 + RANDOM % 300)) payments /api/charge api-1 production
  seed "db connection reset i=${i}" ERROR 503 $((100 + RANDOM % 150)) user-service /api/users api-2 staging
done
seed "circuit breaker open payments" FATAL 502 1200.5 payments /api/charge api-1 production
seed "panic recovered in gateway" FATAL 500 50.0 api-gateway /health api-1 production

echo "seeded; waiting for ClickHouse..."
for i in $(seq 1 45); do
  c="$(docker exec apexio-clickhouse clickhouse-client --query 'SELECT count() FROM apexio.logs' 2>/dev/null || echo 0)"
  echo "logs=${c}"
  if [[ "${c}" -ge 50 ]]; then
    break
  fi
  sleep 2
done

docker exec apexio-clickhouse clickhouse-client --query \
  "SELECT log_level, count() FROM apexio.logs GROUP BY log_level ORDER BY count() DESC"
