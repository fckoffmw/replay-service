package repository

import (
	"context"
	"errors"

	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, login, passwordHash string) (*models.User, error) {
	user := &models.User{}
	query := `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, login, password_hash, created_at
	`
	err := r.db.Pool.QueryRow(ctx, query, login, passwordHash).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password_hash, created_at FROM users WHERE login = $1`
	err := r.db.Pool.QueryRow(ctx, query, login).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
