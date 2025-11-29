#!/bin/bash

mkdir -p storage
chmod 777 storage

export UID=$(id -u)
export GID=$(id -g)

docker compose -f deploy/docker-compose/docker-compose.yml up --build
