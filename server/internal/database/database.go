package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB обертка над пулом соединений PostgreSQL
type DB struct {
	Pool *pgxpool.Pool
}

// Connect создает соединение с PostgreSQL используя DSN строку
func Connect(ctx context.Context, dsn string) (*DB, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Проверяем соединение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close закрывает пул соединений
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
