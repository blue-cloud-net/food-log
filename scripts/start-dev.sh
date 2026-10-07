#!/usr/bin/env bash
# =============================================
# 一键启动开发环境
# - 本地模式：启动本地 docker PostgreSQL
# - 远程模式：DATABASE_URL 指向远程库时直接跳过 docker
#
# 用法:
#   bash scripts/start-dev.sh            # 只起 db，提示手动起后端/前端
#   bash scripts/start-dev.sh --bg       # 后台启动 db+后端+前端，日志写 logs/
#   bash scripts/start-dev.sh --logs     # 同 --bg，再 tail 所有日志
#   bash scripts/start-dev.sh --stop     # 停所有后台组件
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"
LOG_DIR="$ROOT_DIR/logs"
mkdir -p "$LOG_DIR"

ACTION="${1:-}"

# --stop 转交
case "$ACTION" in
  --stop|stop)
    exec bash "$ROOT_DIR/scripts/stop-dev.sh"
    ;;
esac

echo "======================================"
echo " 🍽️  Food Log 开发环境启动"
echo "======================================"

# 1. 检查 .env（优先 .env， fallback .env.development）
if [ -f .env ]; then
  set -a; # shellcheck disable=SC1091
  source .env; set +a
elif [ -f .env.development ]; then
  echo "⚠️  未找到 .env，使用 .env.development"
  set -a; # shellcheck disable=SC1091
  source .env.development; set +a
else
  echo "⚠️  未找到 .env，正在从 .env.example 创建..."
  cp .env.example .env
  set -a; # shellcheck disable=SC1091
  source .env; set +a
fi

# 2. 判断数据库模式
DB_MODE="local"
DB_HOST_DISPLAY="(docker)"
if [ -n "${DATABASE_URL:-}" ]; then
  REMOTE_HOST=$(printf '%s' "$DATABASE_URL" | sed -n 's|.*@\([^:/?]*\).*|\1|p')
  case "$REMOTE_HOST" in
    db|localhost|127.0.0.1|"") : ;;
    *) DB_MODE="remote"; DB_HOST_DISPLAY="$REMOTE_HOST" ;;
  esac
fi

# 3. 准备数据库
if [ "$DB_MODE" = "remote" ]; then
  echo "▶️  检测到远程 DATABASE_URL → host=${DB_HOST_DISPLAY}"
  echo "   跳过本地 docker 数据库；请确保远程库可访问"
  if ! command -v psql >/dev/null 2>&1; then
    echo "❌ 远程模式需要本机 psql 客户端（postgresql-client），请先安装"
    exit 1
  fi
else
  echo "▶️  启动本地 PostgreSQL (docker)..."
  docker compose up -d db
  echo "⏳ 等待数据库就绪..."
  until docker compose exec db pg_isready -U "${POSTGRES_USER:-foodlog}" -d "${POSTGRES_DB:-foodlog}" >/dev/null 2>&1; do
    sleep 1
  done
  echo "✅ 数据库已就绪"
fi

# 4. 运行迁移
echo "▶️  执行数据库迁移..."
bash scripts/migrate.sh

# 5. 默认模式：只起 db，提示手动起后端/前端
if [ "$ACTION" != "--bg" ] && [ "$ACTION" != "--logs" ]; then
  echo ""
  echo "======================================"
  echo " ✅ 数据库就绪 (mode: ${DB_MODE}, host: ${DB_HOST_DISPLAY})"
  echo ""
  echo " 一键后台启动全部组件（推荐）："
  echo "   bash scripts/start-dev.sh --bg"
  echo ""
  echo " 后端启动（终端 1）："
  echo "   cd ${ROOT_DIR}/server && go run cmd/main.go"
  echo ""
  echo " 前端启动（终端 2）："
  echo "   cd ${ROOT_DIR}/client && npm install && npm run dev"
  echo "   访问 http://localhost:5173"
  echo ""
  echo " 默认账号（仅 seed 导入后可用）：admin / admin"
  echo "======================================"
  exit 0
