package models

import (
	"time"
)

type User struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	IsAdmin      bool       `json:"isAdmin"`
	CreatedAt    time.Time  `json:"createdAt,omitempty"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
}
