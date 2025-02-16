package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Add a Gin-specific version
func WriteJSONGin(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}
