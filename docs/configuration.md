# Configuration Guide

## Environment Variables

Сервис настраивается через переменные окружения.

### Основные переменные

| Переменная | Описание | По умолчанию | Обязательная |
|------------|----------|--------------|--------------|
| `PORT` | Порт HTTP сервера | `8080` | Нет |
| `DB_DSN` | Строка подключения к PostgreSQL | - | **Да** |
| `STORAGE_DIR` | Директория для хранения файлов | `./storage` | Нет |
| `LOG_LEVEL` | Уровень логирования (debug/info/warn/error) | `debug` | Нет |
| `GIN_MODE` | Режим Gin (debug/release) | `debug` | Нет |

### Формат DB_DSN

```
postgres://user:password@host:port/database?sslmode=disable
```

Примеры:
```bash
# Локальная разработка
DB_DSN=postgres://replay:replay@localhost:5432/replay?sslmode=disable

# Production с SSL
DB_DSN=postgres://prod_user:secure_pass@db.example.com:5432/replay_prod?sslmode=require

# Docker Compose
DB_DSN=postgres://replay:replay@postgres:5432/replay?sslmode=disable
```

## Способы конфигурации

### 1. Файл .env (Development)

Создайте `.env` в корне проекта:

```env
PORT=8080
DB_DSN=postgres://replay:replay@localhost:5432/replay?sslmode=disable
STORAGE_DIR=./storage
LOG_LEVEL=debug
```

### 2. Переменные окружения (Production)

```bash
export DB_DSN="postgres://user:pass@prod-db:5432/replay?sslmode=require"
export STORAGE_DIR="/mnt/replays"
export LOG_LEVEL="info"
export GIN_MODE="release"

./replay-service
```

### 3. Docker Compose

В `docker-compose.yml`:

```yaml
services:
  api:
    environment:
      DB_DSN: postgres://replay:replay@postgres:5432/replay?sslmode=disable
      STORAGE_DIR: /app/storage
      LOG_LEVEL: info
      GIN_MODE: release
```

### 4. Systemd Service (Linux)

Создайте `/etc/systemd/system/replay-service.service`:

```ini
[Unit]
Description=Replay Service
After=network.target postgresql.service

[Service]
Type=simple
User=replay
WorkingDirectory=/opt/replay-service
Environment="DB_DSN=postgres://replay:pass@localhost:5432/replay?sslmode=require"
Environment="STORAGE_DIR=/var/lib/replay-service/storage"
Environment="LOG_LEVEL=info"
Environment="GIN_MODE=release"
ExecStart=/opt/replay-service/replay-service
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

## Приоритет конфигурации

1. **Переменные окружения** (высший приоритет)
2. **Docker Compose environment**
3. **Файл .env**
4. **Значения по умолчанию** в коде

## Production конфигурация

### Минимальная безопасная конфигурация:

```env
PORT=8080
DB_DSN=postgres://replay_user:STRONG_PASSWORD@db-host:5432/replay_prod?sslmode=require
STORAGE_DIR=/mnt/replays
LOG_LEVEL=info
GIN_MODE=release
```

### Рекомендации:

1. **Используйте SSL для БД** - `sslmode=require`
2. **Сильный пароль БД** - минимум 16 символов
3. **Отдельный пользователь БД** - не используйте superuser
4. **Логи на уровне info** - `LOG_LEVEL=info`
5. **Release режим** - `GIN_MODE=release`
6. **Отдельный volume для storage** - `/mnt/replays` или NFS

### Docker Compose для production:

```yaml
services:
  api:
    image: replay-service:latest
    environment:
      DB_DSN: postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=require
      STORAGE_DIR: /app/storage
      LOG_LEVEL: info
      GIN_MODE: release
    volumes:
      - /mnt/replays:/app/storage
    restart: unless-stopped
    
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      POSTGRES_DB: ${DB_NAME}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
```

Создайте `.env` с секретами:

```env
DB_USER=replay_prod
DB_PASSWORD=your_strong_password_here
DB_NAME=replay_production
```

## Проверка конфигурации

После запуска проверьте логи:

```bash
# Docker
docker logs replay_api

# Должно быть:
# Successfully connected to database
# [GIN-debug] Listening and serving HTTP on :8080
```

## PostgreSQL Configuration

### Где настраиваются credentials БД

В `deploy/docker-compose/docker-compose.yml`:

```yaml
postgres:
  environment:
    POSTGRES_USER: ${POSTGRES_USER:-replay}
    POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-replay}
    POSTGRES_DB: ${POSTGRES_DB:-replay}
```

**По умолчанию:**
- User: `replay`
- Password: `replay`
- Database: `replay`

### Изменение credentials

**Через .env файл:**

```env
POSTGRES_USER=my_user
POSTGRES_PASSWORD=strong_password
POSTGRES_DB=my_database
```

**Важно:** При изменении credentials нужно пересоздать БД:

```bash
docker compose -f deploy/docker-compose/docker-compose.yml down -v
docker compose -f deploy/docker-compose/docker-compose.yml up
```

Флаг `-v` удаляет volumes (все данные будут потеряны!).

### Проверка credentials

```bash
# Посмотреть переменные окружения
docker exec replay_postgres env | grep POSTGRES

# Подключиться к БД
docker exec -it replay_postgres psql -U replay -d replay

# Список пользователей
docker exec -it replay_postgres psql -U replay -d replay -c "\du"
```

## Troubleshooting

### Ошибка: "DB_DSN is required"

Убедитесь что переменная `DB_DSN` установлена:

```bash
echo $DB_DSN
```

### Ошибка: "Failed to connect to database"

Проверьте:
1. Доступность БД: `docker exec -it replay_postgres psql -U replay -d replay`
2. Правильность credentials в `DB_DSN`
3. Что контейнер postgres запущен: `docker ps | grep postgres`

### Ошибка: "Permission denied" при сохранении файлов

Проверьте права на `STORAGE_DIR`:

```bash
ls -la $STORAGE_DIR
chmod 755 $STORAGE_DIR
```

### Credentials не совпадают

Если `DB_DSN` использует одни credentials, а PostgreSQL настроен с другими:

```bash
# Проверить что в PostgreSQL
docker exec replay_postgres env | grep POSTGRES_USER

# Проверить что в приложении
docker exec replay_api env | grep DB_DSN

# Должны совпадать!
```
