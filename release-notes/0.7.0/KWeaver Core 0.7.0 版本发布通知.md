# KWeaver Core 0.7.0 版本发布通知

---

## 版本概述

### KWeaver Core 0.7.0 版本正式发布

**指标语义融合，流式构建就绪，智能体能力多维跃升**

KWeaver Core 0.7.0 以 **BKN 引擎双向升级**为核心主题。在语义建模层，指标类型（Metric）正式成为业务知识网络的第四类一等公民，与对象类、关系类、行动类共同表达业务实体与量化逻辑；关系映射新增 `FilteredCrossJoinMapping` 类型，支持在两个数据源的笛卡尔积上施加过滤条件精确筛选关联实例，覆盖复杂多表关联场景；BKN Specification SDK 同步升级至 v0.1.3，并新增 Exchange 邮件恢复领域知识网络模板。本版本同步完成 VEGA 对 PostgreSQL 流式索引构建的支持，三种构建模式（全量、增量、流式）至此完整覆盖；逻辑视图引擎新增多表 JOIN、UNION 视图类型及自定义 SQL 能力；执行工厂完成 Skill 生命周期管理闭环，支持整包更新、版本控制与历史回滚；Context Loader 完成工具合并并新增指标类型概念召回，覆盖对象、关系、行动、指标四类实体的统一语义搜索能力；Decision Agent 新增 React Agent 运行模式。整体系统在 API 错误语义、并发稳定性与部署运维体验方面同步完成多项专项改进。

---

## KWeaver Core 0.7.0 版本核心亮点速览

**1. BKN 引擎双向升级：指标模型语义化与过滤交叉关联映射**

BKN 引擎在两个方向完成重要突破。**指标模型**：指标类型（Metric）正式成为 BKN 第四类一等公民，与对象类、关系类、行动类并列，支持计算公式、分组维度、时间维度与多种查询模式（即时、趋势、同环比、占比）；Context Loader 同步新增指标语义召回，打通"指标建模→语义召回→智能体问答"链路。**关系映射**：新增 `FilteredCrossJoinMapping` 类型，支持在两个数据源的笛卡尔积上施加过滤条件精确筛选关联实例，覆盖物料-库存-订单等复杂多表关联场景，与原有直接映射、间接映射共同形成完整映射体系。

**2. PostgreSQL 流式索引构建正式上线，数据变更实时入库**

VEGA新增 PostgreSQL 流式索引构建，在既有全量构建与增量构建的基础上补齐流式实时模式：数据库中的行级变更（INSERT / UPDATE / DELETE）通过 PG 变更监听机制实时同步到搜索索引，无需定时全量刷新。至此，VEGA 对 PG 数据源的"批次（全量、增量）+ 流式"三种构建模式完整覆盖，满足对数据时效性有严格要求的实时分析场景。

**3. 执行工厂 Skill 生命周期全管理，版本控制与历史回滚就绪**

执行工厂 将 Skill 的管理能力从"注册-执行"升级为完整的生命周期管理：支持对已发布 Skill 进行元数据编辑（草稿态与发布态分离）、整包版本更新、历史版本查询与一键回滚，同时完成 Skill Dataset 重建补偿与增量构建能力。企业可在不中断线上服务的情况下进行 Skill 迭代，版本管理机制确保每次变更可溯源、可回退。

**4. Decision Agent 新增 React 运行模式，复杂任务推理能力升级**

Decision Agent引入 React Agent 运行模式（`agent_mode: react`），在原有 Dolphin 模式基础上新增独立的 ReactConfig 配置（支持对话历史开关与 LLM 缓存控制）。React 模式在多步推理、工具调用链较长的场景下具有更强的上下文连续性，配合新增的可配置内置 Skill 注册规则，智能体在复杂决策任务中的执行效率显著提升。

**5. 逻辑视图引擎扩展：新增多表 JOIN / UNION 与自定义 SQL**

