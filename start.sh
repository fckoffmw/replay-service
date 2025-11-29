#!/bin/bash

mkdir -p storage

sudo docker compose -f deploy/docker-compose/docker-compose.yml up --build
