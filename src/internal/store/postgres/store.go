package postgres

import (
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

// Store binds the core repository interfaces to a Postgres connection pool.
// Each accessor returns a thin repo sharing that pool, mirroring the shape of
// the in-memory store so the composition root can swap one for the other.
//
// Note there is no Buffer() here: the write-behind HitBuffer stays in memory
// even against Postgres, so the pixel hot path never touches the database.
type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Users() core.UserRepository           { return userRepo{s.pool} }
func (s *Store) Projects() core.ProjectRepository     { return projectRepo{s.pool} }
func (s *Store) Links() core.LinkRepository           { return linkRepo{s.pool} }
func (s *Store) Statistics() core.StatisticRepository { return statRepo{s.pool} }

// mapError translates driver-specific errors into the core sentinels the rest
// of the app understands, so no pgx detail leaks past this package. A missing
// row becomes ErrNotFound; a unique-violation (SQLSTATE 23505) becomes
// ErrConflict.
func mapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return core.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return core.ErrConflict
	}
	return err
}

// nilIfZero maps a zero time to a SQL NULL so a missing range bound reads as
// "unbounded", matching the in-memory store's [from, to) semantics.
func nilIfZero(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}
