#!/usr/bin/env bash
# =============================================
# 构建前后端
# 1. Go 后端编译为 bin/server
# 2. Vue 前端构建到 client/dist
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"

echo "======================================"
echo " 🔨 构建 Food Log"
echo "======================================"

# 1. 后端
echo "▶️  构建 Go 后端..."
cd "$ROOT_DIR/server"
mkdir -p bin
go build -o bin/server ./cmd/main.go
echo "  ✅ server/bin/server"

# 2. 前端
echo "▶️  构建 Vue 前端..."
cd "$ROOT_DIR/client"
if [ ! -d node_modules ]; then
  echo "  📦 安装依赖..."
  npm install
fi
npm run build
echo "  ✅ client/dist"

echo ""
echo "✅ 构建完成！"
echo "  - 后端可执行文件: server/bin/server"
echo "  - 前端静态文件:   client/dist"
echo ""
echo "  Docker 部署：bash scripts/docker-build.sh"
