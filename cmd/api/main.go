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
	// ctx := context.Background() // Create a context
	db.Init(cfg.DatabaseURL)
	defer db.Close()

	// Router
	r := mux.NewRouter()

	// ======================
	// Models (Phase 4)
	// ======================
	modelRepo := models.NewPostgresRepository(db.Pool)
	modelService := models.NewService(modelRepo)
	modelHandler := &handlers.ModelHandler{
		Service: modelService,
	}
	modelHandler.Register(r)

	// ======================
	// Trainer (Python)
	// ======================
	// trainerRunner := trainer.NewPythonRunner(
	// 	cfg.PythonPath,    // e.g. "python"
	// 	cfg.TrainerScript, // e.g. "trainer/train.py"
	// )
	trainerRunner := trainer.NewPythonRunner(
		"python",
		"./trainer/trainer.py",
	)

	// ======================
	// Training (Phase 2–4)
	// ======================
	trainingRepo := training.NewPostgresRepo()
	trainingService := training.NewService(
		trainingRepo,
		trainerRunner,
		modelService,
	)

	trainingHandler := &handlers.TrainingHandler{
		TrainingService: trainingService,
		ModelService:    modelService,
		TrainerRunner:   trainerRunner,
	}

	trainingHandler.Register(r)

	// ======================
	// Server
	// ======================
	log.Println("API listening on", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, r); err != nil {
		log.Fatal(err)
	}
}
