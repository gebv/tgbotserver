#!/bin/bash

set -e

docker exec database psql -U postgres -c "CREATE DATABASE app;"
docker exec database psql -U postgres -c "CREATE USER app WITH password 'apppassword';"
docker exec database psql -U postgres -c "GRANT ALL privileges ON DATABASE app TO app;"
docker exec database psql -U postgres -c "\list"