package repository

import (
	"context"
	"fmt"
	"log"

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

func (r *ReplayRepository) GetGamesByUserID(ctx context.Context, userID uuid.UUID) ([]models.Game, error) {
	query := `
		SELECT g.id, g.name, g.created_at, COUNT(r.id) as replay_count
		FROM games g
		LEFT JOIN replays r ON r.game_id = g.id
		WHERE g.user_id = $1
		GROUP BY g.id, g.name, g.created_at
		ORDER BY g.created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query games: %w", err)
	}
	defer rows.Close()

	games := make([]models.Game, 0)
	for rows.Next() {
		var game models.Game
		if err := rows.Scan(&game.ID, &game.Name, &game.CreatedAt, &game.ReplayCount); err != nil {
			return nil, fmt.Errorf("failed to scan game: %w", err)
		}
		games = append(games, game)
	}

	return games, rows.Err()
}

func (r *ReplayRepository) GetReplaysByGameID(ctx context.Context, gameID, userID uuid.UUID, limit int) ([]models.Replay, error) {
	query := `
		SELECT r.id, r.title, r.original_name, r.uploaded_at, r.size_bytes, r.compression, r.compressed, r.comment, r.game_id
		FROM replays r
		WHERE r.game_id = $1 AND r.user_id = $2
		ORDER BY r.uploaded_at DESC
		LIMIT $3
	`

	log.Printf("[REPO] GetReplaysByGameID: game_id=%s, user_id=%s, limit=%d", gameID, userID, limit)

	rows, err := r.db.Pool.Query(ctx, query, gameID, userID, limit)
	if err != nil {
		log.Printf("[REPO] GetReplaysByGameID ERROR: %v", err)
		return nil, fmt.Errorf("failed to query replays: %w", err)
	}
	defer rows.Close()

	replays := make([]models.Replay, 0)
	for rows.Next() {
		var replay models.Replay
		if err := rows.Scan(&replay.ID, &replay.Title, &replay.OriginalName, &replay.UploadedAt, &replay.SizeBytes, &replay.Compression, &replay.Compressed, &replay.Comment, &replay.GameID); err != nil {
			log.Printf("[REPO] GetReplaysByGameID SCAN ERROR: %v", err)
			return nil, fmt.Errorf("failed to scan replay: %w", err)
		}
		replays = append(replays, replay)
	}

	log.Printf("[REPO] GetReplaysByGameID SUCCESS: found %d replays", len(replays))
	return replays, rows.Err()
}

func (r *ReplayRepository) GetReplayByID(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
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
		return nil, fmt.Errorf("failed to get replay: %w", err)
	}

	replay.UserID = userID
	return &replay, nil
}

func (r *ReplayRepository) CreateGame(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error) {
	query := `
		INSERT INTO games (name, user_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, created_at
	`

	var game models.Game
	err := r.db.Pool.QueryRow(ctx, query, name, userID).Scan(&game.ID, &game.Name, &game.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create game: %w", err)
	}

	game.UserID = userID
	return &game, nil
}

func (r *ReplayRepository) CreateReplay(ctx context.Context, replay *models.Replay) error {
	query := `
		INSERT INTO replays (id, title, original_name, file_path, size_bytes, compression, compressed, comment, game_id, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING uploaded_at
	`

	log.Printf("[REPO] CreateReplay: id=%s, title=%v, file=%s, path=%s, size=%d, game_id=%s, user_id=%s",
		replay.ID, replay.Title, replay.OriginalName, replay.FilePath, replay.SizeBytes, replay.GameID, replay.UserID)

	err := r.db.Pool.QueryRow(ctx, query,
		replay.ID, replay.Title, replay.OriginalName, replay.FilePath, replay.SizeBytes,
		replay.Compression, replay.Compressed, replay.Comment, replay.GameID, replay.UserID,
	).Scan(&replay.UploadedAt)

	if err != nil {
		log.Printf("[REPO] CreateReplay ERROR: %v", err)
		return fmt.Errorf("failed to create replay: %w", err)
	}

	log.Printf("[REPO] CreateReplay SUCCESS: id=%s", replay.ID)
	return nil
}

func (r *ReplayRepository) DeleteReplay(ctx context.Context, replayID, userID uuid.UUID) (string, error) {
	query := `
		DELETE FROM replays
		WHERE id = $1 AND user_id = $2
		RETURNING file_path
	`

	var filePath string
	err := r.db.Pool.QueryRow(ctx, query, replayID, userID).Scan(&filePath)
	if err != nil {
		return "", fmt.Errorf("failed to delete replay: %w", err)
	}

	return filePath, nil
}

func (r *ReplayRepository) DeleteGame(ctx context.Context, gameID, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT file_path FROM replays WHERE game_id = $1 AND user_id = $2
	`

	rows, err := r.db.Pool.Query(ctx, query, gameID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query replay paths: %w", err)
	}
	defer rows.Close()

	var filePaths []string
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("failed to scan file path: %w", err)
		}
		filePaths = append(filePaths, path)
	}

	deleteQuery := `DELETE FROM games WHERE id = $1 AND user_id = $2`
	_, err = r.db.Pool.Exec(ctx, deleteQuery, gameID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete game: %w", err)
	}

	return filePaths, nil
}

func (r *ReplayRepository) UpdateReplay(ctx context.Context, replayID, userID uuid.UUID, title, comment *string) error {
	query := `
		UPDATE replays
		SET title = COALESCE($1, title), comment = COALESCE($2, comment)
		WHERE id = $3 AND user_id = $4
	`

	_, err := r.db.Pool.Exec(ctx, query, title, comment, replayID, userID)
	if err != nil {
		return fmt.Errorf("failed to update replay: %w", err)
	}

	return nil
}
