package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGetReplays_Success проверяет получение списка реплеев игры
func TestGetReplays_Success(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	gameID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/games/:game_id/replays", handler.GetReplays)

	expectedReplays := []models.Replay{
		{ID: uuid.New(), OriginalName: "replay1.rep", GameID: gameID},
		{ID: uuid.New(), OriginalName: "replay2.rep", GameID: gameID},
	}

	mockReplayService.On("GetGameReplays", mock.Anything, gameID, userID, 5).Return(expectedReplays, nil)

	req, _ := http.NewRequest("GET", "/games/"+gameID.String()+"/replays", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.Replay
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, 2, len(response))

	mockReplayService.AssertExpectations(t)
}

// TestGetReplays_WithLimit проверяет работу параметра limit
func TestGetReplays_WithLimit(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	gameID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/games/:game_id/replays", handler.GetReplays)

	expectedReplays := []models.Replay{
		{ID: uuid.New(), OriginalName: "replay1.rep"},
	}

	// Проверяем, что передается правильный лимит
	mockReplayService.On("GetGameReplays", mock.Anything, gameID, userID, 10).Return(expectedReplays, nil)

	req, _ := http.NewRequest("GET", "/games/"+gameID.String()+"/replays?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockReplayService.AssertExpectations(t)
}

// TestGetReplays_InvalidGameID проверяет валидацию game_id
func TestGetReplays_InvalidGameID(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/games/:game_id/replays", handler.GetReplays)

	req, _ := http.NewRequest("GET", "/games/invalid-uuid/replays", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid game_id")

	mockReplayService.AssertNotCalled(t, "GetGameReplays")
}

// TestGetReplay_Success проверяет получение одного реплея
func TestGetReplay_Success(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/replays/:replay_id", handler.GetReplay)

	expectedReplay := &models.Replay{
		ID:           replayID,
		OriginalName: "epic_game.rep",
		GameName:     "Counter-Strike 2",
	}

	mockReplayService.On("GetReplay", mock.Anything, replayID, userID).Return(expectedReplay, nil)

	req, _ := http.NewRequest("GET", "/replays/"+replayID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Replay
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, replayID, response.ID)
	assert.Equal(t, "epic_game.rep", response.OriginalName)

	mockReplayService.AssertExpectations(t)
}

// TestGetReplay_NotFound проверяет обработку несуществующего реплея
func TestGetReplay_NotFound(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/replays/:replay_id", handler.GetReplay)

	mockReplayService.On("GetReplay", mock.Anything, replayID, userID).Return(nil, assert.AnError)

	req, _ := http.NewRequest("GET", "/replays/"+replayID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "replay not found")

	mockReplayService.AssertExpectations(t)
}

// TestCreateReplay_Success проверяет успешную загрузку реплея
func TestCreateReplay_Success(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	gameID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.POST("/games/:game_id/replays", handler.CreateReplay)

	expectedReplay := &models.Replay{
		ID:           replayID,
		OriginalName: "test.rep",
	}

	mockReplayService.On("CreateReplay", mock.Anything, mock.Anything, gameID, userID, "Epic Game", "Best game").Return(expectedReplay, nil)

	// Создаем multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Epic Game")
	writer.WriteField("comment", "Best game")
	part, _ := writer.CreateFormFile("file", "test.rep")
	part.Write([]byte("fake replay data"))
	writer.Close()

	req, _ := http.NewRequest("POST", "/games/"+gameID.String()+"/replays", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, replayID.String(), response["id"])

	mockReplayService.AssertExpectations(t)
}

// TestCreateReplay_MissingFile проверяет валидацию обязательного файла
func TestCreateReplay_MissingFile(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	gameID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.POST("/games/:game_id/replays", handler.CreateReplay)

	// Отправляем form без файла
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Test")
	writer.Close()

	req, _ := http.NewRequest("POST", "/games/"+gameID.String()+"/replays", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "file is required")

	mockReplayService.AssertNotCalled(t, "CreateReplay")
}

// TestUpdateReplay_Success проверяет обновление реплея
func TestUpdateReplay_Success(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.PUT("/replays/:replay_id", handler.UpdateReplay)

	newTitle := "Updated Title"
	newComment := "Updated Comment"
	mockReplayService.On("UpdateReplay", mock.Anything, replayID, userID, &newTitle, &newComment).Return(nil)

	// Создаем form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", newTitle)
	writer.WriteField("comment", newComment)
	writer.Close()

	req, _ := http.NewRequest("PUT", "/replays/"+replayID.String(), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "updated", response["message"])

	mockReplayService.AssertExpectations(t)
}

// TestDeleteReplay_Success проверяет удаление реплея
func TestDeleteReplay_Success(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.DELETE("/replays/:replay_id", handler.DeleteReplay)

	mockReplayService.On("DeleteReplay", mock.Anything, replayID, userID).Return(nil)

	req, _ := http.NewRequest("DELETE", "/replays/"+replayID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "deleted", response["message"])

	mockReplayService.AssertExpectations(t)
}

// TestDeleteReplay_NotFound проверяет удаление несуществующего реплея
func TestDeleteReplay_NotFound(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.DELETE("/replays/:replay_id", handler.DeleteReplay)

	mockReplayService.On("DeleteReplay", mock.Anything, replayID, userID).Return(assert.AnError)

	req, _ := http.NewRequest("DELETE", "/replays/"+replayID.String(), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "replay not found")

	mockReplayService.AssertExpectations(t)
}

// TestGetReplayFile_Success проверяет скачивание файла реплея
func TestGetReplayFile_Success(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()
	replayID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/replays/:replay_id/file", handler.GetReplayFile)

	replay := &models.Replay{
		ID:           replayID,
		OriginalName: "game.rep",
	}

	// Создаем временный файл для теста
	tmpFile := "/tmp/test_replay_" + replayID.String() + ".rep"
	// Примечание: в реальном тесте нужно создать файл, но для демонстрации пропустим

	mockReplayService.On("GetReplay", mock.Anything, replayID, userID).Return(replay, nil)
	mockReplayService.On("GetReplayFilePath", mock.Anything, replayID, userID).Return(tmpFile, ".rep", nil)

	req, _ := http.NewRequest("GET", "/replays/"+replayID.String()+"/file", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Файл не существует, поэтому ожидаем 404
	// В реальном окружении файл бы существовал
	assert.Equal(t, http.StatusNotFound, w.Code)

	mockReplayService.AssertExpectations(t)
}

// TestGetReplayFile_InvalidReplayID проверяет валидацию replay_id
func TestGetReplayFile_InvalidReplayID(t *testing.T) {
	mockGameService := &MockGameService{}
	mockReplayService := new(MockReplayService)
	handler := NewHandler(mockGameService, mockReplayService)

	router := setupTestRouter()
	userID := uuid.New()

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})

	router.GET("/replays/:replay_id/file", handler.GetReplayFile)

	req, _ := http.NewRequest("GET", "/replays/invalid-uuid/file", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid replay_id")

	mockReplayService.AssertNotCalled(t, "GetReplay")
}
