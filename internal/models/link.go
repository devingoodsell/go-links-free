package models

import (
	"time"
)

type LinkStats struct {
	DailyCount     int        `json:"daily_count"`
	WeeklyCount    int        `json:"weekly_count"`
	TotalCount     int        `json:"total_count"`
	LastAccessedAt *time.Time `json:"lastAccessedAt,omitempty"`
}

type Link struct {
	ID             int64      `json:"id"`
	Alias          string     `json:"alias"`
	DestinationURL string     `json:"destinationUrl"`
	CreatedBy      int64      `json:"createdBy"`
	ExpiresAt      *time.Time `json:"expiresAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	IsActive       bool       `json:"isActive"`
	Stats          *LinkStats `json:"stats,omitempty"`
}

type ListOptions struct {
	Search   string `json:"search"`
	Status   string `json:"status"`
	SortBy   string `json:"sort_by"`
	Domain   string `json:"domain"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}
