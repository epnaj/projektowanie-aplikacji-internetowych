package web_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/auth"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/store/memory"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/web"
)

type env struct {
	srv  *httptest.Server
	stat *core.StatisticService
}

// newEnv builds a full HTTP server over the in-memory store, exercising the
// real router, middleware, handlers and session auth end to end.
func newEnv(t *testing.T, ready func(context.Context) error) *env {
	t.Helper()
	store := memory.New()
	sessions := auth.NewSessionManager([]byte("test-session-key-0123456789abcdef"))
	stat := core.NewStatisticService(store.Statistics(), store.Links(), store.Projects(), store.Buffer())
	h := web.NewHandler(
		core.NewUserService(store.Users(), auth.NewBcryptHasher()),
		core.NewProjectService(store.Projects()),
		core.NewLinkService(store.Links(), store.Projects()),
		stat,
		sessions,
		ready,
	)
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)
	return &env{srv: srv, stat: stat}
}

func (e *env) client(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	// Do not auto-follow redirects; we assert on status codes directly.
	return &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}
}

func (e *env) doJSON(t *testing.T, c *http.Client, method, path string, body any) *http.Response {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, e.srv.URL+path, rdr)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func (e *env) doForm(t *testing.T, c *http.Client, method, path string, form url.Values) *http.Response {
	t.Helper()
	req, err := http.NewRequest(method, e.srv.URL+path, strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	return resp
}

func decode[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	defer resp.Body.Close()
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		t.Fatalf("decode: %v", err)
	}
	return v
}

func bodyString(t *testing.T, resp *http.Response) string {
	t.Helper()
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return string(b)
}

func registerAndLogin(t *testing.T, e *env, email string) *http.Client {
	t.Helper()
	c := e.client(t)
	creds := map[string]string{"email": email, "password": "password123"}
	if resp := e.doJSON(t, c, http.MethodPost, "/api/register", creds); resp.StatusCode != http.StatusCreated {
		t.Fatalf("register %s: want 201, got %d", email, resp.StatusCode)
	}
	if resp := e.doJSON(t, c, http.MethodPost, "/api/login", creds); resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s: want 200, got %d", email, resp.StatusCode)
	}
	return c
}

func TestHealthAndReadiness(t *testing.T) {
	e := newEnv(t, nil)
	c := e.client(t)

	if resp := e.doJSON(t, c, http.MethodGet, "/healthz", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("healthz: want 200, got %d", resp.StatusCode)
	}
	if resp := e.doJSON(t, c, http.MethodGet, "/readyz", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("readyz (nil check): want 200, got %d", resp.StatusCode)
	}

	// A failing readiness check must surface 503.
	down := newEnv(t, func(context.Context) error { return errors.New("db down") })
	if resp := down.doJSON(t, down.client(t), http.MethodGet, "/readyz", nil); resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("readyz (failing check): want 503, got %d", resp.StatusCode)
	}
}

func TestAPIFlowAndOwnershipCollapse(t *testing.T) {
	e := newEnv(t, nil)

	alice := registerAndLogin(t, e, "alice@example.com")
	project := decode[map[string]any](t, e.doJSON(t, alice, http.MethodPost, "/api/projects", map[string]string{"name": "Campaign"}))
	pid := strconv.Itoa(int(project["id"].(float64)))

	if resp := e.doJSON(t, alice, http.MethodGet, "/api/projects/"+pid, nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("owner get: want 200, got %d", resp.StatusCode)
	}

	// Unauthenticated access is rejected.
	anon := e.client(t)
	if resp := e.doJSON(t, anon, http.MethodGet, "/api/me", nil); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("anon /api/me: want 401, got %d", resp.StatusCode)
	}

	// Bob sees Alice's project as 404, not 403 (no foreign-ID enumeration).
	bob := registerAndLogin(t, e, "bob@example.com")
	if resp := e.doJSON(t, bob, http.MethodGet, "/api/projects/"+pid, nil); resp.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign get: want 404, got %d", resp.StatusCode)
	}

	// Duplicate email -> 409, bad email -> 422.
	if resp := e.doJSON(t, e.client(t), http.MethodPost, "/api/register", map[string]string{"email": "alice@example.com", "password": "password123"}); resp.StatusCode != http.StatusConflict {
		t.Fatalf("duplicate register: want 409, got %d", resp.StatusCode)
	}
	if resp := e.doJSON(t, e.client(t), http.MethodPost, "/api/register", map[string]string{"email": "not-an-email", "password": "password123"}); resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("bad email: want 422, got %d", resp.StatusCode)
	}
}

func TestPixelAlwaysSucceedsAndStatsFlush(t *testing.T) {
	e := newEnv(t, nil)
	alice := registerAndLogin(t, e, "alice@example.com")
	project := decode[map[string]any](t, e.doJSON(t, alice, http.MethodPost, "/api/projects", map[string]string{"name": "Campaign"}))
	pid := strconv.Itoa(int(project["id"].(float64)))
	link := decode[map[string]any](t, e.doJSON(t, alice, http.MethodPost, "/api/projects/"+pid+"/links", map[string]string{"name": "Newsletter"}))
	hash := link["linkHash"].(string)
	lid := strconv.Itoa(int(link["id"].(float64)))

	anon := e.client(t)
	for i := 0; i < 2; i++ {
		resp := e.doJSON(t, anon, http.MethodGet, "/pixel/"+hash+".gif", nil)
		if resp.StatusCode != http.StatusOK || resp.Header.Get("Content-Type") != "image/gif" {
			t.Fatalf("pixel: want 200 image/gif, got %d %q", resp.StatusCode, resp.Header.Get("Content-Type"))
		}
		resp.Body.Close()
	}
	// Unknown hash still returns a 200 pixel (never leaks, never breaks the page).
	if resp := e.doJSON(t, anon, http.MethodGet, "/pixel/deadbeef.gif", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("unknown pixel: want 200, got %d", resp.StatusCode)
	}

	if err := e.stat.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	stats := decode[[]map[string]any](t, e.doJSON(t, alice, http.MethodGet, "/api/links/"+lid+"/stats", nil))
	if len(stats) != 1 || int(stats[0]["hits"].(float64)) != 2 {
		t.Fatalf("stats: want one bucket of 2 hits, got %+v", stats)
	}
}
