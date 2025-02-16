package handlers

import (
	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/config"
	"github.com/devingoodsell/go-links-free/internal/middleware"
	"github.com/devingoodsell/go-links-free/internal/models"
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
	router := gin.Default()
	// ... rest of your route setup
	return router
}
