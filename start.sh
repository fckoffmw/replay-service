#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

mkdir -p storage

# Освобождаем порт 3000 если он занят
echo "Checking port 3000..."
EXISTING_PID=$(lsof -ti:3000)
if [ ! -z "$EXISTING_PID" ]; then
    echo "Killing process on port 3000 (PID: $EXISTING_PID)"
    kill -9 $EXISTING_PID 2>/dev/null
    sleep 1
fi

# Запуск Python HTTP сервера для фронтенда в фоне
(cd client && python3 -m http.server 3000) &
FRONTEND_PID=$!

echo "Frontend server started on http://localhost:3000 (PID: $FRONTEND_PID)"

# Функция для остановки фронтенд сервера при завершении
cleanup() {
    echo "Stopping frontend server..."
    kill $FRONTEND_PID 2>/dev/null
    exit
}

trap cleanup SIGINT SIGTERM

sudo docker compose -f deploy/docker-compose/docker-compose.yml up --build
