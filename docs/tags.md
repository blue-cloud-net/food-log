# 标签体系

项目有**三套互相独立**的标签字典，互不影响：

| 字典 | 字典表 | 关联表 | 词表接口 |
|---|---|---|---|
| 菜谱标签 | `tag_categories` / `tags` / `tag_rules` | `recipe_tags`（手选）、`recipe_ingredient_tags`（规则派生） | `GET /api/recipes/tags` |
| 餐厅标签 | `restaurant_tag_categories` / `restaurant_tags` | `restaurant_tag_links` | `GET /api/restaurants/tags` |
| 菜品标签 | `dish_tag_categories` / `dish_tags` | `dish_tag_links` | `GET /api/dishes/tags` |

菜谱标签由**数据库字典**驱动，支持「全局预设 + 用户自定义」，并区分**菜谱级（用户手选）**与**食材级（规则自动派生）**两套关联。
探店（餐厅 / 菜品）标签只有**手选**一种，没有自动规则与 AI 兜底。

## 表结构（菜谱标签，`server/migrations/001_schema.sql`）

| 表 | 说明 |
|---|---|
| `tag_categories` | 标签分类。`owner_id IS NULL` = 全局预设；`owner_id = 用户` = 用户自定义。含 `color`（el-tag 色型）、`sort_order`、`is_system` |
| `tags` | 标签。`category_id` 必填；`mutex_group` 用于选择器互斥（如荤菜/素菜同为 `diet`）；`is_system` 标记预设标签 |
| `recipe_tags` | 菜谱 ↔ 标签（**菜谱级**，用户手选） |
| `recipe_ingredient_tags` | 菜谱 ↔ 标签（**食材级**，由规则自动派生） |
| `tag_rules` | 自动标签规则 |

> **不使用短英文 key**：全局预设标签使用固定 UUID（下文表格），关联一律用 id。
> `name` 在同一 owner 下唯一（`UNIQUE NULLS NOT DISTINCT (owner_id, name)`）；
> 全局预设与用户自定义可以同名（例如预设「川菜」与用户自建「川菜」并存）。

## 可见性与权限

- 用户可见的字典 = `owner_id IS NULL`（全局预设） ∪ `owner_id = 当前用户`
- `is_system = true` 的分类/标签：所有用户**只读**，修改返回 403
- 自定义分类/标签：仅创建者可改删；删除分类会**级联删除**其下标签，并解除相关菜谱的关联
- 菜谱只能关联「当前用户可见」的标签，否则创建/更新返回 400

## 探店标签（餐厅 / 菜品）

由 `004_shop_tags.sql` 建立，餐厅与菜品**各一套结构同构的表**，与菜谱标签体系完全隔离。

| 饭店字典 | 表（分类 / 标签 / 关联） | 词表接口 | 管理接口 |
|---|---|---|---|
| 餐厅 | `restaurant_tag_categories` / `restaurant_tags` / `restaurant_tag_links` | `GET /api/restaurants/tags` | `/restaurant-tags`、`/restaurant-tag-categories` |
| 菜品 | `dish_tag_categories` / `dish_tags` / `dish_tag_links` | `GET /api/dishes/tags` | `/dish-tags`、`/dish-tag-categories` |

与菜谱标签相同的约定：

- 可见性：`owner_id IS NULL`（全局预设） ∪ `owner_id = 当前用户`（私有自定义）
- `name` 在同一 owner 下唯一；`is_system = true` 只读
- `mutex_group` 用于选择器互斥（如菜品「份量足 / 份量适中 / 份量少」同为 `portion`）
- 实体只能关联「当前用户可见」的标签，否则创建/更新返回 400（错误码 1001）

不同点：

- **没有自动规则表**（无 `shop_tag_rules`）、**不接 AI 兜底**，完全手选
- 按标签筛选：餐厅列表用 `?tag=<餐厅标签 id>`；餐厅详情内的菜品列表用 `?dish_tag=<菜品标签 id>`
- 全局搜索会命中餐厅 / 菜品标签名

## 自动标签规则（`tag_rules`）

按 `sort_order` 升序求值：

| `rule_type` | 语义 | 用到的列 |
|---|---|---|
| `ingredient_keyword` | 逐个食材名匹配关键词 | `keywords`、`exclude_keywords` |
| `ingredient_text_keyword` | 食材拼接文本包含关键词 | `keywords` |
| `text_keyword` | 按 `match_field` 指定文本匹配 | `match_field`（name/description/name_description）、`keywords` |
| `cook_time_max` | `0 < 烹饪耗时 ≤ max_minutes` | `max_minutes` |
| `group_mutex` | 已命中标签落在 `member_tag_ids` 内则命中本标签；同一 `tag_group` 只取**首个**命中 | `tag_group`、`member_tag_ids` |

- `group_mutex` 必须排在 `ingredient_*` 之后（如荤菜 210 / 素菜 220），用于表达「荤菜优先、素菜其次」
- **去重规则**：自动结果会剔除与用户手选标签重复的项；若用户已手选某互斥组（如素菜）内的标签，自动结果中同组标签（荤菜）一并剔除，避免出现矛盾展示
- 规则识别不足（未命中任何互斥组成员标签）且已配置 AI 时，用 AI 从可见词表中挑选标签兜底；AI 未配置/失败则静默降级
- v1 规则仅由 SQL/迁移维护，暂无管理界面

