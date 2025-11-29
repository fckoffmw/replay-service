package models

import (
	"time"

	"github.com/google/uuid"
)

type Replay struct {
	ID           uuid.UUID `json:"id"`
	Title        *string   `json:"title,omitempty"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"-"`
	SizeBytes    int64     `json:"size_bytes"`
	UploadedAt   time.Time `json:"uploaded_at"`
	Compression  string    `json:"compression"`
	Compressed   bool      `json:"compressed"`
	Comment      *string   `json:"comment,omitempty"`
	GameID       uuid.UUID `json:"game_id"`
	GameName     string    `json:"game_name,omitempty"`
	UserID       uuid.UUID `json:"-"`
}
