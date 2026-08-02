package main

import (
	"context"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

func TestProcessorWritesToStore(t *testing.T) {
	db := store.NewMemory()
	defer db.Close()
	proc := newBatchProcessor(db, batchConfig{BatchSize: 1, FlushInterval: time.Second}, &writerMetrics{})

	ev, err := schema.FromREST(schema.RESTPayload{
		ID:        3,
		Timestamp: uint64(time.Now().UTC().UnixMilli()),
		LogLevel:  "ERROR",
		Message:   "writer-unit",
		Source:    schema.RESTSource{Service: "gateway"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.flush(context.Background(), []schema.LogEvent{ev}); err != nil {
		t.Fatal(err)
	}
	if db.Len() != 1 {
		t.Fatalf("len=%d", db.Len())
	}
	if db.All()[0].Message != "writer-unit" {
		t.Fatalf("got=%+v", db.All()[0])
	}
}
