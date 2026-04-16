# KWeaver Core 文档

KWeaver Core 为**纯后台**平台，请通过 CLI、各语言 SDK 或 HTTP API 操作各子系统。

## 入门

- [环境要求](installation/prerequisites.md) — 硬件、操作系统、网络与工具
- [部署](installation/deploy.md) — 使用 `deploy.sh` 一键安装（含安装后检查）
- [快速开始](quick-start.md) — 从部署到首次 BKN 与智能体操作的端到端路径

## 模块

| 文档 | 说明 |
| --- | --- |
| [数据源管理](datasource.md) | 数据库连接、表发现、CSV 导入与生命周期 |
| [模型管理](model.md) | LLM、Embedding、Reranker 的注册与管理 |
| [BKN 引擎](bkn.md) | 业务知识网络 — 对象类型、关系、动作与实例 |
| [VEGA 引擎](vega.md) | 数据虚拟化 — 连接、模型、视图与统一查询 |
| [Context Loader](context-loader.md) | 面向智能体的上下文组装 |
| [Execution Factory](execution-factory.md) | 工具、算子与技能 |
| [Dataflow](dataflow.md) | 流程编排与自动化 |
| [Decision Agent](decision-agent.md) | 目标驱动智能体、运行时与可观测 |
| [Trace AI](trace-ai.md) | 链路追踪、指标与证据链式可观测 |
| [Info Security Fabric](isf.md) | 身份、权限、策略与审计（启用时） |

## 交流群

<img src="../qrcode.png" width="200" alt="KWeaver 交流群二维码" />

更完整的部署说明见 [部署指南](../../deploy/README.zh.md)。`kweaver` CLI 与各语言 SDK 见 [kweaver-sdk](https://github.com/kweaver-ai/kweaver-sdk)。
