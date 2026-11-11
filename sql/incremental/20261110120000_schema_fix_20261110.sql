-- ============================================================
-- 2026-11-06 ~ 2026-11-10 schema 收口合集（幂等，可重复执行）
-- 路径: server/sql/incremental/20261110120000_schema_fix_20261110.sql
-- 整合日期: 2026-11-11
--
-- 本文件由以下增量按时间序整合而成（正文逐字保留；第四节原非幂等 ALTER
-- 已包成 information_schema + PREPARE，便于合集重复空跑）：
--   20261106000000_add_notification_settings_message_notify.sql → 第一节
--   20261110093000_add_lost_pets_shares_count.sql               → 第二节
--   20261110094500_add_notification_settings_security_notify.sql → 第三节
--   20261110120000_add_pets_visibility.sql                      → 第四节
--   （原分文件已删除；曾按分文件执行过的库，本合集为幂等空跑，可跳过。）
--
-- 适用对象：已有库按时间序执行；全新库只需 init/create.sql（列已同步进去）。
-- ============================================================

USE pet_community;

-- ############################################################
-- 一、通知设置私信开关：notification_settings.message_notify
-- （原 20261106000000_add_notification_settings_message_notify.sql）
-- ############################################################

-- ============================================================
-- [增量·结构] 2026-11-06 通知设置私信开关：notification_settings.message_notify
-- 路径: server/sql/incremental/20261106000000_add_notification_settings_message_notify.sql
-- 创建日期: 2026-11-06
-- 适用: 已有库按时间序执行；全新库只需 init/create.sql（该列已同步进 create.sql）
--
-- 背景：N8 通知开关收口 ——
--   1) settings 各开关此前只存库、创建通知时不生效，本批次在
--      NotificationService / FollowService / AdminNotificationService 侧统一过滤；
--   2) 规格 N8-01/N8-07 要求设置页含「私信」开关，补 message_notify 字段
--      （关闭后不发送实时推送；消息仍入库、未读计数仍累加）。
-- ============================================================

-- 仅当列不存在时执行 ALTER（MySQL 8.0 不支持 ADD COLUMN IF NOT EXISTS）
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notification_settings'
    AND COLUMN_NAME = 'message_notify'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE notification_settings ADD COLUMN message_notify TINYINT DEFAULT 1 COMMENT ''私信提醒（0-关闭实时推送，1-开启；消息仍入库并计未读）'' AFTER system_notify',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- 存量行补默认值（列默认 1，这里只兜底历史 NULL 行）
UPDATE notification_settings SET message_notify = 1 WHERE message_notify IS NULL;

-- ############################################################
-- 二、寻宠分享数：lost_pets.shares_count
-- （原 20261110093000_add_lost_pets_shares_count.sql）
-- ############################################################

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

-- ############################################################
-- 三、通知设置安全提醒开关：notification_settings.security_notify
-- （原 20261110094500_add_notification_settings_security_notify.sql）
-- ############################################################

-- ========================================
-- [增量·结构] 2026-11-10 通知设置安全提醒开关
-- 路径: server/sql/incremental/20261110094500_add_notification_settings_security_notify.sql
-- 适用: 已有库按时间序执行；全新库只需 init/create.sql（该列已同步进 create.sql）
--
-- 背景：R4 新设备登录通知需独立「安全提醒」开关，默认开启。
-- 说明：幂等，可重复执行。
-- ========================================

SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'notification_settings'
    AND COLUMN_NAME = 'security_notify'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE notification_settings ADD COLUMN security_notify TINYINT DEFAULT 1 COMMENT ''安全提醒（0-关闭新设备登录通知，1-开启）'' AFTER message_notify',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE notification_settings SET security_notify = 1 WHERE security_notify IS NULL;

-- 校验：应返回 1 行
SELECT COLUMN_NAME, COLUMN_TYPE, COLUMN_DEFAULT
FROM information_schema.COLUMNS
WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'notification_settings' AND COLUMN_NAME = 'security_notify';

-- 回滚（确认要还原时执行）
-- ALTER TABLE notification_settings DROP COLUMN security_notify;

-- ############################################################
-- 四、宠物可见范围：pets.visibility
-- （原 20261110120000_add_pets_visibility.sql；合集内改为幂等）
-- ############################################################

-- R5：宠物可见范围（0公开 / 1仅粉丝 / 2仅自己），默认公开
SET @col_exists := (
  SELECT COUNT(*) FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'pets'
    AND COLUMN_NAME = 'visibility'
);
SET @sql := IF(@col_exists = 0,
  'ALTER TABLE pets ADD COLUMN visibility TINYINT NOT NULL DEFAULT 0 COMMENT ''可见范围：0公开 1仅粉丝 2仅自己'' AFTER status',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @idx_exists := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'pets'
    AND INDEX_NAME = 'idx_pets_visibility'
);
SET @sql := IF(@idx_exists = 0,
  'CREATE INDEX idx_pets_visibility ON pets (visibility)',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
