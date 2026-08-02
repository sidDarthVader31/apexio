# Broker (Redpanda / Kafka API)

## Interface

`Broker` combines `Publisher` and `Consumer`. Swap implementations via `NewRedpanda` or your own type satisfying the interface.

## Delivery policy (Redpanda default)

| Setting | Value | Meaning |
|---------|-------|---------|
| `RequiredAcks` | `RequireOne` (acks=1) | Leader ack before publish returns success |
| `Async` | `false` | Gateway waits for broker ack; publish errors surface to clients |
| Consumer commit | After batch flush succeeds | Writer fetches up to `WRITER_BATCH_SIZE` (or `WRITER_FLUSH_INTERVAL`), writes to ClickHouse, then commits offsets |

## Backpressure

- **Publish**: synchronous `WriteMessages`; slow broker blocks gateway ingest handlers (REST/OTLP return 502 on failure).
- **Consume**: writer batches fetches per partition; slow ClickHouse writes delay commits for that partition (at-least-once).

## Bring your own broker

Implement `broker.Broker` in a new file (see `memory.go` / `redpanda.go`), then wire it in `cmd/gateway` and `cmd/writer` `main.go` instead of `NewRedpanda`.
