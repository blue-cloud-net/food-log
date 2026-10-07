# 系统架构设计 (Architecture)

> 更新日期：2026-08-11

## 1. 系统概述

美食日志（Food Log）是一个个人美食记录应用，用于记录自制菜谱、餐厅探店和菜品点评，支持多端访问。

## 2. 架构图

```mermaid
graph TB
    subgraph 客户端
        A[Web 浏览器]
        B[手机浏览器 PWA]
        C[Android 安装版 PWA]
    end

    subgraph Nginx 前端服务器
        D[静态资源: Vue SPA]
        E[反向代理 /api → Go 后端]
    end

    subgraph Go 后端 :8080
        F[Gin Router]
        G[JWT 中间件]
        H[Handler 层]
        I[Service 层]
        J[Repository 层]
    end

    subgraph 数据层
        K[PostgreSQL 16]
        L[本地 data/images/ 图片存储]
    end

    A --> D
    B --> D
    C --> D
    D --> E
    E --> F
    F --> G
    G --> H
    H --> I
    I --> J
    J --> K
    I --> L
```

## 3. 技术选型与理由

| 领域 | 选型 | 理由 |
|---|---|---|
| 前端框架 | Vue 3 + Vite | 轻量、组合式 API、构建快 |
| UI 组件库 | Element Plus | Vue 3 生态、表单/上传组件成熟、响应式适配 |
| 状态管理 | Pinia | Vue 官方推荐、TypeScript 友好 |
| PWA | vite-plugin-pwa | 自动生成 manifest + Service Worker，可安装到 Android |
| 后端 | Go + Gin | 高性能、并发强、单二进制部署方便、静态类型安全 |
| 数据库驱动 | pgx/v5 | PostgreSQL 官方推荐驱动，连接池性能好 |
| 认证 | JWT (golang-jwt) | 无状态认证，适合前后端分离 + PWA |
| 密码加密 | bcrypt | 安全的密码哈希算法 |
| 图片处理 | imaging | 纯 Go 图片缩放/格式转换，无 CGO 依赖 |
| AI 标签/识别 | OpenAI 兼容协议 + Ollama | 通用 Provider 接口（chat + vision），词典优先 + AI 兜底 |
| 数据库 | PostgreSQL 16 | 功能完善，JSONB 支持灵活的结构化数据 |
| 部署 | Docker Compose | 一键启动数据库 + 后端 + 前端 |

## 4. 分层架构

```
┌─────────────────────────────────────────────┐
│ Router 层    路由注册、中间件装配            │
├─────────────────────────────────────────────┤
│ Middleware 层  JWT 认证 / CORS / 日志 / 限流  │
├─────────────────────────────────────────────┤
│ Handler 层    解析请求、校验参数、组装响应     │
├─────────────────────────────────────────────┤
│ Service 层    业务逻辑、事务、图片处理         │
├─────────────────────────────────────────────┤
│ Repository 层  SQL 查询、数据持久化           │
├─────────────────────────────────────────────┤
│ PostgreSQL / 文件存储                        │
└─────────────────────────────────────────────┘
```

依赖方向：Router → Middleware → Handler → Service → Repository（单向依赖，易于测试和维护）。

## 5. 前端架构

```
client/src/
├── router/       # 路由配置 + 登录守卫
├── stores/       # Pinia 状态（auth / recipe / restaurant）
├── api/          # Axios 封装 + 各模块 API
├── views/        # 页面组件
├── components/   # 布局 + 业务 + 通用组件
└── utils/        # 工具函数
```

### 5.1 路由设计

| 路径 | 页面 | 需登录 |
|---|---|---|
| `/login` | 登录 | ✗ |
| `/register` | 注册 | ✗ |
| `/` | 首页 Dashboard | ✓ |
| `/recipes` | 菜谱列表 | ✓ |
| `/recipes/new` | 新建菜谱 | ✓ |
| `/recipes/:id` | 菜谱详情 | ✓ |
| `/recipes/:id/edit` | 编辑菜谱 | ✓ |
| `/restaurants` | 餐厅列表 | ✓ |
| `/restaurants/new` | 新建餐厅 | ✓ |
| `/restaurants/:id` | 餐厅详情 | ✓ |
| `/restaurants/:id/edit` | 编辑餐厅 | ✓ |
| `/restaurants/:id/dishes/new` | 添加菜品 | ✓ |
| `/profile` | 个人中心 | ✓ |

