# Архитектура проекта Replay Service

## Обзор

Проект построен на основе паттерна **MVC (Model-View-Controller)** с дополнительными слоями для разделения ответственности и улучшения поддерживаемости кода.

## Архитектурные слои

```
┌─────────────────────────────────────────────────────────┐
│                    Client (View)                        │
│                   HTML/CSS/JavaScript                   │
└─────────────────────┬───────────────────────────────────┘
                      │ HTTP Requests
                      ▼
┌─────────────────────────────────────────────────────────┐
│              Handlers (Controller)                      │
│          Маршрутизация HTTP запросов                    │
└─────────────────────┬───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│              Services (Business Logic)                  │
│         Вся бизнес-логика приложения                    │
└─────────┬───────────────────────────┬───────────────────┘
          │                           │
          ▼                           ▼
┌──────────────────────┐    ┌──────────────────────┐
│   Repository Layer   │    │   Storage Layer      │
│   (Database Access)  │    │   (File System)      │
└──────────┬───────────┘    └──────────┬───────────┘
           │                           │
           ▼                           ▼
    ┌──────────┐              ┌──────────────┐
    │PostgreSQL│              │File System   │
    └──────────┘              └──────────────┘
```

## Описание слоев

### 1. Model (Модель данных)

**Расположение:** `server/internal/models/`

**Ответственность:**
- Определение структур данных
- Описание полей и их типов
- JSON/DB маппинг через теги

**Файлы:**
- `game.go` - модель игры
- `replay.go` - модель реплея

**Пример:**
```go
type Replay struct {
    ID           uuid.UUID `json:"id"`
    Title        *string   `json:"title,omitempty"`
    OriginalName string    `json:"original_name"`
    FilePath     string    `json:"-"`
    SizeBytes    int64     `json:"size_bytes"`
    UploadedAt   time.Time `json:"uploaded_at"`
    GameID       uuid.UUID `json:"game_id"`
    UserID       uuid.UUID `json:"-"`
}
```

**Правила:**
- ❌ НЕ содержит бизнес-логику
- ❌ НЕ содержит методы работы с БД
- ✅ Только определения структур

---

### 2. View (Представление)

**Расположение:** `client/`

**Ответственность:**
- Отображение данных пользователю
- Обработка пользовательского ввода
- Взаимодействие с API через HTTP

**Файлы:**
- `index.html` - главная страница со списком игр
- `player.html` - страница проигрывателя реплеев
- `script.js` - клиентская логика
- `style.css` - стили интерфейса

**Технологии:**
- Vanilla JavaScript (без фреймворков)
- HTML5
- CSS3

---

### 3. Controller (Контроллер)

**Расположение:** `server/internal/handlers/`

**Ответственность:**
- Прием HTTP запросов
- Парсинг параметров и тела запроса
- Базовая валидация входных данных
- Вызов методов сервисов
- Формирование HTTP ответов

**Файлы:**
- `replay.go` - все HTTP обработчики

**Пример:**
```go
func (h *Handler) GetGames(c *gin.Context) {
    userID := c.MustGet("user_id").(uuid.UUID)
    
    games, err := h.gameService.GetUserGames(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get games"})
        return
    }
    
    c.JSON(http.StatusOK, games)
}
```

**Правила:**
- ✅ Тонкий слой (5-15 строк на метод)
- ❌ НЕ содержит бизнес-логику
- ❌ НЕ работает напрямую с БД или файлами
- ✅ Только маршрутизация к сервисам

---

### 4. Service Layer (Бизнес-логика)

**Расположение:** `server/internal/services/`

**Ответственность:**
- Вся бизнес-логика приложения
- Валидация бизнес-правил
- Координация между repository и storage
- Транзакционность операций
- Обработка ошибок и логирование

**Файлы:**
- `game_service.go` - логика работы с играми
- `replay_service.go` - логика работы с реплеями

**Пример:**
```go
func (s *ReplayService) CreateReplay(
    ctx context.Context,
    file *multipart.FileHeader,
    gameID, userID uuid.UUID,
    title, comment string,
) (*models.Replay, error) {
    // Создание модели
    replay := &models.Replay{...}
    
    // Сохранение файла
    filePath, err := s.storage.SaveReplayFile(file, userID, gameID, replay.ID)
    if err != nil {
        return nil, fmt.Errorf("failed to save file: %w", err)
    }
    
    // Сохранение в БД
    if err := s.replayRepo.Create(ctx, replay); err != nil {
        s.storage.DeleteFile(filePath) // Rollback
        return nil, fmt.Errorf("failed to create replay: %w", err)
    }
    
    return replay, nil
}
```

**Правила:**
- ✅ Содержит всю бизнес-логику
- ✅ Координирует работу repository и storage
- ✅ Обрабатывает транзакции и откаты
- ✅ Логирует важные операции
- ❌ НЕ работает напрямую с HTTP

---

### 5. Repository Layer (Доступ к данным)

**Расположение:** `server/internal/repository/`

**Ответственность:**
- Инкапсуляция SQL запросов
- CRUD операции с БД
- Маппинг результатов в модели
- Работа только с одной таблицей/сущностью

**Файлы:**
- `game_repository.go` - работа с таблицей `games`
- `replay_repository.go` - работа с таблицей `replays`

