-- Скрипт для выдачи прав доступа к таблицам в существующей базе данных
-- Выполните этот скрипт от имени пользователя postgres (владельца БД)

-- Выдаем права на таблицы для всех пользователей (для разработки)
GRANT SELECT, INSERT, UPDATE, DELETE ON users TO PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON replays TO PUBLIC;

-- Альтернативно, если нужно выдать права конкретному пользователю:
-- GRANT SELECT, INSERT, UPDATE, DELETE ON users TO replay;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON replays TO replay;
