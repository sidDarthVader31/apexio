package main

import (
	"encoding/json"
	"net/http"

	"github.com/sidDarthVader31/apexio/pkg/schema"
)

type ingestHandler struct {
	pub     *publisher
	metrics *ingestMetrics
}

func (h *ingestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var payload schema.RESTPayload
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&payload); err != nil {
		h.metrics.incRestErr()
		writeErr(w, http.StatusBadRequest, "invalid json", err)
		return
	}
	ev, err := schema.FromREST(payload)
	if err != nil {
		h.metrics.incRestErr()
		writeErr(w, http.StatusBadRequest, "validation failed", err)
		return
	}
	if err := h.pub.publishEvents(r.Context(), []schema.LogEvent{ev}); err != nil {
		h.metrics.incRestErr()
		writeErr(w, http.StatusBadGateway, "publish failed", err)
		return
	}
	h.metrics.incRestOK()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"id":      ev.ID,
		"service": ev.Service,
	})
}