### 5.2 响应式布局

- **移动端（<768px）**：底部 TabBar（首页/菜谱/餐厅/我的）
- **桌面端（≥768px）**：侧边栏导航
- 同一套组件，通过 CSS Media Query 自适应

### 5.3 状态管理

- `useAuthStore`：token、用户信息、登录/登出
- `useRecipeStore`：菜谱列表、当前菜谱、分页状态
- `useRestaurantStore`：餐厅列表、当前餐厅、菜品列表

## 6. 认证流程

```mermaid
sequenceDiagram
    participant U as 用户
    participant C as Vue 前端
    participant S as Go 后端
    participant DB as PostgreSQL

    U->>C: 输入账号密码
    C->>S: POST /api/auth/login
    S->>DB: 校验用户 + bcrypt 比对
    DB-->>S: 用户信息
    S-->>C: JWT token (7天)
    C->>C: token 存入 localStorage
    C->>S: 后续请求带 Authorization: Bearer <token>
    S->>S: 中间件验证 JWT 签名
    S-->>C: 业务数据
```

## 7. 图片处理流程

```mermaid
graph LR
    A[用户上传大图] --> B[multipart 落盘 data/tmp/]
    B --> C[解码图像]
    C --> D[重采样至 ≤1920px 宽]
    D --> E[编码为 JPEG quality 82]
    E --> F[另存 400px 宽缩略图 _thumb.jpg]
    F --> G[原图/缩略图原子重命名至 images/]
    G --> H[返回 原图URL + 缩略图URL]
```

存储规则：`images/{type}/YYYY/MM/uuid.jpg`（原图）与 `images/{type}/YYYY/MM/uuid_thumb.jpg`（缩略图），
其中 `type` 为图片用途（`recipe` / `restaurant`），由上传接口的 `type` 字段决定；
multipart 文件与编码结果先写入 `data/tmp/` 再原子重命名到位，服务启动时清空 `data/tmp/` 残留。

## 7.1 自动标签与 AI 识别

### 自动标签流程

```mermaid
graph TD
    A[创建/更新菜谱] --> B[校验并解析手选标签 id]
    B --> C[从 DB 加载词表与 tag_rules]
    C --> D{命中互斥组成员标签?}
    D -->|是| E[规则推导: 快手/汤/辣等]
    D -->|否| F[调用 AI Provider 兜底]
    E --> G[剔除与手选重复/同互斥组的项]
    F --> G
    G --> H[写入 recipe_ingredient_tags]
    B --> I[写入 recipe_tags]
    F -.AI 未配置/超时/失败.-> G
```

- **词表**：由数据库提供（`tag_categories` / `tags` / `tag_rules`），支持「全局预设 + 用户自定义」，见 [标签体系](./tags.md)
- **核心**（`internal/tagging/tagging.go`）：`Generate(in, rules)` 纯函数，按 `sort_order` 求值规则并返回标签 id；`HasCoreTags` 判断是否需要 AI 兜底
- **服务**（`internal/service/tag_service.go`）：词表读取（30s 进程内缓存 + 写入失效）、`ComputeIngredientTags`（去重与互斥组剔除）、`RecognizeTagNames`（名称→id，未收录则建自定义标签）、自定义分类/标签 CRUD
- **存储**：菜谱级手选标签写入 `recipe_tags`，食材级自动标签写入 `recipe_ingredient_tags`，两者分开
- **词表接口** `GET /api/recipes/tags` 供前端选择器与 id→显示名解析使用；管理接口见 `/api/tags`、`/api/tag-categories`

### AI Provider 架构

```mermaid
graph LR
    A[TagService / 识别接口] --> B[Provider 接口]
    B --> C[OpenAI 兼容实现]
    B --> D[Ollama 实现]
    C --> E[OpenAI/DeepSeek/通义/智谱]
    D --> F[本地 llava/qwen2.5-vl]
```

