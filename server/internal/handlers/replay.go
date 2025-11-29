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

const (
	handlerPath = "server/internal/handlers/replay.go"
)

type Handler struct {
	repo       *repository.ReplayRepository
	storageDir string
}

func NewHandler(repo *repository.ReplayRepository, storageDir string) *Handler {
	log.Printf("[%s/NewHandler] Initialized with storage dir: %s", handlerPath, storageDir)
	return &Handler{
		repo:       repo,
		storageDir: storageDir,
	}
}

func (h *Handler) GetGames(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	log.Printf("[%s/GetGames] user_id=%s", handlerPath, userID)

	games, err := h.repo.GetGamesByUserID(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[%s/GetGames] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get games"})
		return
	}

	log.Printf("[%s/GetGames] SUCCESS: found %d games", handlerPath, len(games))
	c.JSON(http.StatusOK, games)
}

func (h *Handler) GetReplays(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[%s/GetReplays] ERROR: invalid game_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	limit := 5
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	log.Printf("[%s/GetReplays] user_id=%s, game_id=%s, limit=%d", handlerPath, userID, gameID, limit)

	replays, err := h.repo.GetReplaysByGameID(c.Request.Context(), gameID, userID, limit)
	if err != nil {
		log.Printf("[%s/GetReplays] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get replays"})
		return
	}

	log.Printf("[%s/GetReplays] SUCCESS: found %d replays", handlerPath, len(replays))
	c.JSON(http.StatusOK, replays)
}

