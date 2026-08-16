#!/usr/bin/env bash
# =============================================
# 一键启动开发环境
# 启动 PostgreSQL → 后端 → 前端
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

echo "======================================"
echo " 🍽️  Food Log 开发环境启动"
echo "======================================"

# 1. 检查 .env
if [ ! -f .env ]; then
  echo "⚠️  未找到 .env，正在从 .env.example 创建..."
  cp .env.example .env
fi

# 2. 启动数据库
echo "▶️  启动 PostgreSQL..."
docker compose up -d db

# 3. 等待数据库就绪
echo "⏳ 等待数据库就绪..."
until docker compose exec db pg_isready -U foodlog -d foodlog >/dev/null 2>&1; do
  sleep 1
done
echo "✅ 数据库已就绪"

# 4. 运行迁移
echo "▶️  执行数据库迁移..."
bash scripts/migrate.sh

# 5. 提示用户启动后端和前端
echo ""
echo "======================================"
echo " ✅ 数据库已就绪"
echo ""
echo " 后端启动（终端 1）："
echo "   cd ${ROOT_DIR}/server && go run cmd/main.go"
echo ""
echo " 前端启动（终端 2）："
echo "   cd ${ROOT_DIR}/client && npm install && npm run dev"
echo "   访问 http://localhost:5173"
echo "======================================"
