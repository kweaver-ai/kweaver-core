---
issue: "#304"
branch: "feature/304-concept-group-schema-scope"
module: "context-loader/agent-retrieval"
status: "draft"
author: "@criller"
created: "2026-04-27"
pr: ""
---

# Design: 【context-loader】支持 BKN concept_group 概念分组语义，限定 Schema 召回范围

> 文档说明: 本设计文档定义工程实现方案、数据流、边界处理、测试策略和验证清单，用于说明如何证明 PRD 验收矩阵已经满足。本文件不新增或覆盖 PRD 的验收口径；如验证清单与 PRD 不一致，以 PRD 为准，并应先回到 PRD 修正需求语义。

## 背景与目标

GitHub issue `#304` 要求 ContextLoader 在 Schema 探索阶段消费 BKN `concept_group` 语义。本需求只作用于概念层发现：`concept_groups` 用于限制候选 `object_types`、`relation_types`、`action_types`；`metric_types` 在请求层透传同一范围参数，功能过滤依赖 BKN 侧实现。

目标：

- 在标准 `search_schema` MCP 与 HTTP 契约中新增 `search_scope.concept_groups`。
- 在 `search_schema` Schema 探索链路中透传同一组概念分组范围。
- 当调用方省略 `concept_groups` 或传入空数组时，保持现有行为不变。
- 当 `concept_groups` 中包含未知分组时，把 BKN 返回的错误向上透传，而非吞错并伪装为空结果。
- 明确 `concept_groups` 是 BKN 概念层的扁平分组，不是实例过滤器，也不是分组树遍历机制。

## 设计

### 概要

`concept_groups` 新增在现有 `search_scope` 对象下。请求在 `SearchSchema` 中统一归一化（trim + 去重）后，根据其是否非空，进入两条互斥的概念召回路径：

- **BKN 直查路径（`concept_groups` 非空）**：ContextLoader 直接调用 BKN 的 `SearchObjectTypes` / `SearchRelationTypes` / `SearchActionTypes`，在请求体里带上 `concept_groups`，由 BKN 在分组范围内完成召回。**不再调用 `GetKnowledgeNetworkDetail`，也不再做本地分组过滤**。任一搜索失败立即向上透传错误，调用方据此区分"分组合法但无概念"与"分组不存在"。relation/action 返回后，ContextLoader 会补齐它们引用但未被 object search 命中的对象详情。
- **全量召回路径（`concept_groups` 为空）**：沿用原有的"`GetKnowledgeNetworkDetail` → 可选粗召回 → 关系排序 → 对象选择 → 属性裁剪"。该路径保持与本期改造前一致，零行为变化。

BKN 的 concept_group 以 `object_type` 集合作为直接边界；`relation_type` 由 source / target 对象都在范围内推导，`action_type` 由绑定对象在范围内推导。ContextLoader 本期不复制这套分组推导逻辑，而是在 `concept_groups` 非空时调用 BKN typed search，把 BKN 作为分组语义事实来源。

`metric_types` 路径不变：`SearchSchema` 在 direct metric recall 与 metric expansion recall 的 `QueryConceptsReq` 中携带 `concept_groups`，复用 BKN metrics 搜索接口对该字段的接收。BKN 端当前接口定义支持该字段，功能过滤仍在同步实现（详见"开放问题与跨仓库依赖"）。ContextLoader 本期只保证透传，不为 metrics 单独做本地过滤。

### API 变更

标准 HTTP 与 MCP 请求形态：

```json
{
  "kn_id": "kn_supply_chain",
  "query": "inventory risk",
  "search_scope": {
    "concept_groups": ["supply_chain"],
    "include_object_types": true,
    "include_relation_types": true,
    "include_action_types": true,
    "include_metric_types": true
  },
  "max_concepts": 10
}
```

兼容说明：

