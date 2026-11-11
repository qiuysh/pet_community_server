-- ============================================================
-- 2026-10-19 ~ 2026-11-05 schema 收口合集（幂等，可重复执行）
-- 路径: server/sql/incremental/20261105000000_schema_fix_through_20261105.sql
-- 整合日期: 2026-11-11
--
-- 本文件由以下已提交合集按时间序整合而成（各合集正文逐字保留，仅统一总头部）：
--   20261019154000_schema_fix_20261019.sql  → 卷一
--   20261020113000_fix_20261020.sql         → 卷二
--   20261105000000_schema_fix_20261105.sql  → 卷三（内含当日 8 节）
--   （原分文件 / 分日合集已删除；曾按旧文件名执行过的库，本合集为幂等空跑，可跳过。）
--
-- 适用对象：只跑 init/create.sql 建库的存量环境（导库 / QA / BI 只读库）。
--           全新库**无需执行**：以下变更已全部合并进 init/create.sql
--           （卷二第二节为存量 URL 前缀改写，新库无旧前缀可跳过或空跑）。
-- 执行要求：卷三含 DELIMITER 块，请在支持该指令的客户端中整文件执行。
-- ============================================================


-- ################################################################
-- 卷一：2026-10-19 结构补丁合集
-- （原 20261019154000_schema_fix_20261019.sql）
-- ################################################################

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

-- ################################################################
-- 卷二：2026-10-20 结构 / 数据补丁合集
-- （原 20261020113000_fix_20261020.sql）
-- ################################################################

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

-- ################################################################
-- 卷三：2026-11-05 schema 收口合集
-- （原 20261105000000_schema_fix_20261105.sql）
-- ################################################################

-- ============================================================
-- 2026-11-05 schema 收口合集（幂等，可重复执行）
-- 路径: server/sql/incremental/20261105000000_schema_fix_20261105.sql
-- 创建日期: 2026-11-05 | 整合日期: 2026-11-06
--
-- 本文件由当日 8 个增量脚本按时间序整合而成（内容未改动，仅统一头部与编号）：
--   113000 add_topic_shares_count        → 第一节
--   161000 drop_counter_triggers         → 第二节
--   161100 recompute_counter_columns     → 第三节
--   163000 fix_topics_category_id        → 第四节
--   163100 fix_stats_status_semantics    → 第五节
--   170000 add_missing_columns           → 第六节
--   170100 fix_v_user_stats_cumulative   → 第七节
--   172000 normalize_unique_indexes      → 第八节
--   （原有 8 个分文件已删除；曾按分文件执行过的库，本合集为幂等空跑，直接跳过即可。）
--
-- 适用对象：只跑 init/create.sql 建库的存量环境（导库 / QA / BI 只读库）。
--           全新库**无需执行**：以下变更已全部合并进 init/create.sql。
-- 执行要求：第五、八节含 DELIMITER，请在支持该指令的客户端（mysql 命令行 /
--           Navicat / DataGrip）中整文件执行，不要按分号逐句拆跑。
-- 依赖版本：MySQL 8.0.4+（第七节用到窗口函数套聚合）。
-- ============================================================

USE pet_community;



-- ############################################################
-- 一、话题分享数：topics.shares_count
-- （原 20261105113000_add_topic_shares_count.sql）
-- ############################################################

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


-- ############################################################
-- 二、下线 6 个计数触发器
-- （原 20261105161000_drop_counter_triggers.sql）
-- ############################################################

-- ========================================
-- 下线计数触发器（幂等）
-- 日期：2026-11-05
-- 背景：计数列统一由 Go 服务层事务维护，触发器会造成同一次写操作 +2。
--       本脚本可重复执行：DROP TRIGGER IF EXISTS 天然幂等。
-- 影响：删除后 likes_count / comments_count / reply_count / votes_count 仅由 Go 侧更新。
--       执行本节后请继续执行下一节（重算聚合计数列）修正历史虚高值。
-- ========================================


DROP TRIGGER IF EXISTS tr_like_add;
DROP TRIGGER IF EXISTS tr_like_remove;
DROP TRIGGER IF EXISTS tr_comment_add;
DROP TRIGGER IF EXISTS tr_comment_remove;
DROP TRIGGER IF EXISTS tr_vote_add;
DROP TRIGGER IF EXISTS tr_vote_remove;

