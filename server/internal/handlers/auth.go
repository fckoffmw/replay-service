package handlers

import (
	"errors"
	"net/http"

	"github.com/fckoffmw/replay-service/server/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) SetLogger(logger interface{}) {
	// For future use if needed
}

type RegisterRequest struct {
	Login    string `json:"login" binding:"required,min=3"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Register(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Пользователь уже существует"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка регистрации"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{Token: token})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.authService.Login(c.Request.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный логин или пароль"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка входа"})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{Token: token})
}
