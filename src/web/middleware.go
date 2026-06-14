package web

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

type contextKey string

const userIDContextKey contextKey = "userID"

func userIDFromContext(ctx context.Context) (core.ID, bool) {
	id, ok := ctx.Value(userIDContextKey).(core.ID)
	return id, ok
}

// requireAuth rejects requests without a valid session (401) and, on success,
// stashes the user's ID in the request context for handlers to read
func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := h.sessions.UserID(r)
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		ctx := context.WithValue(r.Context(), userIDContextKey, id)
		next(w, r.WithContext(ctx))
	}
}

// statusRecorder captures the response code for structured access logging
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

// requestLogger emits one structured slog line per request with method, path,
// status, and duration
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
		)
	})
}
