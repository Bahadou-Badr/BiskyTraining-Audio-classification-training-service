CREATE TABLE model_versions (
    id UUID PRIMARY KEY,
    training_job_id UUID NOT NULL,
    name TEXT NOT NULL,
    version INTEGER NOT NULL,
    metrics JSONB NOT NULL,
    hyperparameters JSONB NOT NULL,
    artifact_path TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT false
);

CREATE UNIQUE INDEX model_name_version_idx
ON model_versions(name, version);