-- 校验：确认计数表上已无残留触发器。
-- 历史库若还建过 favorites / shares / follows 之类的计数触发器，本脚本不会自动删（怕误删业务触发器），
-- 需要人工确认后 DROP；否则重算之后再点赞仍会 +2。
SELECT TRIGGER_NAME, EVENT_OBJECT_TABLE, EVENT_MANIPULATION, ACTION_TIMING
FROM information_schema.TRIGGERS
WHERE TRIGGER_SCHEMA = DATABASE()
  AND EVENT_OBJECT_TABLE IN ('likes', 'comments', 'activity_votes', 'favorites', 'share_records', 'posts')
ORDER BY EVENT_OBJECT_TABLE, TRIGGER_NAME;


-- ############################################################
-- 三、重算聚合计数列（修历史虚高）
-- （原 20261105161100_recompute_counter_columns.sql）
-- ############################################################

-- ========================================
-- 重算聚合计数列（幂等）
-- 日期：2026-11-05
-- 用法：在上一节（下线计数触发器）之后执行，可重复执行。
--
-- 语义：所有计数从明细表全量重算，结果与 Go 服务层的写入口径一致：
--       posts.comments_count  只统计顶级评论（parent_id 为 0 / NULL），与 CommentService.CreateComment 一致；
--       comments.reply_count  统计直接子回复；
--       不按 status 过滤（与 Go 侧「创建即计数」口径一致，避免与线上实时计数出现偏差）；
--       其余按明细行数。
--       话题分享数 topics.shares_count 无明细表，无法重算，不在本脚本范围内。
--
-- 幂等：纯重算，不累加；重复执行结果相同。
--
-- 修订（ERROR 1093 / 1051）：原 `UPDATE comments c SET c.reply_count = (SELECT ... FROM comments r ...)`
--       在子查询 FROM 中引用了正在更新的目标表，MySQL 直接拒绝（1093）；
--       改用时又踩到临时表在某些客户端/代理下不可靠（1051、跨会话 1146）。
--       最终统一为「聚合派生表 + LEFT JOIN 更新」——这是 MySQL 官方对 1093 的推荐绕法：
--       目标表在 UPDATE 的 JOIN 里只作为被更新对象，同表出现在聚合派生表中会被物化，
--       属于多表 UPDATE 的合法形式，无状态、无临时表、可被任意客户端拆句执行。
--       顺带把相关子查询（逐行扫明细）改为一次聚合，大表下更快。
--
-- 注意：dev seed（20260724100000_seed_dev_data.sql）只写了聚合值、没有明细行，
--       执行后这些帖子的点赞/收藏数会归 0 —— 这是「明细为准」的正确结果，如需演示数据请重跑 seed 并补明细。
-- ========================================


-- ---------- posts：点赞 / 评论 / 收藏 / 分享 ----------

UPDATE posts p
LEFT JOIN (
    SELECT target_id AS post_id, COUNT(*) AS cnt
    FROM likes
    WHERE target_type = 'post'
    GROUP BY target_id
) t ON t.post_id = p.id
SET p.likes_count = COALESCE(t.cnt, 0);

UPDATE posts p
LEFT JOIN (
    SELECT post_id, COUNT(*) AS cnt
    FROM comments
    -- 顶级评论 = parent_id 为 0（哨兵）或 NULL；写成 OR 而不是 IFNULL()，避免函数包裹导致索引失效
    WHERE parent_id = 0 OR parent_id IS NULL
    GROUP BY post_id
) t ON t.post_id = p.id
SET p.comments_count = COALESCE(t.cnt, 0);

UPDATE posts p
LEFT JOIN (
    SELECT post_id, COUNT(*) AS cnt
    FROM favorites
    GROUP BY post_id
) t ON t.post_id = p.id
SET p.favorites_count = COALESCE(t.cnt, 0);

UPDATE posts p
LEFT JOIN (
    SELECT post_id, COUNT(*) AS cnt
    FROM share_records
    GROUP BY post_id
) t ON t.post_id = p.id
SET p.shares_count = COALESCE(t.cnt, 0);

-- ---------- comments：点赞 ----------

UPDATE comments c
LEFT JOIN (
    SELECT target_id AS comment_id, COUNT(*) AS cnt
    FROM likes
    WHERE target_type = 'comment'
    GROUP BY target_id
) t ON t.comment_id = c.id
SET c.likes_count = COALESCE(t.cnt, 0);

