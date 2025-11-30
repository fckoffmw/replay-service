# Структура проекта Replay Service

## Полная структура директорий

```
replay-service/
│
├── 📄 README.md                      # Основная документация
├── 📄 spec.md                        # Спецификация проекта
├── 📄 .env                           # Переменные окружения
├── 📄 .env.example                   # Пример конфигурации
├── 📄 go.mod                         # Go модули
├── 📄 go.sum                         # Зависимости
├── 🚀 start.sh                       # Скрипт запуска
│
├── 📁 client/                        # 🎨 VIEW - Frontend
│   ├── index.html                    # Главная страница
│   ├── player.html                   # Проигрыватель реплеев
│   ├── script.js                     # Клиентская логика
│   └── style.css                     # Стили
│
├── 📁 server/                        # 🔧 Backend (Go)
│   │
│   ├── 📁 cmd/                       # Точки входа
│   │   └── replay-service/
│   │       └── main.go               # 🚪 Главный файл приложения
│   │
│   ├── 📁 config/                    # ⚙️ Конфигурация
│   │   └── config.go                 # Загрузка настроек из .env
│   │
│   ├── 📁 internal/                  # 🔒 Внутренняя логика
│   │   │
│   │   ├── 📁 models/                # 📦 MODEL - Структуры данных
│   │   │   ├── game.go               # Модель игры
│   │   │   └── replay.go             # Модель реплея
│   │   │
│   │   ├── 📁 handlers/              # 🎮 CONTROLLER - HTTP обработчики
│   │   │   └── replay.go             # Все эндпоинты API
│   │   │
│   │   ├── 📁 services/              # 💼 BUSINESS LOGIC - Бизнес-логика
│   │   │   ├── game_service.go       # Логика работы с играми
│   │   │   └── replay_service.go     # Логика работы с реплеями
│   │   │
│   │   ├── 📁 repository/            # 🗄️ DATA ACCESS - Работа с БД
│   │   │   ├── game_repository.go    # SQL запросы для игр
│   │   │   └── replay_repository.go  # SQL запросы для реплеев
│   │   │
│   │   ├── 📁 storage/               # 💾 FILE STORAGE - Файловая система
│   │   │   └── file_storage.go       # Сохранение/удаление файлов
│   │   │
│   │   ├── 📁 middleware/            # 🔐 HTTP Middleware
│   │   │   └── auth.go               # Аутентификация по X-User-ID
│   │   │
│   │   ├── 📁 database/              # 🔌 Database Connection
│   │   │   └── database.go           # Подключение к PostgreSQL
│   │   │
│   │   └── 📁 logger/                # 📝 Logging
│   │       └── logger.go             # Структурированное логирование
│   │
│   └── 📁 migrations/                # 🔄 SQL Миграции
│       ├── 0001_init.up.sql          # Создание таблиц
│       └── 0001_init.down.sql        # Откат миграций
│
├── 📁 storage/                       # 💿 Файловое хранилище
│   └── {user_id}/
│       └── {game_id}/
│           └── {replay_id}.ext       # Файлы реплеев
│
├── 📁 deploy/                        # 🐳 Deployment
│   ├── docker/
│   │   └── Dockerfile                # Образ приложения
│   └── docker-compose/
│       └── docker-compose.yml        # Оркестрация контейнеров
│
└── 📁 docs/                          # 📚 Документация
    ├── README.md                     # Обзор документации
    ├── architecture.md               # Подробное описание архитектуры
    ├── mvc-guide.md                  # Краткая шпаргалка по MVC
    ├── project-structure.md          # Этот файл
    ├── api-specification.md          # Спецификация REST API
    ├── api-examples.http             # Примеры HTTP запросов
    ├── storage-structure.md          # Структура файлового хранилища
    └── configuration.md              # Настройка приложения
```

---

## Архитектурные слои

### 🎨 View Layer (Представление)
```
client/
├── index.html      → Список игр и реплеев
├── player.html     → Проигрыватель видео
├── script.js       → API клиент, DOM манипуляции
└── style.css       → Визуальное оформление
```

**Технологии:** Vanilla JavaScript, HTML5, CSS3

---

### 🎮 Controller Layer (Контроллер)
```
server/internal/handlers/
└── replay.go
    ├── GetGames()         → GET /api/v1/games
    ├── CreateGame()       → POST /api/v1/games
    ├── UpdateGame()       → PUT /api/v1/games/:id
    ├── DeleteGame()       → DELETE /api/v1/games/:id
    ├── GetReplays()       → GET /api/v1/games/:id/replays
    ├── CreateReplay()     → POST /api/v1/games/:id/replays
    ├── GetReplay()        → GET /api/v1/replays/:id
    ├── UpdateReplay()     → PUT /api/v1/replays/:id
    ├── DeleteReplay()     → DELETE /api/v1/replays/:id
    └── GetReplayFile()    → GET /api/v1/replays/:id/file
```

