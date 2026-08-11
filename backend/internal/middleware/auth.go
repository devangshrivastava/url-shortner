package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"url-shortener/internal/service"
)

type AuthMiddleware struct {
	jwtService *service.JWTService
}

func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtService: jwtService}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Fields(authorization)
		if len(parts) != 2 || parts[0] != "Bearer" {
			unauthorized(c)
			return
		}

		userID, err := m.jwtService.ParseAndValidateToken(parts[1])
		if err != nil {
			unauthorized(c)
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.Abort()
	c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
