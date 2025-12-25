package training

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) StartJob(ctx context.Context, datasetSource, modelName string) (*Job, error) {
	job := &Job{
		ID:            uuid.New(),
		Status:        StatusQueued,
		DatasetSource: datasetSource,
		ModelName:     modelName,
		CreatedAt:     time.Now(),
	}
	if err := s.repo.Create(ctx, job); err != nil {
		return nil, err
	}
	return job, nil
}
