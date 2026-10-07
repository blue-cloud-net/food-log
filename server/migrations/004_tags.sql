-- =============================================
-- 004_tags.sql - 菜谱标签字典化
--
-- 设计说明：
--   1) 标签词表由 Go 硬编码改为数据库存储，支持「全局预设 + 用户自定义」
--      - owner_id IS NULL  → 全局预设（所有用户可见）
--      - owner_id = 某用户 → 该用户私有自定义
--   2) 菜谱与标签按 id 关联，且分两套：
--      - recipe_tags             菜谱级（用户手选）
--      - recipe_ingredient_tags  食材级（自动匹配规则派生）
--   3) 不使用短英文 key：全局预设使用固定 UUID，便于 seed / 迁移 / 文档引用
--      固定 UUID 与名称的映射见 docs/tags.md
--   4) 本文件幂等，可重复执行
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

-- ===== updated_at 自动更新触发器（复用 001 的 set_updated_at） =====
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

-- =============================================
-- 全局预设分类（固定 UUID：10000000-0000-4000-8000-0000000000NN）
-- =============================================
INSERT INTO tag_categories (id, owner_id, name, color, sort_order, is_system) VALUES
    ('10000000-0000-4000-8000-000000000001', NULL, '荤素', 'danger',  10, true),
    ('10000000-0000-4000-8000-000000000002', NULL, '食材', 'success', 20, true),
    ('10000000-0000-4000-8000-000000000003', NULL, '场景', 'warning', 30, true),
    ('10000000-0000-4000-8000-000000000004', NULL, '时段', 'info',    40, true),
    ('10000000-0000-4000-8000-000000000005', NULL, '菜系', 'primary', 50, true),
    ('10000000-0000-4000-8000-000000000006', NULL, '口味', 'danger',  60, true)
ON CONFLICT (id) DO UPDATE SET
    owner_id   = EXCLUDED.owner_id,
    name       = EXCLUDED.name,
    color      = EXCLUDED.color,
    sort_order = EXCLUDED.sort_order,
    is_system  = EXCLUDED.is_system;

