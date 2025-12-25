package preprocessing

import "context"

type Repository interface {
	MarkProcessing(ctx context.Context, audioID int64) error
	MarkDone(ctx context.Context, audioID int64) error
	MarkFailed(ctx context.Context, audioID int64, err string) error
	InsertFeature(
		ctx context.Context,
		audioID int64,
		segment int,
		s3Path string,
		featureType string,
	) error
}
