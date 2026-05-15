# KWeaver Core 0.8.0 版本发布通知

---

## 版本概述

### KWeaver Core 0.8.0 版本正式发布

**工具链开箱即用，全链路可观测性落地，多语言运行时就绪，单机部署降低准入门槛**

KWeaver Core 0.8.0 以 **开发者体验提升** 为核心主线，在工具链交付、可观测性基础设施与沙箱运行时三个方向完成重要突破。Context Loader 工具集正式内置为平台默认工具箱，平台部署后自动完成工具箱注册，无需在执行工厂中手动导入；`search_schema` 新增 `concept_groups` 参数，支持按 BKN 概念分组精确限定语义召回范围，解决多业务场景共用同一知识网络时的召回边界问题。BKN 引擎将 ActionType 语义升级为"行动意图（Action Intent）"声明，并将影响对象类扩展为支持多目标的"影响契约（Impact Contract）"，进一步精化智能体对业务动作副作用的建模能力。可观测性方面，`bkn-backend`、`ontology-query`、`vega-backend` 三大服务完成 OpenTelemetry 链路追踪全量接入，跨服务调用链可通过 OpenSearch 统一关联查看，为 TraceAI 证据链能力的持续建设奠定基础。沙箱运行时升级至 v0.4.0，新增多语言复合模板环境并正式补齐 Go 运行支持。VEGA 同步完成生产级限流控制能力上线，为数据查询并发场景提供双层保护；同时对 BuildTask、DiscoverTask、Dataset、DiscoverSchedule 等核心资源域完成 API 体系全面重构，统一设计风格，消除历史遗留的路径不一致问题。部署方面，本版本新增基于 K3s（Linux）和 Docker + Kind（macOS）的单机部署方案，并重新设计部署流程，提供预检脚本与安装后上架向导，显著降低首次部署与运维的试错成本。

---

## KWeaver Core 0.8.0 版本核心亮点速览

**1. Context Loader 工具集开箱即用，concept_group 语义精确召回**

Context Loader 工具集正式内置为平台默认工具箱，随平台启动自动完成注册，无需在执行工厂中手动导入工具箱。`search_schema` 工具新增 `concept_groups` 参数，支持将 Schema 召回范围限定在指定 BKN 概念分组内，解决多业务场景共用同一知识网络时的语义边界混淆问题。

**2. 全链路可观测性基础就绪：三大核心服务接入 OpenTelemetry**

`bkn-backend`、`ontology-query`、`vega-backend` 完成 OpenTelemetry 链路追踪全量改造，HTTP/OAuth/OpenSearch 等访问层统一切换为 OTel 适配客户端。一次数据查询请求涉及的多个微服务 span 现已可通过 OpenSearch 关联查看，为生产环境问题定位建立基础。基于 TraceAI 的智能体侧调用链可视化查询能力将在后续版本持续完善。

**3. BKN 行动语义深化：行动意图与影响契约**

ActionType 字段语义升级为"行动意图（Action Intent）"，明确其为业务动作意图的声明，而非执行指令。影响对象类扩展为"影响契约（Impact Contract）"，支持同一行动声明多个影响目标及操作类型，使智能体在决策前能够完整理解行动的业务副作用。

**4. Sandbox 多语言运行时升级，Go 语言正式支持**

沙箱运行时升级至 v0.4.0，新增多语言复合模板环境，Go 语言运行支持正式交付。Python 依赖安装新增超时配置，避免长耗时安装任务阻塞会话。修复 s3fs 启动脚本沉默失败导致工作目录异常的问题，提升多种部署场景下的执行稳定性。

**5. VEGA 生产级限流控制上线，双层保护数据源稳定性**

VEGA Resource 数据查询新增完整限流控制机制，支持全局并发限制与 Catalog 级别并发限制双层独立生效，同时返回标准 429 Too Many Requests 错误码，从根本上解决并发查询场景下数据源过载与资源抢占问题。

**6. 单机部署方案就绪，一键完成本地安装**

KWeaver Core 新增基于 K3s（Linux）和 Docker + Kind（macOS）的低资源单机部署方案，一键完成本地或单机服务器安装，适用于开发测试、评估验证与资源受限的生产环境。同步提供部署前预检脚本与安装后上架向导，降低首次部署与重复运维的试错成本。

---

## 详细功能说明

### 【Context Loader】

0.8.0 版本完成 Context Loader 工具链的交付闭环，工具集内置、按需分组召回与开箱即用体验三项能力同步落地。

