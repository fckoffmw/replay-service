# REST API Спецификация

## Эндпоинты для работы с реплеями

### 1. Загрузка реплея
```http
POST /replays/upload
Content-Type: multipart/form-data
```

**Параметры:**
- `file` (required) - файл реплея
- `title` (optional) - название реплея
- `game_name` (optional) - название игры

**Пример запроса:**
```bash
curl -X POST http://localhost:8080/replays/upload \
  -F "file=@replay.rep" \
  -F "title=Epic comeback" \
  -F "game_name=Dota 2"
```

**Ответ (201 Created):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Epic comeback",
  "game_name": "Dota 2",
  "original_name": "replay.rep",
  "compression": "gzip",
  "compressed": true
}
```

---

### 2. Получить список всех реплеев
```http
GET /replays
```

**Query параметры:**
- `game` (optional) - фильтр по названию игры

**Примеры запросов:**
```bash
# Все реплеи
curl http://localhost:8080/replays

# Реплеи конкретной игры
curl http://localhost:8080/replays?game=dota-2
```

**Ответ (200 OK):**
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Epic comeback",
    "game_name": "Dota 2",
    "original_name": "replay.rep",
    "compression": "gzip",
    "compressed": true
  },
  {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "title": "First win",
    "game_name": "Counter-Strike",
    "original_name": "match_01.rep",
    "compression": "gzip",
    "compressed": true
  }
]
```

---

### 3. Получить конкретный реплей
```http
GET /replays/:id
```

**Пример запроса:**
```bash
curl http://localhost:8080/replays/550e8400-e29b-41d4-a716-446655440000
```

**Ответ (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Epic comeback",
  "game_name": "Dota 2",
  "original_name": "replay.rep",
  "compression": "gzip",
  "compressed": true,
  "download_url": "/replays/550e8400-e29b-41d4-a716-446655440000/download"
}
```

---

### 4. Скачать файл реплея
```http
GET /replays/:id/download
```

**Пример запроса:**
```bash
curl -O http://localhost:8080/replays/550e8400-e29b-41d4-a716-446655440000/download
```

**Ответ (200 OK):**
- Content-Type: `application/gzip` или `application/octet-stream`
- Content-Disposition: `attachment; filename="epic-comeback.rep.gz"`
- Body: бинарный файл

---

### 5. Обновить метаданные реплея
```http
PATCH /replays/:id
Content-Type: application/json
```

**Body:**
```json
{
  "title": "New title",
  "game_name": "Dota 2"
}
```

**Пример запроса:**
```bash
curl -X PATCH http://localhost:8080/replays/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"title": "Updated title", "game_name": "Dota 2"}'
```

**Ответ (200 OK):**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "title": "Updated title",
  "game_name": "Dota 2",
  "original_name": "replay.rep",
  "compression": "gzip",
  "compressed": true
}
```

---

### 6. Удалить реплей
```http
DELETE /replays/:id
```

**Пример запроса:**
```bash
curl -X DELETE http://localhost:8080/replays/550e8400-e29b-41d4-a716-446655440000
```

**Ответ (204 No Content)**

---

## Эндпоинты для работы с играми

### 7. Получить список игр
```http
GET /games
```

**Пример запроса:**
```bash
curl http://localhost:8080/games
```

**Ответ (200 OK):**
```json
{
  "games": [
    {
      "name": "Dota 2",
      "normalized_name": "dota-2",
      "replay_count": 15
    },
    {
      "name": "Counter-Strike",
      "normalized_name": "counter-strike",
      "replay_count": 8
    }
  ]
}
```

---

### 8. Получить статистику по игре
```http
GET /games/:game_name
```

**Пример запроса:**
```bash
curl http://localhost:8080/games/dota-2
```

**Ответ (200 OK):**
```json
{
  "name": "Dota 2",
  "normalized_name": "dota-2",
  "replay_count": 15,
  "total_size_bytes": 524288000,
  "latest_replay": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "Epic comeback",
    "uploaded_at": "2025-11-28T10:30:00Z"
  }
}
```

---

## Коды ошибок

### 400 Bad Request
```json
{
  "error": "Invalid request",
  "details": "game_name is required"
}
```

### 404 Not Found
```json
{
  "error": "Replay not found",
  "replay_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "message": "Failed to save file"
}
```

---

## Примеры использования

### Загрузить реплей с метаданными
```bash
curl -X POST http://localhost:8080/replays/upload \
  -F "file=@my_game.rep" \
  -F "title=My best game" \
  -F "game_name=Dota 2"
```

### Получить все реплеи Dota 2
```bash
curl http://localhost:8080/replays?game=dota-2
```

### Скачать реплей
```bash
curl -O -J http://localhost:8080/replays/550e8400-e29b-41d4-a716-446655440000/download
```

### Переименовать реплей
```bash
curl -X PATCH http://localhost:8080/replays/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{"title": "Best game ever"}'
```
