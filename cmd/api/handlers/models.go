package handlers

import (
	"encoding/json"
	"net/http"

	"audioml/internal/models"

	"github.com/gorilla/mux"
)

type ModelHandler struct {
	Service *models.Service
}

func (h *ModelHandler) Register(r *mux.Router) {
	r.HandleFunc("/ml/models/{name}/versions", h.ListVersions).Methods("GET")
	r.HandleFunc("/ml/models/{name}/active", h.GetActive).Methods("GET")
	r.HandleFunc("/ml/models/{id}/activate", h.Activate).Methods("POST")
}

// GET /ml/models/{name}/versions
func (h *ModelHandler) ListVersions(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	versions, err := h.Service.ListVersions(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	json.NewEncoder(w).Encode(versions)
}

// GET /ml/models/{name}/active
func (h *ModelHandler) GetActive(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	model, err := h.Service.GetActive(r.Context(), name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if model == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_ = json.NewEncoder(w).Encode(model)
}

// POST /ml/models/{id}/activate
func (h *ModelHandler) Activate(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.Service.Activate(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
