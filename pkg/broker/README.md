# 🔗 Broker (Redpanda / Kafka API)

---

Pluggable message queue between **gateway** (publish) and **writer** (consume).

## 📐 Interface

`Broker` combines `Publisher` and `Consumer`. Swap implementations via `NewRedpanda` or your own type satisfying the interface.

```go
type Broker interface {
    Publisher
    Consumer
}
```

See [`memory.go`](memory.go) for a test double and [`redpanda.go`](redpanda.go) for the default.

---

## 📬 Delivery policy (Redpanda default)

| Setting | Value | Meaning |
|---------|-------|---------|
| `RequiredAcks` | `RequireOne` (acks=1) | Leader ack before publish returns success |
| `Async` | `false` | Gateway waits for broker ack; publish errors surface to clients |
| Consumer commit | After batch flush succeeds | Writer fetches up to `WRITER_BATCH_SIZE` (or `WRITER_FLUSH_INTERVAL`), writes to ClickHouse, then commits offsets |

---

## ⏳ Backpressure

- **Publish** — synchronous `WriteMessages`; slow broker blocks gateway ingest handlers (REST/OTLP return 502 on failure).
- **Consume** — writer batches fetches per partition; slow ClickHouse writes delay commits for that partition (at-least-once).

---

## 🔧 Bring your own broker

Implement `broker.Broker` in a new file (see `memory.go` / `redpanda.go`), then wire it in `cmd/gateway` and `cmd/writer` `main.go` instead of `NewRedpanda`.

Supported alternatives: any Kafka-compatible broker, or fully custom (NATS, RabbitMQ, etc.) behind the same interface.
