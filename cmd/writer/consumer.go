package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// runConsumer batches Kafka messages, writes to ClickHouse, then commits offsets.
func runConsumer(ctx context.Context, cfg broker.RedpandaConfig, topic string, proc *batchProcessor) error {
	if topic == "" {
		return errors.New("writer: topic is required")
	}
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0,
		StartOffset:    kafka.FirstOffset,
	})
	defer reader.Close()

	var events []schema.LogEvent
	var msgs []kafka.Message
	ticker := time.NewTicker(proc.cfg.FlushInterval)
	defer ticker.Stop()

	flush := func() error {
		if len(events) == 0 {
			return nil
		}
		if err := proc.flush(ctx, events); err != nil {
			return err
		}
		if err := reader.CommitMessages(ctx, msgs...); err != nil {
			return fmt.Errorf("writer: commit: %w", err)
		}
		events = events[:0]
		msgs = msgs[:0]
		return nil
	}

	for {
		readCtx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
		msg, err := reader.FetchMessage(readCtx)
		cancel()

		if err == nil {
			ev, decErr := schema.UnmarshalEvent(msg.Value)
			if decErr != nil {
				proc.metrics.incDecodeErrors()
				return fmt.Errorf("writer: decode: %w", decErr)
			}
			events = append(events, ev)
			msgs = append(msgs, msg)
			if len(events) >= proc.cfg.BatchSize {
				if err := flush(); err != nil {
					return err
				}
			}
			continue
		}

		if errors.Is(err, context.DeadlineExceeded) {
			select {
			case <-ticker.C:
				if err := flush(); err != nil {
					return err
				}
			default:
			}
			if ctx.Err() != nil {
				return flush()
			}
			continue
		}

		if errors.Is(err, context.Canceled) {
			return flush()
		}
		return fmt.Errorf("writer: fetch: %w", err)
	}
}
