CREATE TABLE IF NOT EXISTS training_jobs (
  id UUID PRIMARY KEY,
  status VARCHAR(32) NOT NULL,
  dataset_source TEXT NOT NULL,
  model_name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  error TEXT
);
