# Flow Automation (数据流编排)

[中文](README.zh.md) | [English](README.md)

Flow Automation 是一个负责编排和执行自动化数据管道的核心服务。它管理数据流的完整生命周期，集成各种基础设施组件，并提供数据管道管理的 API 接口。

## 核心架构

本项目采用六边形架构（Hexagonal Architecture），将核心业务逻辑与外部接口分离：

```
┌─────────────────────────────────────────────────────────┐
│                    Driver Adapters                      │
│          (HTTP API, Message Queue Consumers)            │
│  ┌───────────────────────────────────────────────────┐  │
│  │              Core Business Logic                  │  │
│  │         (logics/, module/, pkg/)                  │  │
│  └───────────────────────────────────────────────────┘  │
│                   Driven Adapters                       │
│        (Database, External Services, Cache)             │
└─────────────────────────────────────────────────────────┘
```

### 目录结构说明

- **`logics/`**: 核心业务逻辑层，包含数据管道管理、执行器、触发器等业务逻辑
- **`driveradapters/`**: 驱动适配器层（入站），处理 HTTP 请求、消息队列消费等
- **`drivenadapters/`**: 被驱动适配器层（出站），与数据库、外部服务、缓存等交互
- **`module/`**: 领域模块，包含初始化和配置
- **`pkg/`**: 共享工具包和实体定义

## 主要功能

### 1. 数据管道管理 (DAG Management)
- **创建与编辑**: 支持创建、更新、删除数据管道（DAG）
- **版本控制**: 数据管道版本管理
- **执行管理**: 启动、取消、暂停管道实例
- **调试模式**: 支持单步调试和完整调试
- **批量操作**: 批量查询和操作数据管道

### 2. 自定义执行器 (Custom Executors)
- **执行器管理**: 创建、更新、删除自定义执行器
- **动作管理**: 为执行器添加、修改、删除动作
- **权限控制**: 基于用户权限的执行器访问控制
- **Agent 导入**: 支持从应用商店导入 Agent

### 3. 触发器系统 (Trigger System)
- **定时触发**: 基于 Cron 表达式的定时任务触发
- **事件触发**: 支持外部事件触发数据管道
- **手动触发**: 支持用户手动触发数据管道执行
- **回调处理**: 处理异步任务回调

### 4. 认证与授权 (Authentication & Authorization)
- **Token 认证**: 基于 Hydra 的 OAuth2 认证
- **权限验证**: 细粒度的资源访问权限控制
- **业务域隔离**: 多租户业务域隔离
- **最小依赖模式**: 支持通过 `auth.enabled` / `SERVER_AUTH_ENABLED` 关闭 ISF 依赖；关闭后服务可在无 Hydra、Authentication、Authorization、User Management 的情况下运行，但会绕过原有权限与业务域隔离

### 5. 安全策略 (Security Policies)
- **访问控制**: API 访问安全策略
- **数据隔离**: 租户数据隔离
- **审计日志**: 操作审计追踪

### 6. 可观测性 (Observability)
- **指标监控**: Prometheus 指标暴露
- **链路追踪**: OpenTelemetry 分布式追踪
- **日志管理**: 结构化日志和流式日志
- **健康检查**: 服务健康状态检查

### 7. 数据流管理 (Data Flow)
- **数据连接**: 管理数据源连接配置
- **数据转换**: 支持数据流转换和处理

### 8. 算子管理 (Operators)
- **算子注册**: 注册和管理数据处理算子
- **算子执行**: 执行各类数据处理算子

## 技术栈

- **语言**: Go 1.24.0
- **Web 框架**: Gin
- **数据库**: 
  - MongoDB (数据管道定义和实例存储)
  - MariaDB/MySQL (关系型数据)
  - Redis (缓存和分布式锁)
- **消息队列**: Kafka
- **配置管理**: Viper, godotenv
- **可观测性**: OpenTelemetry, Prometheus
- **容器编排**: Kubernetes (Helm Charts)

## 前置要求

- Go 1.24+
- Docker & Docker Compose (本地基础设施)

## 快速开始

详细的本地开发环境搭建指南（包括运行基础设施和模拟服务），请参考 [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md)。

### 快速运行

1. **启动依赖服务**:
   ```bash
   docker-compose up -d
   ```

2. **运行应用**:
   ```bash
   go run main.go
   ```

应用将启动两个服务：
- 公共 API 服务: `http://localhost:8082`
- 私有 API 服务: `http://localhost:8083`

