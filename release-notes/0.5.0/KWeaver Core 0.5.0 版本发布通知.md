# KWeaver Core 0.5.0 版本发布通知

---

## 版本概述

KWeaver Core 0.5.0 是平台能力全面提速的重要版本。本次发布以 **KWeaver SDK 正式上线**为标志，将平台核心能力首次向开发者完整开放；**TraceAI** 完成全链路 TraceID 机制建设，智能体执行轨迹数据采集与查询能力正式上线；**Context Loader** 完成 Token 压缩（≥30%）与行动驱动统一纳管；**VEGA** 与 **Dataflow** 在数据构建与数据湖解耦方向均取得重要突破。与此同时，BKN、执行工厂、Context Loader、Dataflow 等核心组件完成与业务域及 ISF 的架构解耦，平台整体模块化程度与部署灵活性进一步提升。

---

## 核心亮点速览

**1. KWeaver SDK 正式发布，平台能力全面开放**

KWeaver SDK 首次正式发布，同时提供 TypeScript 与 Python 双语言实现，涵盖 CLI 工具、SDK 库与配套 Skill。开发者可通过统一客户端访问业务知识网络、Agent 对话、数据管理、Context Loader 等平台核心能力；配套 Skill 支持 AI 智能体直接调用平台 API，全程自动获取技术参数，无需用户手动输入。

**2. TraceAI 全链路可观测正式上线**

基于 OpenTelemetry 标准完成 TraceAI v0.1.1 正式发布，建立全链路 TraceID 机制，实现智能体执行轨迹数据的统一采集与管理。提供统一轨迹资源查询接口，支持按 TraceID 获取执行轨迹及被访问 BKN 资源信息，AI 可自动分析执行证据链。

**3. Context Loader 深度优化，Token 压缩 ≥30%**

全面优化 Context Loader 返回内容结构，减少冗余信息传递，Token 平均压缩率 ≥30%，智能体执行准确率 ≥90%。行动召回机制完成架构升级，由直接对接执行工厂/MCP 改为通过 BKN 行动驱动统一接入，为后续全链路风险管控奠定统一的行动管控基础。

**4. VEGA 数据构建能力闭环**

VEGA 新增 Resource 数据构建能力，支持全量与增量索引建立及自定义视图，Catalog 扩展支持 PostgreSQL 数据源与扫描任务管理，数据虚拟化层的数据构建能力形成完整闭环。

**5. Dataflow 解耦内容数据湖**

Dataflow 解耦内容数据湖，引入 DFS 文件协议，支持外部文件直接触发数据处理流程，大幅降低 Dataflow 的部署依赖与使用门槛。

---

## 详细功能说明

### 【KWeaver SDK】

KWeaver SDK 是 KWeaver 平台面向开发者的官方集成工具包，本版本首次正式发布，同时提供 TypeScript 与 Python 双语言实现，覆盖 CLI 工具、SDK 库与配套 Skill，将平台核心能力全面开放给开发者与 AI 智能体。

**1. 双语言 SDK 与 CLI**

同步发布 TypeScript 和 Python 两套实现：

- **TypeScript SDK**：`npm install @kweaver-ai/kweaver-sdk`（需 Node.js 22+），提供 `KWeaverClient` 客户端，支持流式 Agent 对话与交互式 TUI；CLI 安装：`npm install -g @kweaver-ai/kweaver-sdk`
- **Python SDK**：`pip install kweaver-sdk`（需 Python 3.10+），提供 `from kweaver import KWeaverClient`；CLI 安装：`pip install kweaver-sdk[cli]`
- **统一鉴权**：支持 OAuth2 浏览器登录（`kweaver auth login`）与客户端凭证两种认证方式，满足交互式与无头（headless）两类使用场景

**2. 平台核心 API 覆盖**

SDK 封装 KWeaver 平台的主要 API 能力，开发者可通过统一客户端访问：

- **业务知识网络**：BKN 列表查询、实例检索、子图查询、行动调用
- **Agent 对话**：单次问答与流式对话，支持 Session 管理
- **数据管理**：数据源管理、数据视图、Dataflow 流程触发、CSV 导入
- **可观测性**：VEGA Catalog 查询、TraceAI 轨迹数据获取
- **Context Loader**：语义检索与上下文装配

**3. 配套 Skill**

随 SDK 同步发布对应 Skill，AI 智能体可通过 Skill 直接调用 KWeaver 平台能力，全程自动获取技术参数，无需用户手动输入。

---

### 【BKN 引擎】

0.5.0 版本在 BKN 的语义建模体系中新增"风险类"领域扩展，并完成 BKN 与业务域的架构解耦，进一步丰富知识网络的业务表达能力与部署灵活性。

**1. 风险类语义建模**

在 BKN 业务知识网络中完成"风险类"语义建模，新增风险类知识对象的建模支持，用户可在业务知识网络中定义和管理风险类相关的业务语义结构。该能力为智能体在风险识别、合规审查、异常检测等业务场景中提供标准化的知识支撑，扩展 BKN 的业务语义覆盖范围。

