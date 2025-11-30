package repository

import (
	"context"

	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
)

type GameRepository struct {
	db *database.DB
}

func NewGameRepository(db *database.DB) *GameRepository {
	return &GameRepository{db: db}
}

func (r *GameRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Game, error) {
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
		return nil, wrapQueryError("query games", err)
	}
	defer rows.Close()

	games := make([]models.Game, 0)
	for rows.Next() {
		var game models.Game
		if err := rows.Scan(&game.ID, &game.Name, &game.CreatedAt, &game.ReplayCount); err != nil {
			return nil, wrapScanError("game", err)
		}
		games = append(games, game)
	}

	return games, rows.Err()
}

func (r *GameRepository) Create(ctx context.Context, userID uuid.UUID, name string) (*models.Game, error) {
	query := `
		INSERT INTO games (name, user_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, created_at
	`

	var game models.Game
	err := r.db.Pool.QueryRow(ctx, query, name, userID).Scan(&game.ID, &game.Name, &game.CreatedAt)
	if err != nil {
		return nil, wrapQueryError("create game", err)
	}

	game.UserID = userID
	return &game, nil
}

func (r *GameRepository) Update(ctx context.Context, gameID, userID uuid.UUID, name string) error {
	query := `
		UPDATE games
		SET name = $1
		WHERE id = $2 AND user_id = $3
	`

	result, err := r.db.Pool.Exec(ctx, query, name, gameID, userID)
	if err != nil {
		return wrapQueryError("update game", err)
	}

	if result.RowsAffected() == 0 {
		return wrapNotFoundError("game")
	}

	return nil
}

func (r *GameRepository) Delete(ctx context.Context, gameID, userID uuid.UUID) error {
	query := `DELETE FROM games WHERE id = $1 AND user_id = $2`

	result, err := r.db.Pool.Exec(ctx, query, gameID, userID)
	if err != nil {
		return wrapQueryError("delete game", err)
	}

	if result.RowsAffected() == 0 {
		return wrapNotFoundError("game")
	}

	return nil
}