-- ---------- comments：回复数（自引用，走派生表绕开 1093） ----------
-- 聚合派生表会被物化，因而「更新 comments 同时从 comments 聚合」是合法的多表 UPDATE；
-- 没有子回复的评论靠 LEFT JOIN + COALESCE 归 0，一次扫描全表即可。

UPDATE comments c
LEFT JOIN (
    SELECT parent_id, COUNT(*) AS cnt
    FROM comments
    WHERE parent_id > 0
    GROUP BY parent_id
) r ON r.parent_id = c.id
SET c.reply_count = COALESCE(r.cnt, 0);

-- ---------- activity_submissions：投票数 ----------

UPDATE activity_submissions s
LEFT JOIN (
    SELECT submission_id, COUNT(*) AS cnt
    FROM activity_votes
    GROUP BY submission_id
) t ON t.submission_id = s.id
SET s.votes_count = COALESCE(t.cnt, 0);

-- ---------- 执行后校验（只读）：aggregate_total 与 detail_total 应逐行相等 ----------
-- 若不等，通常是明细表里有「孤儿行」（如评论已删但 likes 记录还在），需要人工确认后再决定是否清孤儿。
-- 不需要校验时，把这一段整体注释掉即可。

SELECT 'posts.likes_count' AS metric,
       (SELECT COALESCE(SUM(likes_count), 0) FROM posts) AS aggregate_total,
       (SELECT COUNT(*) FROM likes WHERE target_type = 'post') AS detail_total
UNION ALL
SELECT 'posts.comments_count',
       (SELECT COALESCE(SUM(comments_count), 0) FROM posts),
       (SELECT COUNT(*) FROM comments WHERE parent_id = 0 OR parent_id IS NULL)
UNION ALL
SELECT 'posts.favorites_count',
       (SELECT COALESCE(SUM(favorites_count), 0) FROM posts),
       (SELECT COUNT(*) FROM favorites)
UNION ALL
SELECT 'posts.shares_count',
       (SELECT COALESCE(SUM(shares_count), 0) FROM posts),
       (SELECT COUNT(*) FROM share_records)
UNION ALL
SELECT 'comments.likes_count',
       (SELECT COALESCE(SUM(likes_count), 0) FROM comments),
       (SELECT COUNT(*) FROM likes WHERE target_type = 'comment')
UNION ALL
SELECT 'comments.reply_count',
       (SELECT COALESCE(SUM(reply_count), 0) FROM comments),
       (SELECT COUNT(*) FROM comments WHERE parent_id > 0)
UNION ALL
SELECT 'activity_submissions.votes_count',
       (SELECT COALESCE(SUM(votes_count), 0) FROM activity_submissions),
       (SELECT COUNT(*) FROM activity_votes);


-- ############################################################
-- 四、topics.category_id：NULL 归 0 哨兵
-- （原 20261105163000_fix_topics_category_id.sql）
-- ############################################################

-- ========================================
-- 修正 topics.category_id：可空 + 外键 → 0 哨兵（幂等）
-- 日期：2026-11-05
--
-- 背景：
--   线上现象是「新建话题后，话题详情/后台待审列表 500」。
--   两个原因叠加：TopicRepository.Create 的 Select 白名单漏了 CategoryID；
--   而 topics.category_id 在 create.sql 里是「可空、无默认值、带外键 ON DELETE SET NULL」。
--   于是 INSERT 不写该列 → 落 NULL → 之后 SELECT * 把 NULL 扫进 model.Topic.CategoryID（uint）
--   直接报 converting NULL to uint is unsupported。
--
-- 约定：category_id 改用 0 作为「未分类」哨兵（与 pets.breed_id 一致），并去掉外键
--       —— 0 不存在于 topic_categories(id)，保留外键会在新建话题时报 1452。
--       分类有效性由业务层在 CategoryID > 0 时校验。
--
-- 幂等：UPDATE 只动 NULL；外键名动态查找；MODIFY 重复执行结果相同。
-- 用法：存量库执行本脚本；全新库直接用 init/create.sql（已同步）。
-- ========================================


-- 1) 存量 NULL 归 0（必须先做，否则下一步 MODIFY NOT NULL 会失败）
UPDATE topics SET category_id = 0 WHERE category_id IS NULL;

-- 2) 动态删除 category_id 上的外键（create.sql 里未命名，MySQL 自动生成 topics_ibfk_N）
SET @fk_name := (
    SELECT CONSTRAINT_NAME
    FROM information_schema.KEY_COLUMN_USAGE
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'topics'
      AND COLUMN_NAME = 'category_id'
      AND REFERENCED_TABLE_NAME IS NOT NULL
    LIMIT 1
);

