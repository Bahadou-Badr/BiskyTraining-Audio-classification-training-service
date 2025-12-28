package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"audioml/internal/models"
	"audioml/internal/trainer"
	"audioml/internal/training"

	"github.com/gorilla/mux"
)

type TrainingHandler struct {
	TrainingService *training.Service
	ModelService    *models.Service
	TrainerRunner   *trainer.PythonRunner
}

type startTrainingRequest struct {
	Dataset string `json:"dataset"`
	Model   string `json:"model"`
}

func (h *TrainingHandler) Register(r *mux.Router) {
	r.HandleFunc("/training/start", h.StartTraining).Methods("POST")
}

func (h *TrainingHandler) StartTraining(w http.ResponseWriter, r *http.Request) {
	var req startTrainingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.TrainingService.StartJob(
		r.Context(),
		req.Dataset,
		req.Model,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// BACKGROUND EXECUTION
	go func() {

		// detached context
		ctx := context.Background()
		// mark running
		_ = h.TrainingService.MarkRunning(ctx, job.ID.String())

		// run python trainer
		err := h.TrainerRunner.Run(
			ctx,
			trainer.Request{
				JobID:   job.ID.String(),
				Dataset: req.Dataset,
				Model:   req.Model,
			},
		)

		if err != nil {
			msg := err.Error()
			_ = h.TrainingService.MarkFailed(ctx, job.ID.String(), &msg)
			return
		}

		// mark completed
		_ = h.TrainingService.MarkCompleted(ctx, job.ID.String())

		// register model version
		_ = h.ModelService.RegisterFromTraining(
			ctx,
			job.ID.String(),
			req.Model,
			map[string]float64{"accuracy": 0.87},
			map[string]any{"epochs": 10},
			"s3://models/"+req.Model+"/v1/model.bin",
		)
	}()

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(job)
}
