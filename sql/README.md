# 数据库 SQL 脚本

脚本按用途分为两类，**不要混用**：

| 目录 | 用途 | 何时执行 |
|------|------|----------|
| `init/` | 初始化：建库建表 + 系统基础数据 | **全新空库**一次性执行 |
| `incremental/` | 增量：开发种子 / 补丁 / 结构变更 | 在 `init` 之后按文件名时间序执行 |

```
sql/
├── init/                                              # 初始化（新库）
│   └── create.sql                                     # 建库建表 + 系统基础数据（必选）
└── incremental/                                       # 增量（按时间戳顺序）
    ├── 20260724100000_seed_dev_data.sql               # 开发/E2E 测试种子（Phase1/2/5）
    ├── 20261105000000_schema_fix_through_20261105.sql # 结构合集：2026-10-19 ~ 11-05（已执行，见下方）
    └── 20261110120000_schema_fix_20261110.sql         # 结构合集：2026-11-06 ~ 11-10（见下方）
```

## 新库初始化

```bash
# 1. 必选：结构 + 系统账号/配置/分类等
mysql -u root -p < server/sql/init/create.sql

# 2. 可选：本地开发 / E2E 测试数据
mysql -u root -p pet_community < server/sql/incremental/20260724100000_seed_dev_data.sql
```

`create.sql` 已含后台账号（密码 `admin123`）。  
开发种子前台账号密码为 `user123456`（见脚本内注释）。

新库执行完 `create.sql` 后，**不必**再跑结构类增量（两份 `schema_fix` 合集），其内容已同步进 init。  
`through_20261105` 卷二第二节（上传前缀数据改写）对新库通常无旧 `/uploads/` 数据，可跳过或幂等空跑。

## 已有库升级

按文件名时间序执行尚未跑过的增量：

```bash
# 若尚未跑过 2026-10-19 / 10-20 / 11-05 批次（或对应旧分文件），执行：
mysqldump -u root -p pet_community > backup_$(date +%Y%m%d%H%M).sql
mysql -u root -p pet_community < server/sql/incremental/20261105000000_schema_fix_through_20261105.sql

# 若尚未跑过 2026-11-06 / 11-10 批次（或对应旧分文件），执行：
mysql -u root -p pet_community < server/sql/incremental/20261110120000_schema_fix_20261110.sql
```

开发种子仅开发/E2E 需要，生产不跑。

### 合集一：through 2026-11-05（已整合，曾执行过可跳过）

由三日合集再收口为单文件（正文逐字保留）：

| 卷 | 原文件 | 内容摘要 |
|----|--------|----------|
| 一 | `schema_fix_20261019` | messages.duration；comments/pets 哨兵 FK；system_notifications / follows 对齐；死表清理；users 唯一索引 |
| 二 | `fix_20261020` | 唯一索引名与 GORM 对齐；存量 `/uploads/` → `/api/v1/uploads/` |
| 三 | `schema_fix_20261105` | 话题分享数、删计数触发器、重算计数、topics.category_id、统计口径、补列、累计用户视图、唯一索引归一（含 DELIMITER） |

- **当前库若已按旧分日合集 / 分文件跑过**：对本库幂等空跑，不必再执行。
- 含 `DELIMITER` 块，请整文件执行、不要按分号逐句拆跑。

卷二速查：

| 节 | 内容 | 新库是否需要 |
|----|------|--------------|
| 一 | `user_levels` / `notification_settings` / `daily_stats` 唯一索引名与 GORM 对齐 | 否（已进 `create.sql`） |
| 二 | 存量 URL `/uploads/` → `/api/v1/uploads/`（M4-04） | 否（无旧前缀数据） |

卷三（原 11-05 八节）追溯：

| 节 | 原脚本 | 内容 |
|----|--------|------|
| 一 | `113000_add_topic_shares_count` | `topics.shares_count` |
| 二 | `161000_drop_counter_triggers` | 删除 6 个计数器触发器 |
| 三 | `161100_recompute_counter_columns` | 重算历史计数列 |
| 四 | `163000_fix_topics_category_id` | `topics.category_id` NOT NULL DEFAULT 0 |
| 五 | `163100_fix_stats_status_semantics` | 统计视图/过程状态口径 |
| 六 | `170000_add_missing_columns` | 补模型有、库中缺失的列 |
| 七 | `170100_fix_v_user_stats_cumulative` | `cumulative_users` 窗口累计 |
| 八 | `172000_normalize_unique_indexes` | 唯一索引名归一 |

### 合集二：2026-11-10（11-06 ~ 11-10）

| 节 | 原脚本 | 内容 |
|----|--------|------|
| 一 | `add_notification_settings_message_notify` | `notification_settings.message_notify` |
| 二 | `add_lost_pets_shares_count` | `lost_pets.shares_count` |
| 三 | `add_notification_settings_security_notify` | `notification_settings.security_notify` |
| 四 | `add_pets_visibility` | `pets.visibility` + 索引（合集内已幂等化） |

全新库无需执行（已进 `create.sql`）。曾按分文件执行过的库可跳过或空跑。

### 命名

```
{yyyyMMddHHmmss}_{description}.sql
```

约定：

- 只写对已有库必要的变更或补充数据
- 新表、新字段若需长期保留，应同时合并进 `init/create.sql`，保证新库一次建全
- 已提交的增量脚本**禁止改内容**；有误则新增下一条增量
- 当日未发布的多份结构/数据补丁可合并进合集（如 `fix_YYYYMMDD`），开发种子与结构补丁勿混文件
- 已执行过的多个补丁事后整合成合集时**内容逐字保留**、只统一头部与分节；合集重复执行须幂等（非幂等 ALTER 整合时包成 information_schema + PREPARE）

## 与迁移规范的关系

完整规范见 [docs/backend/migration-guide.md](../../docs/backend/migration-guide.md)。  
当前仓库以本目录手写 SQL 为准；自动化 migrate 命令若未落地，请用上方 `mysql <` 方式执行。
