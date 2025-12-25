CREATE TABLE IF NOT EXISTS audio_features (
  id SERIAL PRIMARY KEY,
  audio_file_id INTEGER REFERENCES audio_files(id) ON DELETE CASCADE,
  segment_index INTEGER NOT NULL,
  s3_path TEXT NOT NULL,
  feature_type VARCHAR(32) NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);