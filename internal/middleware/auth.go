package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/devingoodsell/go-links-free/internal/auth"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) AuthenticateGin(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}

	// Extract token from Bearer scheme
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid authorization header"})
		return
	}

	claims, err := m.jwtManager.ValidateToken(tokenParts[1])
	if err != nil {
		c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
		return
	}

	// Add claims to context
	c.Set("user", claims)
	c.Next()
}

func (m *AuthMiddleware) RequireAdminGin(c *gin.Context) {
	claims, exists := c.Get("user")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}

	userClaims, ok := claims.(*auth.Claims)
	if !ok || !userClaims.IsAdmin {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		return
	}

	c.Next()
}
