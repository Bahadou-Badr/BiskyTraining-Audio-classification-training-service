package models

type Repository interface {
	Create(model *ModelVersion) error
	ListByName(name string) ([]ModelVersion, error)
	SetActive(modelID string) error
	GetActive(name string) (*ModelVersion, error)
}
