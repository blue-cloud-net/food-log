-- =============================================
-- 001_init.sql - 初始建表
-- =============================================

-- 扩展：UUID 生成
CREATE EXTENSION IF NOT EXISTS pgcrypto;

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
    tags              JSONB NOT NULL DEFAULT '[]',
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

-- ===== 索引 =====
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email    ON users(email);

CREATE INDEX IF NOT EXISTS idx_recipes_user_created ON recipes(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_recipes_name         ON recipes(name);
CREATE INDEX IF NOT EXISTS idx_recipes_tags         ON recipes USING GIN (tags);

CREATE INDEX IF NOT EXISTS idx_restaurants_user_created ON restaurants(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_restaurants_cuisine      ON restaurants(cuisine_type);

CREATE INDEX IF NOT EXISTS idx_dishes_restaurant ON dishes(restaurant_id, eaten_at DESC);
CREATE INDEX IF NOT EXISTS idx_dishes_user       ON dishes(user_id);

-- ===== updated_at 自动更新触发器 =====
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

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
