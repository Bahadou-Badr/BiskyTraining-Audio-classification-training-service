package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	natsio "github.com/nats-io/nats.go"

	"audioml/cmd/api/handlers"
	"audioml/internal/config"
	"audioml/internal/db"
	"audioml/internal/logger"
	"audioml/internal/nats"
	s3pkg "audioml/internal/s3"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	logger.InitLogger(cfg.Env)

	// DB
	if err := db.Init(ctx, cfg.DatabaseURL); err != nil {
		logger.L.Fatalf("db init: %v", err)
	}
	defer db.Close()

	// NATS connect and JetStream context
	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.L.Fatalf("nats connect: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		logger.L.Fatalf("jetstream: %v", err)
	}

	// ensure stream exists (simple stream)
	streamCfg := &natsio.StreamConfig{
		Name:     "AUDIO_STREAM",
		Subjects: []string{"audio.ingest.raw"},
		Storage:  natsio.FileStorage,
	}
	if _, err := js.AddStream(streamCfg); err != nil {
		// stream might already exist; log only
		logger.L.Printf("add stream: %v", err)
	}

	// S3 client
	s3c, err := s3pkg.NewMinioClient(cfg)
	if err != nil {
		logger.L.Fatalf("s3 init: %v", err)
	}
	if err := s3c.MakeBucketIfNotExists(cfg.MinioBucket); err != nil {
		logger.L.Printf("make bucket: %v", err)
	}

	// Router
	r := mux.NewRouter()
	// health
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}).Methods("GET")

	// upload handler
	uh := &handlers.UploadHandler{
		S3Client: s3c,
		JS:       js,
	}
	uh.Register(r)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + cfg.Port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// run server
	go func() {
		logger.L.Printf("api listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Fatalf("server error: %v", err)
		}
	}()

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	logger.L.Printf("shutting down")
	ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShut); err != nil {
		logger.L.Printf("shutdown error: %v", err)
	}
}
