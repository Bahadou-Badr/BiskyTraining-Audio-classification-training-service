package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

// Create inserts a new model version
func (r *PostgresRepository) Create(ctx context.Context, m *ModelVersion) error {
	metricsJSON, err := json.Marshal(m.Metrics)
	if err != nil {
		return err
	}

	hyperJSON, err := json.Marshal(m.Hyperparams)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO model_versions (
			id,
			training_job_id,
			name,
			version,
			metrics,
			hyperparameters,
			artifact_path,
			is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`

	_, err = r.db.Exec(
		ctx,
		query,
		m.ID,
		m.TrainingJobID,
		m.Name,
		m.Version,
		metricsJSON,
		hyperJSON,
		m.ArtifactPath,
		m.IsActive,
	)

	return err
}

// ListByName returns all versions of a model
func (r *PostgresRepository) ListByName(ctx context.Context, name string) ([]ModelVersion, error) {
	query := `
		SELECT
			id,
			training_job_id,
			name,
			version,
			metrics,
			hyperparameters,
			artifact_path,
			is_active,
			created_at
		FROM model_versions
		WHERE name = $1
		ORDER BY version DESC
	`

	rows, err := r.db.Query(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []ModelVersion

	for rows.Next() {
		var m ModelVersion
		var metricsJSON, hyperJSON []byte

		err := rows.Scan(
			&m.ID,
			&m.TrainingJobID,
			&m.Name,
			&m.Version,
			&metricsJSON,
			&hyperJSON,
			&m.ArtifactPath,
			&m.IsActive,
			&m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		json.Unmarshal(metricsJSON, &m.Metrics)
		json.Unmarshal(hyperJSON, &m.Hyperparams)

		models = append(models, m)
	}

	return models, nil
}

// GetActive returns the active model for a given name
func (r *PostgresRepository) GetActive(ctx context.Context, name string) (*ModelVersion, error) {
	query := `
		SELECT
			id,
			training_job_id,
			name,
			version,
			metrics,
			hyperparameters,
			artifact_path,
			is_active,
			created_at
		FROM model_versions
		WHERE name = $1 AND is_active = true
		LIMIT 1
	`

	var m ModelVersion
	var metricsJSON, hyperJSON []byte

	err := r.db.QueryRow(ctx, query, name).Scan(
		&m.ID,
		&m.TrainingJobID,
		&m.Name,
		&m.Version,
		&metricsJSON,
		&hyperJSON,
		&m.ArtifactPath,
		&m.IsActive,
		&m.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	json.Unmarshal(metricsJSON, &m.Metrics)
	json.Unmarshal(hyperJSON, &m.Hyperparams)

	return &m, nil
}

// SetActive marks a model active and deactivates others
func (r *PostgresRepository) SetActive(ctx context.Context, modelID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Find model name
	var name string
	err = tx.QueryRow(
		ctx,
		`SELECT name FROM model_versions WHERE id = $1`,
		modelID,
	).Scan(&name)

	if err != nil {
		return err
	}

	// Deactivate all models with same name
	_, err = tx.Exec(
		ctx,
		`UPDATE model_versions SET is_active = false WHERE name = $1`,
		name,
	)
	if err != nil {
		return err
	}

	// Activate selected model
	res, err := tx.Exec(
		ctx,
		`UPDATE model_versions SET is_active = true WHERE id = $1`,
		modelID,
	)
	if err != nil {
		return err
	}

	rows := res.RowsAffected()
	if rows == 0 {
		return errors.New("model not found")
	}

	return tx.Commit(ctx)
}
