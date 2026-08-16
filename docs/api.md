# API 接口文档

> 更新日期：2026-08-11
> Base URL：`http://localhost:8080/api`

## 1. 通用约定

### 1.1 响应格式

所有接口返回统一格式：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

| 字段 | 类型 | 说明 |
|---|---|---|
| code | int | 0 成功，非 0 失败 |
| message | string | 提示信息 |
| data | any | 业务数据 |

### 1.2 错误码

| code | 含义 |
|---|---|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 未认证 / token 无效 |
| 1003 | 无权限 |
| 1004 | 资源不存在 |
| 1005 | 数据已存在 |
| 1006 | 服务器内部错误 |
| 1007 | 上传失败 |

### 1.3 认证

登录后所有业务接口需在请求头携带：

```
Authorization: Bearer <token>
```

## 2. 认证模块 `/auth`

### 2.1 注册

`POST /api/auth/register`

请求：
```json
{
  "username": "小明",
  "email": "ming@example.com",
  "password": "123456"
}
```

响应 `data`：
```json
{
  "token": "eyJhbGciOi...",
  "user": {
    "id": "uuid",
    "username": "小明",
    "email": "ming@example.com",
    "avatar_url": null,
    "created_at": "2026-08-11T10:00:00Z"
  }
}
```

### 2.2 登录

`POST /api/auth/login`

请求：
```json
{
  "username": "小明",
  "password": "123456"
}
```

响应同注册。

### 2.3 获取当前用户

`GET /api/auth/me` 🔒

响应 `data`：
```json
{
  "id": "uuid",
  "username": "小明",
  "email": "ming@example.com",
  "avatar_url": null,
  "recipe_count": 12,
  "restaurant_count": 8,
  "dish_count": 25,
  "created_at": "2026-08-11T10:00:00Z"
}
```

### 2.4 更新当前用户

`PUT /api/auth/me` 🔒

请求（部分字段）：
```json
{
  "username": "新名字",
  "email": "new@example.com",
  "password": "新密码(可选)",
  "avatar_url": "/uploads/avatar.jpg"
}
```

## 3. 菜谱模块 `/recipes`

### 3.1 列表

`GET /api/recipes?page=1&page_size=10&keyword=&difficulty=&tag=&sort=created_at&favorite=false`

| 参数 | 类型 | 说明 |
|---|---|---|
| page | int | 页码，默认 1 |
| page_size | int | 每页数量，默认 10，最大 50 |
| keyword | string | 菜名搜索 |
| difficulty | string | 难度筛选 easy/medium/hard |
| tag | string | 标签筛选（如 `素菜`） |
| sort | string | created_at/rating/cook_time |
| favorite | bool | `true` 时只返回已收藏菜谱 |

响应 `data`：
```json
{
  "list": [
    {
      "id": "uuid",
      "name": "麻婆豆腐",
      "description": "家常做法",
      "ingredients": [{"name": "豆腐", "amount": "1", "unit": "块"}],
      "steps": [{"order": 1, "content": "切豆腐", "image": null}],
      "cook_time_minutes": 20,
      "difficulty": "medium",
      "rating": 5,
      "tags": ["川菜", "下饭菜"],
      "images": ["/uploads/2026/08/xxx.jpg", "/uploads/2026/08/xxx_thumb.jpg"],
      "is_favorited": true,
      "created_at": "2026-08-11T10:00:00Z",
      "updated_at": "2026-08-11T10:00:00Z"
    }
  ],
  "total": 12,
  "page": 1,
  "page_size": 10
}
```

### 3.2 创建菜谱

`POST /api/recipes` 🔒

请求：
```json
{
  "name": "麻婆豆腐",
  "description": "家常做法",
  "ingredients": [{"name": "豆腐", "amount": "1", "unit": "块"}],
  "steps": [{"order": 1, "content": "切豆腐", "image": null}],
  "cook_time_minutes": 20,
  "difficulty": "medium",
  "rating": 5,
  "tags": ["川菜"],
  "images": ["/uploads/2026/08/xxx.jpg"]
}
```

响应 `data`：完整菜谱对象。

### 3.3 菜谱详情

`GET /api/recipes/:id` 🔒

响应 `data`：完整菜谱对象（同上）。

### 3.4 更新菜谱

`PUT /api/recipes/:id` 🔒

