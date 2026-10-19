-- ========================================
-- [增量·结构] 2026-10-19 结构补丁合集
-- 路径: server/sql/incremental/20261019154000_schema_fix_20261019.sql
-- 创建日期: 2026-10-19
-- 适用: 已有库按时间序执行；全新库只需 init/create.sql（本文件内容已同步进去）
--
-- 包含 8 项变更（由当日多份增量合并，全部幂等）：
--   一、messages 新增 duration（私信语音时长，秒）
--   二、comments.parent_id     移除外键（0 哨兵值冲突）
--   三、comments.reply_to_id   移除外键（同上）
--   四、pets.breed_id          移除外键（同上）
--   五、system_notifications   与 model.SystemNotification 对齐（varchar → 数值枚举、列重命名）
--   六、follows                与 model.Follow 对齐（followed_id → following_id、补 status）
--   七、清除死表 + 补齐活表
--   八、users.phone/email 索引与 GORM uniqueIndex 对齐（消 1061）
--
-- 二~四 的背景：
--   0 在各处被当作「无」的语义使用（GORM 默认值、`parent_id = 0` / `BreedID > 0`
--   查询、前端 `parentId === 0`、触发器 `NEW.parent_id > 0`），但这三列都挂了外键，
--   而 0 不存在于被引用表 → 任何写入 0 的语句都报 1452。
--   因此保留 0 语义、移除外键；parent_id 的级联删除改由
--   CommentService.DeleteComment 在事务内显式执行。
--
-- 说明：可重复执行；不改业务数据（七仅 DROP 无业务死表）。
-- ========================================

USE pet_community;

-- ========================================
-- 一、messages.duration（私信语音时长）
-- ========================================
SET @col_exists := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'messages'
    AND COLUMN_NAME = 'duration'
);

SET @ddl := IF(
  @col_exists = 0,
  'ALTER TABLE messages ADD COLUMN duration INT NOT NULL DEFAULT 0 COMMENT ''语音时长（秒，非语音为0）'' AFTER media_url',
  'SELECT 1'
);

PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 二、comments.parent_id
-- ========================================
SET @fk_parent := (SELECT CONSTRAINT_NAME
                   FROM information_schema.KEY_COLUMN_USAGE
                   WHERE TABLE_SCHEMA = DATABASE()
                     AND TABLE_NAME = 'comments'
                     AND COLUMN_NAME = 'parent_id'
                     AND REFERENCED_TABLE_NAME IS NOT NULL
                   LIMIT 1);
SET @sql := IF(@fk_parent IS NULL, 'SELECT 1', CONCAT('ALTER TABLE comments DROP FOREIGN KEY `', @fk_parent, '`'));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 三、comments.reply_to_id
-- ========================================
SET @fk_reply := (SELECT CONSTRAINT_NAME
                  FROM information_schema.KEY_COLUMN_USAGE
                  WHERE TABLE_SCHEMA = DATABASE()
                    AND TABLE_NAME = 'comments'
                    AND COLUMN_NAME = 'reply_to_id'
                    AND REFERENCED_TABLE_NAME IS NOT NULL
                  LIMIT 1);
SET @sql := IF(@fk_reply IS NULL, 'SELECT 1', CONCAT('ALTER TABLE comments DROP FOREIGN KEY `', @fk_reply, '`'));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE comments
    MODIFY COLUMN parent_id BIGINT UNSIGNED DEFAULT 0 COMMENT '父评论ID（0=顶级评论）';

ALTER TABLE comments
    MODIFY COLUMN reply_to_id BIGINT UNSIGNED DEFAULT 0 COMMENT '回复的用户ID（0=无）';

-- ========================================
-- 四、pets.breed_id
-- ========================================
SET @fk_breed := (SELECT CONSTRAINT_NAME
                  FROM information_schema.KEY_COLUMN_USAGE
                  WHERE TABLE_SCHEMA = DATABASE()
                    AND TABLE_NAME = 'pets'
                    AND COLUMN_NAME = 'breed_id'
                    AND REFERENCED_TABLE_NAME IS NOT NULL
                  LIMIT 1);
SET @sql := IF(@fk_breed IS NULL, 'SELECT 1', CONCAT('ALTER TABLE pets DROP FOREIGN KEY `', @fk_breed, '`'));
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE pets
    MODIFY COLUMN breed_id BIGINT UNSIGNED DEFAULT 0 COMMENT '品种ID（0=未选品种）';

-- ========================================
-- 五、system_notifications 与 model.SystemNotification 对齐
-- ========================================
-- 旧：type/target_type VARCHAR；target_users / push_time / created_by
-- 新：type/target_type TINYINT；target_ids / publish_at / admin_id

