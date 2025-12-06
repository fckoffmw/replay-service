package services

import (
	"context"
	"log/slog"
	"mime/multipart"
	"path/filepath"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
)

const (
	compressionNone = "none"
)

type ReplayService struct {
	replayRepo ReplayRepositoryInterface
	storage    FileStorageInterface
	logger     *slog.Logger
}

func NewReplayService(
	replayRepo ReplayRepositoryInterface,
	storage FileStorageInterface,
	logger *slog.Logger,
) *ReplayService {
	return &ReplayService{
		replayRepo: replayRepo,
		storage:    storage,
		logger:     logger,
	}
}

func (s *ReplayService) GetGameReplays(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error) {
	s.logger.Info("getting game replays",
		slog.String("game_id", gameID.String()),
		slog.String("user_id", userID.String()),
		slog.Int("limit", limit))

	replays, err := s.replayRepo.GetByGameID(ctx, gameID, userID, limit)
	if err != nil {
		s.logger.Error("failed to get replays", slog.String("error", err.Error()))
		return nil, wrapError("get replays", err)
	}

	s.logger.Info("replays retrieved", slog.Int("count", len(replays)))
	return replays, nil
}

func (s *ReplayService) GetReplay(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
	s.logger.Info("getting replay",
		slog.String("replay_id", replayID.String()),
		slog.String("user_id", userID.String()))

	replay, err := s.replayRepo.GetByID(ctx, replayID, userID)
	if err != nil {
		s.logger.Error("replay not found", slog.String("error", err.Error()))
		return nil, notFoundError("replay", err)
	}

	s.logger.Info("replay retrieved", slog.String("filename", replay.OriginalName))
	return replay, nil
}

func (s *ReplayService) CreateReplay(
	ctx context.Context,
	file *multipart.FileHeader,
	gameID, userID uuid.UUID,
	title, comment string,
) (*models.Replay, error) {
	s.logger.Info("creating replay",
		slog.String("game_id", gameID.String()),
		slog.String("user_id", userID.String()),
		slog.String("filename", file.Filename),
		slog.String("title", title))

	replay := &models.Replay{
		ID:           uuid.New(),
		Title:        stringPtr(title),
		OriginalName: file.Filename,
		SizeBytes:    file.Size,
		Compression:  compressionNone,
		Compressed:   false,
		Comment:      stringPtr(comment),
		GameID:       gameID,
		UserID:       userID,
	}

	filePath, err := s.storage.SaveReplayFile(file, userID, gameID, replay.ID)
	if err != nil {
		s.logger.Error("failed to save file", slog.String("error", err.Error()))
		return nil, wrapError("save file", err)
	}
	replay.FilePath = filePath

	if err := s.replayRepo.Create(ctx, replay); err != nil {
		s.logger.Error("failed to save replay to database", slog.String("error", err.Error()))
		s.storage.DeleteFile(filePath)
		return nil, wrapError("create replay", err)
	}

	s.logger.Info("replay created", slog.String("replay_id", replay.ID.String()))
	return replay, nil
}

func (s *ReplayService) UpdateReplay(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error {
	s.logger.Info("updating replay",
		slog.String("replay_id", replayID.String()),
		slog.String("user_id", userID.String()))

	if err := s.replayRepo.Update(ctx, replayID, userID, title, comment); err != nil {
		s.logger.Error("failed to update replay", slog.String("error", err.Error()))
		return wrapError("update replay", err)
	}

	s.logger.Info("replay updated")
	return nil
}

func (s *ReplayService) DeleteReplay(ctx context.Context, replayID, userID uuid.UUID) error {
	s.logger.Info("deleting replay",
		slog.String("replay_id", replayID.String()),
		slog.String("user_id", userID.String()))

	filePath, err := s.replayRepo.Delete(ctx, replayID, userID)
	if err != nil {
		s.logger.Error("replay not found", slog.String("error", err.Error()))
		return notFoundError("replay", err)
	}

	if err := s.storage.DeleteFile(filePath); err != nil {
		s.logger.Warn("failed to delete file", slog.String("error", err.Error()))
	}

	s.logger.Info("replay deleted")
	return nil
}

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
