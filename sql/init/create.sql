-- ========================================
-- [初始化] 数据库建库建表 + 系统基础数据
-- 路径: server/sql/init/create.sql
-- 适用: 全新环境库一次性执行
-- 版本: v1.3
-- 创建日期: 2026-07-24
-- 更新: 2026-10-19 清除死表（pet_photos/pet_records/topic_follows/notifications/feedback/user_login_logs），
--       补齐活表（user_notifications/announcements/user_activity_logs/admin_login_logs）
-- ========================================

CREATE DATABASE IF NOT EXISTS pet_community 
  CHARACTER SET utf8mb4 
  COLLATE utf8mb4_unicode_ci;

USE pet_community;

-- ========================================
-- 用户表（users）
-- ========================================
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '用户ID（主键）',
    phone VARCHAR(20) COMMENT '手机号',
    email VARCHAR(100) COMMENT '邮箱',
    password VARCHAR(255) NOT NULL COMMENT '密码（BCrypt加密）',
    nickname VARCHAR(50) NOT NULL COMMENT '昵称',
    avatar VARCHAR(255) COMMENT '头像URL',
    signature VARCHAR(100) COMMENT '个性签名',
    gender TINYINT DEFAULT 0 COMMENT '性别（0-保密，1-男，2-女）',
    location VARCHAR(100) COMMENT '所在地',
    pet_years VARCHAR(20) COMMENT '养宠年限',
    privacy_settings JSON DEFAULT (JSON_OBJECT('visibility', 'public', 'commentPermission', 'all', 'atNotification', true, 'messagePermission', 'all', 'recommendEnabled', true)) COMMENT '隐私设置',
    status TINYINT DEFAULT 0 COMMENT '状态（0-正常，1-封禁）',
    points INT DEFAULT 0 COMMENT '积分',
    level TINYINT DEFAULT 1 COMMENT '等级（1-10级）',
    last_login_at DATETIME COMMENT '最后登录时间',
    last_login_ip VARCHAR(50) COMMENT '最后登录IP',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    -- phone/email 唯一索引名与 model.User uniqueIndex（GORM 默认 idx_users_phone/email）对齐，
    -- 勿再额外建同名非唯一 INDEX，否则 AutoMigrate 会报 Error 1061 Duplicate key name
    UNIQUE KEY idx_users_phone (phone),
    UNIQUE KEY idx_users_email (email),
    INDEX idx_users_status (status),
    INDEX idx_users_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- ========================================