SET @type_is_char := (SELECT COUNT(*)
                      FROM information_schema.COLUMNS
                      WHERE TABLE_SCHEMA = DATABASE()
                        AND TABLE_NAME = 'system_notifications'
                        AND COLUMN_NAME = 'type'
                        AND DATA_TYPE IN ('char', 'varchar', 'tinytext', 'text', 'mediumtext', 'longtext'));
SET @sql := IF(@type_is_char = 0, 'SELECT 1',
  "UPDATE system_notifications SET type = CASE LOWER(TRIM(type))
     WHEN 'announcement' THEN '2'
     WHEN 'maintenance'  THEN '3'
     ELSE '1' END");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @tt_is_char := (SELECT COUNT(*)
                    FROM information_schema.COLUMNS
                    WHERE TABLE_SCHEMA = DATABASE()
                      AND TABLE_NAME = 'system_notifications'
                      AND COLUMN_NAME = 'target_type'
                      AND DATA_TYPE IN ('char', 'varchar', 'tinytext', 'text', 'mediumtext', 'longtext'));
SET @sql := IF(@tt_is_char = 0, 'SELECT 1',
  "UPDATE system_notifications SET target_type = CASE LOWER(TRIM(target_type))
     WHEN 'user_group' THEN '1'
     WHEN 'specific'   THEN '2'
     ELSE '0' END");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

ALTER TABLE system_notifications
    MODIFY COLUMN type TINYINT NOT NULL DEFAULT 1 COMMENT '通知类型（1-系统通知，2-公告，3-维护通知）';

ALTER TABLE system_notifications
    MODIFY COLUMN target_type TINYINT NOT NULL DEFAULT 0 COMMENT '发送对象（0-全部用户，1-指定用户组，2-特定用户）';

ALTER TABLE system_notifications
    MODIFY COLUMN title VARCHAR(200) NOT NULL COMMENT '通知标题';

SET @c := (SELECT COUNT(*) FROM information_schema.COLUMNS
           WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'system_notifications' AND COLUMN_NAME = 'target_users');
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE system_notifications RENAME COLUMN target_users TO target_ids');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (SELECT COUNT(*) FROM information_schema.COLUMNS
           WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'system_notifications' AND COLUMN_NAME = 'push_time');
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE system_notifications RENAME COLUMN push_time TO publish_at');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (SELECT COUNT(*) FROM information_schema.COLUMNS
           WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'system_notifications' AND COLUMN_NAME = 'created_by');
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE system_notifications RENAME COLUMN created_by TO admin_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 六、follows 与 model.Follow 对齐
-- ========================================
SET @c := (SELECT COUNT(*) FROM information_schema.COLUMNS
           WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'follows' AND COLUMN_NAME = 'followed_id');
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE follows RENAME COLUMN followed_id TO following_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (SELECT COUNT(*) FROM information_schema.STATISTICS
           WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'follows' AND INDEX_NAME = 'idx_follows_followed_id');
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE follows RENAME INDEX idx_follows_followed_id TO idx_follows_following_id');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (SELECT COUNT(*) FROM information_schema.COLUMNS
           WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'follows' AND COLUMN_NAME = 'status');
SET @sql := IF(@c > 0, 'SELECT 1',
  "ALTER TABLE follows ADD COLUMN status TINYINT NOT NULL DEFAULT 1 COMMENT '状态（1-正常，0-已取消）' AFTER following_id");
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 七、清除死表 + 补齐活表
-- ========================================
-- 死表：pet_photos / pet_records / topic_follows / notifications / feedback / user_login_logs
-- 补表：user_notifications / announcements / user_activity_logs / admin_login_logs

DROP TRIGGER IF EXISTS tr_topic_follow_add;
DROP TRIGGER IF EXISTS tr_topic_follow_remove;

DROP TABLE IF EXISTS pet_photos;
DROP TABLE IF EXISTS pet_records;
DROP TABLE IF EXISTS topic_follows;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS feedback;
DROP TABLE IF EXISTS user_login_logs;

