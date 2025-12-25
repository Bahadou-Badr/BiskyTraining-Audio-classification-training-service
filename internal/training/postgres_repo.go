package training

import (
	"context"
	"time"

	"audioml/internal/db"
)

type PostgresRepo struct{}

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) Create(ctx context.Context, job *Job) error {
	q := `
INSERT INTO training_jobs
(id, status, dataset_source, model_name, created_at)
VALUES ($1, $2, $3, $4, $5)
`
	_, err := db.Pool.Exec(
		ctx, q,
		job.ID,
		job.Status,
		job.DatasetSource,
		job.ModelName,
		job.CreatedAt,
	)
	return err
}

func (r *PostgresRepo) UpdateStatus(ctx context.Context, id string, status Status, errMsg *string) error {
	var q string
	if status == StatusRunning {
		q = `UPDATE training_jobs SET status=$1, started_at=$2 WHERE id=$3`
		_, err := db.Pool.Exec(ctx, q, status, time.Now(), id)
		return err
	}

	if status == StatusCompleted || status == StatusFailed {
		q = `UPDATE training_jobs SET status=$1, finished_at=$2, error=$3 WHERE id=$4`
		_, err := db.Pool.Exec(ctx, q, status, time.Now(), errMsg, id)
		return err
	}

	q = `UPDATE training_jobs SET status=$1 WHERE id=$2`
	_, err := db.Pool.Exec(ctx, q, status, id)
	return err
}

func (r *PostgresRepo) GetByID(ctx context.Context, id string) (*Job, error) {
	row := db.Pool.QueryRow(ctx, `
SELECT id, status, dataset_source, model_name,
       created_at, started_at, finished_at, error
FROM training_jobs WHERE id=$1
`, id)

	var job Job
	err := row.Scan(
		&job.ID,
		&job.Status,
		&job.DatasetSource,
		&job.ModelName,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
		&job.Error,
	)
	if err != nil {
		return nil, err
	}
	return &job, nil
}
