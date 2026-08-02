package store_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

func sampleEvent(t *testing.T) schema.LogEvent {
	t.Helper()
	ev, err := schema.FromREST(schema.RESTPayload{
		ID:        7,
		Timestamp: uint64(time.Now().UTC().UnixMilli()),
		LogLevel:  "WARN",
		Message:   "slow",
		Source:    schema.RESTSource{Service: "api"},
		Metadata:  schema.RESTMetadata{ResponseStatus: 200, ResponseDuration: 900},
	})
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

func TestMemoryWriteBatch(t *testing.T) {
	mem := store.NewMemory()
	defer mem.Close()

	batch := []schema.LogEvent{sampleEvent(t), sampleEvent(t)}
	if err := mem.WriteBatch(context.Background(), batch); err != nil {
		t.Fatal(err)
	}
	if mem.Len() != 2 {
		t.Fatalf("len=%d", mem.Len())
	}
	all := mem.All()
	if all[0].Message != "slow" {
		t.Fatalf("event=%+v", all[0])
	}
}

func TestMemoryRejectsInvalid(t *testing.T) {
	mem := store.NewMemory()
	defer mem.Close()
	bad := schema.LogEvent{Message: "x"} // missing service + timestamp
	if err := mem.WriteBatch(context.Background(), []schema.LogEvent{bad}); err == nil {
		t.Fatal("expected validation error")
	}
	if mem.Len() != 0 {
		t.Fatal("partial write should not stick")
	}
}

func TestMemoryClosed(t *testing.T) {
	mem := store.NewMemory()
	_ = mem.Close()
	if err := mem.WriteBatch(context.Background(), []schema.LogEvent{sampleEvent(t)}); !errors.Is(err, store.ErrStoreClosed) {
		t.Fatalf("err=%v", err)
	}
}

func TestMemoryImplementsStore(t *testing.T) {
	var _ store.Store = store.NewMemory()
}

func TestClickHouseStub(t *testing.T) {
	if _, err := store.NewClickHouse(store.ClickHouseConfig{}); err == nil {
		t.Fatal("expected error for empty DSN")
	}
	ch, err := store.NewClickHouse(store.ClickHouseConfig{DSN: "clickhouse://default@localhost:9000/apexio"})
	if err != nil {
		t.Fatal(err)
	}
	defer ch.Close()

	var _ store.Store = ch
	if ch.Config().Database != "apexio" || ch.Config().Table != "logs" {
		t.Fatalf("defaults=%+v", ch.Config())
	}

	err = ch.WriteBatch(context.Background(), []schema.LogEvent{sampleEvent(t)})
	if !errors.Is(err, store.ErrNotImplemented) {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}