- 省略 `concept_groups`：不启用分组范围，保持现有行为。
- 传入空数组 `concept_groups: []`：归一化后等价于省略，仍走全量召回路径。
- 分组 ID 均不存在：BKN 当前返回 `BknBackend.ObjectType.InternalError` (HTTP 500) + `error_details: "all concept group not found ..."`。ContextLoader 直接向上传递该错误，HTTP 层据此返回 5xx；调用方对未知分组的错误恢复语义保持与 BKN 一致。BKN 把该错误改为 4xx 是另立 issue 推进的优化，不在本期改造内做翻译。部分分组不存在但仍有有效分组时，以 BKN 当前行为为准。
- 分组合法但范围内没有匹配概念：BKN 返回成功 + 空 entries，ContextLoader 返回空 Schema 桶。
- `concept_groups` 只作用于 Schema / 概念发现，不过滤对象实例、关系实例、逻辑属性值或行动执行。
- `/kn/kn_search` 不属于本 issue 变更范围；legacy `/kn/semantic-search` 不纳入本次改造。

### 数据模型变更

不涉及持久化数据模型变更。

ContextLoader 仅扩展接口结构体以承载 BKN 既有字段：

- `SearchSchemaScope.ConceptGroups`
- 内部 `KnSearchLocalRequest` / `KnSearchConceptRetrievalConfig` 的 `ConceptGroups`
- `QueryConceptsReq.ConceptGroups`

不再保留"用于过滤 BKN 导出详情的本地分组成员字段"，相关本地过滤函数（`filterNetworkDetailByConceptGroups` 与 `filter*TypesByID`）已删除。

### 核心流程

1. MCP 或 HTTP 调用方调用 `search_schema`，可选传入 `search_scope.concept_groups`。
2. `NormalizeSearchSchemaReq` 对分组 ID 执行 trim + 去重。
3. `SearchSchema` 委托到 `KnSearch`。
4. **概念召回路由**：
   - 若 `config.ConceptGroups` 非空 → 进入 `conceptRetrievalByGroups`：
     a. 用查询 `query` 构造 BKN `QueryConceptsReq`（`knn` + `match` 子条件、`concept_groups` 透传）；
     b. 顺序调用 `SearchObjectTypes` / `SearchRelationTypes` / `SearchActionTypes`，任一失败立即返回错误；
     c. 收集 relation source/target 与 action object_type_id 中缺失的对象 ID，调用 `GetObjectTypeDetail` 补齐对象详情；
     d. 对补齐后的召回结果复用既有的 `rankRelationTypes`、`selectObjectTypesForConceptRetrieval` 与 `convert*ToLocal`，并在需要时调用 `fetchSampleData`。
   - 否则进入历史 `conceptRetrieval` 路径，行为与本期改造前完全一致。
5. metric direct recall 与 metric expansion 请求始终携带 `concept_groups`（与是否进入直查路径无关）；BKN metrics 功能过滤完成前，ContextLoader 只保证请求透传。
6. 最终输出过滤继续应用 include 开关与 `max_concepts`。

## 工程验证清单

本清单用于证明 PRD 验收矩阵已满足。每一项都需要能追溯到 PRD 中的验收场景；不能追溯的项只作为工程质量检查，不能改变需求范围。

