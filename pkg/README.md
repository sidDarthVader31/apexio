# Apexio packages

Shared contracts for the log pipeline.

| Package | Role |
|---------|------|
| [`pkg/schema`](schema/) | Canonical `LogEvent`, REST + OTLP-like mapping, broker JSON codec |
| [`pkg/broker`](broker/) | `Broker` / `Publisher` / `Consumer` + `Memory` + **Redpanda** |
| [`pkg/store`](store/) | `Store.WriteBatch` + `Memory` + **ClickHouse** |
| [`pkg/auth`](auth/) | Optional HTTP middleware (API-key example) |

## Canonical event

`schema.LogEvent` aligns with ClickHouse `apexio.logs`:

- identity: `timestamp`, `id`, `log_level`, `message`
- source: `service`, `host`, `environment`
- HTTP fields: `request_*`, `response_status`, `response_duration_ms`
- `attrs` map for extras / OTLP attributes

REST ingest (`/api/v1/log`) maps through `schema.FromREST` / `ToREST`.
Broker wire format is flat JSON via `MarshalEvent` / `UnmarshalEvent`.
Default topic: `logs.ingestion.raw.v1` (`schema.DefaultTopic`).

## Test

```bash
make test-unit
# or
go test ./pkg/...
```
