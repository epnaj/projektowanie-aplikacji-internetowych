package web

import (
	"net/http"
	"strconv"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// pathID reads the named path segment as a core.ID.
func pathID(r *http.Request, name string) (core.ID, bool) {
	id, err := strconv.ParseUint(r.PathValue(name), 10, 64)
	if err != nil {
		return 0, false
	}
	return core.ID(id), true
}

// parseID reads the {id} path segment as a core.ID.
func parseID(r *http.Request) (core.ID, bool) {
	return pathID(r, "id")
}

func (h *Handler) listProjects(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	projects, err := h.projects.List(r.Context(), uid)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponses(projects))
}

func (h *Handler) createProject(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	uid, _ := userIDFromContext(r.Context())
	p, err := h.projects.Create(r.Context(), uid, req.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	w.Header().Set("Location", "/api/projects/"+strconv.FormatUint(uint64(p.Id), 10))
	writeJSON(w, http.StatusCreated, toProjectResponse(p))
}

func (h *Handler) getProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	uid, _ := userIDFromContext(r.Context())
	p, err := h.projects.Get(r.Context(), uid, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

func (h *Handler) updateProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	var req updateProjectRequest
	if err := decodeAndValidate(r, &req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	uid, _ := userIDFromContext(r.Context())
	p, err := h.projects.Rename(r.Context(), uid, id, req.Name)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

func (h *Handler) deleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusNotFound, http.StatusText(http.StatusNotFound))
		return
	}
	uid, _ := userIDFromContext(r.Context())
	if err := h.projects.Delete(r.Context(), uid, id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
