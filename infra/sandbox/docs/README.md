# Documentation

项目文档按产品、设计、API、开发、运维五类组织，避免需求文档、架构文档和操作手册混放在根目录。

## Structure

```text
docs/
├── README.md
├── product/
│   ├── roadmap.md
│   ├── prd/
│   └── use-cases/
├── design/
│   ├── architecture/
│   ├── modules/
│   ├── features/
│   ├── decisions/
│   └── archive/
├── api/
│   ├── rest/
│   ├── grpc/
│   └── websocket/
├── dev/
├── ops/
└── assets/
```

## Navigation

### Product

- [产品路线图](./product/roadmap.md)
- [PRD 模板](./product/prd/template.md)
- [Session Python 依赖管理 PRD](./product/prd/session-python-dependency-management.md)
- [场景：PTC 数据分析与上下文问答](./product/use-cases/ptc-data-analysis-and-context-qa.md)
- [背景：统一沙箱临时区](./product/use-cases/unified-sandbox-background.md)

### Design

- [架构总览](./design/architecture/overview.md)
- [系统上下文](./design/architecture/system-context.md)
- [逻辑架构与流程](./design/architecture/logical-architecture.md)
- [安全与性能](./design/architecture/security-and-performance.md)
- [存储架构](./design/architecture/storage-architecture.md)
- [模块设计：Control Plane](./design/modules/control-plane.md)
- [模块设计：Container Scheduler](./design/modules/container-scheduler.md)
- [模块设计：Executor](./design/modules/executor.md)
- [模块设计：Python Dependencies](./design/modules/python-dependencies.md)
- [需求设计：Session Python 依赖管理](./design/features/session-python-dependency-management.md)
- [需求设计：MCP Server](./design/features/mcp-server-implementation.md)
- [Design 模板](./design/features/template.md)
- [ADR 模板](./design/decisions/template.md)

### API

- [API 总览](./api/README.md)
- [REST OpenAPI](./api/rest/sandbox-openapi.json)
- [Execute Sync OpenAPI](./api/rest/execute-sync-openapi.yaml)

### Dev

- [环境准备](./dev/setup.md)
- [构建](./dev/build.md)
- [测试](./dev/test.md)
- [发布](./dev/release.md)
- [贡献指南](./dev/contributing.md)
- [项目结构](./dev/project-structure.md)

### Ops

- [部署](./ops/deploy.md)
- [配置](./ops/config.md)
- [监控](./ops/monitoring.md)
- [故障排查](./ops/troubleshooting.md)
- [历史文档：监控与部署](./ops/monitoring-and-deployment.md)
- [历史文档：Runtime Node 注册](./ops/runtime-node-registration.md)

## PRD <-> Design Linking Convention

每个需求默认生成两份文档：

1. `docs/product/prd/<feature>.md`
2. `docs/design/features/<feature>.md`

两份文档必须使用同一个 `<feature>` slug，并互相链接。

### PRD 必备区块

```md
## 关联设计

- Design: [<标题>](../../design/features/<feature>.md)
```

### Design 必备区块

```md
## 关联需求

- PRD: [<标题>](../../product/prd/<feature>.md)
```

建议在文档创建时直接从模板复制，避免后补链接。
