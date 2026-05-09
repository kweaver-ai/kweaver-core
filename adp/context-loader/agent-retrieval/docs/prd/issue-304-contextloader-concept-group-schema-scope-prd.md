# PRD: 【context-loader】支持 BKN concept_group 概念分组语义，限定 Schema 召回范围

> 状态: Draft
> 关联 Issue: [#304](https://github.com/kweaver-ai/kweaver-core/issues/304)
> 文档说明: 本 PRD 定义当前需求的背景、目标、范围、业务语义、兼容要求和验收标准；其中“验收矩阵”是判断本需求是否完成的权威口径。对应设计文档只说明工程方案和验证清单，不定义第二套验收口径。

## 1. 问题

ContextLoader 的 `search_schema` 是 Agent / MCP / HTTP 调用方探索 BKN Schema 的统一入口。当前调用方可以控制是否返回对象类、关系类、行动类和指标类，但不能通过 BKN `concept_group` 限定探索范围。

在业务知识网络较大或存在清晰业务域分组时，调用方需要只在指定概念分组内探索相关 Schema。如果 `search_schema` 不能消费 `concept_groups`，Agent 可能召回分组外概念，导致后续 `query_*`、`get_*` 或执行规划进入错误业务边界。

## 2. 用户 / 调用方

- MCP / Agent 调用方：通过 `search_schema` 发现对象类、关系类、行动类和指标类。
- HTTP 集成调用方：直接调用 `/kn/search_schema`。
- 后续 SDK / CLI 包装方：需要稳定、可文档化的输入契约。

## 3. 目标

- 在 `search_schema.search_scope` 中支持 `concept_groups`。
- `concept_groups` 限定对象类、关系类和行动类的 Schema 召回范围；`metric_types` 请求携带同一范围参数，实际过滤命中依赖 BKN 侧实现。
- 不传或传空 `concept_groups` 时保持现有行为。
- 当 `concept_groups` 中含有未知分组时，将 BKN 返回的错误透传给调用方，而不是包装为空结果。
- 更新 MCP schema、HTTP OpenAPI、release 文档和 toolset 契约。

## 4. 非目标

- 不修改 BKN 概念分组的管理、创建、更新或校验能力。
- 不把 `concept_groups` 用作实例数据过滤条件。
- 不改造 `query_object_instance`、`query_instance_subgraph`、`get_logic_properties_values`、`get_action_info` 的实例级语义。
- 不实现概念分组树递归展开；本期按 BKN 当前扁平成员关系处理。
- 不修改兼容 `/kn/kn_search` 的历史契约。
- 不修改 legacy `/kn/semantic-search` 的历史契约。

## 5. 业务语义

- `concept_groups` 表示 BKN 概念层分组 ID 列表；BKN 分组以 `object_type` 集合作为直接边界。
- 多个分组采用并集语义。
- 省略 `concept_groups` 或传入空数组表示不启用分组范围。
- `object_types` 直接按分组成员召回。
- `relation_types` 由 BKN 按对象边界推导：关系两端 source / target 对象类都在分组内时，关系类属于该分组范围。
- `action_types` 由 BKN 按绑定对象推导：行动类绑定的 `object_type_id` 在分组内时，行动类属于该分组范围。
- ContextLoader 在 `concept_groups` 非空时调用 BKN typed search 作为分组语义事实来源；对 relation / action 引用但未被 object search 命中的对象类，按 ID 补齐对象详情，避免返回引用缺失的 schema。
- 当 BKN 判定请求分组均不存在时，当前会返回 5xx 错误（语义 `all concept group not found`）。ContextLoader 直接透传 BKN 错误，不吞错并伪装为空结果；部分分组不存在但仍有有效分组时，以 BKN 当前行为为准。
- `metric_types` 纳入 `/kn/search_schema` 的接口契约。当前 BKN metrics 搜索接口定义支持接收 `concept_groups` 参数，但功能过滤仍在 BKN 侧同步实现；ContextLoader 本期只保证 direct metric recall 与 expansion metric recall 都透传该字段，不在本地为 metrics 单独做过滤。

## 6. 兼容要求

- 现有不传 `concept_groups` 的调用方无需修改。
- 新增字段为 additive change。
- MCP 工具与 HTTP OpenAPI 使用相同字段路径：`search_scope.concept_groups`。
- 本需求的标准 HTTP 入口仅为 `/kn/search_schema`；新接入方应使用 `search_schema`。

## 7. 验收标准 / 验收矩阵

本矩阵是本需求 e2e 完成状态的判定标准。AI 交付时需要把实现、测试、文档和命令输出证据回填到这些场景上；若某项无法自动化验证，需要明确人工验证责任和残余风险。

| 场景 | 输入 / 准备 | 预期行为 | 证据要求 |
| --- | --- | --- | --- |
| 默认兼容 | 不传 `concept_groups` | `search_schema` 现有对象类、关系类、行动类、指标类行为不变 | 现有单元测试继续通过 |
| 分组范围（BKN 直查） | `search_scope.concept_groups=["cg-a"]` | ContextLoader 直接调用 BKN `SearchObjectTypes` / `SearchRelationTypes` / `SearchActionTypes`（带 `concept_groups`），不调用 `GetKnowledgeNetworkDetail`，结果中 `object_types`、`relation_types`、`action_types` 落在 `cg-a` 范围内 | 新增单元测试断言 BKN 入参与 detail 调用次数 |
| 引用对象补齐 | BKN relation/action 分组搜索返回的对象引用未出现在 object search 命中中 | ContextLoader 调用 `GetObjectTypeDetail` 补齐 relation source/target 与 action object_type_id 引用对象，再做对象选择、action name 映射和 metric expansion | 新增单元测试 |
| 多分组 | `concept_groups=["cg-a","cg-b"]` | 透传 BKN 并集语义 | 新增单元测试 |
| 空数组 | `concept_groups=[]` | 归一化后等价于未传，不进入 BKN 直查分支 | 新增归一化测试 |
| 未知分组（错误透传） | `concept_groups=["missing"]` | BKN 返回错误（当前为 5xx + `all concept group not found`），ContextLoader 直接向上传递错误，不返回空结果 | 新增边界测试断言 `err != nil` |
| 指标召回 | 开启 `include_metric_types` 且传 `concept_groups` | direct metric recall 与 expansion metric recall 都携带 `concept_groups` | mock 断言请求体 |
| 文档契约 | 查看 MCP / `/kn/search_schema` / release 文档 | 字段路径与错误语义（含 5xx 透传）一致 | 文档 diff review |

## 8. 开放问题与跨仓库依赖

| 问题 | 结论 |
| --- | --- |
| `metric_types` 是否受 `concept_groups` 约束 | `/kn/search_schema` 接口契约上保留该约束；ContextLoader 本期验收为透传 `concept_groups`。BKN metrics 当前接口定义支持该字段，但功能过滤仍在 BKN 侧同步实现；BKN 完成前，metrics 召回结果可能超出分组范围，这是已知跨仓库暂态。 |
| 未知分组是否返回 4xx | 否，ContextLoader 端不做错误码翻译。BKN 当前在判定请求分组均不存在时返回 5xx（`BknBackend.ObjectType.InternalError` + `error_details: "all concept group not found ..."`），ContextLoader 直接透传该响应。语义上引用了不存在的资源更适合用 4xx（建议 400 Bad Request 或 404 Not Found）表达，这是 BKN 侧推进的优化方向；BKN 调整前调用方按 5xx 处理。 |

跨仓库跟进：上述两项依赖均与 BKN 侧实现相关。BKN 团队已通过外部渠道知会并自行在 BKN 仓库登记 issue 跟踪，本 PRD 仅记录依赖事实与暂态行为，不在此处维护具体 issue 链接。