-- =============================================
-- 全局预设标签（固定 UUID：20000000-0000-4000-8000-0000000000NN）
--   NN 分段：01-0f 荤素，11-1e 食材，21-25 场景，31-34 时段，41-4a 菜系，51-55 口味
-- =============================================
INSERT INTO tags (id, category_id, owner_id, name, mutex_group, sort_order, is_system) VALUES
    -- 荤素（互斥组 diet）
    ('20000000-0000-4000-8000-000000000001', '10000000-0000-4000-8000-000000000001', NULL, '荤菜', 'diet', 0, true),
    ('20000000-0000-4000-8000-000000000002', '10000000-0000-4000-8000-000000000001', NULL, '素菜', 'diet', 1, true),
    -- 食材
    ('20000000-0000-4000-8000-000000000011', '10000000-0000-4000-8000-000000000002', NULL, '牛肉',   NULL, 0,  true),
    ('20000000-0000-4000-8000-000000000012', '10000000-0000-4000-8000-000000000002', NULL, '猪肉',   NULL, 1,  true),
    ('20000000-0000-4000-8000-000000000013', '10000000-0000-4000-8000-000000000002', NULL, '鸡肉',   NULL, 2,  true),
    ('20000000-0000-4000-8000-000000000014', '10000000-0000-4000-8000-000000000002', NULL, '羊肉',   NULL, 3,  true),
    ('20000000-0000-4000-8000-000000000015', '10000000-0000-4000-8000-000000000002', NULL, '鸭肉',   NULL, 4,  true),
    ('20000000-0000-4000-8000-000000000016', '10000000-0000-4000-8000-000000000002', NULL, '鱼',     NULL, 5,  true),
    ('20000000-0000-4000-8000-000000000017', '10000000-0000-4000-8000-000000000002', NULL, '海鲜',   NULL, 6,  true),
    ('20000000-0000-4000-8000-000000000018', '10000000-0000-4000-8000-000000000002', NULL, '虾',     NULL, 7,  true),
    ('20000000-0000-4000-8000-000000000019', '10000000-0000-4000-8000-000000000002', NULL, '蟹',     NULL, 8,  true),
    ('20000000-0000-4000-8000-00000000001a', '10000000-0000-4000-8000-000000000002', NULL, '蛋',     NULL, 9,  true),
    ('20000000-0000-4000-8000-00000000001b', '10000000-0000-4000-8000-000000000002', NULL, '豆制品', NULL, 10, true),
    ('20000000-0000-4000-8000-00000000001c', '10000000-0000-4000-8000-000000000002', NULL, '蔬菜',   NULL, 11, true),
    ('20000000-0000-4000-8000-00000000001d', '10000000-0000-4000-8000-000000000002', NULL, '菌菇',   NULL, 12, true),
    ('20000000-0000-4000-8000-00000000001e', '10000000-0000-4000-8000-000000000002', NULL, '主食',   NULL, 13, true),
    -- 场景
    ('20000000-0000-4000-8000-000000000021', '10000000-0000-4000-8000-000000000003', NULL, '快手', NULL, 0, true),
    ('20000000-0000-4000-8000-000000000022', '10000000-0000-4000-8000-000000000003', NULL, '汤',   NULL, 1, true),
    ('20000000-0000-4000-8000-000000000023', '10000000-0000-4000-8000-000000000003', NULL, '凉菜', NULL, 2, true),
    ('20000000-0000-4000-8000-000000000024', '10000000-0000-4000-8000-000000000003', NULL, '面食', NULL, 3, true),
    ('20000000-0000-4000-8000-000000000025', '10000000-0000-4000-8000-000000000003', NULL, '甜点', NULL, 4, true),
    -- 时段
    ('20000000-0000-4000-8000-000000000031', '10000000-0000-4000-8000-000000000004', NULL, '早餐', NULL, 0, true),
    ('20000000-0000-4000-8000-000000000032', '10000000-0000-4000-8000-000000000004', NULL, '午餐', NULL, 1, true),
    ('20000000-0000-4000-8000-000000000033', '10000000-0000-4000-8000-000000000004', NULL, '晚餐', NULL, 2, true),
    ('20000000-0000-4000-8000-000000000034', '10000000-0000-4000-8000-000000000004', NULL, '夜宵', NULL, 3, true),
    -- 菜系
    ('20000000-0000-4000-8000-000000000041', '10000000-0000-4000-8000-000000000005', NULL, '川菜',   NULL, 0,  true),
    ('20000000-0000-4000-8000-000000000042', '10000000-0000-4000-8000-000000000005', NULL, '粤菜',   NULL, 1,  true),
    ('20000000-0000-4000-8000-000000000043', '10000000-0000-4000-8000-000000000005', NULL, '湘菜',   NULL, 2,  true),
    ('20000000-0000-4000-8000-000000000044', '10000000-0000-4000-8000-000000000005', NULL, '鲁菜',   NULL, 3,  true),
    ('20000000-0000-4000-8000-000000000045', '10000000-0000-4000-8000-000000000005', NULL, '苏菜',   NULL, 4,  true),
    ('20000000-0000-4000-8000-000000000046', '10000000-0000-4000-8000-000000000005', NULL, '浙菜',   NULL, 5,  true),
    ('20000000-0000-4000-8000-000000000047', '10000000-0000-4000-8000-000000000005', NULL, '闽菜',   NULL, 6,  true),
    ('20000000-0000-4000-8000-000000000048', '10000000-0000-4000-8000-000000000005', NULL, '徽菜',   NULL, 7,  true),
    ('20000000-0000-4000-8000-000000000049', '10000000-0000-4000-8000-000000000005', NULL, '东北菜', NULL, 8,  true),
    ('20000000-0000-4000-8000-00000000004a', '10000000-0000-4000-8000-000000000005', NULL, '西北菜', NULL, 9,  true),
    -- 口味
    ('20000000-0000-4000-8000-000000000051', '10000000-0000-4000-8000-000000000006', NULL, '辣',   NULL, 0, true),
    ('20000000-0000-4000-8000-000000000052', '10000000-0000-4000-8000-000000000006', NULL, '清淡', NULL, 1, true),
    ('20000000-0000-4000-8000-000000000053', '10000000-0000-4000-8000-000000000006', NULL, '甜',   NULL, 2, true),
    ('20000000-0000-4000-8000-000000000054', '10000000-0000-4000-8000-000000000006', NULL, '酸',   NULL, 3, true),
    ('20000000-0000-4000-8000-000000000055', '10000000-0000-4000-8000-000000000006', NULL, '咸鲜', NULL, 4, true)
ON CONFLICT (id) DO UPDATE SET
    category_id = EXCLUDED.category_id,
    owner_id    = EXCLUDED.owner_id,
    name        = EXCLUDED.name,
    mutex_group = EXCLUDED.mutex_group,
    sort_order  = EXCLUDED.sort_order,
    is_system   = EXCLUDED.is_system;

-- =============================================
-- 全局预设自动标签规则（固定 UUID：30000000-0000-4000-8000-0000000000NN）
-- 求值顺序按 sort_order 升序；group_mutex 必须排在 ingredient_* 之后
-- =============================================

