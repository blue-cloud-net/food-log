-- =============================================
-- 004_shop_tags.sql - 探店标签体系（餐厅 / 菜品各一套）+ 餐厅评分维度
--
-- 背景：菜谱标签体系（tag_categories / tags / tag_rules + recipe_tags /
-- recipe_ingredient_tags）只服务菜谱。探店（餐厅、店内菜品）需要独立的字典与
-- 独立接口，因此各自建一套结构同构的表，与菜谱标签体系互不干扰。
--
-- 内容：
--   1) 餐厅标签：restaurant_tag_categories / restaurant_tags / restaurant_tag_links
--   2) 菜品标签：dish_tag_categories     / dish_tags     / dish_tag_links
--   3) 餐厅评分维度：新增 4 个 1-5 评分列；移除 avg_rating（由「推荐度」取代）
--      与 cuisine_type（由「品类 / 菜系」标签承载）
--
-- 可见性规则与菜谱标签一致：owner_id IS NULL = 全局预设，否则为该用户私有自定义。
-- 全部语句幂等（IF NOT EXISTS / IF EXISTS），可重复执行。
-- 预设标签数据见 005_shop_tag_catalog.sql。
-- =============================================

-- =============================================
-- 1) 餐厅标签字典
-- =============================================

-- ===== 餐厅标签分类 =====
CREATE TABLE IF NOT EXISTS restaurant_tag_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL = 全局预设
    name       VARCHAR(50) NOT NULL,
    color      VARCHAR(20) NOT NULL DEFAULT 'info',          -- el-tag 色型：primary/success/warning/danger/info
    sort_order INT NOT NULL DEFAULT 0,
    is_system  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_restaurant_tag_categories_owner_name UNIQUE NULLS NOT DISTINCT (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_restaurant_tag_categories_owner ON restaurant_tag_categories(owner_id);

-- ===== 餐厅标签 =====
CREATE TABLE IF NOT EXISTS restaurant_tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES restaurant_tag_categories(id) ON DELETE CASCADE,
    owner_id    UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL = 全局预设
    name        VARCHAR(50) NOT NULL,
    mutex_group VARCHAR(32),                                 -- 同组标签互斥（同一组只应选一个）
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    is_system   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_restaurant_tags_owner_name UNIQUE NULLS NOT DISTINCT (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_restaurant_tags_category ON restaurant_tags(category_id);
CREATE INDEX IF NOT EXISTS idx_restaurant_tags_owner    ON restaurant_tags(owner_id);
CREATE INDEX IF NOT EXISTS idx_restaurant_tags_mutex    ON restaurant_tags(mutex_group);

-- ===== 餐厅 ↔ 标签（用户手选） =====
CREATE TABLE IF NOT EXISTS restaurant_tag_links (
    restaurant_id UUID NOT NULL REFERENCES restaurants(id)     ON DELETE CASCADE,
    tag_id        UUID NOT NULL REFERENCES restaurant_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (restaurant_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_restaurant_tag_links_tag ON restaurant_tag_links(tag_id);

-- =============================================
-- 2) 菜品标签字典
-- =============================================

-- ===== 菜品标签分类 =====
CREATE TABLE IF NOT EXISTS dish_tag_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL = 全局预设
    name       VARCHAR(50) NOT NULL,
    color      VARCHAR(20) NOT NULL DEFAULT 'info',
    sort_order INT NOT NULL DEFAULT 0,
    is_system  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_dish_tag_categories_owner_name UNIQUE NULLS NOT DISTINCT (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_dish_tag_categories_owner ON dish_tag_categories(owner_id);

-- ===== 菜品标签 =====
CREATE TABLE IF NOT EXISTS dish_tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES dish_tag_categories(id) ON DELETE CASCADE,
    owner_id    UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL = 全局预设
    name        VARCHAR(50) NOT NULL,
    mutex_group VARCHAR(32),
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    is_system   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_dish_tags_owner_name UNIQUE NULLS NOT DISTINCT (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_dish_tags_category ON dish_tags(category_id);
CREATE INDEX IF NOT EXISTS idx_dish_tags_owner    ON dish_tags(owner_id);
CREATE INDEX IF NOT EXISTS idx_dish_tags_mutex    ON dish_tags(mutex_group);

-- ===== 菜品 ↔ 标签（用户手选） =====
CREATE TABLE IF NOT EXISTS dish_tag_links (
    dish_id UUID NOT NULL REFERENCES dishes(id)     ON DELETE CASCADE,
    tag_id  UUID NOT NULL REFERENCES dish_tags(id)  ON DELETE CASCADE,
    PRIMARY KEY (dish_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_dish_tag_links_tag ON dish_tag_links(tag_id);

-- =============================================
-- 3) 餐厅评分维度
--    avg_rating 由「推荐度」取代，dish 评分不再是餐厅均分；
--    cuisine_type 由「品类 / 菜系」标签承载。
-- =============================================
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS recommend_rating SMALLINT CHECK (recommend_rating BETWEEN 1 AND 5);
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS value_rating     SMALLINT CHECK (value_rating     BETWEEN 1 AND 5);
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS ambience_rating  SMALLINT CHECK (ambience_rating  BETWEEN 1 AND 5);
ALTER TABLE restaurants ADD COLUMN IF NOT EXISTS service_rating   SMALLINT CHECK (service_rating   BETWEEN 1 AND 5);

ALTER TABLE restaurants DROP COLUMN IF EXISTS avg_rating;
ALTER TABLE restaurants DROP COLUMN IF EXISTS cuisine_type;

DROP INDEX IF EXISTS idx_restaurants_cuisine;

-- =============================================
-- updated_at 触发器（仅字典表需要；关联表无 updated_at）
-- =============================================
DROP TRIGGER IF EXISTS trg_restaurant_tag_categories_updated ON restaurant_tag_categories;
CREATE TRIGGER trg_restaurant_tag_categories_updated
    BEFORE UPDATE ON restaurant_tag_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_restaurant_tags_updated ON restaurant_tags;
CREATE TRIGGER trg_restaurant_tags_updated
    BEFORE UPDATE ON restaurant_tags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_dish_tag_categories_updated ON dish_tag_categories;
CREATE TRIGGER trg_dish_tag_categories_updated
    BEFORE UPDATE ON dish_tag_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_dish_tags_updated ON dish_tags;
CREATE TRIGGER trg_dish_tags_updated
    BEFORE UPDATE ON dish_tags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
