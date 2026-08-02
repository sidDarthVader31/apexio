package main

import (
	"context"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

func sampleEvent(msg string) schema.LogEvent {
	ev, _ := schema.FromREST(schema.RESTPayload{
		ID:        uint64(time.Now().UnixNano()),
		Timestamp: uint64(time.Now().UTC().UnixMilli()),
		LogLevel:  "INFO",
		Message:   msg,
		Source:    schema.RESTSource{Service: "batch-test"},
	})
	return ev
}

func TestBatchProcessorFlush(t *testing.T) {
	db := store.NewMemory()
	defer db.Close()
	metrics := &writerMetrics{}
	proc := newBatchProcessor(db, batchConfig{
		BatchSize:     3,
		FlushInterval: time.Hour,
		MaxRetries:    2,
		RetryBackoff:  time.Millisecond,
	}, metrics)

	ctx := context.Background()
	events := []schema.LogEvent{
		sampleEvent("a"),
		sampleEvent("b"),
		sampleEvent("c"),
	}
	if err := proc.flush(ctx, events); err != nil {
		t.Fatal(err)
	}
	if db.Len() != 3 {
		t.Fatalf("len=%d", db.Len())
	}
	if metrics.batchesFlush.Load() < 1 {
		t.Fatal("expected flush metric")
	}
}

func TestBatchProcessorRejectsBadPayload(t *testing.T) {
	_, err := schema.UnmarshalEvent([]byte(`{"message":"x"}`))
	if err == nil {
		t.Fatal("expected decode error")
	}
}
