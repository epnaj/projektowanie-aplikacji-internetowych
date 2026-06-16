package web

import (
	"context"
	"net/http"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/auth"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// Handler holds the dependencies every HTTP handler needs. Handlers stay thin:
// decode -> validate -> call a service -> translate -> write
type Handler struct {
	users    *core.UserService
	projects *core.ProjectService
	links    *core.LinkService
	stats    *core.StatisticService
	sessions *auth.SessionManager
	// ready reports whether backing dependencies are reachable, for /readyz.
	// nil means "always ready" (in-memory mode).
	ready func(context.Context) error
}

func NewHandler(
	users *core.UserService,
	projects *core.ProjectService,
	links *core.LinkService,
	stats *core.StatisticService,
	sessions *auth.SessionManager,
	ready func(context.Context) error,
) *Handler {
	return &Handler{
		users:    users,
		projects: projects,
		links:    links,
		stats:    stats,
		sessions: sessions,
		ready:    ready,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Health probes (unauthenticated)
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET /readyz", h.readyz)

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

	// Links: collection routes nest under the project, item routes are top-level
	mux.HandleFunc("GET /api/projects/{projectId}/links", h.requireAuth(h.listLinks))
	mux.HandleFunc("POST /api/projects/{projectId}/links", h.requireAuth(h.createLink))
	mux.HandleFunc("GET /api/links/{id}", h.requireAuth(h.getLink))
	mux.HandleFunc("PATCH /api/links/{id}", h.requireAuth(h.updateLink))
	mux.HandleFunc("DELETE /api/links/{id}", h.requireAuth(h.deleteLink))

	// Statistics (read-only, owner-scoped)
	mux.HandleFunc("GET /api/projects/{projectId}/stats", h.requireAuth(h.projectStats))
	mux.HandleFunc("GET /api/links/{id}/stats", h.requireAuth(h.linkStats))

	// Public tracking pixel, embedded on third-party pages so it carries no auth
	mux.HandleFunc("GET /pixel/{hash}", h.trackPixel)

	// Static assets (htmx)
	mux.Handle("GET /static/", http.FileServerFS(staticFS))

	// HTML UI (htmx). Talks to the same core services as the JSON API above
	mux.HandleFunc("GET /{$}", h.index)
	mux.HandleFunc("GET /login", h.loginPage)
	mux.HandleFunc("POST /login", h.loginForm)
	mux.HandleFunc("GET /register", h.registerPage)
	mux.HandleFunc("POST /register", h.registerForm)
	mux.HandleFunc("POST /logout", h.logoutForm)

	mux.HandleFunc("GET /app", h.requireAuthPage(h.dashboard))
	mux.HandleFunc("POST /app/projects", h.requireAuthPage(h.uiCreateProject))
	mux.HandleFunc("GET /app/projects/{id}", h.requireAuthPage(h.projectPage))
	mux.HandleFunc("DELETE /app/projects/{id}", h.requireAuthPage(h.uiDeleteProject))
	mux.HandleFunc("GET /app/projects/{id}/stats", h.requireAuthPage(h.uiProjectStats))
	mux.HandleFunc("POST /app/projects/{projectId}/links", h.requireAuthPage(h.uiCreateLink))
	mux.HandleFunc("POST /app/links/{id}/toggle", h.requireAuthPage(h.uiToggleLink))
	mux.HandleFunc("DELETE /app/links/{id}", h.requireAuthPage(h.uiDeleteLink))

	return requestLogger(mux)
}
