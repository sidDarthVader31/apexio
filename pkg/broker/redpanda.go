package broker

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/segmentio/kafka-go"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// RedpandaConfig holds connection settings for Redpanda (Kafka API).
type RedpandaConfig struct {
	Brokers  []string
	ClientID string
	GroupID  string // consumer group for Subscribe
}

// Redpanda implements Broker against Redpanda / Kafka-compatible brokers.
type Redpanda struct {
	cfg    RedpandaConfig
	writer *kafka.Writer

	mu     sync.Mutex
	closed bool
}

// NewRedpanda validates config and opens a Kafka writer (lazy topic auto-create).
func NewRedpanda(cfg RedpandaConfig) (*Redpanda, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("redpanda: at least one broker address is required")
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "apexio"
	}
	if cfg.GroupID == "" {
		cfg.GroupID = "apexio-writer"
	}
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		RequiredAcks:           kafka.RequireOne,
		AllowAutoTopicCreation: true,
		Async:                  false,
	}
	return &Redpanda{cfg: cfg, writer: w}, nil
}

// Publish marshals the event and writes it to the topic with ack=1.
func (r *Redpanda) Publish(ctx context.Context, topic string, event schema.LogEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	closed := r.closed
	w := r.writer
	r.mu.Unlock()
	if closed || w == nil {
		return ErrBrokerClosed
	}
	if topic == "" {
		return errors.New("redpanda: topic is required")
	}
	payload, err := schema.MarshalEvent(event)
	if err != nil {
		return fmt.Errorf("redpanda: %w", err)
	}
	err = w.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Value: payload,
	})
	if err != nil {
		return fmt.Errorf("redpanda: publish: %w", err)
	}
	return nil
}

// Subscribe consumes topics in a loop until ctx is cancelled.
// Messages are committed only after handler returns nil.
func (r *Redpanda) Subscribe(ctx context.Context, topics []string, handler Handler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	closed := r.closed
	cfg := r.cfg
	r.mu.Unlock()
	if closed {
		return ErrBrokerClosed
	}
	if len(topics) == 0 {
		return errors.New("redpanda: at least one topic is required")
	}
	if handler == nil {
		return errors.New("redpanda: handler is nil")
	}

	// kafka-go Reader is single-topic; fan-in with one goroutine per topic.
	errCh := make(chan error, len(topics))
	subCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	for _, topic := range topics {
		topic := topic
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := r.consumeTopic(subCtx, cfg, topic, handler); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
				cancel()
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-done
		return ctx.Err()
	case err := <-errCh:
		cancel()
		<-done
		return err
	case <-done:
		return nil
	}
}

func (r *Redpanda) consumeTopic(ctx context.Context, cfg RedpandaConfig, topic string, handler Handler) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // manual commit after handler success
		StartOffset:    kafka.FirstOffset,
	})
	defer reader.Close()

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return ctx.Err()
			}
			return fmt.Errorf("redpanda: fetch %s: %w", topic, err)
		}
		err = handler(ctx, Message{
			Topic: msg.Topic,
			Key:   msg.Key,
			Value: msg.Value,
		})
		if err != nil {
			// Do not commit — message can be redelivered.
			return fmt.Errorf("redpanda: handler %s: %w", topic, err)
		}
		if err := reader.CommitMessages(ctx, msg); err != nil {
			return fmt.Errorf("redpanda: commit %s: %w", topic, err)
		}
	}
}

// Close shuts down the Kafka writer.
func (r *Redpanda) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closed = true
	if r.writer == nil {
		return nil
	}
	err := r.writer.Close()
	r.writer = nil
	return err
}

// Config returns a copy of the configuration (test helper).
func (r *Redpanda) Config() RedpandaConfig {
	return r.cfg
}
