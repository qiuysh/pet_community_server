-- ========================================
-- [增量·开发种子] Phase1/2/5 测试数据合集
-- 路径: server/sql/incremental/20260724100000_seed_dev_data.sql
-- 适用: 在 init/create.sql 之后执行（开发 / E2E）
-- 内容: 原 02_seed_dev_phase1 + 03_seed_dev_phase2 + 04_seed_dev_phase5
-- 创建日期: 2026-07-24
-- 说明:
--   1. 所有 INSERT 均为幂等写法：目标表已有数据时跳过，可重复执行
--   2. 跨表外键不再硬编码主键，改用 @uN / @petN / @postN 变量按业务键解析
--   3. pets.breed_id 用 0 表示未选品种（与 schema_fix / model.Pet 哨兵语义一致）
--   4. comments.parent_id 先置 0，插入完成后按评论内容回填父子关系
--   5. 不引用死表（pet_photos / pet_records / topic_follows / notifications / feedback / user_login_logs）
-- ========================================

USE pet_community;

-- ========================================
-- Phase1：测试用户
-- ========================================
-- 密码统一为: user123456 (BCrypt cost=10，已与 golang.org/x/crypto/bcrypt 校验)
-- status: 0=未激活/禁用, 1=正常, 2=封禁（登录要求 status=1）
INSERT INTO users (phone, email, password, nickname, avatar, signature, gender, location, pet_years, status, privacy_settings, created_at, updated_at) VALUES
('13800138001', 'user1@petcommunity.com', '$2a$10$oa7nYcQSArcz1ZdlWtS9tOML3p5M8slt3d3fynN9hm71z/mwacNuy', '萌宠小达人', '', '家有两猫一狗，生活乐无忧~', 2, '北京市朝阳区', '3-5年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138002', 'user2@petcommunity.com', '$2a$10$oa7nYcQSArcz1ZdlWtS9tOML3p5M8slt3d3fynN9hm71z/mwacNuy', '铲屎官小王', '', '每天都是被猫咪治愈的一天', 1, '上海市浦东新区', '1-3年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138003', 'user3@petcommunity.com', '$2a$10$oa7nYcQSArcz1ZdlWtS9tOML3p5M8slt3d3fynN9hm71z/mwacNuy', '金毛妈妈', '', '我家金毛叫旺财，今年3岁啦', 2, '广州市天河区', '3-5年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138004', 'user4@petcommunity.com', '$2a$10$oa7nYcQSArcz1ZdlWtS9tOML3p5M8slt3d3fynN9hm71z/mwacNuy', '猫奴一枚', '', '三只猫的铲屎官，痛并快乐着', 0, '深圳市南山区', '5年以上', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138005', 'user5@petcommunity.com', '$2a$10$oa7nYcQSArcz1ZdlWtS9tOML3p5M8slt3d3fynN9hm71z/mwacNuy', '新手铲屎官', '', '刚养了一只小柯基，求指导~', 1, '杭州市西湖区', '0-1年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW())
ON DUPLICATE KEY UPDATE password = VALUES(password), status = VALUES(status), updated_at = NOW();

-- 用户 ID 映射（后续外键引用）
SET @u1 := (SELECT id FROM users WHERE phone = '13800138001' LIMIT 1);
SET @u2 := (SELECT id FROM users WHERE phone = '13800138002' LIMIT 1);
SET @u3 := (SELECT id FROM users WHERE phone = '13800138003' LIMIT 1);
SET @u4 := (SELECT id FROM users WHERE phone = '13800138004' LIMIT 1);
SET @u5 := (SELECT id FROM users WHERE phone = '13800138005' LIMIT 1);

