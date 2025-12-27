package models

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo PostgresRepository
}

func NewService(repo PostgresRepository) *Service {
	return &Service{repo: repo}
}

// RegisterFromTraining creates a new model version from a completed training job
func (s *Service) RegisterFromTraining(
	ctx context.Context,
	trainingJobID string,
	modelName string,
	metrics map[string]float64,
	hyperparams map[string]any,
	artifactPath string,
) error {

	versions, err := s.repo.ListByName(ctx, modelName)
	if err != nil {
		return err
	}

	nextVersion := 1
	if len(versions) > 0 {
		nextVersion = versions[0].Version + 1
	}

	model := &ModelVersion{
		ID:            uuid.NewString(),
		TrainingJobID: trainingJobID,
		Name:          modelName,
		Version:       nextVersion,
		Metrics:       metrics,
		Hyperparams:   hyperparams,
		ArtifactPath:  artifactPath,
		IsActive:      false,
	}

	return s.repo.Create(ctx, model)
}

// Activate sets a model version as active
func (s *Service) Activate(ctx context.Context, modelID string) error {
	return s.repo.SetActive(ctx, modelID)
}

// ListVersions lists all versions of a model
func (s *Service) ListVersions(ctx context.Context, name string) ([]ModelVersion, error) {
	return s.repo.ListByName(ctx, name)
}

// GetActive returns active model version
func (s *Service) GetActive(ctx context.Context, name string) (*ModelVersion, error) {
	return s.repo.GetActive(ctx, name)
}
