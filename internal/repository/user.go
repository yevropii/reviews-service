package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo interface {
	Exists(ctx context.Context, id int64) (bool, error)
}

type userRepo struct{ db *pgxpool.Pool }

func NewUserRepo(db *pgxpool.Pool) UserRepo { return &userRepo{db} }

func (r *userRepo) Exists(ctx context.Context, id int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `SELECT true FROM users WHERE id=$1`, id).Scan(&ok)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return ok, err
}
