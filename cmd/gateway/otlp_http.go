package main

import (
	"io"
	"net/http"
	"strings"

	collogsv1 "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

type otlpHTTPHandler struct {
	pub     *publisher
	metrics *ingestMetrics
}

func (h *otlpHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		h.metrics.incOTLPErr()
		writeErr(w, http.StatusBadRequest, "read body failed", err)
		return
	}

	var req collogsv1.ExportLogsServiceRequest
	ct := strings.ToLower(r.Header.Get("Content-Type"))
	switch {
	case strings.Contains(ct, "json"):
		if err := protojson.Unmarshal(body, &req); err != nil {
			h.metrics.incOTLPErr()
			writeErr(w, http.StatusBadRequest, "invalid otlp json", err)
			return
		}
	default:
		if err := proto.Unmarshal(body, &req); err != nil {
			h.metrics.incOTLPErr()
			writeErr(w, http.StatusBadRequest, "invalid otlp protobuf", err)
			return
		}
	}

	events, err := schema.LogEventsFromOTLP(&req)
	if err != nil {
		h.metrics.incOTLPErr()
		writeErr(w, http.StatusBadRequest, "otlp mapping failed", err)
		return
	}
	if len(events) == 0 {
		h.metrics.incOTLPOK(0)
		w.WriteHeader(http.StatusOK)
		return
	}
	if err := h.pub.publishEvents(r.Context(), events); err != nil {
		h.metrics.incOTLPErr()
		writeErr(w, http.StatusBadGateway, "publish failed", err)
		return
	}
	h.metrics.incOTLPOK(uint64(len(events)))
	w.WriteHeader(http.StatusOK)
}
