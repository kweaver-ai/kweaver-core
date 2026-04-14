# KWeaver Core 0.6.0 版本发布通知

---

## 版本概述

### KWeaver Core 0.6.0 版本正式发布

**企业决策智能生态，Skill 能力全链路贯通**

KWeaver Core 0.6.0 以 **Skill 全链路打通**为核心主题，将Skill能力从业务知识网络建模、Context Loader 语义召回、执行工厂统一执行，到 Decision Agent 智能体感知与调用，形成端到端完整闭环。本版本同步完成 VEGA 对 AnyShare 企业知识库的无缝接入，将主流企业非结构化数据源纳入业务知识网络统一管理体系；沙箱运行时新增压缩包上传自动解压与 Shell 脚本执行能力，大幅提升智能体代码执行的灵活性；整体部署内存压缩至 24G，显著降低企业落地门槛。与此同时，`kweaver-eval` 首版 Acceptance 模块正式建成，覆盖 6 大核心模块 104 个用例，以"确定性断言 + Agent Judge 语义评估"双维度体系对平台全栈进行独立验收，标志着 KWeaver Core 的工程质量治理能力进入系统化阶段。

---

## KWeaver Core 0.6.0 版本核心亮点速览

**1. Skill 全链路打通，技能能力贯穿平台各层**

0.6.0 版本完成 Skill 在 KWeaver 核心平台的全链路打通：BKN 支持在知识网络中建模 Skill 对象类；Context Loader 新增 `find_skills` 工具，支持在知识网络边界内语义召回 Skill 候选；执行工厂新增 Skill 执行接口并支持双写 Dataset；Decision Agent 完成对执行工厂 Skill 的读取与加载；KWeaver SDK 同步对接执行工厂 Skill 管理模块。技能不再是孤立的功能单元，而是与业务知识网络深度融合的可治理、可召回、可执行的智能体能力。

**2. AnyShare 企业知识库接入，非结构化数据源统一纳管**

VEGA 完成对 AnyShare 企业知识库的 Catalog 对接，支持通过 Discover 命令扫描发现 AnyShare 知识库（单页上限 1000 条），并通过 BKN 对象类建模将 AnyShare 文档数据纳入业务知识网络统一管理。支持应用账号、永久 Token 及 SSO 一键登录三种认证方式，实现企业级权限穿透查询。Resource 层同步新增实时流式增量构建与自定义视图能力，数据虚拟化构建能力持续完善。

**3. Sandbox 沙箱运行时升级，智能体执行边界大幅拓展**

沙箱运行时 0.3.2 版本新增压缩包上传与自动解压能力，支持将 ZIP 包整体上传至 Session Workspace 并自动解压到指定路径，内置路径安全校验防止路径穿越；新增 Shell 脚本执行支持，提供 `working_directory` 工作目录参数，覆盖链式命令、相对路径执行等复杂场景。两项能力联合，使智能体在代码执行场景下具备完整的 Skill 包部署与脚本执行能力。

**4. KWeaver SDK Skill 管理与 Dataflow CLI 全面上线，开发者体验大幅升级**

KWeaver SDK 本版完成对执行工厂 Skill 管理模块的全面对接，新增 `kweaver skill` 命令族，覆盖 Skill 注册、安装、查找、读取、状态查询及内容下载全生命周期。新增 `kweaver dataflow` 命令族，支持直接上传本地文件或通过远程 URL 触发 Dataflow 非结构化数据处理流程，SDK 现可端到端覆盖「文件上传 → 文档解析 → 知识网络构建 → 智能体问答」全链路。新增 `kweaver explore` 命令，在本地启动浏览器 SPA，提供平台状态概览、BKN 浏览器、Decision Agent 流式对话（含 Trace）与 VEGA 数据预览四个视图；同步引入多账号 Profile 管理，支持 `--alias` 命名、`--user` 单命令级凭证覆盖与过期客户端自动恢复，显著提升 CI/容器环境下的开发者体验。

**5. 部署内存降至 24G，企业落地门槛大幅降低**

0.6.0 版本完成整体部署优化，ADP + ISF + Core 全量部署内存降低至 24G 以内，相比此前版本显著降低。配合 `kweaver-eval` 首版 Acceptance 模块建成，覆盖 6 大核心模块 104 个用例，采用"确定性断言 + Agent Judge"双维度评测体系，为平台质量持续治理提供工程基础。

---

## 详细功能说明

### 【VEGA 数据虚拟化】

