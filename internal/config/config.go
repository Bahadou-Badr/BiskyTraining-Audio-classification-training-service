package config

import (
	"os"
)

type Config struct {
	Port           string
	DatabaseURL    string
	NatsURL        string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	Env            string
}

func Load() (*Config, error) {
	// sensible local defaults (127.0.0.1 to avoid socket/localhost ambiguity on Windows)
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://backdev:pa55word@127.0.0.1:5432/audioml?sslmode=disable"),
		NatsURL:        getEnv("NATS_URL", "nats://0.0.0.0:4222"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "miniouser"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "miniopass"),
		MinioBucket:    getEnv("MINIO_BUCKET", "audio-raw"),
		Env:            getEnv("ENV", "dev"),
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
