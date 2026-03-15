# KWeaver Core 0.4.0 版本发布通知


## 版本概述

### KWeaver Core 0.4.0 版本正式发布

**面向 AI Agent 的企业级智能数据与决策平台**

KWeaver Core 0.4.0 是 KWeaver 平台架构深度优化的重要里程碑版本。本次发布以"工程架构升级、效果显著提升、质量全面加固"为核心目标，在 BKN 引擎能力扩展、VEGA 数据虚拟化自研升级、全链路可观测性建立等方面取得重大突破，同时完成平台整体工程架构精简——内存占用降至 48G、微服务数目减少 12 个，在结构化与非结构化场景下的准确率和延迟指标均达到显著提升。

---

## KWeaver Core 0.4.0 版本核心亮点速览

**1. BKN 引擎全面扩展，知识网络构建能力跃升**

新增 Markdown 格式文档导入与管理支持，完成基于 Dataset 的概念索引体系重构，修正 Action 创建 Objects 的查询逻辑，并新增指定对象类构建索引能力。BKN 引擎正式从 ontology-manager 迁移至 bkn-backend，完成服务架构规范化，知识网络构建效率与准确性全面提升。

**2. VEGA 数据虚拟化自研引擎上线，多源元数据采集扩展**

全新 VEGA 引擎采用自研数据查询框架，实现对 MariaDB 的库表扫描与原生查询能力，脱离第三方依赖；同步扩展 MariaDB、Oracle、MySQL、OpenSearch 的元数据采集功能，支持跨数据源 JOIN 查询，大幅提升数据虚拟化层的灵活性与覆盖范围。

**3. TraceAI 全链路可观测性体系建立**

完成 TracingCollector 数据采集规范，覆盖 Agent 对话、Session、Run 各级别数据及工具调用链路的完整采集与 TracingID 关联；同步推进 BKN、执行工厂、Dataflow 的 Tracing 埋点落地，建立统一的跨组件可观测性数据标准。

**4. 架构工程持续精简，平台向轻量化演进**

全面去除 MongoDB 依赖，重构 eacp-single 服务，沙箱运行时引入会话级 Python 依赖管理；内存占用降至 48G，微服务数目减少 12 个，架构持续向轻量化演进。

---

### 【BKN 引擎】

BKN（Business Knowledge Network）是一种基于 Markdown 的声明式建模语言，用于定义业务知识网络中的对象、关系与行动。0.4.0 版本完成 BKN 引擎服务重命名（ontology-manager → bkn-backend），并在文档导入、索引构建、行动执行等核心能力上实现重要扩展。

**1. 支持 Markdown 格式文档导入与管理**

本版本新增对 Markdown（`.md`）格式文档的导入与管理能力，配合产品层需求完整实现文档管理入口。用户可通过统一界面完成 Markdown 文档的上传、解析与知识网络绑定，将非结构化业务文档纳入 BKN 的知识来源体系，扩展知识网络的文档覆盖范围。

**2. 基于 Dataset 的概念索引体系重构**

完成概念索引构建逻辑的全面重构，将索引生成过程与 Dataset 数据集紧密绑定，同时新增支持用户指定特定对象类构建索引的能力，提升索引构建的精细化程度与构建效率。

**3. Action 行动查询逻辑修正**

修正 Action 创建 Objects 过程中的查询逻辑问题，增强行动在无佐证材料场景下的执行能力，确保行动查询结果的准确性与完整性，为业务流程的自动化执行提供更可靠的数据基础。

**4. bkn-backend 服务规范化迁移**

完成 ontology-manager 服务向 bkn-backend 的正式重命名与迁移，保持现有 API 接口的完整向后兼容性，同步更新错误类型命名规范，并完成相关文档与错误包的重构，为后续 BKN 引擎能力扩展奠定架构基础。

---

### 【VEGA 数据虚拟化】

VEGA 数据虚拟化层负责多源数据的统一接入、查询与虚拟化管理。0.4.0 版本完成自研查询引擎上线，并大幅扩展数据源元数据采集覆盖范围。

**1. 新 VEGA 引擎：自研数据查询与 MariaDB 库表扫描**

