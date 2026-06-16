package web_test

import (
	"context"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

var (
	reProjectID = regexp.MustCompile(`project-(\d+)`)
	rePixelHash = regexp.MustCompile(`/pixel/([0-9a-f]+)\.gif`)
)

// TestHTMXEndToEnd walks the whole HTML UI the way a browser would: register
// via the form, land on the dashboard, create a project then a link (each a
// returned htmx fragment), fire the tracking pixel, and confirm the stats panel
// renders the hit after a flush.
func TestHTMXEndToEnd(t *testing.T) {
	e := newEnv(t, nil)
	c := e.client(t)

	form := url.Values{"email": {"alice@example.com"}, "password": {"password123"}}
	if resp := e.doForm(t, c, http.MethodPost, "/register", form); resp.StatusCode != http.StatusNoContent || resp.Header.Get("HX-Redirect") != "/app" {
		t.Fatalf("register form: want 204 + HX-Redirect /app, got %d %q", resp.StatusCode, resp.Header.Get("HX-Redirect"))
	}

	dash := bodyString(t, e.doJSON(t, c, http.MethodGet, "/app", nil))
	if !strings.Contains(dash, `id="projects"`) {
		t.Fatalf("dashboard missing projects list:\n%s", dash)
	}

	// Create a project; the response is just the new <li> fragment.
	row := bodyString(t, e.doForm(t, c, http.MethodPost, "/app/projects", url.Values{"name": {"Demo Campaign"}}))
	if !strings.Contains(row, "Demo Campaign") {
		t.Fatalf("project fragment missing name:\n%s", row)
	}
	m := reProjectID.FindStringSubmatch(row)
	if m == nil {
		t.Fatalf("could not find project id in:\n%s", row)
	}
	pid := m[1]

	page := bodyString(t, e.doJSON(t, c, http.MethodGet, "/app/projects/"+pid, nil))
	if !strings.Contains(page, "Links") {
		t.Fatalf("project page missing Links section:\n%s", page)
	}

	// Create a link; fragment carries the embeddable pixel snippet.
	linkRow := bodyString(t, e.doForm(t, c, http.MethodPost, "/app/projects/"+pid+"/links", url.Values{"name": {"Hero Banner"}}))
	hm := rePixelHash.FindStringSubmatch(linkRow)
	if hm == nil {
		t.Fatalf("link fragment missing pixel snippet:\n%s", linkRow)
	}
	hash := hm[1]

	if resp := e.doJSON(t, e.client(t), http.MethodGet, "/pixel/"+hash+".gif", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("pixel: want 200, got %d", resp.StatusCode)
	}

	if err := e.stat.Flush(context.Background()); err != nil {
		t.Fatalf("flush: %v", err)
	}

	stats := bodyString(t, e.doJSON(t, c, http.MethodGet, "/app/projects/"+pid+"/stats", nil))
	if strings.Contains(stats, "No hits recorded") || !strings.Contains(stats, "<td>1</td>") {
		t.Fatalf("stats fragment should show one hit:\n%s", stats)
	}
}

// TestAppRequiresAuth confirms the page guard redirects anonymous visitors.
func TestAppRequiresAuth(t *testing.T) {
	e := newEnv(t, nil)
	resp := e.doJSON(t, e.client(t), http.MethodGet, "/app", nil)
	if resp.StatusCode != http.StatusSeeOther || resp.Header.Get("Location") != "/login" {
		t.Fatalf("anon /app: want 303 -> /login, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
}
