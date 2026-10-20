-- ========================================
-- [增量·合集] 2026-10-20 结构 / 数据补丁
-- 路径: server/sql/incremental/20261020113000_fix_20261020.sql
-- 创建日期: 2026-10-20
-- 适用: 已有库按时间序执行；全新库只需 init/create.sql
--       （第一节已同步进 create.sql；第二节为存量数据改写，新库无旧前缀可跳过或幂等空跑）
--
-- 包含 2 项（由当日两份增量合并，全部幂等）：
--   一、唯一索引名与 GORM AutoMigrate 对齐（消 Error 1061）
--       user_levels / notification_settings / daily_stats
--   二、存量上传地址前缀迁移 /uploads/ → /api/v1/uploads/
--       （M4-04 存储层重构后的数据修正）
--
-- 说明：可重复执行；执行前建议备份（见第二节第 0 段）。
-- ========================================

USE pet_community;

-- ############################################################################
-- 一、唯一索引名与 GORM AutoMigrate 对齐
-- ############################################################################
-- 背景：旧表对同列同时/仅有非唯一 INDEX idx_*（GORM uniqueIndex 也叫这个名字）。
-- AutoMigrate 再建 UNIQUE idx_* → Error 1061。user_id 外键占用索引时需先 DROP FK。
--
-- helper 模式：对「列上只要最终留下 UNIQUE idx_*」：
--   1) 若有 FK，先删 FK
--   2) 丢掉非唯一 idx_*
--   3) 若已有 uk_* 且无 idx_*，rename；否则若无 idx_*，ADD UNIQUE idx_*
--   4) 清掉残留 uk_*
--   5) 若原先有 FK，按原引用重建

-- ========================================
-- 1.1 user_levels.user_id
-- ========================================
SET @fk := (
  SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_levels'
    AND COLUMN_NAME = 'user_id' AND REFERENCED_TABLE_NAME IS NOT NULL
  LIMIT 1
);
SET @sql := IF(@fk IS NULL, 'SELECT 1',
  CONCAT('ALTER TABLE user_levels DROP FOREIGN KEY `', @fk, '`'));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_levels'
    AND INDEX_NAME = 'idx_user_levels_user_id' AND NON_UNIQUE = 1
);
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE user_levels DROP INDEX idx_user_levels_user_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_levels' AND INDEX_NAME = 'uk_user_levels_user_id'
);
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_levels' AND INDEX_NAME = 'idx_user_levels_user_id'
);
SET @sql := CASE
  WHEN @has_uk > 0 AND @has_idx = 0 THEN
    'ALTER TABLE user_levels RENAME INDEX uk_user_levels_user_id TO idx_user_levels_user_id'
  WHEN @has_idx = 0 THEN
    'ALTER TABLE user_levels ADD UNIQUE KEY idx_user_levels_user_id (user_id)'
  ELSE 'SELECT 1'
END;
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_levels' AND INDEX_NAME = 'uk_user_levels_user_id'
);
SET @sql := IF(@has_uk = 0, 'SELECT 1', 'ALTER TABLE user_levels DROP INDEX uk_user_levels_user_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists := (
  SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'user_levels'
    AND CONSTRAINT_TYPE = 'FOREIGN KEY' AND CONSTRAINT_NAME = 'user_levels_ibfk_1'
);
SET @sql := IF(@fk_exists > 0, 'SELECT 1',
  'ALTER TABLE user_levels ADD CONSTRAINT user_levels_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 1.2 notification_settings.user_id
-- ========================================
SET @fk := (
  SELECT CONSTRAINT_NAME FROM information_schema.KEY_COLUMN_USAGE
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings'
    AND COLUMN_NAME = 'user_id' AND REFERENCED_TABLE_NAME IS NOT NULL
  LIMIT 1
);
SET @sql := IF(@fk IS NULL, 'SELECT 1',
  CONCAT('ALTER TABLE notification_settings DROP FOREIGN KEY `', @fk, '`'));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings'
    AND INDEX_NAME = 'idx_notification_settings_user_id' AND NON_UNIQUE = 1
);
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE notification_settings DROP INDEX idx_notification_settings_user_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings' AND INDEX_NAME = 'uk_notification_settings_user_id'
);
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings' AND INDEX_NAME = 'idx_notification_settings_user_id'
);
SET @sql := CASE
  WHEN @has_uk > 0 AND @has_idx = 0 THEN
    'ALTER TABLE notification_settings RENAME INDEX uk_notification_settings_user_id TO idx_notification_settings_user_id'
  WHEN @has_idx = 0 THEN
    'ALTER TABLE notification_settings ADD UNIQUE KEY idx_notification_settings_user_id (user_id)'
  ELSE 'SELECT 1'
