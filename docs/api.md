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

`GET /api/recipes?page=1&page_size=10&keyword=&difficulty=&tag=&sort=created_at`

| 参数 | 类型 | 说明 |
|---|---|---|
| page | int | 页码，默认 1 |
| page_size | int | 每页数量，默认 10，最大 50 |
| keyword | string | 菜名搜索 |
| difficulty | string | 难度筛选 easy/medium/hard |
| tag | string | 标签筛选 |
| sort | string | created_at/rating/cook_time |

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

## 7. 数据模型速查

```typescript
interface Ingredient { name: string; amount: string; unit: string }
interface Step { order: number; content: string; image?: string }

interface Recipe {
  id: string; name: string; description?: string;
  ingredients: Ingredient[]; steps: Step[];
  cook_time_minutes?: number;
  difficulty: 'easy' | 'medium' | 'hard';
  rating?: number; tags: string[]; images: string[];
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
