package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

type batchConfig struct {
	BatchSize     int
	FlushInterval time.Duration
	MaxRetries    int
	RetryBackoff  time.Duration
}

type batchProcessor struct {
	store   store.Store
	cfg     batchConfig
	metrics *writerMetrics
}

func newBatchProcessor(s store.Store, cfg batchConfig, metrics *writerMetrics) *batchProcessor {
	if cfg.BatchSize < 1 {
		cfg.BatchSize = 1
	}
	if cfg.FlushInterval <= 0 {
		cfg.FlushInterval = time.Second
	}
	if cfg.MaxRetries < 1 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = 200 * time.Millisecond
	}
	return &batchProcessor{
		store:   s,
		cfg:     cfg,
		metrics: metrics,
	}
}

func (p *batchProcessor) flush(ctx context.Context, events []schema.LogEvent) error {
	if len(events) == 0 {
		return nil
	}
	err := p.writeWithRetry(ctx, events)
	if err != nil {
		p.metrics.incWriteErrors()
		return err
	}
	p.metrics.incEventsWritten(uint64(len(events)))
	p.metrics.incBatchesFlushed()
	return nil
}

func (p *batchProcessor) writeWithRetry(ctx context.Context, events []schema.LogEvent) error {
	var err error
	for attempt := 1; attempt <= p.cfg.MaxRetries; attempt++ {
		err = p.store.WriteBatch(ctx, events)
		if err == nil {
			return nil
		}
		if attempt < p.cfg.MaxRetries {
			time.Sleep(p.cfg.RetryBackoff)
		}
	}
	return fmt.Errorf("write batch after %d retries: %w", p.cfg.MaxRetries, err)
}
