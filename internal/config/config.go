package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	PublicUserURL string
	JWTSecret     string
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://booking:booking@postgres:5432/booking_doan?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		PublicUserURL: getEnv("PUBLIC_USER_URL", "http://localhost"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-only-insecure-secret-change-me"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
