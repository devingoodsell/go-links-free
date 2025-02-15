package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"url"

	"github.com/gorilla/mux"
	"github.com/yourusername/go-links/internal/auth"
	"github.com/yourusername/go-links/internal/models"
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
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destination_url"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
}

type linkResponse struct {
	ID             int64      `json:"id"`
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destination_url"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Stats          *models.LinkStats `json:"stats,omitempty"`
}

type listResponse struct {
	Links      []linkResponse `json:"links"`
	TotalCount int           `json:"total_count"`
	HasMore    bool          `json:"has_more"`
	NextOffset int           `json:"next_offset,omitempty"`
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

func (h *LinkHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := validateCreateLinkRequest(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get user from context
	claims := r.Context().Value("user").(*auth.Claims)

	link := &models.Link{
		Alias:          req.Alias,
		DestinationURL: req.DestinationURL,
		CreatedBy:      claims.UserID,
		ExpiresAt:      req.ExpiresAt,
	}

	if err := h.linkRepo.Create(r.Context(), link); err != nil {
		// Check for duplicate alias
		if strings.Contains(err.Error(), "unique constraint") {
			http.Error(w, "alias already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, linkResponse{
		ID:             link.ID,
		Alias:          link.Alias,
		DestinationURL: link.DestinationURL,
		ExpiresAt:      link.ExpiresAt,
		CreatedAt:      link.CreatedAt,
		UpdatedAt:      link.UpdatedAt,
	})
}

func (h *LinkHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]

	link, err := h.linkRepo.GetByAlias(r.Context(), alias)
	if err != nil {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}

	// Check if link is expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		// Return a special response for expired links
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusGone)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "This link has expired",
			"destination_url": link.DestinationURL,
			"expired_at":      link.ExpiresAt,
		})
		return
	}

	// Increment stats asynchronously
	go h.linkRepo.IncrementStats(r.Context(), link.ID)

	// Perform redirect
	http.Redirect(w, r, link.DestinationURL, http.StatusTemporaryRedirect)
}

func (h *LinkHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := getUserIDFromContext(r.Context())

	opts := models.ListOptions{
		Limit:  getIntQueryParam(r, "pageSize", 10),
		Offset: getIntQueryParam(r, "page", 0) * getIntQueryParam(r, "pageSize", 10),
		Search: r.URL.Query().Get("search"),
		Status: r.URL.Query().Get("status"),
		SortBy: r.URL.Query().Get("sortBy"),
	}

	links, err := h.linkRepo.ListByUserWithFilters(r.Context(), userID, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, links)
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

func (h *LinkHandler) Update(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]
	claims := r.Context().Value("user").(*auth.Claims)

	var req createLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	link, err := h.linkRepo.GetByAlias(r.Context(), alias)
	if err != nil {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}

	// Verify ownership or admin status
	if link.CreatedBy != claims.UserID && !claims.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Update link
	link.DestinationURL = req.DestinationURL
	link.ExpiresAt = req.ExpiresAt

	if err := h.linkRepo.Update(r.Context(), link); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, linkResponse{
		ID:             link.ID,
		Alias:          link.Alias,
		DestinationURL: link.DestinationURL,
		ExpiresAt:      link.ExpiresAt,
		CreatedAt:      link.CreatedAt,
		UpdatedAt:      link.UpdatedAt,
		Stats:          link.Stats,
	})
}

func (h *LinkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	alias := mux.Vars(r)["alias"]
	claims := r.Context().Value("user").(*auth.Claims)

	// Get the link first to check ownership
	link, err := h.linkRepo.GetByAlias(r.Context(), alias)
	if err != nil {
		http.Error(w, "link not found", http.StatusNotFound)
		return
	}

	// Verify ownership or admin status
	if link.CreatedBy != claims.UserID && !claims.IsAdmin {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := h.linkRepo.Delete(r.Context(), link.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LinkHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	var req bulkActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "no links specified", http.StatusBadRequest)
		return
	}

	if err := h.linkRepo.BulkDelete(r.Context(), req.IDs); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *LinkHandler) BulkUpdateStatus(w http.ResponseWriter, r *http.Request) {
	var req bulkStatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.IDs) == 0 {
		http.Error(w, "no links specified", http.StatusBadRequest)
		return
	}

	if err := h.linkRepo.BulkUpdateStatus(r.Context(), req.IDs, req.IsActive); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
} 