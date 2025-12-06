package handlers

import (
	"context"
	"mime/multipart"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
)

// GameServiceInterface определяет методы для работы с играми
type GameServiceInterface interface {
	GetUserGames(ctx context.Context, userID uuid.UUID) ([]models.Game, error)
	CreateGame(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error)
	UpdateGame(ctx context.Context, gameID, userID uuid.UUID, name string) error
	DeleteGame(ctx context.Context, gameID, userID uuid.UUID) error
}

// ReplayServiceInterface определяет методы для работы с реплеями
type ReplayServiceInterface interface {
	GetGameReplays(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error)
	GetReplay(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error)
	CreateReplay(ctx context.Context, file *multipart.FileHeader, gameID, userID uuid.UUID, title, comment string) (*models.Replay, error)
	UpdateReplay(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error
	DeleteReplay(ctx context.Context, replayID, userID uuid.UUID) error
	GetReplayFilePath(ctx context.Context, replayID, userID uuid.UUID) (string, string, error)
}