-- ========================================
-- Phase1：测试宠物
-- ========================================
INSERT INTO pets (user_id, name, avatar, pet_type, breed_id, breed_name, gender, weight, personality, bio, is_sterilized, is_vaccinated, status, view_count, created_at, updated_at)
SELECT * FROM (
    SELECT @u1 AS user_id, '橘座' AS name, '' AS avatar, 2 AS pet_type, 0 AS breed_id, '橘猫' AS breed_name, 1 AS gender, 5.5 AS weight, '吃货、懒、爱撒娇' AS personality, '一只胖橘，每天除了吃就是睡，偶尔卖萌求抚摸。' AS bio, 1 AS is_sterilized, 1 AS is_vaccinated, 1 AS status, 128 AS view_count, NOW() AS created_at, NOW() AS updated_at
    UNION ALL SELECT @u1, '小黑', '', 2, 0, '狸花猫', 2, 3.2, '活泼、好动、爱探险', '家里的小调皮，上蹿下停不下来，但是很亲人。', 0, 1, 1, 86, NOW(), NOW()
    UNION ALL SELECT @u1, '豆豆', '', 1, 0, '柯基', 1, 10.0, '憨厚、可爱、小短腿', '柯基豆豆，屁股扭扭超可爱，最爱出去玩。', 1, 1, 1, 256, NOW(), NOW()
    UNION ALL SELECT @u2, '咪咪', '', 2, 0, '英国短毛猫', 2, 4.0, '高冷、优雅、爱干净', '英短蓝猫咪咪，高贵冷艳的小公主，只在吃饭时才理人。', 1, 1, 1, 168, NOW(), NOW()
    UNION ALL SELECT @u3, '旺财', '', 1, 0, '金毛寻回犬', 1, 28.5, '温顺、聪明、大暖男', '金毛旺财，超级大暖男，见谁都摇尾巴，是小区里的明星狗狗。', 1, 1, 1, 312, NOW(), NOW()
    UNION ALL SELECT @u4, '布偶公主', '', 2, 0, '布偶猫', 2, 4.8, '粘人、温柔、仙女猫', '布偶猫小仙女，温柔粘人，抱起来就不想撒手。', 1, 1, 1, 234, NOW(), NOW()
    UNION ALL SELECT @u4, '暹罗王子', '', 2, 0, '暹罗猫', 1, 3.5, '话痨、聪明、好动', '暹罗猫话痨本痨，每天喵喵喵说个不停，但是很聪明。', 0, 1, 1, 145, NOW(), NOW()
    UNION ALL SELECT @u4, '加菲', '', 2, 0, '加菲猫', 1, 5.0, '懒、爱吃、爱睡觉', '加菲猫懒癌晚期，能躺着绝不坐着，能坐着绝不站着。', 1, 1, 1, 178, NOW(), NOW()
    UNION ALL SELECT @u5, '小短腿', '', 1, 0, '柯基', 2, 8.5, '活泼、可爱、贪吃', '刚接回家的小柯基，每天活力满满，什么都想吃。', 0, 1, 1, 98, NOW(), NOW()
) AS seed
WHERE NOT EXISTS (SELECT 1 FROM pets LIMIT 1);

SET @pet1 := (SELECT id FROM pets WHERE name = '橘座' AND user_id = @u1 LIMIT 1);
SET @pet2 := (SELECT id FROM pets WHERE name = '小黑' AND user_id = @u1 LIMIT 1);
SET @pet3 := (SELECT id FROM pets WHERE name = '豆豆' AND user_id = @u1 LIMIT 1);
SET @pet5 := (SELECT id FROM pets WHERE name = '旺财' AND user_id = @u3 LIMIT 1);

-- ========================================
-- Phase1：宠物相册示例
-- ========================================
INSERT INTO pet_albums (pet_id, image_url, thumb_url, title, sort_order, created_at)
SELECT * FROM (
    SELECT @pet1 AS pet_id, '' AS image_url, '' AS thumb_url, '日常照' AS title, 1 AS sort_order, NOW() AS created_at
    UNION ALL SELECT @pet1, '', '', '睡觉照', 2, NOW()
    UNION ALL SELECT @pet1, '', '', '吃饭照', 3, NOW()
    UNION ALL SELECT @pet2, '', '', '日常照', 1, NOW()
    UNION ALL SELECT @pet3, '', '', '出去玩', 1, NOW()
    UNION ALL SELECT @pet3, '', '', '海边', 2, NOW()
    UNION ALL SELECT @pet5, '', '', '小时候', 1, NOW()
    UNION ALL SELECT @pet5, '', '', '现在', 2, NOW()
    UNION ALL SELECT @pet5, '', '', '和小伙伴', 3, NOW()
) AS seed
WHERE @pet1 IS NOT NULL AND NOT EXISTS (SELECT 1 FROM pet_albums LIMIT 1);

