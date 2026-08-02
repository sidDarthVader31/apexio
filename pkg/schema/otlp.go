package schema

import (
	"fmt"
	"strconv"
	"time"

	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
)

// LogEventsFromOTLP converts an OTLP ExportLogsServiceRequest into canonical events.
func LogEventsFromOTLP(req *collogsv1.ExportLogsServiceRequest) ([]LogEvent, error) {
	if req == nil {
		return nil, fmt.Errorf("otlp: request is nil")
	}
	var events []LogEvent
	for _, rl := range req.ResourceLogs {
		resourceAttrs := keyValuesToMap(rl.GetResource().GetAttributes())
		for _, sl := range rl.GetScopeLogs() {
			for _, lr := range sl.GetLogRecords() {
				logAttrs := keyValuesToMap(lr.GetAttributes())
				ts := time.Unix(0, int64(lr.GetTimeUnixNano())).UTC()
				if lr.GetTimeUnixNano() == 0 {
					ts = time.Now().UTC()
				}
				severity := lr.GetSeverityText()
				if severity == "" {
					severity = severityNumberToText(lr.GetSeverityNumber())
				}
				ev, err := FromOTLPLike(ts, severity, anyValueToString(lr.GetBody()), resourceAttrs, logAttrs)
				if err != nil {
					return nil, fmt.Errorf("otlp: log record: %w", err)
				}
				events = append(events, ev)
			}
		}
	}
	return events, nil
}

func keyValuesToMap(kvs []*commonv1.KeyValue) map[string]string {
	if len(kvs) == 0 {
		return nil
	}
	out := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		if kv == nil || kv.Key == "" {
			continue
		}
		out[kv.Key] = anyValueToString(kv.Value)
	}
	return out
}

func anyValueToString(v *commonv1.AnyValue) string {
	if v == nil {
		return ""
	}
	switch val := v.Value.(type) {
	case *commonv1.AnyValue_StringValue:
		return val.StringValue
	case *commonv1.AnyValue_BoolValue:
		return strconv.FormatBool(val.BoolValue)
	case *commonv1.AnyValue_IntValue:
		return strconv.FormatInt(val.IntValue, 10)
	case *commonv1.AnyValue_DoubleValue:
		return strconv.FormatFloat(val.DoubleValue, 'f', -1, 64)
	case *commonv1.AnyValue_BytesValue:
		return string(val.BytesValue)
	case *commonv1.AnyValue_ArrayValue:
		return fmt.Sprint(val.ArrayValue)
	case *commonv1.AnyValue_KvlistValue:
		return fmt.Sprint(val.KvlistValue)
	default:
		return ""
	}
}

func severityNumberToText(n logsv1.SeverityNumber) string {
	switch {
	case n >= logsv1.SeverityNumber_SEVERITY_NUMBER_FATAL:
		return LevelFatal
	case n >= logsv1.SeverityNumber_SEVERITY_NUMBER_ERROR:
		return LevelError
	case n >= logsv1.SeverityNumber_SEVERITY_NUMBER_WARN:
		return LevelWarn
	case n >= logsv1.SeverityNumber_SEVERITY_NUMBER_INFO:
		return LevelInfo
	case n >= logsv1.SeverityNumber_SEVERITY_NUMBER_DEBUG:
		return LevelDebug
	default:
		return LevelInfo
	}
}
