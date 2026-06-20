#!/bin/sh
set -e

if [ -d "./migrations" ] && [ "$(ls -A ./migrations 2>/dev/null)" ]; then
    echo "Running migrations..."
    migrate -path ./migrations -database "$DATABASE_URL" up
    echo "Migrations completed"
else
    echo "No migrations found, skipping..."
fi

echo "Starting air..."
exec air