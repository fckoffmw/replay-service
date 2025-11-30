package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/fckoffmw/replay-service/server/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
	logger    *slog.Logger
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, logger *slog.Logger) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		logger:    logger,
	}
}

type Claims struct {
	UserID string `json:"user_id"`
	Login  string `json:"login"`
	jwt.RegisteredClaims
}

func (s *AuthService) Register(ctx context.Context, login, password string) (string, error) {
	existing, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		s.logger.Error("failed to check existing user", slog.String("error", err.Error()))
		return "", err
	}
	if existing != nil {
		return "", ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("failed to hash password", slog.String("error", err.Error()))
		return "", err
	}

	user, err := s.userRepo.Create(ctx, login, string(passwordHash))
	if err != nil {
		s.logger.Error("failed to create user", slog.String("error", err.Error()))
		return "", err
	}

	token, err := s.generateToken(user)
	if err != nil {
		s.logger.Error("failed to generate token", slog.String("error", err.Error()))
		return "", err
	}

	s.logger.Info("user registered", slog.String("user_id", user.ID.String()), slog.String("login", login))
	return token, nil
}

func (s *AuthService) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.userRepo.GetByLogin(ctx, login)
	if err != nil {
		s.logger.Error("failed to get user", slog.String("error", err.Error()))
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		s.logger.Error("failed to generate token", slog.String("error", err.Error()))
		return "", err
	}

	s.logger.Info("user logged in", slog.String("user_id", user.ID.String()), slog.String("login", login))
	return token, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := Claims{
		UserID: user.ID.String(),
		Login:  user.Login,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *AuthService) ValidateToken(tokenString string) (*uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return nil, err
		}
		return &userID, nil
	}

	return nil, errors.New("invalid token")
}
