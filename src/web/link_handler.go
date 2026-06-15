package web

import (
	"net/http"
	"strconv"
)

func (h *Handler) listLinks(w http.ResponseWriter, r *http.Request) {
	projectId, ok := pathID(r, "projectId")
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	uid, _ := userIDFromContext(r.Context())
	links, err := h.links.ListByProject(r.Context(), uid, projectId)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toLinkResponses(links))
}

func (h *Handler) createLink(w http.ResponseWriter, r *http.Request) {
	projectId, ok := pathID(r, "projectId")
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	var req createLinkRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	uid, _ := userIDFromContext(r.Context())
	l, err := h.links.Create(r.Context(), uid, projectId, req.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Location", "/api/links/"+strconv.FormatUint(uint64(l.Id), 10))
	writeJSON(w, http.StatusCreated, toLinkResponse(l))
}

func (h *Handler) getLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	uid, _ := userIDFromContext(r.Context())
	l, err := h.links.Get(r.Context(), uid, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toLinkResponse(l))
}

func (h *Handler) updateLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	var req updateLinkRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	uid, _ := userIDFromContext(r.Context())
	l, err := h.links.Update(r.Context(), uid, id, req.Name, req.Active)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toLinkResponse(l))
}

func (h *Handler) deleteLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	uid, _ := userIDFromContext(r.Context())
	if err := h.links.Delete(r.Context(), uid, id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
