package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// ErrNotImplemented is returned by thin Phase-2 stubs until Phase 3 wires clients.
var ErrNotImplemented = errors.New("store backend not implemented yet")

// ClickHouseConfig holds connection settings for the future ClickHouse client.
type ClickHouseConfig struct {
	DSN      string // e.g. clickhouse://default@localhost:9000/apexio
	Database string
	Table    string
}

// ClickHouse is a thin stub satisfying Store. Real client wiring is Phase 3.
type ClickHouse struct {
	cfg    ClickHouseConfig
	closed bool
}

// NewClickHouse validates config and returns a stub store.
func NewClickHouse(cfg ClickHouseConfig) (*ClickHouse, error) {
	if cfg.DSN == "" {
		return nil, errors.New("clickhouse: DSN is required")
	}
	if cfg.Database == "" {
		cfg.Database = "apexio"
	}
	if cfg.Table == "" {
		cfg.Table = "logs"
	}
	return &ClickHouse{cfg: cfg}, nil
}

// WriteBatch is not implemented in Phase 2.
func (c *ClickHouse) WriteBatch(ctx context.Context, events []schema.LogEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if c.closed {
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
	return fmt.Errorf("%w: clickhouse write batch (phase 3)", ErrNotImplemented)
}

// Close marks the stub closed.
func (c *ClickHouse) Close() error {
	c.closed = true
	return nil
}

// Config returns a copy of the stub configuration (test helper).
func (c *ClickHouse) Config() ClickHouseConfig {
	return c.cfg
}
