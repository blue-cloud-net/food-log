# ===== 构建阶段 =====
FROM node:20-alpine AS builder

WORKDIR /app

# 依赖缓存
COPY package.json package-lock.json ./
RUN npm ci

# 构建
COPY . .
RUN npm run build

# ===== 运行阶段：Nginx =====
FROM nginx:1.27-alpine

# 复制构建产物与 nginx 配置
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
