#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INFRA_DIR="$ROOT_DIR/infrastructure"
BACKEND_DIR="$ROOT_DIR/backend"
FRONTEND_DIR="$ROOT_DIR/frontend"
ADMIN_DIR="$ROOT_DIR/admin-frontend"

BACKEND_LOG="/tmp/review-company-backend.log"
FRONTEND_LOG="/tmp/review-company-frontend.log"
ADMIN_LOG="/tmp/review-company-admin.log"

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

is_port_listening() {
  local port="$1"
  lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
}

start_if_not_running() {
  local port="$1"
  local workdir="$2"
  local cmd="$3"
  local log_file="$4"
  local service_name="$5"

  if is_port_listening "$port"; then
    echo "$service_name is already running on port $port"
    return
  fi

  echo "Starting $service_name on port $port"
  nohup bash -lc "$cmd" > "$log_file" 2>&1 &
}

wait_for_http() {
  local url="$1"
  local name="$2"
  local retries=40
  local sleep_sec=2

  for _ in $(seq 1 "$retries"); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      echo "$name is ready"
      return
    fi
    sleep "$sleep_sec"
  done

  echo "$name is not ready after waiting"
  exit 1
}

wait_for_postgres() {
  local retries=40
  local sleep_sec=2
  for _ in $(seq 1 "$retries"); do
    if docker exec review_postgres pg_isready -U postgres -d review_db >/dev/null 2>&1; then
      echo "PostgreSQL is ready"
      return
    fi
    sleep "$sleep_sec"
  done

  echo "PostgreSQL is not ready after waiting"
  exit 1
}

install_node_deps() {
  local dir="$1"
  if [ -f "$dir/package-lock.json" ]; then
    npm --prefix "$dir" ci
  else
    npm --prefix "$dir" install
  fi
}

main() {
  require_cmd docker
  require_cmd go
  require_cmd npm
  require_cmd curl
  require_cmd lsof

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

  echo "Starting infrastructure containers"
  $compose_cmd -f "$INFRA_DIR/docker-compose.yml" --env-file "$INFRA_DIR/.env" up -d

  wait_for_postgres
  wait_for_http "http://localhost:9200" "Elasticsearch"

  echo "Installing frontend dependencies"
  install_node_deps "$FRONTEND_DIR"
  install_node_deps "$ADMIN_DIR"

  start_if_not_running 3000 "$BACKEND_DIR" "cd '$BACKEND_DIR' && go run ./cmd/api/main.go" "$BACKEND_LOG" "Backend API"
  start_if_not_running 5173 "$FRONTEND_DIR" "cd '$FRONTEND_DIR' && npm run dev -- --host 0.0.0.0 --port 5173" "$FRONTEND_LOG" "Portal frontend"
  start_if_not_running 5174 "$ADMIN_DIR" "cd '$ADMIN_DIR' && npm run dev -- --host 0.0.0.0 --port 5174" "$ADMIN_LOG" "Admin frontend"

  sleep 2

  echo ""
  echo "All services are starting."
  echo "Backend:        http://localhost:3000"
  echo "Portal:         http://localhost:5173"
  echo "Admin:          http://localhost:5174"
  echo ""
  echo "Logs:"
  echo "  $BACKEND_LOG"
  echo "  $FRONTEND_LOG"
  echo "  $ADMIN_LOG"
}

main "$@"
