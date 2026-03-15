#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INFRA_DIR="$ROOT_DIR/infrastructure"
FULL_COMPOSE_FILE="$ROOT_DIR/docker-compose.full.yml"
BACKUP_DIR="$ROOT_DIR/infrastructure/backup"
PG_DUMP_FILE="$BACKUP_DIR/review_db.dump"
ES_JSON_FILE="$BACKUP_DIR/es_companies.json"
ES_BULK_FILE="$BACKUP_DIR/es_companies.bulk.ndjson"

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
  local retries="${WAIT_RETRIES:-180}"
  local sleep_sec=2

  for _ in $(seq 1 "$retries"); do
    local code
    code="$(curl -s -o /dev/null -w "%{http_code}" "$url" || true)"
    if [ "$expected_code" = "any" ] && [ "$code" != "000" ]; then
      echo "$name is ready (HTTP $code)"
      return
    fi
    if [ "$code" = "$expected_code" ]; then
      echo "$name is ready"
      return
    fi
    sleep "$sleep_sec"
  done

  echo "$name is not ready after waiting"
  exit 1
}

ensure_es_bulk_file() {
  if [ -f "$ES_BULK_FILE" ]; then
    return
  fi

  if [ ! -f "$ES_JSON_FILE" ]; then
    return
  fi

  if command -v jq >/dev/null 2>&1; then
    jq -c '.hits.hits[] | {"index":{"_index":"companies","_id":._id}}, ._source' "$ES_JSON_FILE" > "$ES_BULK_FILE"
  else
    python3 - "$ES_JSON_FILE" "$ES_BULK_FILE" <<'PY'
import json
import sys

src = sys.argv[1]
dst = sys.argv[2]

with open(src, "r", encoding="utf-8") as f:
    data = json.load(f)

hits = data.get("hits", {}).get("hits", [])
with open(dst, "w", encoding="utf-8") as out:
    for hit in hits:
        idx = {"index": {"_index": "companies", "_id": hit.get("_id")}}
        out.write(json.dumps(idx, ensure_ascii=False) + "\n")
        out.write(json.dumps(hit.get("_source", {}), ensure_ascii=False) + "\n")
PY
  fi
}

restore_postgres_if_dump_exists() {
  if [ ! -f "$PG_DUMP_FILE" ]; then
    echo "Postgres dump not found, skipping DB restore"
    return
  fi

  echo "Restoring PostgreSQL data from $PG_DUMP_FILE"
  docker cp "$PG_DUMP_FILE" review_postgres:/tmp/review_db.dump
  docker exec review_postgres sh -lc "pg_restore -U postgres -d review_db --clean --if-exists /tmp/review_db.dump"
  docker exec review_postgres sh -lc "rm -f /tmp/review_db.dump"
}

restore_es_if_backup_exists() {
  ensure_es_bulk_file

  if [ ! -f "$ES_BULK_FILE" ]; then
    echo "Elasticsearch backup not found, skipping ES restore"
    return
  fi

  echo "Restoring Elasticsearch companies index from $ES_BULK_FILE"
  docker cp "$ES_BULK_FILE" review_es:/tmp/es_companies.bulk.ndjson
  docker exec review_es sh -lc "curl -s -X DELETE 'http://localhost:9200/companies' >/dev/null || true"
  docker exec review_es sh -lc "curl -s -X PUT 'http://localhost:9200/companies' -H 'Content-Type: application/json' -d '{}' >/dev/null"
  docker exec review_es sh -lc "curl -s -H 'Content-Type: application/x-ndjson' -X POST 'http://localhost:9200/_bulk?refresh=true' --data-binary '@/tmp/es_companies.bulk.ndjson' >/dev/null"
  docker exec review_es sh -lc "true" >/dev/null 2>&1 || true
}

ensure_default_admin_exists() {
  local count
  count="$(docker exec review_postgres psql -U postgres -d review_db -tAc "select count(*) from admin_users;")"
  count="$(echo "$count" | tr -d '[:space:]')"
  if [ "$count" != "0" ]; then
    echo "Admin user(s) already exist, skipping bootstrap admin"
    return
  fi

  echo "No admin user found, creating default admin account"
  local transport_hash
  transport_hash="$(printf "admin123" | sha256sum | awk '{print $1}')"
  local bcrypt_hash
  bcrypt_hash="$(python3 - "$transport_hash" <<'PY'
import bcrypt
import sys
print(bcrypt.hashpw(sys.argv[1].encode(), bcrypt.gensalt(rounds=10)).decode())
PY
)"

  docker exec review_postgres psql -U postgres -d review_db -c "INSERT INTO admin_users (id,email,password,name,role,created_at,updated_at) VALUES (gen_random_uuid(),'admin@reviewct.local','$bcrypt_hash','Super Admin','ADMIN',now(),now()) ON CONFLICT (email) DO NOTHING;" >/dev/null
  echo "Default admin created: admin@reviewct.local / admin123"
}

main() {
  require_cmd docker
  require_cmd curl
  require_cmd python3

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

  if [ "${RESTORE_DATA:-1}" = "1" ]; then
    restore_postgres_if_dump_exists
    restore_es_if_backup_exists
    docker compose -f "$FULL_COMPOSE_FILE" --env-file "$INFRA_DIR/.env" restart backend >/dev/null 2>&1 || true
  else
    echo "RESTORE_DATA=0, skipping data restore"
  fi

  ensure_default_admin_exists

  wait_for_http "http://localhost:3000/api/data-mode" "Backend API"
  wait_for_http "http://localhost:5173" "Portal frontend" "any"
  wait_for_http "http://localhost:5174" "Admin frontend" "any"

  echo ""
  echo "All services are running."
  echo "Backend:        http://localhost:3000"
  echo "Portal:         http://localhost:5173"
  echo "Admin:          http://localhost:5174"
  echo "Kibana:         http://localhost:5601"
  echo "Admin default:  admin@reviewct.local / admin123"
}

main "$@"