- 通过 `AI_PROVIDER`（openai | ollama）选择实现；`AI_BASE_URL`/`AI_API_KEY`/`AI_MODEL`/`AI_TIMEOUT` 配置
- OpenAI 兼容实现：标准库 `net/http` 直连 `/chat/completions`，支持文本与 `image_url` 视觉输入
- Ollama 实现：原生 `/api/chat`，图片下载转 base64 传 `images` 字段
- 前端可通过 `X-AI-Key` 请求头覆盖 API Key；`AI_API_KEY` 为空即禁用 AI
- 图片识别：`POST /api/recipes/ai/recognize` 识别菜名/食材/标签，回填菜谱表单

## 8. 安全设计

- 密码使用 bcrypt 加盐哈希存储
- JWT 密钥通过环境变量注入，生产环境强制更换
- 所有写操作接口均需 JWT 认证，且校验资源归属（user_id）
- 图片上传校验 MIME 类型与大小，重编码防止恶意文件
- CORS 白名单配置，仅允许可信来源

## 9. 部署架构

```mermaid
graph LR
    A[用户] -->|8080| B[app 容器 / Go 单进程]
    B -->|/api| C[API 路由]
    B -->|/images| D[图片静态服务]
    B -->|其余路径| E[内嵌前端 SPA]
    B -->|SQL| F[PostgreSQL :5432]
    B -->|读写| G[./data 数据目录]
    B -->|启动时| H[Schema 迁移 + admin 引导]
```

两个容器通过 Docker Compose 编排（`db` + `app`）：

- **app 单镜像**：前端构建产物通过 `go:embed` 编进 Go 二进制，同一进程提供 API、图片静态资源与 SPA（含深链接回退），不再需要 Nginx
- **数据持久化**：PostgreSQL 使用命名卷 `foodlog-pgdata`；应用数据（凭证 / 上传中转 / 图片）使用宿主机目录 `./data`（容器内 `/app/data`）
- **启动自愈**：app 启动时检测空库并初始化，随后执行增量迁移；生产模式还会按环境变量引导 admin 账号

### 发布流程

- `main` 分支 push：GitHub Actions 执行后端 `go vet` / `go build` / `go test` 与前端 `typecheck` / `build`
- 推送 `v*` tag：构建单镜像并推送 `ghcr.io/blue-cloud-net/food-log:<tag>` 与 `:latest`

### 数据目录

| 路径 | 内容 |
|---|---|
| `data/credentials/` | 自动生成的 admin 密码（仅生产、未设 `ADMIN_PASSWORD` 时） |
| `data/tmp/` | 上传中转与原子写入中间文件，服务启动时清空 |
| `data/images/recipe/` | 菜谱图片（原图 + `_thumb` 缩略图） |
| `data/images/restaurant/` | 餐厅与菜品图片 |

## 10. 目录结构

```
food-log/
├── Dockerfile               # 单镜像构建（前端构建 → go:embed → Go 二进制）
├── docker-compose.yml       # 生产编排（db + app）
├── docker-compose.dev.yml   # 开发编排（db + server[air] + client[Vite]）
├── data/                    # 运行时数据（credentials / tmp / images）
├── docs/                    # 开发文档
├── scripts/                 # 工具脚本（start-dev / stop-dev / build）
├── server/                  # Go 后端
│   ├── cmd/main.go          # 入口（建目录 → 迁移 → 引导 admin → 启动 HTTP）
│   ├── Dockerfile.dev       # 开发镜像（air 热重载）
│   ├── .air.toml            # air 配置
│   ├── internal/
│   │   ├── config/          # 配置加载（DATA_DIR / 运行模式 / admin）
│   │   ├── database/        # 连接、迁移执行、admin 引导
│   │   ├── storage/         # 数据目录布局与中转清理
│   │   ├── web/             # 内嵌前端产物 + SPA 服务
│   │   ├── model/           # 数据模型
│   │   ├── repository/      # 数据访问
│   │   ├── service/         # 业务逻辑（含图片上传处理）
│   │   ├── handler/         # HTTP 处理器
│   │   ├── middleware/      # 中间件
│   │   ├── tagging/         # 自动标签（词典 + 核心逻辑）
│   │   ├── ai/              # AI Provider（接口 + OpenAI/Ollama 实现）
│   │   └── router/          # 路由
│   └── migrations/          # SQL 迁移（embed，启动时执行）
└── client/                  # Vue 前端
    ├── Dockerfile.dev       # 开发镜像（Vite dev server）
    ├── public/              # 静态资源 + PWA 图标
    └── src/
```