**2. 关系类命名约束放开**

同一业务知识网络内，关系类名称不再强制唯一，允许同名关系类存在。该调整适配了业务建模中的实际需求，避免因命名冲突导致的建模限制。

**3. BKN 与业务域及 ISF 架构解耦**

完成业务知识网络（BKN）与业务域及 ISF 的架构解耦，BKN 核心服务不再强依赖业务域配置，支持更灵活的独立部署与跨域复用。同步修复未开启认证状态下数据视图数据预览的权限判断异常问题。

---

### 【TraceAI】

0.5.0 版本完成 TraceAI v0.1.1 的正式发布，基于 OpenTelemetry 标准构建完整的可观测性基础设施，将智能体可观测能力提升至全覆盖水平。

**1. 全链路 TraceID 机制建立**

建立完整的全链路 TraceID 机制，实现智能体执行轨迹数据的统一采集与管理。每次 Agent 执行均生成唯一 TraceID，通过 TraceID 串联执行链路中的调用步骤、工具使用与 BKN 资源访问记录，为执行过程的追溯与分析提供数据基础。

**2. 基于 OpenTelemetry 的可观测性基础设施**

TraceAI 采用 OpenTelemetry 标准构建采集与存储链路：

- **采集链路**：LLM 应用/Agent 以 OTLP 协议上报轨迹数据 → OpenTelemetry Collector（Kubernetes Helm Chart 部署）→ OpenSearch 持久化存储
- **查询服务**：agent-observability 服务提供 REST API，支持原始 DSL 查询与按对话（Conversation）检索轨迹，默认返回上限 1000 条
- **Agent-Factory 集成**：Decision Agent 对话消息与 Agent 执行链路完成关联，支持从对话维度追溯完整执行记录
- **部署支持**：提供 Docker 镜像、Helm Chart 与 GitHub Actions 工作流，覆盖容器化部署全流程

**3. 统一轨迹资源查询接口**

提供统一的轨迹资源查询接口，支持：

- **按 TraceID 获取完整执行轨迹**：返回 Agent 执行过程中的全部步骤记录，包括工具调用、中间推理与最终结果
- **被访问 BKN 资源查询**：查询本次执行中访问的业务知识网络资源
- **执行证据链分析支持**：轨迹数据结构化程度满足 AI 自动分析需求，可供下游智能体读取并生成可解释的执行证据链报告

---

### 【VEGA 数据虚拟化】

0.5.0 版本在 Catalog 元数据管理与 Resource 数据构建两个方向均完成重要能力扩展。

**1. Catalog 支持 PostgreSQL 及扫描任务管理**

VEGA Catalog 扩展对 PostgreSQL 数据源的元数据采集与管理支持，用户可通过统一 Catalog 管理界面完成对 PostgreSQL 库表结构、字段信息的扫描与注册。同步新增扫描任务管理功能，支持对元数据扫描任务的创建、查看与状态追踪，提升元数据治理的可操作性与可见性。

**2. Resource 数据构建（索引建立）**

新增 Resource 数据构建能力，支持在 VEGA 虚拟化层对数据源执行向量索引构建：

- **全量构建**：对数据源全量数据执行一次完整的索引建立
- **增量构建**：基于配置的增量字段，仅同步自上次构建以来新增或变更的数据，降低构建成本

当前版本支持 MySQL 数据源，PostgreSQL 支持同步纳入规划。

**3. 支持对 Resource 定义自定义视图**

支持对 Resource 定义自定义视图，用户可灵活配置字段映射与向量索引结构，满足不同业务场景的检索需求。

**4. VEGA 与 ISF 架构解耦**

完成 VEGA 与 ISF 的架构解耦，VEGA 的数据处理流程不再依赖 ISF 服务，降低跨服务耦合，提升 VEGA 在独立部署场景下的适用性。

---

### 【Dataflow】

0.5.0 版本完成内容数据湖的彻底解耦，并引入统一文件协议，打通外部数据源接入通路。

**1. 解耦内容数据湖**

完成 Dataflow 对内容数据湖的解耦，通过引入 DFS（分布式文件服务）文件协议，Dataflow 的数据流转中间层统一基于 DFS 文件引用实现，节点之间的数据传递不再绑定特定存储服务，大幅简化 Dataflow 的部署依赖与架构复杂度。

**2. DFS 文件协议与外部数据源接入**

建立统一的 DFS 文件协议，贯通 Dataflow 各处理节点：

- **本地文件触发**：支持通过接口直接上传本地文件触发 Dataflow 流程，文件自动进入非结构化处理链路（解析、分块、向量化等）
- **外部 URL 接入**：支持传入外部数据源 URL，通过实现对应连接器即可接入任意外部数据源
- **节点全面适配**：Dataflow 各处理节点已完成对 DFS 文件协议的适配，中间数据转换过程统一基于 DFS 文件信息流转
- **底层实现**：文件子系统基于 OssGateway 实现，中间临时数据存储适配 OSS Gateway，不再依赖内置内容数据湖存储服务

