package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	// Postgres
	PostgresDSN string

	// NATS
	NATSUrl string

	// MinIO
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

	// Transfer
	DefaultChunkSizeMB int
}

func Load(envFile string) (*Config, error) {
	_ = godotenv.Load(envFile)

	cfg := &Config{
		PostgresDSN:        requireEnv("POSTGRES_DSN"),
		NATSUrl:            getEnv("NATS_URL", "nats://127.0.0.1:4222"),
		MinIOEndpoint:      getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinIOAccessKey:     requireEnv("MINIO_ACCESS_KEY"),
		MinIOSecretKey:     requireEnv("MINIO_SECRET_KEY"),
		MinIOBucket:        getEnv("MINIO_BUCKET", "moveric-transfers"),
		MinIOUseSSL:        getBool("MINIO_USE_SSL", false),
		DefaultChunkSizeMB: getInt("CHUNK_SIZE_MB", 64),
	}

	return cfg, nil
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required env var %s is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
