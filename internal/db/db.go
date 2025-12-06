package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var Pool *pgxpool.Pool

func Init(ctx context.Context, databaseURL string) error {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse db url: %w", err)
	}
	// configure pool options if desired
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create pool: %w", err)
	}
	Pool = pool

	// run migrations / create tables
	if err := CreateTables(ctx); err != nil {
		return err
	}
	return nil
}

func Close() {
	if Pool != nil {
		Pool.Close()
	}
}

func CreateTables(ctx context.Context) error {
	createSQL := `
CREATE TABLE IF NOT EXISTS audio_files (
  id SERIAL PRIMARY KEY,
  s3_path_raw TEXT NOT NULL,
  filename TEXT,
  duration_seconds DOUBLE PRECISION,
  sample_rate INTEGER,
  status VARCHAR(32) DEFAULT 'uploaded',
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS ingestion_jobs (
  id SERIAL PRIMARY KEY,
  audio_file_id INTEGER REFERENCES audio_files(id) ON DELETE CASCADE,
  subject TEXT,
  status VARCHAR(32) DEFAULT 'queued',
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
`
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_, err := Pool.Exec(ctx, createSQL)
	return err
}
