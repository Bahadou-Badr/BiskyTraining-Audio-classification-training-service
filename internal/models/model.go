package models

import "time"

type ModelVersion struct {
	ID            string
	TrainingJobID string
	Name          string
	Version       int
	Metrics       map[string]float64
	Hyperparams   map[string]any
	ArtifactPath  string
	IsActive      bool
	CreatedAt     time.Time
}
