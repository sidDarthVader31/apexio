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

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("POST /api/v1/log", &ingestHandler{bus: bus, topic: topic})

	srv := &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("gateway listening on %s topic=%s brokers=%v", addr, topic, brokers)
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

type ingestHandler struct {
	bus   broker.Publisher
	topic string
}

func (h *ingestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload schema.RESTPayload
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&payload); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json", err)
		return
	}
	ev, err := schema.FromREST(payload)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "validation failed", err)
		return
	}
	if err := h.bus.Publish(r.Context(), h.topic, ev); err != nil {
		writeErr(w, http.StatusBadGateway, "publish failed", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"id":      ev.ID,
		"service": ev.Service,
	})
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
