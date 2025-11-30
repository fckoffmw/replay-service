package services

import (
	"context"
	"fmt"
	"log"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/fckoffmw/replay-service/server/internal/storage"
	"github.com/google/uuid"
)

// GameService handles business logic for games
type GameService struct {
	gameRepo   *repository.GameRepository
	replayRepo *repository.ReplayRepository
	storage    *storage.FileStorage
}

func NewGameService(
	gameRepo *repository.GameRepository,
	replayRepo *repository.ReplayRepository,
	storage *storage.FileStorage,
) *GameService {
	return &GameService{
		gameRepo:   gameRepo,
		replayRepo: replayRepo,
		storage:    storage,
	}
}

// GetUserGames returns all games for a user
func (s *GameService) GetUserGames(ctx context.Context, userID uuid.UUID) ([]models.Game, error) {
	log.Printf("[GameService] GetUserGames: user_id=%s", userID)
	
	games, err := s.gameRepo.GetByUserID(ctx, userID)
	if err != nil {
		log.Printf("[GameService] GetUserGames ERROR: %v", err)
		return nil, fmt.Errorf("failed to get games: %w", err)
	}
	
	log.Printf("[GameService] GetUserGames SUCCESS: found %d games", len(games))
	return games, nil
}

// CreateGame creates a new game
func (s *GameService) CreateGame(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error) {
	log.Printf("[GameService] CreateGame: user_id=%s, name=%s", userID, name)
	
	game, err := s.gameRepo.Create(ctx, userID, name)
	if err != nil {
		log.Printf("[GameService] CreateGame ERROR: %v", err)
		return nil, fmt.Errorf("failed to create game: %w", err)
	}
	
	log.Printf("[GameService] CreateGame SUCCESS: game_id=%s", game.ID)
	return game, nil
}

// UpdateGame updates game name
func (s *GameService) UpdateGame(ctx context.Context, gameID, userID uuid.UUID, name string) error {
	log.Printf("[GameService] UpdateGame: game_id=%s, user_id=%s, name=%s", gameID, userID, name)
	
	if err := s.gameRepo.Update(ctx, gameID, userID, name); err != nil {
		log.Printf("[GameService] UpdateGame ERROR: %v", err)
		return fmt.Errorf("failed to update game: %w", err)
	}
	
	log.Printf("[GameService] UpdateGame SUCCESS")
	return nil
}

// DeleteGame deletes a game and all its replays
func (s *GameService) DeleteGame(ctx context.Context, gameID, userID uuid.UUID) error {
	log.Printf("[GameService] DeleteGame: game_id=%s, user_id=%s", gameID, userID)
	
	// Get all replay file paths before deleting
	filePaths, err := s.replayRepo.GetFilePathsByGameID(ctx, gameID, userID)
	if err != nil {
		log.Printf("[GameService] DeleteGame ERROR getting file paths: %v", err)
		return fmt.Errorf("failed to get replay files: %w", err)
	}
	
	// Delete game (cascades to replays via DB constraint)
	if err := s.gameRepo.Delete(ctx, gameID, userID); err != nil {
		log.Printf("[GameService] DeleteGame ERROR: %v", err)
		return fmt.Errorf("failed to delete game: %w", err)
	}
	
	// Delete files from storage
	log.Printf("[GameService] DeleteGame: removing %d files", len(filePaths))
	if errs := s.storage.DeleteFiles(filePaths); len(errs) > 0 {
		for _, err := range errs {
			log.Printf("[GameService] DeleteGame WARNING: %v", err)
		}
	}
	
	log.Printf("[GameService] DeleteGame SUCCESS")
	return nil
}
