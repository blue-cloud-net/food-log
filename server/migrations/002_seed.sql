-- =============================================
-- 002_seed.sql - 测试种子数据
-- 密码均为: password123 (bcrypt 哈希)
-- =============================================

-- 测试用户
INSERT INTO users (id, username, email, password_hash)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'demo', 'demo@example.com',
     '$2a$10$X7UrE7E9Q5c6hQwGf6mJXeZ0z0z0z0z0z0z0z0z0z0z0z0z0z0z0z0z0'),
    ('00000000-0000-0000-0000-000000000002', '小明', 'ming@example.com',
     '$2a$10$X7UrE7E9Q5c6hQwGf6mJXeZ0z0z0z0z0z0z0z0z0z0z0z0z0z0z0z0z0')
ON CONFLICT (email) DO NOTHING;

-- 示例菜谱
INSERT INTO recipes (user_id, name, description, ingredients, steps, cook_time_minutes, difficulty, rating, tags)
VALUES
    ('00000000-0000-0000-0000-000000000001', '麻婆豆腐',
     '经典的川菜家常做法，麻辣鲜香超下饭。',
     '[{"name":"嫩豆腐","amount":"1","unit":"盒"},{"name":"牛肉末","amount":"100","unit":"克"},{"name":"豆瓣酱","amount":"2","unit":"勺"}]',
     '[{"order":1,"content":"豆腐切块，焯水去豆腥"},{"order":2,"content":"炒香肉末和豆瓣酱"},{"order":3,"content":"下豆腐小火煮5分钟"},{"order":4,"content":"勾芡撒花椒面出锅"}]',
     20, 'medium', 5, '["川菜","下饭菜"]'),
    ('00000000-0000-0000-0000-000000000001', '番茄炒蛋',
     '十分钟搞定的快手家常菜。',
     '[{"name":"番茄","amount":"2","unit":"个"},{"name":"鸡蛋","amount":"3","unit":"个"},{"name":"葱花","amount":"少许","unit":""}]',
     '[{"order":1,"content":"鸡蛋打散炒熟盛出"},{"order":2,"content":"番茄炒出汁"},{"order":3,"content":"倒回鸡蛋翻炒调味"}]',
     10, 'easy', 4, '["快手菜","家常"]');

-- 示例餐厅
INSERT INTO restaurants (user_id, name, address, cuisine_type, description, avg_rating)
VALUES
    ('00000000-0000-0000-0000-000000000001', '老四川火锅', '建设路 100 号', '川菜',
     '牛油锅底很香，毛肚新鲜。', 4.5),
    ('00000000-0000-0000-0000-000000000001', '海边渔村', '滨海大道 66 号', '海鲜',
     '食材新鲜，蒸海鲜一绝。', 4.8);

-- 示例菜品
INSERT INTO dishes (restaurant_id, user_id, name, description, price, rating, eaten_at)
VALUES
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
     '麻辣毛肚', '爽脆入味，分量足', 48.00, 5, '2026-07-20'),
    ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001',
     '冰粉', '解辣神器，红糖味浓', 8.00, 4, '2026-07-20'),
    ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001',
     '蒜蓉蒸生蚝', '肥美多汁，蒜香十足', 68.00, 5, '2026-07-25');