**1. Context Loader 工具集内置为平台默认工具箱**

Context Loader 工具集正式成为平台内置默认工具箱，随平台部署自动注册，无需用户手动安装或配置：

- **自动注册**：平台启动时自动检查并同步内置工具依赖包，确保工具集始终与当前版本契约对齐
- **免手动导入**：无需在执行工厂中手动导入工具箱，消除 fresh install 后 kn_* 工具不可用的问题；Agent 侧仍需按需配置工具参数
- **版本化交付**：工具箱描述中内置契约版本（当前为 0.8.0），平台升级时自动触发工具集更新，避免版本漂移

**2. search_schema 支持 concept_group 精确召回**

`search_schema` 工具新增 `search_scope.concept_groups` 参数，允许将 Schema 召回范围限定在指定 BKN 概念分组内：

- **按分组限定**：指定 `concept_groups` 后，对象类、关系类、行动类与指标类的召回均限制在该分组范围内，不同业务场景可在同一知识网络中精确隔离语义边界
- **全类型覆盖**：`concept_groups` 与现有 `include_object_types`、`include_relation_types`、`include_action_types`、`include_metric_types` 开关正交组合，灵活控制召回精度
- **向后兼容**：未传 `concept_groups` 时，保持原有全网范围召回行为，存量 Agent 无需改动
- **说明**：指标类（Metric）的 concept_group 过滤依赖 BKN metrics 侧支持，当前版本指标召回会携带概念分组范围参数，实际过滤效果以 BKN 侧行为为准

**3. CLI 命令帮助体验优化**

Context Loader CLI 帮助体系参照 GitHub / Docker 命令风格进行重构：

- **分场景说明**：11 个核心工具场景均配有独立的用法说明、参数说明与推荐工作流提示
- **工具分组**：schema、query、instance、action、concept_group 等命令按功能族分组展示，降低认知负担
- **示例驱动**：每个工具提供完整的 `example` 示例，包含常见参数组合，可直接复制执行

---

### 【BKN 引擎】

0.8.0 版本在行动语义建模、指标能力与 SDK 三个方向完成重要升级，并修复若干影响可靠性的边界问题。

**1. ActionType 语义升级为行动意图（Action Intent）**

行动类型（ActionType）字段语义升级为行动意图，明确其定位为业务动作意图的声明，而非对执行过程的约束：

- **语义精化**：行动意图仅声明"这个行动打算做什么"（新增/更新/删除等），工具的实际执行逻辑由绑定的 Action Source 决定，两者职责分离
- **值域保留**：当前支持值域与原行动类型一致，后续可按业务需要扩展，存量配置无需迁移

**2. 影响对象类升级为影响契约（Impact Contract）**

影响对象类（Impact Object Type）升级为影响契约（Impact Contract），支持在同一行动上声明多个影响目标：

- **多目标声明**：单个行动可声明对多个对象类的影响，每条契约记录影响的对象类、操作类型及影响描述，完整表达行动的业务副作用
- **辅助决策**：Agent 在执行行动前可通过影响契约评估该行动可能涉及的对象范围，在风险敏感场景下提供可读性更强的决策依据
- **非硬约束**：影响契约为声明性信息，不约束工具的实际执行行为，降低建模门槛

**3. 指标类型支持 concept_group 分组检索**

`SearchMetrics` 接口新增 concept_group 过滤支持，与 `SearchObjectTypes` 等接口保持一致：

- 支持在指定概念分组范围内检索指标类型，避免多场景知识网络中的指标语义混淆
- 当所有指定概念分组均不存在时，返回 404 错误而非 500，提升异常可诊断性

**4. BKN 范式指标导入/导出**

BKN 范式完成指标类型（Metric Definition）的导入/导出支持：

- **tar 导出**：BKN tar 导出包现包含 MetricDefinition，支持指标定义随知识网络整体迁移与分发
- **tar 导入**：导入时自动解析并还原 MetricDefinition，支持批量导入与集成测试覆盖
- **示例更新**：`k8s-network/metrics` 等标准示例同步新增指标定义示例

**5. BKN 三大服务接入 OpenTelemetry Trace 埋点**

`bkn-backend`、`ontology-query`、`vega-backend` 完成 OpenTelemetry 链路追踪全量改造，为 TraceAI 证据链能力持续建设奠定数据基础：

