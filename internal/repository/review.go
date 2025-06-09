package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrConflict = errors.New("duplicate review")

type Review struct {
	ID        int64
	UserID    int64
	ProductID int64
	Text      string
	Eval      int
}

type ReviewFilter struct {
	UserID    *int64
	ProductID *int64
}

type ReviewRow struct {
	ID           int64     `json:"id"`
	Text         string    `json:"text"`
	Eval         int       `json:"user_evaluation"`
	AIEval       *bool     `json:"ai_evaluation,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UserName     string    `json:"user"`
	ProductID    int64     `json:"product_id"`
	ProductTitle string    `json:"product_title"`
}

type ReviewRepo interface {
	Insert(ctx context.Context, rv Review) (int64, error)
	SetAIEval(ctx context.Context, id int64, verdict bool) error
	FindByFilter(ctx context.Context, f ReviewFilter) ([]ReviewRow, error)
	AggregateRating(ctx context.Context, productID int64) (votes int64, sum float64, err error)
}

type reviewRepo struct{ db *pgxpool.Pool }

func NewReviewRepo(db *pgxpool.Pool) ReviewRepo { return &reviewRepo{db} }

func (r *reviewRepo) Insert(ctx context.Context, rv Review) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		`INSERT INTO reviews (user_id,product_id,text,user_evaluation,created_at)
		 VALUES ($1,$2,$3,$4,now()) RETURNING id`,
		rv.UserID, rv.ProductID, rv.Text, rv.Eval).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *reviewRepo) SetAIEval(ctx context.Context, id int64, verdict bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE reviews SET ai_evaluation=$2 WHERE id=$1`, id, verdict)
	return err
}

func (r *reviewRepo) FindByFilter(ctx context.Context, f ReviewFilter) ([]ReviewRow, error) {
	sb := strings.Builder{}
	args := []any{}
	sb.WriteString(`
		SELECT rv.id, rv.text, rv.user_evaluation, rv.ai_evaluation, rv.created_at,
		       u.name, p.id, p.title
		FROM reviews rv
		JOIN users u    ON u.id = rv.user_id
		JOIN products p ON p.id = rv.product_id`)

	where := []string{}
	if f.UserID != nil {
		args = append(args, *f.UserID)
		where = append(where, fmt.Sprintf("rv.user_id=$%d", len(args)))
	}
	if f.ProductID != nil {
		args = append(args, *f.ProductID)
		where = append(where, fmt.Sprintf("rv.product_id=$%d", len(args)))
	}
	if len(where) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(where, " AND "))
	}

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ReviewRow
	for rows.Next() {
		var row ReviewRow
		err = rows.Scan(&row.ID, &row.Text, &row.Eval, &row.AIEval, &row.CreatedAt,
			&row.UserName, &row.ProductID, &row.ProductTitle)
		if err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *reviewRepo) AggregateRating(ctx context.Context, productID int64) (int64, float64, error) {
	var votes int64
	var sum float64
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) AS votes,
		       COALESCE(SUM(
		         CASE
		           WHEN ai_evaluation IS FALSE THEN user_evaluation * 0.5
		           ELSE user_evaluation
		         END), 0) AS sum
		FROM reviews
		WHERE product_id = $1
		  AND ai_evaluation IS NOT NULL`, productID).Scan(&votes, &sum)
	return votes, sum, err
}
