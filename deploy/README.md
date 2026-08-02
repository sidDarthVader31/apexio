# Apexio deploy

Self-hosted pipeline: **gateway** → **Redpanda** → **writer** → **ClickHouse**, plus **Grafana**.

## Quick start

From the repository root:

```bash
make up
# or
docker compose -f deploy/compose/docker-compose.yml up -d --build
```

Verify:

```bash
make test-phase1   # infra
make test-phase2   # contracts
make test-phase3   # REST vertical slice
make test-phase4   # OTLP + sample client
make test-phase5   # Grafana dashboards (ClickHouse)
make test-phase6   # Auth, writer batching, broker docs
```

### Grafana dashboards (Phase 5)

After `make up`, open [http://127.0.0.1:3000/d/apexio-logs/apexio-logs](http://127.0.0.1:3000/d/apexio-logs/apexio-logs) (`admin` / `admin`).

Panels: log volume, error rate, recent errors, response-time distribution, status-code distribution — all backed by `apexio.logs` in ClickHouse. Dashboards are provisioned from [`deploy/grafana/provisioning/dashboards/json/apexio-logs.json`](grafana/provisioning/dashboards/json/apexio-logs.json); no manual token or Job step.

### Ingest via REST

```bash
curl -sS -X POST http://127.0.0.1:18080/api/v1/log \
  -H 'Content-Type: application/json' \
  -d '{
    "id": 1,
    "timestamp": 1732974309000,
    "logLevel": "INFO",
    "message": "hello apexio",
    "metadata": {
      "requestId": "req-1",
      "requestMethod": "GET",
      "requestPath": "/health",
      "responseStatus": 200,
      "responseDuration": 12.5
    },
    "source": {
      "host": "localhost",
      "service": "demo",
      "environment": "dev"
    }
  }'
```

### Ingest via OTLP (HTTP)

Use the sample client:

```bash
go run ./examples/sample-client -mode otlp -message "hello otlp" -service demo
# or both REST + OTLP
go run ./examples/sample-client -mode both
```

OTLP HTTP endpoint: `POST /v1/logs` with `Content-Type: application/x-protobuf`.  
OTLP gRPC: port `4317` (standard OTLP logs export).

Gateway metrics: `GET /metrics`.  
Writer metrics: `GET http://127.0.0.1:8081/metrics` (batch flushes, events written, errors).

## Bring your own auth / broker

Apexio is a **clone-and-adapt backend** — auth and messaging are extension points, not a fixed product surface.

### Auth

Optional API-key middleware lives in [`pkg/auth`](../pkg/auth/). It is **disabled by default** (`GATEWAY_API_KEY` empty).

```bash
cp .env.example .env
# edit .env
export GATEWAY_API_KEY=your-secret
export GATEWAY_API_KEY_HEADER=X-API-Key   # optional, default shown
make up
```

Clients must send the header on ingest routes (`POST /api/v1/log`, `POST /v1/logs`). Health (`/healthz`) and metrics (`/metrics`) stay open.

Replace or chain middleware in [`cmd/gateway/main.go`](../cmd/gateway/main.go) for JWT, mTLS, or an org IdP — the `auth.Middleware` type is a thin `func(http.Handler) http.Handler` wrapper.

OTLP gRPC (`:4317`) is unauthenticated in the default build; terminate TLS and auth at a proxy or extend the gRPC server if you need it.

### Broker

Default: Redpanda via [`pkg/broker/redpanda.go`](../pkg/broker/redpanda.go). Delivery policy, acks, and backpressure are documented in [`pkg/broker/README.md`](../pkg/broker/README.md).

Implement `broker.Broker` (see `memory.go` for a test double) and wire it in `cmd/gateway` and `cmd/writer` to swap Kafka, NATS, RabbitMQ, etc.

### Writer tuning

Batch insert settings (env or `.env`):

| Variable | Default | Purpose |
|----------|---------|---------|
| `WRITER_BATCH_SIZE` | `50` | Max events per ClickHouse flush |
| `WRITER_FLUSH_INTERVAL` | `1s` | Time-based flush |
| `WRITER_MAX_RETRIES` | `3` | Retries on batch write failure |
| `WRITER_RETRY_BACKOFF` | `200ms` | Delay between retries |

Shutdown flushes the in-flight batch before exit.

Then query ClickHouse:

```bash
docker exec apexio-clickhouse clickhouse-client --query \
  "SELECT timestamp, service, log_level, message FROM apexio.logs ORDER BY timestamp DESC LIMIT 5"
```

Stop (keeps data volumes):

```bash
make down
```

Reset data volumes:

```bash
make clean-volumes
```

## Ports

| Service    | Host port | Purpose                          |
|------------|-----------|----------------------------------|
| Gateway    | 18080     | REST `/api/v1/log`, OTLP HTTP `/v1/logs`, `/metrics` |
| Gateway    | 4317      | OTLP gRPC logs export                                |
| Writer     | 8081      | Health only                      |
| Redpanda   | 19092     | Kafka API                        |
| Redpanda   | 18081     | Schema Registry                  |
| Redpanda   | 18082     | HTTP proxy (Pandaproxy)          |
| ClickHouse | 8123      | HTTP interface                   |
| ClickHouse | 9000      | Native protocol                  |
| Grafana    | 3000      | UI + provisioned **Apexio Logs** dashboard (`/d/apexio-logs`) |

## Layout

- [`compose/docker-compose.yml`](compose/docker-compose.yml) — stack definition
- [`../.env.example`](../.env.example) — configuration template (copy to `.env`, never commit secrets)
- [`clickhouse/init/01_schema.sql`](clickhouse/init/01_schema.sql) — `apexio.logs` table
- [`grafana/provisioning/`](grafana/provisioning/) — ClickHouse datasource + **Apexio Logs** dashboard (as code)

## Notes

- ClickHouse init SQL runs only on first start of an empty data volume.
- Grafana installs the ClickHouse plugin on first start (`GF_INSTALL_PLUGINS`); allow ~30–60s before datasource API checks.
- Shared Go contracts: [`pkg/`](../pkg/README.md) (`schema.LogEvent`, `broker.Broker`, `store.Store`).
- App services: [`cmd/gateway`](../cmd/gateway/), [`cmd/writer`](../cmd/writer/).
