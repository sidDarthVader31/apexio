# Apexio deploy (Phase 1)

Self-hosted infra for the redesigned pipeline: **Redpanda**, **ClickHouse**, and **Grafana**.

## Quick start

From the repository root:

```bash
make up
# or
docker compose -f deploy/compose/docker-compose.yml up -d
```

Verify:

```bash
make test-phase1
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
- Application services (gateway/writer) arrive in later phases; Phase 1 is infra only.
