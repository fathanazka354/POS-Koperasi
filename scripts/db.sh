#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if [[ -f "$ROOT_DIR/.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT_DIR/.env"
  set +a
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-pos_koperasi}"

PSQL=(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME")

run_migrate() {
  echo "Running schema migration..."
  "${PSQL[@]}" -f "$ROOT_DIR/migrations/001_schema.sql"
  echo "Running chat migration..."
  "${PSQL[@]}" -f "$ROOT_DIR/migrations/002_chat.sql"
  echo "Running shop migration..."
  "${PSQL[@]}" -f "$ROOT_DIR/migrations/003_shop.sql"
}

run_seed() {
  echo "Running Go seed (internal/seed per entitas)..."
  (cd "$ROOT_DIR" && go run ./cmd/seed)
}

run_fresh() {
  echo "Recreating public schema..."
  "${PSQL[@]}" -c "DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;"
  run_migrate
  run_seed
}

run_reset() {
  echo "Resetting with legacy init script..."
  "${PSQL[@]}" -f "$ROOT_DIR/migrations/001_init.sql"
}

usage() {
  cat <<'EOF'
Usage: scripts/db.sh <command>

Commands:
  migrate   Run schema migration only
  seed      Run seed data only
  fresh     Drop schema, migrate, then seed
  reset     Run legacy migrations/001_init.sql
EOF
}

cmd="${1:-}"
case "$cmd" in
  migrate) run_migrate ;;
  seed) run_seed ;;
  fresh) run_fresh ;;
  reset) run_reset ;;
  *) usage; exit 1 ;;
esac
