package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type ingestMetrics struct {
	restOK    atomic.Uint64
	restErr   atomic.Uint64
	otlpOK    atomic.Uint64
	otlpErr   atomic.Uint64
	otlpLogs  atomic.Uint64
}

func (m *ingestMetrics) incRestOK()    { m.restOK.Add(1) }
func (m *ingestMetrics) incRestErr()   { m.restErr.Add(1) }
func (m *ingestMetrics) incOTLPOK(n uint64) {
	m.otlpOK.Add(1)
	m.otlpLogs.Add(n)
}
func (m *ingestMetrics) incOTLPErr() { m.otlpErr.Add(1) }

func (m *ingestMetrics) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP apexio_ingest_requests_total Ingest requests by protocol and outcome\n")
		fmt.Fprintf(w, "# TYPE apexio_ingest_requests_total counter\n")
		fmt.Fprintf(w, "apexio_ingest_requests_total{protocol=\"rest\",result=\"ok\"} %d\n", m.restOK.Load())
		fmt.Fprintf(w, "apexio_ingest_requests_total{protocol=\"rest\",result=\"error\"} %d\n", m.restErr.Load())
		fmt.Fprintf(w, "apexio_ingest_requests_total{protocol=\"otlp\",result=\"ok\"} %d\n", m.otlpOK.Load())
		fmt.Fprintf(w, "apexio_ingest_requests_total{protocol=\"otlp\",result=\"error\"} %d\n", m.otlpErr.Load())
		fmt.Fprintf(w, "# HELP apexio_ingest_logs_total OTLP log records accepted\n")
		fmt.Fprintf(w, "# TYPE apexio_ingest_logs_total counter\n")
		fmt.Fprintf(w, "apexio_ingest_logs_total %d\n", m.otlpLogs.Load())
	}
}