## AI 图片识别

`POST /api/recipes/ai/recognize` 返回的标签为**标签 id**：AI 给出的名称先按可见词表匹配，未收录的名称会在该用户的「自定义」分类下创建为自定义标签。

## 导出 / 导入

JSON 备份与 CSV 导出中的标签为**标签名称**（而非 id），因此备份可跨数据库导入；导入时按名称重新映射，未收录的名称会创建为用户自定义标签。

## 全局预设分类与标签（固定 UUID）

分类 UUID 形如 `10000000-0000-4000-8000-00000000000N`，标签形如 `20000000-0000-4000-8000-0000000000NN`。

| 分类 (id 尾号) | color | 标签 (id 尾号) |
|---|---|---|
| 荤素 `0001` | danger | 荤菜 `0001`（mutex `diet`）、素菜 `0002`（mutex `diet`） |
| 食材 `0002` | success | 牛肉 `0011`、猪肉 `0012`、鸡肉 `0013`、羊肉 `0014`、鸭肉 `0015`、鱼 `0016`、海鲜 `0017`、虾 `0018`、蟹 `0019`、蛋 `001a`、豆制品 `001b`、蔬菜 `001c`、菌菇 `001d`、主食 `001e` |
| 场景 `0003` | warning | 快手 `0021`、汤 `0022`、凉菜 `0023`、面食 `0024`、甜点 `0025` |
| 时段 `0004` | info | 早餐 `0031`、午餐 `0032`、晚餐 `0033`、夜宵 `0034` |
| 菜系 `0005` | primary | 川菜 `0041`、粤菜 `0042`、湘菜 `0043`、鲁菜 `0044`、苏菜 `0045`、浙菜 `0046`、闽菜 `0047`、徽菜 `0048`、东北菜 `0049`、西北菜 `004a` |
| 口味 `0006` | danger | 辣 `0051`、清淡 `0052`、甜 `0053`、酸 `0054`、咸鲜 `0055` |

规则 UUID 形如 `30000000-0000-4000-8000-0000000000NN`（`0001`~`0016` 共 22 条）。

## 预设词表初始化与演示数据

| 文件 | 内容 | 执行环境 |
|---|---|---|
| `001_schema.sql` | 全部表 / 索引 / 触发器 | 全部 |
| `002_catalog.sql` | 菜谱全局预设分类 / 标签 / 规则 | 全部 |
| `004_shop_tags.sql` | 探店标签表结构 + 餐厅评分维度变更 | 全部 |
| `005_shop_tag_catalog.sql` | 探店全局预设分类 / 标签 | 全部 |
| `009_demo_seed.sql` | 演示账号 + 示例菜谱 / 餐厅 / 菜品 + 标签补充 | 仅 `APP_ENV=development` |

- 全部使用固定 UUID + `ON CONFLICT DO UPDATE`，可重复执行且不会产生重复数据
- 迁移由 `internal/database/migrate.go` 在服务启动时按文件名升序执行，并用 `schema_migrations` 记账，因此**新增迁移无需手动重建数据库**
- `009_demo_seed.sql` 引用了 `004` / `005` 新增的列与表，所以文件名必须排在它们之后（故为 `009` 而非早期的 `003`）；演示数据文件名由 `migrations.DemoSeedFile` 常量指定
- 旧版 `recipes.tags` JSONB 列已废弃，`001_schema.sql` 不再建该列

## 探店全局预设分类与标签（固定 UUID）

| 域 | 分类 (id 尾号) | color | 标签 (id 尾号) |
|---|---|---|---|
| 餐厅 | 品类 `0001` | warning | 西餐 `0001`、中式快餐 `0002`、火锅 `0003`、烧烤 `0004`、日料 `0005`、韩餐 `0006`、东南亚菜 `0007`、面馆 `0008`、小吃 `0009`、甜品店 `000a`、咖啡馆 `000b`、自助餐 `000c` |
| 餐厅 | 菜系 `0002` | primary | 川菜 `0011`、粤菜 `0012`、湘菜 `0013`、鲁菜 `0014`、苏菜 `0015`、浙菜 `0016`、闽菜 `0017`、徽菜 `0018`、东北菜 `0019`、西北菜 `001a` |
| 餐厅 | 场景 `0003` | info | 适合聚餐 `0021`、适合约会 `0022`、一人食 `0023`、外卖友好 `0024`、宵夜 `0025` |
| 菜品 | 口味 `0001` | danger | 辣 `0001`、清淡 `0002`、鲜香 `0003`、重口 `0004`、偏甜 `0005`、偏咸 `0006` |
| 菜品 | 份量 `0002` | success | 份量足 `0011`（mutex `portion`）、份量适中 `0012`（mutex `portion`）、份量少 `0013`（mutex `portion`） |
| 菜品 | 特色 `0003` | warning | 招牌菜 `0021`、隐藏菜单 `0022`、季节限定 `0023` |

UUID 分段：`40000000-` 餐厅分类、`50000000-` 餐厅标签、`60000000-` 菜品分类、`70000000-` 菜品标签（均为 `-0000-4000-8000-` 中段）。
