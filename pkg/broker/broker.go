package broker

import (
	"context"
	"io"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// Message is a single broker record delivered to a consumer.
type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// Handler processes one consumed message. Returning an error signals the
// implementation to retry or not commit, depending on the backend.
type Handler func(ctx context.Context, msg Message) error

// Publisher publishes log events to a topic.
type Publisher interface {
	Publish(ctx context.Context, topic string, event schema.LogEvent) error
	io.Closer
}

// Consumer subscribes to topics and invokes a handler per message.
type Consumer interface {
	Subscribe(ctx context.Context, topics []string, handler Handler) error
	io.Closer
}

// Broker is the full messaging abstraction (produce + consume).
// Default production wiring uses Redpanda; tests use the in-memory implementation.
type Broker interface {
	Publisher
	Consumer
}
