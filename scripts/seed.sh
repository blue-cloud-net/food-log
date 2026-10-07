#!/usr/bin/env bash
# =============================================
# 导入测试种子数据
# 连接模式优先级：docker > remote(DATABASE_URL) > 本机 psql
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

if [ -f .env ]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
fi

POSTGRES_USER="${POSTGRES_USER:-foodlog}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-foodlog123}"
POSTGRES_DB="${POSTGRES_DB:-foodlog}"

SEED_FILE="$ROOT_DIR/server/migrations/002_seed.sql"
if [ ! -f "$SEED_FILE" ]; then
  echo "⚠️  未找到种子数据文件: $SEED_FILE"
  exit 0
fi

# 判断连接模式（与 migrate.sh 保持一致）
MODE="local"
if docker compose ps db >/dev/null 2>&1 && docker compose ps db | grep -q Up; then
  MODE="docker"
elif [ -n "${DATABASE_URL:-}" ]; then
  REMOTE_HOST=$(printf '%s' "$DATABASE_URL" | sed -n 's|.*@\([^:/?]*\).*|\1|p')
  case "$REMOTE_HOST" in
    db|localhost|127.0.0.1|"") : ;;
    *) MODE="remote" ;;
  esac
fi

case "$MODE" in
  docker)
    echo "▶️  通过 Docker 导入种子数据..."
    docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f /dev/stdin < "$SEED_FILE"
    ;;
  remote)
    echo "▶️  通过远程 psql 导入种子数据 (${DATABASE_URL})..."
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$SEED_FILE"
    ;;
  *)
    echo "▶️  通过本机 psql 导入种子数据..."
    export PGPASSWORD="$POSTGRES_PASSWORD"
    psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f "$SEED_FILE"
    ;;
esac

echo "✅ 种子数据导入完成 (mode: $MODE)"
echo "   默认账号：admin / admin"
