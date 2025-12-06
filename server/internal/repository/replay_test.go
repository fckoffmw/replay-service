package repository

import (
	"context"
	"testing"

	"github.com/fckoffmw/replay-service/server/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReplayRepository_Create проверяет создание реплея
func TestReplayRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру для реплея
	game, err := gameRepo.Create(ctx, userID, "Test Game")
	require.NoError(t, err)
	
	// Создаем реплей
	title := "Epic Match"
	comment := "Best game ever"
	replay := &models.Replay{
		ID:           uuid.New(),
		Title:        &title,
		OriginalName: "match.rep",
		FilePath:     "user/game/replay.rep",
		SizeBytes:    1024,
		Compression:  "none",
		Compressed:   false,
		Comment:      &comment,
		GameID:       game.ID,
		UserID:       userID,
	}
	
	err = replayRepo.Create(ctx, replay)
	
	assert.NoError(t, err)
	assert.False(t, replay.UploadedAt.IsZero(), "UploadedAt должен быть установлен")
}

// TestReplayRepository_GetByGameID проверяет получение реплеев игры
func TestReplayRepository_GetByGameID(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	
	// Создаем несколько реплеев
	for i := 0; i < 3; i++ {
		replay := &models.Replay{
			ID:           uuid.New(),
			OriginalName: "replay.rep",
			FilePath:     "path/to/file",
			SizeBytes:    1024,
			Compression:  "none",
			Compressed:   false,
			GameID:       game.ID,
			UserID:       userID,
		}
		replayRepo.Create(ctx, replay)
	}
	
	// Получаем реплеи с лимитом 5
	replays, err := replayRepo.GetByGameID(ctx, game.ID, userID, 5)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, len(replays), "должно быть 3 реплея")
}

// TestReplayRepository_GetByGameID_Limit проверяет работу лимита
func TestReplayRepository_GetByGameID_Limit(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	
	// Создаем 10 реплеев
	for i := 0; i < 10; i++ {
		replay := &models.Replay{
			ID:           uuid.New(),
			OriginalName: "replay.rep",
			FilePath:     "path/to/file",
			SizeBytes:    1024,
			Compression:  "none",
			Compressed:   false,
			GameID:       game.ID,
			UserID:       userID,
		}
		replayRepo.Create(ctx, replay)
	}
	
	// Получаем только 3 реплея
	replays, err := replayRepo.GetByGameID(ctx, game.ID, userID, 3)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, len(replays), "должно вернуться ровно 3 реплея (лимит)")
}

// TestReplayRepository_GetByID проверяет получение одного реплея
func TestReplayRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	
	// Создаем реплей
	title := "Test Replay"
	comment := "Test Comment"
	originalReplay := &models.Replay{
		ID:           uuid.New(),
		Title:        &title,
		OriginalName: "test.rep",
		FilePath:     "path/to/test.rep",
		SizeBytes:    2048,
		Compression:  "none",
		Compressed:   false,
		Comment:      &comment,
		GameID:       game.ID,
		UserID:       userID,
	}
	replayRepo.Create(ctx, originalReplay)
	
	// Получаем реплей по ID
	replay, err := replayRepo.GetByID(ctx, originalReplay.ID, userID)
	
	assert.NoError(t, err)
	assert.NotNil(t, replay)
	assert.Equal(t, originalReplay.ID, replay.ID)
	assert.Equal(t, "Test Replay", *replay.Title)
	assert.Equal(t, "test.rep", replay.OriginalName)
	assert.Equal(t, int64(2048), replay.SizeBytes)
	assert.Equal(t, "Test Game", replay.GameName, "должно быть название игры из JOIN")
}

