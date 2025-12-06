package services

import (
	"context"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
	"mime/multipart"
)

// GameRepositoryInterface определяет методы для работы с играми в БД
// Зачем: позволяет использовать моки в тестах вместо реального репозитория
type GameRepositoryInterface interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Game, error)
	Create(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error)
	Update(ctx context.Context, gameID, userID uuid.UUID, name string) error
	Delete(ctx context.Context, gameID, userID uuid.UUID) error
}

// ReplayRepositoryInterface определяет методы для работы с реплеями в БД
type ReplayRepositoryInterface interface {
	GetByGameID(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error)
	GetByID(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error)
	Create(ctx context.Context, replay *models.Replay) error
	Update(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error
	Delete(ctx context.Context, replayID, userID uuid.UUID) (string, error)
	GetFilePathsByGameID(ctx context.Context, gameID, userID uuid.UUID) ([]string, error)
}

// FileStorageInterface определяет методы для работы с файловой системой
type FileStorageInterface interface {
	SaveReplayFile(file *multipart.FileHeader, userID, gameID, replayID uuid.UUID) (string, error)
	DeleteFile(filePath string) error
	DeleteFiles(filePaths []string) []error
	GetFilePath(relativePath string) string
}
