package broker

import (
	"context"
	"errors"
	"fmt"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// ErrNotImplemented is returned by thin Phase-2 stubs until Phase 3 wires clients.
var ErrNotImplemented = errors.New("broker backend not implemented yet")

// RedpandaConfig holds connection settings for the future Redpanda client.
type RedpandaConfig struct {
	Brokers []string
	ClientID string
}

// Redpanda is a thin stub satisfying Broker. Real client wiring is Phase 3.
type Redpanda struct {
	cfg    RedpandaConfig
	closed bool
}

// NewRedpanda validates config and returns a stub broker.
func NewRedpanda(cfg RedpandaConfig) (*Redpanda, error) {
	if len(cfg.Brokers) == 0 {
		return nil, errors.New("redpanda: at least one broker address is required")
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "apexio"
	}
	return &Redpanda{cfg: cfg}, nil
}

// Publish is not implemented in Phase 2.
func (r *Redpanda) Publish(ctx context.Context, topic string, event schema.LogEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.closed {
		return ErrBrokerClosed
	}
	if topic == "" {
		return errors.New("redpanda: topic is required")
	}
	if err := event.Validate(); err != nil {
		return fmt.Errorf("redpanda: %w", err)
	}
	return fmt.Errorf("%w: redpanda publish (phase 3)", ErrNotImplemented)
}

// Subscribe is not implemented in Phase 2.
func (r *Redpanda) Subscribe(ctx context.Context, topics []string, handler Handler) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if r.closed {
		return ErrBrokerClosed
	}
	if len(topics) == 0 {
		return errors.New("redpanda: at least one topic is required")
	}
	if handler == nil {
		return errors.New("redpanda: handler is nil")
	}
	return fmt.Errorf("%w: redpanda subscribe (phase 3)", ErrNotImplemented)
}

// Close marks the stub closed.
func (r *Redpanda) Close() error {
	r.closed = true
	return nil
}

// Config returns a copy of the stub configuration (test helper).
func (r *Redpanda) Config() RedpandaConfig {
	return r.cfg
}
