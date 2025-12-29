package training

import (
	"context"
	"os"
	"path/filepath"
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

// StartJob creates + runs a training job asynchronously
func (s *Service) StartJob(
	ctx context.Context,
	datasetSource string,
	modelName string,
) (*Job, error) {

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

	// Run training async
	go s.run(ctx, job)

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

	err := s.trainerRunner.Run(ctx, trainer.Request{
		JobID:   job.ID.String(),
		Dataset: job.DatasetSource,
		Model:   job.ModelName,
	})

	if err != nil {
		msg := err.Error()
		_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusFailed, &msg)
		return
	}

	// === Phase 4: register trained model ===
	metrics := map[string]float64{
		"accuracy": 0.91, // TODO: real metrics from trainer
		"loss":     0.08,
	}

	hyperparams := map[string]any{
		"epochs":  10,
		"lr":      0.001,
		"backend": "python",
	}

	artifactPath := "s3://models/" + job.ModelName + "/" + job.ID.String() + "/model.bin"

	err = s.modelService.RegisterFromTraining(
		ctx,
		job.ID.String(),
		job.ModelName,
		metrics,
		hyperparams,
		artifactPath,
	)

	if err != nil {
		msg := "model registration failed: " + err.Error()
		_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusFailed, &msg)
		return
	}

	_ = s.repo.UpdateStatus(ctx, job.ID.String(), StatusCompleted, nil)
}

func (s *Service) MarkRunning(ctx context.Context, jobID string) error {
	return s.repo.UpdateStatus(ctx, jobID, StatusRunning, nil)
}

func (s *Service) MarkCompleted(ctx context.Context, jobID string) error {
	return s.repo.UpdateStatus(ctx, jobID, StatusCompleted, nil)
}

func (s *Service) MarkFailed(ctx context.Context, jobID string, err *string) error {
	return s.repo.UpdateStatus(ctx, jobID, StatusFailed, err)
}
