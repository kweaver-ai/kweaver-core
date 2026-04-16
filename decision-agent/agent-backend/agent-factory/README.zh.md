# Agent Factory

[English](README.md) | 中文

Agent Factory 是 Decision Agent 中负责 Agent 配置、发布、API 暴露和运行编排元数据管理的 Go 服务。本文重点放在当前 API Chat 对接指南，以及与 OpenAPI / DIP 页面相关的接入说明。

## 快速入口

- 对外 API 文档说明：[docs/api/README.md](./docs/api/README.md)
- 静态 OpenAPI HTML：[docs/api/agent-factory.html](./docs/api/agent-factory.html)
- 静态 OpenAPI JSON：[docs/api/agent-factory.json](./docs/api/agent-factory.json)
- 静态 OpenAPI YAML：[docs/api/agent-factory.yaml](./docs/api/agent-factory.yaml)
- API Chat 入口代码：[src/driveradapter/api/httphandler/agenthandler/api_chat.go](./src/driveradapter/api/httphandler/agenthandler/api_chat.go)
- 统一 Chat 服务：[src/domain/service/agentrunsvc/chat.go](./src/domain/service/agentrunsvc/chat.go)

常用命令：

```bash
make gen-api-docs
make validate-api-docs
make goTest
make ciLint
make ciLintFix
make fmt
```

## 项目结构

Agent Factory 是一个采用 DDD + 六边形架构的 Go 服务。

核心目录如下：
- `main.go`
  - 进程入口；初始化 OpenTelemetry 并启动 HTTP 服务
- `conf/`
  - 运行时配置与本地示例配置，包含数据库 / Redis / 端口 / Swagger 开关等
- `src/boot/`
  - 配置、DB、Redis、日志、权限、审计等启动初始化逻辑
- `src/domain/`
  - 领域层：实体、服务、聚合、值对象、枚举与核心业务逻辑
- `src/port/`
  - driver / driven 接口定义，用于保持领域逻辑与适配器解耦
- `src/drivenadapter/`
  - 基础设施适配器，例如数据库访问、外部 HTTP 访问
- `src/driveradapter/api/`
  - HTTP handler 与请求 / 响应 DTO
- `src/infra/server/httpserver/`
  - HTTP 服务装配、路由注册、中间件挂载与优雅关闭生命周期
- `cmd/openapi-docs/` 与 `internal/openapidoc/`
  - OpenAPI 生成、比对、校验与静态文档渲染
- `docs/`
  - 对接文档、截图与对外 OpenAPI 产物

从路由视角看，服务主要分为两类接口面：
- Management 侧接口
  - 主要是 V3 handler，负责 agent 配置、发布、广场、产品、分类等
- Run 侧接口
  - 主要是 V1 handler，负责 chat、API Chat、conversation、session 等运行态能力

## 如何运行项目

Agent Factory 实际上有两种比较实用的运行方式。


### 在项目目录运行

最小直跑流程：

1. 准备本地配置
   - 从 `conf/agent-factory.example.yaml` 出发
   - 重点关注：
     - `project.port`（示例配置里是 `30777`）
     - `db.*`
     - `redis.*`
     - `agent_executor.*`
     - `enable_swagger`
     - 本地开发开关，例如 `auth_enable`、`switch_fields.disable_biz_domain` 以及 mock 相关配置
2. 确保依赖可达
   - MySQL
   - Redis
   - Agent Executor
3. 启动服务

直接启动命令可以是：

```bash
go run ./
```

由于 `src/boot/init.go` 会在启动时初始化 DB、Redis、日志、权限、业务域和可选审计，因此只有在相关配置和依赖都已准备好的情况下，直跑方式才会成功。一些本地开发开关（如 `auth_enable`、`switch_fields.*` 等）可以帮助你跳过某些依赖。

## 开发与校验

代码质量：

```bash
make fmt
make ciLint
make ciLintFix
```

测试：

```bash
make goTest
make ut
make utExclude
```

OpenAPI 文档：

```bash
make gen-api-docs
make validate-api-docs
make compare-api-docs
```

补充说明：
- `main.go` 中声明的 Swagger host 是 `localhost:30777`，`conf/agent-factory.example.yaml` 里的示例端口也是 `30777`
- Run 侧公开接口基路径是 `/api/agent-factory/v1`
- Run 侧内部接口基路径是 `/api/agent-factory/internal/v1`

## API Chat 对接

完整的 API Chat 对接说明请查看仓库 `docs/` 根目录下的文档：

- English: [docs/api-chat-integration.md](./docs/api-chat-integration.md)
- 中文: [docs/api-chat-integration.zh.md](./docs/api-chat-integration.zh.md)

该文档包含：
- `agent_id` / `agent_key` / `app_key` 的区别
- 普通 Chat 与 API Chat 的路由差异
- 页面真实请求示例
- 请求 / 响应结构说明
- 完整快照流与增量流的说明
- 过程态 / 工具调用结构说明
- API 文档入口
