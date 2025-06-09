package models

import "time"

type User struct {
	ID        int64     `db:"id"        json:"id"`
	UUID      string    `db:"uuid"      json:"uuid"`
	Name      string    `db:"name"      json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Product struct {
	ID          int64   `db:"id"            json:"id"`
	Title       string  `db:"title"         json:"title"`
	Description string  `db:"description"   json:"description"`
	Price       float64 `db:"price"         json:"price"`
	RatingVotes int64   `db:"rating_votes"  json:"rating_votes"`
	Rating      float64 `db:"rating"        json:"rating"`
}

type Review struct {
	ID        int64  `db:"id"             json:"id"`
	UserID    int64  `db:"user_id"        json:"user_id"`
	ProductID int64  `db:"product_id"     json:"product_id"`
	Text      string `db:"text"           json:"text"`
	UserEval  bool   `db:"user_evaluation" json:"user_evaluation"`
	AIEval    bool   `db:"ai_evaluation"   json:"ai_evaluation"`
}
