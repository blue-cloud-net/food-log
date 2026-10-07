# API 接口文档

> 更新日期：2026-10-07
> Base URL：`http://localhost:8080/api`
> 相关文档：[标签体系](./tags.md)、[库存食材](./inventory.md)、[数据库设计](./database.md)、[架构设计](./architecture.md)

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
  "avatar_url": "https://example.com/avatar.jpg"
}
```

## 3. 菜谱模块 `/recipes`

### 3.1 列表

`GET /api/recipes?page=1&page_size=10&keyword=&difficulty=&tag=&ingredient_tag=&sort=created_at&favorite=false&made=&liked=false`

| 参数 | 类型 | 说明 |
|---|---|---|
| page | int | 页码，默认 1 |
| page_size | int | 每页数量，默认 10，最大 50 |
| keyword | string | 菜名搜索 |
| difficulty | string | 难度筛选 easy/medium/hard |
| tag | string | **菜谱级**标签筛选（标签 id） |
| ingredient_tag | string | **食材级**标签筛选（标签 id） |
| sort | string | created_at/rating/cook_time |
| favorite | bool | `true` 时只返回已收藏菜谱 |
| made | string | `true` 只返回**已做**（`made_at` 非空）；`false` 只返回**未做**；缺省全部 |
| liked | bool | `true` 时只返回**喜欢**的菜谱（`is_liked`，独立于收藏） |

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
      "tags": ["<标签 id>"],
      "ingredient_tags": ["<标签 id>"],
      "images": ["/images/recipe/2026/08/xxx.jpg", "/images/recipe/2026/08/xxx_thumb.jpg"],
      "made_at": "2026-08-10",
      "is_liked": true,
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

> `tags` 为用户手选的**菜谱级**标签 id，`ingredient_tags` 为服务端按规则自动派生的**食材级**标签 id（只读）。
> 接口只返回 id，显示名/颜色通过 `GET /api/recipes/tags` + 前端 store 解析。详见 [标签体系](./tags.md)。

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
  "tags": ["<标签 id>"],
  "images": ["/images/recipe/2026/08/xxx.jpg"],
  "made_at": "2026-08-10",
  "is_liked": false
}
```

> 请求中的 `tags` 为标签 id 数组（仅**菜谱级**手选标签）。`ingredient_tags` 由服务端计算，请求中传入会被忽略。
> `made_at` 为做过日期（`YYYY-MM-DD`，可空，空/缺省 = 未做）；`is_liked` 为「喜欢」标记（独立于收藏）。
> 无效或不可见的标签 id 返回 400。响应 `data`：完整菜谱对象（含 `tags` 与 `ingredient_tags`）。

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

创建/更新菜谱时，服务端按数据库中的 `tag_rules` 从菜名、描述、食材、耗时推导**食材级标签**（写入 `ingredient_tags`），与用户手选的**菜谱级标签**（`tags`）分开存储：

- 自动结果会剔除与手选重复的标签
- 若手选已占用某互斥组（如荤菜/素菜），自动结果中同组标签一并剔除，避免矛盾展示
- 规则识别不足且已配置 AI 时用 AI 兜底（失败静默降级）
- 规则与词表详见 [标签体系](./tags.md)

### 3.7 标签词表

`GET /api/recipes/tags` 🔒

响应 `data`：当前用户可见的分类（全局预设 + 本人自定义）及其下标签，供前端选择器与 id→显示名解析使用。

```json
[
  {
    "id": "10000000-0000-4000-8000-000000000001",
    "name": "荤素",
    "color": "danger",
    "sort_order": 10,
    "is_system": true,
    "tags": [
      { "id": "20000000-0000-4000-8000-000000000001", "category_id": "10000000-0000-4000-8000-000000000001", "name": "荤菜", "mutex_group": "diet", "sort_order": 0, "is_system": true },
      { "id": "20000000-0000-4000-8000-000000000002", "category_id": "10000000-0000-4000-8000-000000000001", "name": "素菜", "mutex_group": "diet", "sort_order": 1, "is_system": true }
    ]
  }
]
```

