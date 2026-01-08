package config

import (
	"log"
	"os"
)

type Config struct {
	HTTPPort    string
	PostgresDSN string
}

func Load() *Config {
	cfg := &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		PostgresDSN: getEnv("POSTGRES_DSN", ""),
	}

	if cfg.PostgresDSN == "" {
		log.Fatal("POSTGRES_DSN is required")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
