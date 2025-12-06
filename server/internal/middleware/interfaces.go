package middleware

import "github.com/google/uuid"

// AuthServiceInterface определяет методы для аутентификации
type AuthServiceInterface interface {
	ValidateToken(token string) (*uuid.UUID, error)
}
