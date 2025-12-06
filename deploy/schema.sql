-- audio_files
CREATE TABLE IF NOT EXISTS audio_files (
  id SERIAL PRIMARY KEY,
  s3_path_raw TEXT NOT NULL,
  filename TEXT,
  duration_seconds DOUBLE PRECISION,
  sample_rate INTEGER,
  status VARCHAR(32) DEFAULT 'uploaded',
  created_at TIMESTAMPTZ DEFAULT now()
);

-- ingestion_jobs
CREATE TABLE IF NOT EXISTS ingestion_jobs (
  id SERIAL PRIMARY KEY,
  audio_file_id INTEGER NOT NULL REFERENCES audio_files(id) ON DELETE CASCADE,
  subject TEXT,
  status VARCHAR(32) DEFAULT 'queued',
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
