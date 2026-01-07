package training

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"audioml/internal/models"
	"audioml/internal/trainer"

	"github.com/google/uuid"
)

type Service struct {
	repo          Repository
	trainerRunner *trainer.PythonRunner
	modelService  *models.Service
}

func NewService(
	repo Repository,
	trainerRunner *trainer.PythonRunner,
	modelService *models.Service,
) *Service {
	return &Service{
		repo:          repo,
		trainerRunner: trainerRunner,
		modelService:  modelService,
	}
}

func (s *Service) StartJob(
	ctx context.Context,
	datasetSource string,
	modelName string,
) (*Job, error) {

	// DEMO CONTRACT
	if !strings.HasPrefix(datasetSource, "local-audio/") {
		return nil, errors.New("only local-audio datasets are supported")
	}

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

	go s.run(context.Background(), job)

	return job, nil
}

func (s *Service) run(ctx context.Context, job *Job) {
	_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusRunning, nil)

	datasetPath := filepath.Join("datasets", job.DatasetSource)

	if _, err := os.Stat(datasetPath); err != nil {
		msg := "dataset not found: " + datasetPath
		_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusFailed, &msg)
		return
	}

	result, err := s.trainerRunner.Run(ctx, trainer.Request{
		JobID:   job.ID.String(),
		Dataset: datasetPath,
		Model:   job.ModelName,
	})

	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusFailed, &msg)
		return
	}

	err = s.modelService.RegisterFromTraining(
		ctx,
		job.ID.String(),
		job.ModelName,
		result.Metrics,
		result.Params,
		result.ArtifactPath,
	)

	if err != nil {
		msg := "model registration failed: " + err.Error()
		_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusFailed, &msg)
		return
	}

	_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusCompleted, nil)
}
