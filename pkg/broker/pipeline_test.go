package broker_test

import (
	"context"
	"testing"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

// TestPipelineContracts exercises Broker + Store together with memory backends,
// mirroring the gateway→writer handoff without network I/O.
func TestPipelineContracts(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	db := store.NewMemory()
	defer db.Close()

	err := bus.Subscribe(context.Background(), []string{schema.DefaultTopic}, func(ctx context.Context, msg broker.Message) error {
		ev, err := schema.UnmarshalEvent(msg.Value)
		if err != nil {
			return err
		}
		return db.WriteBatch(ctx, []schema.LogEvent{ev})
	})
	if err != nil {
		t.Fatal(err)
	}

	ev := sampleEvent(t)
	if err := bus.Publish(context.Background(), schema.DefaultTopic, ev); err != nil {
		t.Fatal(err)
	}
	if db.Len() != 1 {
		t.Fatalf("store len=%d", db.Len())
	}
	got := db.All()[0]
	if got.Service != ev.Service || got.Message != ev.Message {
		t.Fatalf("stored=%+v", got)
	}
}