END;
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings' AND INDEX_NAME = 'uk_notification_settings_user_id'
);
SET @sql := IF(@has_uk = 0, 'SELECT 1', 'ALTER TABLE notification_settings DROP INDEX uk_notification_settings_user_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @fk_exists := (
  SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings'
    AND CONSTRAINT_TYPE = 'FOREIGN KEY' AND CONSTRAINT_NAME = 'notification_settings_ibfk_1'
);
SET @sql := IF(@fk_exists > 0, 'SELECT 1',
  'ALTER TABLE notification_settings ADD CONSTRAINT notification_settings_ibfk_1 FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 1.3 daily_stats.stat_date（无外键）
-- ========================================
SET @c := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'daily_stats'
    AND INDEX_NAME = 'idx_daily_stats_stat_date' AND NON_UNIQUE = 1
);
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE daily_stats DROP INDEX idx_daily_stats_stat_date');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'daily_stats' AND INDEX_NAME = 'uk_daily_stats_date'
);
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'daily_stats' AND INDEX_NAME = 'idx_daily_stats_stat_date'
);
SET @sql := CASE
  WHEN @has_uk > 0 AND @has_idx = 0 THEN
    'ALTER TABLE daily_stats RENAME INDEX uk_daily_stats_date TO idx_daily_stats_stat_date'
  WHEN @has_idx = 0 THEN
    'ALTER TABLE daily_stats ADD UNIQUE KEY idx_daily_stats_stat_date (stat_date)'
  ELSE 'SELECT 1'
END;
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'daily_stats' AND INDEX_NAME = 'uk_daily_stats_date'
);
SET @sql := IF(@has_uk = 0, 'SELECT 1', 'ALTER TABLE daily_stats DROP INDEX uk_daily_stats_date');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ############################################################################
-- 二、存量上传地址前缀迁移：/uploads/ → /api/v1/uploads/
-- ############################################################################
-- 背景：M4-04 后本地对外前缀为 /api/v1/uploads，旧 /uploads 静态路由已移除。
-- 磁盘文件未动，仅改库内相对路径；http(s):// 与已含新前缀的值不改；幂等。
--
-- 执行步骤
--   1) 备份（第 0 段，shell）
--   2) 第 1 段预览
--   3) 第 2 段事务改写
--   4) 第 3 段校验（残留应为 0）

-- 0. 备份（在 shell 执行，不是 SQL）
-- ---------------------------------------------------------------------------
-- mysqldump -u root -p pet_community \
--   users user_third_accounts pet_breeds pets pet_albums pet_growth_records \
--   topic_categories topics posts drafts messages lost_pets activities \
--   activity_submissions activity_albums activity_reviews adoptions \
--   adoption_follow_ups advertisements \
--   > backup_before_upload_prefix_$(date +%Y%m%d%H%M).sql

SET @old_prefix = '/uploads/';
SET @new_prefix = '/api/v1/uploads/';
-- 数组字段的引号锚定模式
SET @arr_old = CONCAT('"', @old_prefix);
SET @arr_new = CONCAT('"', @new_prefix);


-- ---------------------------------------------------------------------------
-- 1. 预览：各字段待改写行数 + 改写后最大长度（VARCHAR(255) 溢出预警）
--    max_len_after 已按 +7（/api/v1 的长度）计算；>255 的那行需先 ALTER 加长。
-- ---------------------------------------------------------------------------
SELECT 'users.avatar' AS field,
       (SELECT COUNT(*) FROM users
         WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
           AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
           AND avatar NOT LIKE '%://%') AS rows_hit,
       (SELECT IFNULL(MAX(CHAR_LENGTH(avatar)), 0) + 7 FROM users) AS max_len_after
