package repository

import (
	"context"
	"fmt"

	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/fckoffmw/replay-service/server/internal/models"
)

// ReplayRepository предоставляет методы для работы с реплеями в БД
type ReplayRepository struct {
	db *database.DB
}

// NewReplayRepository создает новый экземпляр ReplayRepository
func NewReplayRepository(db *database.DB) *ReplayRepository {
	return &ReplayRepository{db: db}
}

// GetAll возвращает список всех реплеев
func (r *ReplayRepository) GetAll(ctx context.Context) ([]models.Replay, error) {
	query := `
		SELECT id, original_name, file_path, size_bytes, uploaded_at, compression, compressed, user_id
		FROM replays
		ORDER BY uploaded_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query replays: %w", err)
	}
	defer rows.Close()

	var replays []models.Replay
	for rows.Next() {
		var replay models.Replay
		err := rows.Scan(
			&replay.ID,
			&replay.OriginalName,
			&replay.FilePath,
			&replay.SizeBytes,
			&replay.UploadedAt,
			&replay.Compression,
			&replay.Compressed,
			&replay.UserID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan replay: %w", err)
		}
		replays = append(replays, replay)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating replays: %w", err)
	}

	return replays, nil
}
