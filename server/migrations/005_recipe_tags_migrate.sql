-- =============================================
-- 005_recipe_tags_migrate.sql - 存量标签迁移
--
-- 1) 若 recipes.tags 列存在：把每行 JSONB 数组里的中文标签
--    逐用户转成「自定义」分类下的自定义标签，并写入 recipe_tags
--    （按用户决策：一律新建自定义标签，不与全局预设按名称合并）
-- 2) 删除 recipes.tags 列（GIN 索引随之消失）
-- 3) 为演示菜谱补齐标签（仅当该菜谱当前没有任何标签时）
--
-- 幂等：可重复执行。本地/远程 migrate.sh 每次全量重跑本文件也不会产生重复数据。
-- =============================================

-- 由文本生成确定性 UUID（同一输入永远得到同一 id，用于 find-or-create）
-- 使用 pg_temp 会话级函数，不污染业务 schema
CREATE OR REPLACE FUNCTION pg_temp.det_uuid(txt text) RETURNS uuid
LANGUAGE sql IMMUTABLE AS $$
    SELECT (substr(h, 1, 8) || '-' || substr(h, 9, 4) || '-' || substr(h, 13, 4) || '-' ||
            substr(h, 17, 4) || '-' || substr(h, 21, 12))::uuid
    FROM (SELECT md5(txt) AS h) s
$$;

-- =============================================
-- 1) + 2) 存量 recipes.tags → 自定义标签 + recipe_tags，然后删除列
-- =============================================
DO $$
DECLARE
    col_exists boolean;
BEGIN
    SELECT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'recipes' AND column_name = 'tags'
    ) INTO col_exists;

    IF NOT col_exists THEN
        RAISE NOTICE '005: recipes.tags 列不存在，跳过存量迁移';
        RETURN;
    END IF;

    -- 1.1) 为每个「现有标签的用户」建立（或复用）「自定义」分类
    INSERT INTO tag_categories (id, owner_id, name, color, sort_order, is_system)
    SELECT DISTINCT
           pg_temp.det_uuid(r.user_id::text || '|category|自定义'),
           r.user_id,
           '自定义',
           'info',
           900,
           false
    FROM recipes r
    WHERE jsonb_typeof(r.tags) = 'array'
      AND jsonb_array_length(r.tags) > 0
    ON CONFLICT DO NOTHING;

    -- 1.2) 每个 (用户, 标签名) 建一个自定义标签
    WITH src AS (
        SELECT DISTINCT r.user_id AS uid, btrim(v.name) AS tag_name
        FROM recipes r
        CROSS JOIN LATERAL jsonb_array_elements_text(r.tags) AS v(name)
        WHERE jsonb_typeof(r.tags) = 'array'
          AND btrim(v.name) <> ''
    ),
    numbered AS (
        SELECT uid, tag_name,
               (row_number() OVER (PARTITION BY uid ORDER BY tag_name) - 1) AS ord
        FROM src
    )
    INSERT INTO tags (id, category_id, owner_id, name, mutex_group, sort_order, is_system)
    SELECT pg_temp.det_uuid(uid::text || '|tag|' || tag_name),
           pg_temp.det_uuid(uid::text || '|category|自定义'),
           uid,
           tag_name,
           NULL,
           ord,
           false
    FROM numbered
    ON CONFLICT DO NOTHING;

    -- 1.3) 建立菜谱 ↔ 标签关联
    WITH src AS (
        SELECT DISTINCT r.id AS recipe_id, r.user_id AS uid, btrim(v.name) AS tag_name
        FROM recipes r
        CROSS JOIN LATERAL jsonb_array_elements_text(r.tags) AS v(name)
        WHERE jsonb_typeof(r.tags) = 'array'
          AND btrim(v.name) <> ''
    )
    INSERT INTO recipe_tags (recipe_id, tag_id)
    SELECT recipe_id, pg_temp.det_uuid(uid::text || '|tag|' || tag_name)
    FROM src
    ON CONFLICT DO NOTHING;

    RAISE NOTICE '005: 存量标签迁移完成';

    -- 2) 删除旧列
    EXECUTE 'ALTER TABLE recipes DROP COLUMN IF EXISTS tags';
END $$;

DROP INDEX IF EXISTS idx_recipes_tags;

-- =============================================
-- 3) 演示菜谱标签（仅当该菜谱当前没有任何标签）
--    麻婆豆腐：预设「川菜」(20000000-...-000000000041) + 自定义「下饭菜」
--    番茄炒蛋：自定义「快手菜」「家常」
-- =============================================
DO $$
DECLARE
    demo_user uuid := '11111111-1111-1111-1111-111111111111';
    demo_cat  uuid;
    rec_id    uuid;
    tag_id    uuid;
    tag_name  text;
    tag_sort  int;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM users WHERE id = demo_user) THEN
        RAISE NOTICE '005: 演示用户不存在，跳过演示标签';
        RETURN;
    END IF;

    demo_cat := pg_temp.det_uuid(demo_user::text || '|category|自定义');
    INSERT INTO tag_categories (id, owner_id, name, color, sort_order, is_system)
    VALUES (demo_cat, demo_user, '自定义', 'info', 900, false)
    ON CONFLICT DO NOTHING;

    SELECT id INTO rec_id FROM recipes WHERE user_id = demo_user AND name = '麻婆豆腐' LIMIT 1;
    IF rec_id IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM recipe_tags WHERE recipe_id = rec_id)
       AND NOT EXISTS (SELECT 1 FROM recipe_ingredient_tags WHERE recipe_id = rec_id) THEN
        INSERT INTO recipe_tags (recipe_id, tag_id)
        VALUES (rec_id, '20000000-0000-4000-8000-000000000041')  -- 预设：川菜
        ON CONFLICT DO NOTHING;

        tag_name := '下饭菜';
        tag_id := pg_temp.det_uuid(demo_user::text || '|tag|' || tag_name);
        INSERT INTO tags (id, category_id, owner_id, name, sort_order, is_system)
        VALUES (tag_id, demo_cat, demo_user, tag_name, 100, false)
        ON CONFLICT DO NOTHING;
        INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (rec_id, tag_id) ON CONFLICT DO NOTHING;
        RAISE NOTICE '005: 已为「麻婆豆腐」补齐演示标签';
    END IF;

    SELECT id INTO rec_id FROM recipes WHERE user_id = demo_user AND name = '番茄炒蛋' LIMIT 1;
    IF rec_id IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM recipe_tags WHERE recipe_id = rec_id)
       AND NOT EXISTS (SELECT 1 FROM recipe_ingredient_tags WHERE recipe_id = rec_id) THEN
        FOREACH tag_name IN ARRAY ARRAY['快手菜', '家常']
        LOOP
            tag_sort := CASE tag_name WHEN '快手菜' THEN 101 ELSE 102 END;
            tag_id := pg_temp.det_uuid(demo_user::text || '|tag|' || tag_name);
            INSERT INTO tags (id, category_id, owner_id, name, sort_order, is_system)
            VALUES (tag_id, demo_cat, demo_user, tag_name, tag_sort, false)
            ON CONFLICT DO NOTHING;
            INSERT INTO recipe_tags (recipe_id, tag_id) VALUES (rec_id, tag_id) ON CONFLICT DO NOTHING;
        END LOOP;
        RAISE NOTICE '005: 已为「番茄炒蛋」补齐演示标签';
    END IF;
END $$;
