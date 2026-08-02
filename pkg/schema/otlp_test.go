package schema_test

import (
	"testing"
	"time"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

func TestLogEventsFromOTLP(t *testing.T) {
	ts := time.Date(2024, 12, 1, 12, 0, 0, 0, time.UTC)
	req := &collogsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsv1.ResourceLogs{{
			Resource: &resourcev1.Resource{
				Attributes: []*commonv1.KeyValue{
					kv("service.name", "checkout"),
					kv("host.name", "node-1"),
					kv("deployment.environment", "staging"),
				},
			},
			ScopeLogs: []*logsv1.ScopeLogs{{
				LogRecords: []*logsv1.LogRecord{{
					TimeUnixNano:   uint64(ts.UnixNano()),
					SeverityText:   "ERROR",
					Body:           &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "payment failed"}},
					Attributes: []*commonv1.KeyValue{
						kv("http.request.method", "POST"),
						kv("url.path", "/pay"),
						kv("http.response.status_code", "500"),
						kv("traceId", "abc"),
					},
				}},
			}},
		}},
	}

	events, err := schema.LogEventsFromOTLP(req)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events=%d", len(events))
	}
	ev := events[0]
	if ev.Service != "checkout" || ev.Message != "payment failed" || ev.LogLevel != schema.LevelError {
		t.Fatalf("event=%+v", ev)
	}
	if ev.ResponseStatus != 500 || ev.RequestMethod != "POST" {
		t.Fatalf("http fields=%+v", ev)
	}
	if ev.Attrs["traceId"] != "abc" {
		t.Fatalf("attrs=%v", ev.Attrs)
	}
}

func TestLogEventsFromOTLPRequiresService(t *testing.T) {
	req := &collogsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsv1.ResourceLogs{{
			ScopeLogs: []*logsv1.ScopeLogs{{
				LogRecords: []*logsv1.LogRecord{{
					Body: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: "no service"}},
				}},
			}},
		}},
	}
	_, err := schema.LogEventsFromOTLP(req)
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func kv(key, val string) *commonv1.KeyValue {
	return &commonv1.KeyValue{
		Key:   key,
		Value: &commonv1.AnyValue{Value: &commonv1.AnyValue_StringValue{StringValue: val}},
	}
}
