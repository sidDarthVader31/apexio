package schema_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

func sampleREST() schema.RESTPayload {
	return schema.RESTPayload{
		ID:        42,
		Timestamp: 1732974309000,
		LogLevel:  "info",
		Message:   "502 upstream timeout",
		Metadata: schema.RESTMetadata{
			RequestID:        "req-1",
			ClientIP:         "10.0.0.1",
			UserAgent:        "curl/8.0",
			RequestMethod:    "GET",
			RequestPath:      "/payments",
			ResponseStatus:   502,
			ResponseDuration: 194.5,
			Extra:            map[string]string{"traceId": "abc"},
		},
		Source: schema.RESTSource{
			Host:        "api-1",
			Service:     "payments",
			Environment: "production",
			Extra:       map[string]string{"region": "us-east"},
		},
	}
}

func TestFromRESTAndRoundTrip(t *testing.T) {
	ev, err := schema.FromREST(sampleREST())
	if err != nil {
		t.Fatalf("FromREST: %v", err)
	}
	if ev.LogLevel != schema.LevelInfo {
		t.Fatalf("expected INFO, got %q", ev.LogLevel)
	}
	if ev.Service != "payments" {
		t.Fatalf("service=%q", ev.Service)
	}
	if ev.ResponseStatus != 502 {
		t.Fatalf("status=%d", ev.ResponseStatus)
	}
	if ev.Attrs["traceId"] != "abc" || ev.Attrs["region"] != "us-east" {
		t.Fatalf("attrs=%v", ev.Attrs)
	}
	if !ev.Timestamp.Equal(time.UnixMilli(1732974309000).UTC()) {
		t.Fatalf("timestamp=%v", ev.Timestamp)
	}

	rest := schema.ToREST(ev)
	if rest.LogLevel != schema.LevelInfo || rest.Source.Service != "payments" {
		t.Fatalf("ToREST mismatch: %+v", rest)
	}
	if rest.Timestamp != 1732974309000 {
		t.Fatalf("timestamp millis=%d", rest.Timestamp)
	}
}

func TestFromRESTValidation(t *testing.T) {
	p := sampleREST()
	p.Message = ""
	if _, err := schema.FromREST(p); err == nil {
		t.Fatal("expected error for empty message")
	}

	p = sampleREST()
	p.Source.Service = ""
	if _, err := schema.FromREST(p); err == nil {
		t.Fatal("expected error for empty service")
	}

	p = sampleREST()
	p.LogLevel = "NOPE"
	if _, err := schema.FromREST(p); err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestMarshalUnmarshalEvent(t *testing.T) {
	ev, err := schema.FromREST(sampleREST())
	if err != nil {
		t.Fatal(err)
	}
	raw, err := schema.MarshalEvent(ev)
	if err != nil {
		t.Fatal(err)
	}
	got, err := schema.UnmarshalEvent(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Message != ev.Message || got.Service != ev.Service || got.ID != ev.ID {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestRESTJSONCompatibility(t *testing.T) {
	const body = `{
	  "id": 30,
	  "metadata": {
	    "requestId": "2",
	    "clientIp": "36.75.63.226",
	    "userAgent": "Opera",
	    "requestMethod": "DELETE",
	    "requestPath": "/payments",
	    "responseStatus": 502,
	    "responseDuration": 194.67,
	    "extra": {"traceId": "t1"}
	  },
	  "timestamp": 1732974309000,
	  "logLevel": "ERROR",
	  "message": "upstream failed",
	  "source": {
	    "host": "exotic-effector.name",
	    "service": "payments",
	    "environment": "production"
	  }
	}`
	var p schema.RESTPayload
	if err := json.Unmarshal([]byte(body), &p); err != nil {
		t.Fatalf("unmarshal REST: %v", err)
	}
	ev, err := schema.FromREST(p)
	if err != nil {
		t.Fatalf("FromREST: %v", err)
	}
	if ev.LogLevel != schema.LevelError || ev.RequestMethod != "DELETE" {
		t.Fatalf("unexpected event: %+v", ev)
	}
}

func TestFromOTLPLike(t *testing.T) {
	ts := time.Date(2024, 12, 1, 12, 0, 0, 0, time.UTC)
	ev, err := schema.FromOTLPLike(ts, "ERROR", "boom",
		map[string]string{
			"service.name":           "checkout",
			"host.name":              "node-1",
			"deployment.environment": "staging",
		},
		map[string]string{
			"http.request.method":        "POST",
			"url.path":                   "/checkout",
			"http.response.status_code":  "500",
			"http.server.request.duration": "12.5",
			"traceId":                    "xyz",
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if ev.Service != "checkout" || ev.ResponseStatus != 500 {
		t.Fatalf("event=%+v", ev)
	}
	if ev.ResponseDurationMs != 12.5 {
		t.Fatalf("duration=%v", ev.ResponseDurationMs)
	}
	if ev.Attrs["traceId"] != "xyz" {
		t.Fatalf("attrs=%v", ev.Attrs)
	}
}

func TestDefaultTopic(t *testing.T) {
	if schema.DefaultTopic != "logs.ingestion.raw.v1" {
		t.Fatalf("topic=%q", schema.DefaultTopic)
	}
}
