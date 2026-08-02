-- Apexio: canonical logs table for access-log / OTLP-shaped events.
-- Applied on first ClickHouse container start via /docker-entrypoint-initdb.d.

CREATE DATABASE IF NOT EXISTS apexio;

CREATE TABLE IF NOT EXISTS apexio.logs
(
    timestamp DateTime64(3, 'UTC'),
    id UInt64,
    log_level LowCardinality(String),
    message String,
    service LowCardinality(String),
    host LowCardinality(String),
    environment LowCardinality(String),
    request_id String,
    client_ip String,
    user_agent String,
    request_method LowCardinality(String),
    request_path String,
    response_status UInt16,
    response_duration_ms Float64,
    attrs Map(String, String),
    ingested_at DateTime64(3, 'UTC') DEFAULT now64(3)
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (service, timestamp, id)
TTL toDateTime(timestamp) + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;