-- ========================================
-- Phase2：话题 / 动态 / 评论
-- ========================================
INSERT INTO topics (name, description, cover, posts_count, status, created_at, updated_at) VALUES
('宠物求助', '宠物相关问题求助', '', 0, 1, NOW(), NOW()),
('晒宠大赛', '晒出你家最可爱的宠物', '', 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- topic_ids 引用 create.sql 基础话题（萌宠日常=1…宠物趣事=5）；pet_ids 用 @petN 解析
SET @t_daily := (SELECT id FROM topics WHERE name = '萌宠日常' LIMIT 1);
SET @t_knowledge := (SELECT id FROM topics WHERE name = '养宠知识' LIMIT 1);
SET @t_food := (SELECT id FROM topics WHERE name = '宠物美食' LIMIT 1);
SET @t_health := (SELECT id FROM topics WHERE name = '宠物健康' LIMIT 1);
SET @t_fun := (SELECT id FROM topics WHERE name = '宠物趣事' LIMIT 1);

INSERT INTO posts (user_id, content, images, video, pet_ids, topic_ids, location, visibility, status, likes_count, comments_count, shares_count, favorites_count, is_pinned, is_featured, created_at, updated_at)
SELECT * FROM (
    SELECT @u1 AS user_id, '我家猫咪今天早上偷了我的早餐，太调皮了！' AS content, '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20cat%20eating%20bread%20funny%20moment&image_size=portrait_4_3"]' AS images, '' AS video, CONCAT('[', @pet1, ']') AS pet_ids, CONCAT('[', @t_daily, ']') AS topic_ids, '北京市朝阳区' AS location, 0 AS visibility, 1 AS status, 12 AS likes_count, 5 AS comments_count, 3 AS shares_count, 8 AS favorites_count, 0 AS is_pinned, 0 AS is_featured, NOW() AS created_at, NOW() AS updated_at
    UNION ALL SELECT @u1, '分享一下我家狗狗的训练成果，学会了好多新技能！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=smart%20dog%20training%20obedience&image_size=portrait_4_3"]', '', CONCAT('[', @pet3, ']'), CONCAT('[', @t_knowledge, ']'), '北京市海淀区', 0, 1, 28, 12, 8, 15, 0, 1, NOW(), NOW()
    UNION ALL SELECT @u2, '求助！我家猫咪最近不爱吃东西怎么办？', '', '', '', CONCAT('[', @t_health, ']'), '上海市浦东新区', 0, 1, 8, 23, 2, 3, 0, 0, NOW(), NOW()
    UNION ALL SELECT @u2, '自制宠物零食教程，简单又健康！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=homemade%20pet%20treats%20healthy%20snacks&image_size=portrait_4_3"]', '', '', CONCAT('[', @t_food, ']'), '上海市静安区', 0, 1, 45, 18, 12, 25, 0, 1, NOW(), NOW()
    UNION ALL SELECT @u3, '带狗狗去公园玩的一天，它开心极了！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=happy%20dog%20playing%20in%20park&image_size=portrait_4_3"]', '', CONCAT('[', @pet5, ']'), CONCAT('[', @t_daily, ',', @t_fun, ']'), '广州市天河区', 0, 1, 35, 9, 6, 12, 0, 0, NOW(), NOW()
    UNION ALL SELECT @u3, '给大家看看我家的新成员，一只可爱的小奶猫！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=adorable%20kitten%20new%20pet&image_size=portrait_4_3"]', '', '', CONCAT('[', @t_daily, ']'), '广州市越秀区', 0, 1, 67, 32, 15, 38, 1, 1, NOW(), NOW()
    UNION ALL SELECT @u4, '宠物驱虫知识分享，新手必看！', '', '', '', CONCAT('[', @t_knowledge, ',', @t_health, ']'), '深圳市南山区', 0, 1, 89, 45, 22, 56, 0, 1, NOW(), NOW()
    UNION ALL SELECT @u4, '我家狗狗的搞笑瞬间合集，看一次笑一次！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=funny%20dog%20compilation%20cute&image_size=portrait_4_3"]', '', '', CONCAT('[', @t_fun, ']'), '深圳市福田区', 0, 1, 120, 58, 35, 72, 0, 0, NOW(), NOW()
) AS seed
WHERE NOT EXISTS (SELECT 1 FROM posts LIMIT 1);

SET @post1 := (SELECT id FROM posts WHERE content = '我家猫咪今天早上偷了我的早餐，太调皮了！' LIMIT 1);
SET @post2 := (SELECT id FROM posts WHERE content = '分享一下我家狗狗的训练成果，学会了好多新技能！' LIMIT 1);
SET @post3 := (SELECT id FROM posts WHERE content = '求助！我家猫咪最近不爱吃东西怎么办？' LIMIT 1);
SET @post4 := (SELECT id FROM posts WHERE content = '自制宠物零食教程，简单又健康！' LIMIT 1);
SET @post5 := (SELECT id FROM posts WHERE content = '带狗狗去公园玩的一天，它开心极了！' LIMIT 1);
SET @post6 := (SELECT id FROM posts WHERE content = '给大家看看我家的新成员，一只可爱的小奶猫！' LIMIT 1);
SET @post7 := (SELECT id FROM posts WHERE content = '宠物驱虫知识分享，新手必看！' LIMIT 1);
SET @post8 := (SELECT id FROM posts WHERE content = '我家狗狗的搞笑瞬间合集，看一次笑一次！' LIMIT 1);

INSERT INTO comments (post_id, user_id, parent_id, reply_to_id, content, at_users, status, likes_count, created_at, updated_at)
SELECT * FROM (
    SELECT @post1 AS post_id, @u2 AS user_id, 0 AS parent_id, 0 AS reply_to_id, '哈哈太可爱了！我家猫也经常这样' AS content, '' AS at_users, 1 AS status, 3 AS likes_count, NOW() AS created_at, NOW() AS updated_at
    UNION ALL SELECT @post1, @u3, 0, @u2, '同感！猫咪都是小吃货', '', 1, 1, NOW(), NOW()
    UNION ALL SELECT @post2, @u4, 0, 0, '太棒了！请问是什么品种的狗狗？', '', 1, 5, NOW(), NOW()
    UNION ALL SELECT @post2, @u1, 0, @u4, '是金毛哦，特别聪明', '', 1, 2, NOW(), NOW()
    UNION ALL SELECT @post3, @u1, 0, 0, '是不是天气太热了？试试换点清淡的食物', '', 1, 8, NOW(), NOW()
    UNION ALL SELECT @post3, @u4, 0, @u1, '或者看看有没有口腔问题', '', 1, 4, NOW(), NOW()
    UNION ALL SELECT @post4, @u3, 0, 0, '看起来很好吃！收藏了', '', 1, 6, NOW(), NOW()
    UNION ALL SELECT @post5, @u2, 0, 0, '好羡慕！我们这边公园不让狗狗进', '', 1, 4, NOW(), NOW()
    UNION ALL SELECT @post6, @u1, 0, 0, '太可爱了！叫什么名字？', '', 1, 12, NOW(), NOW()
    UNION ALL SELECT @post6, @u3, 0, @u1, '叫咪咪，现在才两个月大', '', 1, 3, NOW(), NOW()
    UNION ALL SELECT @post7, @u2, 0, 0, '非常实用！感谢分享', '', 1, 15, NOW(), NOW()
    UNION ALL SELECT @post8, @u4, 0, 0, '笑死我了！', '', 1, 23, NOW(), NOW()
) AS seed
WHERE @post1 IS NOT NULL AND NOT EXISTS (SELECT 1 FROM comments LIMIT 1);

-- 回复关系（父评论 ID 在插入前无法确定，插入后按内容 + 同帖回填）
UPDATE comments c
    JOIN comments p ON p.content = '哈哈太可爱了！我家猫也经常这样' AND p.post_id = c.post_id
SET c.parent_id = p.id
WHERE c.content = '同感！猫咪都是小吃货';

UPDATE comments c
    JOIN comments p ON p.content = '太棒了！请问是什么品种的狗狗？' AND p.post_id = c.post_id
SET c.parent_id = p.id
WHERE c.content = '是金毛哦，特别聪明';

UPDATE comments c
    JOIN comments p ON p.content = '是不是天气太热了？试试换点清淡的食物' AND p.post_id = c.post_id
SET c.parent_id = p.id
WHERE c.content = '或者看看有没有口腔问题';

UPDATE comments c
    JOIN comments p ON p.content = '太可爱了！叫什么名字？' AND p.post_id = c.post_id
SET c.parent_id = p.id
WHERE c.content = '叫咪咪，现在才两个月大';

-- ========================================
-- Phase5：活动 / 领养
-- ========================================
INSERT INTO activities (title, description, type, cover, status, start_time, end_time, rules, prizes, max_participants, max_submissions, votes_per_day)
SELECT * FROM (
    SELECT '萌宠摄影大赛' AS title, '展示你家毛孩子的最美瞬间，赢取精美奖品！' AS description, 'online_vote' AS type, 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20pet%20photography%20contest%20banner&image_size=landscape_16_9' AS cover, 1 AS status, '2026-07-01 00:00:00' AS start_time, '2026-07-31 23:59:59' AS end_time, '1. 每人最多提交3张照片\n2. 禁止盗用他人作品\n3. 投票每天限10票' AS rules, '[{\"level\":\"first\",\"name\":\"宠物大礼包\",\"description\":\"价值500元宠物用品\"},{\"level\":\"second\",\"name\":\"宠物零食礼盒\",\"description\":\"价值200元宠物零食\"},{\"level\":\"third\",\"name\":\"宠物玩具套装\",\"description\":\"价值100元宠物玩具\"}]' AS prizes, 1000 AS max_participants, 3 AS max_submissions, 10 AS votes_per_day
    UNION ALL SELECT '宠物线下聚会', '周末宠物乐园聚会，让毛孩子们交朋友！', 'offline_meet', 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=pet%20offline%20meeting%20party%20banner&image_size=landscape_16_9', 1, '2026-07-15 14:00:00', '2026-07-15 18:00:00', '1. 请牵好宠物牵引绳\n2. 自带宠物饮水和食物\n3. 保持场地清洁', '[{\"level\":\"participant\",\"name\":\"参与礼品\",\"description\":\"宠物零食礼包\"}]', 50, 0, 0
    UNION ALL SELECT '流浪猫领养活动', '给流浪猫咪一个温暖的家', 'adopt', 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=stray%20cat%20adoption%20event%20banner&image_size=landscape_16_9', 1, '2026-07-01 00:00:00', '2026-08-31 23:59:59', '1. 有稳定住所\n2. 有养宠经验优先\n3. 同意定期回访', '[]', 100, 0, 0
) AS seed
WHERE NOT EXISTS (SELECT 1 FROM activities LIMIT 1);

INSERT INTO adoptions (user_id, pet_name, pet_type, pet_age, pet_gender, health_status, photos, description, adopt_conditions, location, contact_phone, status)
SELECT * FROM (
    SELECT @u1 AS user_id, '小橘' AS pet_name, 'cat' AS pet_type, '2岁' AS pet_age, 1 AS pet_gender, '已绝育，疫苗齐全' AS health_status, '[\"https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20orange%20tabby%20cat&image_size=square\"]' AS photos, '性格温顺，喜欢撒娇，是一只非常亲人的橘猫。' AS description, '1. 有稳定住所\n2. 定期驱虫和疫苗\n3. 接受回访' AS adopt_conditions, '北京市朝阳区' AS location, '13800138001' AS contact_phone, 0 AS status
    UNION ALL SELECT @u1, '旺财', 'dog', '1岁', 1, '健康状况良好', '[\"https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20golden%20retriever%20puppy&image_size=square\"]', '活泼好动，聪明伶俐，喜欢和人互动。', '1. 有足够的活动空间\n2. 每天遛狗\n3. 不丢弃不弃养', '上海市浦东新区', '13800138002', 0
) AS seed
WHERE @u1 IS NOT NULL AND NOT EXISTS (SELECT 1 FROM adoptions LIMIT 1);

-- ========================================
-- 脚本执行完成
-- ========================================
SELECT '开发种子数据（Phase1/2/5）执行完成' AS result;
