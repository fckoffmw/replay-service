package services

import (
	"context"
	"errors"
	"log/slog"
	"mime/multipart"
	"os"
	"testing"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGameRepository - мок для GameRepository
// Зачем: позволяет тестировать GameService изолированно, без реальной БД
type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Game, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Game), args.Error(1)
}

func (m *MockGameRepository) Create(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error) {
	args := m.Called(ctx, userID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Game), args.Error(1)
}

func (m *MockGameRepository) Update(ctx context.Context, gameID, userID uuid.UUID, name string) error {
	args := m.Called(ctx, gameID, userID, name)
	return args.Error(0)
}

func (m *MockGameRepository) Delete(ctx context.Context, gameID, userID uuid.UUID) error {
	args := m.Called(ctx, gameID, userID)
	return args.Error(0)
}

// MockReplayRepository - мок для ReplayRepository
type MockReplayRepository struct {
	mock.Mock
}

func (m *MockReplayRepository) GetByGameID(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error) {
	args := m.Called(ctx, gameID, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Replay), args.Error(1)
}

func (m *MockReplayRepository) GetByID(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
	args := m.Called(ctx, replayID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Replay), args.Error(1)
}

func (m *MockReplayRepository) Create(ctx context.Context, replay *models.Replay) error {
	args := m.Called(ctx, replay)
	return args.Error(0)
}

func (m *MockReplayRepository) Update(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error {
	args := m.Called(ctx, replayID, userID, title, comment)
	return args.Error(0)
}

func (m *MockReplayRepository) Delete(ctx context.Context, replayID, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, replayID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockReplayRepository) GetFilePathsByGameID(ctx context.Context, gameID, userID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, gameID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockFileStorage - мок для FileStorage
type MockFileStorage struct {
	mock.Mock
}

func (m *MockFileStorage) SaveReplayFile(file *multipart.FileHeader, userID, gameID, replayID uuid.UUID) (string, error) {
	args := m.Called(file, userID, gameID, replayID)
	return args.String(0), args.Error(1)
}

func (m *MockFileStorage) DeleteFile(filePath string) error {
	args := m.Called(filePath)
	return args.Error(0)
}

func (m *MockFileStorage) DeleteFiles(filePaths []string) []error {
	args := m.Called(filePaths)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]error)
}

func (m *MockFileStorage) GetFilePath(relativePath string) string {
	args := m.Called(relativePath)
	return args.String(0)
}

// TestGetUserGames_Success проверяет успешное получение списка игр
// Что тестируем: сервис корректно вызывает репозиторий и возвращает данные
func TestGetUserGames_Success(t *testing.T) {
	// Arrange - подготовка тестовых данных
	mockGameRepo := new(MockGameRepository)
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewGameService(mockGameRepo, mockReplayRepo, mockStorage, logger)
	
	userID := uuid.New()
	expectedGames := []models.Game{
		{ID: uuid.New(), Name: "Game 1", UserID: userID},
		{ID: uuid.New(), Name: "Game 2", UserID: userID},
	}
	
	// Настраиваем мок: при вызове GetByUserID вернуть expectedGames
	mockGameRepo.On("GetByUserID", mock.Anything, userID).Return(expectedGames, nil)
	
	// Act - выполнение тестируемого действия
	games, err := service.GetUserGames(context.Background(), userID)
	
	// Assert - проверка результатов
	assert.NoError(t, err, "не должно быть ошибки")
	assert.Equal(t, 2, len(games), "должно вернуться 2 игры")
	assert.Equal(t, expectedGames, games, "игры должны совпадать")
	
	// Проверяем, что мок был вызван с правильными параметрами
	mockGameRepo.AssertExpectations(t)
}

// TestGetUserGames_RepositoryError проверяет обработку ошибки БД
// Что тестируем: сервис корректно обрабатывает ошибки репозитория
func TestGetUserGames_RepositoryError(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewGameService(mockGameRepo, mockReplayRepo, mockStorage, logger)
	
	userID := uuid.New()
	expectedError := errors.New("database connection failed")
	
	// Настраиваем мок: при вызове GetByUserID вернуть ошибку
	mockGameRepo.On("GetByUserID", mock.Anything, userID).Return(nil, expectedError)
	
	games, err := service.GetUserGames(context.Background(), userID)
	
	assert.Error(t, err, "должна быть ошибка")
	assert.Nil(t, games, "игры должны быть nil")
	assert.Contains(t, err.Error(), "get games", "ошибка должна содержать контекст")
	
	mockGameRepo.AssertExpectations(t)
}

// TestCreateGame_Success проверяет успешное создание игры
func TestCreateGame_Success(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewGameService(mockGameRepo, mockReplayRepo, mockStorage, logger)
	
	userID := uuid.New()
	gameName := "Counter-Strike 2"
	expectedGame := &models.Game{
		ID:     uuid.New(),
		Name:   gameName,
		UserID: userID,
	}
	
	mockGameRepo.On("Create", mock.Anything, userID, gameName).Return(expectedGame, nil)
	
	game, err := service.CreateGame(context.Background(), userID, gameName)
	
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.Equal(t, gameName, game.Name)
	assert.Equal(t, userID, game.UserID)
	
	mockGameRepo.AssertExpectations(t)
}

// TestUpdateGame_Success проверяет успешное обновление игры
func TestUpdateGame_Success(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewGameService(mockGameRepo, mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	newName := "Updated Game Name"
	
	mockGameRepo.On("Update", mock.Anything, gameID, userID, newName).Return(nil)
	
	err := service.UpdateGame(context.Background(), gameID, userID, newName)
	
	assert.NoError(t, err)
	mockGameRepo.AssertExpectations(t)
}

// TestDeleteGame_Success проверяет успешное удаление игры
// Что тестируем: сервис удаляет игру из БД и все связанные файлы
func TestDeleteGame_Success(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewGameService(mockGameRepo, mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	filePaths := []string{"path/to/replay1.rep", "path/to/replay2.rep"}
	
	// Настраиваем моки: сначала получаем пути файлов, потом удаляем игру, потом файлы
	mockReplayRepo.On("GetFilePathsByGameID", mock.Anything, gameID, userID).Return(filePaths, nil)
	mockGameRepo.On("Delete", mock.Anything, gameID, userID).Return(nil)
	mockStorage.On("DeleteFiles", filePaths).Return([]error(nil))
	
	err := service.DeleteGame(context.Background(), gameID, userID)
	
	assert.NoError(t, err)
	mockReplayRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestDeleteGame_RepositoryError проверяет обработку ошибки при удалении
// Что тестируем: если не удалось удалить из БД, файлы не удаляются
func TestDeleteGame_RepositoryError(t *testing.T) {
	mockGameRepo := new(MockGameRepository)
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewGameService(mockGameRepo, mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	filePaths := []string{"path/to/replay1.rep"}
	
	mockReplayRepo.On("GetFilePathsByGameID", mock.Anything, gameID, userID).Return(filePaths, nil)
	mockGameRepo.On("Delete", mock.Anything, gameID, userID).Return(errors.New("db error"))
	
	err := service.DeleteGame(context.Background(), gameID, userID)
	
	assert.Error(t, err)
	// Проверяем, что DeleteFiles НЕ был вызван (файлы не удаляются при ошибке БД)
	mockStorage.AssertNotCalled(t, "DeleteFiles")
	
	mockReplayRepo.AssertExpectations(t)
	mockGameRepo.AssertExpectations(t)
}
