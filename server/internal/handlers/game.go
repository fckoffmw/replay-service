package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	contextKeyUserID = "user_id"
	paramGameID      = "game_id"
)

func (h *Handler) GetGames(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)

	games, err := h.gameService.GetUserGames(c.Request.Context(), userID)
	if err != nil {
		respondInternalError(c, "failed to get games")
		return
	}

	respondOK(c, games)
}

func (h *Handler) CreateGame(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "name is required")
		return
	}

	game, err := h.gameService.CreateGame(c.Request.Context(), userID, req.Name)
	if err != nil {
		respondInternalError(c, "failed to create game")
		return
	}

	respondCreated(c, game)
}

func (h *Handler) UpdateGame(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	gameID, err := uuid.Parse(c.Param(paramGameID))
	if err != nil {
		respondBadRequest(c, "invalid game_id")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		respondBadRequest(c, "name is required")
		return
	}

	if err := h.gameService.UpdateGame(c.Request.Context(), gameID, userID, req.Name); err != nil {
		respondInternalError(c, "failed to update game")
		return
	}

	respondSuccess(c, "updated")
}

func (h *Handler) DeleteGame(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	gameID, err := uuid.Parse(c.Param(paramGameID))
	if err != nil {
		respondBadRequest(c, "invalid game_id")
		return
	}

	if err := h.gameService.DeleteGame(c.Request.Context(), gameID, userID); err != nil {
		respondInternalError(c, "failed to delete game")
		return
	}

	respondSuccess(c, "deleted")
}
