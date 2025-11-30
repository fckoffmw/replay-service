package services

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/fckoffmw/replay-service/server/internal/storage"
	"github.com/google/uuid"
)

// ReplayService handles business logic for replays
type ReplayService struct {
	replayRepo *repository.ReplayRepository
	storage    *storage.FileStorage
}

func NewReplayService(
	replayRepo *repository.ReplayRepository,
	storage *storage.FileStorage,
) *ReplayService {
	return &ReplayService{
		replayRepo: replayRepo,
		storage:    storage,
	}
}

// GetGameReplays returns replays for a specific game
func (s *ReplayService) GetGameReplays(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error) {
	log.Printf("[ReplayService] GetGameReplays: game_id=%s, user_id=%s, limit=%d", gameID, userID, limit)
	
	replays, err := s.replayRepo.GetByGameID(ctx, gameID, userID, limit)
	if err != nil {
		log.Printf("[ReplayService] GetGameReplays ERROR: %v", err)
		return nil, fmt.Errorf("failed to get replays: %w", err)
	}
	
	log.Printf("[ReplayService] GetGameReplays SUCCESS: found %d replays", len(replays))
	return replays, nil
}

// GetReplay returns a single replay by ID
func (s *ReplayService) GetReplay(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
	log.Printf("[ReplayService] GetReplay: replay_id=%s, user_id=%s", replayID, userID)
	
	replay, err := s.replayRepo.GetByID(ctx, replayID, userID)
	if err != nil {
		log.Printf("[ReplayService] GetReplay ERROR: %v", err)
		return nil, fmt.Errorf("replay not found: %w", err)
	}
	
	log.Printf("[ReplayService] GetReplay SUCCESS: %s", replay.OriginalName)
	return replay, nil
}

// CreateReplay creates a new replay with file upload
func (s *ReplayService) CreateReplay(
	ctx context.Context,
	file *multipart.FileHeader,
	gameID, userID uuid.UUID,
	title, comment string,
) (*models.Replay, error) {
	log.Printf("[ReplayService] CreateReplay: game_id=%s, user_id=%s, file=%s, title=%s",
		gameID, userID, file.Filename, title)
	
	// Create replay model
	replay := &models.Replay{
		ID:           uuid.New(),
		Title:        stringPtr(title),
		OriginalName: file.Filename,
		SizeBytes:    file.Size,
		Compression:  "none",
		Compressed:   false,
		Comment:      stringPtr(comment),
		GameID:       gameID,
		UserID:       userID,
	}
	
	// Save file to storage
	filePath, err := s.storage.SaveReplayFile(file, userID, gameID, replay.ID)
	if err != nil {
		log.Printf("[ReplayService] CreateReplay ERROR saving file: %v", err)
		return nil, fmt.Errorf("failed to save file: %w", err)
	}
	replay.FilePath = filePath
	
	// Save metadata to database
	if err := s.replayRepo.Create(ctx, replay); err != nil {
		log.Printf("[ReplayService] CreateReplay ERROR saving to DB: %v", err)
		// Rollback: delete file
		s.storage.DeleteFile(filePath)
		return nil, fmt.Errorf("failed to create replay: %w", err)
	}
	
	log.Printf("[ReplayService] CreateReplay SUCCESS: replay_id=%s", replay.ID)
	return replay, nil
}

// UpdateReplay updates replay metadata
func (s *ReplayService) UpdateReplay(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error {
	log.Printf("[ReplayService] UpdateReplay: replay_id=%s, user_id=%s", replayID, userID)
	
	if err := s.replayRepo.Update(ctx, replayID, userID, title, comment); err != nil {
		log.Printf("[ReplayService] UpdateReplay ERROR: %v", err)
		return fmt.Errorf("failed to update replay: %w", err)
	}
	
	log.Printf("[ReplayService] UpdateReplay SUCCESS")
	return nil
}

// DeleteReplay deletes a replay and its file
func (s *ReplayService) DeleteReplay(ctx context.Context, replayID, userID uuid.UUID) error {
	log.Printf("[ReplayService] DeleteReplay: replay_id=%s, user_id=%s", replayID, userID)
	
	// Delete from database and get file path
	filePath, err := s.replayRepo.Delete(ctx, replayID, userID)
	if err != nil {
		log.Printf("[ReplayService] DeleteReplay ERROR: %v", err)
		return fmt.Errorf("replay not found: %w", err)
	}
	
	// Delete file from storage
	if err := s.storage.DeleteFile(filePath); err != nil {
		log.Printf("[ReplayService] DeleteReplay WARNING: failed to delete file: %v", err)
	}
	
	log.Printf("[ReplayService] DeleteReplay SUCCESS")
	return nil
}

// GetReplayFilePath returns the full path to replay file
func (s *ReplayService) GetReplayFilePath(ctx context.Context, replayID, userID uuid.UUID) (string, string, error) {
	replay, err := s.GetReplay(ctx, replayID, userID)
	if err != nil {
		return "", "", err
	}
	
	fullPath := s.storage.GetFilePath(replay.FilePath)
	ext := filepath.Ext(replay.OriginalName)
	
	return fullPath, ext, nil
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