// TestReplayRepository_Update проверяет обновление реплея
func TestReplayRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру и реплей
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	replay := &models.Replay{
		ID:           uuid.New(),
		OriginalName: "test.rep",
		FilePath:     "path/to/file",
		SizeBytes:    1024,
		Compression:  "none",
		Compressed:   false,
		GameID:       game.ID,
		UserID:       userID,
	}
	replayRepo.Create(ctx, replay)
	
	// Обновляем title и comment
	newTitle := "Updated Title"
	newComment := "Updated Comment"
	err := replayRepo.Update(ctx, replay.ID, userID, &newTitle, &newComment)
	
	assert.NoError(t, err)
	
	// Проверяем, что данные обновились
	updated, _ := replayRepo.GetByID(ctx, replay.ID, userID)
	assert.Equal(t, "Updated Title", *updated.Title)
	assert.Equal(t, "Updated Comment", *updated.Comment)
}

// TestReplayRepository_Delete проверяет удаление реплея
func TestReplayRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру и реплей
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	replay := &models.Replay{
		ID:           uuid.New(),
		OriginalName: "test.rep",
		FilePath:     "path/to/file.rep",
		SizeBytes:    1024,
		Compression:  "none",
		Compressed:   false,
		GameID:       game.ID,
		UserID:       userID,
	}
	replayRepo.Create(ctx, replay)
	
	// Удаляем реплей
	filePath, err := replayRepo.Delete(ctx, replay.ID, userID)
	
	assert.NoError(t, err)
	assert.Equal(t, "path/to/file.rep", filePath, "должен вернуться путь к файлу")
	
	// Проверяем, что реплей удален
	_, err = replayRepo.GetByID(ctx, replay.ID, userID)
	assert.Error(t, err, "удаленный реплей не должен быть найден")
}

// TestReplayRepository_GetFilePathsByGameID проверяет получение путей файлов
func TestReplayRepository_GetFilePathsByGameID(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	
	// Создаем несколько реплеев с разными путями
	paths := []string{"path/to/replay1.rep", "path/to/replay2.rep", "path/to/replay3.rep"}
	for _, path := range paths {
		replay := &models.Replay{
			ID:           uuid.New(),
			OriginalName: "test.rep",
			FilePath:     path,
			SizeBytes:    1024,
			Compression:  "none",
			Compressed:   false,
			GameID:       game.ID,
			UserID:       userID,
		}
		replayRepo.Create(ctx, replay)
	}
	
	// Получаем все пути
	filePaths, err := replayRepo.GetFilePathsByGameID(ctx, game.ID, userID)
	
	assert.NoError(t, err)
	assert.Equal(t, 3, len(filePaths), "должно быть 3 пути")
	
	// Проверяем, что все пути присутствуют
	for _, expectedPath := range paths {
		found := false
		for _, actualPath := range filePaths {
			if actualPath == expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "путь %s должен быть в списке", expectedPath)
	}
}

// TestReplayRepository_CascadeDelete проверяет каскадное удаление
// Что тестируем: при удалении игры все её реплеи удаляются автоматически (ON DELETE CASCADE)
func TestReplayRepository_CascadeDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("пропускаем интеграционный тест в режиме short")
	}
	
	db := setupTestDB(t)
	defer db.Close()
	
	gameRepo := NewGameRepository(db)
	replayRepo := NewReplayRepository(db)
	
	userID := uuid.New()
	createTestUser(t, db, userID)
	defer cleanupTestData(t, db, userID)
	
	ctx := context.Background()
	
	// Создаем игру
	game, _ := gameRepo.Create(ctx, userID, "Test Game")
	
	// Создаем реплей
	replay := &models.Replay{
		ID:           uuid.New(),
		OriginalName: "test.rep",
		FilePath:     "path/to/file",
		SizeBytes:    1024,
		Compression:  "none",
		Compressed:   false,
		GameID:       game.ID,
		UserID:       userID,
	}
	replayRepo.Create(ctx, replay)
	
	// Удаляем игру
	gameRepo.Delete(ctx, game.ID, userID)
	
	// Проверяем, что реплей тоже удален (каскадно)
	_, err := replayRepo.GetByID(ctx, replay.ID, userID)
	assert.Error(t, err, "реплей должен быть удален каскадно вместе с игрой")
}
