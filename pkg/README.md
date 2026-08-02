# Apexio packages (Phase 2)

Shared contracts for the redesigned pipeline. No network I/O in this phase —
Redpanda/ClickHouse types are thin stubs; memory backends power unit tests.

## Layout

| Package | Role |
|---------|------|
| [`pkg/schema`](schema/) | Canonical `LogEvent`, REST + OTLP-like mapping, broker JSON codec |
| [`pkg/broker`](broker/) | `Broker` / `Publisher` / `Consumer` + `Memory` + `Redpanda` stub |
| [`pkg/store`](store/) | `Store.WriteBatch` + `Memory` + `ClickHouse` stub |

## Canonical event

`schema.LogEvent` aligns with ClickHouse `apexio.logs`:

- identity: `timestamp`, `id`, `log_level`, `message`
- source: `service`, `host`, `environment`
- HTTP fields: `request_*`, `response_status`, `response_duration_ms`
- `attrs` map for extras / OTLP attributes

REST ingest (legacy `/api/v1/log`) maps through `schema.FromREST` / `ToREST`.
Broker wire format is flat JSON via `MarshalEvent` / `UnmarshalEvent`.
Default topic: `logs.ingestion.raw.v1` (`schema.DefaultTopic`).

## Test

```bash
make test-phase2
# or
go test ./pkg/...
```
