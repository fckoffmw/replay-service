package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/fckoffmw/replay-service/server/config"
	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/fckoffmw/replay-service/server/internal/handlers"
	"github.com/fckoffmw/replay-service/server/internal/logger"
	"github.com/fckoffmw/replay-service/server/internal/middleware"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.NewSlog(cfg.LogLevel)
	_ = logger

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Successfully connected to database")

	replayRepo := repository.NewReplayRepository(db)
	handler := handlers.NewHandler(replayRepo, cfg.StorageDir)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/games", handler.GetGames)
		api.POST("/games", handler.CreateGame)
		api.DELETE("/games/:game_id", handler.DeleteGame)

		api.GET("/games/:game_id/replays", handler.GetReplays)
		api.POST("/games/:game_id/replays", handler.CreateReplay)

		api.GET("/replays/:replay_id", handler.GetReplay)
		api.PUT("/replays/:replay_id", handler.UpdateReplay)
		api.DELETE("/replays/:replay_id", handler.DeleteReplay)
		api.GET("/replays/:replay_id/file", handler.GetReplayFile)
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