0.6.0 版本在数据源接入与 Resource 构建两个方向均完成重要能力扩展，AnyShare 企业知识库接入是本版本 VEGA 最核心的能力突破。

**1. AnyShare 知识库接入（Catalog 对接）**

VEGA Catalog 完成对 AnyShare 企业知识库的接入支持，通过 Discover 命令自动扫描当前账号下可访问的知识库列表（单页检索上限 1000 条），并获取每个知识库的完整 Resource 元数据（名称、ID、类型、创建者、修改者、DNS 地址、字段定义等）。

支持三种认证方式以适配不同企业部署场景：
- **应用账号**：提供 AppID + Secret，适合集成部署
- **永久 Token**：适合高权限运维场景
- **SSO 一键登录**：与 KWeaver 账号体系打通，当前登录者权限透传至 AnyShare，实现用户级数据权限隔离

当前版本优先支持 AnyShare 知识库类型接入，文档库等其他类型将在后续迭代中逐步覆盖。

**2. Resource 实时流式增量构建**

Resource 新增实时流式增量构建模式，在全量构建的基础上支持基于增量字段的实时数据同步。相比全量构建，增量构建仅处理自上次构建以来新增或变更的数据，大幅降低数据更新成本，满足对数据时效性要求较高的场景。

**3. Resource 自定义视图**

支持对 Resource 定义自定义视图，用户可灵活配置字段映射与向量索引结构，满足不同业务场景的检索需求，进一步提升 VEGA 在数据虚拟化层的灵活性。

**4. 大数据量查询稳定性优化**

针对 Catalog 和 Resource 列表接口在大数据量场景下的查询性能问题进行专项优化，提升大规模知识库接入场景下的系统稳定性。

---

### 【BKN 引擎】

0.6.0 版本完成 BKN 与 VEGA Resource 的对接，使业务知识网络中的对象类可直接绑定到 VEGA 数据资源，打通了业务语义建模与数据虚拟化层的连接通路。

**1. BKN 对接 VEGA Resource**

BKN 对象类现可绑定至 VEGA Resource 数据源（支持 AnyShare 知识库、MySQL、Dataset 等类型），实现通过业务知识网络统一查询和召回来自多个异构数据源的数据。绑定完成后，通过 BKN 的 `ontology_query` 接口即可检索对应 Resource 中的真实数据，语义建模与数据访问在同一知识网络中完成闭合。

支持两种对象创建模式：
- **严格模式（strict_mode=true）**：创建对象类时校验关联 Resource 的存在性，不存在则报错拦截
- **非严格模式（strict_mode=false）**：忽略依赖资源存在性检查，直接创建对象类，适合分步建模场景

**2. 稳定性修复**

修复对象类、行动类在无 Branch 状态下更新时报错的问题；修复对象类绑定 Resource 后列表排序异常的问题，提升业务知识网络编辑操作的稳定性。

---

### 【Context Loader】

0.6.0 版本新增 `find_skills` 工具，将 Skill 候选发现能力引入 Context Loader 工具体系，完成技能在语义召回层的闭环。

**1. find_skills：知识网络边界内的 Skill 候选发现**

新增 `find_skills` 工具，支持在指定 `kn_id` 和业务上下文边界内发现 Skill 候选，补全 Context Loader `find_*` 语义工具族的能力覆盖。

核心特性：
- **双召回模式**：支持对象类级召回（在指定对象类下查找 Skill）和实例级召回（在特定实例关联的 Skill 中查找），网络级召回后续开放
- **最小化元数据返回**：返回 `skill_id`、`name`、`description` 三项核心信息，降低 Token 消耗
- **基础契约校验**：入口自动校验知识网络中 `skills` ObjectType 是否存在且至少包含 `skill_id`、`name` 数据属性，不满足时直接返回明确错误，避免无效召回

> 注意：`find_skills` 属于候选资源发现工具，不替代 `kn_search` / `query_object_instance`；`object_type_id` 为必填参数；网络中须已完成 Skill 承接和绑定，召回结果才能稳定可用。

**2. 统一响应格式支持**

新增 `response_format` 参数，支持 `json` 和 `toon` 两种格式：
- HTTP 接口默认返回 `json`，向后兼容
- MCP Tool 默认返回 `toon` 压缩格式，降低 Token 消耗
- 错误响应统一保持 JSON 格式

同步统一了 Context Loader 各工具的 `x-account-id`、`x-account-type` 参数口径，简化调用方接入配置。

---

### 【执行工厂】

0.6.0 版本新增 Skill 执行接口，将技能执行能力纳入执行工厂统一管理，并支持执行结果双写 Dataset。

