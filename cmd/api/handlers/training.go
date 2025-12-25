package handlers

import (
	"encoding/json"
	"net/http"

	"audioml/internal/trainer"
	"audioml/internal/training"

	"github.com/gorilla/mux"
)

type TrainingHandler struct {
	Service *training.Service
	Repo    training.Repository
	Runner  *trainer.PythonRunner
}

type startReq struct {
	Dataset string `json:"dataset"`
	Model   string `json:"model"`
}

func (h *TrainingHandler) Register(r *mux.Router) {
	r.HandleFunc("/training/start", h.Start).Methods("POST")
}

func (h *TrainingHandler) Start(w http.ResponseWriter, r *http.Request) {
	var req startReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	job, err := h.Service.StartJob(r.Context(), req.Dataset, req.Model)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		_ = h.Repo.UpdateStatus(r.Context(), job.ID.String(), training.StatusRunning, nil)
		err := h.Runner.Run(r.Context(), trainer.Request{
			JobID:   job.ID.String(),
			Dataset: req.Dataset,
			Model:   req.Model,
		})
		if err != nil {
			msg := err.Error()
			_ = h.Repo.UpdateStatus(r.Context(), job.ID.String(), training.StatusFailed, &msg)
			return
		}
		_ = h.Repo.UpdateStatus(r.Context(), job.ID.String(), training.StatusCompleted, nil)
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(job)
}
