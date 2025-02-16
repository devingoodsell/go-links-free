package models

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrNotFound     = errors.New("not found")
	ErrDuplicate    = errors.New("duplicate entry")
)
