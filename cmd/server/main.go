package main

import (
	"log"
	"time"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/config"
	"github.com/devingoodsell/go-links-free/internal/db"
	"github.com/devingoodsell/go-links-free/internal/handlers"
	"github.com/devingoodsell/go-links-free/internal/jobs"
	"github.com/devingoodsell/go-links-free/internal/middleware"
	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize DB connection
	database, err := db.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repositories
	linkRepo := models.NewLinkRepository(database)
	analyticsRepo := models.NewAnalyticsRepository(database)
	userRepo := models.NewUserRepository(database)
	requestLogRepo := models.NewRequestLogRepository(database)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 24*time.Hour)

	// Initialize auth service
	authService := auth.NewAuthService(userRepo, jwtManager, cfg.EnableOktaSSO)

	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)
	loggingMiddleware := middleware.NewLoggingMiddleware(requestLogRepo)

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

	// Print all registered routes
	log.Println("\n=== Registered Routes ===")
	for _, route := range router.Routes() {
		log.Printf("Route: %s\t%s", route.Method, route.Path)
	}
	log.Println("=== End Routes ===\n")

	// Print registered routes
	for _, route := range router.Routes() {
		log.Printf("Registered route: %s %s", route.Method, route.Path)
	}

	log.Printf("Configuration: %+v", cfg)

	// Initialize log manager with retention policy
	logManager := models.NewLogManager(database, models.LogRetentionPolicy{
		DetailedRetentionDays:  30,    // Keep detailed logs for 30 days
		AggregateRetentionDays: 90,    // Keep aggregated stats for 90 days
		BatchSize:              1000,  // Delete 1000 records at a time
		MaxDeletionsPerRun:     10000, // Maximum 10000 deletions per cleanup run
	})

	// Start log cleanup job
	cleanupJob := jobs.NewLogCleanupJob(logManager, 24*time.Hour) // Run daily
	cleanupJob.Start()
	defer cleanupJob.Stop()

	// Enable CORS
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Setup routes with all required dependencies
	r = handlers.SetupRoutes(
		cfg,
		authService,
		authMiddleware,
		loggingMiddleware,
		linkRepo,
		analyticsRepo,
		userRepo,
	)

	// Start server
	log.Printf("Starting server on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
