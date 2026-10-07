#!/usr/bin/env bash
# =============================================
# 停止开发环境的全部后台组件（server / client / docker db）
# 用法:
#   bash scripts/stop-dev.sh         # 正常停止，打印日志
#   bash scripts/stop-dev.sh --quiet # 静默模式（被 start-dev.sh --bg 调用）
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"
LOG_DIR="$ROOT_DIR/logs"

QUIET=false
[ "${1:-}" = "--quiet" ] && QUIET=true

log() { [ "$QUIET" = "false" ] && echo "$@"; return 0; }

stop_pid_file() {
  local name="$1" pid_file="$2"
  [ -f "$pid_file" ] || { log "   $name: 无 PID 文件"; return 0; }
  local pid
  pid=$(cat "$pid_file" 2>/dev/null || true)
  if [ -z "$pid" ] || ! kill -0 "$pid" 2>/dev/null; then
    log "   $name: PID $pid 已不在运行"
    rm -f "$pid_file"
    return 0
  fi
  log "▶️  停止 $name (PID $pid)..."
  kill "$pid" 2>/dev/null || true
  for _ in $(seq 1 10); do
    kill -0 "$pid" 2>/dev/null || break
    sleep 0.5
  done
  if kill -0 "$pid" 2>/dev/null; then
    log "   强制 kill -9 $pid"
    kill -9 "$pid" 2>/dev/null || true
  fi
  rm -f "$pid_file"
}

stop_pid_file "server" "$LOG_DIR/server.pid"
stop_pid_file "client" "$LOG_DIR/client.pid"

# 兜底：可能 PID 已失效，按进程名扫一遍
pkill -f "foodlog-server" 2>/dev/null || true
pkill -f "node.*vite" 2>/dev/null || true

# docker db
if docker compose ps db >/dev/null 2>&1 && docker compose ps db | grep -q Up; then
  log "▶️  停止 docker db..."
  docker compose stop db >/dev/null 2>&1 || true
fi

log "✅ 全部停止"
