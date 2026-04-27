# Tracing AI

[English](README.md) | 中文

[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE.txt)

Tracing AI 是一套面向 LLM 应用和智能体系统的可验证性与可观测性框架，目标是通过全链路观测、结构化关联和证据化查询，把 AI 从“黑盒推理”推进到“确定性生产力”。

当前仓库主要包含两个核心组成部分：

- `agent-observability`：基于 Go 实现的 Trace 查询服务，用于从 OpenSearch 检索智能体链路
- `otelcol-contribute-chart`：用于在 Kubernetes 中部署 OpenTelemetry Collector Contrib 的 Helm Chart，负责接收 OTLP 数据并导出到 OpenSearch

## 项目定位

传统应用追踪只能告诉我们“请求是否失败”，但在 AI 系统里，开发者还需要回答更多问题：

- 模型到底看到了什么输入
- 中间调用了哪些工具
- 检索命中了哪些知识
- 某个结论是否有依据、是否可回溯
- 延迟和错误究竟出现在执行链路的哪个节点

Tracing AI 就是为这些问题设计的基础设施。

## 核心特性

### 当前仓库已具备的能力

- 基于 OpenTelemetry 的标准化采集链路
- 基于 OpenTelemetry Collector 的 OTLP 接入能力
- 面向 OpenSearch 的 Trace / Log 导出能力
- 面向 Agent Trace 的查询服务
- Swagger 文档、Docker 构建、Helm 打包与 GitHub Actions 发布流程

### Tracing AI 的目标能力

以下能力代表 Tracing AI 的总体建设方向，其中一部分已经由当前架构打底，另一部分将在后续持续补齐：

- 全链路执行轨迹观测：覆盖输入输出、工具调用、知识检索以及推理步骤
- 决策依据穿透与证据追溯：把 AI 输出与数据来源、知识条目、上下文执行证据关联起来
- 可视化时间轴回放：支持定位复杂 Agent 执行中的耗时节点与坏案例
- 从 Trace 到 Eval 的闭环优化：将失败链路沉淀为评测用例
- 智能根因分析：为多智能体复杂系统提供自动化故障定位基础

## 技术架构

Tracing AI 遵循 OpenTelemetry / OTLP 开放协议，便于以标准化、低侵入方式接入遥测数据。

- `Trace`：一次 AI 交互或一次任务执行的完整生命周期
- `Span`：链路中的单个操作单元，如一次模型调用、一次检索或一次工具执行
- `Collector`：负责接收 OTLP 数据、执行批处理与路由，并导出到底层存储
- `Query Service`：对外提供 Trace 检索与分析接口
- `Storage`：当前仓库以 OpenSearch 为主，整体架构可扩展到更适合 AI 大规模数据的底层存储

当前仓库落地的链路形态如下：

```text
LLM App / Agent
  -> OTLP
OpenTelemetry Collector
  -> OpenSearch
agent-observability
  -> Trace 查询接口 / Swagger
```

## 仓库结构

```text
.
|-- agent-observability/
|   |-- main.go
|   |-- Dockerfile
|   |-- Makefile
|   |-- charts/agent-observability/
|   `-- docs/
|-- otelcol-contribute-chart/
|   |-- charts/otelcol-contrib/
|   `-- scripts/
`-- .github/workflows/
```

## 组件说明

### agent-observability

`agent-observability` 是当前仓库中的 Trace 查询服务，当前已提供：

- `POST /api/v1/traces/_search`：将原始 OpenSearch DSL 代理到配置的 trace index
- `GET /api/v1/traces/by-conversation?conversation_id=...`：按会话维度查询 trace
- `/swagger/`：Swagger 文档访问入口

本地开发：

```bash
cd agent-observability
make test
make gen-swag
make docker-build
```

Helm 部署示例：

```bash
helm upgrade --install agent-observability agent-observability/charts/agent-observability \
  --set image.repository=swr.cn-east-3.myhuaweicloud.com/kweaver-ai/agent-observability \
  --set image.tag=0.1.0 \
  --set opensearch.endpoint=http://opensearch-read.resource.svc.cluster.local:9200 \
  --set opensearch.auth.enabled=false \
  -n observability --create-namespace
```

### otelcol-contribute-chart

`otelcol-contribute-chart` 用于部署 OpenTelemetry Collector Contrib，当前提供：

- 基于 Deployment 的 Collector 部署方式
- OTLP gRPC / HTTP 接收器
- OpenSearch 导出器与可选 Basic Auth
- GHCR OCI Chart 发布流程

本地校验：

```bash
helm lint otelcol-contribute-chart/charts/otelcol-contrib
helm template otelcol-contrib otelcol-contribute-chart/charts/otelcol-contrib
```

安装示例：

```bash
helm upgrade --install otelcol-contrib otelcol-contribute-chart/charts/otelcol-contrib \
  -n observability \
  --create-namespace \
  --set opensearchExporter.http.endpoint=http://opensearch-read.resource.svc.cluster.local:9200
```

创建并发布 Collector 多架构 manifest：

```bash
docker buildx imagetools create \
  -t swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/opentelemetry-collector-contrib:0.148.0 \
  swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.148.0 \
  swr.cn-north-4.myhuaweicloud.com/ddn-k8s/ghcr.io/open-telemetry/opentelemetry-collector-releases/opentelemetry-collector-contrib:0.148.0-linuxarm64
```

校验 manifest 平台信息：

```bash
docker buildx imagetools inspect \
  swr.cn-east-3.myhuaweicloud.com/kweaver-ai/dip/opentelemetry-collector-contrib:0.148.0
```

## 快速开始

1. 部署 `otelcol-contrib`，作为 OTLP 接入与数据导出组件。
2. 让你的 LLM 应用或 Agent Runtime 通过 OTLP 上报遥测数据。
3. 部署 `agent-observability` 查询服务。
4. 通过以下接口查询链路：

```text
POST /api/v1/traces/_search
GET  /api/v1/traces/by-conversation
GET  /swagger/index.html
```

## 业务价值

- 提升 AI 应用可信度，让输出具备可追溯依据
- 降低不确定性风险，在关键节点分析异常行为与潜在幻觉
- 显著缩短问题排查周期，让生产链路可检索、可定位
- 用真实轨迹沉淀评测数据，让后续模型和系统迭代更可量化

## 相关文档

- `agent-observability/README.md`
- `otelcol-contribute-chart/README.md`
- `agent-observability/docs/design/agent-tracing-system-design.md`
- `agent-observability/docs/prd/agent-tracing-system-prd.md`

## License

Apache 2.0，详见 `LICENSE.txt`。
