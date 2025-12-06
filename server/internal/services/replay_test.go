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

// Используем те же моки, что и в game_test.go

// TestGetGameReplays_Success проверяет получение списка реплеев
func TestGetGameReplays_Success(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	limit := 5
	
	expectedReplays := []models.Replay{
		{ID: uuid.New(), OriginalName: "replay1.rep", GameID: gameID},
		{ID: uuid.New(), OriginalName: "replay2.rep", GameID: gameID},
	}
	
	mockReplayRepo.On("GetByGameID", mock.Anything, gameID, userID, limit).Return(expectedReplays, nil)
	
	replays, err := service.GetGameReplays(context.Background(), gameID, userID, limit)
	
	assert.NoError(t, err)
	assert.Equal(t, 2, len(replays))
	assert.Equal(t, expectedReplays, replays)
	
	mockReplayRepo.AssertExpectations(t)
}

// TestGetReplay_Success проверяет получение одного реплея
func TestGetReplay_Success(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	replayID := uuid.New()
	userID := uuid.New()
	
	expectedReplay := &models.Replay{
		ID:           replayID,
		OriginalName: "epic_game.rep",
		UserID:       userID,
	}
	
	mockReplayRepo.On("GetByID", mock.Anything, replayID, userID).Return(expectedReplay, nil)
	
	replay, err := service.GetReplay(context.Background(), replayID, userID)
	
	assert.NoError(t, err)
	assert.NotNil(t, replay)
	assert.Equal(t, replayID, replay.ID)
	assert.Equal(t, "epic_game.rep", replay.OriginalName)
	
	mockReplayRepo.AssertExpectations(t)
}

// TestGetReplay_NotFound проверяет обработку случая, когда реплей не найден
func TestGetReplay_NotFound(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	replayID := uuid.New()
	userID := uuid.New()
	
	mockReplayRepo.On("GetByID", mock.Anything, replayID, userID).Return(nil, errors.New("not found"))
	
	replay, err := service.GetReplay(context.Background(), replayID, userID)
	
	assert.Error(t, err)
	assert.Nil(t, replay)
	
	mockReplayRepo.AssertExpectations(t)
}

// TestCreateReplay_Success проверяет успешное создание реплея
// Что тестируем: файл сохраняется, затем запись создается в БД
func TestCreateReplay_Success(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	
	file := &multipart.FileHeader{
		Filename: "test_replay.rep",
		Size:     1024,
	}
	
	title := "Epic Game"
	comment := "Best game ever"
	filePath := "user/game/replay.rep"
	
	// Настраиваем моки: сначала сохраняется файл, потом создается запись в БД
	mockStorage.On("SaveReplayFile", file, userID, gameID, mock.AnythingOfType("uuid.UUID")).Return(filePath, nil)
	mockReplayRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Replay")).Return(nil)
	
	replay, err := service.CreateReplay(context.Background(), file, gameID, userID, title, comment)
	
	assert.NoError(t, err)
	assert.NotNil(t, replay)
	assert.Equal(t, "test_replay.rep", replay.OriginalName)
	assert.Equal(t, int64(1024), replay.SizeBytes)
	assert.Equal(t, title, *replay.Title)
	assert.Equal(t, comment, *replay.Comment)
	assert.Equal(t, filePath, replay.FilePath)
	assert.Equal(t, "none", replay.Compression)
	assert.False(t, replay.Compressed)
	
	mockStorage.AssertExpectations(t)
	mockReplayRepo.AssertExpectations(t)
}

// TestCreateReplay_StorageError проверяет обработку ошибки сохранения файла
// Что тестируем: если файл не сохранился, запись в БД не создается
func TestCreateReplay_StorageError(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	
	file := &multipart.FileHeader{
		Filename: "test_replay.rep",
		Size:     1024,
	}
	
	// Настраиваем мок: SaveReplayFile возвращает ошибку
	mockStorage.On("SaveReplayFile", file, userID, gameID, mock.AnythingOfType("uuid.UUID")).Return("", errors.New("disk full"))
	
	replay, err := service.CreateReplay(context.Background(), file, gameID, userID, "", "")
	
	assert.Error(t, err)
	assert.Nil(t, replay)
	assert.Contains(t, err.Error(), "save file")
	
	// Проверяем, что Create НЕ был вызван (запись в БД не создается при ошибке файла)
	mockReplayRepo.AssertNotCalled(t, "Create")
	
	mockStorage.AssertExpectations(t)
}