UNION ALL SELECT 'user_third_accounts.avatar',
       (SELECT COUNT(*) FROM user_third_accounts
         WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
           AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
           AND avatar NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(avatar)), 0) + 7 FROM user_third_accounts)
UNION ALL SELECT 'pet_breeds.avatar',
       (SELECT COUNT(*) FROM pet_breeds
         WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
           AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
           AND avatar NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(avatar)), 0) + 7 FROM pet_breeds)
UNION ALL SELECT 'pets.avatar',
       (SELECT COUNT(*) FROM pets
         WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
           AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
           AND avatar NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(avatar)), 0) + 7 FROM pets)
UNION ALL SELECT 'pet_albums.image_url',
       (SELECT COUNT(*) FROM pet_albums
         WHERE image_url LIKE CONCAT('%', @old_prefix, '%')
           AND image_url NOT LIKE CONCAT('%', @new_prefix, '%')
           AND image_url NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(image_url)), 0) + 7 FROM pet_albums)
UNION ALL SELECT 'pet_albums.thumb_url',
       (SELECT COUNT(*) FROM pet_albums
         WHERE thumb_url LIKE CONCAT('%', @old_prefix, '%')
           AND thumb_url NOT LIKE CONCAT('%', @new_prefix, '%')
           AND thumb_url NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(thumb_url)), 0) + 7 FROM pet_albums)
UNION ALL SELECT 'pet_growth_records.images (array)',
       (SELECT COUNT(*) FROM pet_growth_records
         WHERE images LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'topic_categories.icon',
       (SELECT COUNT(*) FROM topic_categories
         WHERE icon LIKE CONCAT('%', @old_prefix, '%')
           AND icon NOT LIKE CONCAT('%', @new_prefix, '%')
           AND icon NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(icon)), 0) + 7 FROM topic_categories)
UNION ALL SELECT 'topics.cover',
       (SELECT COUNT(*) FROM topics
         WHERE cover LIKE CONCAT('%', @old_prefix, '%')
           AND cover NOT LIKE CONCAT('%', @new_prefix, '%')
           AND cover NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(cover)), 0) + 7 FROM topics)
UNION ALL SELECT 'posts.images (array)',
       (SELECT COUNT(*) FROM posts
         WHERE images LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'posts.video',
       (SELECT COUNT(*) FROM posts
         WHERE video LIKE CONCAT('%', @old_prefix, '%')
           AND video NOT LIKE CONCAT('%', @new_prefix, '%')
           AND video NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(video)), 0) + 7 FROM posts)
UNION ALL SELECT 'drafts.images (array)',
       (SELECT COUNT(*) FROM drafts
         WHERE images LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'drafts.video',
       (SELECT COUNT(*) FROM drafts
         WHERE video LIKE CONCAT('%', @old_prefix, '%')
           AND video NOT LIKE CONCAT('%', @new_prefix, '%')
           AND video NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(video)), 0) + 7 FROM drafts)
UNION ALL SELECT 'messages.media_url',
       (SELECT COUNT(*) FROM messages
         WHERE media_url LIKE CONCAT('%', @old_prefix, '%')
           AND media_url NOT LIKE CONCAT('%', @new_prefix, '%')
           AND media_url NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(media_url)), 0) + 7 FROM messages)
UNION ALL SELECT 'lost_pets.images (array)',
       (SELECT COUNT(*) FROM lost_pets
         WHERE images LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'activities.cover',
       (SELECT COUNT(*) FROM activities
         WHERE cover LIKE CONCAT('%', @old_prefix, '%')
           AND cover NOT LIKE CONCAT('%', @new_prefix, '%')
           AND cover NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(cover)), 0) + 7 FROM activities)
UNION ALL SELECT 'activity_submissions.images (array)',
       (SELECT COUNT(*) FROM activity_submissions
         WHERE images LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'activity_submissions.video',
       (SELECT COUNT(*) FROM activity_submissions
         WHERE video LIKE CONCAT('%', @old_prefix, '%')
           AND video NOT LIKE CONCAT('%', @new_prefix, '%')
           AND video NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(video)), 0) + 7 FROM activity_submissions)
