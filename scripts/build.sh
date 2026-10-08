#!/usr/bin/env bash
# =============================================
# 本地一体化构建（不使用 Docker）
#
#   client → pnpm build，产物同步到仓库根 dist/（运行时由 STATIC_DIR 读取）
#   server → 编译为 server/bin/server（纯后端，不含前端产物）
#
# 生产镜像构建请使用：docker compose build
# =============================================
set -euo pipefail
cd "$(dirname "$0")/.."
ROOT_DIR="$(pwd)"
DIST_DIR="$ROOT_DIR/dist"

echo "======================================"
echo " 🔨 构建 Food Log（一体化二进制）"
echo "======================================"

# 1. 前端
echo "▶️  构建 Vue 前端..."
cd "$ROOT_DIR/client"
if [ ! -d node_modules ]; then
  echo "  📦 安装依赖..."
  pnpm install
fi
pnpm build
echo "  ✅ client/dist"

# 2. 同步到仓库根 dist/（保留目录供 STATIC_DIR 默认值 ./dist 使用）
echo "▶️  同步前端产物到 dist/ ..."
rm -rf "$DIST_DIR"
cp -r "$ROOT_DIR/client/dist" "$DIST_DIR"

# 3. 后端
echo "▶️  编译 Go 后端..."
cd "$ROOT_DIR/server"
mkdir -p bin
go build -o bin/server ./cmd/main.go
echo "  ✅ server/bin/server"

echo ""
echo "✅ 构建完成！"
echo "  运行方式：DATA_DIR=./data STATIC_DIR=./dist DATABASE_URL=... ./server/bin/server"
echo "  Docker 部署：docker compose up -d --build"