请求体同创建（部分字段）。响应 `data`：更新后的菜谱。

### 3.5 删除菜谱

`DELETE /api/recipes/:id` 🔒

响应 `data`：`{ "deleted": true }`

### 3.6 自动标签说明

创建/更新菜谱时，后端会依据食材清单自动生成标签（荤菜/素菜、牛肉、鱼、海鲜、豆制品、蔬菜、蛋、快手、辣、汤等），与请求中手动 `tags` 合并去重后保存。词典未识别时若已配置 AI 则调用 AI 兜底（失败静默降级，仅保留词典结果）。

### 3.7 预设标签词表

`GET /api/recipes/tags` 🔒

响应 `data`：按分类返回预设词表，供前端标签选择器/自动补全使用。
```json
[
  { "name": "荤素", "tags": ["荤菜", "素菜"] },
  { "name": "食材", "tags": ["牛肉", "猪肉", "鸡肉", "羊肉", "鸭肉", "鱼", "海鲜", "虾", "蟹", "蛋", "豆制品", "蔬菜", "菌菇", "主食"] },
  { "name": "场景", "tags": ["快手", "汤", "凉菜", "面食", "甜点"] },
  { "name": "时段", "tags": ["早餐", "午餐", "晚餐", "夜宵"] },
  { "name": "菜系", "tags": ["川菜", "粤菜", "湘菜", "鲁菜", "苏菜", "浙菜", "闽菜", "徽菜", "东北菜", "西北菜"] },
  { "name": "口味", "tags": ["辣", "清淡", "甜", "酸", "咸鲜"] }
]
```

### 3.8 随机选菜（今天吃什么）

`GET /api/recipes/random?tag=&difficulty=` 🔒

从符合条件的菜谱中随机返回一条。无匹配时返回 404。响应 `data` 为完整菜谱对象。

### 3.9 收藏 / 取消收藏

`POST /api/recipes/:id/favorite` 🔒

响应 `data`：`{ "favorited": true }`

`DELETE /api/recipes/:id/favorite` 🔒

响应 `data`：`{ "favorited": false }`

列表与详情接口均返回 `is_favorited` 字段；列表可通过 `favorite=true` 过滤。

### 3.10 AI 图片识别

`POST /api/recipes/ai/recognize` 🔒

请求：
```json
{ "image_url": "/uploads/2026/08/xxx.jpg" }
```

`image_url` 支持相对路径（自动拼接请求 Host）或完整 URL。可选请求头 `X-AI-Key` 覆盖后端配置的 API Key（OpenAI 兼容提供者）。

响应 `data`（识别结果，用于回填菜谱表单）：
```json
{
  "name": "清蒸鲈鱼",
  "ingredients": ["鲈鱼", "姜", "葱"],
  "tags": ["荤菜", "鱼"]
}
```

未配置 AI 或不支持视觉模型时返回 400 明确提示。

## 4. 餐厅模块 `/restaurants`

### 4.1 列表

`GET /api/restaurants?page=1&page_size=10&keyword=&cuisine_type=&sort=created_at`

响应 `data`：
```json
{
  "list": [
    {
      "id": "uuid",
      "name": "老四川",
      "address": "建设路 100 号",
      "cuisine_type": "川菜",
      "description": "味道很正宗",
      "avg_rating": 4.5,
      "images": ["/uploads/2026/08/rest.jpg"],
      "lat": 30.5,
      "lng": 104.0,
      "dish_count": 6,
      "created_at": "2026-08-11T10:00:00Z"
    }
  ],
  "total": 8,
  "page": 1,
  "page_size": 10
}
```

### 4.2 创建餐厅

`POST /api/restaurants` 🔒

请求：
```json
{
  "name": "老四川",
  "address": "建设路 100 号",
  "cuisine_type": "川菜",
  "description": "味道很正宗",
  "avg_rating": 4.5,
  "images": ["/uploads/2026/08/rest.jpg"],
  "lat": 30.5,
  "lng": 104.0
}
```

### 4.3 餐厅详情（含菜品）

`GET /api/restaurants/:id` 🔒