**Ответственность:**
- Парсинг HTTP запросов
- Валидация входных данных
- Вызов методов сервисов
- Формирование HTTP ответов

---

### 💼 Service Layer (Бизнес-логика)
```
server/internal/services/
├── game_service.go
│   ├── GetUserGames()     → Получить игры пользователя
│   ├── CreateGame()       → Создать игру
│   ├── UpdateGame()       → Обновить название игры
│   └── DeleteGame()       → Удалить игру + все реплеи
│
└── replay_service.go
    ├── GetGameReplays()   → Получить реплеи игры
    ├── GetReplay()        → Получить один реплей
    ├── CreateReplay()     → Создать реплей (файл + БД)
    ├── UpdateReplay()     → Обновить метаданные
    ├── DeleteReplay()     → Удалить реплей (БД + файл)
    └── GetReplayFilePath() → Получить путь к файлу
```

**Ответственность:**
- Вся бизнес-логика
- Координация repository и storage
- Транзакции и откаты
- Логирование операций

---

### 🗄️ Repository Layer (Доступ к данным)
```
server/internal/repository/
├── game_repository.go
│   ├── GetByUserID()      → SELECT games WHERE user_id = ?
│   ├── Create()           → INSERT INTO games
│   ├── Update()           → UPDATE games SET name = ?
│   └── Delete()           → DELETE FROM games
│
└── replay_repository.go
    ├── GetByGameID()      → SELECT replays WHERE game_id = ?
    ├── GetByID()          → SELECT replays WHERE id = ?
    ├── Create()           → INSERT INTO replays
    ├── Update()           → UPDATE replays
    ├── Delete()           → DELETE FROM replays
    └── GetFilePathsByGameID() → SELECT file_path FROM replays
```

**Ответственность:**
- SQL запросы
- Маппинг результатов в модели
- Работа с pgx

---

### 💾 Storage Layer (Файловое хранилище)
```
server/internal/storage/
└── file_storage.go
    ├── SaveReplayFile()   → Сохранить файл на диск
    ├── DeleteFile()       → Удалить файл
    ├── DeleteFiles()      → Удалить несколько файлов
    ├── GetFilePath()      → Получить полный путь
    └── FileExists()       → Проверить существование
```

**Ответственность:**
- Операции с файловой системой
- Создание директорий
- Копирование файлов

---

### 📦 Model Layer (Модели данных)
```
server/internal/models/
├── game.go
│   └── type Game struct {
│       ID          uuid.UUID
│       Name        string
│       UserID      uuid.UUID
│       CreatedAt   time.Time
│       ReplayCount int
│   }
│
└── replay.go
    └── type Replay struct {
        ID           uuid.UUID
        Title        *string
        OriginalName string
        FilePath     string
        SizeBytes    int64
        UploadedAt   time.Time
        Compression  string
        Compressed   bool
        Comment      *string
        GameID       uuid.UUID
        GameName     string
        UserID       uuid.UUID
    }
```

**Ответственность:**
- Определение структур данных
- JSON/DB маппинг через теги

---

## Поток данных

### Пример: Загрузка реплея

```
1. Client (View)
   ↓ POST /api/v1/games/{id}/replays
   
2. Handler (Controller)
   ├─ Парсит multipart form
   ├─ Извлекает file, title, comment
   └─ Вызывает replayService.CreateReplay()
   
3. ReplayService (Business Logic)
   ├─ Создает модель Replay
   ├─ Вызывает storage.SaveReplayFile()
   │  └─ Storage сохраняет файл → возвращает путь
   ├─ Вызывает replayRepo.Create()
   │  └─ Repository сохраняет в БД
   └─ При ошибке БД: откатывает файл
   
4. Handler
   └─ Возвращает JSON с ID реплея
   
5. Client
   └─ Обновляет список реплеев
```

---

## Зависимости между компонентами

```
main.go
  │
  ├─ Создает DB connection
  ├─ Создает FileStorage
  │
  ├─ Создает Repositories
  │   ├─ GameRepository(db)
  │   └─ ReplayRepository(db)
  │
  ├─ Создает Services
  │   ├─ GameService(gameRepo, replayRepo, storage)
  │   └─ ReplayService(replayRepo, storage)
  │
  ├─ Создает Handler
  │   └─ Handler(gameService, replayService)
  │
  └─ Регистрирует роуты
      └─ Gin Router → Handler methods
```

