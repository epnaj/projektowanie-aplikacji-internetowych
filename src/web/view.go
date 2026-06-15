package web

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
	"time"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

// templates holds every page and fragment, parsed once at startup. Handlers
// render a named template (a full page or an htmx fragment) by name.
var templates = template.Must(template.New("").Funcs(template.FuncMap{
	"fmtTime": func(t time.Time) string { return t.UTC().Format("2006-01-02 15:04") },
}).ParseFS(templatesFS, "templates/*.html"))

func render(w http.ResponseWriter, status int, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := templates.ExecuteTemplate(w, name, data); err != nil {
		slog.Error("render template", "name", name, "err", err)
	}
}

// hxError reports a form error back to htmx: it retargets the swap onto the
// page's #form-error element and writes the message there. The 200 status keeps
// htmx swapping (it ignores 4xx bodies by default).
func hxError(w http.ResponseWriter, msg string) {
	w.Header().Set("HX-Retarget", "#form-error")
	w.Header().Set("HX-Reswap", "innerHTML")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(template.HTMLEscapeString(msg)))
}
