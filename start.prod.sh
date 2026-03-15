#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROD_COMPOSE_FILE="$ROOT_DIR/docker-compose.prod.yml"
DEPLOY_ENV="$ROOT_DIR/config/prod/deploy.env"
BACKEND_ENV="$ROOT_DIR/config/prod/backend.env"
FRONTEND_ENV="$ROOT_DIR/config/prod/frontend.app.env"
ADMIN_ENV="$ROOT_DIR/config/prod/frontend.admin.env"

require_file() {
  local f="$1"
  if [ ! -f "$f" ]; then
    echo "Missing file: $f"
    exit 1
  fi
}

if ! command -v docker >/dev/null 2>&1; then
  echo "Missing docker"
  exit 1
fi

if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon is not running"
  exit 1
fi

if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
  COMPOSE_CMD="docker compose"
elif command -v docker-compose >/dev/null 2>&1; then
  COMPOSE_CMD="docker-compose"
else
  echo "Missing docker compose command"
  exit 1
fi

if [ ! -f "$DEPLOY_ENV" ]; then
  cp -f "$ROOT_DIR/config/prod/deploy.env.example" "$DEPLOY_ENV"
  echo "Created $DEPLOY_ENV"
fi
if [ ! -f "$BACKEND_ENV" ]; then
  cp -f "$ROOT_DIR/config/prod/backend.env.example" "$BACKEND_ENV"
  echo "Created $BACKEND_ENV"
fi
if [ ! -f "$FRONTEND_ENV" ]; then
  cp -f "$ROOT_DIR/config/prod/frontend.app.env.example" "$FRONTEND_ENV"
  echo "Created $FRONTEND_ENV"
fi
if [ ! -f "$ADMIN_ENV" ]; then
  cp -f "$ROOT_DIR/config/prod/frontend.admin.env.example" "$ADMIN_ENV"
  echo "Created $ADMIN_ENV"
fi

require_file "$DEPLOY_ENV"
require_file "$BACKEND_ENV"
require_file "$FRONTEND_ENV"
require_file "$ADMIN_ENV"

$COMPOSE_CMD -f "$PROD_COMPOSE_FILE" --env-file "$DEPLOY_ENV" up -d --build

echo "Production stack is running."
echo "Make sure DNS points to this server for app/admin/api domains."
