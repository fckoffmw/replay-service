# MVC Quick Guide - Replay Service

## Краткая шпаргалка по архитектуре

### 📁 Где что находится?

| Что нужно сделать | Где это делать |
|-------------------|----------------|
| Добавить поле в структуру | `models/` |
| Добавить HTTP эндпоинт | `handlers/` |
| Добавить бизнес-логику | `services/` |
| Добавить SQL запрос | `repository/` |
| Работа с файлами | `storage/` |

---

## 🎯 Правила каждого слоя

### Models (Модели)
```go
// ✅ МОЖНО
type Game struct {
    ID   uuid.UUID `json:"id"`
    Name string    `json:"name"`
}

// ❌ НЕЛЬЗЯ
func (g *Game) Save() error { ... }  // Логика не в модели!
```

### Handlers (Контроллеры)
```go
// ✅ МОЖНО - тонкий слой
func (h *Handler) GetGames(c *gin.Context) {
    userID := c.MustGet("user_id").(uuid.UUID)
    games, err := h.gameService.GetUserGames(c.Request.Context(), userID)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed"})
        return
    }
    c.JSON(200, games)
}

// ❌ НЕЛЬЗЯ - бизнес-логика в контроллере
func (h *Handler) GetGames(c *gin.Context) {
    // Прямая работа с БД
    rows, _ := db.Query("SELECT * FROM games")
    // Работа с файлами
    os.Remove("/path/to/file")
    // Сложная логика
    if game.IsValid() && user.HasAccess() { ... }
}
```

### Services (Бизнес-логика)
```go
// ✅ МОЖНО - вся логика здесь
func (s *ReplayService) CreateReplay(...) (*models.Replay, error) {
    // Валидация
    if file.Size > maxSize {
        return nil, errors.New("file too large")
    }
    
    // Координация
    filePath, err := s.storage.SaveFile(...)
    if err != nil {
        return nil, err
    }
    
    // Транзакция
    if err := s.repo.Create(...); err != nil {
        s.storage.DeleteFile(filePath) // Rollback
        return nil, err
    }
    
    return replay, nil
}

// ❌ НЕЛЬЗЯ - работа с HTTP
func (s *Service) Create(c *gin.Context) { ... }
```

### Repository (Доступ к БД)
```go
// ✅ МОЖНО - только SQL
func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Game, error) {
    query := `SELECT id, name FROM games WHERE id = $1`
    var game models.Game
    err := r.db.QueryRow(ctx, query, id).Scan(&game.ID, &game.Name)
    return &game, err
}

// ❌ НЕЛЬЗЯ - бизнес-логика или работа с файлами
func (r *Repository) Create(game *Game) error {
    // Валидация бизнес-правил
    if game.Name == "" { ... }
    // Работа с файлами
    os.Remove(...)
}
```

### Storage (Файлы)
```go
// ✅ МОЖНО - только файловая система
func (s *FileStorage) SaveFile(file *multipart.FileHeader, path string) error {
    dst, err := os.Create(path)
    // ... копирование файла
    return err
}

// ❌ НЕЛЬЗЯ - работа с БД
func (s *Storage) SaveFile(...) error {
    db.Exec("INSERT INTO files ...")
}
```

---

## 🔄 Типичные сценарии

### Добавить новый эндпоинт

1. **Handler** - добавить метод:
```go
func (h *Handler) GetGameStats(c *gin.Context) {
    gameID := c.Param("game_id")
    stats, err := h.gameService.GetStats(c.Request.Context(), gameID)
    c.JSON(200, stats)
}
```

2. **Service** - добавить логику:
```go
func (s *GameService) GetStats(ctx context.Context, gameID uuid.UUID) (*Stats, error) {
    game, err := s.gameRepo.GetByID(ctx, gameID)
    replays, err := s.replayRepo.GetByGameID(ctx, gameID)
    // Вычисление статистики
    return stats, nil
}
```

3. **main.go** - зарегистрировать роут:
```go
gamesAPI.GET("/:game_id/stats", handler.GetGameStats)
```