// TestCreateReplay_DatabaseError проверяет откат при ошибке БД
// Что тестируем: если запись в БД не создалась, файл удаляется (rollback)
func TestCreateReplay_DatabaseError(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	gameID := uuid.New()
	userID := uuid.New()
	
	file := &multipart.FileHeader{
		Filename: "test_replay.rep",
		Size:     1024,
	}
	
	filePath := "user/game/replay.rep"
	
	// Настраиваем моки: файл сохраняется, но БД возвращает ошибку
	mockStorage.On("SaveReplayFile", file, userID, gameID, mock.AnythingOfType("uuid.UUID")).Return(filePath, nil)
	mockReplayRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.Replay")).Return(errors.New("db constraint violation"))
	// Важно: при ошибке БД файл должен быть удален
	mockStorage.On("DeleteFile", filePath).Return(nil)
	
	replay, err := service.CreateReplay(context.Background(), file, gameID, userID, "", "")
	
	assert.Error(t, err)
	assert.Nil(t, replay)
	assert.Contains(t, err.Error(), "create replay")
	
	// Проверяем, что DeleteFile был вызван (rollback)
	mockStorage.AssertCalled(t, "DeleteFile", filePath)
	
	mockStorage.AssertExpectations(t)
	mockReplayRepo.AssertExpectations(t)
}

// TestUpdateReplay_Success проверяет обновление метаданных реплея
func TestUpdateReplay_Success(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	replayID := uuid.New()
	userID := uuid.New()
	newTitle := "Updated Title"
	newComment := "Updated Comment"
	
	mockReplayRepo.On("Update", mock.Anything, replayID, userID, &newTitle, &newComment).Return(nil)
	
	err := service.UpdateReplay(context.Background(), replayID, userID, &newTitle, &newComment)
	
	assert.NoError(t, err)
	mockReplayRepo.AssertExpectations(t)
}

// TestDeleteReplay_Success проверяет удаление реплея
// Что тестируем: сначала удаляется запись из БД, затем файл
func TestDeleteReplay_Success(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	replayID := uuid.New()
	userID := uuid.New()
	filePath := "user/game/replay.rep"
	
	mockReplayRepo.On("Delete", mock.Anything, replayID, userID).Return(filePath, nil)
	mockStorage.On("DeleteFile", filePath).Return(nil)
	
	err := service.DeleteReplay(context.Background(), replayID, userID)
	
	assert.NoError(t, err)
	mockReplayRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}

// TestDeleteReplay_NotFound проверяет удаление несуществующего реплея
func TestDeleteReplay_NotFound(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	replayID := uuid.New()
	userID := uuid.New()
	
	mockReplayRepo.On("Delete", mock.Anything, replayID, userID).Return("", errors.New("not found"))
	
	err := service.DeleteReplay(context.Background(), replayID, userID)
	
	assert.Error(t, err)
	// Проверяем, что DeleteFile НЕ был вызван (файл не удаляется, если реплей не найден)
	mockStorage.AssertNotCalled(t, "DeleteFile")
	
	mockReplayRepo.AssertExpectations(t)
}

// TestGetReplayFilePath_Success проверяет получение пути к файлу
func TestGetReplayFilePath_Success(t *testing.T) {
	mockReplayRepo := new(MockReplayRepository)
	mockStorage := new(MockFileStorage)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	
	service := NewReplayService(mockReplayRepo, mockStorage, logger)
	
	replayID := uuid.New()
	userID := uuid.New()
	
	replay := &models.Replay{
		ID:           replayID,
		OriginalName: "game.rep",
		FilePath:     "user/game/replay.rep",
	}
	
	mockReplayRepo.On("GetByID", mock.Anything, replayID, userID).Return(replay, nil)
	mockStorage.On("GetFilePath", "user/game/replay.rep").Return("/full/path/user/game/replay.rep")
	
	fullPath, ext, err := service.GetReplayFilePath(context.Background(), replayID, userID)
	
	assert.NoError(t, err)
	assert.Equal(t, "/full/path/user/game/replay.rep", fullPath)
	assert.Equal(t, ".rep", ext)
	
	mockReplayRepo.AssertExpectations(t)
	mockStorage.AssertExpectations(t)
}
