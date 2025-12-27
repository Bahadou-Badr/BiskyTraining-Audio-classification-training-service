package handlers

import (
	"encoding/json"
	"net/http"

	"audioml/internal/training"

	"github.com/gorilla/mux"
)

type TrainingHandler struct {
	Service *training.Service
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

	job, err := h.Service.StartJob(
		r.Context(),
		req.Dataset,
		req.Model,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(job)
}