---

## Конфигурация

### Переменные окружения (.env)
```env
PORT=8080                                    # Порт сервера
DB_DSN=postgres://user:pass@host:port/db    # Строка подключения к БД
STORAGE_DIR=./storage                        # Директория для файлов
LOG_LEVEL=debug                              # Уровень логирования
```

### Загрузка конфигурации
```
config/config.go
  ├─ Ищет .env в корне проекта
  ├─ Загружает переменные окружения
  ├─ Применяет значения по умолчанию
  └─ Валидирует обязательные параметры
```

---

## База данных

### Таблицы
```sql
users
  ├─ id (uuid, PK)
  ├─ created_at (timestamp)
  
games
  ├─ id (uuid, PK)
  ├─ name (varchar)
  ├─ user_id (uuid, FK → users.id)
  ├─ created_at (timestamp)
  └─ UNIQUE(user_id, name)
  
replays
  ├─ id (uuid, PK)
  ├─ title (varchar, nullable)
  ├─ original_name (varchar)
  ├─ file_path (varchar)
  ├─ size_bytes (bigint)
  ├─ uploaded_at (timestamp)
  ├─ compression (varchar)
  ├─ compressed (boolean)
  ├─ comment (text, nullable)
  ├─ game_id (uuid, FK → games.id, ON DELETE CASCADE)
  └─ user_id (uuid, FK → users.id)
```

### Миграции
```
server/migrations/
├── 0001_init.up.sql      → Создание таблиц
└── 0001_init.down.sql    → Откат (DROP TABLE)
```

Применяются автоматически при запуске через Docker Compose.

---

## Middleware

### AuthMiddleware
```
Проверяет заголовок X-User-ID
  ├─ Если отсутствует → 401 Unauthorized
  ├─ Если невалидный UUID → 400 Bad Request
  └─ Если валидный → добавляет в context
```

### CORS
```
Разрешает запросы с любых источников
  ├─ Access-Control-Allow-Origin: *
  ├─ Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
  └─ Access-Control-Allow-Headers: Content-Type, X-User-ID
```

---

## Deployment

### Docker Compose
```yaml
services:
  postgres:
    image: postgres:16
    ports: 5431:5432
    volumes: postgres_data
    
  app:
    build: ./deploy/docker
    ports: 8080:8080
    depends_on: postgres
    volumes: ./storage:/app/storage
```

### Запуск
```bash
./start.sh                    # Запуск всех сервисов
docker compose down           # Остановка
docker compose down -v        # Остановка + удаление данных
```

---

## Логирование

### Уровни логов
- `DEBUG` - детальная информация для отладки
- `INFO` - важные события (запуск, подключение к БД)
- `ERROR` - ошибки выполнения

### Формат
```
[Service/Method] message: details
```

Примеры:
```
[GameService] CreateGame: user_id=..., name=...
[GameService] CreateGame SUCCESS: game_id=...
[GameService] CreateGame ERROR: failed to create game
```

---

## Тестирование

### Структура тестов (будущее)
```
server/internal/
├── services/
│   ├── game_service.go
│   └── game_service_test.go      # Unit-тесты с моками
├── repository/
│   ├── game_repository.go
│   └── game_repository_test.go   # Интеграционные тесты с БД
└── handlers/
    ├── replay.go
    └── replay_test.go             # HTTP тесты
```

---

## Полезные команды

### Разработка
```bash
# Запуск сервера локально
go run server/cmd/replay-service/main.go

# Сборка
go build -o bin/replay-service server/cmd/replay-service/main.go

# Проверка зависимостей
go mod tidy

# Форматирование кода
go fmt ./...

# Линтер
golangci-lint run
```

### Docker
```bash
# Сборка образа
docker build -f deploy/docker/Dockerfile -t replay-service .

# Запуск контейнера
docker run -p 8080:8080 --env-file .env replay-service

# Логи
docker compose logs -f app
```

### База данных
```bash
# Подключение к БД
psql -h localhost -p 5431 -U replay -d replay

# Применить миграции вручную
migrate -path server/migrations -database "postgres://..." up

# Откатить миграции
migrate -path server/migrations -database "postgres://..." down
```

---

## Дополнительные ресурсы

- [Architecture](architecture.md) - подробное описание архитектуры
- [MVC Guide](mvc-guide.md) - шпаргалка по MVC
- [API Specification](api-specification.md) - документация API
- [Configuration](configuration.md) - настройка приложения
