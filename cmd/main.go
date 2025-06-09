package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"reviews-service/internal/server"
	"time"

	"reviews-service/internal/db"
	"reviews-service/internal/llm"
	"reviews-service/internal/repository"
	"reviews-service/internal/service"
)

func main() {
	ctx := context.Background()

	// ---------- PostgreSQL ----------
	pool, err := db.New(ctx, os.Getenv("POSTGRES_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	// ---------- Repositories ----------
	userRepo := repository.NewUserRepo(pool)
	productRepo := repository.NewProductRepo(pool)
	reviewRepo := repository.NewReviewRepo(pool)

	// ---------- Ollama client ----------
	ollamaURL := os.Getenv("OLLAMA_URL")
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	llmClient := ollama.New(ollamaURL, ollamaModel, nil)

	// ---------- Services ----------
	reviewSvc := service.NewReviewService(userRepo, productRepo, reviewRepo, llmClient)

	// ---------- HTTP ----------
	s := server.New(reviewSvc)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      s,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Println("HTTP up on :8080")
	log.Fatal(srv.ListenAndServe())
}
