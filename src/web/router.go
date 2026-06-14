package web

import (
	"net/http"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/auth"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// Handler holds the dependencies every HTTP handler needs. Handlers stay thin:
// decode -> validate -> call a service -> translate -> write
type Handler struct {
	users    *core.UserService
	projects *core.ProjectService
	sessions *auth.SessionManager
}

func NewHandler(users *core.UserService, projects *core.ProjectService, sessions *auth.SessionManager) *Handler {
	return &Handler{users: users, projects: projects, sessions: sessions}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Public auth
	mux.HandleFunc("POST /api/register", h.register)
	mux.HandleFunc("POST /api/login", h.login)

	// Authenticated session
	mux.HandleFunc("POST /api/logout", h.requireAuth(h.logout))
	mux.HandleFunc("GET /api/me", h.requireAuth(h.me))

	// Projects (all owner-scoped)
	mux.HandleFunc("GET /api/projects", h.requireAuth(h.listProjects))
	mux.HandleFunc("POST /api/projects", h.requireAuth(h.createProject))
	mux.HandleFunc("GET /api/projects/{id}", h.requireAuth(h.getProject))
	mux.HandleFunc("PATCH /api/projects/{id}", h.requireAuth(h.updateProject))
	mux.HandleFunc("DELETE /api/projects/{id}", h.requireAuth(h.deleteProject))

	return requestLogger(mux)
}
