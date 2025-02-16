package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/devingoodsell/go-links-free/internal/auth"
	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/gin-gonic/gin"
)

type LinkHandler struct {
	linkRepo *models.LinkRepository
}

func NewLinkHandler(linkRepo *models.LinkRepository) *LinkHandler {
	return &LinkHandler{
		linkRepo: linkRepo,
	}
}

type createLinkRequest struct {
	Alias          string     `json:"alias" binding:"required"`
	DestinationURL string     `json:"destinationUrl" binding:"required,url"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

type updateLinkRequest struct {
	DestinationURL string     `json:"destinationUrl" binding:"required,url"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
}

type linkResponse struct {
	ID             int64             `json:"id"`
	Alias          string            `json:"alias"`
	DestinationURL string            `json:"destination_url"`
	ExpiresAt      *time.Time        `json:"expires_at,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	Stats          *models.LinkStats `json:"stats,omitempty"`
}

type listResponse struct {
	Links      []linkResponse `json:"links"`
	TotalCount int            `json:"total_count"`
	HasMore    bool           `json:"has_more"`
	NextOffset int            `json:"next_offset,omitempty"`
}

// Add new bulk operation handlers
type bulkActionRequest struct {
	IDs []int64 `json:"ids"`
}

type bulkStatusUpdateRequest struct {
	IDs      []int64 `json:"ids"`
	IsActive bool    `json:"isActive"`
}

func validateCreateLinkRequest(req *createLinkRequest) error {
	if req.Alias == "" {
		return errors.New("alias is required")
	}
	if len(req.Alias) > 100 {
		return errors.New("alias must be 100 characters or less")
	}
	if req.DestinationURL == "" {
		return errors.New("destination URL is required")
	}
	if _, err := url.Parse(req.DestinationURL); err != nil {
		return errors.New("invalid destination URL")
	}
	if req.ExpiresAt != nil && req.ExpiresAt.Before(time.Now()) {
		return errors.New("expiration time must be in the future")
	}
	return nil
}

func (h *LinkHandler) Create(c *gin.Context) {
	var req createLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user from context
	userClaims, _ := c.Get("user")
	claims := userClaims.(*auth.Claims)

	link := &models.Link{
		Alias:          req.Alias,
		DestinationURL: req.DestinationURL,
		CreatedBy:      claims.UserID,
		ExpiresAt:      req.ExpiresAt,
		IsActive:       true,
	}

	if err := h.linkRepo.Create(c.Request.Context(), link); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	alias := c.Param("alias")
	link, err := h.linkRepo.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		c.JSON(404, gin.H{"error": "link not found"})
		return
	}

	if link.ExpiresAt != nil && link.ExpiresAt.Before(time.Now()) {
		c.JSON(410, gin.H{"error": "link has expired"})
		return
	}

	if err := h.linkRepo.IncrementStats(c.Request.Context(), link.ID); err != nil {
		// Log error but don't fail the redirect
		// TODO: Add proper logging
	}

	c.Redirect(302, link.DestinationURL)
}

func (h *LinkHandler) List(c *gin.Context) {
	// Get user from context
	userClaims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found in context"})
		return
	}
	claims := userClaims.(*auth.Claims)

	// Parse pagination params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "0"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// Get links for user
	links, total, err := h.linkRepo.ListForUser(c.Request.Context(), claims.UserID, page, pageSize)
	if err != nil {
		log.Printf("Error listing links: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(links))
	for i, link := range links {
		log.Printf("Link %d: createdAt=%v, isActive=%v", i, link.CreatedAt, link.IsActive) // More specific debug
		response[i] = map[string]interface{}{
			"id":             link.ID,
			"alias":          link.Alias,
			"destinationUrl": link.DestinationURL,
			"createdAt":      link.CreatedAt.Format(time.RFC3339), // Format the time explicitly
			"updatedAt":      link.UpdatedAt.Format(time.RFC3339),
			"expiresAt":      link.ExpiresAt,
			"isActive":       link.IsActive,
			"stats":          link.Stats,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"items":      response,
		"totalCount": total,
	})
}

func getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	valueStr := r.URL.Query().Get(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func (h *LinkHandler) Update(c *gin.Context) {
	log.Printf("Update request - URL: %s, Params: %v", c.Request.URL.Path, c.Params)
	log.Printf("Update request body: %s", c.Request.Body)

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		c.JSON(400, gin.H{"error": "invalid link ID"})
		return
	}

	var req updateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(400, gin.H{"error": "invalid request body"})
		return
	}

	log.Printf("Update request for ID %d with data: %+v", id, req)

	// Get the link first to verify ownership
	link, err := h.linkRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "link not found"})
		return
	}

	// Verify ownership
	if link.CreatedBy != getUserIDFromContext(c) {
		c.JSON(403, gin.H{"error": "unauthorized"})
		return
	}

	link.DestinationURL = req.DestinationURL
	link.ExpiresAt = req.ExpiresAt

	if err := h.linkRepo.Update(c.Request.Context(), link); err != nil {
		c.JSON(500, gin.H{"error": "failed to update link"})
		return
	}

	c.JSON(200, link)
}

func (h *LinkHandler) Delete(c *gin.Context) {
	log.Printf("Delete request - URL: %s, Params: %v", c.Request.URL.Path, c.Params)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		c.JSON(400, gin.H{"error": "invalid link ID"})
		return
	}

	userID := getUserIDFromContext(c)
	log.Printf("Attempting to delete link %d for user %d", id, userID)
	if err := h.linkRepo.Delete(c.Request.Context(), id, userID); err != nil {
		log.Printf("Error deleting link: %v", err)
		c.JSON(500, gin.H{"error": "failed to delete link"})
		return
	}

	c.Status(204)
}

func (h *LinkHandler) BulkDelete(c *gin.Context) {
	var req struct {
		IDs []int64 `json:"ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"error": "no links specified"})
		return
	}

	userID := getUserIDFromContext(c)
	if err := h.linkRepo.BulkDelete(c.Request.Context(), userID, req.IDs); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

func (h *LinkHandler) BulkUpdateStatus(c *gin.Context) {
	var req struct {
		IDs      []int64 `json:"ids"`
		IsActive bool    `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	if len(req.IDs) == 0 {
		c.JSON(400, gin.H{"error": "no links specified"})
		return
	}

	userID := getUserIDFromContext(c)
	if err := h.linkRepo.BulkUpdateStatus(c.Request.Context(), userID, req.IDs, req.IsActive); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.Status(204)
}

func getUserIDFromContext(c *gin.Context) int64 {
	if claims, exists := c.Get("user"); exists {
		if userClaims, ok := claims.(*auth.Claims); ok {
			return userClaims.UserID
		}
	}
	return 0
}

func (h *LinkHandler) getLinkFromRequest(c *gin.Context) (*models.Link, error) {
	alias := c.Param("alias")
	link, err := h.linkRepo.GetByAlias(c.Request.Context(), alias)
	if err != nil {
		return nil, errors.New("link not found")
	}

	// Verify ownership
	userID := getUserIDFromContext(c)
	if link.CreatedBy != userID {
		return nil, models.ErrUnauthorized
	}

	return link, nil
}

func (h *LinkHandler) GetStats(c *gin.Context) {
	link, err := h.getLinkFromRequest(c)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if link.Stats == nil {
		c.JSON(200, gin.H{
			"daily_count":  0,
			"weekly_count": 0,
			"total_count":  0,
		})
		return
	}

	c.JSON(200, gin.H{
		"daily_count":  link.Stats.DailyCount,
		"weekly_count": link.Stats.WeeklyCount,
		"total_count":  link.Stats.TotalCount,
	})
}
