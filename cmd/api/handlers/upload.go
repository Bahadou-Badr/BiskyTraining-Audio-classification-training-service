package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	natslib "github.com/nats-io/nats.go"

	"audioml/internal/db"
	ilog "audioml/internal/logger"
	"audioml/internal/nats"
	"audioml/internal/s3"
)

type UploadHandler struct {
	S3Client *s3.MinioClient
	JS       natslib.JetStreamContext
}

func (h *UploadHandler) Register(r *mux.Router) {
	r.HandleFunc("/upload", h.HandleUpload).Methods(http.MethodPost)
}

type registerS3Req struct {
	S3URL string `json:"s3_url"`
}

func (h *UploadHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	contentType := r.Header.Get("Content-Type")

	if contentType == "application/json" {
		var req registerS3Req
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.S3URL == "" {
			http.Error(w, "s3_url required", http.StatusBadRequest)
			return
		}
		if err := insertAudioAndPublish(ctx, req.S3URL, "", h.S3Client, h.JS); err != nil {
			ilog.L.Printf("register s3 error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"queued"}`))
		return
	}

	// Expect multipart file upload under field name "file"
	if err := r.ParseMultipartForm(500 << 20); err != nil {
		http.Error(w, "cannot parse multipart", http.StatusBadRequest)
		return
	}
	file, fh, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file form field 'file' required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := fh.Filename
	if filename == "" {
		filename = uuid.New().String() + ".wav"
	}
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".wav"
	}

	objectName := fmt.Sprintf("raw/%s%s", uuid.New().String(), ext)

	// Stream upload directly to MinIO using PutObject
	var buf bytes.Buffer
	size, err := io.Copy(&buf, file)
	if err != nil {
		ilog.L.Printf("read multipart file: %v", err)
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}

	if err := h.S3Client.UploadFromReader(ctx, objectName, bytes.NewReader(buf.Bytes()), size, fh.Header.Get("Content-Type")); err != nil {
		ilog.L.Printf("s3 upload: %v", err)
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}

	// Insert DB row and publish ingest event
	if err := insertAudioAndPublish(ctx, objectName, filename, h.S3Client, h.JS); err != nil {
		ilog.L.Printf("insert/publish: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"queued"}`))
}

func insertAudioAndPublish(ctx context.Context, s3Path, filename string, s3client *s3.MinioClient, js natslib.JetStreamContext) error {
	// insert audio_files
	var id int64
	sql := `INSERT INTO audio_files (s3_path_raw, filename, status, created_at) VALUES ($1, $2, 'uploaded', now()) RETURNING id`
	err := db.Pool.QueryRow(ctx, sql, s3Path, filename).Scan(&id)
	if err != nil {
		return err
	}
	// create ingestion job
	ins := `INSERT INTO ingestion_jobs (audio_file_id, subject, status, created_at) VALUES ($1, $2, 'queued', now())`
	if _, err := db.Pool.Exec(ctx, ins, id, "audio.ingest.raw"); err != nil {
		return err
	}

	// publish event to JetStream
	ev := nats.IngestEvent{
		AudioID:  id,
		S3Path:   s3Path,
		Filename: filename,
	}
	if err := nats.PublishIngestEvent(js, "audio.ingest.raw", ev); err != nil {
		// Later not Now lol, update ingestion_jobs.status = 'publish_failed'
		return err
	}
	return nil
}
