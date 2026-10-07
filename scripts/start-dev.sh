#!/usr/bin/env bash
# =============================================
# 一键启动开发环境（容器化）
#
# 容器组成（docker-compose.dev.yml）：
#   db      PostgreSQL，独立数据卷（不含生产数据）
#   server  Go + air：源码挂载 /app，改 *.go / *.sql 自动重编译
#   client  Vite + pnpm：源码挂载 /app，浏览器 HMR
#
# 后端以 APP_ENV=development 启动，会自动初始化空库并加载演示数据（admin / admin）。
#
# 用法:
#   bash scripts/start-dev.sh          后台构建并启动
#   bash scripts/start-dev.sh --logs   启动后持续跟踪日志（Ctrl+C 仅退出日志）
#   bash scripts/start-dev.sh --stop   停止全部容器
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"
COMPOSE=(docker compose -f docker-compose.dev.yml)

ACTION="up"
case "${1:-}" in
  ""|--bg|bg)   ACTION="up" ;;
  --logs|logs)  ACTION="logs" ;;
  --stop|stop)  exec bash "$ROOT_DIR/scripts/stop-dev.sh" ;;
  *)
    echo "用法: bash scripts/start-dev.sh [--logs|--stop]"
    exit 1
    ;;
esac

echo "======================================"
echo " 🍽️  Food Log 开发环境启动"
echo "======================================"

# 1. 准备 .env（容器通过 .env 注入数据库账号与 JWT 密钥）
if [ ! -f .env ]; then
  echo "⚠️  未找到 .env，正在从 .env.example 创建..."
  cp .env.example .env
fi

# 2. 读取端口配置
SERVER_PORT="${SERVER_PORT:-8080}"
CLIENT_PORT="${CLIENT_PORT:-5173}"

# 3. 构建并启动（首次会拉取基础镜像、编译 air，耗时较长）
echo "▶️  构建并启动容器（首次较慢）..."
"${COMPOSE[@]}" up -d --build

# 4. 等待后端健康（air 首次编译需要一点时间）
HEALTH_URL="http://localhost:${SERVER_PORT}/api/health"
echo "⏳ 等待后端就绪 (${HEALTH_URL})..."
HEALTHY=false
for _ in $(seq 1 60); do
  if curl -sf "$HEALTH_URL" >/dev/null 2>&1; then
    HEALTHY=true
    break
  fi
  sleep 2
done

echo ""
echo "======================================"
if [ "$HEALTHY" = "true" ]; then
  echo " ✅ 开发环境已就绪"
else
  echo " ⚠️  后端尚未健康，请查看日志排查"
fi
echo ""
echo " 前端: http://localhost:${CLIENT_PORT}   (Vite HMR)"
echo " 后端: http://localhost:${SERVER_PORT}   (air 热重载)"
echo " 默认账号: admin / admin（演示数据）"
echo ""
echo " 查看日志: docker compose -f docker-compose.dev.yml logs -f server client"
echo " 停止环境: bash scripts/start-dev.sh --stop"
echo "======================================"

if [ "$HEALTHY" != "true" ]; then
  echo ""
  echo "----- server 最近日志 -----"
  "${COMPOSE[@]}" logs --tail=30 server || true
  exit 1
fi

if [ "$ACTION" = "logs" ]; then
  echo ""
  echo "📜 跟踪 server / client 日志（Ctrl+C 仅退出日志，容器继续运行）..."
  "${COMPOSE[@]}" logs -f server client
fi