VEGA 逻辑视图新增多表 JOIN（左连接等）与多表 UNION 两种视图类型，以配置化节点方式定义数据来源、关联键与输出字段，无需编写 SQL 即可完成多表联合查询建模；同时新增自定义 SQL 视图模式，支持完整 SQL 语句的灵活查询定义。三种视图类型统一通过 VEGA 统一查询接口对外暴露，底层无缝适配 MariaDB、PostgreSQL 与 OpenSearch 多种数据源。

---

## 详细功能说明

### 【BKN 引擎】

0.7.0 版本在指标模型与关系映射能力两个方向完成重要升级，同时完善边界校验与错误语义。

**1. BKN 指标模型：业务语义化指标定义与查询**

BKN 新增指标类型（Metric Type），支持在知识网络中直接定义具有业务语义的量化指标：

- **指标结构建模**：指标通过 `scope`（所属对象类）、`计算公式`（过滤条件 + 聚合方式 + 分组字段）、`时间维度`、`分析维度`等要素完整描述业务量化逻辑。支持原子指标类型（`atomic`），复合指标类型后续迭代补充
- **多维查询模式**：指标支持即时查询（`instant=true`，获取当前汇总值）、趋势查询（`instant=false` + 时间范围，按日/月/年等日历步长返回时序数据）、同环比分析（`type=parallel`，配置偏移量后计算增长值与增长率）及占比分析（`type=proportion`，按分析维度返回各维度占比百分比）
- **语义检索**：指标支持语义化概念检索，通过 ID、名称、描述生成向量索引，支持基于自然语言的指标发现
- **独立概念地位**：指标在 BKN 中作为独立概念类型存在，不与对象类等其他概念类型产生直接关联，但可通过 scope 引用指定对象类来限定统计主体

**2. Condition 配置重构与校验完善**

重构 BKN Condition 配置结构，将 `ActionCondCfg` 从 `CondCfg` 中分离为独立结构体，字段名从 `Name` 统一调整为 `Field` 以明确语义：

- 完善对象类逻辑属性绑定资源在严格模式（`strict_mode`）下的 Condition 校验
- 自动补全指标类型属性的系统参数，降低手动配置成本
- 优化关系类型校验逻辑，新增非严格模式支持
- 行动来源类型有效值从 `tool/map` 更新为 `tool/mcp`，与当前工具体系对齐

**3. FilteredCrossJoinMapping：过滤式交叉关联映射**

BKN 关系类型新增 `FilteredCrossJoinMapping` 映射类型，支持在两个数据源的笛卡尔积上施加过滤条件，从中精确筛选满足关联条件的实例对：

- 解决了原有关系映射只能表达简单字段等值连接的限制，允许在关系类型中内联声明复杂的多字段过滤逻辑
- 适用于"从资源 A 和资源 B 的组合中，按业务规则筛选有效关联"的多表关联场景，如物料-库存-订单的条件交叉匹配
- 与现有 `DirectMappingRule` / `InDirectMappingRule` 并列，三种映射类型共同覆盖简单映射、间接映射与过滤交叉映射全场景

**4. API 错误语义规范化**

统一资源未找到场景的 HTTP 状态码为 404（原为 403），覆盖知识网络、概念组、行动类型、关系类型、对象类型等全部资源类型。Action-type execute 接口在 input 参数缺失时由静默丢弃改为返回明确的 400 错误，提升智能体错误路径的可诊断性。

---

### 【VEGA 数据虚拟化】

0.7.0 版本在数据构建模式、逻辑视图能力与数据源接入三个方向完成重要能力扩展。

**1. PostgreSQL 流式索引构建**

VEGA 新增 PostgreSQL 流式索引构建，至此 PG 数据源覆盖三种构建模式：

- **全量构建**：一次性将 PG 表全量数据同步到搜索索引
- **增量构建**：基于增量字段仅同步上次构建以来新增或变更的数据
- **流式构建**：通过 PG 变更监听机制实时捕获行级变更（INSERT / UPDATE / DELETE），数据变更实时同步到索引，满足毫秒级数据时效要求

