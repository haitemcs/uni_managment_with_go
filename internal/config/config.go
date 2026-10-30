package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	TokenTTL    time.Duration
}

func Load() Config {
	hours, err := strconv.Atoi(getenv("JWT_TTL_HOURS", "24"))
	if err != nil || hours < 1 {
		hours = 24
	}
	return Config{
		Port:        getenv("APP_PORT", "8080"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://university:university@localhost:5433/university_db?sslmode=disable"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		TokenTTL:    time.Duration(hours) * time.Hour,
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
