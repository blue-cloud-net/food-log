-- =============================================
-- 010_inventory.sql - 库存食材（冰箱 / 外面）
--
-- 目标：记录当前拥有的食材，按「存放位置」归类。位置既覆盖冰箱内的具体位置
-- （冷藏室 / 冷冻室 / 变温室 / 保鲜抽屉 / 门架），也覆盖放在外面的位置
-- （室温台面 / 储物柜 / 荫凉通风处）。
--
-- 内容：
--   1) storage_locations  存放位置字典（area = fridge | outside）
--        owner_id IS NULL = 全局预设（只读），否则为该用户私有自定义位置
--   2) inventory_items    库存食材条目（归属用户，引用一个存放位置）
--
-- 设计说明：
--   - location_id 采用 ON DELETE RESTRICT：位置被库存引用时不可删除，
--     由服务层转换为 409，避免误删位置连带清空库存。
--   - images 沿用菜谱 / 餐厅的 JSONB 数组存法。
--   - 预设位置数据见 011_storage_location_catalog.sql。
--   - 全部语句幂等（IF NOT EXISTS / DROP ... IF EXISTS），可重复执行。
-- =============================================

-- ===== 存放位置字典 =====
CREATE TABLE IF NOT EXISTS storage_locations (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID REFERENCES users(id) ON DELETE CASCADE,  -- NULL = 全局预设
    area       VARCHAR(20) NOT NULL,                         -- fridge（冰箱内）/ outside（外面）
    name       VARCHAR(50) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    is_system  BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ck_storage_locations_area CHECK (area IN ('fridge', 'outside')),
    CONSTRAINT uq_storage_locations_owner_area_name UNIQUE NULLS NOT DISTINCT (owner_id, area, name)
);

CREATE INDEX IF NOT EXISTS idx_storage_locations_owner ON storage_locations(owner_id);

-- ===== 库存食材 =====
CREATE TABLE IF NOT EXISTS inventory_items (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES storage_locations(id) ON DELETE RESTRICT,
    name        VARCHAR(100) NOT NULL,
    amount      VARCHAR(30) NOT NULL DEFAULT '',   -- 数量（自由文本，如 2 / 半 / 500）
    unit        VARCHAR(20) NOT NULL DEFAULT '',   -- 单位（如 个 / 克 / 盒）
    category    VARCHAR(20) NOT NULL DEFAULT '',   -- 食材分类（肉禽/水产/蔬菜…，自由文本）
    expire_at   DATE,                              -- 保质期 / 过期日期
    note        VARCHAR(500) NOT NULL DEFAULT '',
    images      JSONB NOT NULL DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_inventory_items_user_location ON inventory_items(user_id, location_id);
CREATE INDEX IF NOT EXISTS idx_inventory_items_user_expire   ON inventory_items(user_id, expire_at);

-- =============================================
-- updated_at 触发器
-- =============================================
DROP TRIGGER IF EXISTS trg_storage_locations_updated ON storage_locations;
CREATE TRIGGER trg_storage_locations_updated
    BEFORE UPDATE ON storage_locations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

DROP TRIGGER IF EXISTS trg_inventory_items_updated ON inventory_items;
CREATE TRIGGER trg_inventory_items_updated
    BEFORE UPDATE ON inventory_items
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
