#!/usr/bin/env bash
# =============================================
# 停止开发环境容器
#
# 用法:
#   bash scripts/stop-dev.sh          正常停止，打印进度
#   bash scripts/stop-dev.sh --quiet  静默模式
#
# 说明：数据卷（foodlog-pgdata-dev / client_node_modules）会保留，
#       如需清空请手动执行 docker volume rm。
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

QUIET=false
[ "${1:-}" = "--quiet" ] && QUIET=true

log() { [ "$QUIET" = "false" ] && echo "$@"; return 0; }

log "▶️  停止开发环境容器..."
docker compose -f "$ROOT_DIR/docker-compose.dev.yml" down --remove-orphans >/dev/null 2>&1 || true
log "✅ 已停止（数据卷保留）"
