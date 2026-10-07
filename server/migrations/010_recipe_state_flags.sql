-- =============================================
-- 010_recipe_state_flags.sql - 首页分类所需的状态字段
--
-- 新增列：
--   recipes.made_at        DATE     做过日期（NULL = 未做）
--   recipes.is_liked       BOOLEAN  喜欢（自制，独立于 recipe_favorites 收藏）
--   restaurants.is_visited BOOLEAN  是否已探店（false = 未探店）
--   dishes.is_liked        BOOLEAN  喜欢菜品
--
-- 依赖：
--   001_schema.sql   表结构（recipes / restaurants / dishes）
--
-- 全部语句幂等（ADD COLUMN IF NOT EXISTS / CREATE INDEX IF NOT EXISTS），可重复执行。
-- 注意：演示数据 009_demo_seed.sql 在本文件之前执行，不能引用此处新增的列。
-- =============================================

ALTER TABLE recipes     ADD COLUMN IF NOT EXISTS made_at    DATE;
ALTER TABLE recipes     ADD COLUMN IF NOT EXISTS is_liked   BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS is_visited BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE dishes      ADD COLUMN IF NOT EXISTS is_liked   BOOLEAN NOT NULL DEFAULT false;

-- 已做 / 未做：按 (user_id, made_at) 过滤，NULL 表示未做
CREATE INDEX IF NOT EXISTS idx_recipes_user_made ON recipes(user_id, made_at);
-- 喜欢（自制）：部分索引，仅覆盖已喜欢行
CREATE INDEX IF NOT EXISTS idx_recipes_user_liked ON recipes(user_id) WHERE is_liked;
-- 已探店：部分索引
CREATE INDEX IF NOT EXISTS idx_restaurants_user_visited ON restaurants(user_id) WHERE is_visited;
-- 喜欢菜品：部分索引
CREATE INDEX IF NOT EXISTS idx_dishes_user_liked ON dishes(user_id) WHERE is_liked;
