package main

import (
	"context"
	"fmt"
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

const (
	API_V1_PATH         = "/api/v1"
	API_V1_GAMES_PATH   = API_V1_PATH + "/games"
	API_V1_REPLAYS_PATH = API_V1_PATH + "/replays"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logger.NewSlog(cfg.LogLevel)
	logger.Info(fmt.Sprintf("CONFIG=  %s", cfg.String()))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	logger.Info("Successfully connected to database")

	replayRepo := repository.NewReplayRepository(db)
	replayHandler := handlers.NewHandler(replayRepo, cfg.StorageDir)

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

	gamesAPI := r.Group(API_V1_GAMES_PATH)
	gamesAPI.Use(middleware.AuthMiddleware())
	{
		gamesAPI.GET("", replayHandler.GetGames)
		gamesAPI.POST("", replayHandler.CreateGame)
		gamesAPI.PUT("/:game_id", replayHandler.UpdateGame)
		gamesAPI.DELETE("/:game_id", replayHandler.DeleteGame)

		gamesAPI.GET("/:game_id/replays", replayHandler.GetReplays)
		gamesAPI.POST("/:game_id/replays", replayHandler.CreateReplay)
	}

	replaysAPI := r.Group(API_V1_REPLAYS_PATH)
	replaysAPI.Use(middleware.AuthMiddleware())
	{
		replaysAPI.GET("/:replay_id", replayHandler.GetReplay)
		replaysAPI.PUT("/:replay_id", replayHandler.UpdateReplay)
		replaysAPI.DELETE("/:replay_id", replayHandler.DeleteReplay)
		replaysAPI.GET("/:replay_id/file", replayHandler.GetReplayFile)
	}

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