**Пример:**
```go
func (r *ReplayRepository) GetByID(ctx context.Context, replayID, userID uuid.UUID) (*models.Replay, error) {
    query := `
        SELECT r.id, r.title, r.original_name, r.file_path, ...
        FROM replays r
        WHERE r.id = $1 AND r.user_id = $2
    `
    
    var replay models.Replay
    err := r.db.Pool.QueryRow(ctx, query, replayID, userID).Scan(...)
    if err != nil {
        return nil, fmt.Errorf("failed to get replay: %w", err)
    }
    
    return &replay, nil
}
```

**Правила:**
- ✅ Один repository на одну таблицу
- ✅ Только SQL запросы и маппинг
- ❌ НЕ содержит бизнес-логику
- ❌ НЕ работает с файлами

---

### 6. Storage Layer (Файловое хранилище)

**Расположение:** `server/internal/storage/`

**Ответственность:**
- Сохранение файлов на диск
- Удаление файлов
- Управление директориями
- Формирование путей к файлам

**Файлы:**
- `file_storage.go` - работа с файловой системой

**Пример:**
```go
func (s *FileStorage) SaveReplayFile(
    file *multipart.FileHeader,
    userID, gameID, replayID uuid.UUID,
) (string, error) {
    ext := filepath.Ext(file.Filename)
    fileName := replayID.String() + ext
    relPath := filepath.Join(userID.String(), gameID.String(), fileName)
    fullPath := filepath.Join(s.baseDir, relPath)
    
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return "", fmt.Errorf("failed to create directory: %w", err)
    }
    
    // Сохранение файла...
    
    return relPath, nil
}
```

**Правила:**
- ✅ Только операции с файловой системой
- ✅ Возвращает относительные пути
- ❌ НЕ работает с БД
- ❌ НЕ содержит бизнес-логику

---

## Поток данных

### Пример: Создание реплея

1. **Client** отправляет POST запрос с файлом
2. **Handler** (Controller):
   - Парсит multipart form
   - Извлекает параметры
   - Вызывает `replayService.CreateReplay()`
3. **Service**:
   - Создает модель `Replay`
   - Вызывает `storage.SaveReplayFile()` для сохранения файла
   - Вызывает `replayRepo.Create()` для сохранения метаданных
   - При ошибке БД откатывает файл
4. **Storage** сохраняет файл и возвращает путь
5. **Repository** сохраняет запись в БД
6. **Service** возвращает созданный объект
7. **Handler** формирует JSON ответ
8. **Client** получает ID созданного реплея

---

## Преимущества архитектуры

### ✅ Разделение ответственности (Separation of Concerns)
Каждый слой решает только свою задачу, что упрощает понимание кода.

### ✅ Тестируемость
Легко писать unit-тесты, мокируя зависимости:
```go
// Тест сервиса с мок-репозиторием
mockRepo := &MockReplayRepository{}
service := NewReplayService(mockRepo, mockStorage)
```

### ✅ Поддерживаемость
Изменения в одном слое не влияют на другие. Например, можно:
- Заменить PostgreSQL на MySQL (меняем только repository)
- Заменить локальное хранилище на S3 (меняем только storage)
- Добавить GraphQL API (добавляем новые handlers)

### ✅ Масштабируемость
Легко добавлять новую функциональность:
- Новая сущность = новые model + repository + service + handlers
- Новый эндпоинт = новый метод в handler + service

### ✅ Читаемость
Понятная структура проекта - новый разработчик быстро разберется.

---

## Зависимости между слоями

```
Handler → Service → Repository → Database
                 → Storage → File System
```

**Правило:** Зависимости идут только сверху вниз, никогда наоборот.

❌ **Неправильно:**
```go
// Repository вызывает Service
func (r *Repository) Create() {
    service.DoSomething() // ПЛОХО!
}
```

✅ **Правильно:**
```go
// Service вызывает Repository
func (s *Service) Create() {
    repo.Save() // ХОРОШО!
}
```

---

## Рекомендации по разработке

### При добавлении новой функции:

1. **Model** - определите структуру данных
2. **Repository** - добавьте методы работы с БД
3. **Service** - реализуйте бизнес-логику
4. **Handler** - добавьте HTTP обработчик
5. **View** - обновите клиентскую часть

### При рефакторинге:

- Если логика в handler > 15 строк → переносите в service
- Если service работает с БД напрямую → выносите в repository
- Если repository содержит бизнес-логику → переносите в service

---

## Технологии

- **Backend:** Go 1.24, Gin (HTTP), pgx (PostgreSQL)
- **Frontend:** Vanilla JavaScript, HTML5, CSS3
- **Database:** PostgreSQL 16
- **Storage:** Local File System
- **Deployment:** Docker, Docker Compose

---

## Дополнительные слои

### Middleware
**Расположение:** `server/internal/middleware/`

Обработка сквозной функциональности:
- Аутентификация (`AuthMiddleware`)
- CORS
- Логирование запросов
- Rate limiting (в будущем)

### Database
**Расположение:** `server/internal/database/`

Управление подключением к БД:
- Создание connection pool
- Ping проверка
- Graceful shutdown

### Logger
**Расположение:** `server/internal/logger/`

Структурированное логирование:
- Уровни логов (debug, info, error)
- Контекстная информация
- Форматирование вывода

### Config
**Расположение:** `server/config/`

Управление конфигурацией:
- Загрузка из .env
- Валидация параметров
- Значения по умолчанию
