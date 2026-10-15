-- ========================================
-- [增量·开发种子] Phase1/2/5 测试数据合集
-- 路径: server/sql/incremental/20260724100000_seed_dev_data.sql
-- 适用: 在 init/create.sql 之后执行（开发 / E2E）
-- 内容: 原 02_seed_dev_phase1 + 03_seed_dev_phase2 + 04_seed_dev_phase5
-- 创建日期: 2026-07-24
-- ========================================

USE pet_community;

-- ========================================
-- Phase1：测试用户
-- ========================================
-- 密码统一为: user123456 (BCrypt加密)
-- status: 0=未激活/禁用, 1=正常, 2=封禁（登录要求 status=1）
INSERT INTO users (phone, email, password, nickname, avatar, signature, gender, location, pet_years, status, privacy_settings, created_at, updated_at) VALUES
('13800138001', 'user1@petcommunity.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjzqAKL9xL5jvMFVdNJHvGCgTq/VEq', '萌宠小达人', '', '家有两猫一狗，生活乐无忧~', 2, '北京市朝阳区', '3-5年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138002', 'user2@petcommunity.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjzqAKL9xL5jvMFVdNJHvGCgTq/VEq', '铲屎官小王', '', '每天都是被猫咪治愈的一天', 1, '上海市浦东新区', '1-3年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138003', 'user3@petcommunity.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjzqAKL9xL5jvMFVdNJHvGCgTq/VEq', '金毛妈妈', '', '我家金毛叫旺财，今年3岁啦', 2, '广州市天河区', '3-5年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138004', 'user4@petcommunity.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjzqAKL9xL5jvMFVdNJHvGCgTq/VEq', '猫奴一枚', '', '三只猫的铲屎官，痛并快乐着', 0, '深圳市南山区', '5年以上', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW()),
('13800138005', 'user5@petcommunity.com', '$2a$10$N9qo8uLOickgx2ZMRZoMye.IjzqAKL9xL5jvMFVdNJHvGCgTq/VEq', '新手铲屎官', '', '刚养了一只小柯基，求指导~', 1, '杭州市西湖区', '0-1年', 1, '{"visibility":"public","commentPermission":"all","messagePermission":"all"}', NOW(), NOW())
ON DUPLICATE KEY UPDATE status = VALUES(status), updated_at = NOW();

