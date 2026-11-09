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

USE pet_community;

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
