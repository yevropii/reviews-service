package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRating struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Rating      float64 `json:"rating"`
	RatingVotes int64   `json:"rating_votes"`
}

type ProductRepo interface {
	Exists(ctx context.Context, id int64) (bool, error)
	GetRating(ctx context.Context, id int64) (ProductRating, error)
	UpdateRating(ctx context.Context, id int64, rating float64, votes int64) error
}

type productRepo struct{ db *pgxpool.Pool }

func NewProductRepo(db *pgxpool.Pool) ProductRepo { return &productRepo{db} }

func (r *productRepo) Exists(ctx context.Context, id int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `SELECT true FROM products WHERE id=$1`, id).Scan(&ok)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return ok, err
}

func (r *productRepo) GetRating(ctx context.Context, id int64) (ProductRating, error) {
	var pr ProductRating
	err := r.db.QueryRow(ctx,
		`SELECT id,title,rating,rating_votes FROM products WHERE id=$1`, id).
		Scan(&pr.ID, &pr.Title, &pr.Rating, &pr.RatingVotes)
	return pr, err
}

func (r *productRepo) UpdateRating(ctx context.Context, id int64, rating float64, votes int64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE products SET rating=$2, rating_votes=$3 WHERE id=$1`, id, rating, votes)
	return err
}
