package handlers

import (
	"bytes"
	"context"
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

// MockGameService - мок для GameService
// Зачем: тестируем только HTTP слой, без реальной бизнес-логики
type MockGameService struct {
	mock.Mock
}

func (m *MockGameService) GetUserGames(ctx context.Context, userID uuid.UUID) ([]models.Game, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockGameService) CreateGame(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameService) UpdateGame(ctx context.Context, gameID, userID uuid.UUID, name string) error {
	args := m.Called(ctx, gameID, userID, name)
	return args.Error(0)
}

func (m *MockGameService) DeleteGame(ctx context.Context, gameID, userID uuid.UUID) error {
	args := m.Called(ctx, gameID, userID)
	return args.Error(0)
}

// setupTestRouter создает тестовый Gin router
// Зачем: изолированное тестирование HTTP handlers без запуска всего сервера
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestGetGames_Success проверяет успешное получение списка игр
// Что тестируем: HTTP 200, корректный JSON ответ
func TestGetGames_Success(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{} // пустой мок
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	
	// Настраиваем middleware для установки user_id
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.GET("/games", handler.GetGames)
	
	expectedGames := []models.Game{
		{ID: uuid.New(), Name: "Game 1"},
		{ID: uuid.New(), Name: "Game 2"},
	}
	
	mockGameService.On("GetUserGames", mock.Anything, userID).Return(expectedGames, nil)
	
	// Создаем HTTP запрос
	req, _ := http.NewRequest("GET", "/games", nil)
	w := httptest.NewRecorder()
	
	// Выполняем запрос
	router.ServeHTTP(w, req)
	
	// Проверяем ответ
	assert.Equal(t, http.StatusOK, w.Code, "должен вернуться статус 200")
	
	var response []models.Game
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "ответ должен быть валидным JSON")
	assert.Equal(t, 2, len(response), "должно быть 2 игры")
	
	mockGameService.AssertExpectations(t)
}

// TestGetGames_ServiceError проверяет обработку ошибки сервиса
// Что тестируем: HTTP 500 при ошибке бизнес-логики
func TestGetGames_ServiceError(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{}
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.GET("/games", handler.GetGames)
	
	mockGameService.On("GetUserGames", mock.Anything, userID).Return(nil, assert.AnError)
	
	req, _ := http.NewRequest("GET", "/games", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "failed to get games")
	
	mockGameService.AssertExpectations(t)
}

// TestCreateGame_Success проверяет успешное создание игры
func TestCreateGame_Success(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{}
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.POST("/games", handler.CreateGame)
	
	gameName := "Counter-Strike 2"
	expectedGame := &models.Game{
		ID:     uuid.New(),
		Name:   gameName,
		UserID: userID,
	}
	
	mockGameService.On("CreateGame", mock.Anything, userID, gameName).Return(expectedGame, nil)
	
	// Создаем JSON body
	body := map[string]string{"name": gameName}
	jsonBody, _ := json.Marshal(body)
	
	req, _ := http.NewRequest("POST", "/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code, "должен вернуться статус 201")
	
	var response models.Game
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, gameName, response.Name)
	
	mockGameService.AssertExpectations(t)
}

// TestCreateGame_MissingName проверяет валидацию обязательного поля
// Что тестируем: HTTP 400 при отсутствии обязательного поля
func TestCreateGame_MissingName(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{}
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.POST("/games", handler.CreateGame)
	
	// Отправляем пустой body
	body := map[string]string{}
	jsonBody, _ := json.Marshal(body)
	
	req, _ := http.NewRequest("POST", "/games", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "должен вернуться статус 400")
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "name is required")
	
	// Проверяем, что сервис НЕ был вызван
	mockGameService.AssertNotCalled(t, "CreateGame")
}

// TestUpdateGame_Success проверяет успешное обновление игры
func TestUpdateGame_Success(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{}
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	gameID := uuid.New()
	
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.PUT("/games/:game_id", handler.UpdateGame)
	
	newName := "Updated Game Name"
	mockGameService.On("UpdateGame", mock.Anything, gameID, userID, newName).Return(nil)
	
	body := map[string]string{"name": newName}
	jsonBody, _ := json.Marshal(body)
	
	req, _ := http.NewRequest("PUT", "/games/"+gameID.String(), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "updated", response["message"])
	
	mockGameService.AssertExpectations(t)
}

// TestUpdateGame_InvalidGameID проверяет валидацию UUID
// Что тестируем: HTTP 400 при невалидном UUID в URL
func TestUpdateGame_InvalidGameID(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{}
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.PUT("/games/:game_id", handler.UpdateGame)
	
	body := map[string]string{"name": "New Name"}
	jsonBody, _ := json.Marshal(body)
	
	// Отправляем невалидный UUID
	req, _ := http.NewRequest("PUT", "/games/invalid-uuid", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "invalid game_id")
	
	mockGameService.AssertNotCalled(t, "UpdateGame")
}

// TestDeleteGame_Success проверяет успешное удаление игры
func TestDeleteGame_Success(t *testing.T) {
	mockGameService := new(MockGameService)
	mockReplayService := &MockReplayService{}
	handler := NewHandler(mockGameService, mockReplayService)
	
	router := setupTestRouter()
	userID := uuid.New()
	gameID := uuid.New()
	
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	
	router.DELETE("/games/:game_id", handler.DeleteGame)
	
	mockGameService.On("DeleteGame", mock.Anything, gameID, userID).Return(nil)
	
	req, _ := http.NewRequest("DELETE", "/games/"+gameID.String(), nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "deleted", response["message"])
	
	mockGameService.AssertExpectations(t)
}

// MockReplayService - пустой мок для тестов GameHandler
type MockReplayService struct {
	mock.Mock
}

func (m *MockReplayService) GetGameReplays(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error) {
	args := m.Called(ctx, gameID, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Replay), args.Error(1)
}

func (m *MockReplayService) GetReplay(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
	args := m.Called(ctx, replayID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Replay), args.Error(1)
}

func (m *MockReplayService) CreateReplay(ctx context.Context, file *multipart.FileHeader, gameID, userID uuid.UUID, title, comment string) (*models.Replay, error) {
	args := m.Called(ctx, file, gameID, userID, title, comment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Replay), args.Error(1)
}

func (m *MockReplayService) UpdateReplay(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error {
	args := m.Called(ctx, replayID, userID, title, comment)
	return args.Error(0)
}

func (m *MockReplayService) DeleteReplay(ctx context.Context, replayID, userID uuid.UUID) error {
	args := m.Called(ctx, replayID, userID)
	return args.Error(0)
}

func (m *MockReplayService) GetReplayFilePath(ctx context.Context, replayID, userID uuid.UUID) (string, string, error) {
	args := m.Called(ctx, replayID, userID)
	return args.String(0), args.String(1), args.Error(2)
}
