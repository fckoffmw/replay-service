# Документация Replay Service

## 📚 Навигация по документации

### 🏗️ Архитектура

| Документ | Описание | Для кого |
|----------|----------|----------|
| [Architecture](architecture.md) | Подробное описание MVC архитектуры с примерами кода | Разработчики |
| [MVC Quick Guide](mvc-guide.md) | Краткая шпаргалка: что где делать, типичные ошибки | Все разработчики |
| [Project Structure](project-structure.md) | Полная структура проекта, зависимости, команды | Новые участники |

### 🔌 API

| Документ | Описание | Для кого |
|----------|----------|----------|
| [API Specification](api-specification.md) | Полное описание REST API эндпоинтов | Frontend, QA |
| [API Examples](api-examples.http) | Готовые HTTP запросы для тестирования | Разработчики, QA |

### ⚙️ Конфигурация

| Документ | Описание | Для кого |
|----------|----------|----------|
| [Configuration](configuration.md) | Настройка переменных окружения | DevOps, Разработчики |
| [Storage Structure](storage-structure.md) | Структура файлового хранилища | Разработчики, DevOps |

---

## 🚀 Быстрый старт

### Для новых разработчиков

1. **Прочитайте сначала:**
   - [MVC Quick Guide](mvc-guide.md) - 5 минут
   - [Project Structure](project-structure.md) - 10 минут

2. **Затем изучите:**
   - [Architecture](architecture.md) - 20 минут
   - [API Specification](api-specification.md) - 15 минут

3. **Начните разработку:**
   - Следуйте правилам из [MVC Quick Guide](mvc-guide.md)
   - Используйте [API Examples](api-examples.http) для тестирования

### Для Frontend разработчиков

1. [API Specification](api-specification.md) - все эндпоинты
2. [API Examples](api-examples.http) - примеры запросов
3. [Storage Structure](storage-structure.md) - как хранятся файлы

### Для DevOps

1. [Configuration](configuration.md) - переменные окружения
2. [Project Structure](project-structure.md) - команды для сборки и запуска
3. [Storage Structure](storage-structure.md) - структура файлов

---

## 📖 Основные концепции

### MVC Архитектура

Проект следует паттерну **Model-View-Controller** с дополнительными слоями:

```
View (client/) → Controller (handlers/) → Service (services/)
                                              ↓
                                    Repository (repository/)
                                    Storage (storage/)
```

**Ключевые принципы:**
- Handlers - тонкий слой (только маршрутизация)
- Services - вся бизнес-логика
- Repository - только SQL запросы
- Storage - только файловые операции

Подробнее: [Architecture](architecture.md)

---

### Слои приложения

| Слой | Расположение | Ответственность |
|------|--------------|-----------------|
| **Model** | `models/` | Структуры данных |
| **View** | `client/` | Веб-интерфейс |
| **Controller** | `handlers/` | HTTP маршрутизация |
| **Service** | `services/` | Бизнес-логика |
| **Repository** | `repository/` | Доступ к БД |
| **Storage** | `storage/` | Файловая система |

Подробнее: [MVC Quick Guide](mvc-guide.md)

---

## 🎯 Типичные задачи

### Добавить новый эндпоинт

1. Добавить метод в `handlers/replay.go`
2. Добавить логику в соответствующий `services/*.go`
3. При необходимости добавить SQL в `repository/*.go`
4. Зарегистрировать роут в `main.go`

Пример: [MVC Quick Guide - Добавить новый эндпоинт](mvc-guide.md#добавить-новый-эндпоинт)

### Добавить новую сущность

1. Создать модель в `models/`
2. Создать repository в `repository/`
3. Создать service в `services/`
4. Создать handlers в `handlers/`
5. Зарегистрировать роуты в `main.go`

Пример: [MVC Quick Guide - Добавить новую сущность](mvc-guide.md#добавить-новую-сущность-например-user)

### Изменить бизнес-логику

Вся бизнес-логика находится в `services/`:
- `game_service.go` - логика работы с играми
- `replay_service.go` - логика работы с реплеями

**Не изменяйте логику в:**
- ❌ handlers (только маршрутизация)
- ❌ repository (только SQL)
- ❌ storage (только файлы)

---

## 🔍 Поиск информации

### Как работает загрузка реплея?
→ [Architecture - Поток данных](architecture.md#поток-данных)

### Какие есть эндпоинты?
→ [API Specification](api-specification.md)

### Как настроить переменные окружения?
→ [Configuration](configuration.md)

### Где хранятся файлы?
→ [Storage Structure](storage-structure.md)

### Как правильно организовать код?
→ [MVC Quick Guide](mvc-guide.md)

### Какая структура проекта?
→ [Project Structure](project-structure.md)

---

## 📊 Диаграммы

### Архитектура системы

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

Подробнее: [Architecture](architecture.md)

---

### Структура директорий

```
replay-service/
├── client/          # View - Frontend
├── server/
│   ├── cmd/         # Точка входа
│   ├── config/      # Конфигурация
│   └── internal/
│       ├── models/      # Model
│       ├── handlers/    # Controller
│       ├── services/    # Business Logic
│       ├── repository/  # Data Access
│       └── storage/     # File Storage
├── storage/         # Файлы реплеев
├── deploy/          # Docker
└── docs/            # Документация
```

Подробнее: [Project Structure](project-structure.md)

---

## 🛠️ Инструменты разработки

### Тестирование API
Используйте [API Examples](api-examples.http) с расширением REST Client в VS Code.

### Форматирование кода
```bash
go fmt ./...
```

### Линтер
```bash
golangci-lint run
```

### Сборка
```bash
go build -o bin/replay-service server/cmd/replay-service/main.go
```

Подробнее: [Project Structure - Полезные команды](project-structure.md#полезные-команды)

---

## 📝 Соглашения о коде

### Именование

- **Handlers**: `GetGames`, `CreateReplay`, `DeleteGame`
- **Services**: `GetUserGames`, `CreateReplay`, `DeleteGame`
- **Repository**: `GetByID`, `Create`, `Update`, `Delete`
- **Storage**: `SaveFile`, `DeleteFile`, `GetFilePath`

### Логирование

```go
log.Printf("[ServiceName] MethodName: details")
log.Printf("[ServiceName] MethodName SUCCESS: result")
log.Printf("[ServiceName] MethodName ERROR: %v", err)
```

### Обработка ошибок

```go
if err != nil {
    log.Printf("[Service] Method ERROR: %v", err)
    return fmt.Errorf("descriptive message: %w", err)
}
```

Подробнее: [Architecture](architecture.md)

---

## 🤝 Вклад в проект

### Перед началом работы

1. Изучите [MVC Quick Guide](mvc-guide.md)
2. Следуйте архитектурным принципам из [Architecture](architecture.md)
3. Проверяйте код на соответствие слоям

### Чеклист для Pull Request

- [ ] Код следует MVC архитектуре
- [ ] Handlers < 20 строк (логика в services)
- [ ] Services не работают с HTTP
- [ ] Repository содержит только SQL
- [ ] Добавлено логирование в services
- [ ] Обработаны все ошибки
- [ ] Обновлена документация (если нужно)

---

## 📞 Поддержка

### Вопросы по архитектуре
→ [Architecture](architecture.md) или [MVC Quick Guide](mvc-guide.md)

### Вопросы по API
→ [API Specification](api-specification.md)

### Вопросы по настройке
→ [Configuration](configuration.md)

### Вопросы по структуре
→ [Project Structure](project-structure.md)

---

## 🔄 История изменений

См. [CHANGELOG.md](../CHANGELOG.md) в корне проекта.

---

## 📄 Лицензия

MIT
