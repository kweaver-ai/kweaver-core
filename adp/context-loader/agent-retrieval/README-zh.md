# Context Loader

English: [README.md](README.md).

## 模块概述

Agent Retrieval 是 ADP context-loader 中的检索服务组件，面向知识网络语义检索、行动召回、逻辑属性解析以及 MCP（Model Context Protocol）服务。该模块与 ontology-query、ontology-manager、算子集成等后端服务集成，为智能体提供知识检索与工具调用能力。

## 核心功能

核心功能为六个工具及 Context Loader MCP 服务：

1. **kn_search**：知识网络搜索，支持概念与实例级语义检索
2. **kn_schema_search**：知识网络 Schema 搜索
3. **query_object_instance**：按对象类型查询知识网络中的实例
4. **query_instance_subgraph**：按深度与关系配置查询实例周边子图
5. **get_action_info**：获取行动元数据与 Schema，供算子/智能体工具使用
6. **get_logic_properties_values**：解析逻辑属性并获取属性值
7. **Context Loader MCP Server**：对外提供上述六个工具，供 Cursor、Claude Desktop 等 MCP 客户端调用

## 技术架构

### 技术栈
- **语言**：Go 1.24
- **Web 框架**：Gin 1.11
- **可观测**：OpenTelemetry
- **缓存**：Redis（可选，支持哨兵模式）
- **校验**：go-playground/validator

### 项目结构
```
server/
├── driveradapters/     # HTTP/MCP 适配层
│   ├── knactionrecall/
│   ├── knlogicpropertyresolver/
│   ├── knontologyjob/
│   ├── knqueryobjectinstance/
│   ├── knquerysubgraph/
│   ├── knretrieval/
│   ├── knsearch/
│   ├── mcp/            # MCP 服务与 Schema
│   └── mcpproxy/
├── drivenadapters/     # 后端/数据访问
│   ├── ontology_query.go
│   ├── ontology_manager.go
│   ├── operator_integration.go
│   ├── data_retrieval.go
│   ├── agent_app.go
│   └── ...
├── interfaces/         # 端口接口定义
├── logics/             # 业务逻辑
│   ├── knactionrecall/
│   ├── knlogicpropertyresolver/
│   ├── knquerysubgraph/
│   ├── knrerank/
│   ├── knretrieval/
│   └── knsearch/
├── infra/              # 配置、日志、遥测、国际化
│   ├── config/
│   ├── logger/
│   ├── telemetry/
│   └── localize/
└── main.go
```

## 快速开始

### 环境要求
- Go 1.24+
- 可访问的 ontology-query、ontology-manager 等后端（或通过 ktctl 连接远程集群）

### 本地开发

1. **克隆**
   ```bash
   git clone <仓库地址>
   cd context-loader/agent-retrieval
   ```

2. **配置**
   编辑 `server/infra/config/agent-retrieval.yaml`（及按需 `agent-retrieval-secret.yaml`）中的服务地址与端口。若本地连远程集群，需先配置 ktctl，再执行 `make build`，以便将配置复制到 `/sysvol/config` 与 `/sysvol/secret`。

3. **依赖**
   ```bash
   go mod download
   ```

4. **运行**
   ```bash
   # 方式一：配合 ktctl 与 make（推荐开发）
   make build

   # 方式二：仅启动进程（使用默认配置路径）
   cd server && go run main.go
   ```
   默认监听 `http://0.0.0.0:30779`。

### Make 目标
- `make help`：查看目标说明
- `make lint`：执行 golangci-lint
- `make test`：单元测试
- `make test-cover`：带覆盖率的测试
- `make generate-mock`：生成 mock
- `make helm-template`：渲染 Helm 模板
- `make preview`：预览 API 文档（在 `docs/` 下）

### Docker

需在 **仓库根目录**（context-loader）下构建，以便 Dockerfile 能复制 agent-retrieval 目录：

```bash
# 在 context-loader 根目录执行
docker build -t agent-retrieval:latest -f agent-retrieval/docker/Dockerfile .
```

运行：

```bash
docker run -d -p 30779:30779 --name agent-retrieval agent-retrieval:latest
```

### Kubernetes / Helm

在 `agent-retrieval` 目录下：

```bash
helm install agent-retrieval ./helm/agent-retrieval/
# 或仅渲染模板：
helm template agent-retrieval ./helm/agent-retrieval -n <命名空间> -f ./helm/agent-retrieval/values.yaml
```

## 配置说明

主配置文件：`server/infra/config/agent-retrieval.yaml`。

主要配置项：
- **project**：`host`、`port`（默认 30779）、`language`、`logger_level`、`debug`
- **ontology_query**、**ontology_manager**、**data_retrieval**、**operator_integration**：各后端服务地址与端口
- **oauth**：Hydra OAuth（对外 API）
- **redis**：可选 Redis（如哨兵模式）缓存
- **concept_search_config**、**deploy_agent**、**rerank_llm**：功能相关配置

敏感信息：`server/infra/config/agent-retrieval-secret.yaml`（按需加入版本控制忽略）。

## 监控与运维

- **就绪探针**：`GET /health/ready`
- **存活探针**：`GET /health/alive`
- **日志**：结构化日志，级别与输出方式可在配置中调整
- **链路追踪**：OpenTelemetry，在 `observability` 配置中启用与配置

## 开发规范

- 遵循 Go 官方编码规范与项目约定
- 保持清洁架构：driver/driven 适配层、接口与业务逻辑分层清晰
- 提交前执行 `make lint` 与 `make test`（或 `make test-cover`）

## 版本

版本号见项目根目录下的 `VERSION` 文件。

## 支持与联系

- **团队**：AISHU ADP
- **文档**：见 `docs/`（PRD、API、发布说明等）。