| PRD 验收场景 | 工程验证项 | 证据 |
| --- | --- | --- |
| 默认兼容 | 不传 `concept_groups` 时，`search_schema` 保持现有对象类、关系类、行动类和指标类行为 | 回归单元测试 + `go test ./adp/context-loader/agent-retrieval/server/...` |
| 分组范围（BKN 直查） | 传入 `concept_groups` 时进入 `conceptRetrievalByGroups`，BKN 三个 typed search API 均被调用并带上 `concept_groups`，`GetKnowledgeNetworkDetail` 调用次数为 0 | `TestConceptRetrieval_GroupScopeDelegatesToBkn` |
| 引用对象补齐 | relation/action 返回引用了 object search 未命中的对象时，调用 `GetObjectTypeDetail` 补齐对象详情，并用于 action object_type_name 与 metric expansion | `TestConceptRetrieval_GroupScopeCompletesReferencedObjects` |
| 多分组 | `concept_groups` 携带多个 ID 透传到 BKN | 同上单测 + 归一化测试 |
| 空数组 | `concept_groups: []` 归一化为空，进入全量召回路径 | `TestNormalizeSearchSchemaReq_NormalizesConceptGroups` |
| 未知分组（错误透传） | BKN 返回错误时 ContextLoader 立即透传，不调用后续 search API、不调用 `GetKnowledgeNetworkDetail`、不返回空结果 | `TestConceptRetrieval_UnknownConceptGroupPropagatesError` |
| 指标召回 | direct metric recall 与 metric expansion 都携带 `concept_groups` | `TestSearchSchema_MetricQueriesCarryConceptGroups` |
| 文档契约 | MCP schema、HTTP OpenAPI、release 文档与 toolset 契约同步描述错误透传与分组语义 | 文档 diff、JSON/toolset 结构校验 |

## 测试策略

- 单元测试请求归一化：覆盖 trim、去重、省略与空 `concept_groups`（`TestNormalizeSearchSchemaReq_NormalizesConceptGroups`）。
- 单元测试 BKN 直查路径：确认 BKN 三个 typed search API 都被调用、`concept_groups` 透传、`GetKnowledgeNetworkDetail` 不被调用（`TestConceptRetrieval_GroupScopeDelegatesToBkn`）。
- 单元测试引用对象补齐：确认 relation source/target 与 action object_type_id 缺失时批量调用 `GetObjectTypeDetail`，并在响应对象与 action object_type_name 中体现补齐结果（`TestConceptRetrieval_GroupScopeCompletesReferencedObjects`）。
- 单元测试错误透传：mock BKN `SearchObjectTypes` 返回错误，断言 `conceptRetrieval` 直接返回该错误，且不调用后续 API（`TestConceptRetrieval_UnknownConceptGroupPropagatesError`）。
- 单元测试 metric direct / expansion recall 查询构造：确认请求体包含 `concept_groups`（已存在的 `TestSearchSchema_MetricQueriesCarryConceptGroups` 等用例）。
- 命令：`cd adp/context-loader/agent-retrieval/server && go test ./...`

## 影响分析

- **向后兼容**：调用方不传 `concept_groups` 时兼容现有行为；新增字段是增量能力。
- **依赖变化**：不新增依赖；依赖 BKN Backend 既有 typed search API 对 `concept_groups` 的支持，并依赖 BKN metrics 后续完成对 `concept_groups` 的功能过滤。
- **性能影响**：BKN 直查路径替代了原本的"全量 export + 本地过滤"，BKN 在分组范围内召回，候选集普遍更小，避免了对大网络全量加载的开销；代价是把分组语义实施完全交给 BKN，metrics 真正的分组过滤命中依赖 BKN 端的实现完成度。
- **风险**：
  - BKN 对未知分组返回 5xx 而非 4xx，HTTP 层会以 5xx 透传到 ContextLoader 上游；调用方可能误以为是 ContextLoader 内部错误。该错误码语义不符合"客户端引用了不存在的资源"的预期，需要 BKN 侧改为 4xx；在 BKN 调整前，ContextLoader 不做错误码翻译。
  - `metric_types` 的真实分组过滤命中由 BKN 实现完成度决定。BKN 端实施 metrics 分组过滤前，metrics 召回结果可能超出分组范围；ContextLoader 本期只验证请求透传。
  - 直查路径未做结果缓存；高频调用同一 `(kn_id, concept_groups, query)` 时不会复用 BKN 响应，必要时可在后续迭代中加入。

## 参考

- GitHub Issue: https://github.com/kweaver-ai/kweaver-core/issues/304
- `docs/apis/api_private/search_schema.yaml`
- BKN metrics API：`adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml`