- driveradapters / drivenadapters / logics 全层接入 OTel span 追踪与错误日志埋点
- HTTP/OAuth/Hydra/OpenSearch/vega-backend 等外部访问层统一切换为 OTel 适配客户端
- 为各业务流程成功分支补齐 span 状态设置，为失败分支补齐结构化错误日志
- 一次数据查询请求跨越的多个微服务 span（如 vega-backend → bkn-backend → ontology-query 的 24 个 span）现已可在 OpenSearch 中统一关联查看

**6. BKN SDK 升级**

kweaver-sdk 完成 BKN 指标能力的全面接入，并对知识网络构建流程完成若干重要改进：

- **bkn metric CLI**：新增 `bkn metric list` / `bkn metric query` / `bkn metric dry-run` 等命令，支持程序化查询 BKN 指标数据；TypeScript 与 Python 均已完成 `BknMetricsResource`、`MetricQueryResource` API 客户端封装
- **create-from-ds 迁移至 VEGA Catalog**：`bkn create-from-ds` 与 `create-from-csv` 的表结构扫描由旧 data-connection 接口迁移至 VEGA Catalog API，现需传入 VEGA catalog id；CLI 提示与帮助文档同步更新
- **事务性创建**：`create-from-ds` / `create-from-csv` 改为批量 POST ObjectType，后端在事务中执行，all-or-nothing；任意步骤失败自动回滚已创建的知识网络，消除残留孤儿数据问题；CSV import 阶段前增加客户端预校验，提前暴露命名不合规问题
- **PK 自动检测改进**：新增 `--pk-map` 参数支持手动指定主键列；auto-detect 歧义时立即 fail-fast，不再静默取第一列（旧逻辑可能导致低基数列被误作主键，造成大量数据丢失）；优先采用数据源 Schema 中声明的 PRIMARY KEY 约束，减少依赖采样启发的不确定性

**7. 问题修复**

- 修复间接关系类（Indirect Relation Type）在 `BackingDataSource` 为空时的空指针异常，避免特定配置下的服务崩溃
- `action-type execute` 接口在 input 参数缺失时由静默丢弃改为返回明确错误，提升工具调用链路中鉴权参数缺失场景的可诊断性
- `ontology-query` 修复指标趋势查询中 resource 日历分桶返回日期字符串时的解析失败问题

---

### 【TraceAI】

0.8.0 版本完成 TraceAI 证据链能力的基础建设，三大核心服务完成调用链埋点接入。

**1. 三大核心服务调用链埋点接入**

BKN 引擎、VEGA、Ontology Query 三大服务完成基于 OTEL 的 Trace 埋点和上报，TraceAI 的调用链数据采集基础已就绪：

- 通过配置开启 `traceenable` 即可启用相关服务的链路追踪
- 各微服务的 service name、version 等 OTel 标准属性已在 Helm ConfigMap 中统一配置
- 已采集的 Trace 数据可通过 OpenSearch 进行原始数据查询和链路关联分析

**2. 当前阶段说明**

TraceAI 证据链能力处于持续建设中：

- **已完成**：核心服务调用链数据采集与上报，Trace 数据可在 OpenSearch 中查看
- **进行中**：基于 TraceAI agent-observability 服务的智能体侧调用链可视化查询，将在后续版本持续完善

---

### 【VEGA 数据虚拟化】

0.8.0 版本在三个方向完成重要交付：生产级限流保障能力上线、核心资源域 API 体系全面重构（BuildTask/DiscoverTask/Dataset/DiscoverSchedule 统一设计风格并消除历史遗留问题），以及为 Catalog/Resource 引入可扩展的业务字段机制。

**1. Resource 数据查询限流控制**

VEGA Resource 数据查询新增完整的并发限流控制机制：

- **全局并发限制**：限制 VEGA 服务整体最大并发查询数，防止服务自身因并发过高而过载；超出全局限制时返回 `429 Too Many Requests`（`ErrGlobalLimitExceeded`）
- **Catalog 级并发限制**：对每个数据源（Catalog）独立设置最大并发查询数，保护下游数据库/服务不被单一 Catalog 的突发请求压垮；超出时返回 `ErrCatalogLimitExceeded`
- **双层独立生效**：全局限制与 Catalog 限制并行生效，两者同时满足方可获得查询许可
- **动态配置**：全局并发数与 Catalog 并发数均支持通过配置文件动态调整，无需重启服务

**2. Catalog/Resource 扩展字段支持**

