package main

import (
	"context"
	"log"
	"net/http"

	"github.com/thuhangnt2010-create/booking-doan-be/internal/config"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/db"
	"github.com/thuhangnt2010-create/booking-doan-be/internal/handlers"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	pgPool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("postgres connect failed: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := db.NewRedisClient(ctx, cfg.RedisAddr)
	if err != nil {
		log.Fatalf("redis connect failed: %v", err)
	}
	defer redisClient.Close()

	mux := http.NewServeMux()
	mux.Handle("/health", &handlers.HealthHandler{DB: pgPool, Redis: redisClient})

	log.Printf("booking-doan-be listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
