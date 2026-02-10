package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	DBPath        string
	JWTSecret     string
	EncryptionKey string
	TokenTTL      time.Duration
	WorkerPoolSize int
}

func Load() (Config, error) {
	cfg := Config{
		Port:          getEnv("PORT", ":8080"),
		DBPath:        getEnv("DB_PATH", "./data/vault.db"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		EncryptionKey: os.Getenv("VAULT_ENC_KEY"),
		TokenTTL:      parseDurationMinutes(getEnv("TOKEN_TTL_MIN", "60")),
		WorkerPoolSize: parseInt(getEnv("WORKER_POOL_SIZE", "8"), 8),
	}

	if cfg.JWTSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}
	if cfg.EncryptionKey == "" {
		return Config{}, errors.New("VAULT_ENC_KEY is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDurationMinutes(value string) time.Duration {
	minutes, err := strconv.Atoi(value)
	if err != nil || minutes <= 0 {
		minutes = 60
	}
	return time.Duration(minutes) * time.Minute
}

func parseInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
