# Replay Service Documentation

Сервис для хранения и управления игровыми реплеями.

## Быстрый старт

### Запуск проекта

```bash
./start.sh
```

Сервис будет доступен на `http://localhost:8080`

### Остановка проекта

```bash
sudo docker compose -f deploy/docker-compose/docker-compose.yml down
```

### Остановка с удалением данных

```bash
sudo docker compose -f deploy/docker-compose/docker-compose.yml down -v
```

## Тестирование

Откройте `client/index.html` в браузере для интерактивного тестирования API.

## Документация

- [API Specification](api-specification.md) - описание всех эндпоинтов
- [Storage Structure](storage-structure.md) - структура файлового хранилища
- [API Examples](api-examples.http) - примеры запросов

## Структура проекта

```
replay-service/
├── server/
│   ├── cmd/replay-service/     # Точка входа
│   ├── config/                 # Конфигурация
│   ├── internal/
│   │   ├── database/          # Подключение к БД
│   │   ├── handlers/          # HTTP handlers
│   │   ├── middleware/        # Middleware (auth)
│   │   ├── models/            # Модели данных
│   │   └── repository/        # Работа с БД
│   └── migrations/            # SQL миграции
├── storage/                   # Файловое хранилище
├── deploy/
│   ├── docker/               # Dockerfile
│   └── docker-compose/       # Docker Compose
├── docs/                     # Документация
├── test-client.html         # HTML клиент для тестирования
└── start.sh                 # Скрипт запуска
```
