package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	repo       *repository.ReplayRepository
	storageDir string
}

func NewHandler(repo *repository.ReplayRepository, storageDir string) *Handler {
	log.Printf("[HANDLER] Initialized with storage dir: %s", storageDir)
	return &Handler{
		repo:       repo,
		storageDir: storageDir,
	}
}

func (h *Handler) GetGames(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	log.Printf("[GET /games] user_id=%s", userID)

	games, err := h.repo.GetGamesByUserID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[GET /games] ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get games"})
		return
	}

	log.Printf("[GET /games] SUCCESS: found %d games", len(games))
	c.JSON(http.StatusOK, games)
}

func (h *Handler) GetReplays(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[GET /games/:game_id/replays] ERROR: invalid game_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	log.Printf("[GET /games/%s/replays] user_id=%s, limit=%d", gameID, userID, limit)

	replays, err := h.repo.GetReplaysByGameID(c.Request.Context(), gameID, userID, limit)
	if err != nil {
		log.Printf("[GET /games/%s/replays] ERROR: %v", gameID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get replays"})
		return
	}

	log.Printf("[GET /games/%s/replays] SUCCESS: found %d replays", gameID, len(replays))
	c.JSON(http.StatusOK, replays)
}

func (h *Handler) GetReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[GET /replays/:replay_id] ERROR: invalid replay_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	log.Printf("[GET /replays/%s] user_id=%s", replayID, userID)

	replay, err := h.repo.GetReplayByID(c.Request.Context(), replayID, userID)
	if err != nil {
		log.Printf("[GET /replays/%s] ERROR: %v", replayID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	log.Printf("[GET /replays/%s] SUCCESS: %s", replayID, replay.OriginalName)
	c.JSON(http.StatusOK, replay)
}

func (h *Handler) CreateReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[POST /games/:game_id/replays] ERROR: invalid game_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	log.Printf("[POST /games/%s/replays] user_id=%s", gameID, userID)

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("[POST /games/%s/replays] ERROR: no file: %v", gameID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	title := c.PostForm("title")
	comment := c.PostForm("comment")

	log.Printf("[POST /games/%s/replays] file=%s, size=%d, title=%s, comment=%s",
		gameID, file.Filename, file.Size, title, comment)

	replayID := uuid.New()
	ext := filepath.Ext(file.Filename)
	fileName := replayID.String() + ext
	relPath := filepath.Join(userID.String(), gameID.String(), fileName)
	fullPath := filepath.Join(h.storageDir, relPath)

	log.Printf("[POST /games/%s/replays] Creating directory: %s", gameID, filepath.Dir(fullPath))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Printf("[POST /games/%s/replays] ERROR: failed to create directory: %v", gameID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
		return
	}

	log.Printf("[POST /games/%s/replays] Saving file to: %s", gameID, fullPath)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		log.Printf("[POST /games/%s/replays] ERROR: failed to save file: %v", gameID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	replay := &models.Replay{
		Title:        &title,
		OriginalName: file.Filename,
		FilePath:     relPath,
		SizeBytes:    file.Size,
		Compression:  "none",
		Compressed:   false,
		Comment:      &comment,
		GameID:       gameID,
		UserID:       userID,
	}

	log.Printf("[POST /games/%s/replays] Saving to database: replay_id=%s", gameID, replayID)

	if err := h.repo.CreateReplay(c.Request.Context(), replay); err != nil {
		log.Printf("[POST /games/%s/replays] ERROR: failed to save to DB: %v", gameID, err)
		os.Remove(fullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create replay"})
		return
	}

	log.Printf("[POST /games/%s/replays] SUCCESS: created replay_id=%s", gameID, replay.ID)
	c.JSON(http.StatusCreated, gin.H{"id": replay.ID})
}

func (h *Handler) DeleteReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[DELETE /replays/:replay_id] ERROR: invalid replay_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	log.Printf("[DELETE /replays/%s] user_id=%s", replayID, userID)

	filePath, err := h.repo.DeleteReplay(c.Request.Context(), replayID, userID)
	if err != nil {
		log.Printf("[DELETE /replays/%s] ERROR: %v", replayID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	fullPath := filepath.Join(h.storageDir, filePath)
	log.Printf("[DELETE /replays/%s] Removing file: %s", replayID, fullPath)

	if err := os.Remove(fullPath); err != nil {
		log.Printf("[DELETE /replays/%s] WARNING: failed to remove file: %v", replayID, err)
	}

	log.Printf("[DELETE /replays/%s] SUCCESS", replayID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) DeleteGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[DELETE /games/:game_id] ERROR: invalid game_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	log.Printf("[DELETE /games/%s] user_id=%s", gameID, userID)

	filePaths, err := h.repo.DeleteGame(c.Request.Context(), gameID, userID)
	if err != nil {
		log.Printf("[DELETE /games/%s] ERROR: %v", gameID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game"})
		return
	}

	log.Printf("[DELETE /games/%s] Removing %d files", gameID, len(filePaths))

	for _, path := range filePaths {
		fullPath := filepath.Join(h.storageDir, path)
		if err := os.Remove(fullPath); err != nil {
			log.Printf("[DELETE /games/%s] WARNING: failed to remove file %s: %v", gameID, path, err)
		}
	}

	log.Printf("[DELETE /games/%s] SUCCESS", gameID)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) UpdateReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[PUT /replays/:replay_id] ERROR: invalid replay_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	title := c.PostForm("title")
	comment := c.PostForm("comment")

	log.Printf("[PUT /replays/%s] user_id=%s, title=%s, comment=%s", replayID, userID, title, comment)

	var titlePtr, commentPtr *string
	if title != "" {
		titlePtr = &title
	}
	if comment != "" {
		commentPtr = &comment
	}

	if err := h.repo.UpdateReplay(c.Request.Context(), replayID, userID, titlePtr, commentPtr); err != nil {
		log.Printf("[PUT /replays/%s] ERROR: %v", replayID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update replay"})
		return
	}

	log.Printf("[PUT /replays/%s] SUCCESS", replayID)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) GetReplayFile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[GET /replays/:replay_id/file] ERROR: invalid replay_id: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	log.Printf("[GET /replays/%s/file] user_id=%s", replayID, userID)

	replay, err := h.repo.GetReplayByID(c.Request.Context(), replayID, userID)
	if err != nil {
		log.Printf("[GET /replays/%s/file] ERROR: replay not found: %v", replayID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	fullPath := filepath.Join(h.storageDir, replay.FilePath)
	log.Printf("[GET /replays/%s/file] Opening file: %s", replayID, fullPath)

	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("[GET /replays/%s/file] ERROR: file not found: %v", replayID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer file.Close()

	log.Printf("[GET /replays/%s/file] SUCCESS: serving %s", replayID, replay.OriginalName)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", replay.OriginalName))
	c.Header("Content-Type", "application/octet-stream")
	io.Copy(c.Writer, file)
}

func (h *Handler) CreateGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[POST /games] ERROR: invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	log.Printf("[POST /games] user_id=%s, name=%s", userID, req.Name)

	game, err := h.repo.CreateGame(c.Request.Context(), userID, req.Name)
	if err != nil {
		log.Printf("[POST /games] ERROR: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game"})
		return
	}

	log.Printf("[POST /games] SUCCESS: created game_id=%s", game.ID)
	c.JSON(http.StatusCreated, game)
}
