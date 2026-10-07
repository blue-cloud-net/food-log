#!/usr/bin/env bash
# =============================================
# 执行数据库迁移
# 依次执行 server/migrations/*.sql
# 连接模式优先级：docker > remote(DATABASE_URL) > 本机 psql
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

# 读取配置（支持 .env）
if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

POSTGRES_USER="${POSTGRES_USER:-foodlog}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-foodlog123}"
POSTGRES_DB="${POSTGRES_DB:-foodlog}"

# 1. 优先使用 docker compose 中的 db 容器
MODE="local"
if docker compose ps db >/dev/null 2>&1 && docker compose ps db | grep -q Up; then
  MODE="docker"
fi

# 2. 若 docker 未起，但 DATABASE_URL 指向远程库，则走远程 psql
if [ "$MODE" = "local" ] && [ -n "${DATABASE_URL:-}" ]; then
  REMOTE_HOST=$(printf '%s' "$DATABASE_URL" | sed -n 's|.*@\([^:/?]*\).*|\1|p')
  case "$REMOTE_HOST" in
    db|localhost|127.0.0.1|"") : ;;
    *) MODE="remote" ;;
  esac
fi

run_sql() {
  local f="$1"
  case "$MODE" in
    docker)
      echo "  → $(basename "$f")  [docker]"
      docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f /dev/stdin < "$f"
      ;;
    remote)
      echo "  → $(basename "$f")  [remote: $DATABASE_URL]"
      psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$f"
      ;;
    *)
      echo "  → $(basename "$f")  [local psql]"
      export PGPASSWORD="$POSTGRES_PASSWORD"
      psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f "$f"
      ;;
  esac
}

echo "▶️  执行数据库迁移 (mode: $MODE)..."
for f in "$ROOT_DIR"/server/migrations/*.sql; do
  run_sql "$f"
done

echo "✅ 迁移完成 (mode: $MODE)"
