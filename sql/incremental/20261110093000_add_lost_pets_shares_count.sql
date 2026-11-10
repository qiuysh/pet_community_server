-- ========================================
-- [增量·结构] 2026-11-10 寻宠分享数
-- 路径: server/sql/incremental/20261110093000_add_lost_pets_shares_count.sql
-- 适用: 已有库按时间序执行；全新库只需 init/create.sql（该列已同步进 create.sql）
--
-- 背景：R2 寻宠分享复用帖子/话题「选平台后才上报」口径，lost_pets 需 shares_count。
-- 说明：幂等，可重复执行。
-- ========================================

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'lost_pets'
    AND COLUMN_NAME = 'shares_count'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE lost_pets ADD COLUMN shares_count INT NOT NULL DEFAULT 0 COMMENT ''分享数（用户选择分享平台后累加）'' AFTER view_count',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 校验：应返回 1 行
SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'lost_pets' AND COLUMN_NAME = 'shares_count';

-- 回滚（确认要还原时执行）
-- ALTER TABLE lost_pets DROP COLUMN shares_count;
