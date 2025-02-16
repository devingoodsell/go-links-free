package server

import (
	"github.com/devingoodsell/go-links-free/internal/handlers"
	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, linkHandler *handlers.LinkHandler, authMiddleware gin.HandlerFunc) {
	// Protected routes
	protected := r.Group("/api", authMiddleware)
	{
		protected.GET("/links", linkHandler.List)
		protected.POST("/links", linkHandler.Create)
		protected.DELETE("/links/delete/:id", linkHandler.Delete)
		protected.PUT("/links/:alias", linkHandler.Update)
		protected.GET("/links/:alias/stats", linkHandler.GetStats)
		// ... other routes
	}
}
