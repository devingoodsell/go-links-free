package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/devingoodsell/go-links-free/internal/services"
	"github.com/gin-gonic/gin"
)

func AddLinkRoutes(router *gin.Engine, linkService *services.LinkService, authMiddleware gin.HandlerFunc) {
	linkHandler := NewLinkHandler(linkService)

	// Public redirect endpoint
	router.GET("/go/:alias", linkHandler.Redirect)

	// Protected routes
	protected := router.Group("/api/links")
	protected.Use(authMiddleware)
	{
		protected.POST("", linkHandler.Create)
		protected.GET("", linkHandler.List)
		protected.PUT("/:alias", linkHandler.Update)
		protected.DELETE("/:alias", linkHandler.Delete)
		protected.GET("/:alias/stats", linkHandler.GetStats)
		protected.POST("/bulk-delete", linkHandler.BulkDelete)
		protected.POST("/bulk-update-status", linkHandler.BulkUpdateStatus)
	}
}

type LinkHandler struct {
	linkService *services.LinkService
}

func NewLinkHandler(linkService *services.LinkService) *LinkHandler {
	return &LinkHandler{
		linkService: linkService,
	}
}

// Handler implementations...

func (h *LinkHandler) Create(c *gin.Context) {
	var req struct {
		Alias          string     `json:"alias"`
		DestinationURL string     `json:"destination_url"`
		ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := getUserIDFromContext(c)
	link, err := h.linkService.Create(c.Request.Context(), userID, req.Alias, req.DestinationURL, req.ExpiresAt)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicate):
			c.JSON(http.StatusConflict, gin.H{"error": "alias already exists"})
		case errors.Is(err, models.ErrUnauthorized):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	alias := c.Param("alias")
	link, err := h.linkService.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusGone, gin.H{"error": "link has expired"})
		return
	}

	// Increment stats
	if err := h.linkService.IncrementStats(c.Request.Context(), link.ID); err != nil {
		// Log error but don't fail the redirect
		log.Printf("Failed to increment stats: %v", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, link.DestinationURL)
}

func (h *LinkHandler) List(c *gin.Context) {
	userID := getUserIDFromContext(c)
	opts := models.ListOptions{
		Search: c.Query("search"),
		Status: c.Query("status"),
		SortBy: c.Query("sort"),
		Domain: c.Query("domain"),
	}

	links, err := h.linkService.List(c.Request.Context(), userID, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"links": links})
}

func (h *LinkHandler) Update(c *gin.Context) {
	var req struct {
		DestinationURL string     `json:"destination_url"`
		ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	userID := getUserIDFromContext(c)
	alias := c.Param("alias")

	link, err := h.linkService.Update(c.Request.Context(), userID, alias, req.DestinationURL, req.ExpiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, link)
}

func (h *LinkHandler) Delete(c *gin.Context) {
	userID := getUserIDFromContext(c)
	alias := c.Param("alias")

	if err := h.linkService.Delete(c.Request.Context(), userID, alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LinkHandler) GetStats(c *gin.Context) {
	alias := c.Param("alias")
	link, err := h.linkService.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	if link.Stats == nil {
		c.JSON(http.StatusOK, gin.H{
			"daily_count":  0,
			"weekly_count": 0,
			"total_count":  0,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"daily_count":  link.Stats.DailyCount,
		"weekly_count": link.Stats.WeeklyCount,
		"total_count":  link.Stats.TotalCount,
	})
}

func (h *LinkHandler) BulkDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no links specified"})
		return
	}

	userID := getUserIDFromContext(c)
	if err := h.linkService.BulkDelete(c.Request.Context(), userID, req.IDs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LinkHandler) BulkUpdateStatus(c *gin.Context) {
	var req struct {
		IDs      []int64 `json:"ids"`
		IsActive bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no links specified"})
		return
	}

	userID := getUserIDFromContext(c)
	if err := h.linkService.BulkUpdateStatus(c.Request.Context(), userID, req.IDs, req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func getUserIDFromContext(c *gin.Context) int64 {
	if claims, exists := c.Get("user"); exists {
		if userClaims, ok := claims.(*auth.Claims); ok {
			return userClaims.UserID
		}
	}
	return 0
}