> 注意：流式构建依赖 PG 服务端开启相关变更监听开关（Logical Replication），未开启时构建任务将报错提示，请在启用前确认数据库配置。

**2. 逻辑视图扩展：新增 JOIN、UNION 与自定义 SQL**

逻辑视图（Logic View）从原有基础查询扩展为支持三种完整视图类型：

- **JOIN 视图**：支持多表左连接（Left Join）等关联操作。通过配置化节点声明来源 Resource、关联键与输出字段，系统自动将配置转化为对应 SQL 执行；支持 MariaDB 与 PostgreSQL 数据源
- **UNION 视图**：支持多表 UNION 合并查询，将同结构不同来源的数据合并为统一视图；支持 MariaDB 与 OpenSearch 混合数据源
- **自定义 SQL 视图**：支持完整 SQL 语句定义，通过模板化语法（`{{.nodeId}}`）引用配置中声明的 Resource，兼顾灵活性与安全性

所有逻辑视图通过 VEGA 统一查询接口（`/resources/{id}/data`）对外暴露，支持分页、排序、条件过滤与 limit 参数。

**3. AnyShare 文档库接入与过滤条件支持**

VEGA Catalog 在 0.6.0 AnyShare 知识库接入基础上，0.7.0 新增对 **AnyShare 文档库** 类型的支持，实现企业文档数据（Word、PDF 等格式文件）纳入业务知识网络管理；同时新增 `cond` 条件过滤参数，支持在 Catalog 发现阶段基于业务条件筛选文档库/知识库内容，减少不必要的数据摄取量。

**4. Resource 数据查询能力增强**

Resource Data 查询接口（`/resources/{id}/data`）新增聚合分析支持：

- **聚合/分组**：支持 `GROUP BY` 分组统计
- **Having 过滤**：支持对聚合结果的二次条件过滤
- **查询参数修复**：修复 limit 参数在部分场景下未生效的问题，确保查询结果分页行为符合预期

---

### 【执行工厂】

0.7.0 版本完成 Skill 生命周期管理的完整闭环，从单纯的注册-执行能力升级为支持版本控制、元数据迭代与数据补偿的全生命周期管理体系。

**1. Skill 整包更新与元数据编辑**

支持对已发布的 Skill 进行两种更新路径：

- **元数据编辑**：在不替换 Skill 执行包的情况下，对 Skill 的名称、描述、分类等元数据进行修改，更改后生成草稿版本（编辑态），与当前发布版本并存，不影响线上调用
- **整包更新**：重新上传 Skill 完整包，以新版本替代现有实现，同样先进入草稿态，经确认后发布上线

两种更新路径均通过显式发布操作使变更生效，确保线上版本稳定性。

**2. 版本管理与历史回滚**

每次发布操作均记录版本历史，支持以下版本管理操作：

- **历史版本查询**：查看指定 Skill 的完整版本变更记录
- **历史版本回灌**：将任意历史版本回灌至草稿箱，作为下一次发布的候选版本
- **历史版本直接发布**：支持直接将历史版本发布为线上版本，实现快速回滚

线上始终只保留一个正式发布版本，Dataset 中存储的始终是线上最新版本对应的索引数据。

**3. Skill Dataset 重建补偿与增量构建**

新增 Skill Dataset 重建补偿机制：在 Skill 版本更新后，自动触发关联 Dataset 的重建或增量更新任务，确保 Dataset 中的索引数据始终与当前发布版本保持一致。增量构建仅处理自上次构建以来发生变更的部分，降低全量重建的性能开销。

---

### 【Context Loader】

0.7.0 版本完成工具整合与能力扩展，将指标类型纳入统一语义召回体系，进一步降低智能体调用成本。

**1. search_schema 工具合并：四类概念统一召回入口**

将原有 `kn_schema_search` 与 `kn_search` 两个工具合并为统一的 `search_schema` 工具，覆盖对象类、关系类、行动类、**指标类** 四种 BKN 概念类型的统一语义召回：

