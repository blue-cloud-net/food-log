# ===== 构建阶段 =====
FROM golang:1.26-alpine AS builder

WORKDIR /app

# 依赖缓存
COPY go.mod go.sum* ./
RUN go mod download

# 编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/main.go

# ===== 运行阶段 =====
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /bin/server /usr/local/bin/server
COPY migrations ./migrations

# 上传目录
RUN mkdir -p /app/uploads

EXPOSE 8080

CMD ["server"]
