package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/fckoffmw/replay-service/server/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler is the controller layer that handles HTTP requests
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

// GetGames handles GET /games
func (h *Handler) GetGames(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	games, err := h.gameService.GetUserGames(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get games"})
		return
	}

	c.JSON(http.StatusOK, games)
}

// GetReplays handles GET /games/:game_id/replays
func (h *Handler) GetReplays(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	replays, err := h.replayService.GetGameReplays(c.Request.Context(), gameID, userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get replays"})
		return
	}

	c.JSON(http.StatusOK, replays)
}

// GetReplay handles GET /replays/:replay_id
func (h *Handler) GetReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	replay, err := h.replayService.GetReplay(c.Request.Context(), replayID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	c.JSON(http.StatusOK, replay)
}

// CreateReplay handles POST /games/:game_id/replays
func (h *Handler) CreateReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	title := c.PostForm("title")
	comment := c.PostForm("comment")

	replay, err := h.replayService.CreateReplay(c.Request.Context(), file, gameID, userID, title, comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create replay"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": replay.ID})
}

// DeleteReplay handles DELETE /replays/:replay_id
func (h *Handler) DeleteReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	if err := h.replayService.DeleteReplay(c.Request.Context(), replayID, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// DeleteGame handles DELETE /games/:game_id
func (h *Handler) DeleteGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	if err := h.gameService.DeleteGame(c.Request.Context(), gameID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// UpdateReplay handles PUT /replays/:replay_id
func (h *Handler) UpdateReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	title := c.PostForm("title")
	comment := c.PostForm("comment")

	var titlePtr, commentPtr *string
	if title != "" {
		titlePtr = &title
	}
	if comment != "" {
		commentPtr = &comment
	}

	if err := h.replayService.UpdateReplay(c.Request.Context(), replayID, userID, titlePtr, commentPtr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update replay"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// GetReplayFile handles GET /replays/:replay_id/file
func (h *Handler) GetReplayFile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	replay, err := h.replayService.GetReplay(c.Request.Context(), replayID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	fullPath, ext, err := h.replayService.GetReplayFilePath(c.Request.Context(), replayID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer file.Close()

	contentType := getContentType(ext)
	
	if isVideoFile(ext) {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", replay.OriginalName))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", replay.OriginalName))
	}
	
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=31536000")
	
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}
	
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	io.Copy(c.Writer, file)
}

func getContentType(ext string) string {
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".ogg", ".ogv":
		return "video/ogg"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	case ".m4v":
		return "video/x-m4v"
	default:
		return "application/octet-stream"
	}
}

func isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".webm", ".ogg", ".ogv", ".mov", ".avi", ".mkv", ".m4v"}
	for _, ve := range videoExts {
		if ext == ve {
			return true
		}
	}
	return false
}

// CreateGame handles POST /games
func (h *Handler) CreateGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	game, err := h.gameService.CreateGame(c.Request.Context(), userID, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game"})
		return
	}

	c.JSON(http.StatusCreated, game)
}

// UpdateGame handles PUT /games/:game_id
func (h *Handler) UpdateGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	if err := h.gameService.UpdateGame(c.Request.Context(), gameID, userID, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
