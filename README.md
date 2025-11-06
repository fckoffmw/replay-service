# команда для запуска
docker compose -f deploy/docker-compose/docker-compose.yml up --build     
# запрос
http://localhost:8080/healthz
# остановка и удаление
docker compose -f deploy/docker-compose/docker-compose.yml down
