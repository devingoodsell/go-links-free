package handlers

import "github.com/gin-gonic/gin"

// HealthCheck handles health check requests
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
	})
}
