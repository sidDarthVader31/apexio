package main

import (
	"context"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

func TestProcessorWritesToStore(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	db := store.NewMemory()
	defer db.Close()
	proc := &processor{store: db}

	if err := bus.Subscribe(context.Background(), []string{schema.DefaultTopic}, proc.handle); err != nil {
		t.Fatal(err)
	}

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
	if err := bus.Publish(context.Background(), schema.DefaultTopic, ev); err != nil {
		t.Fatal(err)
	}
	if db.Len() != 1 {
		t.Fatalf("len=%d", db.Len())
	}
	if db.All()[0].Message != "writer-unit" {
		t.Fatalf("got=%+v", db.All()[0])
	}
}

func TestProcessorRejectsBadPayload(t *testing.T) {
	db := store.NewMemory()
	defer db.Close()
	proc := &processor{store: db}
	err := proc.handle(context.Background(), broker.Message{Value: []byte(`{"message":"x"}`)})
	if err == nil {
		t.Fatal("expected unmarshal/validation error")
	}
	if db.Len() != 0 {
		t.Fatal("should not write bad events")
	}
}
