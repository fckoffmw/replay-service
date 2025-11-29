package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			userIDStr = "00000000-0000-0000-0000-000000000001"
			log.Printf("[AUTH] No X-User-ID header, using default: %s", userIDStr)
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Printf("[AUTH] ERROR: invalid user_id: %s", userIDStr)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id"})
			c.Abort()
			return
		}

		log.Printf("[AUTH] %s %s - user_id=%s", c.Request.Method, c.Request.URL.Path, userID)
		c.Set("user_id", userID)
		c.Next()
	}
}
