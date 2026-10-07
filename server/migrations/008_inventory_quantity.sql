-- =============================================
-- 008_inventory_quantity.sql - 库存数量改为数值
--
-- 背景：原 inventory_items.amount 是自由文本（"3" / "少许"），无法支持
-- 「3 个苹果只吃 1 个」这类部分消耗。改为数值列 quantity，单位仍由 unit 承载。
--
-- 迁移策略：
--   1) 新增 quantity NUMERIC(10,2) NOT NULL DEFAULT 1
--   2) 从 amount 回填：抽取其中的数字；抽不到（如「少许」「半」）按 1 计
--      —— 按约定「全部强制数字」，不再保留模糊表述
--   3) 删除 amount 列
--
-- 编号说明：必须排在 009_demo_seed.sql 之前，否则演示数据会引用已被删除的 amount 列。
-- 全部语句幂等（IF NOT EXISTS / IF EXISTS / DROP CONSTRAINT IF EXISTS），可重复执行。
-- =============================================

-- ===== 1) 数值列 =====
ALTER TABLE inventory_items ADD COLUMN IF NOT EXISTS quantity NUMERIC(10,2) NOT NULL DEFAULT 1;

-- ===== 2) 回填（仅当 amount 列仍存在时执行；amount 已在旧库中删除则为无操作） =====
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'inventory_items'
          AND column_name = 'amount'
    ) THEN
        -- 抽取数字；抽不到按 1；下限 0.01 以满足 quantity > 0
        EXECUTE $mig$
            UPDATE inventory_items
               SET quantity = GREATEST(
                       COALESCE(NULLIF(regexp_replace(amount, '[^0-9.]', '', 'g'), ''), '1')::numeric,
                       0.01)
             WHERE amount IS NOT NULL
        $mig$;
    END IF;
END $$;

-- ===== 3) 删除旧文本列 =====
ALTER TABLE inventory_items DROP COLUMN IF EXISTS amount;

-- ===== 4) 数量必须为正 =====
ALTER TABLE inventory_items DROP CONSTRAINT IF EXISTS ck_inventory_items_quantity;
ALTER TABLE inventory_items ADD CONSTRAINT ck_inventory_items_quantity CHECK (quantity > 0);
