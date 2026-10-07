#!/usr/bin/env sh
# =============================================
# Docker 环境下的数据库迁移执行器
# 供 docker-compose 的 migrate 一次性服务调用
# 使用 schema_migrations 表记录已执行文件，避免重复执行
# =============================================
set -e

DB_HOST="${DB_HOST:-db}"
DB_USER="${POSTGRES_USER:-foodlog}"
DB_PASSWORD="${POSTGRES_PASSWORD:-foodlog123}"
DB_NAME="${POSTGRES_DB:-foodlog}"
MIGRATIONS_DIR="${MIGRATIONS_DIR:-/migrations}"

export PGPASSWORD="$DB_PASSWORD"
PSQL="psql -h $DB_HOST -U $DB_USER -d $DB_NAME"

# 标记表：记录已执行的迁移文件
echo "▶️  初始化迁移标记表 schema_migrations ..."
$PSQL -v ON_ERROR_STOP=1 <<'SQL'
CREATE TABLE IF NOT EXISTS schema_migrations (
  filename   TEXT PRIMARY KEY,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
SQL

for f in "$MIGRATIONS_DIR"/*.sql; do
  base=$(basename "$f")
  if [ "$($PSQL -tAc "SELECT 1 FROM schema_migrations WHERE filename = '$base'")" = "1" ]; then
    echo "→ 跳过 $base（已执行）"
    continue
  fi
  echo "→ 执行 $base"
  $PSQL -v ON_ERROR_STOP=1 -f "$f"
  $PSQL -c "INSERT INTO schema_migrations (filename) VALUES ('$base')"
done

echo "✅ 迁移完成"
