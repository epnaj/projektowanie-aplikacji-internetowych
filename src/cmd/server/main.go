package main

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/auth"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/store/memory"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/store/postgres"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/web"
)

const flushInterval = 5 * time.Second

// repoProvider is the subset of a backing store the composition root needs.
// Both the in-memory store and the Postgres store satisfy it, so the storage
// engine is a single decision made here and nowhere else.
type repoProvider interface {
	Users() core.UserRepository
	Projects() core.ProjectRepository
	Links() core.LinkRepository
	Statistics() core.StatisticRepository
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	addr := envOr("ADDR", ":8080")

	// Cancel everything on SIGINT/SIGTERM for a graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Storage engine: Postgres when DATABASE_URL is set, otherwise an in-memory
	// store for local development. The write-behind hit buffer is always in
	// memory so the pixel hot path never hits the database.
	store, buffer, closeStore := openStore(ctx)
	defer closeStore()

	hasher := auth.NewBcryptHasher()
	sessions := auth.NewSessionManager(sessionKey())

	// Core services.
	userService := core.NewUserService(store.Users(), hasher)
	projectService := core.NewProjectService(store.Projects())
	linkService := core.NewLinkService(store.Links(), store.Projects())
	statService := core.NewStatisticService(store.Statistics(), store.Links(), store.Projects(), buffer)

	handler := web.NewHandler(userService, projectService, linkService, statService, sessions)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go runFlushWorker(ctx, statService)

	go func() {
		slog.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "err", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "err", err)
	}

	// Persist whatever hits are still buffered before exiting
	if err := statService.Flush(context.Background()); err != nil {
		slog.Error("final flush failed", "err", err)
	}
}

// runFlushWorker drains the hit buffer into the statistics store on a fixed
// interval until the context is cancelled
func runFlushWorker(ctx context.Context, stats *core.StatisticService) {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := stats.Flush(ctx); err != nil {
				slog.Error("flush hits", "err", err)
			}
		}
	}
}

// openStore selects the storage engine. With DATABASE_URL set it opens a
// Postgres pool (failing fast if it cannot connect); otherwise it falls back to
// the in-memory store. It returns the repo provider, the hit buffer, and a
// cleanup func to release resources on shutdown.
func openStore(ctx context.Context) (repoProvider, core.HitBuffer, func()) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Warn("DATABASE_URL not set; using in-memory store (data is not persisted)")
		store := memory.New()
		return store, store.Buffer(), func() {}
	}

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		slog.Error("connect postgres", "err", err)
		os.Exit(1)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("ping postgres", "err", err)
		os.Exit(1)
	}
	slog.Info("using postgres store")
	return postgres.New(pool), memory.NewBuffer(), pool.Close
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