Catalog 与 Resource 新增实体级扩展 KV（`t_entity_extension`），支持在标准字段之外附加业务自定义属性：

- 扩展字段支持创建、查询与更新，接口与校验完整覆盖
- 数据库迁移脚本已对齐 MariaDB 与 DM8 双数据库

**3. API 体系重构**

对 BuildTask、DiscoverTask、Dataset、DiscoverSchedule 等核心资源域完成全面 API 重构，统一设计风格、消除历史遗留的路径不一致与双主键冗余问题：

- **BuildTask 顶层资源化**：从嵌套路径 `/resources/buildtask/...` 迁移为顶层 `/build-tasks`，寻址改为单主键；状态控制由 `PUT .../status` 改为显式动作 `POST .../start` / `POST .../stop`，补全完整的增删查端点
- **DiscoverTask API 收敛**：端点收敛为只读 + 清理（list / get / delete），新增整体事务批量删除，字段命名规范化（`scheduled_id` → `schedule_id`）
- **Dataset & DiscoverSchedule 收敛**：Dataset 文档接口统一收敛至 `/resources/{id}/data`；DiscoverSchedule 从嵌套路径提升为顶层 `/discover-schedules`，去除双主键 path；旧 `SQL` 命名统一重命名为 `Raw`，消除语义误导
- **OpenAPI 全量补齐**：Catalog、Resource、Dataset、BuildTask、DiscoverTask、Query/Discover/TestConnection 等各资源域均已完成 OpenAPI 3.1 独立文档覆盖；修复多处接口文档与实际行为不一致问题

**4. 问题修复**

- 修复 Logic View 相关 bug 并补充 SQL 校验
- 修复指标趋势查询中 resource 日历分桶返回日期字符串时的解析失败问题
- 移除冗余的统一查询接口 `/api/vega-backend/v1/query/execute`，收敛 API 入口

---

### 【执行工厂】

0.8.0 版本完成 Skill 管理侧内容查看能力，建立工具命名规范的自动化前置校验体系，并通过 kweaver-sdk 补齐 Skill 全生命周期的 CLI 与 SDK 编程接口。

**1. Skill 管理态内容查看**

在已有 Skill 生命周期管理（版本控制、整包更新、历史回滚）基础上，新增管理态 Skill 内容读取能力：

- **SKILL.md 内容读取**：通过管理态接口获取 Skill 的完整 SKILL.md 描述内容及内部文件结构，无需下载完整包即可审阅 Skill 实现说明
- **响应模式切换**：支持 `url`（返回 OSS 预签名下载链接）与 `content`（直接返回文件内容）两种响应模式，按需选择
- **OSS 自动同步**：Skill 注册或更新时自动将 SKILL.md 同步到 OSS，确保管理态读取始终获取最新内容
- **发布时重名校验**：名称唯一性校验调整为仅在发布时触发，编辑态不再拦截，降低 Skill 迭代流程中的操作摩擦

**2. 工具名称规范化校验**

为提升 Agent 与大语言模型（尤其是 DeepSeek 系列）的工具调用成功率，新增工具命名规范前置校验：

- **注册阶段校验**：在工具注册、导入、编辑与发布阶段前置检查工具名称是否符合 DeepSeek 工具命名规范（如中文字符等特殊字符的限制）
- **告警提示**：发现不合规名称时输出明确告警，引导用户在运行前修正，避免运行时 tool calling 静默失败
- **覆盖范围**：校验覆盖 API、Agent、MCP 三类工具注册路径

**3. 内置工具箱自动注册与升级**

配合 Context Loader 工具集内置化，执行工厂完成内置工具箱的 ADP 格式自动注册与升级流程，确保平台升级后内置工具集版本与运行时保持一致。

**4. Skill SDK 与 CLI 全生命周期管理**

kweaver-sdk 完成 Skill 编辑与版本历史管理能力的 CLI 和 SDK 封装，支持程序化管理 Skill 全生命周期：

- **元数据与内容包编辑**：`skill update-metadata` 支持更新名称、描述、分类等元信息；`skill update-package` 支持替换内容包（SKILL.md 或 ZIP）
- **状态管理**：`skill status` / `skill set-status` 查看与切换发布状态；`skill delete` 支持删除 Skill
- **版本历史管理**：`skill history` 查看历史版本列表；`skill republish` 将历史版本恢复为草稿；`skill publish-history` 直接发布指定历史版本
- **草稿内容读取**：`skill management-content` / `skill management-read-file` / `skill management-download` 覆盖草稿内容索引、指定文件读取与完整 ZIP 下载
- Python 与 TypeScript SDK 均完成对应 `SkillsResource` 方法封装，覆盖全部新增命令

