package auth

import (
	"crypto/rand"
	"encoding/base64"
)

func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
} 