全新 VEGA 引擎上线，采用自研数据查询框架替代原有第三方依赖，实现对 MariaDB 数据库的库表结构扫描与原生数据查询能力。新引擎在查询性能与灵活性上均有显著提升，为后续多数据库适配扩展打下基础。

**2. 多源元数据采集扩展**

新增对 MariaDB、Oracle、MySQL、OpenSearch 四类数据源的元数据采集功能，用户可通过统一的元数据管理界面获取上述数据源的表结构、字段信息及索引元数据，提升平台对企业多样化数据基础设施的覆盖能力。

**3. 跨数据源 JOIN 查询能力**

修正自定义数据查询框架，正式开放跨数据源的 JOIN 查询能力，支持在 VEGA 虚拟化层对多个数据源的数据进行关联查询，大幅拓展数据分析场景的灵活性。

---

### 【TraceAI】

TraceAI 是 KWeaver 平台的全链路可观测性组件，负责 Agent 执行过程的数据采集、链路追踪与行为审计。0.4.0 版本完成核心采集规范建立与跨组件埋点推进。

**1. TracingCollector 采集规范建立**

完成 TracingCollector 的完整数据采集规范设计，明确 Agent 相关链路的采集标准：

- **对话级采集**：完整记录 Agent 对话内容，关联 Session 与 Run 等级别的上下文数据
- **执行步骤采集**：记录 Agent 执行的每个 Step，包含工具调用 Step 及其响应内容
- **链路关联**：基于 TracingID 实现执行链路数据的跨组件关联调用

**2. 跨组件 Tracing 埋点推进**

同步推进 BKN 引擎、执行工厂、Dataflow 三大组件按照 TracingCollector 规范完善 Tracing 链路数据埋点，建立统一的可观测性数据标准，为后续 Agent 行为分析与问题排查提供完整数据基础。

---

### 【Dataflow】

Dataflow 是 KWeaver 平台的数据流处理引擎，负责文档解析、数据流转与多节点编排。0.4.0 版本完成两项重要架构改进。

**1. 全平台去除 MongoDB 依赖**

完成 Dataflow 层及全系统范围内对 MongoDB 的依赖清除，平台架构中不再存在任何 MongoDB 依赖组件。此次改造大幅简化了系统部署与运维复杂度，降低技术栈多样性带来的维护成本，进一步精简平台整体架构。

**2. 文档解析节点适配 MinerU 官方 API**

文档解析节点完成对 MinerU 官方 API 的适配对接，替代原有的本地解析方案。MinerU 官方 API 在 PDF、图文混排等复杂文档的解析精度上显著提升，为知识网络构建提供更高质量的文档解析结果。

---

### 【Context Loader】

Context Loader 是 KWeaver 平台智能体的上下文加载与管理组件，负责为 Agent 提供精准、高效的知识上下文检索与注入。0.4.0 版本在接口规范与压缩能力上完成重要优化。

**1. 接口形式优化，提升智能体准确率与效率**

对 Context Loader 的接口调用形式进行全面优化，精简接口调用链路，降低冗余数据传输，提升智能体在上下文检索阶段的响应速度与检索准确率。

**2. 新增 TOON 压缩能力**

新增 TOON 格式的上下文压缩能力，支持 HTTP 和 MCP 两种协议的压缩数据传输。通过上下文压缩，在保持知识完整性的同时，有效降低上下文注入的 Token 消耗，提升大规模知识网络场景下的智能体响应效率。

---

### 【执行工厂】

执行工厂（Execution Factory）是 KWeaver 平台的函数计算与沙箱执行调度核心，负责管理用户自定义函数与沙箱运行时的交互。0.4.0 版本完成函数依赖库管理能力的闭环建设。

**1. 函数依赖库安装支持**

完成函数对沙箱依赖库的完整支持，用户在执行工厂中定义的函数可通过统一机制声明并安装所需的 Python 依赖包。系统将自动完成依赖的解析、安装与版本管理，消除函数执行时的依赖缺失问题。

**2. 函数编辑器支持依赖包配置**

算子平台函数编辑器新增依赖包配置界面，用户可在编辑器内直接配置函数所需的依赖包列表，并实时查看安装状态，实现依赖管理与函数开发的一体化体验。

---

### 【ISF 信息安全编织】