- **search_scope 参数**：通过 `include_object_types`、`include_relation_types`、`include_action_types`、`include_metric_types` 四个开关灵活控制召回范围，默认全类型覆盖
- **精简响应模式**：支持 `schema_brief` 参数，开启后返回精简的概念摘要格式，降低 Token 消耗
- **响应格式**：HTTP 接口默认 JSON，MCP Tool 默认 toon 压缩格式；原有两个工具保留在工具集中但不对外通过 MCP 暴露

**2. 指标类型概念召回**

基于 `search_schema` 工具，智能体现可通过自然语言查询直接发现知识网络中定义的指标：

- 返回指标的 `id`、`name`、`comment`（业务说明）及 `metric_type`
- 支持基于指标名称与描述的语义向量匹配
- 支持混合召回：查询 "概念数量" 等语义时可同时返回对象类与相关指标，通过交叉评分确定最终排序

**3. 僵尸依赖清理**

移除 `agent-retrieval` 中被硬编码 feature flag 旁路的 `data_retrieval` 僵尸依赖，清理代码中已失效的配置项，确保工具召回路径行为与代码声明一致。

---

### 【Decision Agent】

0.7.0 版本在运行模式与可配置性两个维度完成重要升级，同时简化了可观测性系统，提升运维可维护性。

**1. React Agent 运行模式**

新增 `agent_mode: react` 运行模式，通过 `/v3/agent/react` 端点创建 React Agent 配置：

- **ReactConfig**：独立的 React 模式配置项，支持 `disable_history_in_a_conversation`（关闭对话内历史）与 `disable_llm_cache`（关闭 LLM 缓存）
- **AgentMode 枚举**：新增 `default / dolphin / react` 三种模式枚举，为后续模式扩展预留扩展点
- React 模式在多步工具调用链场景下具有更强的上下文连续性，适用于需要精确跟踪推理步骤的复杂决策任务

**2. 内置 Skill 注册规则可配置化**

支持通过配置文件控制内置 Skill 的注册行为与调用规则（`skill_enabled` 配置项），灵活适配不同部署场景下对内置能力的开放范围需求。

**3. Agent 模板复制修复**

修复复制 Agent 模板时 `published_at` 和 `published_by` 字段未被清除的问题，确保复制出的新模板始终以未发布状态初始化，避免状态混淆。

**4. OAuth Bearer 转发修复**

修复 `agent-executor` 中 OpenAPI Tool 不转发 OAuth Bearer Token 的问题，确保通过 Toolbox 注册的 OAuth 保护下游服务可被 Decision Agent 正常调用。

**5. 可观测性系统简化**

将可观测性实现从自定义 O11Y 追踪迁移至标准 OpenTelemetry（OTLP）链路，移除过时的 `observability handler` 与相关冗余组件，统一各服务（`agent-factory`、`agent-executor`、`agent-memory`）的遥测配置结构，降低运维复杂度。同时新增 LLM 消息日志功能（`LLMMessageLoggingConfig`），用于开发调试阶段捕获 LLM 输入输出的完整消息。

**6. 废弃配置清理**

移除以下已废弃的服务配置：`kn_data_query`、`kn_knowledge_data`、`data_connection`、`search_engine`、`ecosearch`、`ecoindex_public` 等；移除 `disable_biz_domain_init` 配置项，简化业务域初始化逻辑。

---

### 【沙箱运行时】

0.7.0 Sandbox新增控制平面接管能力，大幅提升 Kubernetes 环境下运维升级的稳定性。

**1. 控制平面接管（Control Plane Takeover）**

新增控制平面启动时对既有会话 Pod 的接管能力：

- 控制平面重启或升级后，自动扫描并接管 Kubernetes 集群中仍在运行的存量会话 Pod，恢复会话与执行器的绑定关系
- 通过 Pod Owner Reference 机制实现执行器 Pod 与控制平面实例的亲和绑定，避免孤立 Pod 持续消耗资源
- 处理同名 Pod 重建冲突：等待处于 Terminating 状态的旧 Pod 完全删除后再创建新 Pod，消除启动状态同步阶段的竞争条件

