package handlers

import (
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	paramReplayID      = "replay_id"
	formFieldFile      = "file"
	formFieldTitle     = "title"
	formFieldComment   = "comment"
	queryLimit         = "limit"
	queryDownload      = "download"
	defaultReplayLimit = 5
)

func (h *Handler) GetReplays(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	gameID, err := uuid.Parse(c.Param(paramGameID))
	if err != nil {
		respondBadRequest(c, "invalid game_id")
		return
	}

	limit := defaultReplayLimit
	if limitStr := c.Query(queryLimit); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	replays, err := h.replayService.GetGameReplays(c.Request.Context(), gameID, userID, limit)
	if err != nil {
		respondInternalError(c, "failed to get replays")
		return
	}

	respondOK(c, replays)
}

func (h *Handler) GetReplay(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	replayID, err := uuid.Parse(c.Param(paramReplayID))
	if err != nil {
		respondBadRequest(c, "invalid replay_id")
		return
	}

	replay, err := h.replayService.GetReplay(c.Request.Context(), replayID, userID)
	if err != nil {
		respondNotFound(c, "replay not found")
		return
	}

	respondOK(c, replay)
}

func (h *Handler) CreateReplay(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	gameID, err := uuid.Parse(c.Param(paramGameID))
	if err != nil {
		respondBadRequest(c, "invalid game_id")
		return
	}

	file, err := c.FormFile(formFieldFile)
	if err != nil {
		respondBadRequest(c, "file is required")
		return
	}

	title := c.PostForm(formFieldTitle)
	comment := c.PostForm(formFieldComment)

	replay, err := h.replayService.CreateReplay(c.Request.Context(), file, gameID, userID, title, comment)
	if err != nil {
		respondInternalError(c, "failed to create replay")
		return
	}

	respondCreated(c, gin.H{"id": replay.ID})
}

func (h *Handler) DeleteReplay(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	replayID, err := uuid.Parse(c.Param(paramReplayID))
	if err != nil {
		respondBadRequest(c, "invalid replay_id")
		return
	}

	if err := h.replayService.DeleteReplay(c.Request.Context(), replayID, userID); err != nil {
		respondNotFound(c, "replay not found")
		return
	}

	respondSuccess(c, "deleted")
}

func (h *Handler) UpdateReplay(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	replayID, err := uuid.Parse(c.Param(paramReplayID))
	if err != nil {
		respondBadRequest(c, "invalid replay_id")
		return
	}

	title := c.PostForm(formFieldTitle)
	comment := c.PostForm(formFieldComment)

	var titlePtr, commentPtr *string
	if title != "" {
		titlePtr = &title
	}
	if comment != "" {
		commentPtr = &comment
	}

	if err := h.replayService.UpdateReplay(c.Request.Context(), replayID, userID, titlePtr, commentPtr); err != nil {
		respondInternalError(c, "failed to update replay")
		return
	}

	respondSuccess(c, "updated")
}

func (h *Handler) GetReplayFile(c *gin.Context) {
	userID := c.MustGet(contextKeyUserID).(uuid.UUID)
	replayID, err := uuid.Parse(c.Param(paramReplayID))
	if err != nil {
		respondBadRequest(c, "invalid replay_id")
		return
	}

	replay, err := h.replayService.GetReplay(c.Request.Context(), replayID, userID)
	if err != nil {
		respondNotFound(c, "replay not found")
		return
	}

	fullPath, ext, err := h.replayService.GetReplayFilePath(c.Request.Context(), replayID, userID)
	if err != nil {
		respondNotFound(c, "file not found")
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		respondNotFound(c, "file not found")
		return
	}
	defer file.Close()

	contentType := getContentType(ext)

	download := c.Query(queryDownload) == "true"

	if download || !isVideoFile(ext) {
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", replay.OriginalName))
	} else {
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s", replay.OriginalName))
	}

	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=31536000")

	fileInfo, err := file.Stat()
	if err != nil {
		respondInternalError(c, "failed to read file")
		return
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	io.Copy(c.Writer, file)
}

func getContentType(ext string) string {
	contentTypes := map[string]string{
		".mp4":  "video/mp4",
		".webm": "video/webm",
		".ogg":  "video/ogg",
		".ogv":  "video/ogg",
		".mov":  "video/quicktime",
		".avi":  "video/x-msvideo",
		".mkv":  "video/x-matroska",
		".m4v":  "video/x-m4v",
	}

	if contentType, ok := contentTypes[ext]; ok {
		return contentType
	}
	return "application/octet-stream"
}

func isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".webm", ".ogg", ".ogv", ".mov", ".avi", ".mkv", ".m4v"}
	return slices.Contains(videoExts, ext)
}
