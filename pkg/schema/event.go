// Package schema defines the canonical log event used across Apexio services.
//
// Wire formats:
//   - REST ingest JSON (legacy /api/v1/log body) via [FromREST] / [ToREST]
//   - Internal broker payload (flat JSON of LogEvent) via [MarshalEvent] / [UnmarshalEvent]
//   - OTLP-like attribute bags via [FromOTLPLike] (full OTLP protobuf arrives in Phase 4)
package schema

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// DefaultTopic is the Redpanda/Kafka topic for raw ingested events.
const DefaultTopic = "logs.ingestion.raw.v1"

// Well-known log levels.
const (
	LevelDebug = "DEBUG"
	LevelInfo  = "INFO"
	LevelWarn  = "WARN"
	LevelError = "ERROR"
	LevelFatal = "FATAL"
)

// LogEvent is the canonical in-process representation of a log.
// Field names align with deploy/clickhouse/init/01_schema.sql (apexio.logs).
type LogEvent struct {
	Timestamp          time.Time         `json:"timestamp"`
	ID                 uint64            `json:"id"`
	LogLevel           string            `json:"log_level"`
	Message            string            `json:"message"`
	Service            string            `json:"service"`
	Host               string            `json:"host"`
	Environment        string            `json:"environment"`
	RequestID          string            `json:"request_id"`
	ClientIP           string            `json:"client_ip"`
	UserAgent          string            `json:"user_agent"`
	RequestMethod      string            `json:"request_method"`
	RequestPath        string            `json:"request_path"`
	ResponseStatus     uint16            `json:"response_status"`
	ResponseDurationMs float64           `json:"response_duration_ms"`
	Attrs              map[string]string `json:"attrs,omitempty"`
}

// RESTPayload is the public HTTP ingest body (compatible with the legacy API).
type RESTPayload struct {
	ID        uint64       `json:"id"`
	Timestamp uint64       `json:"timestamp"` // epoch millis
	LogLevel  string       `json:"logLevel"`
	Message   string       `json:"message"`
	Metadata  RESTMetadata `json:"metadata"`
	Source    RESTSource   `json:"source"`
}

// RESTMetadata is the nested metadata object on REST ingest.
type RESTMetadata struct {
	RequestID          string            `json:"requestId"`
	ClientIP           string            `json:"clientIp"`
	UserAgent          string            `json:"userAgent"`
	RequestMethod      string            `json:"requestMethod"`
	RequestPath        string            `json:"requestPath"`
	ResponseStatus     int               `json:"responseStatus"`
	ResponseDuration   float64           `json:"responseDuration"` // millis (legacy field name)
	Extra              map[string]string `json:"extra"`
}

// RESTSource is the nested source object on REST ingest.
type RESTSource struct {
	Host        string            `json:"host"`
	Service     string            `json:"service"`
	Environment string            `json:"environment"`
	Extra       map[string]string `json:"extra"`
}

var (
	ErrEmptyMessage = errors.New("message is required")
	ErrEmptyService = errors.New("source.service is required")
	ErrInvalidLevel = errors.New("logLevel is invalid")
)

// Validate checks required fields on a LogEvent.
func (e *LogEvent) Validate() error {
	if e == nil {
		return errors.New("event is nil")
	}
	if strings.TrimSpace(e.Message) == "" {
		return ErrEmptyMessage
	}
	if strings.TrimSpace(e.Service) == "" {
		return ErrEmptyService
	}
	if e.LogLevel != "" && !ValidLogLevel(e.LogLevel) {
		return fmt.Errorf("%w: %q", ErrInvalidLevel, e.LogLevel)
	}
	if e.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}
	return nil
}

// ValidLogLevel reports whether level is a known severity (case-insensitive).
func ValidLogLevel(level string) bool {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal:
		return true
	default:
		return false
	}
}

// NormalizeLogLevel uppercases a known level; empty becomes INFO.
func NormalizeLogLevel(level string) string {
	level = strings.ToUpper(strings.TrimSpace(level))
	if level == "" {
		return LevelInfo
	}
	return level
}

// FromREST maps a REST ingest payload into a LogEvent.
func FromREST(p RESTPayload) (LogEvent, error) {
	attrs := mergeAttrs(p.Source.Extra, p.Metadata.Extra)
	ts := time.UnixMilli(int64(p.Timestamp)).UTC()
	if p.Timestamp == 0 {
		ts = time.Now().UTC()
	}
	ev := LogEvent{
		Timestamp:          ts,
		ID:                 p.ID,
		LogLevel:           NormalizeLogLevel(p.LogLevel),
		Message:            p.Message,
		Service:            p.Source.Service,
		Host:               p.Source.Host,
		Environment:        p.Source.Environment,
		RequestID:          p.Metadata.RequestID,
		ClientIP:           p.Metadata.ClientIP,
		UserAgent:          p.Metadata.UserAgent,
		RequestMethod:      p.Metadata.RequestMethod,
		RequestPath:        p.Metadata.RequestPath,
		ResponseStatus:     uint16(p.Metadata.ResponseStatus),
		ResponseDurationMs: p.Metadata.ResponseDuration,
		Attrs:              attrs,
	}
	if err := ev.Validate(); err != nil {
		return LogEvent{}, err
	}
	return ev, nil
}