**2. 会话 Pod 资源配置优化**

K8s 调度模式下，会话 Pod 的 CPU 与内存 Request 调整为零（保留 Limit），减少因 per-session 资源预留导致的 Pod 调度失败，提升密集会话场景下的调度成功率。

---

### 【ISF 信息安全编织】

0.7.0 完成 ISF 服务整合与能力扩展，移除 `sharemgnt-single` 独立服务，相关功能统一收归 ISF 主服务，降低整体部署服务数量与内存占用。

**1. 个人访问令牌（PAT）支持**

Authentication 模块新增个人访问令牌（Personal Access Token）能力，支持在 ISF 管理界面创建、查询与吊销 PAT：

- 每个账户最多可创建 10 个 PAT（可配置）
- 支持永久有效或指定到期时间
- Token 生命周期通过 Hydra OAuth2 统一管理，与平台认证体系深度集成
- 数据库迁移支持 MariaDB、DM8、KDB9 三种数据库

**2. 批量用户导入/导出**

UserManagement 新增批量用户管理能力，支持通过 Excel 文件批量创建账号：

- **批量导入**：通过标准 Excel 模板导入用户，支持姓名、所属部门（多级，以 `/` 分隔）、手机号、有效期、存储配额、初始密码、用户密级等字段；红色字段为必填，黑色字段选填
- **批量导出**：将当前用户列表导出为 Excel，支持异步后台任务执行，大数据量场景下不阻塞前台操作
- 国际化支持：简体中文与繁体中文填写说明随模板内嵌，无需额外文档

---

### 【BKN Specification SDK】

BKN Specification SDK 是独立发布的开放规范工具库，用于以代码方式解析、构建与序列化 BKN 业务知识网络规范文档。0.7.0 周期内发布稳定版本，完成底层重构并扩充示例库。

**1. 解析器与序列化器底层重构**

- **解析器重构**：引入 `extractSectionsWithDesc` / `buildDescription` 统一处理文档结构，摘要（summary）从描述内容首句自动提取，不再需要单独声明 `summary` 字段
- **序列化器优化**：提取通用 `encodeYAMLBlock` 函数，移除冗余字段（`sub_conds`、`value_from`），统一表格格式规范

**2. Exchange 邮件恢复领域模板与示例库扩展**

新增 Exchange 邮件恢复领域知识网络模板，覆盖数据域、Exchange 服务器、备份时间点、恢复任务、恢复作业等完整对象类型，包含备份时间点验证失败、生产数据覆盖两类风险行动及完整脚本，可直接作为企业邮件备份恢复场景的起点模板；k8s-network、供应链（supplychain-hd）、员工入职（mock_system）等现有示例同步适配 `filtered_cross_join` 关系类型与新格式规范。

---

### 【KWeaver SDK】

0.7.0 周期内 KWeaver SDK完成多项能力扩展，重点补全了 Toolbox/Tool 命令体系、Action Type 执行体验与 Python SDK 认证对齐。

**1. Toolbox / Tool 命令正式上线**

新增 `kweaver toolbox` 与 `kweaver tool` 命令族，覆盖工具箱与工具的完整生命周期管理：

- `kweaver toolbox create / list / publish / unpublish / delete`：工具箱创建、发布与下线管理
- `kweaver tool upload / list / enable / disable`：OpenAPI Spec 上传、工具启用/停用

配合新增的 `kweaver call -F <key>=@<file>` multipart 文件上传支持，完整替代了原 `examples/03-action-lifecycle` 中的手工 curl + call 组合流程。

**2. kweaver tool execute / debug**

新增 `kweaver tool execute` 与 `kweaver tool debug` 子命令，封装 Toolbox 代理调用所需的 header / query / body / timeout 枚举结构：

- 自动将当前登录 Bearer Token 注入 envelope，无需手动处理认证头
- 解决 agent-operator-integration 服务对 Authorization 头的转发限制
- Python / TypeScript SDK 同步新增 `ToolboxesResource.execute()` / `.debug()` 方法