## 配置说明

应用使用环境变量进行配置：

- **`.env`**: 默认配置信息（提交到 Git）
- **`.env.local`**: 本地开发覆盖配置（不提交到 Git）

主要配置项：
- `API_SERVER_PORT`: 公共 API 端口（默认 8082）
- `API_SERVER_PRIVATE_PORT`: 私有 API 端口（默认 8083）
- `DB_TYPE`（`dbType`/`db_type`）: 用于选择流程元数据的底层存储后端，可选值为 `mongodb` 或 `mysql`
  - `mongodb`: 使用 MongoDB 存储 DAG 定义、执行实例等文档型流程数据
  - `mysql`: 使用基于 MySQL/MariaDB 的仓库存储流程元数据，适用于关系库存储部署或迁移后的场景
    如果服务之前使用的是 MongoDB 存储，切换到 `mysql` 前需要先迁移历史流程数据，例如执行 `python3 migrations/mariadb/0.4.0/pre/02-mongodb_to_mysql_migration.py`
- `SERVER_AUTH_ENABLED`: 用于控制 `flow-automation` 是否启用鉴权相关 ISF 依赖
  - `true`（默认）: 使用真实的 Hydra、Authentication、Authorization、User Management 服务
  - `false`: 使用内置 mock 适配器，以最小依赖模式运行服务
  - 在 Helm 部署中，该开关由 `auth.enabled` 统一管理
  - 该模式适用于本地开发、演示或最小依赖部署场景，会绕过原有的权限控制与业务域隔离逻辑
- `MONGODB_HOST`: MongoDB 地址
- `REDIS_HOST`: Redis 地址
- `KAFKA_BROKERS`: Kafka 集群地址

本次解耦设计的详细说明见 [#273 Dataflow解耦ISF](docs/features/%23273%20Dataflow%E8%A7%A3%E8%80%A6ISF/Dataflow%E8%A7%A3%E8%80%A6ISF.md)。

## 项目结构

```text
.
├── common/             # 通用工具和常量
├── conf/               # 配置文件
├── docs/               # 文档
├── drivenadapters/     # 出站适配器（数据库、外部 API）
├── driveradapters/     # 入站适配器（HTTP、事件消费者）
│   ├── admin/          # 管理接口
│   ├── alarm/          # 告警接口
│   ├── auth/           # 认证接口
│   ├── executor/       # 执行器接口
│   ├── mgnt/           # 数据管道管理接口
│   ├── trigger/        # 触发器接口
│   └── ...
├── helm/               # Helm 部署图表
├── logics/             # 业务逻辑
│   ├── mgnt/           # 数据管道管理逻辑
│   ├── executor/       # 执行器逻辑
│   ├── cronjob/        # 定时任务逻辑
│   └── ...
├── mock-server/        # 本地开发模拟服务
├── module/             # 领域模块
├── pkg/                # 共享包
│   ├── actions/        # 动作定义
│   ├── entity/         # 实体定义
│   └── ...
├── resource/           # 静态资源
├── schema/             # 数据库 Schema
├── scripts/            # 辅助脚本
├── store/              # 数据存储接口
└── utils/              # 通用工具
```

## API 文档

主要 API 端点：

### 数据管道管理
- `POST /api/automation/v1/dags` - 创建数据管道
- `PUT /api/automation/v1/dags/:id` - 更新数据管道
- `GET /api/automation/v1/dags/:id` - 获取数据管道详情
- `DELETE /api/automation/v1/dags/:id` - 删除数据管道
- `GET /api/automation/v1/dags` - 列出数据管道

### 数据管道执行
- `POST /api/automation/v1/dags/:id/run` - 执行数据管道
- `POST /api/automation/v1/dags/:id/cancel` - 取消执行
- `GET /api/automation/v1/dags/:id/instances` - 查询执行实例

### 执行器管理
- `POST /api/automation/v1/executors` - 创建执行器
- `PUT /api/automation/v1/executors/:id` - 更新执行器
- `GET /api/automation/v1/executors` - 列出执行器

## 部署

项目使用 Helm 进行 Kubernetes 部署：

```bash
cd helm/flow-automation
helm install flow-automation . -f values.yaml
```

## 开发指南

1. **代码风格**: 遵循 Go 标准代码规范，使用 `golangci-lint` 进行代码检查
2. **测试**: 使用 `go test` 运行单元测试
3. **Mock 生成**: 使用 `go.uber.org/mock` 生成测试 Mock
