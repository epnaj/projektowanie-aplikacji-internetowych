package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

type statRepo struct{ pool *pgxpool.Pool }

// UpsertHits adds hits to the (link, hour) bucket, creating it on first sight.
// The ON CONFLICT clause is the durable counterpart of the in-memory buffer's
// accumulate-then-flush, keyed by the UNIQUE (link_id, hour) constraint.
func (r statRepo) UpsertHits(ctx context.Context, linkId core.ID, hour time.Time, hits int64) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO statistics (link_id, hour, hits)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (link_id, hour)
		 DO UPDATE SET hits = statistics.hits + EXCLUDED.hits`,
		linkId, hour, hits)
	return mapError(err)
}

func (r statRepo) ListByLink(ctx context.Context, linkId core.ID, from, to time.Time) ([]core.Statistic, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, link_id, hour, hits
		 FROM statistics
		 WHERE link_id = $1
		   AND ($2::timestamptz IS NULL OR hour >= $2::timestamptz)
		   AND ($3::timestamptz IS NULL OR hour <  $3::timestamptz)
		 ORDER BY hour`,
		linkId, nilIfZero(from), nilIfZero(to))
	if err != nil {
		return nil, mapError(err)
	}
	return scanStatistics(rows)
}

func (r statRepo) ListByProject(ctx context.Context, projectId core.ID, from, to time.Time) ([]core.Statistic, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT s.id, s.link_id, s.hour, s.hits
		 FROM statistics s
		 JOIN links l ON l.id = s.link_id
		 WHERE l.project_id = $1
		   AND ($2::timestamptz IS NULL OR s.hour >= $2::timestamptz)
		   AND ($3::timestamptz IS NULL OR s.hour <  $3::timestamptz)
		 ORDER BY s.hour`,
		projectId, nilIfZero(from), nilIfZero(to))
	if err != nil {
		return nil, mapError(err)
	}
	return scanStatistics(rows)
}

func scanStatistics(rows pgx.Rows) ([]core.Statistic, error) {
	defer rows.Close()
	out := []core.Statistic{}
	for rows.Next() {
		var s core.Statistic
		if err := rows.Scan(&s.Id, &s.LinkId, &s.Hour, &s.Hits); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
