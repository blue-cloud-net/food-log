# syntax=docker/dockerfile:1

# =============================================
# 阶段 1：构建前端（Vue + Vite → dist）
# =============================================
FROM node:24-alpine AS frontend

WORKDIR /fe

# 与 client/package.json 的 packageManager 保持一致
RUN npm i -g pnpm@11.24.0

COPY client/package.json client/pnpm-lock.yaml client/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY client/ ./
RUN pnpm build

# =============================================
# 阶段 2：构建后端（纯 Go 二进制，不含前端产物）
# =============================================
FROM golang:1.26-alpine AS backend

WORKDIR /app

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/server ./cmd/main.go

# =============================================
# 阶段 3：运行时（单进程同时提供 API / 图片 / 前端静态资源）
# =============================================
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget
ENV TZ=Asia/Shanghai

WORKDIR /app

# 数据目录：credentials / tmp / images（生产通过 bind mount 持久化）
ENV DATA_DIR=/app/data
# 前端静态资源目录：由前端构建阶段拷贝，服务启动时从磁盘读取
ENV STATIC_DIR=/app/dist
RUN mkdir -p "$DATA_DIR" "$STATIC_DIR"

COPY --from=backend /bin/server /usr/local/bin/server
# 前端构建产物（html / assets / icons / pwa 文件），运行时按需读取
COPY --from=frontend /fe/dist/ /app/dist/

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1

CMD ["server"]
