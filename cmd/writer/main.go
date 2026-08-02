package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
	"github.com/sidDarthVader31/apexio/pkg/store"
)

func main() {
	brokers := strings.Split(envOr("REDPANDA_BROKERS", "localhost:19092"), ",")
	topic := envOr("LOG_TOPIC", schema.DefaultTopic)
	healthAddr := envOr("WRITER_HEALTH_ADDR", ":8081")
	chAddr := envOr("CLICKHOUSE_ADDR", "localhost:9000")

	bus, err := broker.NewRedpanda(broker.RedpandaConfig{
		Brokers:  brokers,
		ClientID: envOr("REDPANDA_CLIENT_ID", "apexio-writer"),
		GroupID:  envOr("REDPANDA_GROUP_ID", "apexio-writer"),
	})
	if err != nil {
		log.Fatalf("broker: %v", err)
	}
	defer bus.Close()

	db, err := store.NewClickHouse(store.ClickHouseConfig{
		Addr:     chAddr,
		Database: envOr("CLICKHOUSE_DATABASE", "apexio"),
		Table:    envOr("CLICKHOUSE_TABLE", "logs"),
		Username: envOr("CLICKHOUSE_USER", "default"),
		Password: os.Getenv("CLICKHOUSE_PASSWORD"),
	})
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer db.Close()

	proc := &processor{store: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	healthSrv := &http.Server{Addr: healthAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("writer health on %s", healthAddr)
		if err := healthSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("health listen: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("writer consuming topic=%s brokers=%v clickhouse=%s", topic, brokers, chAddr)
	err = bus.Subscribe(ctx, []string{topic}, proc.handle)
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("subscribe: %v", err)
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = healthSrv.Shutdown(shutdownCtx)
}

type processor struct {
	store store.Store
}

func (p *processor) handle(ctx context.Context, msg broker.Message) error {
	ev, err := schema.UnmarshalEvent(msg.Value)
	if err != nil {
		return err
	}
	return p.store.WriteBatch(ctx, []schema.LogEvent{ev})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
