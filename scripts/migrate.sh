#!/usr/bin/env bash
# =============================================
# 执行数据库迁移
# 依次执行 server/migrations/*.sql
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

# 读取配置（支持 .env）
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

POSTGRES_USER="${POSTGRES_USER:-foodlog}"
POSTGRES_PASSWORD="${POSTGRES_PASSWORD:-foodlog123}"
POSTGRES_DB="${POSTGRES_DB:-foodlog}"

# 判断连接方式：Docker 容器内 or 本机
if docker compose ps db >/dev/null 2>&1 && docker compose ps db | grep -q Up; then
  echo "▶️  通过 Docker 执行迁移..."
  for f in "$ROOT_DIR"/server/migrations/*.sql; do
    echo "  → $(basename "$f")"
    docker compose exec -T db psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f /dev/stdin < "$f"
  done
else
  echo "▶️  通过本机 psql 执行迁移..."
  export PGPASSWORD="$POSTGRES_PASSWORD"
  for f in "$ROOT_DIR"/server/migrations/*.sql; do
    echo "  → $(basename "$f")"
    psql -h localhost -U "$POSTGRES_USER" -d "$POSTGRES_DB" -v ON_ERROR_STOP=1 -f "$f"
  done
fi

echo "✅ 迁移完成"
