package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/store/postgres"
)

const migrationPath = "../../../../db/migrations/000001_init.up.sql"

// TestPostgresRepositories runs the real adapter against a live database. It is
// skipped unless TEST_DATABASE_URL is set, so the normal `go test ./...` stays
// hermetic. Run it against the compose stack, e.g.:
//
//	TEST_DATABASE_URL=postgres://pixel:pixel@db:5432/pixeltracker?sslmode=disable go test ./internal/store/postgres/...
func TestPostgresRepositories(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run Postgres integration tests")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer pool.Close()
	ensureSchema(t, ctx, pool)

	st := postgres.New(pool)
	users, projects, links, stats := st.Users(), st.Projects(), st.Links(), st.Statistics()

	email := fmt.Sprintf("it-%d@example.com", time.Now().UnixNano())
	u := &core.User{Email: email, PasswordHash: "hash", CreatedAt: time.Now()}
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if u.Id == 0 {
		t.Fatal("user id not assigned by RETURNING")
	}
	// Cascade delete cleans up the whole graph at the end.
	t.Cleanup(func() { _ = users.Delete(context.Background(), u.Id) })

	if got, err := users.GetByEmail(ctx, email); err != nil || got.Id != u.Id {
		t.Fatalf("get by email: id=%d err=%v", got.Id, err)
	}
	if _, err := users.GetById(ctx, 1<<62); !errors.Is(err, core.ErrNotFound) {
		t.Fatalf("missing user: want ErrNotFound, got %v", err)
	}

	// Unique-violation must surface as ErrConflict (the 23505 mapping).
	dup := &core.User{Email: email, PasswordHash: "other", CreatedAt: time.Now()}
	if err := users.Create(ctx, dup); !errors.Is(err, core.ErrConflict) {
		t.Fatalf("duplicate email: want ErrConflict, got %v", err)
	}

	p := &core.Project{OwnerId: u.Id, Name: "Campaign", CreatedAt: time.Now()}
	if err := projects.Create(ctx, p); err != nil {
		t.Fatalf("create project: %v", err)
	}
	l := &core.Link{
		ProjectId: p.Id,
		Name:      "Newsletter",
		LinkHash:  fmt.Sprintf("hash-%d", time.Now().UnixNano()),
		Active:    true,
		CreatedAt: time.Now(),
	}
	if err := links.Create(ctx, l); err != nil {
		t.Fatalf("create link: %v", err)
	}
	if got, err := links.GetByHash(ctx, l.LinkHash); err != nil || got.Id != l.Id {
		t.Fatalf("get by hash: id=%d err=%v", got.Id, err)
	}

	// UpsertHits must accumulate into the same (link, hour) bucket.
	hour := time.Now().Truncate(time.Hour).UTC()
	if err := stats.UpsertHits(ctx, l.Id, hour, 2); err != nil {
		t.Fatalf("upsert 1: %v", err)
	}
	if err := stats.UpsertHits(ctx, l.Id, hour, 3); err != nil {
		t.Fatalf("upsert 2: %v", err)
	}
	byLink, err := stats.ListByLink(ctx, l.Id, time.Time{}, time.Time{})
	if err != nil || len(byLink) != 1 || byLink[0].Hits != 5 {
		t.Fatalf("list by link: want one bucket of 5, got %+v err=%v", byLink, err)
	}
	byProject, err := stats.ListByProject(ctx, p.Id, time.Time{}, time.Time{})
	if err != nil || len(byProject) != 1 || byProject[0].Hits != 5 {
		t.Fatalf("list by project: want one bucket of 5, got %+v err=%v", byProject, err)
	}
}

// ensureSchema applies the init migration when the target database is empty, so
// the test works against both a freshly-created DB and the already-migrated
// compose stack.
func ensureSchema(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	var reg *string
	if err := pool.QueryRow(ctx, "SELECT to_regclass('public.users')::text").Scan(&reg); err != nil {
		t.Fatalf("probe schema: %v", err)
	}
	if reg != nil {
		return
	}
	sql, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}
