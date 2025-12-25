package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	nats "github.com/nats-io/nats.go"

	"audioml/cmd/api/handlers"
	"audioml/internal/config"
	"audioml/internal/db"
	"audioml/internal/logger"
	natspkg "audioml/internal/nats"
	"audioml/internal/preprocessing"
	s3pkg "audioml/internal/s3"
	"audioml/internal/trainer"
	"audioml/internal/training"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	logger.InitLogger(cfg.Env)

	// DB init with retry: try 8 times with 2s interval
	if err := db.InitWithRetry(ctx, cfg.DatabaseURL, 8, 2*time.Second); err != nil {
		logger.L.Fatalf("db init: %v", err)
	}
	defer db.Close()

	// NATS
	nc, err := natspkg.Connect(cfg.NatsURL)
	if err != nil {
		logger.L.Fatalf("nats connect: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logger.L.Fatalf("jetstream: %v", err)
	}
	// ensure stream exists (best effort)
	streamCfg := &nats.StreamConfig{
		Name:     "AUDIO_STREAM",
		Subjects: []string{"audio.ingest.raw"},
		Storage:  nats.FileStorage,
	}
	if _, err := js.AddStream(streamCfg); err != nil {
		logger.L.Printf("add stream: %v (maybe already exists)", err)
	}

	// MinIO
	s3c, err := s3pkg.NewMinioClient(cfg)
	if err != nil {
		logger.L.Fatalf("s3 init: %v", err)
	}
	if err := s3c.MakeBucketIfNotExists(cfg.MinioBucket); err != nil {
		logger.L.Printf("make bucket: %v", err)
	}

	prepRepo := preprocessing.NewPostgresRepo()

	pipeline := &preprocessing.Pipeline{
		Repo:   prepRepo,
		// S3:     s3c,
		Python: "python",
	}

	worker := &preprocessing.Worker{
		Pipeline: pipeline,
		Repo:     prepRepo,
	}

	if err := worker.Start(js); err != nil {
		logger.L.Fatalf("preprocessing worker: %v", err)
	}

	// Training
	trainingRepo := training.NewPostgresRepo()
	trainingService := training.NewService(trainingRepo)
	trainerRunner := trainer.NewPythonRunner("python", "trainer/train.py")

	th := &handlers.TrainingHandler{
		Service: trainingService,
		Repo:    trainingRepo,
		Runner:  trainerRunner,
	}

	// Router
	r := mux.NewRouter()
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	uh := &handlers.UploadHandler{
		S3Client: s3c,
		JS:       js,
	}
	uh.Register(r) // uh UploadHandler
	th.Register(r) // uh TrainingHandler

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + cfg.Port,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	// start server
	go func() {
		logger.L.Printf("api listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown on Ctrl-C
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.L.Printf("shutting down")
	ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShut); err != nil {
		logger.L.Printf("shutdown error: %v", err)
	}
}