> 用户自定义分类/标签额外带 `owner_id`；`is_system = true` 表示预设项（只读）。

### 3.8 随机选菜（今天吃什么）

`GET /api/recipes/random?tag=&ingredient_tag=&difficulty=` 🔒

从符合条件的菜谱中随机返回一条（`tag` / `ingredient_tag` 均为标签 id）。无匹配时返回 404。响应 `data` 为完整菜谱对象。

### 3.9 收藏 / 已做 / 喜欢

`POST /api/recipes/:id/favorite` 🔒

响应 `data`：`{ "favorited": true }`

`DELETE /api/recipes/:id/favorite` 🔒

响应 `data`：`{ "favorited": false }`

列表与详情接口均返回 `is_favorited` 字段；列表可通过 `favorite=true` 过滤。

`PUT /api/recipes/:id/made` 🔒

请求：`{ "made_at": "2026-08-10" }`；`made_at` 为 `null` / 空字符串时表示**取消已做**。

响应 `data`：`{ "made_at": "2026-08-10" }`（取消时为 `null`）

`PUT /api/recipes/:id/like` 🔒

请求：`{ "liked": true }`

响应 `data`：`{ "is_liked": true }`

列表与详情均返回 `made_at` / `is_liked`；列表可通过 `made=true|false`、`liked=true` 过滤（首页「自制」分组即基于此）。

### 3.10 AI 图片识别

`POST /api/recipes/ai/recognize` 🔒

请求：
```json
{ "image_url": "/images/recipe/2026/08/xxx.jpg" }
```

`image_url` 支持相对路径（自动拼接请求 Host）或完整 URL。可选请求头 `X-AI-Key` 覆盖后端配置的 API Key（OpenAI 兼容提供者）。

响应 `data`（识别结果，用于回填菜谱表单）：
```json
{
  "name": "清蒸鲈鱼",
  "ingredients": ["鲈鱼", "姜", "葱"],
  "tags": ["<标签 id>"]
}
```

> `tags` 为标签 id：AI 给出的名称先按当前用户可见词表匹配，未收录的名称会在该用户的「自定义」分类下创建为自定义标签。

未配置 AI 或不支持视觉模型时返回 400 明确提示。

## 3.11 标签管理 `/tags`、`/tag-categories` 🔒

