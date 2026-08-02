// Sample client for Apexio Phase 4: sends REST and/or OTLP logs to the gateway.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcev1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

func main() {
	gateway := flag.String("gateway", envOr("APEXIO_GATEWAY", "http://127.0.0.1:18080"), "gateway base URL")
	mode := flag.String("mode", "both", "send mode: rest, otlp, or both")
	message := flag.String("message", "", "log message (default includes timestamp)")
	service := flag.String("service", "sample-client", "service name")
	flag.Parse()

	msg := *message
	if msg == "" {
		msg = fmt.Sprintf("sample-client-%d", time.Now().Unix())
	}

	var err error
	switch *mode {
	case "rest":
		err = sendREST(*gateway, msg, *service)
	case "otlp":
		err = sendOTLP(*gateway, msg, *service)
	case "both":
		err = sendREST(*gateway, msg+"-rest", *service)
		if err == nil {
			err = sendOTLP(*gateway, msg+"-otlp", *service)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("sent via %s to %s message=%q service=%s\n", *mode, *gateway, msg, *service)
}

func sendREST(gateway, message, service string) error {
	payload := map[string]any{
		"id":        time.Now().UnixNano(),
		"timestamp": time.Now().UnixMilli(),
		"logLevel":  "INFO",
		"message":   message,
		"metadata": map[string]any{
			"requestId":        "sample-req",
			"requestMethod":    "POST",
			"requestPath":      "/sample",
			"responseStatus":   200,
			"responseDuration": 5.5,
		},
		"source": map[string]any{
			"host":        "localhost",
			"service":     service,
			"environment": "dev",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := http.Post(gateway+"/api/v1/log", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("rest status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func sendOTLP(gateway, message, service string) error {
	req := &collogsv1.ExportLogsServiceRequest{
		ResourceLogs: []*logsv1.ResourceLogs{{
			Resource: &resourcev1.Resource{
				Attributes: []*commonv1.KeyValue{{
					Key: "service.name",
					Value: &commonv1.AnyValue{
						Value: &commonv1.AnyValue_StringValue{StringValue: service},
					},
				}, {
					Key: "deployment.environment",
					Value: &commonv1.AnyValue{
						Value: &commonv1.AnyValue_StringValue{StringValue: "dev"},
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
					Attributes: []*commonv1.KeyValue{{
						Key: "http.request.method",
						Value: &commonv1.AnyValue{
							Value: &commonv1.AnyValue_StringValue{StringValue: "POST"},
						},
					}},
				}},
			}},
		}},
	}
	body, err := proto.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequest(http.MethodPost, gateway+"/v1/logs", bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("otlp status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
