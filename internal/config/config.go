package config

import "os"

type Config struct {
	Port        string
	DatabaseURL string
	RedisAddr   string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://booking:booking@postgres:5432/booking_doan?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "redis:6379"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
