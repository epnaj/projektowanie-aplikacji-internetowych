package web

import (
	"net/http"
	"time"
)

// timeWindow reads optional from/to query params as RFC3339 timestamps. A
// missing param is the zero time, which the store treats as unbounded.
func timeWindow(r *http.Request) (from, to time.Time, err error) {
	if v := r.URL.Query().Get("from"); v != "" {
		from, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return from, to, err
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		to, err = time.Parse(time.RFC3339, v)
		if err != nil {
			return from, to, err
		}
	}
	return from, to, nil
}

func (h *Handler) projectStats(w http.ResponseWriter, r *http.Request) {
	projectId, ok := pathID(r, "projectId")
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	from, to, err := timeWindow(r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "from/to must be RFC3339 timestamps")
		return
	}

	uid, _ := userIDFromContext(r.Context())
	stats, err := h.stats.ListByProject(r.Context(), uid, projectId, from, to)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStatisticResponses(stats))
}

func (h *Handler) linkStats(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	from, to, err := timeWindow(r)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "from/to must be RFC3339 timestamps")
		return
	}

	uid, _ := userIDFromContext(r.Context())
	stats, err := h.stats.ListByLink(r.Context(), uid, id, from, to)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toStatisticResponses(stats))
}