---

### 【Decision Agent】

0.8.0 周期内 Decision Agent 发布 v0.7.1，核心围绕工具命名规范化、Skill 内置工具注册链路重构与用户文档体系建设三个方向。

**1. 工具名称规范化校验与 DeepSeek 兼容**

为提升 Agent 在 DeepSeek 等模型下的工具调用成功率，完成内置工具名称规范整治并建立注册阶段校验机制：

- **存量内置工具重命名**：将不符合命名规范的内置工具名称统一调整为合规格式（如将 `获取agent详情` 重命名为 `get_agent_detail`），移除已下线接口对应的废弃工具
- **注册阶段校验**：在工具注册时校验名称仅包含 `a-z`、`A-Z`、`0-9`、`_`、`-`，且长度不超过 64 字符；发现不合规名称时记录警告日志，避免运行时静默失败
- **Dolphin 依赖升级**：`kweaver-dolphin` 升级至 v0.7.6，修复 DeepSeek v4 兼容性问题

**2. Skill 内置工具注册链路重构**

重构 Skill 内置工具的服务边界与装配机制：

- **职责解耦**：Skill API 能力切换为 Agent Operator Integration 服务，修复原来错误的下游服务依赖
- **注册链路明确**：将 Skill contract tools 移入 agent core logic 层，通用工具装配与平台内置工具注入逻辑解耦
- **工具接口补齐**：新增 Skill HTTP 端点、请求/响应模型与 OpenAPI 工具定义，补充 `skill_tools.json` 初始化配置

**3. React 模式全流程 CLI 优化与增强**

在 0.7.0 新增 React Agent 运行模式基础上，0.8.0 完成 React 模式的 CLI 全流程配置增强：

- **配置文件支持**：支持通过配置文件声明 `agent_mode: react`，无需每次在命令行传参，降低复杂配置场景下的操作成本
- **`disable_history_in_a_conversation`**：明确支持在配置文件中关闭对话内历史记录，适用于不依赖上下文的独立任务场景，可有效降低 token 消耗
- **配置校验**：新增带校验的 Agent `Config` 值对象，为 Agent 配置和 Dolphin/React 模式补充完整的结构校验，配置错误时返回明确错误提示；同步更新 Agent Mode 文档与仓库 Agent 配置说明

**4. 用户文档与示例体系建设**

- **用户手册**：新增完整的 Decision Agent 用户手册，覆盖 API、CLI、概念说明、TypeScript SDK 指南及聚合版文档，提供可直接运行的 API、CLI、SDK 示例
- **Cookbook**：新增集成 Cookbook 场景文档，覆盖合同摘要、Sub-Agent 合同审查、人工干预/终止等完整场景
- **示例重组**：按能力目录拆分 API/CLI/TypeScript SDK 示例与 Makefile target，新增共享环境与状态处理，便于示例流程复用


---

### 【沙箱运行时】

0.8.0 沙箱运行时升级至 v0.4.0，核心交付多语言复合模板环境与 Go 运行时正式支持，同时在镜像分层架构、文件安全处理和依赖安装可靠性方面完成多项改进。

**1. 多语言复合模板环境**

新增内置 `multi-language` 模板，支持 Python、Go、Bash 复合执行环境，一个会话内可混合运行多种语言的 Skill：

- **Go 1.25.2 运行时**：在 multi-language runtime base 中内置 Go 1.25.2，并将 Go build/module 缓存配置到 `/workspace/.cache`，Bubblewrap 与 subprocess 两种执行路径均可直接使用 `go` 命令
- **模板独立 Helm 覆盖**：新增 `image.defaultTemplates.pythonBasic` 与 `image.defaultTemplates.multiLanguage` Helm values，支持部署时分别覆盖两个内置模板的镜像版本，满足离线或定制化部署场景
- **默认模板自动回退**：create-session 请求未传 `template_id` 时自动使用 `DEFAULT_TEMPLATE_ID`，无需显式指定模板

**2. 稳定 Runtime Base 镜像分层**

镜像构建架构完成分层重构，将运行时依赖与执行器应用代码解耦：

