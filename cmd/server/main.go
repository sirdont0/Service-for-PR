package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/you/pr-assign-avito/internal/infra"
	"github.com/you/pr-assign-avito/internal/repository"
	pgrepo "github.com/you/pr-assign-avito/internal/repository/pg"
	transport "github.com/you/pr-assign-avito/internal/transport/http"
	uc "github.com/you/pr-assign-avito/internal/usecase"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL required")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		cancel()
		log.Fatalf("db connect: %v", err)
	}
	defer cancel()
	defer pool.Close()

	repoImpl := pgrepo.NewPGRepo(pool)
	var repo repository.Repo = repoImpl

	logger := infra.NewStdLogger()
	prUC := uc.NewPRUsecase(repo)

	handlers := transport.NewHandlers(prUC, repo, logger)
	router := transport.NewRouter(handlers).(*mux.Router)

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Infof("starting server on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Errorf("server error: %v", err)
	}
}
