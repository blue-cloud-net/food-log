-- =============================================
-- 001_schema.sql - 结构初始化（全部表 / 索引 / 触发器）
--
-- 由原 001_init.sql + 003_favorites.sql + 004_tags.sql 的 DDL 整合而来。
-- 全部语句幂等（IF NOT EXISTS / CREATE OR REPLACE），可重复执行。
--
-- 表一览：
--   users                    用户
--   recipes                  自制菜谱
--   restaurants              餐厅
--   dishes                   店内菜品
--   recipe_favorites         菜谱收藏
--   tag_categories           标签分类（owner_id IS NULL = 全局预设）
--   tags                     标签
--   recipe_tags              菜谱 ↔ 标签（用户手选）
--   recipe_ingredient_tags   菜谱 ↔ 标签（自动规则派生）
--   tag_rules                自动标签规则
-- =============================================

-- 扩展：UUID 生成
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ===== updated_at 自动更新触发器函数 =====
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- =============================================
-- 业务表
-- =============================================

-- ===== users 用户表 =====
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(50)  NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    avatar_url    TEXT,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- ===== recipes 自制菜谱表 =====
CREATE TABLE IF NOT EXISTS recipes (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name              VARCHAR(200) NOT NULL,
    description       TEXT,
    ingredients       JSONB NOT NULL DEFAULT '[]',
    steps             JSONB NOT NULL DEFAULT '[]',
    cook_time_minutes INT,
    difficulty        VARCHAR(20) CHECK (difficulty IN ('easy', 'medium', 'hard')),
    rating            SMALLINT CHECK (rating BETWEEN 1 AND 5),
    images            JSONB NOT NULL DEFAULT '[]',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===== restaurants 餐厅表 =====
CREATE TABLE IF NOT EXISTS restaurants (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          VARCHAR(200) NOT NULL,
    address       TEXT,
    cuisine_type  VARCHAR(100),
    description   TEXT,
    avg_rating    NUMERIC(2,1) CHECK (avg_rating BETWEEN 0 AND 5),
    images        JSONB NOT NULL DEFAULT '[]',
    lat           DOUBLE PRECISION,
    lng           DOUBLE PRECISION,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===== dishes 店内菜品表 =====
CREATE TABLE IF NOT EXISTS dishes (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
    user_id       UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name          VARCHAR(200) NOT NULL,
    description   TEXT,
    price         NUMERIC(10,2),
    rating        SMALLINT CHECK (rating BETWEEN 1 AND 5),
    images        JSONB NOT NULL DEFAULT '[]',
    eaten_at      DATE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ===== recipe_favorites 菜谱收藏表 =====
CREATE TABLE IF NOT EXISTS recipe_favorites (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    recipe_id  UUID NOT NULL REFERENCES recipes(id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, recipe_id)
);

-- =============================================
-- 标签字典化
--   1) owner_id IS NULL  → 全局预设（所有用户可见）
--      owner_id = 某用户 → 该用户私有自定义
--   2) 菜谱与标签按 id 关联，分两套：
--      recipe_tags             菜谱级（用户手选）
--      recipe_ingredient_tags  食材级（自动匹配规则派生）
--   3) 全局预设使用固定 UUID，便于迁移 / 文档引用，映射见 docs/tags.md
-- =============================================

