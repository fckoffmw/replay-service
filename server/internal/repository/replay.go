package repository

import (
	"context"

	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
)

type ReplayRepository struct {
	db *database.DB
}

func NewReplayRepository(db *database.DB) *ReplayRepository {
	return &ReplayRepository{db: db}
}

func (r *ReplayRepository) GetByGameID(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error) {
	query := `
		SELECT r.id, r.title, r.original_name, r.uploaded_at, r.size_bytes, r.compression, r.compressed, r.comment, r.game_id
		FROM replays r
		WHERE r.game_id = $1 AND r.user_id = $2
		ORDER BY r.uploaded_at DESC
		LIMIT $3
	`

	rows, err := r.db.Pool.Query(ctx, query, gameID, userID, limit)
	if err != nil {
		return nil, wrapQueryError("query replays", err)
	}
	defer rows.Close()

	replays := make([]models.Replay, 0)
	for rows.Next() {
		var replay models.Replay
		if err := rows.Scan(&replay.ID, &replay.Title, &replay.OriginalName, &replay.UploadedAt, &replay.SizeBytes, &replay.Compression, &replay.Compressed, &replay.Comment, &replay.GameID); err != nil {
			return nil, wrapScanError("replay", err)
		}
		replays = append(replays, replay)
	}

	return replays, rows.Err()
}

func (r *ReplayRepository) GetByID(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
	query := `
		SELECT r.id, r.title, r.original_name, r.comment, r.uploaded_at, r.size_bytes, 
		       r.compression, r.compressed, r.file_path, r.game_id, g.name as game_name
		FROM replays r
		JOIN games g ON r.game_id = g.id
		WHERE r.id = $1 AND r.user_id = $2
	`

	var replay models.Replay
	err := r.db.Pool.QueryRow(ctx, query, replayID, userID).Scan(
		&replay.ID, &replay.Title, &replay.OriginalName, &replay.Comment, &replay.UploadedAt,
		&replay.SizeBytes, &replay.Compression, &replay.Compressed, &replay.FilePath,
		&replay.GameID, &replay.GameName,
	)
	if err != nil {
		return nil, wrapQueryError("get replay", err)
	}

	replay.UserID = userID
	return &replay, nil
}

func (r *ReplayRepository) Create(ctx context.Context, replay *models.Replay) error {
	query := `
		INSERT INTO replays (id, title, original_name, file_path, size_bytes, compression, compressed, comment, game_id, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING uploaded_at
	`

	err := r.db.Pool.QueryRow(ctx, query,
		replay.ID, replay.Title, replay.OriginalName, replay.FilePath, replay.SizeBytes,
		replay.Compression, replay.Compressed, replay.Comment, replay.GameID, replay.UserID,
	).Scan(&replay.UploadedAt)

	if err != nil {
		return wrapQueryError("create replay", err)
	}

	return nil
}

func (r *ReplayRepository) Delete(ctx context.Context, replayID, userID uuid.UUID) (string, error) {
	query := `
		DELETE FROM replays
		WHERE id = $1 AND user_id = $2
		RETURNING file_path
	`

	var filePath string
	err := r.db.Pool.QueryRow(ctx, query, replayID, userID).Scan(&filePath)
	if err != nil {
		return "", wrapQueryError("delete replay", err)
	}

	return filePath, nil
}

func (r *ReplayRepository) GetFilePathsByGameID(ctx context.Context, gameID, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT file_path FROM replays WHERE game_id = $1 AND user_id = $2
	`

	rows, err := r.db.Pool.Query(ctx, query, gameID, userID)
	if err != nil {
		return nil, wrapQueryError("query replay paths", err)
	}
	defer rows.Close()

	var filePaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, wrapScanError("file path", err)
		}
		filePaths = append(filePaths, path)
	}

	return filePaths, rows.Err()
}

func (r *ReplayRepository) Update(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error {
	query := `
		UPDATE replays
		SET title = COALESCE($1, title), comment = COALESCE($2, comment)
		WHERE id = $3 AND user_id = $4
	`

	_, err := r.db.Pool.Exec(ctx, query, title, comment, replayID, userID)
	if err != nil {
		return wrapQueryError("update replay", err)
	}

	return nil
}
