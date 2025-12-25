package preprocessing

import (
	"context"

	"audioml/internal/db"
)

type PostgresRepo struct{}

func NewPostgresRepo() *PostgresRepo {
	return &PostgresRepo{}
}

func (r *PostgresRepo) MarkProcessing(ctx context.Context, audioID int64) error {
	_, err := db.Pool.Exec(
		ctx,
		`UPDATE audio_files SET status='processing' WHERE id=$1`,
		audioID,
	)
	return err
}

func (r *PostgresRepo) MarkDone(ctx context.Context, audioID int64) error {
	_, err := db.Pool.Exec(
		ctx,
		`UPDATE audio_files SET status='features_ready' WHERE id=$1`,
		audioID,
	)
	return err
}

func (r *PostgresRepo) MarkFailed(ctx context.Context, audioID int64, msg string) error {
	_, err := db.Pool.Exec(
		ctx,
		`UPDATE audio_files SET status='failed' WHERE id=$1`,
		audioID,
	)
	return err
}

func (r *PostgresRepo) InsertFeature(
	ctx context.Context,
	audioID int64,
	segment int,
	s3Path string,
	featureType string,
) error {
	_, err := db.Pool.Exec(ctx, `
INSERT INTO audio_features
(audio_file_id, segment_index, s3_path, feature_type)
VALUES ($1, $2, $3, $4)
`, audioID, segment, s3Path, featureType)
	return err
}
