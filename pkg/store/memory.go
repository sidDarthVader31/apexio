package store

import (
	"context"
	"errors"
	"sync"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

// ErrStoreClosed is returned when WriteBatch is used after Close.
var ErrStoreClosed = errors.New("store is closed")

// Memory is an in-process Store for unit tests.
type Memory struct {
	mu     sync.RWMutex
	closed bool
	events []schema.LogEvent
}

// NewMemory returns an empty in-memory store.
func NewMemory() *Memory {
	return &Memory{}
}

// WriteBatch appends events after validating each one.
func (m *Memory) WriteBatch(ctx context.Context, events []schema.LogEvent) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(events) == 0 {
		return nil
	}
	for i := range events {
		if err := events[i].Validate(); err != nil {
			return err
		}
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return ErrStoreClosed
	}
	m.events = append(m.events, events...)
	return nil
}

// All returns a copy of stored events (test helper).
func (m *Memory) All() []schema.LogEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]schema.LogEvent, len(m.events))
	copy(out, m.events)
	return out
}

// Len returns the number of stored events.
func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.events)
}

// Close marks the store closed.
func (m *Memory) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
