package handlers

import (
	"net/http"
	"time"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/gin-gonic/gin"
)

func AddAuthRoutes(router *gin.Engine, authService *auth.AuthService) {
	authHandler := NewAuthHandler(authService)

	// Public auth routes
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)

		// OKTA routes (if enabled)
		authGroup.GET("/okta/login", authHandler.OktaLogin)
		authGroup.GET("/okta/callback", authHandler.OktaCallback)
	}
}

type AuthHandler struct {
	authService *auth.AuthService
}

func NewAuthHandler(authService *auth.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) OktaLogin(c *gin.Context) {
	state := generateState() // You'll need to implement this
	url, err := h.authService.GetOktaAuthURL(state)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store state in cookie for validation in callback
	c.SetCookie("okta_state", state, int(5*time.Minute), "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"auth_url": url})
}

func (h *AuthHandler) OktaCallback(c *gin.Context) {
	state, err := c.Cookie("okta_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	if state != c.Query("state") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "state mismatch"})
		return
	}

	// Clear state cookie
	c.SetCookie("okta_state", "", -1, "/", "", true, true)

	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no code provided"})
		return
	}

	token, err := h.authService.HandleOktaCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// Helper function to generate a secure random state
func generateState() string {
	// Implementation from previous auth.go
	return "random-state" // Replace with actual implementation
}