ISF（Information Security Fabric，信息安全编织）负责 KWeaver 平台的权限管理、身份认证与访问控制体系。0.4.0 版本完成服务架构重构与授权管理能力扩展。

**1. 重构去除 eacp-single 服务**

通过架构重构完成 eacp-single 服务的移除，采用基于 Kubernetes 的多副本并发执行控制方案替代原有单一服务模式。新方案具备更强的水平扩展能力与更高的执行并发度，消除单点瓶颈，大幅提升平台在高并发场景下的稳定性与吞吐能力。

**2. 授权管理支持资源创建者条件配置**

授权管理新增资源创建者维度的访问控制配置能力：资源的创建者可以对其创建的相关资源进行操作管理，无需额外的管理员授权。该能力进一步细化了平台的权限粒度，使权限策略更贴合实际业务场景中的所有权管理需求。

---

### 【Sandbox 沙箱运行时】

Sandbox 沙箱运行时为 KWeaver 平台提供安全隔离的代码执行环境，支持 Python 函数执行、文件工作空间管理与多运行时调度。0.4.0 版本（对应 Sandbox v0.3.0）新增会话级 Python 依赖管理能力。

**1. 会话级 Python 依赖管理**

新增完整的会话级 Python 依赖配置与管理能力：

- **依赖声明与安装**：支持在会话创建时声明所需依赖包，并通过后台任务触发初始依赖同步安装
- **手动安装**：支持按需手动触发依赖安装，灵活响应运行时环境变化
- **状态追踪**：会话响应中包含已安装依赖明细及错误信息，前端会话管理界面同步展示依赖安装操作与状态

**2. 运行时依赖同步与隔离**

确保运行时执行器与控制平面的依赖状态完整同步，自动处理依赖冲突与环境隔离，增强沙箱在复杂部署环境下的兼容性与稳定性。

**3. 数据库与版本升级支持**

新增数据库自动升级能力，服务启动时自动完成数据库结构的版本对齐，支持从历史版本平滑升级，降低运维成本与升级风险。

---

### 【Decision Agent】

Decision Agent（决策智能体）是 KWeaver 平台面向业务决策场景的核心智能体组件，提供多步骤推理、工具调用与对话管理能力。0.4.0 版本完成权限处理增强与 API 路由修复。

**1. agent-executor 日志目录权限处理增强**

增强 agent-executor 在各类部署环境下的运行稳定性，完善日志目录的权限处理与错误恢复机制，确保智能体执行日志的完整记录。

**2. 修复 agent router v1 注册缺失**

修复历史版本 API 端点不可用的问题，恢复完整的 API 路由功能，保障存量业务系统的正常集成与访问。

---

## 版本发布

### 1. 产品安装包和技术文档

**AI Data Platform (ADP):**
- **GitHub Release**: [https://github.com/kweaver-ai/adp/tree/release/0.4.0](https://github.com/kweaver-ai/adp/tree/release/0.4.0)
- **技术文档**: [https://github.com/kweaver-ai/adp/blob/main/README.md](https://github.com/kweaver-ai/adp/blob/main/README.md)

**Decision Agent:**
- **GitHub Release**: [https://github.com/kweaver-ai/decision-agent/tree/feature/20260312](https://github.com/kweaver-ai/decision-agent/tree/feature/20260312)
- **Changelog**: [https://github.com/kweaver-ai/decision-agent/blob/feature/20260312/changelog.zh.md](https://github.com/kweaver-ai/decision-agent/blob/feature/20260312/changelog.zh.md)

**Sandbox:**
- **GitHub Release**: [https://github.com/kweaver-ai/sandbox/tree/feature/20260312](https://github.com/kweaver-ai/sandbox/tree/feature/20260312)
- **Changelog**: [https://github.com/kweaver-ai/sandbox/blob/feature/20260312/CHANGELOG_ZH.md](https://github.com/kweaver-ai/sandbox/blob/feature/20260312/CHANGELOG_ZH.md)

### 2. 产品发布资料

- **版本发布日期**：2026-03-14
- **版本目标文档**：KWeaver Core 0.4.0 关键计划

---

**KWeaver - 面向 AI Agent 的智能数据平台与决策引擎**
