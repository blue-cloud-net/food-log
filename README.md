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
| 部署 | Docker Compose（PostgreSQL + 单镜像：Go 内嵌前端产物） |
| CI/CD | GitHub Actions（main 编译检查 / tag 发布 GHCR 镜像） |

## 📂 目录结构

```
food-log/
├── Dockerfile               # 单镜像构建（前端构建 → go:embed → Go 二进制）
├── docker-compose.yml       # 生产编排（db + app）
├── docker-compose.dev.yml   # 开发编排（db + server[air] + client[Vite]）
├── data/                    # 运行时数据（credentials / tmp / images）
├── docs/                    # 开发文档（架构 / API / 数据库 / 标签）
├── scripts/                 # 工具脚本（start-dev / stop-dev / build）
├── server/                  # Go 后端（含内嵌迁移与前端产物）
└── client/                  # Vue 前端
```

## 🚀 快速开始

### 方式一：开发（容器化热重载）

```bash
cp .env.example .env         # 首次执行，按需修改 JWT_SECRET
bash scripts/start-dev.sh    # 启动 db + 后端(air) + 前端(Vite)

# 前端 http://localhost:5173  （Vite HMR）
# 后端 http://localhost:8080  （air 监听 *.go / *.sql 自动重编译）
# 默认账号 admin / admin      （开发模式自动加载演示数据）
```

源码目录通过 volume 挂载进容器，改代码即生效；停止用 `bash scripts/start-dev.sh --stop`。

### 方式二：生产部署（单镜像）

```bash
cp .env.example .env         # 必须修改 JWT_SECRET
docker compose up -d --build
# 访问 http://localhost:8080
```

前端产物通过 `go:embed` 编进 Go 二进制，**单一进程**同时提供 API、图片与前端页面，不再需要 Nginx。

首次启动会自动完成：空库初始化 → 增量迁移 → 引导管理员账号。
生产模式（默认）不设置 `ADMIN_PASSWORD` 时会生成随机密码并写入数据目录：

```bash
cat data/credentials/admin-password.txt
```

也可指定固定密码：`ADMIN_PASSWORD=your-password docker compose up -d`

### 可选：使用 GHCR 预构建镜像

```bash
APP_IMAGE=ghcr.io/blue-cloud-net/food-log:latest docker compose up -d
```

## 📚 文档

| 文档 | 说明 |
|---|---|
| [架构设计](docs/architecture.md) | 系统架构、部署架构、技术选型、交互流程 |
| [API 接口](docs/api.md) | 完整 API 文档 |
| [数据库设计](docs/database.md) | 表结构、索引、迁移与初始化流程 |
| [标签体系](docs/tags.md) | 标签分类、预设词表、自动标签规则 |

## 🧩 常用脚本

| 命令 | 说明 |
|---|---|
| `bash scripts/start-dev.sh` | 启动开发环境（db + air 热重载后端 + Vite 前端） |
| `bash scripts/start-dev.sh --logs` | 同上，并持续跟踪容器日志 |
| `bash scripts/start-dev.sh --stop` | 停止开发环境 |
| `bash scripts/build.sh` | 本地一体化构建（前端产物 → 内嵌目录 → Go 二进制） |
| `docker compose up -d --build` | 生产部署 |
| `docker compose logs -f app` | 查看应用日志 |

## 📦 数据目录（`DATA_DIR`）

| 路径 | 说明 |
|---|---|
| `data/credentials/` | 自动生成的 admin 密码（仅生产、未设 `ADMIN_PASSWORD` 时） |
| `data/tmp/` | 上传中转与原子写入的中间文件，服务启动时清空 |
| `data/images/recipe/` | 菜谱图片（`YYYY/MM/<uuid>.jpg` + `_thumb.jpg`） |
| `data/images/restaurant/` | 餐厅与菜品图片 |

> 图片对外路径为 `/images/{type}/YYYY/MM/<uuid>.jpg`，由 Go 进程直接提供静态服务。
> 备份时需同时覆盖 `data/` 目录与 PostgreSQL 命名卷 `foodlog-pgdata`。

## 📄 License

MIT
