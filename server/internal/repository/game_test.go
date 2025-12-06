package repository

import (
	"context"
	"testing"

	"github.com/fckoffmw/replay-service/server/internal/database"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestDB создает подключение к тестовой БД
// Зачем: интеграционные тесты проверяют реальную работу с PostgreSQL
// Примечание: требуется запущенная тестовая БД
func setupTestDB(t *testing.T) *database.DB {
	// Используем переменную окружения для тестовой БД
	// Пример: TEST_DB_DSN=postgres://replay:replay@localhost:5431/replay_test
	dsn := "postgres://replay:replay@localhost:5431/replay?sslmode=disable"
	
	db, err := database.Connect(context.Background(), dsn)
	require.NoError(t, err, "не удалось подключиться к тестовой БД")
	
	return db
}

// createTestUser создает тестового пользователя в БД
func createTestUser(t *testing.T, db *database.DB, userID uuid.UUID) {
	ctx := context.Background()
	
	// Создаем пользователя для тестов
	_, err := db.Pool.Exec(ctx, 
		"INSERT INTO users (id, login, password_hash) VALUES ($1, $2, $3) ON CONFLICT (id) DO NOTHING",
		userID, "test_user_"+userID.String(), "test_hash")
	require.NoError(t, err)
}

// cleanupTestData очищает тестовые данные после теста
func cleanupTestData(t *testing.T, db *database.DB, userID uuid.UUID) {
	ctx := context.Background()
	
	// Удаляем пользователя (игры и реплеи удалятся каскадно)
	_, err := db.Pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	require.NoError(t, err)
}

// TestGameRepository_Create проверяет создание игры
// Что тестируем: SQL INSERT работает корректно, возвращается ID и timestamp
func TestGameRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	gameName := "Test Game"
	
	// Act
	game, err := repo.Create(ctx, userID, gameName)
	
	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, game)
	assert.NotEqual(t, uuid.Nil, game.ID, "ID должен быть сгенерирован")
	assert.Equal(t, gameName, game.Name)
	assert.Equal(t, userID, game.UserID)
	assert.False(t, game.CreatedAt.IsZero(), "CreatedAt должен быть установлен")
}

// TestGameRepository_Create_Duplicate проверяет обработку дубликатов
// Что тестируем: ON CONFLICT работает (игра с таким именем уже существует)
func TestGameRepository_Create_Duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	gameName := "Duplicate Game"
	
	// Создаем игру первый раз
	game1, err := repo.Create(ctx, userID, gameName)
	require.NoError(t, err)
	
	// Создаем игру с тем же именем второй раз
	game2, err := repo.Create(ctx, userID, gameName)
	
	// Должна вернуться та же игра (ON CONFLICT DO UPDATE)
	assert.NoError(t, err)
	assert.Equal(t, game1.ID, game2.ID, "ID должен остаться тем же")
}

// TestGameRepository_GetByUserID проверяет получение списка игр
// Что тестируем: JOIN с replays работает, подсчет реплеев корректен
func TestGameRepository_GetByUserID(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем несколько игр
	game1, _ := repo.Create(ctx, userID, "Game 1")
	game2, _ := repo.Create(ctx, userID, "Game 2")
	
	// Получаем список игр
	games, err := repo.GetByUserID(ctx, userID)
	
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(games), 2, "должно быть минимум 2 игры")
	
	// Проверяем, что наши игры в списке
	foundGame1 := false
	foundGame2 := false
	for _, g := range games {
		if g.ID == game1.ID {
			foundGame1 = true
			assert.Equal(t, 0, g.ReplayCount, "у новой игры должно быть 0 реплеев")
		}
		if g.ID == game2.ID {
			foundGame2 = true
		}
	}
	
	assert.True(t, foundGame1, "Game 1 должна быть в списке")
	assert.True(t, foundGame2, "Game 2 должна быть в списке")
}

// TestGameRepository_Update проверяет обновление игры
func TestGameRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := repo.Create(ctx, userID, "Original Name")
	
	// Обновляем название
	newName := "Updated Name"
	err := repo.Update(ctx, game.ID, userID, newName)
	
	assert.NoError(t, err)
	
	// Проверяем, что название изменилось
	games, _ := repo.GetByUserID(ctx, userID)
	for _, g := range games {
		if g.ID == game.ID {
			assert.Equal(t, newName, g.Name)
		}
	}
}

// TestGameRepository_Update_NotFound проверяет обновление несуществующей игры
func TestGameRepository_Update_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	
	ctx := context.Background()
	fakeGameID := uuid.New()
	fakeUserID := uuid.New()
	
	err := repo.Update(ctx, fakeGameID, fakeUserID, "New Name")
	
	assert.Error(t, err, "должна быть ошибка при обновлении несуществующей игры")
}

// TestGameRepository_Delete проверяет удаление игры
func TestGameRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := repo.Create(ctx, userID, "Game to Delete")
	
	// Удаляем игру
	err := repo.Delete(ctx, game.ID, userID)
	
	assert.NoError(t, err)
	
	// Проверяем, что игра удалена
	games, _ := repo.GetByUserID(ctx, userID)
	for _, g := range games {
		assert.NotEqual(t, game.ID, g.ID, "удаленная игра не должна быть в списке")
	}
}

// TestGameRepository_Delete_NotFound проверяет удаление несуществующей игры
func TestGameRepository_Delete_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	repo := NewGameRepository(db)
	
	ctx := context.Background()
	fakeGameID := uuid.New()
	fakeUserID := uuid.New()
	
	err := repo.Delete(ctx, fakeGameID, fakeUserID)
	
	assert.Error(t, err, "должна быть ошибка при удалении несуществующей игры")
}