### Добавить новую сущность (например, User)

1. **Model**: `models/user.go`
```go
type User struct {
    ID    uuid.UUID
    Name  string
    Email string
}
```

2. **Repository**: `repository/user_repository.go`
```go
type UserRepository struct { ... }
func (r *UserRepository) GetByID(...) { ... }
func (r *UserRepository) Create(...) { ... }
```

3. **Service**: `services/user_service.go`
```go
type UserService struct { ... }
func (s *UserService) RegisterUser(...) { ... }
```

4. **Handler**: добавить методы в `handlers/replay.go` или создать `handlers/user.go`

---

## 🚫 Частые ошибки

### ❌ Бизнес-логика в Handler
```go
// ПЛОХО
func (h *Handler) CreateGame(c *gin.Context) {
    // Валидация
    if name == "" { return }
    // Работа с БД
    db.Exec("INSERT ...")
    // Работа с файлами
    os.MkdirAll(...)
}
```

### ✅ Правильно
```go
// ХОРОШО
func (h *Handler) CreateGame(c *gin.Context) {
    name := c.PostForm("name")
    game, err := h.gameService.CreateGame(c.Request.Context(), userID, name)
    c.JSON(201, game)
}
```

---

### ❌ Repository с бизнес-логикой
```go
// ПЛОХО
func (r *Repository) CreateReplay(replay *Replay) error {
    // Валидация
    if replay.Title == "" {
        return errors.New("title required")
    }
    // Работа с файлами
    os.MkdirAll(...)
    // SQL
    db.Exec(...)
}
```

### ✅ Правильно
```go
// ХОРОШО - только SQL
func (r *Repository) Create(ctx context.Context, replay *Replay) error {
    query := `INSERT INTO replays (...) VALUES (...)`
    _, err := r.db.Exec(ctx, query, ...)
    return err
}
```

---

### ❌ Service работает с HTTP
```go
// ПЛОХО
func (s *Service) GetGames(c *gin.Context) {
    userID := c.Param("user_id")
    c.JSON(200, games)
}
```

### ✅ Правильно
```go
// ХОРОШО - работа с контекстом и данными
func (s *Service) GetUserGames(ctx context.Context, userID uuid.UUID) ([]Game, error) {
    return s.repo.GetByUserID(ctx, userID)
}
```

---

## 📊 Зависимости

```
main.go
  ↓
Handler (зависит от Service)
  ↓
Service (зависит от Repository + Storage)
  ↓
Repository → Database
Storage → File System
```

**Правило:** Зависимости только сверху вниз!

---

## 🧪 Тестирование

### Unit-тест Service
```go
func TestReplayService_CreateReplay(t *testing.T) {
    // Мокируем зависимости
    mockRepo := &MockReplayRepository{}
    mockStorage := &MockFileStorage{}
    
    service := NewReplayService(mockRepo, mockStorage)
    
    // Тестируем
    replay, err := service.CreateReplay(...)
    assert.NoError(t, err)
}
```

### Unit-тест Repository
```go
func TestReplayRepository_GetByID(t *testing.T) {
    // Используем тестовую БД
    db := setupTestDB(t)
    repo := NewReplayRepository(db)
    
    // Тестируем
    replay, err := repo.GetByID(ctx, id)
    assert.NoError(t, err)
}
```

---

## 📝 Чеклист при добавлении функции

- [ ] Определил структуру в `models/`
- [ ] Добавил SQL запросы в `repository/`
- [ ] Реализовал бизнес-логику в `services/`
- [ ] Добавил HTTP обработчик в `handlers/`
- [ ] Зарегистрировал роут в `main.go`
- [ ] Handler < 20 строк (иначе логика в service)
- [ ] Service не работает с HTTP
- [ ] Repository не содержит бизнес-логику
- [ ] Добавил логирование в service
- [ ] Обработал ошибки

---

## 🎓 Дополнительно

Подробное описание архитектуры: [architecture.md](architecture.md)