UNION ALL SELECT 'activity_albums.photo_url',
       (SELECT COUNT(*) FROM activity_albums
         WHERE photo_url LIKE CONCAT('%', @old_prefix, '%')
           AND photo_url NOT LIKE CONCAT('%', @new_prefix, '%')
           AND photo_url NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(photo_url)), 0) + 7 FROM activity_albums)
UNION ALL SELECT 'activity_reviews.photos (array)',
       (SELECT COUNT(*) FROM activity_reviews
         WHERE photos LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'adoptions.photos (array)',
       (SELECT COUNT(*) FROM adoptions
         WHERE photos LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'adoption_follow_ups.photos (array)',
       (SELECT COUNT(*) FROM adoption_follow_ups
         WHERE photos LIKE CONCAT('%', @arr_old, '%')),
       NULL
UNION ALL SELECT 'advertisements.image_url',
       (SELECT COUNT(*) FROM advertisements
         WHERE image_url LIKE CONCAT('%', @old_prefix, '%')
           AND image_url NOT LIKE CONCAT('%', @new_prefix, '%')
           AND image_url NOT LIKE '%://%'),
       (SELECT IFNULL(MAX(CHAR_LENGTH(image_url)), 0) + 7 FROM advertisements);


-- ---------------------------------------------------------------------------
-- 2. 执行：单事务，任一条失败整体回滚
--    若第 1 段出现 max_len_after > 255，先执行注释里的 ALTER 再跑本段。
-- ---------------------------------------------------------------------------
-- ALTER TABLE messages MODIFY media_url VARCHAR(500) COMMENT '媒体URL（图片/语音）';

START TRANSACTION;

-- ---- 单值字段 ----
UPDATE users
   SET avatar = REPLACE(avatar, @old_prefix, @new_prefix)
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
   AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
   AND avatar NOT LIKE '%://%';

UPDATE user_third_accounts
   SET avatar = REPLACE(avatar, @old_prefix, @new_prefix)
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
   AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
   AND avatar NOT LIKE '%://%';

UPDATE pet_breeds
   SET avatar = REPLACE(avatar, @old_prefix, @new_prefix)
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
   AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
   AND avatar NOT LIKE '%://%';

UPDATE pets
   SET avatar = REPLACE(avatar, @old_prefix, @new_prefix)
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%')
   AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
   AND avatar NOT LIKE '%://%';

UPDATE pet_albums
   SET image_url = IF(image_url LIKE CONCAT('%', @old_prefix, '%')
                      AND image_url NOT LIKE CONCAT('%', @new_prefix, '%')
                      AND image_url NOT LIKE '%://%',
                      REPLACE(image_url, @old_prefix, @new_prefix), image_url),
       thumb_url = IF(thumb_url LIKE CONCAT('%', @old_prefix, '%')
                      AND thumb_url NOT LIKE CONCAT('%', @new_prefix, '%')
                      AND thumb_url NOT LIKE '%://%',
                      REPLACE(thumb_url, @old_prefix, @new_prefix), thumb_url)
 WHERE image_url LIKE CONCAT('%', @old_prefix, '%')
    OR thumb_url LIKE CONCAT('%', @old_prefix, '%');

UPDATE topic_categories
   SET icon = REPLACE(icon, @old_prefix, @new_prefix)
 WHERE icon LIKE CONCAT('%', @old_prefix, '%')
   AND icon NOT LIKE CONCAT('%', @new_prefix, '%')
   AND icon NOT LIKE '%://%';

UPDATE topics
   SET cover = REPLACE(cover, @old_prefix, @new_prefix)
 WHERE cover LIKE CONCAT('%', @old_prefix, '%')
   AND cover NOT LIKE CONCAT('%', @new_prefix, '%')
   AND cover NOT LIKE '%://%';

UPDATE messages
   SET media_url = REPLACE(media_url, @old_prefix, @new_prefix)
 WHERE media_url LIKE CONCAT('%', @old_prefix, '%')
   AND media_url NOT LIKE CONCAT('%', @new_prefix, '%')
   AND media_url NOT LIKE '%://%';

UPDATE activities
   SET cover = REPLACE(cover, @old_prefix, @new_prefix)
 WHERE cover LIKE CONCAT('%', @old_prefix, '%')
   AND cover NOT LIKE CONCAT('%', @new_prefix, '%')
   AND cover NOT LIKE '%://%';

UPDATE activity_albums
   SET photo_url = REPLACE(photo_url, @old_prefix, @new_prefix)
 WHERE photo_url LIKE CONCAT('%', @old_prefix, '%')
   AND photo_url NOT LIKE CONCAT('%', @new_prefix, '%')
   AND photo_url NOT LIKE '%://%';

UPDATE advertisements
   SET image_url = REPLACE(image_url, @old_prefix, @new_prefix)
 WHERE image_url LIKE CONCAT('%', @old_prefix, '%')
   AND image_url NOT LIKE CONCAT('%', @new_prefix, '%')
   AND image_url NOT LIKE '%://%';

-- ---- 数组字段（引号锚定，逐元素替换） ----
UPDATE pet_growth_records
   SET images = REPLACE(images, @arr_old, @arr_new)
 WHERE images LIKE CONCAT('%', @arr_old, '%');

UPDATE lost_pets
   SET images = REPLACE(images, @arr_old, @arr_new)
 WHERE images LIKE CONCAT('%', @arr_old, '%');

UPDATE activity_reviews
   SET photos = REPLACE(photos, @arr_old, @arr_new)
 WHERE photos LIKE CONCAT('%', @arr_old, '%');

UPDATE adoptions
   SET photos = REPLACE(photos, @arr_old, @arr_new)
 WHERE photos LIKE CONCAT('%', @arr_old, '%');

UPDATE adoption_follow_ups
   SET photos = REPLACE(photos, @arr_old, @arr_new)
 WHERE photos LIKE CONCAT('%', @arr_old, '%');

-- ---- 混合表：数组列与单值列各用自己的规则 ----
UPDATE posts
   SET images = IF(images LIKE CONCAT('%', @arr_old, '%'),
                   REPLACE(images, @arr_old, @arr_new), images),
       video  = IF(video LIKE CONCAT('%', @old_prefix, '%')
                   AND video NOT LIKE CONCAT('%', @new_prefix, '%')
                   AND video NOT LIKE '%://%',
                   REPLACE(video, @old_prefix, @new_prefix), video)
 WHERE images LIKE CONCAT('%', @arr_old, '%')
    OR video LIKE CONCAT('%', @old_prefix, '%');

UPDATE drafts
   SET images = IF(images LIKE CONCAT('%', @arr_old, '%'),
                   REPLACE(images, @arr_old, @arr_new), images),
       video  = IF(video LIKE CONCAT('%', @old_prefix, '%')
                   AND video NOT LIKE CONCAT('%', @new_prefix, '%')
                   AND video NOT LIKE '%://%',
                   REPLACE(video, @old_prefix, @new_prefix), video)
 WHERE images LIKE CONCAT('%', @arr_old, '%')
    OR video LIKE CONCAT('%', @old_prefix, '%');

UPDATE activity_submissions
   SET images = IF(images LIKE CONCAT('%', @arr_old, '%'),
                   REPLACE(images, @arr_old, @arr_new), images),
       video  = IF(video LIKE CONCAT('%', @old_prefix, '%')
                   AND video NOT LIKE CONCAT('%', @new_prefix, '%')
                   AND video NOT LIKE '%://%',
                   REPLACE(video, @old_prefix, @new_prefix), video)
 WHERE images LIKE CONCAT('%', @arr_old, '%')
    OR video LIKE CONCAT('%', @old_prefix, '%');

COMMIT;


-- ---------------------------------------------------------------------------
-- 3. 校验：remaining 应全为 0；再抽查几条实际数据
-- ---------------------------------------------------------------------------
SELECT 'users.avatar' AS field, COUNT(*) AS remaining FROM users
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%') AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'user_third_accounts.avatar', COUNT(*) FROM user_third_accounts
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%') AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'pet_breeds.avatar', COUNT(*) FROM pet_breeds
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%') AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'pets.avatar', COUNT(*) FROM pets
 WHERE avatar LIKE CONCAT('%', @old_prefix, '%') AND avatar NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'pet_albums.image_url', COUNT(*) FROM pet_albums
 WHERE image_url LIKE CONCAT('%', @old_prefix, '%') AND image_url NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'pet_albums.thumb_url', COUNT(*) FROM pet_albums
 WHERE thumb_url LIKE CONCAT('%', @old_prefix, '%') AND thumb_url NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'pet_growth_records.images', COUNT(*) FROM pet_growth_records
 WHERE images LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'topic_categories.icon', COUNT(*) FROM topic_categories
 WHERE icon LIKE CONCAT('%', @old_prefix, '%') AND icon NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'topics.cover', COUNT(*) FROM topics
 WHERE cover LIKE CONCAT('%', @old_prefix, '%') AND cover NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'posts.images', COUNT(*) FROM posts
 WHERE images LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'posts.video', COUNT(*) FROM posts
 WHERE video LIKE CONCAT('%', @old_prefix, '%') AND video NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'drafts.images', COUNT(*) FROM drafts
 WHERE images LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'drafts.video', COUNT(*) FROM drafts
 WHERE video LIKE CONCAT('%', @old_prefix, '%') AND video NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'messages.media_url', COUNT(*) FROM messages
 WHERE media_url LIKE CONCAT('%', @old_prefix, '%') AND media_url NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'lost_pets.images', COUNT(*) FROM lost_pets
 WHERE images LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'activities.cover', COUNT(*) FROM activities
 WHERE cover LIKE CONCAT('%', @old_prefix, '%') AND cover NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'activity_submissions.images', COUNT(*) FROM activity_submissions
 WHERE images LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'activity_submissions.video', COUNT(*) FROM activity_submissions
 WHERE video LIKE CONCAT('%', @old_prefix, '%') AND video NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'activity_albums.photo_url', COUNT(*) FROM activity_albums
 WHERE photo_url LIKE CONCAT('%', @old_prefix, '%') AND photo_url NOT LIKE CONCAT('%', @new_prefix, '%')
UNION ALL SELECT 'activity_reviews.photos', COUNT(*) FROM activity_reviews
 WHERE photos LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'adoptions.photos', COUNT(*) FROM adoptions
 WHERE photos LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'adoption_follow_ups.photos', COUNT(*) FROM adoption_follow_ups
 WHERE photos LIKE CONCAT('%', @arr_old, '%')
UNION ALL SELECT 'advertisements.image_url', COUNT(*) FROM advertisements
 WHERE image_url LIKE CONCAT('%', @old_prefix, '%') AND image_url NOT LIKE CONCAT('%', @new_prefix, '%');

-- 抽查（改完应为 /api/v1/uploads/... 且前缀不重复）
SELECT id, avatar FROM users ORDER BY id DESC LIMIT 5;
SELECT id, media_url, type FROM messages
 WHERE type IN ('image', 'voice') ORDER BY id DESC LIMIT 5;
SELECT id, images FROM posts WHERE images LIKE '%/uploads/%' ORDER BY id DESC LIMIT 5;


-- ---------------------------------------------------------------------------
-- 4. 回滚（仅在已备份 / 确认要还原时执行，取消注释整段）
--    注意：会连带把迁移后新产生的地址改回旧前缀，务必在停写窗口内操作。
-- ---------------------------------------------------------------------------
-- START TRANSACTION;
-- UPDATE users SET avatar = REPLACE(avatar, @new_prefix, @old_prefix)
--  WHERE avatar LIKE CONCAT(@new_prefix, '%');
-- UPDATE pets SET avatar = REPLACE(avatar, @new_prefix, @old_prefix)
--  WHERE avatar LIKE CONCAT(@new_prefix, '%');
-- UPDATE pet_albums
--    SET image_url = REPLACE(image_url, @new_prefix, @old_prefix),
--        thumb_url = REPLACE(thumb_url, @new_prefix, @old_prefix)
--  WHERE image_url LIKE CONCAT(@new_prefix, '%') OR thumb_url LIKE CONCAT(@new_prefix, '%');
-- UPDATE messages SET media_url = REPLACE(media_url, @new_prefix, @old_prefix)
--  WHERE media_url LIKE CONCAT(@new_prefix, '%');
-- UPDATE posts
--    SET images = REPLACE(images, @arr_new, @arr_old),
--        video  = REPLACE(video, @new_prefix, @old_prefix);
-- UPDATE adoptions SET photos = REPLACE(photos, @arr_new, @arr_old);
-- -- 其余单值表按 users 模式、其余数组表按 adoptions 模式补全
-- COMMIT;
