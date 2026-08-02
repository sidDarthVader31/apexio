package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"

	"github.com/sidDarthVader31/apexio/pkg/broker"
	"github.com/sidDarthVader31/apexio/pkg/schema"
)

func newTestHandlers(t *testing.T) (*ingestHandler, *otlpHTTPHandler, *broker.Memory) {
	t.Helper()
	bus := broker.NewMemory()
	pub := &publisher{bus: bus, topic: schema.DefaultTopic}
	metrics := &ingestMetrics{}
	return &ingestHandler{pub: pub, metrics: metrics}, &otlpHTTPHandler{pub: pub, metrics: metrics}, bus
}

func TestIngestHandlerSuccess(t *testing.T) {
	h, _, bus := newTestHandlers(t)
	defer bus.Close()

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
	h, _, bus := newTestHandlers(t)
	defer bus.Close()

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
	pub := &publisher{bus: bus, topic: schema.DefaultTopic}
	h := &ingestHandler{pub: pub, metrics: &ingestMetrics{}}

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
	h, _, bus := newTestHandlers(t)
	defer bus.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/log", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
}

func TestIngestUsesRequestContext(t *testing.T) {
	h, _, bus := newTestHandlers(t)
	defer bus.Close()
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
	if rr.Code == http.StatusCreated {
		t.Fatalf("expected failure on cancelled context, got %d", rr.Code)
	}
}

func TestResponseShape(t *testing.T) {
	h, _, bus := newTestHandlers(t)
	defer bus.Close()
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

func TestOTLPHttpHandlerProtobuf(t *testing.T) {
	_, oh, bus := newTestHandlers(t)
	defer bus.Close()

	reqPB := buildOTLPRequest(t, "otlp-http-msg", "demo-service")
	body, err := proto.Marshal(reqPB)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rr := httptest.NewRecorder()
	oh.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rr.Code, rr.Body.String())
	}
	if len(bus.Messages(schema.DefaultTopic)) != 1 {
		t.Fatalf("published=%d", len(bus.Messages(schema.DefaultTopic)))
	}
	ev, err := schema.UnmarshalEvent(bus.Messages(schema.DefaultTopic)[0].Value)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Message != "otlp-http-msg" || ev.Service != "demo-service" {
		t.Fatalf("event=%+v", ev)
	}
}

func TestOTLPHttpHandlerInvalid(t *testing.T) {
	_, oh, bus := newTestHandlers(t)
	defer bus.Close()
	req := httptest.NewRequest(http.MethodPost, "/v1/logs", bytes.NewBufferString("not-proto"))
	req.Header.Set("Content-Type", "application/x-protobuf")
	rr := httptest.NewRecorder()
	oh.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rr.Code)
	}
	if len(bus.Messages(schema.DefaultTopic)) != 0 {
		t.Fatal("should not publish")
	}
}

func buildOTLPRequest(t *testing.T, message, service string) *collogsv1.ExportLogsServiceRequest {
	t.Helper()
	return &collogsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsv1.ResourceLogs{{
			Resource: &resourcev1.Resource{
				Attributes: []*commonv1.KeyValue{{
					Key: "service.name",
					Value: &commonv1.AnyValue{
						Value: &commonv1.AnyValue_StringValue{StringValue: service},
					},
				}},
			},
			ScopeLogs: []*logsv1.ScopeLogs{{
				LogRecords: []*logsv1.LogRecord{{
					TimeUnixNano: uint64(time.Now().UnixNano()),
					SeverityText: "INFO",
					Body: &commonv1.AnyValue{
						Value: &commonv1.AnyValue_StringValue{StringValue: message},
					},
				}},
			}},
		}},
	}
}
