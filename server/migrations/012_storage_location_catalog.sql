-- =============================================
-- 011_storage_location_catalog.sql - 全局预设存放位置
--
-- 开发 / 生产都会执行：属于系统运行必需的基础数据（不含任何演示业务数据）。
-- 依赖 010_inventory.sql 建立的 storage_locations。
--
-- 全部 INSERT 使用固定 UUID + ON CONFLICT (id) DO UPDATE，可重复执行。
--
-- UUID 分段约定（沿用 docs/tags.md）：
--   10000000-...  菜谱标签分类
--   20000000-...  菜谱标签
--   30000000-...  自动标签规则
--   40000000-...  餐厅标签分类
--   50000000-...  餐厅标签
--   60000000-...  菜品标签分类
--   70000000-...  菜品标签
--   80000000-...  存放位置（01-0f 冰箱内，11-1f 外面）
-- =============================================

INSERT INTO storage_locations (id, owner_id, area, name, sort_order, is_system) VALUES
    -- 冰箱内
    ('80000000-0000-4000-8000-000000000001', NULL, 'fridge',  '冷藏室',            10, true),
    ('80000000-0000-4000-8000-000000000002', NULL, 'fridge',  '冷冻室',            20, true),
    ('80000000-0000-4000-8000-000000000003', NULL, 'fridge',  '变温室 / 零度保鲜', 30, true),
    ('80000000-0000-4000-8000-000000000004', NULL, 'fridge',  '保鲜抽屉（果蔬）',   40, true),
    ('80000000-0000-4000-8000-000000000005', NULL, 'fridge',  '冰箱门架',          50, true),
    -- 外面
    ('80000000-0000-4000-8000-000000000011', NULL, 'outside', '室温 / 台面',       10, true),
    ('80000000-0000-4000-8000-000000000012', NULL, 'outside', '储物柜 / 橱柜',     20, true),
    ('80000000-0000-4000-8000-000000000013', NULL, 'outside', '荫凉通风处',        30, true)
ON CONFLICT (id) DO UPDATE SET
    owner_id   = EXCLUDED.owner_id,
    area       = EXCLUDED.area,
    name       = EXCLUDED.name,
    sort_order = EXCLUDED.sort_order,
    is_system  = EXCLUDED.is_system;
