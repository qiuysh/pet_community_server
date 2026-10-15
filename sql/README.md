# 数据库 SQL 脚本

脚本按用途分为两类，**不要混用**：

| 目录 | 用途 | 何时执行 |
|------|------|----------|
| `init/` | 初始化：建库建表 + 系统基础数据 | **全新空库**一次性执行 |
| `incremental/` | 增量：开发种子 / 补丁 / 结构变更 | 在 `init` 之后按文件名时间序执行 |

```
sql/
├── init/                                      # 初始化（新库）
│   └── create.sql                          # 建库建表 + 系统基础数据（必选）
└── incremental/                               # 增量
    └── 20260724100000_seed_dev_data.sql       # 开发/E2E 测试种子（Phase1/2/5）
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

## 增量脚本

```bash
# 按文件名时间戳顺序执行尚未应用的脚本
mysql -u root -p pet_community < server/sql/incremental/20260724100000_seed_dev_data.sql
```

### 命名

```
{yyyyMMddHHmmss}_{description}.sql
```

约定：

- 只写对已有库必要的变更或补充数据
- 新表、新字段若需长期保留，应同时合并进 `init/create.sql`，保证新库一次建全
- 已提交的增量脚本**禁止改内容**；有误则新增下一条增量

## 与迁移规范的关系

完整规范见 [docs/backend/migration-guide.md](../../docs/backend/migration-guide.md)。  
当前仓库以本目录手写 SQL 为准；自动化 migrate 命令若未落地，请用上方 `mysql <` 方式执行。
