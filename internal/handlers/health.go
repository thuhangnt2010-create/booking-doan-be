package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

type healthResponse struct {
	Status string `json:"status"`
	DB     string `json:"db"`
	Redis  string `json:"redis"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp := healthResponse{Status: "ok", DB: "ok", Redis: "ok"}
	code := http.StatusOK

	if err := h.DB.Ping(ctx); err != nil {
		resp.DB = "fail"
		resp.Status = "degraded"
		code = http.StatusServiceUnavailable
	}
	if err := h.Redis.Ping(ctx).Err(); err != nil {
		resp.Redis = "fail"
		resp.Status = "degraded"
		code = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}
