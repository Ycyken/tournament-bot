#!/bin/bash
set -e

docker-compose down --remove-orphans -v
docker-compose up --build -d
