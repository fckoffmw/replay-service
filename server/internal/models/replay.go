package models

import (
	"time"

	"github.com/google/uuid"
)

// Replay представляет реплей из базы данных
type Replay struct {
	ID           uuid.UUID `json:"id"`
	OriginalName string    `json:"original_name"`
	FilePath     string    `json:"-"`
	SizeBytes    int64     `json:"-"`
	UploadedAt   time.Time `json:"-"`
	Compression  string    `json:"compression"`
	Compressed   bool      `json:"compressed"`
	UserID       uuid.UUID `json:"-"`
}
