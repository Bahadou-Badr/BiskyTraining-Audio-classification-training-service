package s3

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"audioml/internal/config"
)

type MinioClient struct {
	Client *minio.Client
	Bucket string
}

func NewMinioClient(cfg *config.Config) (*MinioClient, error) {
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	return &MinioClient{
		Client: minioClient,
		Bucket: cfg.MinioBucket,
	}, nil
}

func (m *MinioClient) MakeBucketIfNotExists(bucket string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := m.Client.BucketExists(ctx, bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := m.Client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	return nil
}

func (m *MinioClient) UploadFromReader(ctx context.Context, objectName string, reader io.Reader, objectSize int64, contentType string) error {
	_, err := m.Client.PutObject(ctx, m.Bucket, objectName, reader, objectSize, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}
