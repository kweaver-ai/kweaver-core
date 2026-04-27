# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

### Backend (Go)
```bash
cd server
go mod tidy          # 安装依赖
go run main.go       # 运行服务
go build -o vega-backend  # 构建
go test ./...        # 运行所有测试
go test -v ./logics/catalog/...  # 运行单个包的测试
go vet ./...         # 静态检查
```

### Frontend (React)
```bash
cd frontend
npm install          # 安装依赖
npm run dev          # 开发服务器
npm run build        # 生产构建
npm run lint         # ESLint 检查
npm run preview      # 预览构建结果
```

### Configuration
服务配置文件: `./config/vega-backend-config.yaml`

## Architecture

VEGA Manager 采用六边形架构 (Hexagonal Architecture / Ports & Adapters):

```
server/
├── interfaces/         # 端口定义 (Ports) - 服务和数据访问接口
├── logics/             # 核心业务逻辑 - 服务实现
│   ├── catalog/        # Catalog 服务
│   ├── resource/       # Resource 服务
│   └── connectors/     # 数据源连接器 (Factory 模式)
├── driveradapters/     # 主适配器 (Primary Adapters) - HTTP REST 处理器
└── drivenadapters/     # 次适配器 (Secondary Adapters) - 数据库访问、外部服务
    ├── catalog/        # Catalog 数据访问
    ├── resource/       # Resource 数据访问
    └── opensearch/     # OpenSearch 客户端
```

### 核心概念

**Catalog (数据目录)**: 数据源连接，可以是物理数据源(MySQL, S3, Kafka等)或逻辑目录(虚拟视图)

**Resource (数据资源)**: 数据资源实体，支持9种类型:
- 物理: table, file, fileset, api, metric, topic, index
- 逻辑: logicview, dataset

**Connectors**: 可插拔的数据源连接器，按类型组织在 `logics/connectors/`:
- table/ (MySQL, PostgreSQL, 达梦, Oracle, ClickHouse)
- index/ (OpenSearch, Elasticsearch)
- 计划中: fileset/, topic/, metric/, api/

### REST API

基础路径: `/api/vega-backend/v1`

| 资源 | 端点 |
|------|------|
| Catalog | GET/POST `/catalogs`, GET/PUT/DELETE `/catalogs/:id` |
| Catalog 状态 | GET `/catalogs/:id/status`, POST `/catalogs/:id/test-connection` |
| Resource | GET/POST `/resources`, GET/PUT/DELETE `/resources/:id` |
| Resource 操作 | GET `/resources/:id/schema`, POST `/resources/:id/enable\|disable\|sync` |
| 健康检查 | GET `/health` |

## Database

MariaDB/MySQL 8.0+，主要表:
- `t_catalog` - 目录/数据源配置
- `t_resource` - 数据资源定义，含 schema (JSON)
- `t_resource_schema_history` - Schema 变更历史

迁移脚本: `migrations/mariadb/`

### Schema 字段类型系统

VEGA 统一类型: `integer`, `unsigned_integer`, `float`, `decimal`, `string`, `text`, `date`, `datetime`, `time`, `boolean`, `binary`, `json`, `vector`

字段特征 (Features): `keyword`, `fulltext`, `vector`

## Frontend

React 19 + Vite + Tailwind CSS + React Router

路由:
- `/` - Dashboard
- `/catalogs` - 目录管理
- `/resources` - 资源列表
- `/resources/:id` - 资源详情
- `/query` - 查询工作台

## Development Conventions

- 所有代码注释使用中文
- 依赖注入通过构造函数，接口定义在 `interfaces/`
- 错误码定义在 `errors/`，国际化在 `locale/`
- 新增连接器类型需在 `logics/connectors/factory/` 注册