-- ========================================
-- Phase1：测试宠物
-- ========================================
INSERT INTO pets (user_id, name, avatar, pet_type, breed_id, breed_name, gender, weight, personality, bio, is_sterilized, is_vaccinated, status, view_count, created_at, updated_at) VALUES
(1, '橘座', '', 2, 0, '橘猫', 1, 5.5, '吃货、懒、爱撒娇', '一只胖橘，每天除了吃就是睡，偶尔卖萌求抚摸。', 1, 1, 1, 128, NOW(), NOW()),
(1, '小黑', '', 2, 0, '狸花猫', 2, 3.2, '活泼、好动、爱探险', '家里的小调皮，上蹿下停不下来，但是很亲人。', 0, 1, 1, 86, NOW(), NOW()),
(1, '豆豆', '', 1, 0, '柯基', 1, 10.0, '憨厚、可爱、小短腿', '柯基豆豆，屁股扭扭超可爱，最爱出去玩。', 1, 1, 1, 256, NOW(), NOW()),
(2, '咪咪', '', 2, 0, '英国短毛猫', 2, 4.0, '高冷、优雅、爱干净', '英短蓝猫咪咪，高贵冷艳的小公主，只在吃饭时才理人。', 1, 1, 1, 168, NOW(), NOW()),
(3, '旺财', '', 1, 0, '金毛寻回犬', 1, 28.5, '温顺、聪明、大暖男', '金毛旺财，超级大暖男，见谁都摇尾巴，是小区里的明星狗狗。', 1, 1, 1, 312, NOW(), NOW()),
(4, '布偶公主', '', 2, 0, '布偶猫', 2, 4.8, '粘人、温柔、仙女猫', '布偶猫小仙女，温柔粘人，抱起来就不想撒手。', 1, 1, 1, 234, NOW(), NOW()),
(4, '暹罗王子', '', 2, 0, '暹罗猫', 1, 3.5, '话痨、聪明、好动', '暹罗猫话痨本痨，每天喵喵喵说个不停，但是很聪明。', 0, 1, 1, 145, NOW(), NOW()),
(4, '加菲', '', 2, 0, '加菲猫', 1, 5.0, '懒、爱吃、爱睡觉', '加菲猫懒癌晚期，能躺着绝不坐着，能坐着绝不站着。', 1, 1, 1, 178, NOW(), NOW()),
(5, '小短腿', '', 1, 0, '柯基', 2, 8.5, '活泼、可爱、贪吃', '刚接回家的小柯基，每天活力满满，什么都想吃。', 0, 1, 1, 98, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ========================================
-- Phase1：宠物相册示例
-- ========================================
INSERT INTO pet_albums (pet_id, image_url, thumb_url, title, sort_order, created_at) VALUES
(1, '', '', '日常照', 1, NOW()),
(1, '', '', '睡觉照', 2, NOW()),
(1, '', '', '吃饭照', 3, NOW()),
(2, '', '', '日常照', 1, NOW()),
(3, '', '', '出去玩', 1, NOW()),
(3, '', '', '海边', 2, NOW()),
(5, '', '', '小时候', 1, NOW()),
(5, '', '', '现在', 2, NOW()),
(5, '', '', '和小伙伴', 3, NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ========================================
-- Phase2：话题 / 动态 / 评论
-- ========================================
INSERT INTO topics (name, description, cover, posts_count, status, created_at, updated_at) VALUES
('宠物求助', '宠物相关问题求助', '', 0, 1, NOW(), NOW()),
('晒宠大赛', '晒出你家最可爱的宠物', '', 0, 1, NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = NOW();

INSERT INTO posts (user_id, content, images, video, pet_ids, topic_ids, location, visibility, status, likes_count, comments_count, shares_count, favorites_count, is_pinned, is_featured, created_at, updated_at) VALUES
(1, '我家猫咪今天早上偷了我的早餐，太调皮了！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20cat%20eating%20bread%20funny%20moment&image_size=portrait_4_3"]', '', '[1]', '[1]', '北京市朝阳区', 0, 1, 12, 5, 3, 8, 0, 0, NOW(), NOW()),
(1, '分享一下我家狗狗的训练成果，学会了好多新技能！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=smart%20dog%20training%20obedience&image_size=portrait_4_3"]', '', '[3]', '[2]', '北京市海淀区', 0, 1, 28, 12, 8, 15, 0, 1, NOW(), NOW()),
(2, '求助！我家猫咪最近不爱吃东西怎么办？', '', '', '', '[4]', '上海市浦东新区', 0, 1, 8, 23, 2, 3, 0, 0, NOW(), NOW()),
(2, '自制宠物零食教程，简单又健康！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=homemade%20pet%20treats%20healthy%20snacks&image_size=portrait_4_3"]', '', '', '[3]', '上海市静安区', 0, 1, 45, 18, 12, 25, 0, 1, NOW(), NOW()),
(3, '带狗狗去公园玩的一天，它开心极了！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=happy%20dog%20playing%20in%20park&image_size=portrait_4_3"]', '', '[5]', '[1,5]', '广州市天河区', 0, 1, 35, 9, 6, 12, 0, 0, NOW(), NOW()),
(3, '给大家看看我家的新成员，一只可爱的小奶猫！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=adorable%20kitten%20new%20pet&image_size=portrait_4_3"]', '', '', '[1]', '广州市越秀区', 0, 1, 67, 32, 15, 38, 1, 1, NOW(), NOW()),
(4, '宠物驱虫知识分享，新手必看！', '', '', '', '[2,4]', '深圳市南山区', 0, 1, 89, 45, 22, 56, 0, 1, NOW(), NOW()),
(4, '我家狗狗的搞笑瞬间合集，看一次笑一次！', '["https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=funny%20dog%20compilation%20cute&image_size=portrait_4_3"]', '', '', '[5]', '深圳市福田区', 0, 1, 120, 58, 35, 72, 0, 0, NOW(), NOW());

INSERT INTO comments (post_id, user_id, parent_id, reply_to_id, content, at_users, status, likes_count, created_at, updated_at) VALUES
(1, 2, 0, 0, '哈哈太可爱了！我家猫也经常这样', '', 1, 3, NOW(), NOW()),
(1, 3, 1, 2, '同感！猫咪都是小吃货', '', 1, 1, NOW(), NOW()),
(2, 4, 0, 0, '太棒了！请问是什么品种的狗狗？', '', 1, 5, NOW(), NOW()),
(2, 1, 3, 4, '是金毛哦，特别聪明', '', 1, 2, NOW(), NOW()),
(3, 1, 0, 0, '是不是天气太热了？试试换点清淡的食物', '', 1, 8, NOW(), NOW()),
(3, 4, 4, 1, '或者看看有没有口腔问题', '', 1, 4, NOW(), NOW()),
(4, 3, 0, 0, '看起来很好吃！收藏了', '', 1, 6, NOW(), NOW()),
(5, 2, 0, 0, '好羡慕！我们这边公园不让狗狗进', '', 1, 4, NOW(), NOW()),
(6, 1, 0, 0, '太可爱了！叫什么名字？', '', 1, 12, NOW(), NOW()),
(6, 3, 8, 1, '叫咪咪，现在才两个月大', '', 1, 3, NOW(), NOW()),
(7, 2, 0, 0, '非常实用！感谢分享', '', 1, 15, NOW(), NOW()),
(8, 4, 0, 0, '笑死我了！', '', 1, 23, NOW(), NOW());

-- ========================================
-- Phase5：活动 / 领养
-- ========================================
INSERT INTO activities (title, description, type, cover, status, start_time, end_time, rules, prizes, max_participants, max_submissions, votes_per_day) VALUES
('萌宠摄影大赛', '展示你家毛孩子的最美瞬间，赢取精美奖品！', 'online_vote', 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20pet%20photography%20contest%20banner&image_size=landscape_16_9', 1, '2026-07-01 00:00:00', '2026-07-31 23:59:59', '1. 每人最多提交3张照片\n2. 禁止盗用他人作品\n3. 投票每天限10票', '[{\"level\":\"first\",\"name\":\"宠物大礼包\",\"description\":\"价值500元宠物用品\"},{\"level\":\"second\",\"name\":\"宠物零食礼盒\",\"description\":\"价值200元宠物零食\"},{\"level\":\"third\",\"name\":\"宠物玩具套装\",\"description\":\"价值100元宠物玩具\"}]', 1000, 3, 10),
('宠物线下聚会', '周末宠物乐园聚会，让毛孩子们交朋友！', 'offline_meet', 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=pet%20offline%20meeting%20party%20banner&image_size=landscape_16_9', 1, '2026-07-15 14:00:00', '2026-07-15 18:00:00', '1. 请牵好宠物牵引绳\n2. 自带宠物饮水和食物\n3. 保持场地清洁', '[{\"level\":\"participant\",\"name\":\"参与礼品\",\"description\":\"宠物零食礼包\"}]', 50, 0, 0),
('流浪猫领养活动', '给流浪猫咪一个温暖的家', 'adopt', 'https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=stray%20cat%20adoption%20event%20banner&image_size=landscape_16_9', 1, '2026-07-01 00:00:00', '2026-08-31 23:59:59', '1. 有稳定住所\n2. 有养宠经验优先\n3. 同意定期回访', '[]', 100, 0, 0)
ON DUPLICATE KEY UPDATE updated_at = NOW();

INSERT INTO adoptions (user_id, pet_name, pet_type, pet_age, pet_gender, health_status, photos, description, adopt_conditions, location, contact_phone, status) VALUES
(1, '小橘', 'cat', '2岁', 1, '已绝育，疫苗齐全', '[\"https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20orange%20tabby%20cat&image_size=square\"]', '性格温顺，喜欢撒娇，是一只非常亲人的橘猫。', '1. 有稳定住所\n2. 定期驱虫和疫苗\n3. 接受回访', '北京市朝阳区', '13800138001', 0),
(1, '旺财', 'dog', '1岁', 1, '健康状况良好', '[\"https://trae-api-cn.mchost.guru/api/ide/v1/text_to_image?prompt=cute%20golden%20retriever%20puppy&image_size=square\"]', '活泼好动，聪明伶俐，喜欢和人互动。', '1. 有足够的活动空间\n2. 每天遛狗\n3. 不丢弃不弃养', '上海市浦东新区', '13800138002', 0)
ON DUPLICATE KEY UPDATE updated_at = NOW();

-- ========================================
-- 脚本执行完成
-- ========================================
SELECT '开发种子数据（Phase1/2/5）执行完成' AS result;
