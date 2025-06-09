package service

import (
	"context"
	"errors"
	"reviews-service/internal/repository"
)

type LLMClient interface {
	EvaluateSentiment(ctx context.Context, text string) (bool, error)
}

type ReviewService interface {
	AddReview(ctx context.Context, dto AddDTO) error
	GetProductRating(ctx context.Context, productID int64) (repository.ProductRating, error)
	ListReviews(ctx context.Context, f Filter) ([]repository.ReviewRow, error)
}

type AddDTO struct {
	UserID    int64
	ProductID int64
	Text      string
	Score     int // 1–10
}

type Filter struct {
	UserID    *int64
	ProductID *int64
}

var (
	ErrBadInput  = errors.New("bad input")
	ErrUserNF    = errors.New("user not found")
	ErrProductNF = errors.New("product not found")
)

type reviewSvc struct {
	users    repository.UserRepo
	products repository.ProductRepo
	reviews  repository.ReviewRepo
	llm      LLMClient
}

func NewReviewService(u repository.UserRepo, p repository.ProductRepo, r repository.ReviewRepo, llm LLMClient) ReviewService {
	return &reviewSvc{u, p, r, llm}
}

func (s *reviewSvc) AddReview(ctx context.Context, dto AddDTO) error {
	if dto.Score < 1 || dto.Score > 10 || dto.Text == "" {
		return ErrBadInput
	}
	if ok, _ := s.users.Exists(ctx, dto.UserID); !ok {
		return ErrUserNF
	}
	if ok, _ := s.products.Exists(ctx, dto.ProductID); !ok {
		return ErrProductNF
	}

	id, err := s.reviews.Insert(ctx, repository.Review{
		UserID:    dto.UserID,
		ProductID: dto.ProductID,
		Text:      dto.Text,
		Eval:      dto.Score,
	})
	if err != nil {
		return err
	}

	go s.processLLM(context.Background(), id, dto)

	return nil
}

func (s *reviewSvc) processLLM(ctx context.Context, reviewID int64, dto AddDTO) {
	verdict, err := s.llm.EvaluateSentiment(ctx, dto.Text)
	if err != nil {
		return
	}

	if err = s.reviews.SetAIEval(ctx, reviewID, verdict); err != nil {
		return
	}

	pr, err := s.products.GetRating(ctx, dto.ProductID)
	if err != nil {
		return
	}

	weighted := float64(dto.Score)
	if !verdict { // если llm оценила как негативный, то снижаем вес отзыва
		weighted *= 0.5
	}

	newVotes := pr.RatingVotes + 1
	newRating := (pr.Rating*float64(pr.RatingVotes) + weighted) / float64(newVotes)

	if err = s.products.UpdateRating(ctx, dto.ProductID, newRating, newVotes); err != nil {
		return
	}

	s.recalcProduct(ctx, dto.ProductID)
}

func (s *reviewSvc) GetProductRating(ctx context.Context, productID int64) (repository.ProductRating, error) {
	return s.products.GetRating(ctx, productID)
}

func (s *reviewSvc) ListReviews(ctx context.Context, f Filter) ([]repository.ReviewRow, error) {
	return s.reviews.FindByFilter(ctx, repository.ReviewFilter{
		UserID:    f.UserID,
		ProductID: f.ProductID,
	})
}

func (s *reviewSvc) recalcProduct(ctx context.Context, productID int64) {
	votes, sum, err := s.reviews.AggregateRating(ctx, productID)
	if err != nil || votes == 0 {
		return
	}
	rating := sum / float64(votes)
	_ = s.products.UpdateRating(ctx, productID, rating, votes)
}
