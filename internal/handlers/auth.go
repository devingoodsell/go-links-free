package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"log"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *auth.AuthService
	userRepo    *models.UserRepository
}

func NewAuthHandler(authService *auth.AuthService, userRepo *models.UserRepository) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userRepo:    userRepo,
	}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var loginReq loginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user first to update last login
	user, err := h.authService.GetUserByEmail(c.Request.Context(), loginReq.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Attempt login
	token, err := h.authService.Login(c.Request.Context(), loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Update last login time
	now := time.Now().UTC()

	if err := h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID, &now); err != nil {
		log.Printf("Failed to update last login time for user %d", user.ID)
		// Don't fail the login if this fails
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	token, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, authResponse{Token: token})
}

// OKTA SSO handlers
func (h *AuthHandler) OktaLogin(c *gin.Context) {
	state := generateState()
	url, err := h.authService.GetOktaAuthURL(state)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("okta_state", state, 300, "/", "", true, true)
	c.JSON(200, gin.H{"auth_url": url})
}

func (h *AuthHandler) OktaCallback(c *gin.Context) {
	stateCookie, err := c.Cookie("okta_state")
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid state"})
		return
	}

	state := c.Query("state")
	if state != stateCookie {
		c.JSON(400, gin.H{"error": "state mismatch"})
		return
	}

	c.SetCookie("okta_state", "", -1, "/", "", true, true)

	code := c.Query("code")
	if code == "" {
		c.JSON(400, gin.H{"error": "no code provided"})
		return
	}

	token, err := h.authService.HandleOktaCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, authResponse{Token: token})
}

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GetCurrentUser handles the /api/auth/me endpoint
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user claims from context (set by auth middleware)
	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	userClaims, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user claims"})
		return
	}

	// Get full user details from database
	user, err := h.userRepo.GetByEmail(c.Request.Context(), userClaims.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user details"})
		return
	}

	c.JSON(http.StatusOK, user)
}
