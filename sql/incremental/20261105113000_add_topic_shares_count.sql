-- ========================================
-- [增量·结构] 2026-11-05 话题分享数
-- 路径: server/sql/incremental/20261105113000_add_topic_shares_count.sql
-- 创建日期: 2026-11-05
-- 适用: 已有库按时间序执行；全新库只需 init/create.sql（该列已同步进 create.sql）
--
-- 背景：话题详情的分享此前不计数（ShareTopic 丢弃 platform）。
--       前端改为「选平台后才上报」后，用 topics.shares_count 记录分享次数，口径与 posts.shares_count 一致。
-- 说明：幂等，可重复执行。
-- ========================================

USE pet_community;

-- 仅当列不存在时执行 ALTER（MySQL 8.0 不支持 ADD COLUMN IF NOT EXISTS）
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'topics'
    AND COLUMN_NAME = 'shares_count'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE topics ADD COLUMN shares_count INT NOT NULL DEFAULT 0 COMMENT ''分享数（用户选择分享平台后累加）'' AFTER posts_count',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 校验：应返回 1 行，COLUMN_TYPE=int
SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'topics' AND COLUMN_NAME = 'shares_count';

-- 回滚（确认要还原时执行）
-- ALTER TABLE topics DROP COLUMN shares_count;
