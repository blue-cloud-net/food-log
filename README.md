# 🍽️ 美食日志 (Food Log)

记录自己做过的菜、去过的餐厅、吃过的菜品，支持评分与照片。一套代码同时支持 **Web / 手机 Web / Android PWA**。

## ✨ 功能

- 📝 **自制菜谱** — 记录菜名、食材清单、烹饪步骤、难度、耗时、评分、标签、多张照片
- 🍽️ **餐厅探店** — 记录店名、地址、菜系、环境照片、综合评分
- 🥘 **菜品点评** — 每道菜记录价格、口味描述、评分、就餐日期、照片
- 🖼️ **图片智能处理** — 大图自动压缩，生成缩略图，加载更快
- 🔐 **用户系统** — 注册登录、JWT 认证、个人中心
- 📱 **PWA** — 手机浏览器可安装到桌面，离线可访问

## 🏗️ 技术栈

| 层 | 技术 |
|---|---|
| 前端 | Vue 3 + Vite + TypeScript + Element Plus + Pinia + Vue Router + PWA |
| 后端 | Go (Gin) + pgx + JWT + imaging |
| 数据库 | PostgreSQL 16 |
| 部署 | Docker Compose（PostgreSQL + Go + Nginx） |

## 📂 目录结构

```
food-log/
├── docs/           # 开发文档（架构 / 规划 / 进度 / API / 数据库）
├── scripts/        # 工具脚本（启动 / 迁移 / 构建 / 部署）
├── server/         # Go 后端
├── client/         # Vue 前端
├── docker-compose.yml
└── README.md
```

## 🚀 快速开始

### 方式一：Docker 一键部署

```bash
cp .env.example .env
bash scripts/docker-build.sh
# 访问 http://localhost
```

### 方式二：本地开发

```bash
# 1. 启动数据库
docker compose up -d db

# 2. 运行迁移
bash scripts/migrate.sh

# 3. 启动后端 (终端 1)
cd server && go run cmd/main.go

# 4. 启动前端 (终端 2)
cd client && npm install && npm run dev
# 访问 http://localhost:5173
```

## 📚 文档

| 文档 | 说明 |
|---|---|
| [架构设计](docs/architecture.md) | 系统架构、技术选型、交互流程 |
| [整体规划](docs/planning.md) | 功能清单、里程碑、迭代计划 |
| [开发进度](docs/progress.md) | 进度跟踪表 |
| [API 接口](docs/api.md) | 完整 API 文档 |
| [数据库设计](docs/database.md) | 表结构、索引、ER 图 |

## 🧩 常用脚本

```bash
bash scripts/start-dev.sh      # 一键启动开发环境
bash scripts/migrate.sh        # 执行数据库迁移
bash scripts/seed.sh           # 导入测试数据
bash scripts/build.sh          # 构建前后端
bash scripts/docker-build.sh   # Docker 全容器部署
```

## 📄 License

MIT