-- 1) 食材关键词 → 食材标签
INSERT INTO tag_rules (id, owner_id, tag_id, rule_type, keywords, exclude_keywords, sort_order, is_active) VALUES
    ('30000000-0000-4000-8000-000000000001', NULL, '20000000-0000-4000-8000-000000000011', 'ingredient_keyword',
     ARRAY['牛肉','牛腩','牛腱','牛排','肥牛','牛里脊'], '{}', 10, true),
    ('30000000-0000-4000-8000-000000000002', NULL, '20000000-0000-4000-8000-000000000012', 'ingredient_keyword',
     ARRAY['五花肉','猪里脊','猪蹄','猪肝','猪肚','培根','火腿','香肠','腊肉','猪肉'], '{}', 20, true),
    ('30000000-0000-4000-8000-000000000003', NULL, '20000000-0000-4000-8000-000000000014', 'ingredient_keyword',
     ARRAY['羊肉','羊排','羊蝎子','肥羊'], '{}', 30, true),
    ('30000000-0000-4000-8000-000000000004', NULL, '20000000-0000-4000-8000-000000000013', 'ingredient_keyword',
     ARRAY['鸡腿','鸡翅','鸡胸','鸡爪','鸡肉','三黄鸡','土鸡','乌鸡'], '{}', 40, true),
    ('30000000-0000-4000-8000-000000000005', NULL, '20000000-0000-4000-8000-000000000015', 'ingredient_keyword',
     ARRAY['鸭腿','鸭肉','烤鸭','盐水鸭'], '{}', 50, true),
    ('30000000-0000-4000-8000-000000000006', NULL, '20000000-0000-4000-8000-000000000016', 'ingredient_keyword',
     ARRAY['三文鱼','鲈鱼','草鱼','鲫鱼','带鱼','鳕鱼','鲤鱼','罗非鱼','多宝鱼','黄花鱼','鱼丸','鱼','鱼片'],
     ARRAY['鱼香'], 60, true),
    ('30000000-0000-4000-8000-000000000007', NULL, '20000000-0000-4000-8000-000000000018', 'ingredient_keyword',
     ARRAY['基围虾','大虾','虾仁','虾皮','小龙虾','河虾','虾滑','虾'], '{}', 70, true),
    ('30000000-0000-4000-8000-000000000008', NULL, '20000000-0000-4000-8000-000000000019', 'ingredient_keyword',
     ARRAY['大闸蟹','梭子蟹','螃蟹','蟹'], '{}', 80, true),
    ('30000000-0000-4000-8000-000000000009', NULL, '20000000-0000-4000-8000-000000000017', 'ingredient_keyword',
     ARRAY['鱿鱼','章鱼','扇贝','蛤蜊','花甲','生蚝','牡蛎','鲍鱼','海参','海螺','蛏子','青口','海胆'], '{}', 90, true),
    ('30000000-0000-4000-8000-00000000000a', NULL, '20000000-0000-4000-8000-00000000001a', 'ingredient_keyword',
     ARRAY['鹌鹑蛋','皮蛋','咸蛋','鸭蛋','鸡蛋','蛋'], '{}', 100, true),
    ('30000000-0000-4000-8000-00000000000b', NULL, '20000000-0000-4000-8000-00000000001b', 'ingredient_keyword',
     ARRAY['油豆腐','豆腐干','豆腐','豆干','豆皮','腐竹','千张','豆花','豆浆','素鸡'], '{}', 110, true),
    ('30000000-0000-4000-8000-00000000000c', NULL, '20000000-0000-4000-8000-00000000001d', 'ingredient_keyword',
     ARRAY['金针菇','杏鲍菇','茶树菇','海鲜菇','香菇','蘑菇','平菇','木耳','银耳'], '{}', 120, true),
    ('30000000-0000-4000-8000-00000000000d', NULL, '20000000-0000-4000-8000-00000000001c', 'ingredient_keyword',
     ARRAY['西兰花','花菜','油麦菜','空心菜','生菜','菠菜','青菜','小白菜','白菜','包菜','卷心菜','番茄','西红柿',
           '土豆','马铃薯','胡萝卜','白萝卜','黄瓜','茄子','青椒','彩椒','尖椒','洋葱','冬瓜','丝瓜','苦瓜','南瓜',
           '莲藕','山药','芋头','芹菜','韭菜','豆角','荷兰豆','玉米','竹笋','莴笋','苋菜','蒜薹','秋葵','紫甘蓝',
           '芦笋','西葫芦'], '{}', 130, true),
    ('30000000-0000-4000-8000-00000000000e', NULL, '20000000-0000-4000-8000-00000000001e', 'ingredient_keyword',
     ARRAY['米饭','炒饭','粥','馒头','包子','饺子','馄饨','云吞','面条','挂面','米粉','河粉','年糕','意面','饼'],
     '{}', 140, true)
