# 数据库设计 (Database)

> 更新日期：2026-08-11
> 数据库：PostgreSQL 16

## 1. ER 图

```mermaid
erDiagram
    users ||--o{ recipes : "拥有"
    users ||--o{ restaurants : "拥有"
    users ||--o{ dishes : "评价"
    restaurants ||--o{ dishes : "包含"

    users {
        uuid id PK
        varchar username
        varchar email
        varchar password_hash
        text avatar_url
        timestamptz created_at
        timestamptz updated_at
    }
    recipes {
        uuid id PK
        uuid user_id FK
        varchar name
        text description
        jsonb ingredients
        jsonb steps
        int cook_time_minutes
        varchar difficulty
        smallint rating
        jsonb tags
        jsonb images
        timestamptz created_at
        timestamptz updated_at
    }
    restaurants {
        uuid id PK
        uuid user_id FK
        varchar name
        text address
        varchar cuisine_type
        text description
        numeric avg_rating
        jsonb images
        double lat
        double lng
        timestamptz created_at
        timestamptz updated_at
    }
    dishes {
        uuid id PK
        uuid restaurant_id FK
        uuid user_id FK
        varchar name
        text description
        numeric price
        smallint rating
        jsonb images
        date eaten_at
        timestamptz created_at
    }
```

## 2. 表结构

### 2.1 users — 用户表

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() | 主键 |
| username | VARCHAR(50) | UNIQUE, NOT NULL | 用户名 |
| email | VARCHAR(255) | UNIQUE, NOT NULL | 邮箱 |
| password_hash | VARCHAR(255) | NOT NULL | bcrypt 哈希 |
| avatar_url | TEXT | | 头像 |
| created_at | TIMESTAMPTZ | DEFAULT now() | |
| updated_at | TIMESTAMPTZ | DEFAULT now() | |

### 2.2 recipes — 自制菜谱表

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | UUID | PK | 主键 |
| user_id | UUID | FK → users(id), NOT NULL | 所属用户 |
| name | VARCHAR(200) | NOT NULL | 菜名 |
| description | TEXT | | 描述 |
| ingredients | JSONB | | 食材 `[{name, amount, unit}]` |
| steps | JSONB | | 步骤 `[{order, content, image}]` |
| cook_time_minutes | INT | | 烹饪耗时（分钟） |
| difficulty | VARCHAR(20) | CHECK in (easy,medium,hard) | 难度 |
| rating | SMALLINT | CHECK 1-5 | 自评 |
| tags | JSONB | | 标签 `["川菜","快手菜"]` |
| images | JSONB | | 图片 `["/uploads/xx.jpg"]` |
| created_at | TIMESTAMPTZ | | |
| updated_at | TIMESTAMPTZ | | |

### 2.3 restaurants — 餐厅表

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | UUID | PK | 主键 |
| user_id | UUID | FK → users(id), NOT NULL | 所属用户 |
| name | VARCHAR(200) | NOT NULL | 店名 |
| address | TEXT | | 地址 |
| cuisine_type | VARCHAR(100) | | 菜系 |
| description | TEXT | | 备注 |
| avg_rating | NUMERIC(2,1) | CHECK 0-5 | 综合评分 |
| images | JSONB | | 环境照片 |
| lat | DOUBLE PRECISION | | 纬度 |
| lng | DOUBLE PRECISION | | 经度 |
| created_at | TIMESTAMPTZ | | |
| updated_at | TIMESTAMPTZ | | |

### 2.4 dishes — 店内菜品表

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | UUID | PK | 主键 |
| restaurant_id | UUID | FK → restaurants(id), NOT NULL | 所属餐厅 |
| user_id | UUID | FK → users(id), NOT NULL | 评价用户 |
| name | VARCHAR(200) | NOT NULL | 菜名 |
| description | TEXT | | 口味描述 |
| price | NUMERIC(10,2) | | 价格 |
| rating | SMALLINT | CHECK 1-5 | 评分 |
| images | JSONB | | 照片 |
| eaten_at | DATE | | 就餐日期 |
| created_at | TIMESTAMPTZ | | |

## 3. 索引策略

```sql
-- users
CREATE UNIQUE INDEX idx_users_username ON users(username);
CREATE UNIQUE INDEX idx_users_email ON users(email);

-- recipes（按用户查询 + 按创建时间排序）
CREATE INDEX idx_recipes_user_created ON recipes(user_id, created_at DESC);
CREATE INDEX idx_recipes_tags ON recipes USING GIN (tags);
CREATE INDEX idx_recipes_name ON recipes(name);

-- restaurants
CREATE INDEX idx_restaurants_user_created ON restaurants(user_id, created_at DESC);
CREATE INDEX idx_restaurants_cuisine ON restaurants(cuisine_type);

-- dishes（按餐厅查询 + 按就餐日期排序）
CREATE INDEX idx_dishes_restaurant ON dishes(restaurant_id, eaten_at DESC);
```

## 4. 关键设计说明

- **JSONB 存储结构化数据**：食材、步骤、标签、图片使用 JSONB，避免过度建表，同时支持 GIN 索引查询
- **UUID 主键**：使用 PostgreSQL 内置 `gen_random_uuid()`，避免自增主键暴露数据量
- **软删除策略**：当前采用硬删除（DELETE），后续如需可回收再引入 deleted_at
- **菜品与餐厅强关联**：菜品通过 restaurant_id 归属餐厅，级联删除保证数据一致
- **updated_at 自动更新**：通过触发器或应用层维护

## 5. 迁移文件

| 文件 | 内容 |
|---|---|
| `server/migrations/001_init.sql` | 建表 + 索引 + 触发器 |
| `server/migrations/002_seed.sql` | 测试种子数据 |

## 6. 迁移记录

| 版本 | 日期 | 描述 |
|---|---|---|
| 001 | 2026-08-11 | 初始建表 |
