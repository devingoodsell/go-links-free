package handlers

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *sql.DB  // Your database connection
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	health := gin.H{
		"status": "healthy",
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