SET @drop_fk := IF(
    @fk_name IS NULL,
    -- 空操作：SET 一定在 PREPARE 支持列表内（不用 DO 0，避免个别客户端/代理不认）
    'SET @noop := 0',
    CONCAT('ALTER TABLE topics DROP FOREIGN KEY `', @fk_name, '`')
);

PREPARE drop_fk_stmt FROM @drop_fk;
EXECUTE drop_fk_stmt;
DEALLOCATE PREPARE drop_fk_stmt;

-- 3) 列改为 NOT NULL DEFAULT 0
ALTER TABLE topics
    MODIFY COLUMN category_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '分类ID（0=未分类）';


-- ############################################################
-- 五、统计视图/存储过程 status 口径
-- （原 20261105163100_fix_stats_status_semantics.sql）
-- ############################################################

-- ========================================
-- 修正统计视图/存储过程的 status 口径（幂等）
-- 日期：2026-11-05
--
-- 问题：SQL 侧报表对象的 status 口径与 Go 代码相反，导致导出的活跃/封禁、已发布/已拒绝全错。
--   users.status  ：代码 1=正常、0=禁用/未激活、2=封禁（UserService 登录校验 + 后台封禁写 2）
--                   旧脚本误写成 0=活跃、1=封禁。
--   posts.status  ：代码 0=待审核、1=已发布、2=定时待发、3=已拒绝（service.PostStatus* 常量）
--                   旧脚本把 2 算成「已拒绝」，且完全没统计 3；定时帖被误判为被拒。
--
-- 说明：后台页面（驾驶舱等）走 Go 查询，口径本来就是对的；本脚本修的是 SQL 视图/存储过程
--       这类「DBA、BI、导出工具」会直接使用的对象。
-- 幂等：DROP ... IF EXISTS 后重建，可重复执行。
-- 注意：存储过程需要 DELIMITER，请在支持该指令的客户端（命令行 / Navicat / DataGrip）执行。
-- ========================================


DROP VIEW IF EXISTS v_user_stats;
CREATE VIEW v_user_stats AS
SELECT
    COUNT(*) as total_users,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_users,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as banned_users,
    DATE_FORMAT(created_at, '%Y-%m-%d') as date,
    COUNT(*) OVER (ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d')) as cumulative_users
FROM users
GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d');

DROP VIEW IF EXISTS v_post_stats;
CREATE VIEW v_post_stats AS
SELECT
    COUNT(*) as total_posts,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as published_posts,
    SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as pending_posts,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as scheduled_posts,
    SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END) as rejected_posts,
    SUM(likes_count) as total_likes,
    SUM(comments_count) as total_comments,
    SUM(shares_count) as total_shares,
    DATE_FORMAT(created_at, '%Y-%m-%d') as date
FROM posts
GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d');

DROP PROCEDURE IF EXISTS GetUserStats;
DROP PROCEDURE IF EXISTS GetPostStats;

DELIMITER //

CREATE PROCEDURE GetUserStats(IN start_date DATE, IN end_date DATE)
BEGIN
    SELECT
        COUNT(*) as total_users,
        SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_users,
        SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as banned_users
    FROM users
    WHERE created_at BETWEEN start_date AND DATE_ADD(end_date, INTERVAL 1 DAY);
END //

CREATE PROCEDURE GetPostStats(IN start_date DATE, IN end_date DATE)
BEGIN
    SELECT
        COUNT(*) as total_posts,
        SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as published_posts,
        SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as pending_posts,
        SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as scheduled_posts,
        SUM(CASE WHEN status = 3 THEN 1 ELSE 0 END) as rejected_posts,
        SUM(likes_count) as total_likes,
        SUM(comments_count) as total_comments,
        SUM(shares_count) as total_shares
    FROM posts
    WHERE created_at BETWEEN start_date AND DATE_ADD(end_date, INTERVAL 1 DAY);
END //

DELIMITER ;


-- ############################################################
-- 六、补齐模型与 create.sql 的列漂移（8 列）
-- （原 20261105170000_add_missing_columns.sql）
-- ############################################################

