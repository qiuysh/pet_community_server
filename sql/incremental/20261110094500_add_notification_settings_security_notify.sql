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