-- ===== 标签分类 =====
CREATE TABLE IF NOT EXISTS tag_categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL = 全局预设
    name       VARCHAR(50) NOT NULL,
    color      VARCHAR(20) NOT NULL DEFAULT 'info',          -- el-tag 色型：primary/success/warning/danger/info
    sort_order INT NOT NULL DEFAULT 0,
    is_system  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_tag_categories_owner_name UNIQUE NULLS NOT DISTINCT (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tag_categories_owner ON tag_categories(owner_id);

-- ===== 标签 =====
CREATE TABLE IF NOT EXISTS tags (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_id UUID NOT NULL REFERENCES tag_categories(id) ON DELETE CASCADE,
    owner_id    UUID REFERENCES users(id) ON DELETE CASCADE, -- NULL = 全局预设
    name        VARCHAR(50) NOT NULL,
    mutex_group VARCHAR(32),                                 -- 同组标签互斥（如 diet：荤菜/素菜）
    sort_order  INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    is_system   BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_tags_owner_name UNIQUE NULLS NOT DISTINCT (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category_id);
CREATE INDEX IF NOT EXISTS idx_tags_owner    ON tags(owner_id);
CREATE INDEX IF NOT EXISTS idx_tags_mutex    ON tags(mutex_group);

-- ===== 菜谱 ↔ 标签（菜谱级 / 用户手选） =====
CREATE TABLE IF NOT EXISTS recipe_tags (
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id    UUID NOT NULL REFERENCES tags(id)    ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_recipe_tags_tag ON recipe_tags(tag_id);

-- ===== 菜谱 ↔ 标签（食材级 / 自动派生） =====
CREATE TABLE IF NOT EXISTS recipe_ingredient_tags (
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id    UUID NOT NULL REFERENCES tags(id)    ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_recipe_ingredient_tags_tag ON recipe_ingredient_tags(tag_id);

-- ===== 自动标签规则 =====
-- rule_type 取值：
--   ingredient_keyword       逐个食材名匹配 keywords（含 exclude_keywords）
--   ingredient_text_keyword  食材拼接文本匹配 keywords
--   text_keyword             按 match_field 指定文本匹配 keywords
--   cook_time_max            0 < cook_time_minutes <= max_minutes
--   group_mutex              已命中标签落在 member_tag_ids 内；同 tag_group 只取首个命中
CREATE TABLE IF NOT EXISTS tag_rules (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id         UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL = 全局预设
    tag_id           UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    rule_type        VARCHAR(32) NOT NULL,
    match_field      VARCHAR(32),
    keywords         TEXT[] NOT NULL DEFAULT '{}',
    exclude_keywords TEXT[] NOT NULL DEFAULT '{}',
    max_minutes      INT,
    tag_group        VARCHAR(32),
    member_tag_ids   UUID[] NOT NULL DEFAULT '{}',
    sort_order       INT NOT NULL DEFAULT 0,
    is_active        BOOLEAN NOT NULL DEFAULT true,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_tag_rules_type CHECK (
        rule_type IN ('ingredient_keyword', 'ingredient_text_keyword', 'text_keyword', 'cook_time_max', 'group_mutex')
    ),
    CONSTRAINT ck_tag_rules_field CHECK (
        match_field IS NULL OR match_field IN ('name', 'description', 'name_description')
    )
);

CREATE INDEX IF NOT EXISTS idx_tag_rules_tag    ON tag_rules(tag_id);
CREATE INDEX IF NOT EXISTS idx_tag_rules_active ON tag_rules(is_active, sort_order);

-- =============================================
-- 索引
-- =============================================
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email    ON users(email);

CREATE INDEX IF NOT EXISTS idx_recipes_user_created ON recipes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recipes_name         ON recipes(name);

CREATE INDEX IF NOT EXISTS idx_restaurants_user_created ON restaurants(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_restaurants_cuisine      ON restaurants(cuisine_type);

CREATE INDEX IF NOT EXISTS idx_dishes_restaurant ON dishes(restaurant_id, eaten_at DESC);
CREATE INDEX IF NOT EXISTS idx_dishes_user       ON dishes(user_id);

CREATE INDEX IF NOT EXISTS idx_recipe_favorites_user ON recipe_favorites(user_id, created_at DESC);

-- =============================================
-- updated_at 触发器
-- =============================================
DROP TRIGGER IF EXISTS trg_users_updated ON users;
CREATE TRIGGER trg_users_updated
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_recipes_updated ON recipes;
CREATE TRIGGER trg_recipes_updated
    BEFORE UPDATE ON recipes
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_restaurants_updated ON restaurants;
CREATE TRIGGER trg_restaurants_updated
    BEFORE UPDATE ON restaurants
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tag_categories_updated ON tag_categories;
CREATE TRIGGER trg_tag_categories_updated
    BEFORE UPDATE ON tag_categories
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tags_updated ON tags;
CREATE TRIGGER trg_tags_updated
    BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_tag_rules_updated ON tag_rules;
CREATE TRIGGER trg_tag_rules_updated
    BEFORE UPDATE ON tag_rules
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
