-- =============================================
-- 003_favorites.sql - 菜谱收藏
-- =============================================

CREATE TABLE IF NOT EXISTS recipe_favorites (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id)    ON DELETE CASCADE,
    recipe_id  UUID NOT NULL REFERENCES recipes(id)  ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, recipe_id)
);

CREATE INDEX IF NOT EXISTS idx_recipe_favorites_user ON recipe_favorites(user_id, created_at DESC);
