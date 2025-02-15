package models

import (
	"time"
)

type Link struct {
	ID             int64      `json:"id"`
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destination_url"`
	CreatedBy      int64      `json:"created_by"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	Stats          *LinkStats `json:"stats,omitempty"`
}

type LinkStats struct {
	ID             int64      `json:"id"`
	LinkID         int64      `json:"link_id"`
	DailyCount     int        `json:"daily_count"`
	WeeklyCount    int        `json:"weekly_count"`
	TotalCount     int        `json:"total_count"`
	LastAccessedAt *time.Time `json:"last_accessed_at,omitempty"`
} 