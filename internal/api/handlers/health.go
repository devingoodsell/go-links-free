package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	}

	// Check database connection
	if err := h.db.Ping(); err != nil {
		health["status"] = "unhealthy"
		health["database"] = "disconnected"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	health["database"] = "connected"
	c.JSON(http.StatusOK, health)
}

func AddHealthRoutes(router *gin.Engine) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}
