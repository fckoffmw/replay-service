package handlers

import (
	"github.com/fckoffmw/replay-service/server/internal/services"
)

type Handler struct {
	gameService   *services.GameService
	replayService *services.ReplayService
}

func NewHandler(gameService *services.GameService, replayService *services.ReplayService) *Handler {
	return &Handler{
		gameService:   gameService,
		replayService: replayService,
	}
}