-- 用户设备表（user_devices）
-- ========================================
CREATE TABLE IF NOT EXISTS user_devices (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '设备ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    device_id VARCHAR(100) NOT NULL COMMENT '设备标识符',
    device_name VARCHAR(100) COMMENT '设备名称',
    device_type VARCHAR(20) COMMENT '设备类型',
    ip VARCHAR(50) COMMENT 'IP地址',
    last_active_at DATETIME COMMENT '最后活跃时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_devices_user_id (user_id),
    INDEX idx_user_devices_device_id (device_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户设备表';

-- ========================================
-- 验证码表（verification_codes）
-- ========================================
CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '验证码ID（主键）',
    account VARCHAR(100) NOT NULL COMMENT '账号（手机号/邮箱）',
    code VARCHAR(10) NOT NULL COMMENT '验证码',
    type VARCHAR(20) NOT NULL COMMENT '类型（login/register/reset_password）',
    expires_at DATETIME NOT NULL COMMENT '过期时间',
    used TINYINT DEFAULT 0 COMMENT '是否已使用（0-未使用，1-已使用）',
    ip VARCHAR(50) COMMENT 'IP地址',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_verification_codes_account (account),
    INDEX idx_verification_codes_type (type),
    INDEX idx_verification_codes_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='验证码表';

-- ========================================
-- 第三方账号关联表（user_third_accounts）
-- ========================================
CREATE TABLE IF NOT EXISTS user_third_accounts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    platform VARCHAR(20) NOT NULL COMMENT '平台（wechat/weibo/qq）',
    open_id VARCHAR(100) NOT NULL COMMENT 'OpenID',
    union_id VARCHAR(100) COMMENT 'UnionID',
    nickname VARCHAR(50) COMMENT '昵称',
    avatar VARCHAR(255) COMMENT '头像',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_third_accounts_user_id (user_id),
    INDEX idx_user_third_accounts_platform_openid (platform, open_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='第三方账号关联表';

-- ========================================
-- 宠物品种表（pet_breeds）
-- ========================================
CREATE TABLE IF NOT EXISTS pet_breeds (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '品种ID（主键）',
    pet_type TINYINT NOT NULL COMMENT '宠物类型（1-狗，2-猫）',
    name VARCHAR(50) NOT NULL COMMENT '品种名称',
    english_name VARCHAR(100) COMMENT '英文名称',
    avatar VARCHAR(255) COMMENT '品种图标',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    status TINYINT DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    INDEX idx_pet_breeds_pet_type (pet_type),
    INDEX idx_pet_breeds_sort_order (sort_order),
    INDEX idx_pet_breeds_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='宠物品种表';

-- ========================================
-- 宠物表（pets）
-- ========================================
CREATE TABLE IF NOT EXISTS pets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '宠物ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    name VARCHAR(50) NOT NULL COMMENT '宠物名称',
    avatar VARCHAR(255) COMMENT '宠物头像',
    pet_type TINYINT NOT NULL COMMENT '宠物类型（1-狗，2-猫）',
    breed_id BIGINT UNSIGNED DEFAULT 0 COMMENT '品种ID（0=未选品种）',
    breed_name VARCHAR(50) COMMENT '品种名称',
    gender TINYINT DEFAULT 0 COMMENT '性别（0-未知，1-公，2-母）',
    birthday DATE COMMENT '生日',
    weight DECIMAL(5,2) COMMENT '体重（kg）',
    personality VARCHAR(100) COMMENT '性格描述',
    bio VARCHAR(500) COMMENT '简介',
    is_sterilized TINYINT DEFAULT 0 COMMENT '是否绝育（0-否，1-是）',
    is_vaccinated TINYINT DEFAULT 0 COMMENT '是否疫苗齐全（0-否，1-是）',
    tags VARCHAR(255) COMMENT '标签（JSON数组）',
    description VARCHAR(200) COMMENT '描述',
    status TINYINT DEFAULT 1 COMMENT '状态（0-正常，1-生病，2-走失，3-去世）',
    view_count INT DEFAULT 0 COMMENT '浏览次数',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_pets_user_id (user_id),
    INDEX idx_pets_status (status),
    INDEX idx_pets_pet_type (pet_type),
    -- breed_id 用 0 作为「未选品种」的哨兵值（Pet.BreedID 为 uint，ORM 零值即 0），
    -- 因此不能加外键：0 不存在于 pet_breeds(id)，创建宠物不选品种时会报 1452；
    -- 品种有效性由 PetService 在 BreedID > 0 时调用 GetBreedByID 校验，项目暂无删除品种的功能。
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='宠物表';

-- ========================================
-- 宠物相册表（pet_albums）
-- ========================================
CREATE TABLE IF NOT EXISTS pet_albums (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '相册ID（主键）',
    pet_id BIGINT UNSIGNED NOT NULL COMMENT '宠物ID（外键）',
    image_url VARCHAR(255) NOT NULL COMMENT '照片URL',
    thumb_url VARCHAR(255) COMMENT '缩略图URL',
    title VARCHAR(100) COMMENT '照片标题',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_pet_albums_pet_id (pet_id),
    INDEX idx_pet_albums_sort_order (sort_order),
    FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='宠物相册表';

-- ========================================
-- 宠物成长记录表（pet_growth_records）
-- ========================================
CREATE TABLE IF NOT EXISTS pet_growth_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID（主键）',
    pet_id BIGINT UNSIGNED NOT NULL COMMENT '宠物ID（外键）',
    record_date DATE NOT NULL COMMENT '记录日期',
    weight DECIMAL(5,2) COMMENT '体重（kg）',
    height DECIMAL(5,2) COMMENT '身高（cm）',
    title VARCHAR(100) COMMENT '记录标题',
    content TEXT COMMENT '记录内容',
    images TEXT COMMENT '图片URL（JSON数组）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_pet_growth_records_pet_id (pet_id),
    INDEX idx_pet_growth_records_record_date (record_date),
    FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='宠物成长记录表';

-- ========================================
-- 话题分类表（topic_categories）
-- ========================================
CREATE TABLE IF NOT EXISTS topic_categories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '分类ID（主键）',
    name VARCHAR(50) NOT NULL COMMENT '分类名称',
    description VARCHAR(200) COMMENT '分类描述',
    icon VARCHAR(100) COMMENT '分类图标',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    status TINYINT DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_topic_categories_sort (sort_order),
    UNIQUE KEY uk_topic_categories_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题分类表';

-- ========================================
-- 话题表（topics）
-- ========================================
CREATE TABLE IF NOT EXISTS topics (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '话题ID（主键）',
    category_id BIGINT UNSIGNED COMMENT '分类ID（外键）',
    name VARCHAR(50) NOT NULL COMMENT '话题名称',
    description VARCHAR(200) COMMENT '话题描述',
    cover VARCHAR(255) COMMENT '话题封面',
    posts_count INT DEFAULT 0 COMMENT '讨论数',
    followers_count INT DEFAULT 0 COMMENT '关注数',
    sort_order INT DEFAULT 0 COMMENT '排序顺序',
    status TINYINT DEFAULT 1 COMMENT '状态（0-待审核，1-正常，2-关闭）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_topics_category_id (category_id),
    INDEX idx_topics_name (name),
    INDEX idx_topics_status (status),
    INDEX idx_topics_posts_count (posts_count DESC),
    UNIQUE KEY uk_topics_name (name),
    FOREIGN KEY (category_id) REFERENCES topic_categories(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题表';

-- ========================================
-- 话题参与者表（topic_participants）
-- ========================================
CREATE TABLE IF NOT EXISTS topic_participants (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    topic_id BIGINT UNSIGNED NOT NULL COMMENT '话题ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_topic_participants_topic_id (topic_id),
    INDEX idx_topic_participants_user_id (user_id),
    UNIQUE KEY uk_topic_participants (topic_id, user_id),
    FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='话题参与者表';

-- ========================================
-- 动态表（posts）
-- ========================================
CREATE TABLE IF NOT EXISTS posts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '动态ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    content TEXT COMMENT '内容',
    images TEXT COMMENT '图片URL（JSON数组）',
    video VARCHAR(255) COMMENT '视频URL',
    pet_ids VARCHAR(200) COMMENT '关联宠物ID（JSON数组）',
    topic_ids VARCHAR(200) COMMENT '关联话题ID（JSON数组）',
    location VARCHAR(100) COMMENT '定位',
    visibility TINYINT DEFAULT 0 COMMENT '可见范围（0-公开，1-仅粉丝，2-私密）',
    likes_count INT DEFAULT 0 COMMENT '点赞数',
    comments_count INT DEFAULT 0 COMMENT '评论数',
    shares_count INT DEFAULT 0 COMMENT '分享数',
    favorites_count INT DEFAULT 0 COMMENT '收藏数',
    is_pinned TINYINT DEFAULT 0 COMMENT '是否置顶（0-否，1-是）',
    is_featured TINYINT DEFAULT 0 COMMENT '是否精选（0-否，1-是）',
    publish_at DATETIME COMMENT '定时发布时间',
    status TINYINT DEFAULT 0 COMMENT '状态（0-待审核，1-正常，2-拒绝）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_posts_user_id (user_id),
    INDEX idx_posts_status (status),
    INDEX idx_posts_created_at (created_at DESC),
    INDEX idx_posts_is_pinned (is_pinned),
    INDEX idx_posts_is_featured (is_featured),
    INDEX idx_posts_visibility (visibility),
    INDEX idx_posts_publish_at (publish_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='动态表';

-- ========================================
-- 草稿表（drafts）
-- ========================================
CREATE TABLE IF NOT EXISTS drafts (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '草稿ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    content TEXT COMMENT '内容',
    images TEXT COMMENT '图片URL（JSON数组）',
    video VARCHAR(255) COMMENT '视频URL',
    pet_ids VARCHAR(200) COMMENT '关联宠物ID（JSON数组）',
    topic_ids VARCHAR(200) COMMENT '关联话题ID（JSON数组）',
    location VARCHAR(100) COMMENT '定位',
    visibility TINYINT DEFAULT 0 COMMENT '可见范围（0-公开，1-仅粉丝，2-私密）',
    publish_at DATETIME COMMENT '定时发布时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_drafts_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='草稿表';

-- ========================================
-- 评论表（comments）
-- ========================================
CREATE TABLE IF NOT EXISTS comments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '评论ID（主键）',
    post_id BIGINT UNSIGNED NOT NULL COMMENT '动态ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    parent_id BIGINT UNSIGNED DEFAULT 0 COMMENT '父评论ID（0=顶级评论）',
    reply_to_id BIGINT UNSIGNED DEFAULT 0 COMMENT '回复的用户ID（0=无）',
    content VARCHAR(500) NOT NULL COMMENT '评论内容',
    at_users VARCHAR(200) COMMENT '@的用户列表（JSON数组）',
    likes_count INT DEFAULT 0 COMMENT '点赞数',
    reply_count INT DEFAULT 0 COMMENT '回复数',
    status TINYINT DEFAULT 0 COMMENT '状态（0-待审核，1-正常，2-拒绝）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_comments_post_id (post_id),
    INDEX idx_comments_user_id (user_id),
    INDEX idx_comments_parent_id (parent_id),
    INDEX idx_comments_status (status),
    INDEX idx_comments_created_at (created_at DESC),
    -- parent_id / reply_to_id 用 0 作为「无」的哨兵值（ORM 默认值、查询条件、前端都依赖 0），
    -- 因此这两列不能加外键：0 不存在于 comments(id) / users(id)，任何插入都会触发 1452。
    -- 子评论的级联删除由 CommentService.DeleteComment 在事务内显式处理。
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- ========================================
-- 关注表（follows）
-- ========================================
-- 列名与 model.Follow 一致：following_id（被关注者）、status（1-正常 0-已取消）
CREATE TABLE IF NOT EXISTS follows (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    follower_id BIGINT UNSIGNED NOT NULL COMMENT '关注者ID',
    following_id BIGINT UNSIGNED NOT NULL COMMENT '被关注者ID',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '状态（1-正常，0-已取消）',
    is_secret TINYINT DEFAULT 0 COMMENT '是否悄悄关注（0-否，1-是）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_follows_follower_id (follower_id),
    INDEX idx_follows_following_id (following_id),
    UNIQUE KEY uk_follows (follower_id, following_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (following_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='关注表';

-- ========================================
-- 收藏表（favorites）
-- ========================================
CREATE TABLE IF NOT EXISTS favorites (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '收藏ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    post_id BIGINT UNSIGNED NOT NULL COMMENT '动态ID（外键）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '收藏时间',
    INDEX idx_favorites_user_id (user_id),
    INDEX idx_favorites_post_id (post_id),
    INDEX idx_favorites_created_at (created_at DESC),
    UNIQUE KEY uk_favorites (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收藏表';

-- ========================================
-- 点赞表（likes）
-- ========================================
CREATE TABLE IF NOT EXISTS likes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '点赞ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    target_type VARCHAR(20) NOT NULL COMMENT '目标类型（post/comment）',
    target_id BIGINT UNSIGNED NOT NULL COMMENT '目标ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
    INDEX idx_likes_user_id (user_id),
    INDEX idx_likes_target (target_type, target_id),
    UNIQUE KEY uk_likes (user_id, target_type, target_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞表';

-- ========================================
-- 分享记录表（share_records）
-- ========================================
CREATE TABLE IF NOT EXISTS share_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    post_id BIGINT UNSIGNED NOT NULL COMMENT '动态ID（外键）',
    platform VARCHAR(20) COMMENT '分享平台',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '分享时间',
    INDEX idx_share_records_user_id (user_id),
    INDEX idx_share_records_post_id (post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享记录表';

-- ========================================
-- 举报表（reports）
-- ========================================
CREATE TABLE IF NOT EXISTS reports (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '举报ID（主键）',
    reporter_id BIGINT UNSIGNED NOT NULL COMMENT '举报人ID（外键）',
    target_type VARCHAR(20) NOT NULL COMMENT '被举报类型（user/post/comment）',
    target_id BIGINT UNSIGNED NOT NULL COMMENT '被举报内容ID',
    reason VARCHAR(500) NOT NULL COMMENT '举报原因',
    reason_type VARCHAR(50) COMMENT '举报原因类型',
    status TINYINT DEFAULT 0 COMMENT '状态（0-待处理，1-已处理，2-已忽略）',
    handle_result VARCHAR(200) COMMENT '处理结果',
    handled_by BIGINT UNSIGNED COMMENT '处理人ID（后台用户）',
    handled_at DATETIME COMMENT '处理时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '举报时间',
    INDEX idx_reports_reporter_id (reporter_id),
    INDEX idx_reports_target (target_type, target_id),
    INDEX idx_reports_status (status),
    FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='举报表';

-- ========================================
-- 用户通知表（user_notifications）
-- 与 model.UserNotification 对齐（旧表名 notifications 已废弃）
-- ========================================
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

-- ========================================
-- 通知设置表（notification_settings）
-- ========================================
CREATE TABLE IF NOT EXISTS notification_settings (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    like_notify TINYINT DEFAULT 1 COMMENT '点赞通知（0-关闭，1-开启）',
    comment_notify TINYINT DEFAULT 1 COMMENT '评论通知（0-关闭，1-开启）',
    at_notify TINYINT DEFAULT 1 COMMENT '@通知（0-关闭，1-开启）',
    follow_notify TINYINT DEFAULT 1 COMMENT '关注通知（0-关闭，1-开启）',
    system_notify TINYINT DEFAULT 1 COMMENT '系统通知（0-关闭，1-开启）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_notification_settings_user_id (user_id),
    UNIQUE KEY uk_notification_settings_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='通知设置表';

-- ========================================
-- 会话表（conversations）
-- ========================================
CREATE TABLE IF NOT EXISTS conversations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '会话ID（主键）',
    user1_id BIGINT UNSIGNED NOT NULL COMMENT '用户1ID（外键）',
    user2_id BIGINT UNSIGNED NOT NULL COMMENT '用户2ID（外键）',
    last_message_id BIGINT UNSIGNED COMMENT '最后消息ID',
    last_message VARCHAR(200) COMMENT '最后消息内容',
    last_message_time DATETIME COMMENT '最后消息时间',
    unread_count INT DEFAULT 0 COMMENT '未读消息数',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_conversations_user1_id (user1_id),
    INDEX idx_conversations_user2_id (user2_id),
    UNIQUE KEY uk_conversations (user1_id, user2_id),
    FOREIGN KEY (user1_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (user2_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='会话表';

-- ========================================
-- 私信消息表（messages）
-- ========================================
CREATE TABLE IF NOT EXISTS messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '消息ID（主键）',
    conversation_id BIGINT UNSIGNED NOT NULL COMMENT '会话ID（外键）',
    from_id BIGINT UNSIGNED NOT NULL COMMENT '发送者ID（外键）',
    to_id BIGINT UNSIGNED NOT NULL COMMENT '接收者ID（外键）',
    content TEXT COMMENT '消息内容',
    type VARCHAR(20) DEFAULT 'text' COMMENT '消息类型（text/image/voice）',
    media_url VARCHAR(255) COMMENT '媒体URL（图片/语音）',
    duration INT NOT NULL DEFAULT 0 COMMENT '语音时长（秒，非语音为0）',
    status TINYINT DEFAULT 0 COMMENT '状态（0-未读，1-已读）',
    is_recalled TINYINT DEFAULT 0 COMMENT '是否已撤回（0-否，1-是）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    INDEX idx_messages_conversation_id (conversation_id),
    INDEX idx_messages_from_id (from_id),
    INDEX idx_messages_to_id (to_id),
    INDEX idx_messages_created_at (created_at DESC),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
    FOREIGN KEY (from_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='私信消息表';

-- ========================================
-- 拉黑表（blocks）
-- ========================================
CREATE TABLE IF NOT EXISTS blocks (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    blocker_id BIGINT UNSIGNED NOT NULL COMMENT '拉黑者ID（外键）',
    blocked_id BIGINT UNSIGNED NOT NULL COMMENT '被拉黑者ID（外键）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '拉黑时间',
    INDEX idx_blocks_blocker_id (blocker_id),
    INDEX idx_blocks_blocked_id (blocked_id),
    UNIQUE KEY uk_blocks (blocker_id, blocked_id),
    FOREIGN KEY (blocker_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (blocked_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='拉黑表';

-- ========================================
-- 用户等级表（user_levels）
-- ========================================
CREATE TABLE IF NOT EXISTS user_levels (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    level TINYINT DEFAULT 1 COMMENT '等级（1-10级）',
    points INT DEFAULT 0 COMMENT '当前积分',
    total_points INT DEFAULT 0 COMMENT '累计积分',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_user_levels_user_id (user_id),
    UNIQUE KEY uk_user_levels_user_id (user_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户等级表';

-- ========================================
-- 积分记录表（points_records）
-- ========================================
CREATE TABLE IF NOT EXISTS points_records (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    points INT NOT NULL COMMENT '积分数（正数为增加，负数为扣除）',
    type VARCHAR(20) NOT NULL COMMENT '类型（post_create/comment/like/follow/sign_in等）',
    target_id BIGINT UNSIGNED COMMENT '目标ID',
    description VARCHAR(100) COMMENT '描述',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_points_records_user_id (user_id),
    INDEX idx_points_records_type (type),
    INDEX idx_points_records_created_at (created_at DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分记录表';

-- ========================================
-- 积分规则表（points_rules）
-- ========================================
CREATE TABLE IF NOT EXISTS points_rules (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '规则ID（主键）',
    type VARCHAR(20) NOT NULL COMMENT '规则类型',
    points INT NOT NULL COMMENT '积分值',
    description VARCHAR(100) COMMENT '规则描述',
    is_active TINYINT DEFAULT 1 COMMENT '是否启用（0-禁用，1-启用）',
    daily_limit INT DEFAULT 0 COMMENT '每日上限（0表示无限制）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_points_rules_type (type),
    INDEX idx_points_rules_is_active (is_active),
    UNIQUE KEY uk_points_rules_type (type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='积分规则表';

-- ========================================
-- 等级权益表（level_benefits）
-- ========================================
CREATE TABLE IF NOT EXISTS level_benefits (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '权益ID（主键）',
    level TINYINT NOT NULL COMMENT '等级',
    benefit_type VARCHAR(20) NOT NULL COMMENT '权益类型（post_limit/comment_limit/image_limit等）',
    benefit_value VARCHAR(100) COMMENT '权益值',
    description VARCHAR(200) COMMENT '权益描述',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_level_benefits_level (level),
    INDEX idx_level_benefits_type (benefit_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='等级权益表';

-- ========================================
-- 寻宠启事表（lost_pets）
-- ========================================
CREATE TABLE IF NOT EXISTS lost_pets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '寻宠ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '发布者ID（外键）',
    pet_id BIGINT UNSIGNED COMMENT '关联宠物ID',
    pet_name VARCHAR(50) NOT NULL COMMENT '宠物名称',
    pet_type TINYINT NOT NULL COMMENT '宠物类型（1-狗，2-猫）',
    breed_id BIGINT UNSIGNED COMMENT '品种ID',
    breed_name VARCHAR(50) COMMENT '品种名称',
    gender TINYINT COMMENT '性别',
    age VARCHAR(20) COMMENT '年龄',
    description TEXT COMMENT '特征描述',
    images TEXT COMMENT '照片URL（JSON数组）',
    lost_location VARCHAR(200) NOT NULL COMMENT '走失地点',
    lost_time DATETIME NOT NULL COMMENT '走失时间',
    contact VARCHAR(100) NOT NULL COMMENT '联系方式',
    status TINYINT DEFAULT 1 COMMENT '状态（1-寻找中，2-已找到，3-已结束）',
    view_count INT DEFAULT 0 COMMENT '浏览数',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_lost_pets_user_id (user_id),
    INDEX idx_lost_pets_status (status),
    INDEX idx_lost_pets_lost_time (lost_time DESC),
    INDEX idx_lost_pets_view_count (view_count DESC),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='寻宠启事表';

-- ========================================
-- 活动表（activities）
-- ========================================
CREATE TABLE IF NOT EXISTS activities (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '活动ID（主键）',
    title VARCHAR(100) NOT NULL COMMENT '活动标题',
    description VARCHAR(500) COMMENT '活动描述',
    type VARCHAR(20) NOT NULL COMMENT '活动类型（online_vote/online_collect/offline_meet/adopt）',
    cover VARCHAR(255) COMMENT '活动封面',
    status TINYINT DEFAULT 0 COMMENT '状态（0-草稿，1-进行中，2-已结束，3-已取消）',
    start_time DATETIME NOT NULL COMMENT '开始时间',
    end_time DATETIME NOT NULL COMMENT '结束时间',
    rules TEXT COMMENT '活动规则',
    prizes TEXT COMMENT '奖品信息（JSON格式）',
    max_participants INT DEFAULT 0 COMMENT '最大参与人数',
    max_submissions INT DEFAULT 1 COMMENT '每人作品提交限制',
    votes_per_day INT DEFAULT 10 COMMENT '每日投票限制',
    require_approval TINYINT DEFAULT 0 COMMENT '是否需要审核',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_activities_type (type),
    INDEX idx_activities_status (status),
    INDEX idx_activities_start_time (start_time),
    INDEX idx_activities_end_time (end_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动表';

-- ========================================
-- 活动报名表（activity_registrations）
-- ========================================
CREATE TABLE IF NOT EXISTS activity_registrations (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '记录ID（主键）',
    activity_id BIGINT UNSIGNED NOT NULL COMMENT '活动ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    status TINYINT DEFAULT 0 COMMENT '状态（0-审核中，1-已通过，2-已拒绝，3-已取消）',
    name VARCHAR(50) NOT NULL COMMENT '报名人姓名',
    phone VARCHAR(20) NOT NULL COMMENT '联系电话',
    message VARCHAR(200) COMMENT '备注信息',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '报名时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_activity_registrations_activity_id (activity_id),
    INDEX idx_activity_registrations_user_id (user_id),
    INDEX idx_activity_registrations_status (status),
    UNIQUE KEY uk_activity_registrations (activity_id, user_id),
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动报名表';

-- ========================================
-- 活动作品表（activity_submissions）
-- ========================================
CREATE TABLE IF NOT EXISTS activity_submissions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '投稿ID（主键）',
    activity_id BIGINT UNSIGNED NOT NULL COMMENT '活动ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    title VARCHAR(50) NOT NULL COMMENT '作品标题',
    description VARCHAR(500) COMMENT '作品描述',
    type VARCHAR(20) NOT NULL COMMENT '内容类型（image/video/text）',
    content TEXT COMMENT '作品内容',
    images TEXT COMMENT '图片URL（JSON数组）',
    video VARCHAR(255) COMMENT '视频URL',
    votes_count INT DEFAULT 0 COMMENT '投票数',
    `rank` INT DEFAULT 0 COMMENT '排名',
    is_awarded TINYINT DEFAULT 0 COMMENT '是否获奖',
    award_level VARCHAR(20) COMMENT '获奖等级（first/second/third）',
    status TINYINT DEFAULT 1 COMMENT '状态（0-待审核，1-正常，2-已拒绝）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '投稿时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_activity_submissions_activity_id (activity_id),
    INDEX idx_activity_submissions_user_id (user_id),
    INDEX idx_activity_submissions_status (status),
    INDEX idx_activity_submissions_votes_count (votes_count DESC),
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动作品表';

-- ========================================
-- 活动投票表（activity_votes）
-- ========================================
CREATE TABLE IF NOT EXISTS activity_votes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '投票ID（主键）',
    activity_id BIGINT UNSIGNED NOT NULL COMMENT '活动ID（外键）',
    submission_id BIGINT UNSIGNED NOT NULL COMMENT '投稿ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '投票时间',
    INDEX idx_activity_votes_activity_id (activity_id),
    INDEX idx_activity_votes_submission_id (submission_id),
    INDEX idx_activity_votes_user_id (user_id),
    INDEX idx_activity_votes_created_at (created_at),
    UNIQUE KEY uk_activity_votes_user_submission (user_id, submission_id),
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (submission_id) REFERENCES activity_submissions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动投票表';

-- ========================================
-- 线下活动签到表（offline_activity_sign_ins）
-- ========================================
CREATE TABLE IF NOT EXISTS offline_activity_sign_ins (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '签到ID（主键）',
    activity_id BIGINT UNSIGNED NOT NULL COMMENT '活动ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    sign_in_time DATETIME NOT NULL COMMENT '签到时间',
    sign_in_location VARCHAR(100) COMMENT '签到地点',
    qr_code VARCHAR(100) COMMENT '签到二维码',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_offline_activity_sign_ins_activity_id (activity_id),
    INDEX idx_offline_activity_sign_ins_user_id (user_id),
    UNIQUE KEY uk_offline_activity_sign_ins (activity_id, user_id),
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='线下活动签到表';

-- ========================================
-- 活动相册表（activity_albums）
-- ========================================
CREATE TABLE IF NOT EXISTS activity_albums (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    activity_id BIGINT UNSIGNED NOT NULL COMMENT '活动ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    photo_url VARCHAR(255) NOT NULL COMMENT '照片URL',
    description VARCHAR(200) COMMENT '照片描述',
    status TINYINT DEFAULT 1 COMMENT '状态（0-待审核，1-已通过，2-已拒绝）',
    reviewed_by BIGINT UNSIGNED COMMENT '审核人ID',
    reviewed_at DATETIME COMMENT '审核时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_activity_albums_activity_id (activity_id),
    INDEX idx_activity_albums_user_id (user_id),
    INDEX idx_activity_albums_status (status),
    INDEX idx_activity_albums_created_at (created_at),
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动相册表';

-- ========================================
-- 活动回顾表（activity_reviews）
-- ========================================
CREATE TABLE IF NOT EXISTS activity_reviews (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    activity_id BIGINT UNSIGNED NOT NULL UNIQUE COMMENT '活动ID（外键）',
    title VARCHAR(100) NOT NULL COMMENT '回顾标题',
    content TEXT NOT NULL COMMENT '回顾内容',
    photos TEXT COMMENT '照片URL（JSON数组）',
    published_at DATETIME COMMENT '发布时间',
    published_by BIGINT UNSIGNED NOT NULL COMMENT '发布人ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_activity_reviews_activity_id (activity_id),
    INDEX idx_activity_reviews_published_at (published_at),
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='活动回顾表';

-- ========================================
-- 领养信息表（adoptions）
-- ========================================
CREATE TABLE IF NOT EXISTS adoptions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '领养ID（主键）',
    pet_id BIGINT UNSIGNED COMMENT '关联宠物ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '发布者ID（外键）',
    pet_name VARCHAR(50) NOT NULL COMMENT '宠物名称',
    pet_type VARCHAR(20) NOT NULL COMMENT '宠物类型',
    pet_age VARCHAR(20) COMMENT '宠物年龄',
    pet_gender TINYINT COMMENT '性别（0-未知，1-公，2-母）',
    health_status VARCHAR(100) COMMENT '健康状况',
    photos VARCHAR(500) COMMENT '照片URL（JSON数组）',
    description TEXT COMMENT '描述',
    adopt_conditions TEXT COMMENT '领养条件',
    location VARCHAR(100) COMMENT '所在地',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    status TINYINT DEFAULT 0 COMMENT '状态（0-待领养，1-已领养，2-已取消）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_adoptions_user_id (user_id),
    INDEX idx_adoptions_status (status),
    INDEX idx_adoptions_location (location),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (pet_id) REFERENCES pets(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领养信息表';

-- ========================================
-- 领养申请表（adoption_applications）
-- ========================================
CREATE TABLE IF NOT EXISTS adoption_applications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '申请ID（主键）',
    adoption_id BIGINT UNSIGNED NOT NULL COMMENT '领养信息ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '申请人ID（外键）',
    application_info TEXT COMMENT '申请信息（JSON格式）',
    status TINYINT DEFAULT 0 COMMENT '状态（0-审核中，1-已通过，2-已拒绝）',
    reviewed_by BIGINT UNSIGNED COMMENT '审核人ID',
    reviewed_at DATETIME COMMENT '审核时间',
    agreement_signed TINYINT DEFAULT 0 COMMENT '是否签署协议（0-未签署，1-已签署）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_adoption_applications_adoption_id (adoption_id),
    INDEX idx_adoption_applications_user_id (user_id),
    INDEX idx_adoption_applications_status (status),
    FOREIGN KEY (adoption_id) REFERENCES adoptions(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领养申请表';

-- ========================================
-- 领养回访记录表（adoption_follow_ups）
-- ========================================
CREATE TABLE IF NOT EXISTS adoption_follow_ups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '回访ID（主键）',
    adoption_id BIGINT UNSIGNED NOT NULL COMMENT '领养信息ID（外键）',
    application_id BIGINT UNSIGNED NOT NULL COMMENT '领养申请ID（外键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '回访人ID（外键）',
    follow_up_type VARCHAR(20) NOT NULL COMMENT '回访类型（visit/photo/online/phone）',
    content TEXT COMMENT '回访内容',
    photos TEXT COMMENT '回访照片（JSON数组）',
    status TINYINT DEFAULT 0 COMMENT '状态（0-正常，1-异常）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '回访时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_adoption_follow_ups_adoption_id (adoption_id),
    INDEX idx_adoption_follow_ups_application_id (application_id),
    INDEX idx_adoption_follow_ups_user_id (user_id),
    FOREIGN KEY (adoption_id) REFERENCES adoptions(id) ON DELETE CASCADE,
    FOREIGN KEY (application_id) REFERENCES adoption_applications(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='领养回访记录表';

-- ========================================
-- 广告位表（ad_positions）
-- ========================================
CREATE TABLE IF NOT EXISTS ad_positions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '广告位ID（主键）',
    name VARCHAR(50) NOT NULL COMMENT '广告位名称',
    location VARCHAR(50) NOT NULL COMMENT '广告位位置（home_banner/splash/detail_sidebar/topic_banner）',
    size VARCHAR(20) NOT NULL COMMENT '广告位尺寸',
    status TINYINT DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_ad_positions_location (location),
    INDEX idx_ad_positions_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='广告位表';

-- ========================================
-- 广告投放表（advertisements）
-- ========================================
CREATE TABLE IF NOT EXISTS advertisements (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '广告ID（主键）',
    position_id BIGINT UNSIGNED NOT NULL COMMENT '广告位ID（外键）',
    title VARCHAR(50) NOT NULL COMMENT '广告标题',
    image_url VARCHAR(255) NOT NULL COMMENT '广告图片URL',
    link_url VARCHAR(255) COMMENT '跳转链接',
    start_time DATETIME NOT NULL COMMENT '开始时间',
    end_time DATETIME NOT NULL COMMENT '结束时间',
    view_count INT DEFAULT 0 COMMENT '浏览次数',
    click_count INT DEFAULT 0 COMMENT '点击次数',
    status TINYINT DEFAULT 0 COMMENT '状态（0-待审核，1-上架，2-下架）',
    created_by BIGINT UNSIGNED NOT NULL COMMENT '创建人ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_advertisements_position_id (position_id),
    INDEX idx_advertisements_status (status),
    INDEX idx_advertisements_start_time (start_time),
    INDEX idx_advertisements_end_time (end_time),
    FOREIGN KEY (position_id) REFERENCES ad_positions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='广告投放表';

-- ========================================
-- 系统通知表（system_notifications）
-- ========================================
-- 列定义与 model.SystemNotification + 后台通知表单保持一致（类型为 1/2/3、发送对象为 0/1/2 的数值枚举）
CREATE TABLE IF NOT EXISTS system_notifications (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '通知ID（主键）',
    type TINYINT NOT NULL DEFAULT 1 COMMENT '通知类型（1-系统通知，2-公告，3-维护通知）',
    title VARCHAR(200) NOT NULL COMMENT '通知标题',
    content TEXT NOT NULL COMMENT '通知内容',
    target_type TINYINT NOT NULL DEFAULT 0 COMMENT '发送对象（0-全部用户，1-指定用户组，2-特定用户）',
    target_ids TEXT COMMENT '目标用户ID列表（英文逗号分隔）',
    status TINYINT NOT NULL DEFAULT 0 COMMENT '状态（0-草稿，1-已发布）',
    publish_at DATETIME COMMENT '定时推送时间（空=立即）',
    admin_id BIGINT UNSIGNED NOT NULL COMMENT '发布管理员ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_system_notifications_type (type),
    INDEX idx_system_notifications_status (status),
    INDEX idx_system_notifications_publish_at (publish_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统通知表';

-- ========================================
-- 公告表（announcements）
-- 与 model.Announcement 对齐
-- ========================================
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

-- ========================================
-- 用户行为日志表（user_behavior_logs）
-- ========================================
CREATE TABLE IF NOT EXISTS user_behavior_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID（主键）',
    user_id BIGINT UNSIGNED COMMENT '用户ID（外键）',
    action_type VARCHAR(50) NOT NULL COMMENT '行为类型',
    target_type VARCHAR(50) COMMENT '目标类型',
    target_id BIGINT UNSIGNED COMMENT '目标ID',
    detail TEXT COMMENT '行为详情（JSON格式）',
    ip VARCHAR(50) COMMENT 'IP地址',
    device VARCHAR(100) COMMENT '设备信息',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_user_behavior_logs_user_id (user_id),
    INDEX idx_user_behavior_logs_action_type (action_type),
    INDEX idx_user_behavior_logs_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户行为日志表';

-- ========================================
-- 用户活动日志表（user_activity_logs）
-- 与 model.UserActivityLog 对齐
-- ========================================
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

-- ========================================
-- 搜索历史表（search_histories）
-- ========================================
CREATE TABLE IF NOT EXISTS search_histories (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    keyword VARCHAR(100) NOT NULL COMMENT '搜索关键词',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '搜索时间',
    INDEX idx_search_histories_user_id (user_id),
    INDEX idx_search_histories_created_at (created_at),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='搜索历史表';

-- ========================================
-- 用户兴趣表（user_interests）
-- ========================================
CREATE TABLE IF NOT EXISTS user_interests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户ID（外键）',
    interest VARCHAR(50) NOT NULL COMMENT '兴趣标签',
    weight DECIMAL(5,2) DEFAULT 1 COMMENT '权重',
    last_active DATETIME COMMENT '最后活跃时间',
    INDEX idx_user_interests_user_id (user_id),
    INDEX idx_user_interests_interest (interest),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户兴趣表';

-- ========================================
-- 热关键词表（hot_keywords）
-- ========================================
CREATE TABLE IF NOT EXISTS hot_keywords (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT 'ID（主键）',
    keyword VARCHAR(100) NOT NULL COMMENT '关键词',
    search_count INT DEFAULT 0 COMMENT '搜索次数',
    `rank` INT DEFAULT 0 COMMENT '排名',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_hot_keywords_keyword (keyword),
    INDEX idx_hot_keywords_rank (`rank`),
    UNIQUE KEY uk_hot_keywords_keyword (keyword)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='热关键词表';

-- ========================================
-- 后台用户表（admin_users）
-- ========================================
CREATE TABLE IF NOT EXISTS admin_users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '后台用户ID（主键）',
    username VARCHAR(50) NOT NULL COMMENT '管理员用户名（唯一）',
    password VARCHAR(255) NOT NULL COMMENT '密码（BCrypt加密）',
    real_name VARCHAR(50) COMMENT '真实姓名',
    email VARCHAR(100) COMMENT '邮箱',
    phone VARCHAR(20) COMMENT '手机号',
    role VARCHAR(20) NOT NULL DEFAULT 'content_admin' COMMENT '角色（super_admin/content_admin/data_admin/service）',
    status TINYINT DEFAULT 1 COMMENT '状态（0-禁用，1-正常）',
    last_login_at DATETIME COMMENT '最后登录时间',
    last_login_ip VARCHAR(50) COMMENT '最后登录IP',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME COMMENT '删除时间（软删除）',
    INDEX idx_admin_users_username (username),
    INDEX idx_admin_users_status (status),
    INDEX idx_admin_users_deleted_at (deleted_at),
    UNIQUE KEY uk_admin_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台用户表';

-- ========================================
-- 后台操作日志表（admin_operation_logs）
-- ========================================
CREATE TABLE IF NOT EXISTS admin_operation_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '日志ID（主键）',
    admin_id BIGINT UNSIGNED NOT NULL COMMENT '管理员ID（外键）',
    action VARCHAR(100) NOT NULL COMMENT '操作类型',
    target_type VARCHAR(50) COMMENT '操作对象类型',
    target_id BIGINT UNSIGNED COMMENT '操作对象ID',
    detail TEXT COMMENT '操作详情',
    ip VARCHAR(50) COMMENT 'IP地址',
    user_agent VARCHAR(255) COMMENT '用户代理',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_admin_operation_logs_admin_id (admin_id),
    INDEX idx_admin_operation_logs_action (action),
    INDEX idx_admin_operation_logs_created_at (created_at DESC),
    FOREIGN KEY (admin_id) REFERENCES admin_users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台操作日志表';

-- ========================================
-- 后台登录日志表（admin_login_logs）
-- 与 model.AdminLoginLog 对齐
-- ========================================
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
-- 敏感词表（sensitive_words）
-- ========================================
CREATE TABLE IF NOT EXISTS sensitive_words (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '敏感词ID（主键）',
    word VARCHAR(100) NOT NULL COMMENT '敏感词',
    replacement VARCHAR(100) DEFAULT '*' COMMENT '替换词',
    category VARCHAR(50) COMMENT '分类',
    status TINYINT DEFAULT 1 COMMENT '状态（0-禁用，1-启用）',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_sensitive_words_word (word),
    INDEX idx_sensitive_words_status (status),
    UNIQUE KEY uk_sensitive_words_word (word)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='敏感词表';

-- ========================================
-- 系统配置表（system_config）
-- ========================================
CREATE TABLE IF NOT EXISTS system_config (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '配置ID（主键）',
    config_key VARCHAR(100) NOT NULL COMMENT '配置键名',
    config_value TEXT COMMENT '配置值',
    description VARCHAR(200) COMMENT '配置描述',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_system_config_config_key (config_key),
    UNIQUE KEY uk_system_config_config_key (config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ========================================
-- 每日统计表（daily_stats）
-- ========================================
CREATE TABLE IF NOT EXISTS daily_stats (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY COMMENT '统计ID（主键）',
    stat_date DATE NOT NULL COMMENT '统计日期',
    user_count INT DEFAULT 0 COMMENT '总用户数',
    new_user_count INT DEFAULT 0 COMMENT '新增用户数',
    post_count INT DEFAULT 0 COMMENT '总动态数',
    new_post_count INT DEFAULT 0 COMMENT '新增动态数',
    comment_count INT DEFAULT 0 COMMENT '总评论数',
    new_comment_count INT DEFAULT 0 COMMENT '新增评论数',
    like_count INT DEFAULT 0 COMMENT '总点赞数',
    new_like_count INT DEFAULT 0 COMMENT '新增点赞数',
    favorite_count INT DEFAULT 0 COMMENT '总收藏数',
    new_favorite_count INT DEFAULT 0 COMMENT '新增收藏数',
    share_count INT DEFAULT 0 COMMENT '总分享数',
    new_share_count INT DEFAULT 0 COMMENT '新增分享数',
    topic_count INT DEFAULT 0 COMMENT '总话题数',
    new_topic_count INT DEFAULT 0 COMMENT '新增话题数',
    active_user_count INT DEFAULT 0 COMMENT '活跃用户数',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX idx_daily_stats_stat_date (stat_date),
    UNIQUE KEY uk_daily_stats_date (stat_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='每日统计表';

-- ========================================
-- 初始化默认数据
-- ========================================

-- 密码统一为: admin123 (BCrypt cost=10；与 content_admin 等同明文校验通过)
INSERT INTO admin_users (username, password, real_name, email, role, status)
SELECT 'admin', '$2a$10$PtAxeDxVakM6hfgrX51KOudn1dMKt5bKfP1J4O/JWwAr.4WwNF2bi', '系统管理员', 'admin@petcommunity.com', 'super_admin', 1
WHERE NOT EXISTS (SELECT 1 FROM admin_users WHERE username = 'admin');

INSERT INTO admin_users (username, password, email, role, status) VALUES
('content_admin', '$2a$10$PtAxeDxVakM6hfgrX51KOudn1dMKt5bKfP1J4O/JWwAr.4WwNF2bi', 'content@petcommunity.com', 'content_admin', 1),
('data_admin', '$2a$10$Csc4LowlvrdPeWORZv2xkebtv6.lWDv16rt0pN7R4foul3HXx8AMG', 'data@petcommunity.com', 'data_admin', 1),
('service', '$2a$10$KSq2fiRVmOB.5zxSWqwiOepR/GRcj5YLFurDRjsJ7sV6/uphX8Ira', 'service@petcommunity.com', 'service', 1)
ON DUPLICATE KEY UPDATE password = VALUES(password), updated_at = NOW();

INSERT INTO topic_categories (name, sort_order, status) VALUES
('日常分享', 1, 1),
('养宠知识', 2, 1),
('美食教程', 3, 1),
('健康医疗', 4, 1),
('萌宠趣事', 5, 1)
ON DUPLICATE KEY UPDATE sort_order = VALUES(sort_order), status = VALUES(status);

INSERT INTO topics (category_id, name, description, status) VALUES
(1, '萌宠日常', '分享宠物的日常生活点滴', 1),
(2, '养宠知识', '交流养宠经验和知识', 1),
(3, '宠物美食', '分享宠物食谱和零食制作', 1),
(4, '宠物健康', '关注宠物健康和医疗', 1),
(5, '宠物趣事', '分享宠物的搞笑瞬间', 1)
ON DUPLICATE KEY UPDATE updated_at = CURRENT_TIMESTAMP;

INSERT INTO system_config (config_key, config_value, description) VALUES
('site_name', '宠物社区', '网站名称'),
('site_logo', '', '网站Logo'),
('site_description', '分享宠物生活，结识铲屎官', '网站描述'),
('site_icp', '', 'ICP备案信息'),
('audit_auto_enabled', 'true', '是否开启自动审核'),
('audit_sensitive_enabled', 'true', '是否开启敏感词过滤'),
('register_enabled', 'true', '是否开放用户注册'),
('upload_image_max_size', '2097152', '图片上传最大大小（字节）'),
('upload_video_max_size', '52428800', '视频上传最大大小（字节）'),
('max_pets_per_user', '20', '每个用户最多宠物数'),
('login_fail_max_count', '5', '登录失败最大次数'),
('login_lock_duration', '1800', '登录锁定时长（秒）'),
('code_expire_time', '300', '验证码有效期（秒）'),
('sms_daily_limit', '10', '每日短信发送上限'),
('activity_types', '["online_vote","online_collect","offline_meet","adopt"]', '活动类型'),
('vote_limit_daily', '10', '每日投票限制'),
('vote_limit_hourly', '3', '每小时投票限制'),
('submission_limit', '5', '每人作品提交限制')
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), updated_at = NOW();

INSERT INTO sensitive_words (word, replacement, category) VALUES
('垃圾', '*', '辱骂'),
('傻逼', '**', '辱骂'),
('操', '*', '辱骂'),
('fuck', '***', '辱骂'),
('色情', '**', '色情'),
('赌博', '**', '违法'),
('白痴', '**', '辱骂'),
('笨蛋', '**', '辱骂'),
('蠢货', '**', '辱骂'),
('混蛋', '**', '辱骂'),
('王八蛋', '***', '辱骂'),
('狗娘养的', '****', '辱骂'),
('神经病', '***', '辱骂'),
('变态', '**', '辱骂'),
('垃圾人', '***', '辱骂'),
('废物', '**', '辱骂'),
('黄色', '**', '色情'),
('裸聊', '**', '色情'),
('卖淫', '**', '色情'),
('嫖娼', '**', '色情'),
('包养', '**', '色情'),
('一夜情', '***', '色情'),
('约炮', '**', '色情'),
('毒品', '**', '违法'),
('贩毒', '**', '违法'),
('吸毒', '**', '违法'),
('走私', '**', '违法'),
('诈骗', '**', '违法'),
('传销', '**', '违法'),
('高利贷', '***', '违法'),
('假币', '**', '违法'),
('反动', '**', '政治敏感'),
('分裂', '**', '政治敏感'),
('独立', '**', '政治敏感'),
('代开发票', '****', '广告'),
('代刷信誉', '****', '广告'),
('兼职刷单', '****', '广告'),
('贷款', '**', '广告'),
('信用卡套现', '****', '广告'),
('加微信', '***', '广告'),
('加QQ', '***', '广告'),
('私聊', '**', '广告'),
('自杀', '**', '其他'),
('自残', '**', '其他'),
('血腥', '**', '其他'),
('暴力', '**', '其他'),
('恐怖', '**', '其他')
ON DUPLICATE KEY UPDATE status = 1;

INSERT INTO points_rules (type, points, description, is_active, daily_limit) VALUES
('post_create', 10, '发布动态', 1, 10),
('comment', 2, '发表评论', 1, 50),
('post_liked', 1, '动态被点赞', 1, 0),
('comment_liked', 1, '评论被点赞', 1, 0),
('follow', 5, '关注用户', 1, 10),
('sign_in', 5, '每日签到', 1, 1),
('share', 3, '分享动态', 1, 5),
('favorite', 1, '收藏动态', 1, 20),
('at_reply', 2, '@回复', 1, 30),
('activity_join', 20, '参加活动', 1, 5)
ON DUPLICATE KEY UPDATE points = VALUES(points), description = VALUES(description), is_active = VALUES(is_active), daily_limit = VALUES(daily_limit), updated_at = CURRENT_TIMESTAMP;

INSERT INTO level_benefits (level, benefit_type, benefit_value, description)
SELECT * FROM (
    SELECT 1 AS level, 'level_name' AS benefit_type, '铲屎官' AS benefit_value, '等级名称' AS description
    UNION ALL SELECT 1, 'points_required', '0', '升级所需积分'
    UNION ALL SELECT 1, 'post_limit', '3', '每日发布动态上限'
    UNION ALL SELECT 1, 'image_limit', '3', '单条动态图片上限'
    UNION ALL SELECT 2, 'level_name', '初级铲屎官', '等级名称'
    UNION ALL SELECT 2, 'points_required', '100', '升级所需积分'
    UNION ALL SELECT 2, 'post_limit', '5', '每日发布动态上限'
    UNION ALL SELECT 2, 'image_limit', '5', '单条动态图片上限'
    UNION ALL SELECT 3, 'level_name', '中级铲屎官', '等级名称'
    UNION ALL SELECT 3, 'points_required', '500', '升级所需积分'
    UNION ALL SELECT 3, 'post_limit', '8', '每日发布动态上限'
    UNION ALL SELECT 3, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 4, 'level_name', '高级铲屎官', '等级名称'
    UNION ALL SELECT 4, 'points_required', '1000', '升级所需积分'
    UNION ALL SELECT 4, 'post_limit', '10', '每日发布动态上限'
    UNION ALL SELECT 4, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 5, 'level_name', '资深铲屎官', '等级名称'
    UNION ALL SELECT 5, 'points_required', '2000', '升级所需积分'
    UNION ALL SELECT 5, 'post_limit', '15', '每日发布动态上限'
    UNION ALL SELECT 5, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 5, 'feature_priority', '1', '精选优先'
    UNION ALL SELECT 6, 'level_name', '宠物达人', '等级名称'
    UNION ALL SELECT 6, 'points_required', '5000', '升级所需积分'
    UNION ALL SELECT 6, 'post_limit', '20', '每日发布动态上限'
    UNION ALL SELECT 6, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 6, 'feature_priority', '2', '精选优先'
    UNION ALL SELECT 7, 'level_name', '宠物专家', '等级名称'
    UNION ALL SELECT 7, 'points_required', '10000', '升级所需积分'
    UNION ALL SELECT 7, 'post_limit', '30', '每日发布动态上限'
    UNION ALL SELECT 7, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 7, 'feature_priority', '3', '精选优先'
    UNION ALL SELECT 8, 'level_name', '宠物大师', '等级名称'
    UNION ALL SELECT 8, 'points_required', '20000', '升级所需积分'
    UNION ALL SELECT 8, 'post_limit', '50', '每日发布动态上限'
    UNION ALL SELECT 8, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 8, 'feature_priority', '4', '精选优先'
    UNION ALL SELECT 9, 'level_name', '宠物宗师', '等级名称'
    UNION ALL SELECT 9, 'points_required', '50000', '升级所需积分'
    UNION ALL SELECT 9, 'post_limit', '100', '每日发布动态上限'
    UNION ALL SELECT 9, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 9, 'feature_priority', '5', '精选优先'
    UNION ALL SELECT 10, 'level_name', '宠物之神', '等级名称'
    UNION ALL SELECT 10, 'points_required', '100000', '升级所需积分'
    UNION ALL SELECT 10, 'post_limit', '999', '每日发布动态上限'
    UNION ALL SELECT 10, 'image_limit', '9', '单条动态图片上限'
    UNION ALL SELECT 10, 'feature_priority', '10', '精选优先'
) AS seed
WHERE NOT EXISTS (SELECT 1 FROM level_benefits LIMIT 1);

INSERT INTO ad_positions (name, location, size, status)
SELECT * FROM (
    SELECT '首页Banner' AS name, 'home_banner' AS location, '1920x400' AS size, 1 AS status
    UNION ALL SELECT '开屏广告', 'splash', '1920x1080', 1
    UNION ALL SELECT '详情页广告', 'detail_sidebar', '300x250', 1
    UNION ALL SELECT '话题页广告', 'topic_banner', '800x200', 1
) AS seed
WHERE NOT EXISTS (SELECT 1 FROM ad_positions LIMIT 1);

INSERT INTO hot_keywords (keyword, search_count, `rank`) VALUES
('猫咪', 1200, 1),
('狗狗', 980, 2),
('泰迪', 750, 3),
('布偶猫', 620, 4),
('金毛', 580, 5),
('柯基', 450, 6),
('英短', 420, 7),
('哈士奇', 380, 8),
('龙猫', 320, 9),
('仓鼠', 280, 10)
ON DUPLICATE KEY UPDATE search_count = VALUES(search_count), `rank` = VALUES(`rank`);

INSERT INTO system_notifications (type, title, content, target_type, status, admin_id)
SELECT * FROM (
    SELECT 1 AS type, '欢迎加入宠物社区' AS title, '欢迎您加入我们的宠物社区！在这里您可以分享宠物日常、参加活动、结交志同道合的朋友。' AS content, 0 AS target_type, 1 AS status, 1 AS admin_id
    UNION ALL SELECT 1, '活动报名成功通知', '您已成功报名参加活动，请按时参加！', 2, 0, 1
    UNION ALL SELECT 1, '领养申请审核结果', '您的领养申请已审核，请查看详情。', 2, 0, 1
) AS seed
WHERE NOT EXISTS (SELECT 1 FROM system_notifications LIMIT 1);

-- ========================================
-- 创建视图
-- ========================================

DROP VIEW IF EXISTS v_user_stats;
CREATE VIEW v_user_stats AS
SELECT 
    COUNT(*) as total_users,
    SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as active_users,
    SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as banned_users,
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
    SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as rejected_posts,
    SUM(likes_count) as total_likes,
    SUM(comments_count) as total_comments,
    SUM(shares_count) as total_shares,
    DATE_FORMAT(created_at, '%Y-%m-%d') as date
FROM posts
GROUP BY DATE_FORMAT(created_at, '%Y-%m-%d');

DROP VIEW IF EXISTS v_hot_topics;
CREATE VIEW v_hot_topics AS
SELECT 
    t.id,
    t.name,
    t.description,
    t.cover,
    t.posts_count,
    t.created_at
FROM topics t
WHERE t.status = 1
ORDER BY t.posts_count DESC, t.created_at DESC;

DROP VIEW IF EXISTS v_user_detail;
CREATE VIEW v_user_detail AS
SELECT 
    u.*,
    (SELECT COUNT(*) FROM follows WHERE following_id = u.id AND status = 1) as follower_count,
    (SELECT COUNT(*) FROM follows WHERE follower_id = u.id AND status = 1) as following_count,
    (SELECT COUNT(*) FROM posts WHERE user_id = u.id AND status = 1) as post_count,
    (SELECT COUNT(*) FROM pets WHERE user_id = u.id) as pet_count
FROM users u;

-- ========================================
-- 存储过程
-- ========================================

DELIMITER //
DROP PROCEDURE IF EXISTS GetUserStats //
CREATE PROCEDURE GetUserStats(IN start_date DATE, IN end_date DATE)
BEGIN
    SELECT 
        COUNT(*) as total_users,
        SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as active_users,
        SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as banned_users
    FROM users
    WHERE created_at BETWEEN start_date AND DATE_ADD(end_date, INTERVAL 1 DAY);
END //
DELIMITER ;

DELIMITER //
DROP PROCEDURE IF EXISTS GetPostStats //
CREATE PROCEDURE GetPostStats(IN start_date DATE, IN end_date DATE)
BEGIN
    SELECT 
        COUNT(*) as total_posts,
        SUM(CASE WHEN status = 1 THEN 1 ELSE 0 END) as published_posts,
        SUM(CASE WHEN status = 0 THEN 1 ELSE 0 END) as pending_posts,
        SUM(CASE WHEN status = 2 THEN 1 ELSE 0 END) as rejected_posts,
        SUM(likes_count) as total_likes,
        SUM(comments_count) as total_comments,
        SUM(shares_count) as total_shares
    FROM posts
    WHERE created_at BETWEEN start_date AND DATE_ADD(end_date, INTERVAL 1 DAY);
END //
DELIMITER ;

DELIMITER //
DROP PROCEDURE IF EXISTS GetUserProfile //
CREATE PROCEDURE GetUserProfile(IN user_id BIGINT)
BEGIN
    SELECT 
        u.*,
        COALESCE(fr.follower_count, 0) as follower_count,
        COALESCE(fg.following_count, 0) as following_count,
        COALESCE(p.post_count, 0) as post_count,
        COALESCE(pt.pet_count, 0) as pet_count
    FROM users u
    LEFT JOIN (
        SELECT following_id, COUNT(*) as follower_count
        FROM follows
        WHERE status = 1
        GROUP BY following_id
    ) fr ON u.id = fr.following_id
    LEFT JOIN (
        SELECT follower_id, COUNT(*) as following_count
        FROM follows
        WHERE status = 1
        GROUP BY follower_id
    ) fg ON u.id = fg.follower_id
    LEFT JOIN (
        SELECT user_id, COUNT(*) as post_count
        FROM posts
        WHERE status = 1
        GROUP BY user_id
    ) p ON u.id = p.user_id
    LEFT JOIN (
        SELECT user_id, COUNT(*) as pet_count
        FROM pets
        GROUP BY user_id
    ) pt ON u.id = pt.user_id
    WHERE u.id = user_id;
END //
DELIMITER ;

-- ========================================
-- 触发器
-- ========================================

DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_like_add
AFTER INSERT ON likes
FOR EACH ROW
BEGIN
    IF NEW.target_type = 'post' THEN
        UPDATE posts SET likes_count = likes_count + 1 WHERE id = NEW.target_id;
    ELSEIF NEW.target_type = 'comment' THEN
        UPDATE comments SET likes_count = likes_count + 1 WHERE id = NEW.target_id;
    END IF;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_like_remove
AFTER DELETE ON likes
FOR EACH ROW
BEGIN
    IF OLD.target_type = 'post' THEN
        UPDATE posts SET likes_count = likes_count - 1 WHERE id = OLD.target_id;
    ELSEIF OLD.target_type = 'comment' THEN
        UPDATE comments SET likes_count = likes_count - 1 WHERE id = OLD.target_id;
    END IF;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_comment_add
AFTER INSERT ON comments
FOR EACH ROW
BEGIN
    UPDATE posts SET comments_count = comments_count + 1 WHERE id = NEW.post_id;
    IF NEW.parent_id IS NOT NULL AND NEW.parent_id > 0 THEN
        UPDATE comments SET reply_count = reply_count + 1 WHERE id = NEW.parent_id;
    END IF;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_comment_remove
AFTER DELETE ON comments
FOR EACH ROW
BEGIN
    UPDATE posts SET comments_count = comments_count - 1 WHERE id = OLD.post_id;
    IF OLD.parent_id IS NOT NULL AND OLD.parent_id > 0 THEN
        UPDATE comments SET reply_count = reply_count - 1 WHERE id = OLD.parent_id;
    END IF;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_vote_add
AFTER INSERT ON activity_votes
FOR EACH ROW
BEGIN
    UPDATE activity_submissions SET votes_count = votes_count + 1 WHERE id = NEW.submission_id;
END //
DELIMITER ;

DELIMITER //
CREATE TRIGGER IF NOT EXISTS tr_vote_remove
AFTER DELETE ON activity_votes
FOR EACH ROW
BEGIN
    UPDATE activity_submissions SET votes_count = votes_count - 1 WHERE id = OLD.submission_id;
END //
DELIMITER ;

-- ========================================
-- 创建用户定义函数
-- ========================================

DELIMITER //
DROP FUNCTION IF EXISTS GetUserLevelName //
CREATE FUNCTION GetUserLevelName(level TINYINT) RETURNS VARCHAR(20)
DETERMINISTIC
BEGIN
    DECLARE level_name VARCHAR(20);
    CASE level
        WHEN 1 THEN SET level_name = '铲屎官';
        WHEN 2 THEN SET level_name = '初级铲屎官';
        WHEN 3 THEN SET level_name = '中级铲屎官';
        WHEN 4 THEN SET level_name = '高级铲屎官';
        WHEN 5 THEN SET level_name = '资深铲屎官';
        WHEN 6 THEN SET level_name = '宠物达人';
        WHEN 7 THEN SET level_name = '宠物专家';
        WHEN 8 THEN SET level_name = '宠物大师';
        WHEN 9 THEN SET level_name = '宠物宗师';
        WHEN 10 THEN SET level_name = '宠物之神';
        ELSE SET level_name = '未知';
    END CASE;
    RETURN level_name;
END //
DELIMITER ;

DELIMITER //
DROP FUNCTION IF EXISTS GetRelativeTime //
CREATE FUNCTION GetRelativeTime(datetime_val DATETIME) RETURNS VARCHAR(50)
DETERMINISTIC
BEGIN
    DECLARE diff_seconds INT;
    DECLARE result VARCHAR(50);
    
    SET diff_seconds = TIMESTAMPDIFF(SECOND, datetime_val, NOW());
    
    IF diff_seconds < 60 THEN
        SET result = CONCAT(diff_seconds, '秒前');
    ELSEIF diff_seconds < 3600 THEN
        SET result = CONCAT(FLOOR(diff_seconds / 60), '分钟前');
    ELSEIF diff_seconds < 86400 THEN
        SET result = CONCAT(FLOOR(diff_seconds / 3600), '小时前');
    ELSEIF diff_seconds < 172800 THEN
        SET result = '昨天';
    ELSEIF diff_seconds < 604800 THEN
        SET result = CONCAT(FLOOR(diff_seconds / 86400), '天前');
    ELSE
        SET result = DATE_FORMAT(datetime_val, '%Y-%m-%d');
    END IF;
    
    RETURN result;
END //
DELIMITER ;

-- ========================================
-- 脚本执行完成
-- ========================================
SELECT '数据库初始化脚本执行完成' as result;