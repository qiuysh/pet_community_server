# 宠物社区 · 后端

社区 RESTful API 服务（独立仓库 / 分支）。

## 技术栈

- **语言**: Go 1.22+
- **框架**: Gin
- **ORM**: GORM v2
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+（可选，失败可降级）
- **认证**: JWT

## 目录结构

```
├── cmd/           # 入口
├── config/        # 配置加载、DB / Redis 初始化
├── controller/    # HTTP 控制器
├── service/       # 业务逻辑
├── repository/    # 数据访问
├── model/         # 表模型
├── types/         # 请求/响应 DTO
├── middleware/    # 中间件
├── utils/         # 工具
├── sql/           # 数据库脚本（init / incremental）
├── config.yaml    # 唯一应用配置
└── go.mod
```

## 环境要求

- Go >= 1.22
- MySQL >= 8.0
- Redis >= 7.0（可选）

## 数据库初始化

```bash
# 必选：建库建表 + 系统基础数据
mysql -u root -p < sql/init/create.sql

# 可选：开发 / E2E 测试种子
mysql -u root -p pet_community < sql/incremental/20260724100000_seed_dev_data.sql
```

说明见 [sql/README.md](./sql/README.md)。

- 后台默认账号：`admin` / `admin123`（及脚本内其它角色账号）
- 前台测试账号密码见种子脚本注释（`user123456`）

## 快速开始

```bash
# 按需修改 config.yaml（数据库、Redis、端口等）
go mod download
go run cmd/main.go
```

默认监听：`http://0.0.0.0:8083`（见 `config.yaml` → `server.port`）。

## 配置说明

唯一配置文件：`config.yaml`（不读 `.env`）。主要项：

| 区块 | 说明 |
|------|------|
| `server` | 监听地址、端口、超时 |
| `database` | MySQL 连接 |
| `jwt` | JWT 密钥（长度 ≥ 16） |
| `redis` | 是否启用及连接池 |
| `rate_limit` | 限流 |
| `scheduler` | 定时任务 cron |

## API 认证

在请求头携带：

```
Authorization: Bearer <token>
```

## 协作约定

- 后端规范与 API / 库表设计见关联文档中的 `docs/backend/`、`docs/specs/`
- Agent 协作规范见 `AGENTS.md`（若本分支包含）
