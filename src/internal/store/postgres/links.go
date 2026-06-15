package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

type linkRepo struct{ pool *pgxpool.Pool }

func (r linkRepo) Create(ctx context.Context, link *core.Link) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO links (project_id, name, link_hash, active, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		link.ProjectId, link.Name, link.LinkHash, link.Active, link.CreatedAt,
	).Scan(&link.Id)
	return mapError(err)
}

func (r linkRepo) GetById(ctx context.Context, id core.ID) (*core.Link, error) {
	return r.queryLink(ctx,
		`SELECT id, project_id, name, link_hash, active, created_at FROM links WHERE id = $1`, id)
}

func (r linkRepo) GetByHash(ctx context.Context, hash string) (*core.Link, error) {
	return r.queryLink(ctx,
		`SELECT id, project_id, name, link_hash, active, created_at FROM links WHERE link_hash = $1`, hash)
}

func (r linkRepo) queryLink(ctx context.Context, query string, arg any) (*core.Link, error) {
	var l core.Link
	err := r.pool.QueryRow(ctx, query, arg).
		Scan(&l.Id, &l.ProjectId, &l.Name, &l.LinkHash, &l.Active, &l.CreatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &l, nil
}

func (r linkRepo) ListByProject(ctx context.Context, projectId core.ID) ([]core.Link, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, project_id, name, link_hash, active, created_at
		 FROM links WHERE project_id = $1 ORDER BY id`,
		projectId)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	out := []core.Link{}
	for rows.Next() {
		var l core.Link
		if err := rows.Scan(&l.Id, &l.ProjectId, &l.Name, &l.LinkHash, &l.Active, &l.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

func (r linkRepo) Update(ctx context.Context, link *core.Link) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE links SET name = $1, active = $2 WHERE id = $3`,
		link.Name, link.Active, link.Id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r linkRepo) Delete(ctx context.Context, id core.ID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM links WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}