**3. SQL 写入节点增强**

Dataflow SQL 写入节点新增"清空后追加"写入模式，原有"追加"与"覆盖"模式保持不变；同步优化接口参数校验逻辑，提升写入操作的可靠性与使用体验。

**4. Dataflow 与业务域及 ISF 架构解耦**

完成 Dataflow 与业务域及 ISF 的架构解耦，Dataflow 的数据处理流程不再依赖业务域及 ISF 服务，降低跨服务耦合，提升 Dataflow 在独立部署场景下的适用性。

---

### 【Context Loader】

0.5.0 版本在压缩效率与行动能力架构两个方向均完成重大升级。

**1. 返回内容结构优化，Token 压缩 ≥30%**

对 Context Loader 返回内容结构进行全面优化：

- **冗余信息精简**：移除响应中的冗余字段与重复内容，降低无效 Token 消耗
- **压缩格式支持**：新增 TOON 压缩格式输出，HTTP 和 MCP 协议均支持压缩传输；MCP 接入默认采用 TOON 格式，HTTP 接入默认保持 JSON 格式以兼容旧版本
- **量化效果**：Token 平均压缩率 ≥30%，智能体执行准确率 ≥90%，试错成本降低，推理轮数减少

**2. 行动召回通过 BKN 行动驱动统一接入**

完成行动召回架构的重要升级，改造 `get_action_info` 的召回逻辑：

- **统一接入路径**：行动召回工具由原来直接对接执行工厂/MCP 代理执行接口，全面调整为通过 BKN 行动驱动统一接入
- **统一执行管理**：召回的行动工具通过行动驱动进行统一调度与执行，行动触发记录在业务知识网络中可见
- **风险管控基础**：新架构为后续在行动驱动层统一实施风险管控（执行权限、操作审批、异常拦截）奠定架构基础

**3. Context Loader 与业务域及 ISF 架构解耦**

完成 Context Loader 与业务域及 ISF 的架构解耦，Context Loader 的上下文检索与注入流程不再依赖业务域及 ISF 权限服务，提升独立部署灵活性与请求链路性能。

---

### 【执行工厂】

0.5.0 版本新增 Agent Skill 接入能力，将平台的执行能力进一步向 Skill 生态延伸。

**1. Agent Skill 接入支持**

执行工厂新增对 Agent Skill 的接入、注册与调用支持：

- **Skill 注册**：支持将 Agent Skill 注册到执行工厂，统一纳入执行能力管理体系
- **Skill 配置**：提供 Skill 的基础配置能力，支持参数定义与运行环境配置
- **Skill 调用**：已注册的 Skill 可通过执行工厂标准接口被 Agent 调用，无需额外适配

该能力为 DIP 场景的 Skill 化落地提供核心执行支撑。


**2. 执行工厂与业务域及 ISF 架构解耦**

完成执行工厂与业务域及 ISF 的架构解耦，执行工厂的工具调度与函数执行不再强依赖业务域及 ISF 配置，支持在更灵活的多域、跨域部署场景中使用，同步修复了 Context Loader MCP 获取工具列表失败的问题。

---

### 【Decision Agent】

0.5.0 版本重点完成稳定性修复与 TraceAI 集成。

**1. 稳定性修复**

针对 0.4.0 版本反馈的多个稳定性问题完成修复：

- **长期记忆修复**：修复开启长期记忆后智能体使用报错的问题，确保长期记忆功能在新部署环境下正常运行
- **历史对话修复**：修复对话中断后从历史记录重新进入时已有答案未展示的问题，恢复历史对话的完整可见性
- **Dolphin 调度修复**：修复智能体使用时偶发的 Dolphin 调度异常，提升多智能体协同场景下的稳定性
- **对话历史信息丢失修复**：修复智能体对话过程中历史上下文信息丢失的问题

**2. TraceAI 知识证据链集成**

完成 Decision Agent 与 TraceAI 的深度集成：

- **对话消息关联执行链路**：Agent-Factory 对话消息与 Agent 执行链路完成关联，支持从对话维度追溯完整执行记录
- **知识证据链数据提取**：Context Loader 工具响应中新增对象类数据的提取能力，为 TraceAI 提供结构化的知识来源信息
- **默认业务域配置**：新增默认业务域配置支持，避免无业务域 ID 时接口报错，降低初始化配置门槛

**3. Decision Agent 与业务域及 ISF 架构解耦**

完成 Decision Agent 与业务域及 ISF 的架构解耦，不再强依赖业务域及 ISF 配置，支持在更灵活的多域、跨域部署场景中使用。

---

## 资源下载

### 1. GitHub 安装包和技术文档

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk

**AI Data Platform**
- GitHub Release: https://github.com/kweaver-ai/adp/tree/release/0.5.0

**Decision Agent**
- GitHub Release: https://github.com/kweaver-ai/decision-agent/tree/release/0.5.0

**TraceAI**
- GitHub: https://github.com/kweaver-ai/tracing-ai

---
