package web

import (
	"context"
	"net/http"
	"time"
)

// healthz is a liveness probe: if the process can serve this, it is alive.
func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// readyz is a readiness probe: it reports 503 until backing dependencies (the
// database) answer. With no check configured (in-memory mode) it is always ready.
func (h *Handler) readyz(w http.ResponseWriter, r *http.Request) {
	if h.ready != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := h.ready(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}
