package config

import "os"

type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddr     string
	PublicUserURL string
	JWTSecret     string
	MinioAddr     string
	MinioUser     string
	MinioPassword string
	MinioBuckets  string // comma-separated, empty = uploads backup scope is a no-op
}

func Load() Config {
	return Config{
		Port:          getEnv("PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://booking:booking@postgres:5432/booking_doan?sslmode=disable"),
		RedisAddr:     getEnv("REDIS_ADDR", "redis:6379"),
		PublicUserURL: getEnv("PUBLIC_USER_URL", "http://localhost"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-only-insecure-secret-change-me"),
		MinioAddr:     getEnv("MINIO_ADDR", "minio:9000"),
		MinioUser:     getEnv("MINIO_ROOT_USER", ""),
		MinioPassword: getEnv("MINIO_ROOT_PASSWORD", ""),
		MinioBuckets:  getEnv("MINIO_BUCKETS", ""),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
