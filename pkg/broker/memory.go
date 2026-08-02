package broker

import (
	"context"
	"errors"
	"sync"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// ErrBrokerClosed is returned when Publish/Subscribe is used after Close.
var ErrBrokerClosed = errors.New("broker is closed")

// Memory is an in-process Broker for unit tests and local dry-runs.
type Memory struct {
	mu       sync.RWMutex
	closed   bool
	topics   map[string][]Message
	handlers map[string][]Handler
}

// NewMemory returns an empty in-memory broker.
func NewMemory() *Memory {
	return &Memory{
		topics:   make(map[string][]Message),
		handlers: make(map[string][]Handler),
	}
}

// Publish appends the event to the topic and synchronously invokes handlers.
func (m *Memory) Publish(ctx context.Context, topic string, event schema.LogEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	payload, err := schema.MarshalEvent(event)
	if err != nil {
		return err
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return ErrBrokerClosed
	}
	msg := Message{Topic: topic, Value: payload}
	m.topics[topic] = append(m.topics[topic], msg)
	handlers := append([]Handler(nil), m.handlers[topic]...)
	m.mu.Unlock()

	for _, h := range handlers {
		if err := h(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Subscribe registers a handler. For Memory, messages already buffered on the
// topic are replayed, then the handler receives future Publish calls.
func (m *Memory) Subscribe(ctx context.Context, topics []string, handler Handler) error {
	if handler == nil {
		return errors.New("handler is nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return ErrBrokerClosed
	}

	var replay []Message
	for _, topic := range topics {
		m.handlers[topic] = append(m.handlers[topic], handler)
		replay = append(replay, m.topics[topic]...)
	}
	m.mu.Unlock()

	for _, msg := range replay {
		if err := handler(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

// Messages returns a copy of buffered messages for a topic (test helper).
func (m *Memory) Messages(topic string) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Message, len(m.topics[topic]))
	copy(out, m.topics[topic])
	return out
}

// Close marks the broker closed.
func (m *Memory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
