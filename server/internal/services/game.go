package services

import (
	"context"
	"log/slog"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/fckoffmw/replay-service/server/internal/storage"
	"github.com/google/uuid"
)

type GameService struct {
	gameRepo   *repository.GameRepository
	replayRepo *repository.ReplayRepository
	storage    *storage.FileStorage
	logger     *slog.Logger
}

func NewGameService(
	gameRepo *repository.GameRepository,
	replayRepo *repository.ReplayRepository,
	storage *storage.FileStorage,
	logger *slog.Logger,
) *GameService {
	return &GameService{
		gameRepo:   gameRepo,
		replayRepo: replayRepo,
		storage:    storage,
		logger:     logger,
	}
}

func (s *GameService) GetUserGames(ctx context.Context, userID uuid.UUID) ([]models.Game, error) {
	s.logger.Info("getting user games", slog.String("user_id", userID.String()))

	games, err := s.gameRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get games", slog.String("error", err.Error()))
		return nil, wrapError("get games", err)
	}

	s.logger.Info("games retrieved", slog.Int("count", len(games)))
	return games, nil
}

func (s *GameService) CreateGame(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error) {
	s.logger.Info("creating game",
		slog.String("user_id", userID.String()),
		slog.String("name", name))

	game, err := s.gameRepo.Create(ctx, userID, name)
	if err != nil {
		s.logger.Error("failed to create game", slog.String("error", err.Error()))
		return nil, wrapError("create game", err)
	}

	s.logger.Info("game created", slog.String("game_id", game.ID.String()))
	return game, nil
}

func (s *GameService) UpdateGame(ctx context.Context, gameID, userID uuid.UUID, name string) error {
	s.logger.Info("updating game",
		slog.String("game_id", gameID.String()),
		slog.String("user_id", userID.String()),
		slog.String("name", name))

	if err := s.gameRepo.Update(ctx, gameID, userID, name); err != nil {
		s.logger.Error("failed to update game", slog.String("error", err.Error()))
		return wrapError("update game", err)
	}

	s.logger.Info("game updated")
	return nil
}

func (s *GameService) DeleteGame(ctx context.Context, gameID, userID uuid.UUID) error {
	s.logger.Info("deleting game",
		slog.String("game_id", gameID.String()),
		slog.String("user_id", userID.String()))

	filePaths, err := s.replayRepo.GetFilePathsByGameID(ctx, gameID, userID)
	if err != nil {
		s.logger.Error("failed to get replay files", slog.String("error", err.Error()))
		return wrapError("get replay files", err)
	}

	if err := s.gameRepo.Delete(ctx, gameID, userID); err != nil {
		s.logger.Error("failed to delete game", slog.String("error", err.Error()))
		return wrapError("delete game", err)
	}

	s.logger.Info("removing replay files", slog.Int("count", len(filePaths)))
	if errs := s.storage.DeleteFiles(filePaths); len(errs) > 0 {
		for _, err := range errs {
			s.logger.Warn("failed to delete file", slog.String("error", err.Error()))
		}
	}

	s.logger.Info("game deleted")
	return nil
}