**1. Skill 执行接口与 Dataset 双写**

执行工厂新增专用 Skill 执行接口，支持通过统一接口调用已注册的 Skill，并将执行结果同步写入 Dataset，实现技能产出数据的持久化沉淀。Dataset 数据可供后续 BKN 对象类绑定查询，形成"执行→存储→召回"的完整数据闭环。

**2. 稳定性修复**

修复组合算子创建失败的问题；修复流式转发在缺少 ResponseWriter 时 nil 指针解引用导致 panic 的问题；优化服务启动逻辑，避免索引初始化阻塞启动，改为后台重试机制，提升服务可用性。

---

### 【Decision Agent】

0.6.0 版本完成 Decision Agent 对执行工厂 Skill 能力的全面接入，智能体具备了从知识网络中感知、读取并调用 Skill 的完整能力。

**1. Agent 内置 Skill 读取与加载能力**

`agent-factory` 新增 Skill 类型支持，在 Agent 创建、详情查看、更新处理器及运行服务中全面支持 Skill 配置，智能体可在运行时从执行工厂读取并加载可用 Skill，实现对 Skill 的原生感知与调用。

相关数据库迁移已完成（v0.6.0 Skill 相关表及 `agent-memory` 历史表，覆盖 DM8 和 MariaDB）。

**2. TraceAI Evidence 链路补充**

`agent-executor` 新增 TraceAI Evidence 请求头支持，引入 `enable_traceai_evidence` 功能开关（在 `FeaturesConfig` 中配置），开启后在 API 工具代理请求中自动注入 `X-TraceAi-Enable-Evidence` 请求头，打通 Decision Agent 以外其他调用链的 Evidence 数据采集通路，实现全链路证据提取的统一覆盖。

**3. 发布请求校验优化**

重构 `agent-factory` 发布请求校验逻辑，采用构造函数语义对请求字段进行校验和清洗，校验失败时返回明确的 400 错误信息而非 500，提升接口错误的可诊断性。

---

### 【沙箱运行时】

0.6.0 版本对应 Sandbox 0.3.2，新增压缩包上传与 Shell 执行两项能力，显著提升智能体在代码执行场景下的灵活性与适用范围。

**1. 压缩包上传与自动解压**

新增 Session Workspace 压缩包上传能力，支持将 ZIP 格式压缩包上传至指定 Workspace 路径并自动解压：

- **覆盖控制**：支持配置冲突时的覆盖行为，上传响应中包含覆盖统计和解压结果元数据
- **路径安全**：内置路径安全校验，拒绝非法条目与路径穿越内容（Path Traversal 防护）
- **Skill 包部署**：配合 Skill 执行能力，支持将整个 Skill 依赖包一次性上传部署到 Sandbox 环境

**2. Shell 脚本执行支持**

新增 `language=shell` 执行模式，在 `execute` 与 `execute-sync` 接口中支持 Shell 脚本直接执行：

- **工作目录控制**：支持通过可选 `working_directory` 参数指定脚本执行的工作目录
- **命令规范化**：对误写 `bash/sh` 前缀命令进行自动规范化处理
- **场景覆盖**：支持链式命令、相对路径执行等复杂 Shell 脚本场景，端到端测试覆盖完整

**3. Helm Chart 镜像配置优化**

将 Sandbox Helm Chart 镜像配置统一规范至顶层 `image` values 结构，支持离线打包工具从 Chart values 中完整提取所有镜像，简化私有化部署的离线安装流程。

---

### 【KWeaver SDK】

0.6.0 版本完成 KWeaver SDK 与执行工厂 Skill 管理模块的对接，并新增 `kweaver explore` 交互式探索命令与多账号管理能力，进一步提升开发者与智能体在平台上的操作体验。

**1. Skill 管理命令全面上线**

SDK CLI 新增 `kweaver skill` 命令族，覆盖 Skill 从注册到执行的完整生命周期：

- `skill list`：列出当前可用的 Skill
- `skill market`：浏览 Skill 市场
- `skill get`：获取指定 Skill 的配置详情、输入输出参数定义
- `skill register`：将 Skill 注册到执行工厂，纳入统一执行能力管理体系
- `skill install`：触发 Skill 依赖安装与运行环境初始化
- `skill status`：查询 Skill 安装/运行状态
- `skill content` / `read-file`：读取 Skill 内容与附属文件
- `skill download`：下载 Skill 包

**2. kweaver explore：本地可视化平台探索**

