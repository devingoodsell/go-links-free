package router

import (
	"github.com/devingoodsell/go-links-free/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// New creates a new router with all routes configured
func New(linkHandler *handlers.LinkHandler, authMiddleware gin.HandlerFunc) *gin.Engine {
	r := gin.Default()

	// Public routes
	r.GET("/health", handlers.HealthCheck)
	r.GET("/go/:alias", linkHandler.Redirect)

	// Protected routes
	api := r.Group("/api")
	api.Use(authMiddleware)

	// Link routes
	links := api.Group("/links")
	links.POST("", linkHandler.Create)
	links.GET("", linkHandler.List)
	links.PUT("/:alias", linkHandler.Update)
	links.DELETE("/:alias", linkHandler.Delete)
	links.GET("/:alias/stats", linkHandler.GetStats)

	// Bulk operations
	links.POST("/bulk/delete", linkHandler.BulkDelete)
	links.POST("/bulk/status", linkHandler.BulkUpdateStatus)

	return r
}
