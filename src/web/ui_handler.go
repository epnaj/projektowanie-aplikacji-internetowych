package web

import (
	"net/http"
)

// projectPageData is the view model for the project detail page.
type projectPageData struct {
	Project projectResponse
	Links   []linkResponse
}

// index sends visitors to the app, which in turn bounces to /login when there
// is no session.
func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/app", http.StatusSeeOther)
}

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.UserID(r); ok {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	render(w, http.StatusOK, "page_login", nil)
}

func (h *Handler) registerPage(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.sessions.UserID(r); ok {
		http.Redirect(w, r, "/app", http.StatusSeeOther)
		return
	}
	render(w, http.StatusOK, "page_register", nil)
}

func (h *Handler) loginForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		hxError(w, "invalid form")
		return
	}
	user, err := h.users.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		hxError(w, "invalid email or password")
		return
	}
	if err := h.sessions.Save(w, r, user.Id); err != nil {
		hxError(w, "could not start session")
		return
	}
	w.Header().Set("HX-Redirect", "/app")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) registerForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		hxError(w, "invalid form")
		return
	}
	req := registerRequest{Email: r.FormValue("email"), Password: r.FormValue("password")}
	if err := req.Validate(); err != nil {
		hxError(w, err.Error())
		return
	}
	user, err := h.users.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		hxError(w, "could not register (email may already be taken)")
		return
	}
	if err := h.sessions.Save(w, r, user.Id); err != nil {
		hxError(w, "could not start session")
		return
	}
	w.Header().Set("HX-Redirect", "/app")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) logoutForm(w http.ResponseWriter, r *http.Request) {
	_ = h.sessions.Clear(w, r)
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	uid, _ := userIDFromContext(r.Context())
	projects, err := h.projects.List(r.Context(), uid)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	render(w, http.StatusOK, "page_dashboard", struct{ Projects []projectResponse }{toProjectResponses(projects)})
}

func (h *Handler) uiCreateProject(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		hxError(w, "invalid form")
		return
	}
	req := createProjectRequest{Name: r.FormValue("name")}
	if err := req.Validate(); err != nil {
		hxError(w, err.Error())
		return
	}
	uid, _ := userIDFromContext(r.Context())
	p, err := h.projects.Create(r.Context(), uid, req.Name)
	if err != nil {
		hxError(w, "could not create project")
		return
	}
	render(w, http.StatusOK, "project_row", toProjectResponse(p))
}

func (h *Handler) uiDeleteProject(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	uid, _ := userIDFromContext(r.Context())
	if err := h.projects.Delete(r.Context(), uid, id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK) // empty body; htmx removes the row
}

func (h *Handler) projectPage(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	uid, _ := userIDFromContext(r.Context())
	p, err := h.projects.Get(r.Context(), uid, id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	links, err := h.links.ListByProject(r.Context(), uid, id)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	render(w, http.StatusOK, "page_project", projectPageData{
		Project: toProjectResponse(p),
		Links:   toLinkResponses(links),
	})
}

func (h *Handler) uiCreateLink(w http.ResponseWriter, r *http.Request) {
	projectId, ok := pathID(r, "projectId")
	if !ok {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		hxError(w, "invalid form")
		return
	}
	req := createLinkRequest{Name: r.FormValue("name")}
	if err := req.Validate(); err != nil {
		hxError(w, err.Error())
		return
	}
	uid, _ := userIDFromContext(r.Context())
	l, err := h.links.Create(r.Context(), uid, projectId, req.Name)
	if err != nil {
		hxError(w, "could not create link")
		return
	}
	render(w, http.StatusOK, "link_row", toLinkResponse(l))
}

func (h *Handler) uiToggleLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	uid, _ := userIDFromContext(r.Context())
	current, err := h.links.Get(r.Context(), uid, id)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	flipped := !current.Active
	l, err := h.links.Update(r.Context(), uid, id, nil, &flipped)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	render(w, http.StatusOK, "link_row", toLinkResponse(l))
}

func (h *Handler) uiDeleteLink(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	uid, _ := userIDFromContext(r.Context())
	if err := h.links.Delete(r.Context(), uid, id); err != nil {
		writeServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) uiProjectStats(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	from, to, err := timeWindow(r)
	if err != nil {
		http.Error(w, "bad time window", http.StatusUnprocessableEntity)
		return
	}
	uid, _ := userIDFromContext(r.Context())
	stats, err := h.stats.ListByProject(r.Context(), uid, id, from, to)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	render(w, http.StatusOK, "stats_table", toStatisticResponses(stats))
}
