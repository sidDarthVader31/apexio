package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type writerMetrics struct {
	eventsWritten atomic.Uint64
	batchesFlush  atomic.Uint64
	decodeErrors  atomic.Uint64
	writeErrors   atomic.Uint64
}

func (m *writerMetrics) incEventsWritten(n uint64) { m.eventsWritten.Add(n) }
func (m *writerMetrics) incBatchesFlushed()      { m.batchesFlush.Add(1) }
func (m *writerMetrics) incDecodeErrors()        { m.decodeErrors.Add(1) }
func (m *writerMetrics) incWriteErrors()         { m.writeErrors.Add(1) }

func (m *writerMetrics) handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		fmt.Fprintf(w, "# HELP apexio_writer_events_written_total Events persisted to ClickHouse\n")
		fmt.Fprintf(w, "# TYPE apexio_writer_events_written_total counter\n")
		fmt.Fprintf(w, "apexio_writer_events_written_total %d\n", m.eventsWritten.Load())
		fmt.Fprintf(w, "# HELP apexio_writer_batches_flushed_total ClickHouse batch flushes\n")
		fmt.Fprintf(w, "# TYPE apexio_writer_batches_flushed_total counter\n")
		fmt.Fprintf(w, "apexio_writer_batches_flushed_total %d\n", m.batchesFlush.Load())
		fmt.Fprintf(w, "# HELP apexio_writer_decode_errors_total Failed event decodes\n")
		fmt.Fprintf(w, "# TYPE apexio_writer_decode_errors_total counter\n")
		fmt.Fprintf(w, "apexio_writer_decode_errors_total %d\n", m.decodeErrors.Load())
		fmt.Fprintf(w, "# HELP apexio_writer_write_errors_total Failed batch writes (after retries)\n")
		fmt.Fprintf(w, "# TYPE apexio_writer_write_errors_total counter\n")
		fmt.Fprintf(w, "apexio_writer_write_errors_total %d\n", m.writeErrors.Load())
	}
}
