package trainer

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

func NewPythonRunner(pythonBin, script, workDir string) *PythonRunner {
	return &PythonRunner{
		PythonBin:     pythonBin,
		TrainerScript: script,
		WorkDir:       workDir,
	}
}

func (r *PythonRunner) Run(ctx context.Context, req Request) (*Result, error) {
	cmd := exec.CommandContext(
		ctx,
		r.PythonBin,
		r.TrainerScript,
		"--job-id", req.JobID,
		"--dataset", req.Dataset,
		"--model", req.Model,
		"--out", "artifacts",
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	fmt.Println("Python raw output:", string(out))

	var result Result
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
