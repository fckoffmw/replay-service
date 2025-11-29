# REST API Specification

Base URL: `http://localhost:8080/api/v1`

Все запросы требуют заголовок `X-User-ID` с UUID пользователя.
По умолчанию используется: `00000000-0000-0000-0000-000000000001`

## Games

### Получить список игр

```http
GET /api/v1/games
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Response 200:**
```json
[
  {
    "id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "name": "Counter-Strike 2",
    "created_at": "2025-11-09T14:00:00Z",
    "replay_count": 2
  },
  {
    "id": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
    "name": "Dota 2",
    "created_at": "2025-11-14T14:00:00Z",
    "replay_count": 1
  }
]
```

### Создать игру

```http
POST /api/v1/games
Content-Type: application/json
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Body:**
```json
{
  "name": "Valorant"
}
```

**Response 201:**
```json
{
  "id": "cccccccc-cccc-cccc-cccc-cccccccccccc",
  "name": "Valorant",
  "created_at": "2025-11-29T15:00:00Z"
}
```

### Обновить игру

```http
PUT /api/v1/games/{game_id}
Content-Type: application/json
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Body:**
```json
{
  "name": "Counter-Strike 2 Updated"
}
```

**Response 200:**
```json
{
  "message": "updated"
}
```

### Удалить игру

```http
DELETE /api/v1/games/{game_id}
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Response 200:**
```json
{
  "message": "deleted"
}
```

Удаляет игру и все её реплеи (файлы и записи в БД).

## Replays

### Получить реплеи игры

```http
GET /api/v1/games/{game_id}/replays?limit=5
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Query Parameters:**
- `limit` (optional, default: 5) - количество реплеев

**Response 200:**
```json
[
  {
    "id": "10000000-0000-0000-0000-000000000001",
    "title": "Epic comeback",
    "original_name": "match_2024_01_15.rep",
    "size_bytes": 1048576,
    "uploaded_at": "2025-11-24T14:00:00Z",
    "compression": "none",
    "compressed": false,
    "game_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
  }
]
```

### Получить детали реплея

```http
GET /api/v1/replays/{replay_id}
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Response 200:**
```json
{
  "id": "10000000-0000-0000-0000-000000000001",
  "title": "Epic comeback",
  "original_name": "match_2024_01_15.rep",
  "comment": "Amazing clutch in overtime",
  "size_bytes": 1048576,
  "uploaded_at": "2025-11-24T14:00:00Z",
  "compression": "none",
  "compressed": false,
  "game_id": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "game_name": "Counter-Strike 2"
}
```

### Загрузить реплей

```http
POST /api/v1/games/{game_id}/replays
Content-Type: multipart/form-data
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Form Data:**
- `file` (required) - файл реплея
- `title` (optional) - название реплея
- `comment` (optional) - комментарий

**Response 201:**
```json
{
  "id": "20000000-0000-0000-0000-000000000001"
}
```

### Обновить реплей

```http
PUT /api/v1/replays/{replay_id}
Content-Type: multipart/form-data
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Form Data:**
- `title` (optional) - новое название
- `comment` (optional) - новый комментарий

**Response 200:**
```json
{
  "message": "updated"
}
```

### Удалить реплей

```http
DELETE /api/v1/replays/{replay_id}
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Response 200:**
```json
{
  "message": "deleted"
}
```

### Скачать файл реплея

```http
GET /api/v1/replays/{replay_id}/file
```

**Headers:**
```
X-User-ID: 00000000-0000-0000-0000-000000000001
```

**Response 200:**
- Content-Type: `application/octet-stream`
- Content-Disposition: `attachment; filename="original_name.rep"`
- Body: binary file

## Health Check

```http
GET /healthz
```

**Response 200:**
```json
{
  "status": "ok"
}
```

## Error Responses

### 400 Bad Request
```json
{
  "error": "invalid game_id"
}
```

### 404 Not Found
```json
{
  "error": "replay not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "failed to create replay"
}
```

## Examples

### Создать игру и загрузить реплей

```bash
# 1. Создать игру
GAME_ID=$(curl -s -X POST http://localhost:8080/api/v1/games \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -H "Content-Type: application/json" \
  -d '{"name":"Dota 2"}' | jq -r '.id')

# 2. Загрузить реплей
curl -X POST http://localhost:8080/api/v1/games/$GAME_ID/replays \
  -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -F "file=@replay.rep" \
  -F "title=My best game" \
  -F "comment=Won with rampage"
```

### Получить все реплеи игры

```bash
curl -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  http://localhost:8080/api/v1/games/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa/replays?limit=10
```

### Скачать реплей

```bash
curl -H "X-User-ID: 00000000-0000-0000-0000-000000000001" \
  -O -J http://localhost:8080/api/v1/replays/10000000-0000-0000-0000-000000000001/file
```
