package store

import (
	"context"
	"io"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// Store persists batches of log events (ClickHouse in production).
type Store interface {
	WriteBatch(ctx context.Context, events []schema.LogEvent) error
	io.Closer
}
