#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INFRA_DIR="$ROOT_DIR/infrastructure"
FULL_COMPOSE_FILE="$ROOT_DIR/docker-compose.full.yml"

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1"
    exit 1
  fi
}

pick_compose_cmd() {
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    echo "docker compose"
    return
  fi
  if command -v docker-compose >/dev/null 2>&1; then
    echo "docker-compose"
    return
  fi
  echo ""
}

wait_for_http() {
  local url="$1"
  local name="$2"
  local expected_code="${3:-200}"
  local retries=60
  local sleep_sec=2

  for _ in $(seq 1 "$retries"); do
    local code
    code="$(curl -s -o /dev/null -w "%{http_code}" "$url" || true)"
    if [ "$code" = "$expected_code" ]; then
      echo "$name is ready"
      return
    fi
    sleep "$sleep_sec"
  done

  echo "$name is not ready after waiting"
  exit 1
}

main() {
  require_cmd docker
  require_cmd curl

  if ! docker info >/dev/null 2>&1; then
    echo "Docker daemon is not running. Please start Docker first."
    exit 1
  fi

  local compose_cmd
  compose_cmd="$(pick_compose_cmd)"
  if [ -z "$compose_cmd" ]; then
    echo "Missing docker compose command (docker compose or docker-compose)."
    exit 1
  fi

  if [ ! -f "$INFRA_DIR/.env" ] && [ -f "$INFRA_DIR/.env.example" ]; then
    cp -f "$INFRA_DIR/.env.example" "$INFRA_DIR/.env"
    echo "Created infrastructure/.env from .env.example"
  fi

  echo "Starting all services via Docker Compose"
  $compose_cmd -f "$FULL_COMPOSE_FILE" --env-file "$INFRA_DIR/.env" up -d --build

  wait_for_http "http://localhost:3000/api/data-mode" "Backend API"
  wait_for_http "http://localhost:5173" "Portal frontend"
  wait_for_http "http://localhost:5174" "Admin frontend"

  echo ""
  echo "All services are running."
  echo "Backend:        http://localhost:3000"
  echo "Portal:         http://localhost:5173"
  echo "Admin:          http://localhost:5174"
  echo "Kibana:         http://localhost:5601"
}

main "$@"
