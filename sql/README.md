# 数据库 SQL 脚本

脚本按用途分为两类，**不要混用**：

| 目录 | 用途 | 何时执行 |
|------|------|----------|
| `init/` | 初始化：建库建表 + 系统基础数据 | **全新空库**一次性执行 |
| `incremental/` | 增量：开发种子 / 补丁 / 结构变更 | 在 `init` 之后按文件名时间序执行 |

```
sql/
├── init/                                        # 初始化（新库）
│   └── create.sql                               # 建库建表 + 系统基础数据（必选）
└── incremental/                                 # 增量（按时间戳顺序）
    ├── 20260724100000_seed_dev_data.sql         # 开发/E2E 测试种子（Phase1/2/5）
    ├── 20261019154000_schema_fix_20261019.sql  # 结构补丁合集（幂等，已同步进 create.sql）
    ├── 20261020113000_fix_20261020.sql         # 2026-10-20 合集：唯一索引对齐 + 上传前缀迁移
    └── 20261105000000_schema_fix_20261105.sql       # 2026-11-05 schema 收口合集（8 合 1，已执行，见下方速查表）
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

新库执行完 `create.sql` 后，**不必**再跑结构类增量（`schema_fix` / 合集第一节），其内容已同步进 init。  
合集第二节（上传前缀数据改写）对新库通常无旧 `/uploads/` 数据，可跳过或幂等空跑。

## 已有库升级

按文件名时间序执行尚未跑过的增量：

```bash
mysql -u root -p pet_community < server/sql/incremental/20261019154000_schema_fix_20261019.sql
# 先备份再建 2026-10-20 合集（含索引对齐 + 上传前缀迁移）
mysqldump -u root -p pet_community > backup_$(date +%Y%m%d%H%M).sql
mysql -u root -p pet_community < server/sql/incremental/20261020113000_fix_20261020.sql
```

开发种子仅开发/E2E 需要，生产不跑。

### 2026-10-20 合集说明

| 节 | 内容 | 新库是否需要 |
|----|------|--------------|
| 一 | `user_levels` / `notification_settings` / `daily_stats` 唯一索引名与 GORM 对齐（消 Error 1061） | 否（已进 `create.sql`） |
| 二 | 存量 URL `/uploads/` → `/api/v1/uploads/`（M4-04） | 否（无旧前缀数据） |

第二节只改相对路径；`http(s)://` 外链、已含新前缀的值不受影响。

### 2026-11-05 schema 收口合集（已执行，勿重复跑）

当日 8 个增量脚本已整合为单个文件 `20261105000000_schema_fix_20261105.sql`（分文件已删除，内容逐字保留，按原时间序排成八节）。

- **当前库已按分文件执行过全部内容**：合集对本库是幂等空跑，不必再执行；新库只需跑 `init/create.sql`（内容已合并进去）。
- 合集面向「只跑 init 建库」的存量环境（导库 / QA / BI 只读库）一次执行到位；含两处 `DELIMITER` 块，请整文件执行、不要按分号逐句拆跑。
- 留此表仅作变更追溯：

| 节 | 原脚本 | 内容 |
|----|--------|------|
| 一 | `113000_add_topic_shares_count` | `topics.shares_count`，支撑话题分享计数 |
| 二 | `161000_drop_counter_triggers` | 删除 6 个计数器触发器（改为事务内自增，消除双写口径） |
| 三 | `161100_recompute_counter_columns` | 重算历史计数列（评论/回复/点赞/收藏/分享） |
| 四 | `163000_fix_topics_category_id` | `topics.category_id` 补 `NOT NULL DEFAULT 0`（原可空导致扫描崩溃） |
| 五 | `163100_fix_stats_status_semantics` | 统计视图/过程的状态口径修正（上架 vs 审核态反转） |
| 六 | `170000_add_missing_columns` | 补 8 个模型有、库中缺失的列（schema 漂移） |
| 七 | `170100_fix_v_user_stats_cumulative` | `v_user_stats.cumulative_users` 改为窗口累计 |
| 八 | `172000_normalize_unique_indexes` | 唯一索引名归一（`uni_*` → `uk_*`，清理重复索引） |

### 命名

```
{yyyyMMddHHmmss}_{description}.sql
```

约定：

- 只写对已有库必要的变更或补充数据
- 新表、新字段若需长期保留，应同时合并进 `init/create.sql`，保证新库一次建全
- 已提交的增量脚本**禁止改内容**；有误则新增下一条增量
- 当日未发布的多份结构/数据补丁可合并进合集（如 `fix_YYYYMMDD`），开发种子与结构补丁勿混文件
- 同一天已执行过的多个补丁，事后整合成合集时**内容逐字保留**、只统一头部与分节（参照 2026-11-05 合集做法），保证曾执行过的库幂等空跑结果不变

## 与迁移规范的关系

完整规范见 [docs/backend/migration-guide.md](../../docs/backend/migration-guide.md)。  
当前仓库以本目录手写 SQL 为准；自动化 migrate 命令若未落地，请用上方 `mysql <` 方式执行。
