package main

import (
	"context"
	"encoding/json"
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
)

func main() {
	addr := envOr("GATEWAY_ADDR", ":8080")
	otlpGRPCAddr := envOr("GATEWAY_OTLP_GRPC_ADDR", ":4317")
	brokers := strings.Split(envOr("REDPANDA_BROKERS", "localhost:19092"), ",")
	topic := envOr("LOG_TOPIC", schema.DefaultTopic)

	bus, err := broker.NewRedpanda(broker.RedpandaConfig{
		Brokers:  brokers,
		ClientID: envOr("REDPANDA_CLIENT_ID", "apexio-gateway"),
	})
	if err != nil {
		log.Fatalf("broker: %v", err)
	}
	defer bus.Close()

	pub := &publisher{bus: bus, topic: topic}
	metrics := &ingestMetrics{}

	grpcSrv, err := startOTLPGRPC(otlpGRPCAddr, pub, metrics)
	if err != nil {
		log.Fatalf("otlp grpc: %v", err)
	}
	if grpcSrv != nil {
		defer grpcSrv.GracefulStop()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /metrics", metrics.handler())
	mux.Handle("POST /api/v1/log", &ingestHandler{pub: pub, metrics: metrics})
	mux.Handle("POST /v1/logs", &otlpHTTPHandler{pub: pub, metrics: metrics})

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("gateway http on %s (rest=/api/v1/log otlp=/v1/logs) topic=%s brokers=%v", addr, topic, brokers)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func writeErr(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   msg,
		"details": err.Error(),
	})
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
