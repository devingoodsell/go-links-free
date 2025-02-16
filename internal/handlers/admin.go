package handlers

import (
	"strconv"
	"time"

	"github.com/devingoodsell/go-links-free/internal/models"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	analyticsRepo *models.AnalyticsRepository
	linkRepo      *models.LinkRepository
	userRepo      *models.UserRepository
}

func NewAdminHandler(
	analyticsRepo *models.AnalyticsRepository,
	linkRepo *models.LinkRepository,
	userRepo *models.UserRepository,
) *AdminHandler {
	return &AdminHandler{
		analyticsRepo: analyticsRepo,
		linkRepo:      linkRepo,
		userRepo:      userRepo,
	}
}

func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	stats, err := h.analyticsRepo.GetSystemStats(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, stats)
}

func (h *AdminHandler) GetRedirectsOverTime(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "daily"
	}

	data, err := h.analyticsRepo.GetRedirectsOverTime(c.Request.Context(), period)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, data)
}

// Add admin-specific link management endpoints
func (h *AdminHandler) ListAllLinks(c *gin.Context) {
	// TODO: Implement with pagination and filtering options
	c.JSON(501, gin.H{"error": "not implemented"})
}

func (h *AdminHandler) UpdateLinkAdmin(c *gin.Context) {
	// TODO: Implement
	c.JSON(501, gin.H{"error": "not implemented"})
}

func (h *AdminHandler) GetPopularLinks(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		period = "daily"
	}

	limit := 10 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 { // Maximum limit
				limit = 100
			}
		}
	}

	links, err := h.analyticsRepo.GetPopularLinks(c.Request.Context(), limit, period)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, links)
}

func (h *AdminHandler) GetUserActivity(c *gin.Context) {
	days := 30 // Default to last 30 days
	if daysStr := c.Query("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
			if days > 365 { // Maximum lookback period
				days = 365
			}
		}
	}

	activities, err := h.analyticsRepo.GetUserActivity(c.Request.Context(), days)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, activities)
}

func (h *AdminHandler) GetTopDomains(c *gin.Context) {
	limit := 10 // Default limit
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 50 { // Maximum limit
				limit = 50
			}
		}
	}

	stats, err := h.analyticsRepo.GetTopDomains(c.Request.Context(), limit)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stats)
}

func (h *AdminHandler) GetPeakUsage(c *gin.Context) {
	dateStr := c.Query("date")
	var date time.Time
	var err error

	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid date format (use YYYY-MM-DD)"})
			return
		}
	}

	stats, err := h.analyticsRepo.GetPeakUsage(c.Request.Context(), date)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, stats)
}

func (h *AdminHandler) GetPerformanceMetrics(c *gin.Context) {
	window := c.Query("window")
	if window == "" {
		window = "hour"
	}

	if window != "hour" && window != "day" && window != "week" && window != "month" {
		c.JSON(400, gin.H{"error": "invalid window (use hour, day, week, or month)"})
		return
	}

	metrics, err := h.analyticsRepo.GetPerformanceMetrics(c.Request.Context(), window)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, metrics)
}

func (h *AdminHandler) GetUsers(c *gin.Context) {
	// ... existing code ...
}

func (h *AdminHandler) GetUser(c *gin.Context) {
	// ... existing code ...
}
