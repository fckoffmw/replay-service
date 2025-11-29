package models

import (
	"time"

	"github.com/google/uuid"
)

type Game struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	UserID      uuid.UUID `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	ReplayCount int       `json:"replay_count,omitempty"`
}