// ToREST maps a LogEvent back to the public REST shape.
func ToREST(e LogEvent) RESTPayload {
	extra := cloneMap(e.Attrs)
	return RESTPayload{
		ID:        e.ID,
		Timestamp: uint64(e.Timestamp.UTC().UnixMilli()),
		LogLevel:  e.LogLevel,
		Message:   e.Message,
		Metadata: RESTMetadata{
			RequestID:        e.RequestID,
			ClientIP:         e.ClientIP,
			UserAgent:        e.UserAgent,
			RequestMethod:    e.RequestMethod,
			RequestPath:      e.RequestPath,
			ResponseStatus:   int(e.ResponseStatus),
			ResponseDuration: e.ResponseDurationMs,
			Extra:            extra,
		},
		Source: RESTSource{
			Host:        e.Host,
			Service:     e.Service,
			Environment: e.Environment,
		},
	}
}

// FromOTLPLike maps OTLP-style resource/log attributes into a LogEvent.
// Full OTLP protobuf decode is Phase 4; this covers the attribute contract early.
func FromOTLPLike(ts time.Time, severity, body string, resourceAttrs, logAttrs map[string]string) (LogEvent, error) {
	attrs := mergeAttrs(resourceAttrs, logAttrs)
	ev := LogEvent{
		Timestamp:          ts.UTC(),
		LogLevel:           NormalizeLogLevel(severity),
		Message:            body,
		Service:            firstAttr(attrs, "service.name", "service"),
		Host:               firstAttr(attrs, "host.name", "host"),
		Environment:        firstAttr(attrs, "deployment.environment", "environment"),
		RequestID:          firstAttr(attrs, "http.request_id", "request_id", "requestId"),
		ClientIP:           firstAttr(attrs, "client.address", "client_ip", "clientIp"),
		UserAgent:          firstAttr(attrs, "user_agent.original", "user_agent", "userAgent"),
		RequestMethod:      firstAttr(attrs, "http.request.method", "request_method", "requestMethod"),
		RequestPath:        firstAttr(attrs, "url.path", "request_path", "requestPath"),
		ResponseStatus:     parseUint16Attr(attrs, "http.response.status_code", "response_status", "responseStatus"),
		ResponseDurationMs: parseFloatAttr(attrs, "http.server.request.duration", "response_duration_ms", "responseDuration"),
		Attrs:              attrs,
	}
	if ev.Timestamp.IsZero() {
		ev.Timestamp = time.Now().UTC()
	}
	if err := ev.Validate(); err != nil {
		return LogEvent{}, err
	}
	return ev, nil
}

// MarshalEvent encodes a LogEvent for the broker wire format.
func MarshalEvent(e LogEvent) ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(e)
}

// UnmarshalEvent decodes a broker payload into a LogEvent.
func UnmarshalEvent(data []byte) (LogEvent, error) {
	var e LogEvent
	if err := json.Unmarshal(data, &e); err != nil {
		return LogEvent{}, err
	}
	e.LogLevel = NormalizeLogLevel(e.LogLevel)
	if err := e.Validate(); err != nil {
		return LogEvent{}, err
	}
	return e, nil
}

func mergeAttrs(maps ...map[string]string) map[string]string {
	out := make(map[string]string)
	for _, m := range maps {
		for k, v := range m {
			if k == "" {
				continue
			}
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func cloneMap(m map[string]string) map[string]string {
	if len(m) == 0 {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func firstAttr(attrs map[string]string, keys ...string) string {
	for _, k := range keys {
		if v, ok := attrs[k]; ok && v != "" {
			return v
		}
	}
	return ""
}

func parseUint16Attr(attrs map[string]string, keys ...string) uint16 {
	for _, k := range keys {
		if v, ok := attrs[k]; ok && v != "" {
			var n uint64
			if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n <= 65535 {
				return uint16(n)
			}
		}
	}
	return 0
}

func parseFloatAttr(attrs map[string]string, keys ...string) float64 {
	for _, k := range keys {
		if v, ok := attrs[k]; ok && v != "" {
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				return f
			}
		}
	}
	return 0
}
