#!/bin/sh
set -e

echo "Waiting for database to be ready..."
sleep 5

echo "Starting API (migrations will be applied automatically)..."
exec go run ./cmd/api
