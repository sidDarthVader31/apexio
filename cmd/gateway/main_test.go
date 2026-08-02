package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

func TestIngestHandlerSuccess(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	h := &ingestHandler{bus: bus, topic: schema.DefaultTopic}

	body := `{
	  "id": 99,
	  "timestamp": 1732974309000,
	  "logLevel": "INFO",
	  "message": "ok",
	  "metadata": {"requestId": "r1", "responseStatus": 200, "responseDuration": 10},
	  "source": {"service": "payments", "host": "h1", "environment": "dev"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/log", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	msgs := bus.Messages(schema.DefaultTopic)
	if len(msgs) != 1 {
		t.Fatalf("published=%d", len(msgs))
	}
	ev, err := schema.UnmarshalEvent(msgs[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Service != "payments" || ev.Message != "ok" {
		t.Fatalf("event=%+v", ev)
	}
}

func TestIngestHandlerValidationError(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	h := &ingestHandler{bus: bus, topic: schema.DefaultTopic}

	body := `{"id":1,"timestamp":1,"logLevel":"INFO","message":"x","source":{"service":""}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/log", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(bus.Messages(schema.DefaultTopic)) != 0 {
		t.Fatal("should not publish invalid events")
	}
}

func TestIngestHandlerPublishFailure(t *testing.T) {
	bus := broker.NewMemory()
	_ = bus.Close()
	h := &ingestHandler{bus: bus, topic: schema.DefaultTopic}

	body := `{
	  "id": 1,
	  "timestamp": 1732974309000,
	  "logLevel": "INFO",
	  "message": "ok",
	  "source": {"service": "api"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/log", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadGateway {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
}

func TestIngestHandlerInvalidJSON(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	h := &ingestHandler{bus: bus, topic: schema.DefaultTopic}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/log", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestIngestUsesRequestContext(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	h := &ingestHandler{bus: bus, topic: schema.DefaultTopic}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	cancel()
	body := `{
	  "id": 1,
	  "timestamp": 1732974309000,
	  "logLevel": "INFO",
	  "message": "ok",
	  "source": {"service": "api"}
	}`
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/api/v1/log", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	// Cancelled context should fail publish.
	if rr.Code == http.StatusCreated {
		t.Fatalf("expected failure on cancelled context, got %d", rr.Code)
	}
}

func TestResponseShape(t *testing.T) {
	bus := broker.NewMemory()
	defer bus.Close()
	h := &ingestHandler{bus: bus, topic: schema.DefaultTopic}
	body := `{
	  "id": 5,
	  "timestamp": 1732974309000,
	  "logLevel": "ERROR",
	  "message": "fail",
	  "source": {"service": "api"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/log", bytes.NewBufferString(body))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["success"] != true {
		t.Fatalf("resp=%v", resp)
	}
}
