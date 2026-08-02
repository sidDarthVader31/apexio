# 📦 Apexio packages

---

Shared contracts for the log pipeline. All services import from here — **fork and extend** these interfaces for your org.

## 🗂️ Layout

| Package | Role |
|---------|------|
| [`pkg/schema`](schema/) | Canonical `LogEvent`, REST + OTLP mapping, broker JSON codec |
| [`pkg/broker`](broker/) | `Broker` / `Publisher` / `Consumer` + `Memory` + **Redpanda** |
| [`pkg/store`](store/) | `Store.WriteBatch` + `Memory` + **ClickHouse** |
| [`pkg/auth`](auth/) | Optional HTTP middleware (API-key example) |

---

## 📋 Canonical event

`schema.LogEvent` aligns with ClickHouse `apexio.logs`:

- **identity** — `timestamp`, `id`, `log_level`, `message`
- **source** — `service`, `host`, `environment`
- **HTTP** — `request_*`, `response_status`, `response_duration_ms`
- **extras** — `attrs` map for OTLP attributes and custom fields

**Wire formats**

- REST ingest (`/api/v1/log`) → `schema.FromREST` / `ToREST`
- Broker payload → flat JSON via `MarshalEvent` / `UnmarshalEvent`
- OTLP → `LogEventsFromOTLP` (HTTP `/v1/logs`, gRPC `:4317`)

Default topic: `logs.ingestion.raw.v1` (`schema.DefaultTopic`).

---

## 🧪 Test

```bash
make test-unit
# or
go test ./pkg/... -count=1
```
