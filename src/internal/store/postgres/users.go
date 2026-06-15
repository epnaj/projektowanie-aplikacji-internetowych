package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/epnaj/projektowanie-aplikacji-internetowych/internal/core"
)

type userRepo struct{ pool *pgxpool.Pool }

func (r userRepo) Create(ctx context.Context, user *core.User) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, created_at)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		user.Email, user.PasswordHash, user.CreatedAt,
	).Scan(&user.Id)
	return mapError(err)
}

func (r userRepo) GetById(ctx context.Context, id core.ID) (*core.User, error) {
	return r.queryUser(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE id = $1`, id)
}

func (r userRepo) GetByEmail(ctx context.Context, email string) (*core.User, error) {
	return r.queryUser(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email)
}

func (r userRepo) queryUser(ctx context.Context, query string, arg any) (*core.User, error) {
	var u core.User
	err := r.pool.QueryRow(ctx, query, arg).
		Scan(&u.Id, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return nil, mapError(err)
	}
	return &u, nil
}

func (r userRepo) Update(ctx context.Context, user *core.User) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET email = $1, password_hash = $2 WHERE id = $3`,
		user.Email, user.PasswordHash, user.Id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r userRepo) Delete(ctx context.Context, id core.ID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return mapError(err)
	}
	if tag.RowsAffected() == 0 {
		return core.ErrNotFound
	}
	return nil
}
