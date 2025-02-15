package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/go-links/internal/models"
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

func (h *AdminHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.analyticsRepo.GetSystemStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, stats)
}

func (h *AdminHandler) GetRedirectsOverTime(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}

	data, err := h.analyticsRepo.GetRedirectsOverTime(r.Context(), period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, data)
}

// Add admin-specific link management endpoints
func (h *AdminHandler) ListAllLinks(w http.ResponseWriter, r *http.Request) {
	// Similar to regular list but without user filtering
	// TODO: Implement with pagination and filtering options
}

func (h *AdminHandler) UpdateLinkAdmin(w http.ResponseWriter, r *http.Request) {
	// Similar to regular update but with additional admin capabilities
	// TODO: Implement
}

func (h *AdminHandler) GetPopularLinks(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "daily"
	}

	limit := 10 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 100 { // Maximum limit
				limit = 100
			}
		}
	}

	links, err := h.analyticsRepo.GetPopularLinks(r.Context(), limit, period)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, links)
}

func (h *AdminHandler) GetUserActivity(w http.ResponseWriter, r *http.Request) {
	days := 30 // Default to last 30 days
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if parsedDays, err := strconv.Atoi(daysStr); err == nil && parsedDays > 0 {
			days = parsedDays
			if days > 365 { // Maximum lookback period
				days = 365
			}
		}
	}

	activities, err := h.analyticsRepo.GetUserActivity(r.Context(), days)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, activities)
}

func (h *AdminHandler) GetTopDomains(w http.ResponseWriter, r *http.Request) {
	limit := 10 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
			if limit > 50 { // Maximum limit
				limit = 50
			}
		}
	}

	stats, err := h.analyticsRepo.GetTopDomains(r.Context(), limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, stats)
}

func (h *AdminHandler) GetPeakUsage(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	var date time.Time
	var err error

	if dateStr == "" {
		date = time.Now()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "invalid date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
	}

	stats, err := h.analyticsRepo.GetPeakUsage(r.Context(), date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, stats)
}

func (h *AdminHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	window := r.URL.Query().Get("window")
	if window == "" {
		window = "hour"
	}

	if window != "hour" && window != "day" && window != "week" && window != "month" {
		http.Error(w, "invalid window (use hour, day, week, or month)", http.StatusBadRequest)
		return
	}

	metrics, err := h.analyticsRepo.GetPerformanceMetrics(r.Context(), window)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, metrics)
} 