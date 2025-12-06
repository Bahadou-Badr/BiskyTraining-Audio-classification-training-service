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
	natspkg "audioml/internal/nats"
	"audioml/internal/s3"
)

// UploadHandler handles file uploads or registering S3 URLs.
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

	// If JSON body with s3_url
	if r.Header.Get("Content-Type") == "application/json" {
		var req registerS3Req
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if req.S3URL == "" {
			http.Error(w, "s3_url required", http.StatusBadRequest)
			return
		}
		// create DB row and publish
		if err := insertAudioAndPublish(ctx, req.S3URL, "", h.S3Client, h.JS); err != nil {
			ilog.L.Printf("register s3 error: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(`{"status":"queued"}`))
		return
	}

	// otherwise multipart file
	if err := r.ParseMultipartForm(200 << 20); err != nil {
		http.Error(w, "cannot parse multipart", http.StatusBadRequest)
		return
	}
	fh, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file form field 'file' required", http.StatusBadRequest)
		return
	}
	defer fh.Close()

	// read file bytes
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, fh); err != nil {
		ilog.L.Printf("read file: %v", err)
		http.Error(w, "read file error", http.StatusInternalServerError)
		return
	}
	filename := r.FormValue("filename")
	if filename == "" {
		filename = uuid.New().String() + ".wav"
	}
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".wav"
	}
	objectName := fmt.Sprintf("raw/%s%s", uuid.New().String(), ext)

	// upload to minio
	if err := h.S3Client.UploadFromReader(ctx, objectName, bytes.NewReader(buf.Bytes()), int64(buf.Len()), "audio/wav"); err != nil {
		ilog.L.Printf("s3 upload: %v", err)
		http.Error(w, "upload error", http.StatusInternalServerError)
		return
	}
	s3Path := objectName

	// insert DB row and publish event
	if err := insertAudioAndPublish(ctx, s3Path, filename, h.S3Client, h.JS); err != nil {
		ilog.L.Printf("insert/publish: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"queued"}`))
}

// insertAudioAndPublish inserts audio_files and ingestion_job and publishes a NATS ingest event.
func insertAudioAndPublish(ctx context.Context, s3Path, filename string, s3client *s3.MinioClient, js natslib.JetStreamContext) error {
	// insert into audio_files
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

	// publish NATS JetStream event
	ev := natspkg.IngestEvent{
		AudioID:  id,
		S3Path:   s3Path,
		Filename: filename,
	}
	if err := natspkg.PublishIngestEvent(js, "audio.ingest.raw", ev); err != nil {
		// log but do not fail hard — up to you (could set job status to 'publish_failed')
		ilog.L.Printf("publish ingest event: %v", err)
		return err
	}

	return nil
}
