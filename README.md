# 🐾 宠物社区 Web 应用

一个面向宠物爱好者的社区平台，提供宠物档案管理、社区互动、社交交流等功能。

## 🛠️ 技术栈

- **语言**: Go 1.22+
- **框架**: Gin Web Framework
- **ORM**: GORM v2
- **数据库**: MySQL 8.0+
- **缓存**: Redis 7.0+
- **认证**: JWT

## 📁 项目结构

```bash

pet_community/
├── server/                # 后端代码
│   ├── cmd/               # 命令入口
│   ├── controller/        # HTTP 控制器
│   ├── service/           # 业务逻辑
│   ├── repository/        # 数据访问层
│   ├── model/             # 数据库模型
│   ├── types/             # 类型定义
│   ├── middleware/        # 中间件
│   ├── util/              # 工具函数
│   ├── config/            # 配置文件
│   └── sql/               # 数据库脚本（init 初始化 / incremental 增量）

```

## 🚀 快速开始

### 环境要求

- Go >= 1.22.0
- MySQL >= 8.0.0
- Redis >= 7.0.0


### 后端启动

```bash
cd server
go mod download
go run cmd/main.go
```

### 数据库初始化

```bash
# 全新空库：结构 + 系统基础数据
mysql -u root -p < sql/init/create.sql

# 可选：开发/E2E 测试种子
mysql -u root -p pet_community < sql/incremental/20260724100000_seed_dev_data.sql
```

说明见 [sql/README.md](./sql/README.md)。

## 🔧 配置说明


### 后端配置

后端唯一配置文件为 `server/config.yaml`（不分环境，不再读取 `.env`）。

## 🔐 API 认证

API 使用 JWT Token 认证，请求时在 Header 中携带：

```
Authorization: Bearer <token>
```

## 📖 文档

- **需求文档**: `docs/requirements/pet-community-requirements.md`
- **前端规范**: `docs/frontend/`
- **后端规范**: `docs/backend/`
- **API 文档**: `docs/backend/api-spec.md`
- **数据库设计**: `docs/backend/database-design.md`
