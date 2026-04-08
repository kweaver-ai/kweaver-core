# 更新日志

本文件记录项目的重要变更。

## [0.2.1] - 2026-04-07

### 优化

- 统一 `agent-observability` 与 `otelcol-contrib` Helm Chart 的镜像 values 结构，改为使用 `image.registry`、`image.repository` 与 `image.tag`，以支持离线打包从默认 values 中提取镜像。
- 更新 `agent-observability` 发布流程，在打包 Chart 前将解析出的发布版本同步写入 Chart 默认 `image.tag`。

### 升级说明

- 如果部署时将包含 registry 的完整镜像地址覆盖到 `image.repository`，需要按新的 Chart values 结构拆分为 `image.registry` 与 `image.repository`。

## [0.2.0] - 2026-03-31

### 优化

- 更新 `.github/workflows/` 下的发布流程，统一改用迁移后的 `trace-ai/` 路径，覆盖版本读取、Go 模块定位、Docker 构建上下文与 Helm Chart 打包路径。

### 文档

- 在 `README.md` 中补充 `opentelemetry-collector-contrib` 多架构 manifest 的创建与校验命令，便于后续镜像发布复用。

## [0.1.1] - 2026-03-27

### 优化

- 将按 `conversation_id` 查询 Trace 的默认结果上限提升到 `1000`，使单次请求可以返回更多匹配记录。

## [0.1.0] - 2026-03-25

项目首次发布。

### 新增

- 新增 `agent-observability` 服务，用于从 OpenSearch 查询智能体 Trace 数据。
- 新增 Trace 查询接口，支持原始 DSL 检索和按 `conversation_id` 查询，并生成 Swagger 文档。
- 新增 `agent-observability` 的 Docker、Helm 和 GitHub Actions 构建发布流程。
- 新增 `otelcol-contribute-chart` Helm Chart，用于在 Kubernetes 中部署 OpenTelemetry Collector Contrib。
- 新增 Collector 默认 OTLP 接入与 OpenSearch 导出配置，支持 Trace 和 Log 链路处理。
- 新增仓库级中英文 README，说明 Tracing AI 的定位、架构、核心能力与快速开始方式。

### 文档

- 新增 `agent-observability/docs/` 下的 Agent 链路系统文档，包括 PRD、设计文档、API 描述与 Swagger 产物。
