package server

import (
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"reviews-service/internal/metrics"
	"reviews-service/internal/telemetry"
	"time"

	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	repo "reviews-service/internal/repository"
	svc "reviews-service/internal/service"
	"strconv"
)

func New(s svc.ReviewService) http.Handler {
	r := chi.NewRouter()

	r.Use(
		middleware.RequestID,
		telemetry.TracingMiddleware,
		metrics.PrometheusMiddleware,
		RequestLoggerMiddleware(&log.Logger),
	)

	r.Get("/products/{id}/rating", getProductRating(s))
	r.Post("/reviews", addReview(s))
	r.Get("/reviews", listReviews(s))

	return r
}

func getProductRating(s svc.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
		if err != nil {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}

		pr, err := s.GetProductRating(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(pr)
	}
}

type addReviewReq struct {
	UserID     int64  `json:"user_id"`
	ProductID  int64  `json:"product_id"`
	Text       string `json:"text"`
	Evaluation int    `json:"evaluation"` // 1–10
}

func addReview(s svc.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req addReviewReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}

		err := s.AddReview(r.Context(), svc.AddDTO{
			UserID:    req.UserID,
			ProductID: req.ProductID,
			Text:      req.Text,
			Score:     req.Evaluation,
		})
		switch err {
		case nil:
			w.WriteHeader(http.StatusCreated)
		case repo.ErrConflict:
			http.Error(w, "duplicate review", http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func listReviews(s svc.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		var f svc.Filter
		if v := q.Get("user_id"); v != "" {
			id, _ := strconv.ParseInt(v, 10, 64)
			f.UserID = &id
		}
		if v := q.Get("product_id"); v != "" {
			id, _ := strconv.ParseInt(v, 10, 64)
			f.ProductID = &id
		}

		items, err := s.ListReviews(r.Context(), f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(items)
	}
}

// RequestLoggerMiddleware - middleware для логирования запросов
func RequestLoggerMiddleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration", time.Since(start)).
				Str("request_id", middleware.GetReqID(r.Context())).
				Msg("Обработан HTTP запрос")
		})
	}
}
