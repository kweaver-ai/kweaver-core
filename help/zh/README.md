# 📘 KWeaver Core 文档

KWeaver Core 为**纯后台**平台，请通过 CLI、各语言 SDK 或 HTTP API 操作各子系统。

---

## 🚀 入门

**部署：** 完整安装以 **Linux** 为主；**macOS** 仅可选本机 kind 验证 — 见 [`deploy/dev/README.md`](../../deploy/dev/README.md)。

| 文档 | 说明 |
| --- | --- |
| [安装与部署](install.md) | 环境要求、`deploy.sh` 安装与安装后检查 |
| [快速开始](quick-start.md) | 从部署到首次 BKN 与智能体操作的端到端路径 |

---

## 🧩 模块

| 文档 | 说明 |
| --- | --- |
| 📂 [数据源管理](datasource.md) | 数据库连接、表发现、CSV 导入与生命周期 |
| 🧠 [模型管理](model.md) | LLM、Embedding、Reranker 的注册与管理 |
| 🕸️ [BKN 引擎](bkn.md) | 业务知识网络 — 对象类、关系类、行动类与实例 |
| 🗄️ [VEGA 引擎](vega.md) | 数据虚拟化 — 连接、模型、视图与统一查询 |
| 📚 [Context Loader](context-loader.md) | 面向智能体的上下文组装 |
| 🛠️ [Execution Factory](execution-factory.md) | 工具、算子与技能 |
| 🔁 [Dataflow](dataflow.md) | 流程编排与自动化 |
| 🤖 [Decision Agent](decision-agent.md) | 目标驱动智能体、运行时与可观测 |
| 🔭 [Trace AI](trace-ai.md) | 链路追踪、指标与证据链式可观测 |
| 🔐 [Info Security Fabric](isf.md) | 身份、权限、策略与审计（启用时） |
| 🛡️ [平台管理员工具](install.md#-完整安装后的管理员工具kweaver-admin) | `kweaver-admin` — 用户/组织/角色/模型/审计（完整安装后） |

---

## 💬 交流群

<img src="../qrcode.png" width="200" alt="KWeaver 交流群二维码" />

> 💡 CLI 安装：业务用户用 `npm install -g @kweaver-ai/kweaver-sdk`；平台管理员另装 `npm install -g @kweaver-ai/kweaver-admin`。更详细的集群运维说明以随产品提供的部署文档为准。
