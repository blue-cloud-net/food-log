#!/usr/bin/env bash
# =============================================
# Docker 全容器部署
# 构建并启动 db + server + client
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

echo "======================================"
echo " 🐳 Docker 部署 Food Log"
echo "======================================"

# 检查 .env
if [ ! -f .env ]; then
  echo "⚠️  未找到 .env，正在创建..."
  cp .env.example .env
  echo "   ⚠️  请修改 .env 中的 JWT_SECRET 后重新运行！"
fi

echo "▶️  构建并启动全部服务..."
docker compose up -d --build

echo ""
echo "✅ 部署完成！"
echo "  前端:  http://localhost${CLIENT_PORT:-:80}"
echo "  后端:  http://localhost:${SERVER_PORT:-8080}"
echo ""
echo "  查看日志:  docker compose logs -f"
echo "  停止服务:  docker compose down"
