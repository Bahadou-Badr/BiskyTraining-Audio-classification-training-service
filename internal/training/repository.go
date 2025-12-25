package training

import "context"

type Repository interface {
	Create(ctx context.Context, job *Job) error
	UpdateStatus(ctx context.Context, id string, status Status, err *string) error
	GetByID(ctx context.Context, id string) (*Job, error)
}
