package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"audioml/internal/dataset"

	"github.com/gorilla/mux"
)

type DatasetUploadHandler struct{}

func (h *DatasetUploadHandler) Register(r *mux.Router) {
	r.HandleFunc("/datasets/upload", h.Upload).Methods(http.MethodPost)
}

type datasetUploadResponse struct {
	Dataset string `json:"dataset"`
	Files   int    `json:"files"`
}

func (h *DatasetUploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(1 << 30); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	datasetName := r.FormValue("dataset")
	if datasetName == "" {
		http.Error(w, "dataset field is required", http.StatusBadRequest)
		return
	}

	basePath := filepath.Join("datasets", "local-audio", datasetName)

	if err := os.MkdirAll(basePath, 0755); err != nil {
		http.Error(w, "failed to create dataset directory", http.StatusInternalServerError)
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		http.Error(w, "no files uploaded (use field name 'files')", http.StatusBadRequest)
		return
	}

	saved := 0

	for _, fh := range files {
		src, err := fh.Open()
		if err != nil {
			continue
		}

		dstPath := filepath.Join(basePath, fh.Filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			src.Close()
			continue
		}

		if _, err := io.Copy(dst, src); err == nil {
			// 🔍 FFmpeg validation
			if err := dataset.ValidateAudio(dstPath); err == nil {
				saved++
			} else {
				// Invalid audio → delete file
				os.Remove(dstPath)
			}
		}

		dst.Close()
		src.Close()
	}

	if saved == 0 {
		http.Error(w, "no valid audio files were saved", http.StatusBadRequest)
		return
	}

	resp := datasetUploadResponse{
		Dataset: fmt.Sprintf("local-audio/%s", datasetName),
		Files:   saved,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
