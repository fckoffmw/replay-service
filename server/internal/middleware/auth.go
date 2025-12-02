package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/fckoffmw/replay-service/server/internal/services"
	"github.com/gin-gonic/gin"
)

const (
	contextKeyUserID = "user_id"
)

func AuthMiddleware(authService *services.AuthService, logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			logger.Warn("missing authorization token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		logger.Debug("extracted token", slog.String("token_preview", token[:min(20, len(token))]+"..."))

		userID, err := authService.ValidateToken(token)
		if err != nil {
			logger.Warn("invalid token",
				slog.String("error", err.Error()),
				slog.String("token_preview", token[:min(20, len(token))]))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный или истекший токен"})
			c.Abort()
			return
		}

		logger.Info("authenticated request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("user_id", userID.String()))

		c.Set(contextKeyUserID, *userID)
		c.Next()
	}
}

func extractToken(c *gin.Context) string {
	// Try query parameter first (for video/file requests)
	if token := c.Query("token"); token != "" {
		return token
	}

	// Try Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
