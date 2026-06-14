package main

import (
	"crypto/rand"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/auth"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/store/memory"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/web"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	addr := envOr("ADDR", ":8080")

	// Adapters (swap memory for Postgres/Redis here later).
	store := memory.New()
	hasher := auth.NewBcryptHasher()
	sessions := auth.NewSessionManager(sessionKey())

	// Core services.
	userService := core.NewUserService(store.Users(), hasher)
	projectService := core.NewProjectService(store.Projects())

	handler := web.NewHandler(userService, projectService, sessions)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("server starting", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("server failed", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// sessionKey reads SESSION_KEY or, in development, generates an ephemeral key.
// A generated key invalidates all sessions on restart, which is fine locally
// but must be set from a stable secret in production
func sessionKey() []byte {
	if v := os.Getenv("SESSION_KEY"); v != "" {
		return []byte(v)
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		slog.Error("generate session key", "err", err)
		os.Exit(1)
	}
	slog.Warn("SESSION_KEY not set; using an ephemeral key (sessions drop on restart)")
	return key
}
