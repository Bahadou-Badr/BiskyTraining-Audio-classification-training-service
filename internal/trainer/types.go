package trainer

type Request struct {
	JobID   string
	Dataset string
	Model   string
}

type Result struct {
	Metrics      map[string]float64 `json:"metrics"`
	ArtifactPath string             `json:"artifact_path"`
	Params       map[string]any     `json:"params"`
}

type PythonRunner struct {
	PythonBin     string
	TrainerScript string
	WorkDir       string // where artifacts/metrics go
}
