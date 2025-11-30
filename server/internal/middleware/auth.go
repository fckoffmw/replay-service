package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	headerUserID     = "X-User-ID"
	contextKeyUserID = "user_id"
	defaultUserID    = "00000000-0000-0000-0000-000000000001"
)

func AuthMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader(headerUserID)
		if userIDStr == "" {
			userIDStr = defaultUserID
			logger.Debug("no user_id header, using default",
				slog.String("default_user_id", userIDStr))
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			logger.Error("invalid user_id",
				slog.String("user_id", userIDStr),
				slog.String("error", err.Error()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
			c.Abort()
			return
		}

		logger.Info("authenticated request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.String("user_id", userID.String()))

		c.Set(contextKeyUserID, userID)
		c.Next()
	}
}
