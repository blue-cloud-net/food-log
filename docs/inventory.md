# 库存食材（冰箱 / 外面）

> 更新日期：2026-10-07
> 相关文档：[数据库设计](./database.md)、[API 接口](./api.md)、[架构设计](./architecture.md)

「库存食材」记录当前拥有、还没用完的食材，并按**存放位置**归类。位置分两大类：

- `fridge`（冰箱）：冷藏室 / 冷冻室 / 变温室·零度保鲜 / 保鲜抽屉（果蔬）/ 冰箱门架
- `outside`（外面）：室温·台面 / 储物柜·橱柜 / 荫凉通风处

位置本身是**字典**（可用户自定义），而不是写死的枚举，便于各家按自己的冰箱结构命名。

## 1. 表结构

见 `server/migrations/011_inventory.sql`（结构）与 `012_storage_location_catalog.sql`（预设位置）。

| 表 | 说明 |
|---|---|
| `storage_locations` | 存放位置字典，`area ∈ {fridge, outside}` |
| `inventory_items` | 库存食材条目，`user_id` 归属用户，`location_id` 指向一个存放位置 |

### storage_locations

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | UUID | PK, DEFAULT gen_random_uuid() | 主键 |
| owner_id | UUID | FK → users(id), ON DELETE CASCADE | `NULL` = 全局预设 |
| area | VARCHAR(20) | CHECK in (fridge, outside) | 大类 |
| name | VARCHAR(50) | NOT NULL | 位置名 |
| sort_order | INT | DEFAULT 0 | 排序（自定义位置固定 900，排在预设之后） |
| is_system | BOOLEAN | DEFAULT false | 全局预设 = true（只读） |
| created_at / updated_at | TIMESTAMPTZ | DEFAULT now() | 带 `updated_at` 触发器 |

- `UNIQUE NULLS NOT DISTINCT (owner_id, area, name)`：同一 owner 同一大类下名称唯一
- 服务层额外阻止与全局预设重名，避免列表里出现两个「冷藏室」

### inventory_items

| 列 | 类型 | 约束 | 说明 |
|---|---|---|---|
| id | UUID | PK | 主键 |
| user_id | UUID | FK → users(id), ON DELETE CASCADE, NOT NULL | 所属用户 |
| location_id | UUID | FK → storage_locations(id), **ON DELETE RESTRICT**, NOT NULL | 存放位置 |
| name | VARCHAR(100) | NOT NULL | 食材名称 |
| amount | VARCHAR(30) | DEFAULT '' | 数量（自由文本，沿用菜谱 `ingredients.amount` 语义） |
| unit | VARCHAR(20) | DEFAULT '' | 单位（个 / 克 / 盒…） |
| category | VARCHAR(20) | DEFAULT '' | 食材分类（自由文本，前端给预设建议） |
| expire_at | DATE | | 保质期 / 过期日期，可空 |
| note | VARCHAR(500) | DEFAULT '' | 备注 |
| images | JSONB | DEFAULT '[]' | 图片 `["/images/inventory/2026/10/xx.jpg"]` |
| created_at / updated_at | TIMESTAMPTZ | DEFAULT now() | 带 `updated_at` 触发器 |

## 2. 可见性与权限

存放位置与标签字典同构：

- `owner_id IS NULL` → 全局预设位置，`is_system = true`，对所有人**只读**（改 / 删返回 403）
- `owner_id = 当前用户` → 私有自定义位置，可改可删
- 他人私有位置对本用户按「不存在」处理（404），不泄露存在性

**删除自定义位置**：位置下仍有库存食材时返回 **409**，不做级联删除（`ON DELETE RESTRICT`），避免误删位置连带清空库存。需先移走或删除该位置下的食材。

## 3. 库存条目

- `amount` / `unit` / `category` 都是自由文本，前端给建议值：数量/单位随手输入，分类提供 `肉禽 / 水产 / 蔬菜 / 水果 / 蛋奶 / 豆制品 / 主食 / 调味 / 干货 / 其他` 且允许自定义
- `expire_at` 为空表示未记录保质期；请求体传空串会被归一化为 `NULL`
- 条目归属校验：详情 / 更新 / 删除按 `user_id` 判定，非本人返回 403，不存在返回 404

## 4. 交互行为

- **两级分组**：列表按 `冰箱 / 外面` → 具体位置 分组渲染；接口返回平铺列表 + `location_id`，前端用位置词表（`stores/storageLocations.ts`）分组，符合「词表只传 id，显示名由前端解析」的既有约定
- **不分页**：家庭库存量级小（几十项），一次返回全部，避免分组与分页冲突
- **保质期状态**：已过期（红）/ 临期 ≤3 天（黄）/ 充足（绿），阈值与文案见 `client/src/utils/inventory.ts`
- **筛选**：关键词（名称 / 备注）、大类、指定位置、N 天内到期（含已过期）
- **默认排序**：按过期日升序，未设置过期日的排最后（`expire_at ASC NULLS LAST`）

## 5. 全局预设固定 UUID

| 大类 | 位置 | UUID |
|---|---|---|
| fridge | 冷藏室 | `80000000-0000-4000-8000-000000000001` |
| fridge | 冷冻室 | `80000000-0000-4000-8000-000000000002` |
| fridge | 变温室 / 零度保鲜 | `80000000-0000-4000-8000-000000000003` |
| fridge | 保鲜抽屉（果蔬） | `80000000-0000-4000-8000-000000000004` |
| fridge | 冰箱门架 | `80000000-0000-4000-8000-000000000005` |
| outside | 室温 / 台面 | `80000000-0000-4000-8000-000000000011` |
| outside | 储物柜 / 橱柜 | `80000000-0000-4000-8000-000000000012` |
| outside | 荫凉通风处 | `80000000-0000-4000-8000-000000000013` |

沿用 [标签体系](./tags.md) 的 UUID 分段约定，`80000000-` 号段专用于存放位置（01-0f 冰箱内，11-1f 外面）。

## 6. 初始化与数据

| 文件 | 内容 | 执行环境 |
|---|---|---|
| `011_inventory.sql` | `storage_locations` / `inventory_items` + 索引 + 触发器 | 开发 + 生产 |
| `012_storage_location_catalog.sql` | 8 个全局预设存放位置（固定 UUID + `ON CONFLICT (id) DO UPDATE`） | 开发 + 生产 |

> 本功能**暂无演示数据**：演示文件 `009_demo_seed.sql` 按文件名排序早于 `011_inventory.sql`，
> 全新库执行到 009 时库存表尚未建立，无法在其中播种库存数据。

## 7. 图片上传

库存食材图片复用 `POST /api/upload`，`type` 传 `inventory`，落地到 `{DATA_DIR}/images/inventory/YYYY/MM/`。

## 8. 边界（当前不支持）

- 数据导出 / 导入（`/export`）暂不包含库存食材
- 全局搜索（`/search`）暂不检索库存食材
- 首页统计（`/auth/me` 的 `recipe_count` 等）暂不含库存计数
- 暂无「用完 / 减量」快捷操作（直接编辑数量或删除条目）