CREATE TABLE IF NOT EXISTS user_notifications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '通知ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '接收用户ID',
    type VARCHAR(20) NOT NULL COMMENT '通知类型（like/comment/at/system）',
    title VARCHAR(200) COMMENT '通知标题',
    content VARCHAR(500) COMMENT '通知内容',
    trigger_type VARCHAR(20) COMMENT '触发对象类型（post/comment）',
    trigger_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '触发对象ID',
    trigger_user_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '触发用户ID',
    is_read TINYINT NOT NULL DEFAULT 0 COMMENT '是否已读（0-未读，1-已读）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_notifications_user_id (user_id),
    INDEX idx_user_notifications_is_read (is_read),
    INDEX idx_user_notifications_created_at (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户通知表';

CREATE TABLE IF NOT EXISTS announcements (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '公告ID（主键）',
    title VARCHAR(200) NOT NULL COMMENT '公告标题',
    content TEXT NOT NULL COMMENT '公告内容',
    type TINYINT NOT NULL DEFAULT 1 COMMENT '公告类型',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态（0-下线，1-上线）',
    sort_order INT NOT NULL DEFAULT 0 COMMENT '排序（越大越靠前）',
    admin_id BIGINT UNSIGNED NOT NULL COMMENT '发布管理员ID',
    start_at DATETIME COMMENT '生效开始时间',
    end_at DATETIME COMMENT '生效结束时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_announcements_type (type),
    INDEX idx_announcements_status (status),
    INDEX idx_announcements_sort_order (sort_order DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='公告表';

CREATE TABLE IF NOT EXISTS user_activity_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
    action VARCHAR(100) NOT NULL COMMENT '动作',
    target_type VARCHAR(50) COMMENT '目标类型',
    target_id BIGINT UNSIGNED COMMENT '目标ID',
    ip VARCHAR(50) COMMENT 'IP地址',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_activity_logs_user_id (user_id),
    INDEX idx_user_activity_logs_action (action),
    INDEX idx_user_activity_logs_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户活动日志表';

CREATE TABLE IF NOT EXISTS admin_login_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID（主键）',
    admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID',
    username VARCHAR(50) NOT NULL COMMENT '登录用户名',
    ip VARCHAR(50) COMMENT '登录IP',
    user_agent VARCHAR(500) COMMENT '用户代理',
    device VARCHAR(100) COMMENT '设备信息',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态（1-成功，0-失败）',
    fail_reason VARCHAR(255) COMMENT '失败原因',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '登录时间',
    INDEX idx_admin_login_logs_admin_id (admin_id),
    INDEX idx_admin_login_logs_status (status),
    INDEX idx_admin_login_logs_created_at (created_at DESC)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台登录日志表';

-- ========================================
-- 八、users.phone/email 索引与 GORM uniqueIndex 对齐
-- ========================================
-- 旧 create.sql 曾同时有非唯一 idx_users_* 与 UNIQUE uk_users_*，
-- AutoMigrate 再建 UNIQUE idx_users_phone → Error 1061。
-- 注意：若无唯一索引时多次跑 seed 导致 phone/email 重复，建唯一索引会报 1062，
-- 开发库需先去重（保留较小 id）再启服务。

SET @c := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_phone'
    AND NON_UNIQUE = 1
);
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE users DROP INDEX idx_users_phone');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'uk_users_phone'
);
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'idx_users_phone'
);
SET @sql := IF(@has_uk > 0 AND @has_idx = 0,
  'ALTER TABLE users RENAME INDEX uk_users_phone TO idx_users_phone',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'uk_users_phone'
);
SET @sql := IF(@has_uk = 0, 'SELECT 1', 'ALTER TABLE users DROP INDEX uk_users_phone');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @c := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_email'
    AND NON_UNIQUE = 1
);
SET @sql := IF(@c = 0, 'SELECT 1', 'ALTER TABLE users DROP INDEX idx_users_email');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'uk_users_email'
);
SET @has_idx := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'idx_users_email'
);
SET @sql := IF(@has_uk > 0 AND @has_idx = 0,
  'ALTER TABLE users RENAME INDEX uk_users_email TO idx_users_email',
  'SELECT 1');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @has_uk := (
  SELECT COUNT(*) FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND INDEX_NAME = 'uk_users_email'
);
SET @sql := IF(@has_uk = 0, 'SELECT 1', 'ALTER TABLE users DROP INDEX uk_users_email');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- ========================================
-- 校验（可选）
-- SELECT COLUMN_NAME FROM information_schema.COLUMNS
-- WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'messages' AND COLUMN_NAME = 'duration';
-- SELECT TABLE_NAME FROM information_schema.TABLES
-- WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME IN (
--   'pet_photos','notifications','user_notifications','announcements','admin_login_logs'
-- );
-- SELECT INDEX_NAME, NON_UNIQUE FROM information_schema.STATISTICS
-- WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users'
--   AND INDEX_NAME IN ('idx_users_phone','idx_users_email','uk_users_phone','uk_users_email');
