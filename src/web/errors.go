package web

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// statusFor is the single place where granular core errors collapse into HTTP
// status codes. ErrNotOwner deliberately maps to 404 (identical to
// ErrNotFound) so a caller cannot tell "exists but not yours" from "does not
// exist" and enumerate foreign IDs
func statusFor(err error) int {
	switch {
	case errors.Is(err, core.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, core.ErrNotOwner):
		return http.StatusNotFound
	case errors.Is(err, core.ErrLinkInactive):
		return http.StatusNotFound
	case errors.Is(err, core.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, core.ErrInvalidCredentials):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("encode response", "err", err)
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, errorResponse{Error: msg})
}

// writeServiceError maps a core error to its status. 500s are logged with the
// real cause but never leak it to the client.
func writeServiceError(w http.ResponseWriter, err error) {
	code := statusFor(err)
	if code == http.StatusInternalServerError {
		slog.Error("internal error", "err", err)
		writeError(w, code, "internal error")
		return
	}
	writeError(w, code, http.StatusText(code))
}
