package handlers

import (
	"log"
	"time"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/config"
	"github.com/devingoodsell/go-links-free/internal/middleware"
	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	cfg *config.Config,
	authService *auth.AuthService,
	authMiddleware *middleware.AuthMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	linkRepo *models.LinkRepository,
	analyticsRepo *models.AnalyticsRepository,
	userRepo *models.UserRepository,
) *gin.Engine {
	log.Println("Setting up routes...")
	gin.SetMode(gin.DebugMode)
	router := gin.New()

	// Add recovery middleware
	router.Use(gin.Recovery())

	// Add logger middleware
	router.Use(gin.Logger())

	// Debug middleware to log all requests and headers
	router.Use(func(c *gin.Context) {
		log.Printf("DEBUG: Incoming request: %s %s", c.Request.Method, c.Request.URL.Path)
		log.Printf("DEBUG: Request headers: %+v", c.Request.Header)
		c.Next()
	})

	// Add CORS middleware using gin-contrib/cors
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8081"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add health check endpoint after CORS middleware
	router.GET("/api/health", func(c *gin.Context) {
		log.Printf("DEBUG: Health check handler called from IP: %s", c.ClientIP())
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Add root handler for testing
	router.GET("/", func(c *gin.Context) {
		c.String(200, "Server is running")
	})

	// Test route to verify routing is working
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Auth routes
	authHandler := NewAuthHandler(authService, userRepo)
	// Protected routes with auth middleware
	protected := router.Group("/api")
	protected.Use(authMiddleware.AuthenticateGin)

	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	protected.GET("/auth/me", authHandler.GetCurrentUser)

	// OKTA routes (only if OKTA is enabled)
	if cfg.EnableOktaSSO {
		router.GET("/api/auth/okta/login", authHandler.OktaLogin)
		router.GET("/api/auth/okta/callback", authHandler.OktaCallback)
	}

	// Link routes
	linkHandler := NewLinkHandler(linkRepo)

	// Public redirect endpoint
	router.GET("/go/:alias", linkHandler.Redirect)

	// Link management endpoints
	protected.GET("/links", linkHandler.List)
	protected.POST("/links", linkHandler.Create)
	protected.DELETE("/links/delete/:id", linkHandler.Delete)
	protected.PUT("/links/:id", linkHandler.Update)
	protected.GET("/links/:alias/stats", linkHandler.GetStats)

	// Bulk operations
	protected.POST("/links/bulk/delete", linkHandler.BulkDelete)
	protected.POST("/links/bulk/status", linkHandler.BulkUpdateStatus)

	// Admin routes
	admin := protected.Group("/admin")
	admin.Use(authMiddleware.RequireAdminGin)

	adminHandler := NewAdminHandler(analyticsRepo, linkRepo, userRepo)
	admin.GET("/stats", adminHandler.GetSystemStats)
	admin.GET("/stats/redirects", adminHandler.GetRedirectsOverTime)
	admin.GET("/stats/popular", adminHandler.GetPopularLinks)
	admin.GET("/stats/users", adminHandler.GetUserActivity)
	admin.GET("/stats/domains", adminHandler.GetTopDomains)
	admin.GET("/stats/peak-usage", adminHandler.GetPeakUsage)
	admin.GET("/stats/performance", adminHandler.GetPerformanceMetrics)
	admin.GET("/links", adminHandler.ListAllLinks)
	admin.PUT("/links/:alias", adminHandler.UpdateLinkAdmin)

	// Print all routes at the end
	routes := router.Routes()
	log.Println("\n=== Registered Routes ===")
	for _, route := range routes {
		log.Printf("Route: %s\t%s", route.Method, route.Path)
	}
	log.Println("=== End Routes ===\n")

	return router
}