**3. BKN Action-Type 执行体验升级**

新增 `kweaver bkn action-type inputs <kn> <at>` 命令，列出所有 `value_from=="input"` 的参数及其类型感知的起步模板，方便直接复制到 `--dynamic-params` 中使用；`bkn action-type execute` 同步新增 Flag 表单模式（`--dynamic-params / --instance / --trigger-type`），无需手工拼接 JSON envelope，与原位置参数形式保持向后兼容。

**4. Python SDK 认证完整对齐**

Python SDK 完成与 TypeScript CLI 的认证机制对齐：

- 新增 `HttpSigninAuth`，通过 RSA 密码加密 + OAuth2 redirect chain 实现无浏览器 HTTP 登录，覆盖 200+JSON redirect 等边缘场景
- `~/.kweaver` 存储格式（`displayName`、`logoutRedirectUri`、ISO-Z 时间戳）与 TS CLI 字节级对齐，`kweaver auth list/whoami` 可正常识别 Python 写入的会话
- 新增 `change_password` 方法，支持初始密码强制修改（HTTP 401 code `401001017` 自动触发引导）

**5. kweaver agent skill add / remove / list**

新增 Agent Skill 成员管理命令族，支持通过 CLI 直接为 Agent 配置挂载 Skill，替代原有手动编辑 Agent JSON config 的工作流：

- `kweaver agent skill add <agent_id> <skill_id>`：为 Agent 添加 Skill 绑定
- `kweaver agent skill remove <agent_id> <skill_id>`：移除 Skill 绑定
- `kweaver agent skill list <agent_id>`：查看当前 Agent 已绑定的 Skill 列表

---

### 【部署与基础设施】

**1. kweaver-admin：平台管理员 CLI 正式发布**

`kweaver-admin` 是面向平台管理员的独立命令行工具，在安装了 ISF（信息安全编织）的环境下，支持通过 CLI 完成用户、角色与模型的完整管理操作，无需登录 Web 控制台。

```bash
npm install -g @kweaver-ai/kweaver-admin
kweaver-admin auth login https://your-platform.example/
```

核心能力覆盖以下管理域：

- **用户管理**（`user`）：列出、查询、创建、更新、删除用户，管理员重置密码，查看用户已分配角色
- **角色管理**（`role`）：列出角色与成员，为用户分配/撤销角色
- **组织/部门管理**（`org`）：列表、树形结构、创建、更新、删除部门，查看部门成员
- **大模型管理**（`llm`）：新增、编辑、删除、测试大模型配置
- **小模型管理**（`small-model`）：新增、编辑、删除、测试小模型配置
- **审计查询**（`audit`）：按用户、时间范围查询登录审计事件
- **原始 HTTP 调用**（`call` / `curl`）：携带认证头的任意 API 调用

`kweaver-admin` 同时提供 Agent Skill 形态，可通过 `npx skills add` 安装到支持 skill 加载的 Agent 工作流中，让 AI 助手直接以自然语言完成平台管理操作：

```bash
npx skills add https://github.com/kweaver-ai/kweaver-admin --skill kweaver-admin
```

**2. 稳定性修复**

- 修复 DAG 变量创建（`CreateDagVars`）与索引刷新（`refreshDagIndexes`）中的死锁问题，消除并发操作下的概率性阻塞
- 修复 vega-backend 中因冗余唯一索引（`uk_catalog_source_identifier`）导致的数据插入异常
- 修复模型工厂应用账户无法删除小模型的问题

---

## 版本发布

### 1. GitHub 安装包和技术文档

**KWeaver Core**
- GitHub Release: https://github.com/kweaver-ai/kweaver-core/tree/release/0.7.0

**KWeaver SDK**
- GitHub: https://github.com/kweaver-ai/kweaver-sdk

**kweaver-admin**
- GitHub: https://github.com/kweaver-ai/kweaver-admin
- npm: `npm install -g @kweaver-ai/kweaver-admin`

---