-- ========================================
-- 补齐 create.sql 与 Go 模型之间的列漂移（幂等）
-- 日期：2026-11-05
--
-- 背景：
--   create.sql 落后于 model/*.go，导致「只跑 SQL 建库」的环境（导库、QA、BI 只读库）
--   在业务读取时直接 1054 Unknown column：
--     users.token_version / failed_login_count / locked_until  → 登录、鉴权、封禁失效逻辑
--     admin_users.failed_login_count / locked_until            → 后台登录锁定
--     topics.creator_id                                        → 话题创建（Create 白名单显式写该列）
--     activity_registrations.priority                           → 报名优先排队
--     pet_albums.media_type                                     → 相册视频（上传/列表读取）
--   通过 Go 启动（AutoMigrate）的实例不受影响，本脚本对齐纯 SQL 环境。
--
-- 幂等：每列先查 information_schema.COLUMNS，缺失才 ADD；索引同理查 STATISTICS。
--       MySQL 8.0 的 ADD COLUMN 不支持 IF NOT EXISTS，故用 PREPARE 动态执行。
-- 用法：存量库执行本脚本；全新库直接用 init/create.sql（已同步）。
-- ========================================


-- ---------- users ----------
SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'token_version');
SET @ddl := IF(@has = 0,
    'ALTER TABLE users ADD COLUMN token_version INT NOT NULL DEFAULT 0 COMMENT ''Token 版本（递增使旧 token 全部失效）'' AFTER level',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'failed_login_count');
SET @ddl := IF(@has = 0,
    'ALTER TABLE users ADD COLUMN failed_login_count INT NOT NULL DEFAULT 0 COMMENT ''连续登录失败次数'' AFTER token_version',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'users' AND COLUMN_NAME = 'locked_until');
SET @ddl := IF(@has = 0,
    'ALTER TABLE users ADD COLUMN locked_until DATETIME NULL COMMENT ''登录锁定截止时间（NULL=未锁定）'' AFTER failed_login_count',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------- admin_users ----------
SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_users' AND COLUMN_NAME = 'failed_login_count');
SET @ddl := IF(@has = 0,
    'ALTER TABLE admin_users ADD COLUMN failed_login_count INT NOT NULL DEFAULT 0 COMMENT ''连续登录失败次数'' AFTER status',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_users' AND COLUMN_NAME = 'locked_until');
SET @ddl := IF(@has = 0,
    'ALTER TABLE admin_users ADD COLUMN locked_until DATETIME NULL COMMENT ''登录锁定截止时间（NULL=未锁定）'' AFTER failed_login_count',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------- topics.creator_id ----------
SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'topics' AND COLUMN_NAME = 'creator_id');
SET @ddl := IF(@has = 0,
    'ALTER TABLE topics ADD COLUMN creator_id BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT ''创建者ID（0=系统/运营）'' AFTER category_id',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has := (SELECT COUNT(*) FROM information_schema.STATISTICS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'topics' AND INDEX_NAME = 'idx_topics_creator_id');
SET @ddl := IF(@has = 0,
    'ALTER TABLE topics ADD INDEX idx_topics_creator_id (creator_id)',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------- activity_registrations.priority ----------
SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'activity_registrations' AND COLUMN_NAME = 'priority');
SET @ddl := IF(@has = 0,
    'ALTER TABLE activity_registrations ADD COLUMN priority TINYINT NOT NULL DEFAULT 0 COMMENT ''报名优先级（1=高等级用户优先排队，0=普通）'' AFTER status',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

SET @has := (SELECT COUNT(*) FROM information_schema.STATISTICS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'activity_registrations'
               AND INDEX_NAME = 'idx_activity_registrations_priority');
SET @ddl := IF(@has = 0,
    'ALTER TABLE activity_registrations ADD INDEX idx_activity_registrations_priority (priority)',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------- pet_albums.media_type ----------
SET @has := (SELECT COUNT(*) FROM information_schema.COLUMNS
             WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'pet_albums' AND COLUMN_NAME = 'media_type');
SET @ddl := IF(@has = 0,
    'ALTER TABLE pet_albums ADD COLUMN media_type VARCHAR(20) NOT NULL DEFAULT ''image'' COMMENT ''媒体类型（image/video）'' AFTER image_url',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;

-- ---------- 执行后校验：应返回 0 行 ----------
SELECT t.table_name, t.column_name
FROM (
    SELECT 'users' AS table_name, 'token_version' AS column_name
    UNION ALL SELECT 'users', 'failed_login_count'
    UNION ALL SELECT 'users', 'locked_until'
    UNION ALL SELECT 'admin_users', 'failed_login_count'
    UNION ALL SELECT 'admin_users', 'locked_until'
    UNION ALL SELECT 'topics', 'creator_id'
    UNION ALL SELECT 'activity_registrations', 'priority'
    UNION ALL SELECT 'pet_albums', 'media_type'
) t
WHERE NOT EXISTS (
    SELECT 1 FROM information_schema.COLUMNS c
    WHERE c.TABLE_SCHEMA = DATABASE()
      AND c.TABLE_NAME = t.table_name
      AND c.COLUMN_NAME = t.column_name
);


-- ############################################################
-- 七、v_user_stats.cumulative_users 语义修正
-- （原 20261105170100_fix_v_user_stats_cumulative.sql）
-- ############################################################

-- ========================================
-- 修正 v_user_stats.cumulative_users 语义（幂等：DROP + CREATE）
-- 日期：2026-11-05
--
-- 问题：
--   旧定义 `COUNT(*) OVER (ORDER BY DATE_FORMAT(created_at,'%Y-%m-%d'))`
--   在 GROUP BY（按天分组）之后开窗，COUNT(*) 取的是「当天新增数」，
--   窗口帧却是「从第一行到当前行」→ 结果是行号（累计天数），不是累计用户数。
--   例：连续 3 天每天新增 1 人，旧值 1/2/3，正确值 1/2/3 恰好相同；
--       某天新增 5 人后，旧值下一行只 +1，正确值应 +5，此时才暴露。
--
-- 修法：
--   窗口函数套聚合 `SUM(COUNT(*)) OVER (ORDER BY 日期)` —— MySQL 8.0 允许
--   「聚合函数作为窗口函数的参数」，先按天聚合、再按日期累加。
--   8.0 的视图不支持 FROM 子句里放派生表（ERROR 1349），因此不能用「先聚合成
--   临时结果再开窗」的常规写法，只能走这条语法。
--
-- 兼容性：需要 8.0.4+（窗口函数）。执行完请跑文件末尾的校验查询确认视图可用。
--        若个别 8.0.x 报 3593（窗口/聚合组合不支持），回滚方式见文末注释。
-- 幂等：DROP VIEW IF EXISTS + CREATE，可重复执行。
-- ========================================


DROP VIEW IF EXISTS v_user_stats;
CREATE VIEW v_user_stats AS
SELECT
    COUNT(*) as total_users,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_users,
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as banned_users,
    DATE_FORMAT(created_at, '%Y-%m-%d') as date,
    SUM(COUNT(*)) OVER (ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d')) as cumulative_users
FROM users
GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d');

-- ---------- 校验：应能正常返回，且最后一行 cumulative_users = 用户总数 ----------
SELECT date, total_users, cumulative_users
FROM v_user_stats
ORDER BY date;

-- 回滚（若上面 CREATE VIEW 报错，取消下面注释后整段执行，恢复旧定义避免视图缺失）：
-- DROP VIEW IF EXISTS v_user_stats;
-- CREATE VIEW v_user_stats AS
-- SELECT
--     COUNT(*) as total_users,
--     SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as active_users,
--     SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as banned_users,
--     DATE_FORMAT(created_at, '%Y-%m-%d') as date,
--     COUNT(*) OVER (ORDER BY DATE_FORMAT(created_at, '%Y-%m-%d')) as cumulative_users
-- FROM users
-- GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d');


-- ############################################################
-- 八、唯一键重复索引归一
-- （原 20261105172000_normalize_unique_indexes.sql）
-- ############################################################

-- ========================================
-- 收敛唯一键的重复索引（幂等）
-- 日期：2026-11-05
--
-- 背景：
--   1) create.sql 里同时声明了 UNIQUE KEY uk_x (col) 与 INDEX idx_x (col)，
--      普通索引完全被唯一键覆盖，属于纯写入开销。
--   2) 模型长期用不带名字的 gorm:"unique"，AutoMigrate 会按 GORM 约定再建一套
--      uni_<table>_<col> 唯一约束；与 create.sql 的 uk_<table>_<col> 同列重复，
--      跑过 AutoMigrate 的库里同列可能有三份索引（uk_ + uni_ + idx_）。
--   3) 现模型已改为显式 uniqueIndex:uk_*（activity_reviews 沿用 SQL 里的 activity_id），
--      本脚本把存量库的命名一并对齐，避免 AutoMigrate 继续新增重复唯一键。
--
-- 处理顺序（每张表）：
--   有 uni_ 且无 uk_ → RENAME uni_ 为 uk_（保留唯一性）
--   有 uni_ 且有 uk_ → DROP uni_
--   有 idx_ 且有 uk_ → DROP idx_（唯一键已覆盖该列）
--
-- 用法：存量库执行一次；全新库直接用 init/create.sql（已同步，不再有 INDEX idx_*）。
-- 说明：MySQL 8.0 的 DROP INDEX / RENAME INDEX 都不支持 IF EXISTS，因此统一查
--       information_schema.STATISTICS 后动态执行。
-- ========================================


DROP PROCEDURE IF EXISTS normalize_unique_index;

DELIMITER //
CREATE PROCEDURE normalize_unique_index(
    IN tbl VARCHAR(64),
    IN legacy_idx VARCHAR(64),
    IN legacy_uni VARCHAR(64),
    IN canon VARCHAR(64)
)
BEGIN
    DECLARE has_canon INT DEFAULT 0;
    DECLARE has_legacy_idx INT DEFAULT 0;
    DECLARE has_legacy_uni INT DEFAULT 0;

    SELECT COUNT(*) INTO has_canon FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl AND INDEX_NAME = canon;
    SELECT COUNT(*) INTO has_legacy_idx FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl AND INDEX_NAME = legacy_idx;
    SELECT COUNT(*) INTO has_legacy_uni FROM information_schema.STATISTICS
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = tbl AND INDEX_NAME = legacy_uni;

    -- 只有 uni_ 时改名为规范名，保证唯一约束不丢
    IF has_legacy_uni > 0 AND has_canon = 0 THEN
        SET @ddl = CONCAT('ALTER TABLE `', tbl, '` RENAME INDEX `', legacy_uni, '` TO `', canon, '`');
        PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
        SET has_canon = 1;
        SET has_legacy_uni = 0;
    END IF;

    IF has_legacy_uni > 0 AND has_canon > 0 THEN
        SET @ddl = CONCAT('ALTER TABLE `', tbl, '` DROP INDEX `', legacy_uni, '`');
        PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;

    IF has_legacy_idx > 0 AND has_canon > 0 THEN
        SET @ddl = CONCAT('ALTER TABLE `', tbl, '` DROP INDEX `', legacy_idx, '`');
        PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
    END IF;
END //
DELIMITER ;

CALL normalize_unique_index('topics', 'idx_topics_name', 'uni_topics_name', 'uk_topics_name');
CALL normalize_unique_index('points_rules', 'idx_points_rules_type', 'uni_points_rules_type', 'uk_points_rules_type');
CALL normalize_unique_index('hot_keywords', 'idx_hot_keywords_keyword', 'uni_hot_keywords_keyword', 'uk_hot_keywords_keyword');
CALL normalize_unique_index('admin_users', 'idx_admin_users_username', 'uni_admin_users_username', 'uk_admin_users_username');
CALL normalize_unique_index('sensitive_words', 'idx_sensitive_words_word', 'uni_sensitive_words_word', 'uk_sensitive_words_word');
CALL normalize_unique_index('system_config', 'idx_system_config_config_key', 'uni_system_config_config_key', 'uk_system_config_config_key');
CALL normalize_unique_index('topic_categories', '', 'uni_topic_categories_name', 'uk_topic_categories_name');
-- activity_reviews 的规范索引名沿用建表脚本里的列名 activity_id
CALL normalize_unique_index('activity_reviews', 'idx_activity_reviews_activity_id', 'uni_activity_reviews_activity_id', 'activity_id');

DROP PROCEDURE normalize_unique_index;

-- ---------- admin_users.deleted_at ----------
-- 该列为历史残留（当前 Delete 为硬删除，代码零引用），只为审计兼容保留列；
-- 对应的 idx_admin_users_deleted_at 永远是同一批 NULL 值，没有查询价值，直接删除。
SET @has_idx := (SELECT COUNT(*) FROM information_schema.STATISTICS
                 WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'admin_users'
                   AND INDEX_NAME = 'idx_admin_users_deleted_at');
SET @ddl := IF(@has_idx > 0,
    'ALTER TABLE admin_users DROP INDEX idx_admin_users_deleted_at',
    'SET @noop := 0');
PREPARE stmt FROM @ddl; EXECUTE stmt; DEALLOCATE PREPARE stmt;