响应 `data`：
```json
{
  "id": "uuid",
  "name": "老四川",
  "address": "建设路 100 号",
  "cuisine_type": "川菜",
  "description": "味道很正宗",
  "avg_rating": 4.5,
  "images": [],
  "lat": 30.5,
  "lng": 104.0,
  "created_at": "2026-08-11T10:00:00Z",
  "dishes": [
    {
      "id": "uuid",
      "name": "水煮鱼",
      "description": "麻辣鲜香",
      "price": 68.00,
      "rating": 5,
      "images": [],
      "eaten_at": "2026-08-10",
      "created_at": "2026-08-11T10:00:00Z"
    }
  ]
}
```

### 4.4 更新餐厅

`PUT /api/restaurants/:id` 🔒

### 4.5 删除餐厅

`DELETE /api/restaurants/:id` 🔒（级联删除其菜品）

## 5. 菜品模块 `/dishes`

### 5.1 添加菜品

`POST /api/restaurants/:id/dishes` 🔒

请求：
```json
{
  "name": "水煮鱼",
  "description": "麻辣鲜香",
  "price": 68.00,
  "rating": 5,
  "images": ["/uploads/2026/08/fish.jpg"],
  "eaten_at": "2026-08-10"
}
```

### 5.2 更新菜品

`PUT /api/dishes/:id` 🔒

### 5.3 删除菜品

`DELETE /api/dishes/:id` 🔒

## 6. 图片上传 `/upload`

### 6.1 上传图片

`POST /api/upload` 🔒
Content-Type: `multipart/form-data`

| 字段 | 说明 |
|---|---|
| images | 文件数组，支持 jpg/png/webp/gif，单文件 ≤20MB |

响应 `data`：
```json
{
  "files": [
    {
      "url": "/uploads/2026/08/uuid.jpg",
      "thumb_url": "/uploads/2026/08/uuid_thumb.jpg"
    }
  ]
}
```

### 6.2 图片访问

上传后的图片通过静态服务直接访问：`GET /uploads/2026/08/uuid.jpg`（无需认证，浏览器可直接展示）。

## 7. 全局搜索 `/search`

### 7.1 搜索

`GET /api/search?keyword=牛肉` 🔒

聚合搜索菜谱、餐厅、菜品，各类返回前 10 条。响应 `data`：
```json
{
  "recipes": [ { "id": "uuid", "name": "土豆炖牛肉", "tags": [...] } ],
  "restaurants": [ { "id": "uuid", "name": "牛肉面馆", "cuisine_type": "面食" } ],
  "dishes": [ { "id": "uuid", "name": "红烧牛肉", "restaurant_id": "uuid" } ]
}
```

## 8. 数据导出 / 导入 `/export`

### 8.1 导出 JSON（完整备份）

`GET /api/export?format=json` 🔒

直接返回备份文件（Content-Disposition 提示下载），内容为 `ExportData` 格式（见下），包含菜谱/餐厅/菜品/收藏。

### 8.2 导出 CSV（菜谱单表）

`GET /api/export?format=csv` 🔒

返回 `recipes.csv`，列为：菜名、描述、食材、步骤、耗时(分钟)、难度、评分、标签、创建时间。

### 8.3 导入恢复

`POST /api/export/import?mode=append` 🔒

`body` 直接为导出的 JSON 备份内容。

| mode | 说明 |
|---|---|
| append（默认） | 按菜名去重追加，同名菜谱跳过 |
| overwrite | 先清空当前用户全部数据（收藏/菜品/菜谱/餐厅）再导入 |

响应 `data`：`{ "imported": true, "mode": "append" }`

## 9. 数据模型速查

```typescript
interface Ingredient { name: string; amount: string; unit: string }
interface Step { order: number; content: string; image?: string }

interface Recipe {
  id: string; name: string; description?: string;
  ingredients: Ingredient[]; steps: Step[];
  cook_time_minutes?: number;
  difficulty: 'easy' | 'medium' | 'hard' | '';
  rating?: number; tags: string[]; images: string[];
  is_favorited: boolean;
  created_at: string; updated_at: string;
}

interface Restaurant {
  id: string; name: string; address?: string;
  cuisine_type?: string; description?: string;
  avg_rating?: number; images: string[];
  lat?: number; lng?: number;
  dish_count?: number; created_at: string;
}

interface Dish {
  id: string; restaurant_id: string; name: string;
  description?: string; price?: number; rating?: number;
  images: string[]; eaten_at?: string; created_at: string;
}
```
