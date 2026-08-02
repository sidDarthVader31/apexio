package broker_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

func sampleEvent(t *testing.T) schema.LogEvent {
	t.Helper()
	ev, err := schema.FromREST(schema.RESTPayload{
		ID:        1,
		Timestamp: uint64(time.Now().UTC().UnixMilli()),
		LogLevel:  "INFO",
		Message:   "hello",
		Source:    schema.RESTSource{Service: "api", Host: "h1", Environment: "dev"},
	})
	if err != nil {
		t.Fatal(err)
	}
	return ev
}

func TestMemoryPublishSubscribe(t *testing.T) {
	mem := broker.NewMemory()
	defer mem.Close()

	var saw atomic.Int32
	err := mem.Subscribe(context.Background(), []string{schema.DefaultTopic}, func(ctx context.Context, msg broker.Message) error {
		ev, err := schema.UnmarshalEvent(msg.Value)
		if err != nil {
			return err
		}
		if ev.Message != "hello" {
			return errors.New("bad message")
		}
		saw.Add(1)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := mem.Publish(context.Background(), schema.DefaultTopic, sampleEvent(t)); err != nil {
		t.Fatal(err)
	}
	if saw.Load() != 1 {
		t.Fatalf("handler calls=%d", saw.Load())
	}
	if len(mem.Messages(schema.DefaultTopic)) != 1 {
		t.Fatalf("buffered=%d", len(mem.Messages(schema.DefaultTopic)))
	}
}

func TestMemoryReplayOnSubscribe(t *testing.T) {
	mem := broker.NewMemory()
	defer mem.Close()

	if err := mem.Publish(context.Background(), schema.DefaultTopic, sampleEvent(t)); err != nil {
		t.Fatal(err)
	}

	var saw atomic.Int32
	if err := mem.Subscribe(context.Background(), []string{schema.DefaultTopic}, func(ctx context.Context, msg broker.Message) error {
		saw.Add(1)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if saw.Load() != 1 {
		t.Fatalf("expected replay, got %d", saw.Load())
	}
}

func TestMemoryClosed(t *testing.T) {
	mem := broker.NewMemory()
	_ = mem.Close()
	if err := mem.Publish(context.Background(), schema.DefaultTopic, sampleEvent(t)); !errors.Is(err, broker.ErrBrokerClosed) {
		t.Fatalf("err=%v", err)
	}
}

func TestMemoryImplementsBroker(t *testing.T) {
	var _ broker.Broker = broker.NewMemory()
}

func TestRedpandaConfigValidation(t *testing.T) {
	if _, err := broker.NewRedpanda(broker.RedpandaConfig{}); err == nil {
		t.Fatal("expected error for empty brokers")
	}
	rp, err := broker.NewRedpanda(broker.RedpandaConfig{Brokers: []string{"localhost:19092"}})
	if err != nil {
		t.Fatal(err)
	}
	defer rp.Close()

	var _ broker.Broker = rp
	if rp.Config().GroupID != "apexio-writer" {
		t.Fatalf("default group=%q", rp.Config().GroupID)
	}
	if rp.Config().ClientID != "apexio" {
		t.Fatalf("default client=%q", rp.Config().ClientID)
	}
}
