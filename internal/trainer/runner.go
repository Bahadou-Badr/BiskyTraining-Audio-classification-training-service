package trainer

import (
	"context"
	"os/exec"
)

type PythonRunner struct {
	Python string
	Script string
}

func NewPythonRunner(python, script string) *PythonRunner {
	return &PythonRunner{
		Python: python,
		Script: script,
	}
}

func (r *PythonRunner) Run(ctx context.Context, req Request) error {
	cmd := exec.CommandContext(
		ctx,
		r.Python,
		r.Script,
		"--job-id", req.JobID,
		"--dataset", req.Dataset,
		"--model", req.Model,
	)
	return cmd.Run()
}
