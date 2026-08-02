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
make test-phase3   # unit + E2E vertical slice
```

### Ingest a log (Phase 3)

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
| Gateway    | 18080     | REST ingest `POST /api/v1/log`   |
| Writer     | 8081      | Health only                      |
| Redpanda   | 19092     | Kafka API                        |
| Redpanda   | 18081     | Schema Registry                  |
| Redpanda   | 18082     | HTTP proxy (Pandaproxy)          |
| ClickHouse | 8123      | HTTP interface                   |
| ClickHouse | 9000      | Native protocol                  |
| Grafana    | 3000      | UI (`admin` / `admin`)           |

## Layout

- [`compose/docker-compose.yml`](compose/docker-compose.yml) — stack definition
- [`clickhouse/init/01_schema.sql`](clickhouse/init/01_schema.sql) — `apexio.logs` table
- [`grafana/provisioning/`](grafana/provisioning/) — ClickHouse datasource + dashboard provider

## Notes

- ClickHouse init SQL runs only on first start of an empty data volume.
- Grafana installs the ClickHouse plugin on first start (`GF_INSTALL_PLUGINS`); allow ~30–60s before datasource API checks.
- Shared Go contracts: [`pkg/`](../pkg/README.md) (`schema.LogEvent`, `broker.Broker`, `store.Store`).
- App services: [`cmd/gateway`](../cmd/gateway/), [`cmd/writer`](../cmd/writer/).