ON CONFLICT (id) DO UPDATE SET
    tag_id           = EXCLUDED.tag_id,
    rule_type        = EXCLUDED.rule_type,
    keywords         = EXCLUDED.keywords,
    exclude_keywords = EXCLUDED.exclude_keywords,
    sort_order       = EXCLUDED.sort_order,
    is_active        = EXCLUDED.is_active;

-- 2) 食材文本 / 菜名文本 / 耗时 → 场景与口味标签
INSERT INTO tag_rules (id, owner_id, tag_id, rule_type, match_field, keywords, exclude_keywords, max_minutes, sort_order, is_active) VALUES
    ('30000000-0000-4000-8000-00000000000f', NULL, '20000000-0000-4000-8000-000000000051', 'ingredient_text_keyword', NULL,
     ARRAY['辣椒','小米辣','尖椒','螺丝椒','花椒','藤椒','豆瓣','剁椒','泡椒','辣椒粉','辣椒面','油泼辣子'], '{}', NULL, 150, true),
    ('30000000-0000-4000-8000-000000000010', NULL, '20000000-0000-4000-8000-000000000024', 'ingredient_text_keyword', NULL,
     ARRAY['面条','挂面','米粉','河粉','意面','凉面','炒面','拉面','馒头','包子','饺子','馄饨','云吞','饼'], '{}', NULL, 160, true),
    ('30000000-0000-4000-8000-000000000011', NULL, '20000000-0000-4000-8000-000000000022', 'text_keyword', 'name_description',
     ARRAY['汤'], '{}', NULL, 170, true),
    ('30000000-0000-4000-8000-000000000012', NULL, '20000000-0000-4000-8000-000000000023', 'text_keyword', 'name',
     ARRAY['凉拌','凉菜'], '{}', NULL, 180, true),
    ('30000000-0000-4000-8000-000000000013', NULL, '20000000-0000-4000-8000-000000000025', 'text_keyword', 'name_description',
     ARRAY['蛋糕','甜点','布丁','冰淇淋','雪糕','糖水','曲奇','饼干'], '{}', NULL, 190, true),
    ('30000000-0000-4000-8000-000000000014', NULL, '20000000-0000-4000-8000-000000000021', 'cook_time_max', NULL,
     '{}', '{}', 15, 200, true)
ON CONFLICT (id) DO UPDATE SET
    tag_id           = EXCLUDED.tag_id,
    rule_type        = EXCLUDED.rule_type,
    match_field      = EXCLUDED.match_field,
    keywords         = EXCLUDED.keywords,
    exclude_keywords = EXCLUDED.exclude_keywords,
    max_minutes      = EXCLUDED.max_minutes,
    sort_order       = EXCLUDED.sort_order,
    is_active        = EXCLUDED.is_active;

-- 3) 荤素互斥（荤菜优先：sort_order 更小；同 tag_group 只取首个命中）
INSERT INTO tag_rules (id, owner_id, tag_id, rule_type, tag_group, member_tag_ids, sort_order, is_active) VALUES
    ('30000000-0000-4000-8000-000000000015', NULL, '20000000-0000-4000-8000-000000000001', 'group_mutex', 'diet',
     ARRAY['20000000-0000-4000-8000-000000000011','20000000-0000-4000-8000-000000000012',
           '20000000-0000-4000-8000-000000000013','20000000-0000-4000-8000-000000000014',
           '20000000-0000-4000-8000-000000000015','20000000-0000-4000-8000-000000000016',
           '20000000-0000-4000-8000-000000000017','20000000-0000-4000-8000-000000000018',
           '20000000-0000-4000-8000-000000000019','20000000-0000-4000-8000-00000000001a']::uuid[],
     210, true),
    ('30000000-0000-4000-8000-000000000016', NULL, '20000000-0000-4000-8000-000000000002', 'group_mutex', 'diet',
     ARRAY['20000000-0000-4000-8000-00000000001b','20000000-0000-4000-8000-00000000001c',
           '20000000-0000-4000-8000-00000000001d']::uuid[],
     220, true)
ON CONFLICT (id) DO UPDATE SET
    tag_id         = EXCLUDED.tag_id,
    rule_type      = EXCLUDED.rule_type,
    tag_group      = EXCLUDED.tag_group,
    member_tag_ids = EXCLUDED.member_tag_ids,
    sort_order     = EXCLUDED.sort_order,
    is_active      = EXCLUDED.is_active;
