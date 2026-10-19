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
    └── 20261019154000_schema_fix_20261019.sql  # 结构补丁合集（幂等，已同步进 create.sql）
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

新库执行完 `create.sql` 后，**不必**再跑结构类增量（`schema_fix`），其内容已同步进 init。

## 已有库升级

```bash
mysql -u root -p pet_community < server/sql/incremental/20261019154000_schema_fix_20261019.sql
```

开发种子仅开发/E2E 需要，生产不跑。

### 命名

```
{yyyyMMddHHmmss}_{description}.sql
```

约定：

- 只写对已有库必要的变更或补充数据
- 新表、新字段若需长期保留，应同时合并进 `init/create.sql`，保证新库一次建全
- 已提交的增量脚本**禁止改内容**；有误则新增下一条增量
- 当日未发布的多份结构补丁可合并进合集（如 `schema_fix`），开发种子与结构补丁勿混文件

## 与迁移规范的关系

完整规范见 [docs/backend/migration-guide.md](../../docs/backend/migration-guide.md)。  
当前仓库以本目录手写 SQL 为准；自动化 migrate 命令若未落地，请用上方 `mysql <` 方式执行。
