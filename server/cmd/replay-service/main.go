package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Mock: список реплеев
	r.GET("/replays", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{"id": "11111111-1111-1111-1111-111111111111", "original_name": "example.rep", "compression": "gzip", "compressed": true},
		})
	})

	// Mock: получить реплей
	r.GET("/replays/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{"id": id, "url": "/download/not-implemented"})
	})

	// Mock: загрузка
	r.POST("/replays/upload", func(c *gin.Context) {
		// В мок-версии просто отвечаем успехом
		c.JSON(http.StatusCreated, gin.H{"id": "22222222-2222-2222-2222-222222222222"})
	})

	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
