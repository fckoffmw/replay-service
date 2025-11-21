package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fckoffmw/replay-service/server/config"
	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/fckoffmw/replay-service/server/internal/logger"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфигурацию из .env файла
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.NewSlog(cfg.LogLevel)
	logger.Info("logger created!")
	logger.Debug("AHAHHA")

	// Подключаемся к базе данных
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	// Создаем репозиторий для работы с реплеями
	replayRepo := repository.NewReplayRepository(db)

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Список реплеев из базы данных
	r.GET("/replays", func(c *gin.Context) {
		replays, err := replayRepo.GetAll(c.Request.Context())
		if err != nil {
			log.Printf("Failed to get replays: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get replays"})
			return
		}

		c.JSON(http.StatusOK, replays)
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

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
