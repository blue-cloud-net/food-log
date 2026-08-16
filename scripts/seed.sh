#!/usr/bin/env bash
# =============================================
# 导入测试种子数据
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

if [ -f .env ]; then
  set -a
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

if docker compose ps db >/dev/null 2>&1 && docker compose ps db | grep -q Up; then
  echo "▶️  通过 Docker 导入种子数据..."
  docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f /dev/stdin < "$SEED_FILE"
else
  echo "▶️  通过本机 psql 导入种子数据..."
  export PGPASSWORD="$POSTGRES_PASSWORD"
  psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f "$SEED_FILE"
fi

echo "✅ 种子数据导入完成"
