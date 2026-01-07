package config

import (
	"log"
	"os"
)

type Config struct {
	Env            string
	Port           string
	HTTPAddr       string
	DatabaseURL    string
	NatsURL        string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	PythonPath     string
	TrainerScript  string
}

func Load() *Config {
	cfg := &Config{
		Env:            getEnv("APP_ENV", "local"),
		Port:           getEnv("PORT", "8080"),
		HTTPAddr:       getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:    getEnv("DATABASE_URL", "postgres://backdev:pa55word@localhost:5432/audioml?sslmode=disable"),
		NatsURL:        getEnv("NATS_URL", "nats://0.0.0.0:4222"),
		MinioEndpoint:  getEnv("MINIO_ENDPOINT", "127.0.0.1:9000"),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "miniouser"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "miniopass"),
		MinioBucket:    getEnv("MINIO_BUCKET", "audio-raw"),
		PythonPath:     getEnv("PYTHON_PATH", "python"),
		TrainerScript:  getEnv("TRAINER_SCRIPT", "./trainer/trainer.py"),
	}

	log.Printf("Config loaded (env=%s)", cfg.Env)
	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
