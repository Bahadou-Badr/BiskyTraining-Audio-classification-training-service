package main

import (
	"log"
	"net/http"

	"audioml/cmd/api/handlers"
	"audioml/internal/config"
	"audioml/internal/db"
	"audioml/internal/logger"
	"audioml/internal/models"

	"audioml/internal/trainer"
	"audioml/internal/training"

	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	// Logger
	logger.InitLogger(cfg.Env)

	// Database
	db.Init(cfg.DatabaseURL)
	defer db.Close()

	r := mux.NewRouter()

	// Dataset Upload
	datasetHandler := &handlers.DatasetUploadHandler{}
	datasetHandler.Register(r)

	// Models
	modelRepo := models.NewPostgresRepository(db.Pool)
	modelService := models.NewService(modelRepo)

	modelHandler := &handlers.ModelHandler{
		Service: modelService,
	}
	modelHandler.Register(r)

	// Trainer (Python)
	trainerRunner := trainer.NewPythonRunner(
		cfg.PythonPath,
		cfg.TrainerScript,
		"artifacts", // shared output directory
	)

	// Training (Job lifecycle only)
	trainingRepo := training.NewPostgresRepo()
	trainingService := training.NewService(trainingRepo, trainerRunner, modelService)

	trainingHandler := &handlers.TrainingHandler{
		TrainingService: trainingService,
	}
	trainingHandler.Register(r)

	// Server
	log.Println("API listening on", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		log.Fatal(err)
	}
}