新增 `kweaver explore` 命令，在本地启动一个浏览器 SPA，提供四个交互式探索视图：平台状态聚合概览、BKN 知识网络浏览器（对象类/关系类/实例/子图）、Decision Agent 流式对话（含实时执行 Trace 展示）、VEGA Catalog 与数据预览。开发者无需打开完整平台 UI，即可快速验证平台数据与智能体效果。

**3. Dataflow CLI 上线，非结构化数据构建 KN 全链路打通**

新增 `kweaver dataflow` 命令族，支持通过 SDK 直接操作 Dataflow 文档处理流程：

- `dataflow list`：列举当前所有 Dataflow DAG
- `dataflow run <dagId> --file <path>`：上传本地文件（PDF、Word 等）直接触发非结构化数据处理流程
- `dataflow run <dagId> --url <url> --name <filename>`：通过远程 URL 触发，无需本地上传
- `dataflow runs <dagId>`：查看指定 DAG 的运行记录，支持 `--since` 按时间过滤
- `dataflow logs <dagId> <instanceId>`：查看逐步骤执行日志，`--detail` 可展开每步 input/output 完整载荷

结合 Dataflow 文档解析节点、执行工厂 Skill 双写 Dataset、BKN 对象类绑定与 Context Loader 语义召回，SDK 现可端到端覆盖「文件上传 → 文档解析 → 知识网络构建 → 智能体问答」全链路，并可通过 OpenClaw 完整演示效果。

**4. 多账号管理与认证增强**

新增多账号 Profile 管理支持，开发者可在同一机器上管理多个 KWeaver 实例的登录态：

- `--alias` 参数支持为登录账号命名，`auth use` 在多账号间快速切换
- 全局 `--user` 标志支持单条命令级别的凭证覆盖，无需切换全局账号
- 支持通过 `--port` 和 `--redirect-uri` 自定义 OAuth2 回调地址，适配复杂网络环境
- 过期客户端自动恢复（stale client auto-recovery），减少重新登录频率

---

### 【TraceAI】

0.6.0 对应 TraceAI 0.2.2，完成 Helm Chart 镜像版本管理优化，确保打包出的 Chart 可以继承发布流程解析出的版本号，镜像标签与实际发布镜像保持强一致。

---

### 【kweaver-eval】

0.6.0 版本完成 `kweaver-eval` 模块初版建设，对 KWeaver Core 全栈进行独立验收，覆盖 **Agent、BKN、VEGA、数据源（DS）、数据视图（Dataview）、Context Loader** 6 大核心模块，共 **104 个用例**，当前 **79 个用例通过（76%）**。

**1. 用例覆盖概览**

| 模块 | 用例数 | 通过 | 已知缺陷 |
|------|--------|------|----------|
| Agent（Decision Agent） | 33 | 33 | 0 |
| BKN 引擎 | 26 | 23 | 3 |
| VEGA + DS + Dataview | 27 | 19 | 6 |
| Context Loader | 3 | 3 | 0 |
| Dataflow | 14 | 待统计 | — |
| Token 刷新 | 1 | 1 | 0 |

Agent 模块覆盖了 CRUD 生命周期、单轮/多轮/流式对话、对话健壮性（并发会话、特殊字符输入、长消息、上下文跨轮保持）及错误路径共 33 个用例，全部通过。Dataflow 模块 14 个测试用例已随 `kweaver dataflow` CLI 于 0.6.0 发布当日正式激活。

**2. 双维度评测体系**

每个测试用例均产出两个评分维度：

- **确定性断言**：验证退出码、JSON 结构、字段值，覆盖所有用例
- **Agent Judge 语义评估**：通过 Claude API 对输出结果进行语义正确性评分，支持 CRITICAL / HIGH / MEDIUM / LOW 四级严重度定级，按用例选启

**3. 跨运行 Issue 追踪与上游缺陷反馈**

内置跨运行 Issue 追踪机制（`feedback.json` 持久化），自动识别连续出现的缺陷并升级为需人工跟进状态。已通过验收测试发现并跟踪反馈 **20+ 个上游缺陷**（含 adp#427、adp#428、adp#442、adp#445、adp#447、adp#448 等），形成 kweaver-eval 与上游仓库之间的工程质量反馈闭环。

---

## 版本发布

### 1. GitHub 安装包和技术文档

**KWeaver Core**
- GitHub Release: https://github.com/kweaver-ai/kweaver-core/tree/release/0.6.0

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk

**kweaver-eval**
- GitHub: https://github.com/kweaver-ai/kweaver-eval

---
