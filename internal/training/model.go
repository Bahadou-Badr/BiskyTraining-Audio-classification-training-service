package training

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Job struct {
	ID            uuid.UUID
	Status        Status
	DatasetSource string
	ModelName     string
	CreatedAt     time.Time
	StartedAt     *time.Time
	FinishedAt    *time.Time
	Error         *string
}
