package store

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// ClickHouseConfig holds connection settings for ClickHouse.
// Provide either DSN or Addr (+ optional auth/database).
type ClickHouseConfig struct {
	DSN      string // e.g. clickhouse://default@localhost:9000/apexio
	Addr     string // e.g. localhost:9000 or clickhouse:9000
	Database string
	Table    string
	Username string
	Password string
	// PingAttempts is the number of Ping tries on connect (default 30, ~60s total).
	PingAttempts int
}

// ClickHouse implements Store with bulk inserts into apexio.logs.
type ClickHouse struct {
	cfg  ClickHouseConfig
	conn driver.Conn

	mu     sync.Mutex
	closed bool
}

// NewClickHouse opens a native-protocol connection.
func NewClickHouse(cfg ClickHouseConfig) (*ClickHouse, error) {
	if cfg.Database == "" {
		cfg.Database = "apexio"
	}
	if cfg.Table == "" {
		cfg.Table = "logs"
	}
	if cfg.Username == "" {
		cfg.Username = "default"
	}

	var (
		conn driver.Conn
		err  error
	)
	switch {
	case cfg.DSN != "":
		opts, parseErr := clickhouse.ParseDSN(cfg.DSN)
		if parseErr != nil {
			return nil, fmt.Errorf("clickhouse: parse DSN: %w", parseErr)
		}
		if opts.Auth.Database == "" {
			opts.Auth.Database = cfg.Database
		}
		conn, err = clickhouse.Open(opts)
	case cfg.Addr != "":
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{cfg.Addr},
			Auth: clickhouse.Auth{
				Database: cfg.Database,
				Username: cfg.Username,
				Password: cfg.Password,
			},
		})
	default:
		return nil, errors.New("clickhouse: DSN or Addr is required")
	}
	if err != nil {
		return nil, fmt.Errorf("clickhouse: open: %w", err)
	}
	var pingErr error
	attempts := cfg.PingAttempts
	if attempts < 1 {
		attempts = 30
	}
	for attempt := 1; attempt <= attempts; attempt++ {
		pingErr = conn.Ping(context.Background())
		if pingErr == nil {
			return &ClickHouse{cfg: cfg, conn: conn}, nil
		}
		time.Sleep(2 * time.Second)
	}
	_ = conn.Close()
	return nil, fmt.Errorf("clickhouse: ping: %w", pingErr)
}

// WriteBatch inserts events in a single prepared batch.
func (c *ClickHouse) WriteBatch(ctx context.Context, events []schema.LogEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	c.mu.Lock()
	closed := c.closed
	conn := c.conn
	cfg := c.cfg
	c.mu.Unlock()
	if closed || conn == nil {
		return ErrStoreClosed
	}
	if len(events) == 0 {
		return nil
	}
	for i := range events {
		if err := events[i].Validate(); err != nil {
			return fmt.Errorf("clickhouse: event[%d]: %w", i, err)
		}
	}

	query := fmt.Sprintf(
		`INSERT INTO %s.%s (
			timestamp, id, log_level, message, service, host, environment,
			request_id, client_ip, user_agent, request_method, request_path,
			response_status, response_duration_ms, attrs
		)`,
		cfg.Database, cfg.Table,
	)
	batch, err := conn.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("clickhouse: prepare batch: %w", err)
	}
	for _, e := range events {
		attrs := e.Attrs
		if attrs == nil {
			attrs = map[string]string{}
		}
		if err := batch.Append(
			e.Timestamp.UTC(),
			e.ID,
			e.LogLevel,
			e.Message,
			e.Service,
			e.Host,
			e.Environment,
			e.RequestID,
			e.ClientIP,
			e.UserAgent,
			e.RequestMethod,
			e.RequestPath,
			e.ResponseStatus,
			e.ResponseDurationMs,
			attrs,
		); err != nil {
			return fmt.Errorf("clickhouse: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("clickhouse: send: %w", err)
	}
	return nil
}

// Close closes the ClickHouse connection.
func (c *ClickHouse) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

// Config returns a copy of the configuration (test helper).
func (c *ClickHouse) Config() ClickHouseConfig {
	return c.cfg
}
