-- =============================================
-- 003_demo_seed.sql - 演示数据（仅开发模式执行）
--
-- 由原 002_seed.sql + 005_recipe_tags_migrate.sql 第 3 节整合而来。
-- 仅在 APP_ENV=development 时被迁移器执行，生产环境不会加载。
--
-- 内容：
--   1) 演示账号 admin / admin
--   2) 示例菜谱 / 餐厅 / 菜品
--   3) 演示菜谱的标签补齐
--
-- 依赖 001_schema.sql（表结构）与 002_catalog.sql（全局预设标签）。
-- 全部语句幂等，可重复执行。
-- =============================================

-- =============================================
-- 1) 演示账号 admin / admin
-- 密码哈希: $2a$10$6jl7SFSLBZRwC1BPgxThS.JpTsRZe8BHuzj.BRl.HWpNoxMepl7Ze (bcrypt, cost=10)
-- 重新生成方式:
--   python3 -c "import bcrypt; print(bcrypt.hashpw(b'admin', bcrypt.gensalt(rounds=10)).decode().replace('$2b$','$2a$'))"
-- =============================================
INSERT INTO users (id, username, email, password_hash)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'admin', 'admin@example.com',
     '$2a$10$6jl7SFSLBZRwC1BPgxThS.JpTsRZe8BHuzj.BRl.HWpNoxMepl7Ze')
ON CONFLICT (email) DO NOTHING;

-- =============================================
-- 2) 示例菜谱（固定 UUID，便于标签种子引用）
-- =============================================
INSERT INTO recipes (id, user_id, name, description, ingredients, steps, cook_time_minutes, difficulty, rating)
VALUES
    ('22222222-2222-2222-2222-222222222201', '11111111-1111-1111-1111-111111111111', '麻婆豆腐',
     '经典的川菜家常做法，麻辣鲜香超下饭。',
     '[{"name":"嫩豆腐","amount":"1","unit":"盒"},{"name":"牛肉末","amount":"100","unit":"克"},{"name":"豆瓣酱","amount":"2","unit":"勺"}]',
     '[{"order":1,"content":"豆腐切块，焯水去豆腥"},{"order":2,"content":"炒香肉末和豆瓣酱"},{"order":3,"content":"下豆腐小火煮5分钟"},{"order":4,"content":"勾芡撒花椒面出锅"}]',
     20, 'medium', 5),
    ('22222222-2222-2222-2222-222222222202', '11111111-1111-1111-1111-111111111111', '番茄炒蛋',
     '十分钟搞定的快手家常菜。',
     '[{"name":"番茄","amount":"2","unit":"个"},{"name":"鸡蛋","amount":"3","unit":"个"},{"name":"葱花","amount":"少许","unit":""}]',
     '[{"order":1,"content":"鸡蛋打散炒熟盛出"},{"order":2,"content":"番茄炒出汁"},{"order":3,"content":"倒回鸡蛋翻炒调味"}]',
     10, 'easy', 4)
ON CONFLICT (id) DO NOTHING;

-- ===== 示例餐厅 =====
INSERT INTO restaurants (id, user_id, name, address, cuisine_type, description, avg_rating)
VALUES
    ('00000000-0000-0000-0000-000000000011', '11111111-1111-1111-1111-111111111111', '老四川火锅', '建设路 100 号', '川菜',
     '牛油锅底很香，毛肚新鲜。', 4.5),
    ('00000000-0000-0000-0000-000000000012', '11111111-1111-1111-1111-111111111111', '海边渔村', '滨海大道 66 号', '海鲜',
     '食材新鲜，蒸海鲜一绝。', 4.8)
ON CONFLICT (id) DO NOTHING;

-- ===== 示例菜品 =====
INSERT INTO dishes (restaurant_id, user_id, name, description, price, rating, eaten_at)
VALUES
    ('00000000-0000-0000-0000-000000000011', '11111111-1111-1111-1111-111111111111',
     '麻辣毛肚', '爽脆入味，分量足', 48.00, 5, '2026-07-20'),
    ('00000000-0000-0000-0000-000000000011', '11111111-1111-1111-1111-111111111111',
     '冰粉', '解辣神器，红糖味浓', 8.00, 4, '2026-07-20'),
    ('00000000-0000-0000-0000-000000000012', '11111111-1111-1111-1111-111111111111',
     '蒜蓉蒸生蚝', '肥美多汁，蒜香十足', 68.00, 5, '2026-07-25');

-- =============================================
-- 3) 演示菜谱标签（仅当该菜谱当前没有任何标签）
--    麻婆豆腐 → 预设「川菜」+ 自定义「下饭菜」
--    番茄炒蛋 → 自定义「快手菜」「家常」
-- =============================================

-- 由文本生成确定性 UUID（同一输入永远得到同一 id，用于 find-or-create）
-- 使用 pg_temp 会话级函数，不污染业务 schema
CREATE OR REPLACE FUNCTION pg_temp.det_uuid(txt text) RETURNS uuid
LANGUAGE sql IMMUTABLE AS $$
    SELECT (substr(h, 1, 8) || '-' || substr(h, 9, 4) || '-' || substr(h, 13, 4) || '-' ||
            substr(h, 17, 4) || '-' || substr(h, 21, 12))::uuid
    FROM (SELECT md5(txt) AS h) s
$$;

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
        RAISE NOTICE '003: 演示用户不存在，跳过演示标签';
        RETURN;
    END IF;

    -- 演示用户的「自定义」分类
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
        RAISE NOTICE '003: 已为「麻婆豆腐」补齐演示标签';
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
        RAISE NOTICE '003: 已为「番茄炒蛋」补齐演示标签';
    END IF;
END $$;
