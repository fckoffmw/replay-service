# Клиентская часть Replay Service

## Структура файлов

```
client/
├── index.html          # Главная страница (список игр и реплеев)
├── player.html         # Страница проигрывателя
├── login.html          # Страница входа
├── register.html       # Страница регистрации
├── style.css           # Стили для главной страницы
├── auth.css            # Стили для страниц авторизации
├── script.js           # Логика главной страницы
└── auth.js             # Логика авторизации
```

## Авторизация

### JWT Токены

Приложение использует JWT (JSON Web Tokens) для аутентификации:

- Токен сохраняется в `localStorage` после успешного входа/регистрации
- Токен автоматически добавляется в заголовок `Authorization: Bearer <token>` для всех API запросов
- При истечении срока действия токена пользователь перенаправляется на страницу входа

### Структура JWT токена

```json
{
  "user_id": "uuid",
  "email": "user@example.com",
  "username": "username",
  "exp": 1234567890
}
```

### Страницы

#### login.html - Вход
- Email и пароль
- Чекбокс "Запомнить меня"
- Ссылка на регистрацию

#### register.html - Регистрация
- Имя пользователя (минимум 3 символа)
- Email
- Пароль (минимум 6 символов)
- Подтверждение пароля
- Ссылка на вход

## API Endpoints для авторизации

### Регистрация
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "user",
  "email": "user@example.com",
  "password": "password123"
}

Response 201:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "user",
    "email": "user@example.com"
  }
}
```

### Вход
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123",
  "remember_me": true
}

Response 200:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "uuid",
    "username": "user",
    "email": "user@example.com"
  }
}
```

## TokenManager

Утилита для работы с JWT токенами:

```javascript
// Сохранить токен
TokenManager.setToken(token);

// Получить токен
const token = TokenManager.getToken();

// Удалить токен
TokenManager.removeToken();

// Проверить авторизацию
if (TokenManager.isAuthenticated()) {
  // Пользователь авторизован
}

// Получить данные пользователя из токена
const user = TokenManager.getUserFromToken();
// { id: "uuid", email: "...", username: "..." }
```

## Защита страниц

Все страницы, кроме `login.html` и `register.html`, проверяют авторизацию:

```javascript
// В script.js
if (!TokenManager.isAuthenticated()) {
    window.location.href = 'login.html';
}
```

## Использование в запросах

```javascript
// Получить заголовки с токеном
function getAuthHeaders() {
    const token = TokenManager.getToken();
    return {
        'Authorization': `Bearer ${token}`
    };
}

// Использование
const response = await fetch(`${API_BASE}/games`, {
    headers: getAuthHeaders()
});
```

## Выход из системы

```javascript
function logout() {
    TokenManager.removeToken();
    window.location.href = 'login.html';
}
```

## Обработка ошибок

При ошибке авторизации (401) пользователь должен быть перенаправлен на страницу входа:

```javascript
if (response.status === 401) {
    TokenManager.removeToken();
    window.location.href = 'login.html';
}
```

## Стилизация

Страницы авторизации используют отдельный файл стилей `auth.css` с:
- Градиентным фоном
- Анимацией появления
- Адаптивным дизайном
- Состояниями загрузки
- Сообщениями об ошибках

## Безопасность

- Пароли никогда не сохраняются на клиенте
- JWT токен хранится в localStorage (для production рекомендуется httpOnly cookie)
- Токен автоматически проверяется на истечение срока действия
- При выходе токен удаляется из localStorage

## Будущие улучшения

- [ ] Refresh токены для автоматического обновления
- [ ] Хранение токена в httpOnly cookie
- [ ] Двухфакторная аутентификация
- [ ] Восстановление пароля
- [ ] OAuth провайдеры (Google, GitHub)