fi

# =====================================================
# 6. --bg / --logs 模式：后台启动后端 + 前端，日志到 logs/
# =====================================================

# 先停旧实例（避免端口冲突）
if [ -f "$LOG_DIR/server.pid" ] || [ -f "$LOG_DIR/client.pid" ]; then
  echo "▶️  检测到旧实例，先停..."
  bash "$ROOT_DIR/scripts/stop-dev.sh" --quiet || true
fi

SERVER_PORT="${SERVER_PORT:-8080}"

# ---- 编译后端（从项目根目录运行，保证 UPLOAD_DIR=./server/uploads 路径正确）----
echo "▶️  编译后端..."
if ! (cd "$ROOT_DIR/server" && go build -o "$LOG_DIR/foodlog-server" ./cmd/main.go) 2>"$LOG_DIR/build.err"; then
  echo "❌ 编译失败，查看 $LOG_DIR/build.err:"
  cat "$LOG_DIR/build.err"
  exit 1
fi

# ---- 启动后端 ----
echo "▶️  启动后端 (后台)..."
: > "$LOG_DIR/server.log"
cd "$ROOT_DIR"
nohup "$LOG_DIR/foodlog-server" >> "$LOG_DIR/server.log" 2>&1 &
SERVER_PID=$!
echo "$SERVER_PID" > "$LOG_DIR/server.pid"
disown "$SERVER_PID" 2>/dev/null || true

# ---- 启动前端 ----
echo "▶️  启动前端 (后台)..."
if [ ! -d "$ROOT_DIR/client/node_modules" ]; then
  echo "   安装前端依赖（首次会较慢）..."
  (cd "$ROOT_DIR/client" && npm install) >> "$LOG_DIR/client.log" 2>&1
fi
: > "$LOG_DIR/client.log"
(cd "$ROOT_DIR/client" && nohup npm run dev >> "$LOG_DIR/client.log" 2>&1 &)
CLIENT_PID=$(pgrep -f "vite$" | head -n1 || echo "")
[ -n "$CLIENT_PID" ] && echo "$CLIENT_PID" > "$LOG_DIR/client.pid"

# ---- 健康检查 ----
echo "⏳ 等待后端健康..."
HEALTH_URL="http://localhost:${SERVER_PORT}/api/health"
HEALTHY=false
for i in $(seq 1 30); do
  if curl -sf "$HEALTH_URL" >/dev/null 2>&1; then
    HEALTHY=true
    break
  fi
  sleep 1
done

# ---- 汇总 ----
echo ""
echo "======================================"
if [ "$HEALTHY" = "true" ]; then
  echo " ✅ 全部就绪"
else
  echo " ⚠️  后端 30s 内未健康，请检查 $LOG_DIR/server.log"
fi
echo ""
echo " 后端: http://localhost:${SERVER_PORT}   (PID $SERVER_PID)"
echo " 前端: http://localhost:5173   (PID ${CLIENT_PID:-?})"
echo ""
echo " 日志文件："
echo "   tail -f $LOG_DIR/server.log"
echo "   tail -f $LOG_DIR/client.log"
echo "   tail -f $LOG_DIR/*.log    # 全部"
echo ""
echo " 停止全部：bash scripts/start-dev.sh --stop"
echo "======================================"

# 打印各组件启动末尾（"启动信号"）
echo ""
echo "----- 🍽️  server.log (last 8) -----"
tail -n 8 "$LOG_DIR/server.log" 2>/dev/null || echo "(空)"
echo "----- 🎨 client.log (last 8) -----"
tail -n 8 "$LOG_DIR/client.log" 2>/dev/null || echo "(空)"

# --logs: 持续跟踪日志（Ctrl+C 只退出 tail，不影响后台服务）
if [ "$ACTION" = "--logs" ]; then
  echo ""
  echo "📜 跟踪所有日志（Ctrl+C 仅退出 tail，服务继续运行）..."
  tail -F "$LOG_DIR/server.log" "$LOG_DIR/client.log"
fi
