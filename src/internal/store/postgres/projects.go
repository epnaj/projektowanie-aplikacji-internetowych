package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

type projectRepo struct{ pool *pgxpool.Pool }

func (r projectRepo) Create(ctx context.Context, project *core.Project) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO projects (owner_id, name, created_at)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		project.OwnerId, project.Name, project.CreatedAt,
	).Scan(&project.Id)
	return mapError(err)
}

func (r projectRepo) GetById(ctx context.Context, id core.ID) (*core.Project, error) {
	var p core.Project
	err := r.pool.QueryRow(ctx,
		`SELECT id, owner_id, name, created_at FROM projects WHERE id = $1`, id).
		Scan(&p.Id, &p.OwnerId, &p.Name, &p.CreatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &p, nil
}

func (r projectRepo) ListByOwner(ctx context.Context, ownerId core.ID) ([]core.Project, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, owner_id, name, created_at FROM projects WHERE owner_id = $1 ORDER BY id`,
		ownerId)
	if err != nil {
		return nil, mapError(err)
	}
	defer rows.Close()

	out := []core.Project{}
	for rows.Next() {
		var p core.Project
		if err := rows.Scan(&p.Id, &p.OwnerId, &p.Name, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r projectRepo) Update(ctx context.Context, project *core.Project) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE projects SET name = $1 WHERE id = $2`,
		project.Name, project.Id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r projectRepo) Delete(ctx context.Context, id core.ID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}
