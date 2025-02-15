package main

import (
	"log"
	"net/http"
	"time"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/config"
	"github.com/devingoodsell/go-links-free/internal/handlers"
	"github.com/devingoodsell/go-links-free/internal/jobs"
	"github.com/devingoodsell/go-links-free/internal/middleware"
	"github.com/devingoodsell/go-links-free/internal/models"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize DB connection
	db, err := models.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	linkRepo := models.NewLinkRepository(db)
	analyticsRepo := models.NewAnalyticsRepository(db)
	userRepo := models.NewUserRepository(db)

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret)

	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(authService)
	loggingMiddleware := middleware.NewLoggingMiddleware()

	// Setup routes with all required dependencies
	router := handlers.SetupRoutes(
		cfg,
		authService,
		authMiddleware,
		loggingMiddleware,
		linkRepo,
		analyticsRepo,
		userRepo,
	)

	// Initialize log manager with retention policy
	logManager := models.NewLogManager(db, models.LogRetentionPolicy{
		DetailedRetentionDays:  30,    // Keep detailed logs for 30 days
		AggregateRetentionDays: 90,    // Keep aggregated stats for 90 days
		BatchSize:              1000,  // Delete 1000 records at a time
		MaxDeletionsPerRun:     10000, // Maximum 10000 deletions per cleanup run
	})

	// Start log cleanup job
	cleanupJob := jobs.NewLogCleanupJob(logManager, 24*time.Hour) // Run daily
	cleanupJob.Start()
	defer cleanupJob.Stop()

	log.Printf("Starting server on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