- **稳定 base 层**：新增只包含系统依赖和语言运行时的稳定 Python 与 multi-language runtime base 镜像，不随应用代码变更重建，显著减少镜像层变动范围
- **版本化 executor 层**：最终 executor/template 镜像 tag 跟随项目 `VERSION`，重型 runtime 层保持稳定，升级时只更新轻量应用层
- **共享 executor Dockerfile**：移除旧的按模板拆分 Dockerfile，统一为共享 executor template Dockerfile，简化镜像维护

**3. Session 依赖安装超时配置**

手动增量安装 session 依赖请求新增 `install_timeout` 参数支持，超时时间透传到 executor session-config sync 调用：

- 避免大型 Skill 依赖包安装时间过长被 executor client 默认超时限制截断导致的安装失败
- 文件上传大小校验同步改为读取配置项，不再硬编码 100 MB 限制，支持按部署需要调整

**4. 文件上传安全增强**

ZIP 压缩包解压处理新增多项安全防护：

- 新增文件数量与总解压大小双重限制，防止解压炸弹
- 解压时拒绝符号链接条目，降低不安全压缩包处理风险

**5. 问题修复**

- 修复默认 seed 的模板镜像地址固定为 `v1.0.0` 或 `latest` 的问题，现改为跟随 `VERSION`、`TEMPLATE_IMAGE_TAG` 或 `PROJECT_VERSION`，确保版本升级后模板镜像自动对齐

---

### 【部署与基础设施】

**1. KWeaver Core 部署流程优化**

重新设计 KWeaver Core 部署流程，降低出错风险

- 提供部署前预检脚本，便于提前发现环境问题并在需要时辅助修复
- 提供安装完成后的上架向导脚本，基础初始化等收尾步骤
- 安装流程与文档说明同步更新，降低首次部署与重复运维的试错成本

**2. KWeaver Core 单机部署支持**

KWeaver Core 新增基于K3s（Linux） 和 Docker + Kind（MacOS） 的低资源单机部署方案：

- 一键完成 KWeaver Core 的本地或单机服务器安装，适用于开发测试、评估验证与资源受限的生产环境
- 一键完成环境检查和修复和初始化配置
- 详细安装指引见：[https://github.com/kweaver-ai/kweaver-core/tree/main/deploy](https://github.com/kweaver-ai/kweaver-core/tree/main/deploy)

**3. OSS 网关支持火山云 TOS**

对象存储网关（OSS Gateway）新增火山云 TOS 适配，补充国内主流对象存储方案覆盖。

---

### 【Dataflow】

0.8.0 版本 Dataflow 完成版本升级并新增 Dataset 写入节点能力。

**1. Dataset 写入节点支持（@dataset/write-docs）**

Dataflow 流水线新增 `@dataset/write-docs` 写入节点，支持将流水线处理结果直接写入 VEGA Dataset，打通数据流转至知识网络的最后一公里：

- 可在 Dataflow 流程中配置 Dataset 写入节点，将上游节点输出持久化到指定 Dataset
- 与现有 VEGA Resource/Dataset API 统一对接，复用已有的数据审计与版本机制

---

### 【ISF 信息安全编织】

0.8.0 版本在身份安全与部署依赖两个方向完成若干改进。

**1. 多端并发登录管控**

新增同一账号的多端登录安全控制：

- 禁止同一账号在多个终端或多个浏览器上同时登录，避免身份凭据的并发滥用
- 登录成功后展示上一次登录信息（包括 IP 地址与登录时间），帮助用户及时发现异常登录

**2. 认证与组织架构同步插件解耦对象存储**

认证与组织架构同步插件不再依赖对象存储（OSS），降低安装部署门槛，简化涉密场景下的服务依赖配置。

---

### 【示例库更新】

0.8.0 版本新增两个完整端到端示例：

**Example 04：多智能体 session_id 传递（Multi-Agent Custom-Input Propagation）**

演示在多智能体调用链中通过自定义输入传递 `session_id`，实现跨 Agent 对话上下文共享，覆盖完整的 session_id E2E 场景。

**Example 05：基于知识网络的 Skill 路由（KN-Driven Skill Routing）**

演示 Agent 如何基于 BKN 知识网络语义动态路由到合适的 Skill，提供可直接运行的端到端 Demo 与配套博客说明。

---

## 版本发布

### 1. GitHub 安装包和技术文档

**KWeaver Core**
- GitHub Release: https://github.com/kweaver-ai/kweaver-core/tree/release/0.8.0

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk


---
