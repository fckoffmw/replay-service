package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService - мок для AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(token string) (*uuid.UUID, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*uuid.UUID), args.Error(1)
}

func (m *MockAuthService) Register(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

// setupTestRouter создает тестовый router
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// TestAuthMiddleware_ValidToken проверяет успешную аутентификацию
func TestAuthMiddleware_ValidToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()
	userID := uuid.New()

	// Настраиваем мок: токен валиден
	mockAuthService.On("ValidateToken", "valid-token").Return(&userID, nil)

	// Применяем middleware
	router.Use(AuthMiddleware(mockAuthService, logger))

	// Добавляем тестовый эндпоинт
	router.GET("/test", func(c *gin.Context) {
		// Проверяем, что user_id установлен в контексте
		contextUserID, exists := c.Get("user_id")
		assert.True(t, exists, "user_id должен быть в контексте")
		assert.Equal(t, userID, contextUserID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Отправляем запрос с валидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestAuthMiddleware_MissingToken проверяет отсутствие токена
func TestAuthMiddleware_MissingToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockAuthService, logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Отправляем запрос БЕЗ токена
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Требуется авторизация")

	// ValidateToken не должен быть вызван
	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

// TestAuthMiddleware_InvalidToken проверяет невалидный токен
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()

	// Настраиваем мок: токен невалиден
	mockAuthService.On("ValidateToken", "invalid-token").Return(nil, errors.New("invalid token"))

	router.Use(AuthMiddleware(mockAuthService, logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Отправляем запрос с невалидным токеном
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Неверный или истекший токен")

	mockAuthService.AssertExpectations(t)
}

// TestAuthMiddleware_ExpiredToken проверяет истекший токен
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()

	// Настраиваем мок: токен истек
	mockAuthService.On("ValidateToken", "expired-token").Return(nil, errors.New("token expired"))

	router.Use(AuthMiddleware(mockAuthService, logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestAuthMiddleware_TokenFromQuery проверяет токен из query параметра
func TestAuthMiddleware_TokenFromQuery(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()
	userID := uuid.New()

	// Настраиваем мок
	mockAuthService.On("ValidateToken", "query-token").Return(&userID, nil)

	router.Use(AuthMiddleware(mockAuthService, logger))

	router.GET("/test", func(c *gin.Context) {
		contextUserID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, userID, contextUserID)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Отправляем запрос с токеном в query параметре
	req, _ := http.NewRequest("GET", "/test?token=query-token", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestAuthMiddleware_InvalidAuthorizationFormat проверяет неправильный формат заголовка
func TestAuthMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()
	router.Use(AuthMiddleware(mockAuthService, logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	testCases := []struct {
		name   string
		header string
	}{
		{"No Bearer prefix", "just-token"},
		{"Wrong prefix", "Basic token"},
		{"Empty Bearer", "Bearer "},
		{"Only Bearer", "Bearer"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}

	mockAuthService.AssertNotCalled(t, "ValidateToken")
}

// TestAuthMiddleware_QueryTokenPriority проверяет приоритет query токена
func TestAuthMiddleware_QueryTokenPriority(t *testing.T) {
	mockAuthService := new(MockAuthService)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	router := setupTestRouter()
	userID := uuid.New()

	// Query токен имеет приоритет над header токеном
	mockAuthService.On("ValidateToken", "query-token").Return(&userID, nil)

	router.Use(AuthMiddleware(mockAuthService, logger))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Отправляем запрос с токеном и в query, и в header
	req, _ := http.NewRequest("GET", "/test?token=query-token", nil)
	req.Header.Set("Authorization", "Bearer header-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Проверяем, что использовался query токен, а не header
	mockAuthService.AssertCalled(t, "ValidateToken", "query-token")
	mockAuthService.AssertNotCalled(t, "ValidateToken", "header-token")
}

// TestExtractToken проверяет функцию извлечения токена
func TestExtractToken(t *testing.T) {
	testCases := []struct {
		name          string
		setupContext  func(*gin.Context)
		expectedToken string
	}{
		{
			name: "Token from query",
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/test?token=query-token", nil)
			},
			expectedToken: "query-token",
		},
		{
			name: "Token from header",
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/test", nil)
				c.Request.Header.Set("Authorization", "Bearer header-token")
			},
			expectedToken: "header-token",
		},
		{
			name: "No token",
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/test", nil)
			},
			expectedToken: "",
		},
		{
			name: "Query priority",
			setupContext: func(c *gin.Context) {
				c.Request, _ = http.NewRequest("GET", "/test?token=query-token", nil)
				c.Request.Header.Set("Authorization", "Bearer header-token")
			},
			expectedToken: "query-token",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tc.setupContext(c)

			token := extractToken(c)

			assert.Equal(t, tc.expectedToken, token)
		})
	}
}

// TestMin проверяет вспомогательную функцию min
func TestMin(t *testing.T) {
	testCases := []struct {
		a, b     int
		expected int
	}{
		{5, 10, 5},
		{10, 5, 5},
		{5, 5, 5},
		{0, 10, 0},
		{-5, 5, -5},
	}

	for _, tc := range testCases {
		result := min(tc.a, tc.b)
		assert.Equal(t, tc.expected, result)
	}
}