用于维护**用户自定义**分类与标签。全局预设项（`is_system = true`）只读，修改返回 403；他人的自定义项返回 403，不存在的返回 404。

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/tags` | 新建自定义标签 |
| PUT | `/api/tags/:id` | 重命名/改互斥组/改排序 |
| DELETE | `/api/tags/:id` | 删除标签（级联解除菜谱关联） |
| POST | `/api/tag-categories` | 新建自定义分类 |
| PUT | `/api/tag-categories/:id` | 改名/改色/改排序 |
| DELETE | `/api/tag-categories/:id` | 删除分类（级联删除其下标签） |

新建标签请求（`category_id` 省略时自动归入该用户的「自定义」分类，不存在则自动创建）：
```json
{ "name": "外婆菜", "category_id": "<分类 id>", "mutex_group": "optional" }
```

新建分类请求：
```json
{ "name": "我的分类", "color": "warning" }
```

`color` 仅支持 `primary` / `success` / `warning` / `danger` / `info`。同名自定义标签/分类返回 409。

> 菜谱标签与探店标签（餐厅 / 菜品）是**三套独立字典**，互不影响：菜谱用 `/tags`、`/tag-categories`；餐厅用 `/restaurant-tags`、`/restaurant-tag-categories`；菜品用 `/dish-tags`、`/dish-tag-categories`。

## 4. 餐厅模块 `/restaurants`

### 4.1 列表

`GET /api/restaurants?page=1&page_size=10&keyword=&tag=&sort=&visited=`

| 参数 | 说明 |
|---|---|
| keyword | 模糊匹配店名 / 地址 |
| tag | **餐厅标签 id**，按标签筛选 |
| sort | `recommend`（推荐度）\| `value`（性价比）\| `ambience`（环境）\| `service`（服务）；缺省按创建时间倒序 |
| visited | `true` 只返回**已探店**；`false` 只返回**未探店**；缺省全部 |

响应 `data`：
```json
{
  "list": [
    {
      "id": "uuid",
      "name": "老四川",
      "address": "建设路 100 号",
      "description": "味道很正宗",
      "tags": ["<餐厅标签 id>"],
      "recommend_rating": 5,
      "value_rating": 4,
      "ambience_rating": 3,
      "service_rating": 4,
      "images": ["/images/restaurant/2026/08/rest.jpg"],
      "lat": 30.5,
      "lng": 104.0,
      "is_visited": true,
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
  "description": "味道很正宗",
  "tags": ["<餐厅标签 id>"],
  "recommend_rating": 5,
  "value_rating": 4,
  "ambience_rating": 3,
  "service_rating": 4,
  "images": ["/images/restaurant/2026/08/rest.jpg"],
  "lat": 30.5,
  "lng": 104.0,
  "is_visited": true
}
```

四个评分字段均为 1-5，`0` / 缺省表示未评分（存 NULL）。`is_visited` 为是否已探店（缺省 `false` = 未探店）。`tags` 为餐厅标签 id，必须是当前用户可见的标签，否则返回 400。

### 4.3 餐厅详情（含菜品）

`GET /api/restaurants/:id` 🔒

可选查询参数 `dish_tag`（**菜品标签 id**），用于只返回该标签下的菜品。

响应 `data`：
```json
{
  "id": "uuid",
  "name": "老四川",
  "address": "建设路 100 号",
  "description": "味道很正宗",
  "tags": ["<餐厅标签 id>"],
  "recommend_rating": 5,
  "value_rating": 4,
  "ambience_rating": 3,
  "service_rating": 4,
  "images": [],
  "lat": 30.5,
  "lng": 104.0,
  "is_visited": true,
  "dish_count": 1,
  "created_at": "2026-08-11T10:00:00Z",
  "dishes": [
    {
      "id": "uuid",
      "name": "水煮鱼",
      "description": "麻辣鲜香",
      "price": 68.00,
      "rating": 5,
      "tags": ["<菜品标签 id>"],
      "images": [],
      "eaten_at": "2026-08-10",
      "is_liked": false,
      "created_at": "2026-08-11T10:00:00Z"
    }
  ]
}
```

### 4.4 更新餐厅

`PUT /api/restaurants/:id` 🔒（请求体同创建）

### 4.5 删除餐厅

`DELETE /api/restaurants/:id` 🔒（级联删除其菜品与标签关联）

### 4.6 餐厅标签词表

`GET /api/restaurants/tags` 🔒

返回**餐厅标签**分类树（全局预设 + 本人自定义），结构与菜谱标签词表一致：`TagCategory[]`。

### 4.7 餐厅标签维护

`/restaurant-tags`、`/restaurant-tag-categories` 🔒

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/restaurant-tags` | 新建自定义餐厅标签 |
| PUT | `/api/restaurant-tags/:id` | 重命名/改互斥组/改排序 |
| DELETE | `/api/restaurant-tags/:id` | 删除标签（级联解除餐厅关联） |
| POST | `/api/restaurant-tag-categories` | 新建自定义分类 |
| PUT | `/api/restaurant-tag-categories/:id` | 改名/改色/改排序 |
| DELETE | `/api/restaurant-tag-categories/:id` | 删除分类（级联删除其下标签） |

请求体与 `/api/tags`、`/api/tag-categories` 完全一致。

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
  "tags": ["<菜品标签 id>"],
  "images": ["/images/restaurant/2026/08/fish.jpg"],
  "eaten_at": "2026-08-10",
  "is_liked": false
}
```

`rating` 为 1-5，界面上展示为**推荐度**；`0` / 缺省表示未评分（存 NULL）。`is_liked` 为「喜欢该菜品」标记。`tags` 为菜品标签 id，必须是当前用户可见的标签，否则返回 400。

### 5.2 更新菜品

`PUT /api/dishes/:id` 🔒（请求体同添加）

### 5.3 删除菜品

`DELETE /api/dishes/:id` 🔒

### 5.4 菜品标签词表

`GET /api/dishes/tags` 🔒

返回**菜品标签**分类树（全局预设 + 本人自定义）：`TagCategory[]`。

### 5.5 菜品标签维护

`/dish-tags`、`/dish-tag-categories` 🔒

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/dish-tags` | 新建自定义菜品标签 |
| PUT | `/api/dish-tags/:id` | 重命名/改互斥组/改排序 |
| DELETE | `/api/dish-tags/:id` | 删除标签（级联解除菜品关联） |
| POST | `/api/dish-tag-categories` | 新建自定义分类 |
| PUT | `/api/dish-tag-categories/:id` | 改名/改色/改排序 |
| DELETE | `/api/dish-tag-categories/:id` | 删除分类（级联删除其下标签） |

### 5.6 菜品列表（跨餐厅）

`GET /api/dishes?page=1&page_size=10&liked=&keyword=&tag=&restaurant_id=` 🔒

| 参数 | 说明 |
|---|---|
| liked | `true` 只返回**喜欢**的菜品 |
| keyword | 模糊匹配菜名 |
| tag | **菜品标签 id**，按标签筛选 |
| restaurant_id | 限定某餐厅下的菜品 |

返回分页菜品列表，每项额外带 `restaurant_name`（所属餐厅名），字段与详情内菜品一致。首页「喜欢菜品」分组使用 `liked=true`。

### 5.7 标记菜品喜欢

`PUT /api/dishes/:id/like` 🔒

请求：`{ "liked": true }`

响应 `data`：`{ "is_liked": true }`

## 6. 库存食材模块 `/inventory`、`/storage-locations`

记录当前拥有的食材，按「存放位置」分组。位置分 `fridge`（冰箱，含冷藏室/冷冻室/变温室/保鲜抽屉/门架）与 `outside`（外面，含室温台面/储物柜/荫凉通风处）两大类。**列表不分页**，分组由前端按位置词表完成。详见 [库存食材](./inventory.md)。

### 6.1 存放位置词表

`GET /api/storage-locations` 🔒

返回**当前用户可见**的存放位置平铺列表 `StorageLocation[]`（全局预设 + 本人自定义），已按 `area, sort_order` 排序：

```json
[
  { "id": "80000000-0000-4000-8000-000000000001", "area": "fridge", "name": "冷藏室", "sort_order": 10, "is_system": true },
  { "id": "<uuid>", "area": "fridge", "name": "冰箱上层", "sort_order": 900, "is_system": false }
]
```

`owner_id` 为空且 `is_system = true` 的为全局预设位置（只读）。

### 6.2 存放位置维护

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/storage-locations` | 新建自定义位置，请求体 `{ "area": "fridge", "name": "冰箱上层" }` |
| PUT | `/api/storage-locations/:id` | 改名 / 调整大类（仅本人自定义，预设返回 403） |
| DELETE | `/api/storage-locations/:id` | 删除位置（仅本人自定义；位置下仍有库存食材时返回 **409**） |

`area` 只接受 `fridge` / `outside`；与全局预设或本人已有位置重名返回 409。

### 6.3 库存列表

`GET /api/inventory?keyword=&area=&location_id=&expiring_within_days=&sort=` 🔒

| 参数 | 说明 |
|---|---|
| keyword | 模糊匹配食材名称 / 备注 |
| area | 存放位置大类：`fridge` \| `outside` |
| location_id | 指定存放位置 id |
| expiring_within_days | >0 时仅返回 N 天内到期（含已过期）的条目 |
| sort | `expire`（默认，按过期日升序，未设置排最后）\| `created` \| `name` |

响应 `data`：`{ "list": [InventoryItem], "total": 3 }`。条目字段：

```json
{
  "id": "<uuid>",
  "user_id": "<uuid>",
  "location_id": "<uuid>",
  "name": "鸡蛋",
  "amount": "6",
  "unit": "个",
  "category": "蛋奶",
  "expire_at": "2026-10-20",
  "note": "开封后冷藏",
  "images": ["/images/inventory/2026/10/uuid.jpg"],
  "created_at": "...",
  "updated_at": "..."
}
```

`amount` / `unit` / `category` 均为自由文本（`category` 前端给预设建议）；`expire_at` 为 `YYYY-MM-DD`，未设置为 `null`。

### 6.4 新增 / 更新 / 删除条目

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/inventory` | 新增条目 |
| GET | `/api/inventory/:id` | 条目详情 |
| PUT | `/api/inventory/:id` | 更新条目 |
| DELETE | `/api/inventory/:id` | 删除条目（食材用完） |

请求体：

```json
{
  "location_id": "<uuid>",
  "name": "鸡蛋",
  "amount": "6",
  "unit": "个",
  "category": "蛋奶",
  "expire_at": "2026-10-20",
  "note": "开封后冷藏",
  "images": []
}
```

`location_id` 必须是当前用户可见的位置，否则返回 400（参数错误）；条目归属校验失败返回 404 / 403。

### 6.5 图片上传

库存食材图片复用 `POST /api/upload`，`type` 传 `inventory`。

## 7. 图片上传 `/upload`

### 7.1 上传图片

`POST /api/upload` 🔒
Content-Type: `multipart/form-data`

| 字段 | 说明 |
|---|---|
| images | 文件数组，支持 jpg/png/webp/gif，单文件 ≤20MB，一次最多 9 张 |
| type | 图片用途：`recipe`（菜谱，默认）\| `restaurant`（餐厅与菜品）\| `inventory`（库存食材）；其它取值返回 400 |

图片按用途与日期落盘：`{DATA_DIR}/images/{type}/YYYY/MM/<uuid>.jpg`，同时生成 `_thumb.jpg` 缩略图。

响应 `data`：
```json
{
  "files": [
    {
      "url": "/images/recipe/2026/08/uuid.jpg",
      "thumb_url": "/images/recipe/2026/08/uuid_thumb.jpg"
    }
  ]
}
```

### 7.2 图片访问

上传后的图片通过静态服务直接访问：`GET /images/recipe/2026/08/uuid.jpg`（无需认证，浏览器可直接展示）。
路径不存在时返回 404，不回退到前端页面。

## 8. 全局搜索 `/search`

### 8.1 搜索

`GET /api/search?keyword=牛肉` 🔒

聚合搜索菜谱、餐厅、菜品，各类返回前 10 条。关键词会匹配：

- 菜谱：菜名 / 描述（含标签名）
- 餐厅：店名 / 地址 / **餐厅标签名**
- 菜品：菜名 / **菜品标签名**

响应 `data`：
```json
{
  "recipes": [ { "id": "uuid", "name": "土豆炖牛肉", "tags": [...] } ],
  "restaurants": [ { "id": "uuid", "name": "牛肉面馆", "tags": [...] } ],
  "dishes": [ { "id": "uuid", "name": "红烧牛肉", "restaurant_id": "uuid", "tags": [...] } ]
}
```

## 9. 数据导出 / 导入 `/export`

### 9.1 导出 JSON（完整备份）

`GET /api/export?format=json` 🔒

直接返回备份文件（Content-Disposition 提示下载），内容为 `ExportData` 格式（见下），包含菜谱/餐厅/菜品/收藏。

> **备份中的标签为标签名称（而非 id）**，因此备份可跨数据库导入；导入时按名称重新映射，未收录的名称会创建为用户自定义标签。

### 9.2 导出 CSV（菜谱单表）

`GET /api/export?format=csv` 🔒

返回 `recipes.csv`，列为：菜名、描述、食材、步骤、耗时(分钟)、难度、评分、**菜谱标签**、**食材标签**、创建时间（标签列输出标签名称，以 `/` 分隔）。

### 9.3 导入恢复

`POST /api/export/import?mode=append` 🔒

`body` 直接为导出的 JSON 备份内容。

| mode | 说明 |
|---|---|
| append（默认） | 按菜名去重追加，同名菜谱跳过 |
| overwrite | 先清空当前用户全部数据（收藏/菜品/菜谱/餐厅）再导入 |

响应 `data`：`{ "imported": true, "mode": "append" }`

## 10. 数据模型速查

```typescript
interface Ingredient { name: string; amount: string; unit: string }
interface Step { order: number; content: string; image?: string }

interface Tag {
  id: string; category_id: string; owner_id?: string;
  name: string; mutex_group?: string;
  sort_order: number; is_system: boolean;
}

interface TagCategory {
  id: string; owner_id?: string; name: string;
  color: 'primary' | 'success' | 'warning' | 'danger' | 'info';
  sort_order: number; is_system: boolean; tags: Tag[];
}

interface Recipe {
  id: string; name: string; description?: string;
  ingredients: Ingredient[]; steps: Step[];
  cook_time_minutes?: number;
  difficulty: 'easy' | 'medium' | 'hard' | '';
  rating?: number;
  /** 菜谱级标签 id（用户手选） */
  tags: string[];
  /** 食材级标签 id（服务端自动派生） */
  ingredient_tags: string[];
  images: string[];
  /** 做过日期 YYYY-MM-DD（null = 未做） */
  made_at: string | null;
  /** 喜欢（独立于收藏） */
  is_liked: boolean;
  is_favorited: boolean;
  created_at: string; updated_at: string;
}

interface Restaurant {
  id: string; name: string; address?: string;
  description?: string;
  /** 餐厅标签 id */
  tags: string[];
  /** 1-5 推荐度 */
  recommend_rating: number;
  /** 1-5 性价比 */
  value_rating: number;
  /** 1-5 环境 */
  ambience_rating: number;
  /** 1-5 服务 */
  service_rating: number;
  images: string[];
  lat?: number; lng?: number;
  /** 是否已探店（false = 未探店） */
  is_visited: boolean;
  dish_count?: number; created_at: string;
}

interface Dish {
  id: string; restaurant_id: string; name: string;
  description?: string; price?: number;
  /** 1-5（界面展示为「推荐度」） */
  rating?: number;
  /** 菜品标签 id */
  tags: string[];
  images: string[]; eaten_at?: string;
  /** 喜欢该菜品 */
  is_liked: boolean;
  /** 列表接口返回的所属餐厅名 */
  restaurant_name?: string;
  created_at: string;
}

interface StorageLocation {
  id: string;
  /** 为空表示全局预设位置（is_system = true，只读） */
  owner_id?: string;
  /** fridge 冰箱内 | outside 外面 */
  area: 'fridge' | 'outside';
  name: string;
  sort_order: number;
  is_system: boolean;
}

interface InventoryItem {
  id: string; user_id: string;
  /** 存放位置 id（前端用位置词表解析名称与分组） */
  location_id: string;
  name: string;
  amount: string; unit: string;
  /** 食材分类（自由文本） */
  category: string;
  /** 过期日期 YYYY-MM-DD（null = 未设置） */
  expire_at: string | null;
  note: string;
  images: string[];
  created_at: string; updated_at: string;
}
```