func (h *Handler) GetReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[%s/GetReplay] ERROR: invalid replay_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	log.Printf("[%s/GetReplay] user_id=%s, replay_id=%s", handlerPath, userID, replayID)

	replay, err := h.repo.GetReplayByID(c.Request.Context(), replayID, userID)
	if err != nil {
		log.Printf("[%s/GetReplay] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	log.Printf("[%s/GetReplay] SUCCESS: %s", handlerPath, replay.OriginalName)
	c.JSON(http.StatusOK, replay)
}

func (h *Handler) CreateReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[%s/CreateReplay] ERROR: invalid game_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	log.Printf("[%s/CreateReplay] user_id=%s, game_id=%s", handlerPath, userID, gameID)

	file, err := c.FormFile("file")
	if err != nil {
		log.Printf("[%s/CreateReplay] ERROR: no file: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}

	title := c.PostForm("title")
	comment := c.PostForm("comment")

	log.Printf("[%s/CreateReplay] file=%s, size=%d, title=%s, comment=%s",
		handlerPath, file.Filename, file.Size, title, comment)

	replay := &models.Replay{
		ID:           uuid.New(),
		Title:        &title,
		OriginalName: file.Filename,
		SizeBytes:    file.Size,
		Compression:  "none",
		Compressed:   false,
		Comment:      &comment,
		GameID:       gameID,
		UserID:       userID,
	}

	ext := filepath.Ext(file.Filename)
	fileName := replay.ID.String() + ext
	relPath := filepath.Join(userID.String(), gameID.String(), fileName)
	fullPath := filepath.Join(h.storageDir, relPath)
	replay.FilePath = relPath

	log.Printf("[%s/CreateReplay] Creating directory: %s", handlerPath, filepath.Dir(fullPath))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Printf("[%s/CreateReplay] ERROR: failed to create directory: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
		return
	}

	log.Printf("[%s/CreateReplay] Saving file to: %s", handlerPath, fullPath)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		log.Printf("[%s/CreateReplay] ERROR: failed to save file: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
		return
	}

	log.Printf("[%s/CreateReplay] Saving to database: replay_id=%s", handlerPath, replay.ID)

	if err := h.repo.CreateReplay(c.Request.Context(), replay); err != nil {
		log.Printf("[%s/CreateReplay] ERROR: failed to save to DB: %v", handlerPath, err)
		os.Remove(fullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create replay"})
		return
	}

	log.Printf("[%s/CreateReplay] SUCCESS: created replay_id=%s", handlerPath, replay.ID)
	c.JSON(http.StatusCreated, gin.H{"id": replay.ID})
}

func (h *Handler) DeleteReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[%s/DeleteReplay] ERROR: invalid replay_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	log.Printf("[%s/DeleteReplay] user_id=%s, replay_id=%s", handlerPath, userID, replayID)

	filePath, err := h.repo.DeleteReplay(c.Request.Context(), replayID, userID)
	if err != nil {
		log.Printf("[%s/DeleteReplay] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	fullPath := filepath.Join(h.storageDir, filePath)
	log.Printf("[%s/DeleteReplay] Removing file: %s", handlerPath, fullPath)

	if err := os.Remove(fullPath); err != nil {
		log.Printf("[%s/DeleteReplay] WARNING: failed to remove file: %v", handlerPath, err)
	}

	log.Printf("[%s/DeleteReplay] SUCCESS", handlerPath)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) DeleteGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[%s/DeleteGame] ERROR: invalid game_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	log.Printf("[%s/DeleteGame] user_id=%s, game_id=%s", handlerPath, userID, gameID)

	filePaths, err := h.repo.DeleteGame(c.Request.Context(), gameID, userID)
	if err != nil {
		log.Printf("[%s/DeleteGame] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete game"})
		return
	}

	log.Printf("[%s/DeleteGame] Removing %d files", handlerPath, len(filePaths))

	for _, path := range filePaths {
		fullPath := filepath.Join(h.storageDir, path)
		if err := os.Remove(fullPath); err != nil {
			log.Printf("[%s/DeleteGame] WARNING: failed to remove file %s: %v", handlerPath, path, err)
		}
	}

	log.Printf("[%s/DeleteGame] SUCCESS", handlerPath)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *Handler) UpdateReplay(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[%s/UpdateReplay] ERROR: invalid replay_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	title := c.PostForm("title")
	comment := c.PostForm("comment")

	log.Printf("[%s/UpdateReplay] user_id=%s, replay_id=%s, title=%s, comment=%s", handlerPath, userID, replayID, title, comment)

	var titlePtr, commentPtr *string
	if title != "" {
		titlePtr = &title
	}
	if comment != "" {
		commentPtr = &comment
	}

	if err := h.repo.UpdateReplay(c.Request.Context(), replayID, userID, titlePtr, commentPtr); err != nil {
		log.Printf("[%s/UpdateReplay] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update replay"})
		return
	}

	log.Printf("[%s/UpdateReplay] SUCCESS", handlerPath)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) GetReplayFile(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	replayID, err := uuid.Parse(c.Param("replay_id"))
	if err != nil {
		log.Printf("[%s/GetReplayFile] ERROR: invalid replay_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid replay_id"})
		return
	}

	log.Printf("[%s/GetReplayFile] user_id=%s, replay_id=%s", handlerPath, userID, replayID)

	replay, err := h.repo.GetReplayByID(c.Request.Context(), replayID, userID)
	if err != nil {
		log.Printf("[%s/GetReplayFile] ERROR: replay not found: %v", handlerPath, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "replay not found"})
		return
	}

	fullPath := filepath.Join(h.storageDir, replay.FilePath)
	log.Printf("[%s/GetReplayFile] Opening file: %s", handlerPath, fullPath)

	file, err := os.Open(fullPath)
	if err != nil {
		log.Printf("[%s/GetReplayFile] ERROR: file not found: %v", handlerPath, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	defer file.Close()

	log.Printf("[%s/GetReplayFile] SUCCESS: serving %s", handlerPath, replay.OriginalName)
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
		log.Printf("[%s/CreateGame] ERROR: invalid request: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	log.Printf("[%s/CreateGame] user_id=%s, name=%s", handlerPath, userID, req.Name)

	game, err := h.repo.CreateGame(c.Request.Context(), userID, req.Name)
	if err != nil {
		log.Printf("[%s/CreateGame] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create game"})
		return
	}

	log.Printf("[%s/CreateGame] SUCCESS: created game_id=%s", handlerPath, game.ID)
	c.JSON(http.StatusCreated, game)
}

func (h *Handler) UpdateGame(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	gameID, err := uuid.Parse(c.Param("game_id"))
	if err != nil {
		log.Printf("[%s/UpdateGame] ERROR: invalid game_id: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid game_id"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[%s/UpdateGame] ERROR: invalid request: %v", handlerPath, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	log.Printf("[%s/UpdateGame] user_id=%s, game_id=%s, name=%s", handlerPath, userID, gameID, req.Name)

	if err := h.repo.UpdateGame(c.Request.Context(), gameID, userID, req.Name); err != nil {
		log.Printf("[%s/UpdateGame] ERROR: %v", handlerPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update game"})
		return
	}

	log.Printf("[%s/UpdateGame] SUCCESS", handlerPath)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}
