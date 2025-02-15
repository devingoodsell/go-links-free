package handlers

import (
	"github.com/gorilla/mux"
	"github.com/yourusername/go-links/internal/auth"
	"github.com/yourusername/go-links/internal/config"
	"github.com/yourusername/go-links/internal/middleware"
	"github.com/yourusername/go-links/internal/models"
)

func SetupRoutes(
	cfg *config.Config,
	authService *auth.AuthService,
	authMiddleware *middleware.AuthMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	linkRepo *models.LinkRepository,
	analyticsRepo *models.AnalyticsRepository,
	userRepo *models.UserRepository,
) *mux.Router {
	router := mux.NewRouter()

	// Add logging middleware to all routes
	router.Use(loggingMiddleware.LogRequest)

	// Auth routes
	authHandler := NewAuthHandler(authService)
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")

	// OKTA routes (only if OKTA is enabled)
	if cfg.EnableOktaSSO {
		router.HandleFunc("/api/auth/okta/login", authHandler.OktaLogin).Methods("GET")
		router.HandleFunc("/api/auth/okta/callback", authHandler.OktaCallback).Methods("GET")
	}

	// Link routes
	linkHandler := NewLinkHandler(linkRepo)
	
	// Public redirect endpoint
	router.HandleFunc("/go/{alias}", linkHandler.Redirect).Methods("GET")

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(authMiddleware.Authenticate)

	// Link management endpoints
	protected.HandleFunc("/links", linkHandler.Create).Methods("POST")
	protected.HandleFunc("/links", linkHandler.List).Methods("GET")
	protected.HandleFunc("/links/{alias}", linkHandler.Update).Methods("PUT")
	protected.HandleFunc("/links/{alias}", linkHandler.Delete).Methods("DELETE")

	// Admin routes
	admin := protected.PathPrefix("/admin").Subrouter()
	admin.Use(authMiddleware.RequireAdmin)

	adminHandler := NewAdminHandler(analyticsRepo, linkRepo, userRepo)
	admin.HandleFunc("/stats", adminHandler.GetSystemStats).Methods("GET")
	admin.HandleFunc("/stats/redirects", adminHandler.GetRedirectsOverTime).Methods("GET")
	admin.HandleFunc("/stats/popular", adminHandler.GetPopularLinks).Methods("GET")
	admin.HandleFunc("/stats/users", adminHandler.GetUserActivity).Methods("GET")
	admin.HandleFunc("/stats/domains", adminHandler.GetTopDomains).Methods("GET")
	admin.HandleFunc("/stats/peak-usage", adminHandler.GetPeakUsage).Methods("GET")
	admin.HandleFunc("/stats/performance", adminHandler.GetPerformanceMetrics).Methods("GET")
	admin.HandleFunc("/links", adminHandler.ListAllLinks).Methods("GET")
	admin.HandleFunc("/links/{alias}", adminHandler.UpdateLinkAdmin).Methods("PUT")

	return router
} 