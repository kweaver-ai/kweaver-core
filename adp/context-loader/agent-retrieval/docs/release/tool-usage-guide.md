# context-loader 工具使用指南

面向开发人员的 context-loader 工具集说明文档，用于理解 context-loader 的定位、能力边界，以及在 Agent/服务中如何调用这些接口。

## 文档信息

| 字段 | 值 |
| :--- | :--- |
| 文档版本 | v1.6 |
| 适用版本 | context-loader v0.8.0 |
| 发布日期 | 2026-04-28 |
| 状态 | 正式发布 |

| 修订日期 | 修订说明 |
| :--- | :--- |
| 2026-04-28 | 更新为 context-loader `0.8.0`；`search_schema` 增加 `search_scope.concept_groups`，用于按 BKN 概念分组限定 object / relation / action schema 召回范围，并向 metric schema 检索透传分组条件 |
| 2026-04-23 | 更新为 context-loader `0.7.0`；本版 release 仅纳入 `issue-189` / `issue-234`；`search_schema` 的 HTTP `kn_id` 改为通过 body 传递，并补充 `metric_types` 发布口径 |
| 2026-04-16 | Schema 探索入口统一为 `search_schema`；MCP 不再暴露 `kn_search` / `kn_schema_search`；补充标准 / 兼容 / legacy 接口分层说明 |
| 2026-04-10 | 更新为 context-loader `0.6.0`，新增 `find_skills` 工具说明，并补充 `search/query/find/get` 四类工具语义 |
| 2026-03-26 | 根据 `docs/apis/api_private` OpenAPI 更新 6 个工具的依赖说明与参数配置 |
| 2026-01-04 | 首次发布 |

## 1. 什么是 context-loader

### 1.1 定位

context-loader 的目标不是直接回答用户问题，而是为 Agent 提供来自 BKN（业务知识网络）的高质量、最小且完备的上下文子集，让最终回答尽可能基于事实、降低幻觉。

### 1.2 工具语义分层

- `search_*`：探索发现 Schema 入口，解决“有哪些对象类 / 关系类 / 行动类”的问题
- `query_*`：对已知 Schema 做精确实例与事实查询，获取可审计结果
- `find_*`：在指定 `kn_id` 和业务上下文边界内发现候选资源，解决“当前场景下可考虑装配什么”的问题
- `get_*`：对已知目标做确定性获取、计算或信息物化

## 2. 能力边界

| 维度 | context-loader 负责 | context-loader 不负责 |
| :--- | :--- | :--- |
| 意图与推理 | 提供可用的检索工具与稳定输出 | 用户意图理解、规划与复杂推理 |
| 数据获取 | Schema 检索、实例检索、候选资源发现、逻辑属性解析、行动信息物化 | 最终自然语言答案生成 |
| 可靠性 | 提供确定性的结构化查询原子能力 | “自动把所有参数都推断出来”的完全自治 |

## 3. 快速开始

### 3.1 服务地址

默认服务地址：

```
http://agent-retrieval:30779
```

### 3.2 认证与通用 Header

多数接口要求在 Header 中携带以下认证信息：

| Header | 必填 | 说明 |
| :--- | :--- | :--- |
| `x-account-id` | 是（以接口定义为准） | 账户 ID |
| `x-account-type` | 是（以接口定义为准） | 账户类型（如 user/app/system/anonymous） |

### 3.3 最小调用示例：先查概念，再找入口实例

1）查概念（Schema，推荐：search_schema）：

```bash
curl -X POST "http://agent-retrieval:30779/api/agent-retrieval/in/v1/kn/search_schema" \
  -H "Content-Type: application/json" \
  -H "x-account-id: <your-account-id>" \
  -H "x-account-type: user" \
  -d '{
    "kn_id": "kn_medical",
    "query": "头晕吃什么药",
    "search_scope": {
      "concept_groups": ["medical_core"],
      "include_object_types": true,
      "include_relation_types": true,
      "include_action_types": true,
      "include_metric_types": true
    },
    "max_concepts": 10
  }'
```

2）精确查询实例（用 query_object_instance 定位入口实例）：

```bash
curl -X POST "http://agent-retrieval:30779/api/agent-retrieval/in/v1/kn/query_object_instance?kn_id=kn_medical&ot_id=disease" \
  -H "Content-Type: application/json" \
  -H "x-account-id: <your-account-id>" \
  -H "x-account-type: user" \
  -d '{
    "limit": 10,
    "condition": {
      "operation": "and",
      "sub_conditions": [
        { "field": "name", "operation": "like", "value_from": "const", "value": "高血压" }
      ]
    }
  }'
```

## 4. 用法概览（如何选择工具）

### 4.1 典型调用链

```
用户问题
      └─ Agent 规划
      ├─ 探索发现：search_schema
      ├─ 精确查询：query_object_instance / query_instance_subgraph
      ├─ 候选资源发现：find_skills（需要发现当前场景下可挂载的 Skill 时）
      ├─ 逻辑属性：get_logic_properties_values（需要动态参数时）
      └─ 行动信息物化：get_action_info（需要动态工具发现时）
```

### 4.2 工具总览

| 工具 | 核心作用 | 何时用 |
| :--- | :--- | :--- |
| `search_schema` | 统一的 Schema 探索入口 | 不确定有哪些对象类/关系类/动作类时 |
| `query_object_instance` | 单对象类实例过滤查询 | 已知对象类与过滤条件，要查列表时 |
| `query_instance_subgraph` | 沿关系路径拉取子图 | 需要跨关系找关联对象/多跳事实时 |
| `find_skills` | 在业务边界内发现 Skill 候选 | 已知 `kn_id`，想知道当前场景下可考虑装配哪些 Skill 时 |
| `get_logic_properties_values` | 逻辑属性解析（指标/算子） | 值需要按上下文动态计算时 |
| `get_action_info` | 动态工具发现（Function Call 定义） | 针对具体对象实例，想知道“能做什么动作”时 |

### 4.3 工具依赖（四类工具如何衔接）

- 精确查询依赖探索发现：
  - 结构化查询需要 `ot_id`（对象类 ID）与 Schema 信息（字段/主键/关系方向/动作绑定对象类）
  - `ot_id` 与 Schema 通常来自 `search_schema` 的返回（object_types / relation_types / action_types）
- 候选资源发现依赖 KN 边界与可选上下文：
  - `find_skills` 至少需要 `kn_id`
  - 若已通过探索发现确认对象类，可继续传 `object_type_id`
  - 若已通过精确查询获取实例，可继续传 `instance_identities`
  - `skill_query` 用于当前业务边界内过滤和排序，不替代 `search_*`
- 逻辑属性与行动召回依赖 Schema + 精确查询数据：
  - `get_logic_properties_values` 需要 `ot_id` + `_instance_identities`，其中数组元素应来自 `query_object_instance` 或 `query_instance_subgraph` 返回结果中的 `_instance_identity` 字段，以及逻辑属性定义（来自 Schema）
  - `get_action_info` 需要 `at_id`（来自 Schema）+ `_instance_identities`，其中数组元素应来自精确查询结果中的 `_instance_identity` 字段

## 5. 实现原理概览

### 5.1 Schema 与 Data 分层

- Schema：对象类/关系类/动作类定义（用于让 Agent“理解世界结构”）
- Data：对象实例与关联事实（用于让 Agent“拿到确定性证据”）

### 5.2 为什么要“先探索，再结构化，再发现候选资源”

- 探索发现用于降低盲区：先找到候选概念/入口实例
- 结构化查询用于保证可追溯：推理链的每一步都有确定输入与确定输出
- 候选资源发现用于补齐运行时装配：在既定业务边界内返回可考虑使用的 Skill 等资源，而不是暴露底层承载细节

## 6. 工具参考（Tool Reference）

本节仅给出开发接入时最常用的信息：用途、关键参数与最小示例。完整字段与响应结构以本目录下对应的 OpenAPI YAML 文件为准。

### 6.1 search_schema（统一 Schema 探索入口）

> 接口定义：[docs/apis/api_private/search_schema.yaml](../apis/api_private/search_schema.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/search_schema`
- 作用：根据 query 返回与之相关的 `object_types / relation_types / action_types / metric_types`
- 说明：这是新版本标准 Schema 探索接口，也是 MCP / Agent 唯一推荐入口。
- HTTP 口径：`kn_id` 通过 request body 传入，不再使用 `x-kn-id` Header。
- 概念分组：`search_scope.concept_groups` 可按 BKN 概念分组限定 Schema 召回范围；该范围作用于对象类、关系类和动作类，并向指标类检索透传分组条件，不作为实例数据过滤条件。

请求体（关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `query` | 是 | 用户自然语言查询 |
| `kn_id` | 是 | 知识网络 ID，通过 request body 传入 |
| `search_scope` | 否 | 是否包含对象类/关系类/动作类/指标类；`concept_groups` 用于限定 BKN 概念分组，其中对象类/关系类/动作类按分组生效，指标类透传分组条件；至少开启一种资源类型，默认全开 |
| `max_concepts` | 否 | 最大候选概念数量（默认 10） |
| `schema_brief` | 否 | 是否返回精简 Schema（默认 false） |
| `enable_rerank` | 否 | 是否启用关系类型 Rerank（默认 true） |

返回要点：

- `object_types / relation_types / action_types / metric_types`：Schema 结果；对象类/关系类/动作类可按概念分组限定，指标类会透传同一组分组条件
- 不返回实例数据，不返回 `nodes` / `message`
- 不传或传空 `concept_groups` 时不限定分组；分组语义实际由 BKN 完成，ContextLoader 直接调用 BKN 的 typed search 接口并把列表透传下去
- BKN 概念分组以对象类为直接边界：`object_types` 直接按组内对象类召回，`relation_types` 按 source / target 对象类均在组内推导，`action_types` 按绑定对象类在组内推导
- 当关系类或动作类结果引用了对象检索未命中的对象类时，ContextLoader 会补齐对应对象类详情，保证返回 Schema 引用完整
- `metric_types` 会携带同一组 `concept_groups` 调用 BKN metrics 检索；接口字段已支持，实际分组过滤依赖 BKN metrics 侧实现
- 当 BKN 判定请求分组均不存在时，当前会返回 5xx 错误（含 `error_details: "all concept group not found ..."`），`search_schema` 会直接向上透传该错误，而不是包装为成功的空结果；调用方可据此区分“分组不存在”与“分组合法但范围内无概念”

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 请求体参数（必填） | `"kn_medical"` |
| `query` | 模型生成 | 用户问题/关键词 | `模型生成` |
| `search_scope` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `max_concepts` | 固定值 | 最大概念数 | `10` |
| `schema_brief` | 固定值 | 默认返回相对完整 Schema | `false` |
| `enable_rerank` | 固定值 | 请求体参数（可选） | `true` |

### 6.2 兼容与 Legacy 说明

- `kn_search`
  - 兼容 HTTP 接口：`POST /api/agent-retrieval/in/v1/kn/kn_search`
  - 与 `search_schema` 共用收敛后的 Schema-only logic
  - 本次 `0.8.0` 的概念分组需求不改造该接口；新接入方如需按 BKN 概念分组探索 Schema，应使用 `search_schema.search_scope.concept_groups`
  - 旧字段可传，但不再恢复实例检索或 `nodes / message`
- `kn_schema_search`
  - legacy HTTP 接口：`POST /api/agent-retrieval/in/v1/kn/semantic-search`
  - 保持历史 `concepts[]` 输出形态
  - 不参与本次 shared logic 收敛
  - 虽然历史请求结构中已有 `concept_groups` 表达，但本次不改造该 legacy 链路；新接入方应使用 `search_schema`

### 6.3 query_object_instance（对象实例查询）

> 接口定义：[docs/apis/api_private/query_object_instance.yaml](../apis/api_private/query_object_instance.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/query_object_instance`
- Query 参数：
  - `kn_id`（必填）：业务知识网络 ID
  - `ot_id`（必填）：对象类 ID
  - `include_logic_params`（可选）：是否返回逻辑属性计算参数（默认 false）
- 作用：在指定对象类内，按过滤条件查询实例列表（支持分页）

请求体（FirstQueryWithSearchAfter，关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `limit` | 否 | 返回数量（默认 10，范围 1-100） |
| `condition` | 否 | 过滤条件（支持 and/or/比较/集合/like/match 等） |
| `sort` | 否 | 排序字段列表 |
| `need_total` | 否 | 是否返回总数 |
| `properties` | 否 | 指定返回的属性字段列表 |

Condition 规则要点：

- `value_from` 与 `value` 必须同时出现
- `value_from` 当前仅支持 `"const"`

示例：

```bash
curl -X POST "http://agent-retrieval:30779/api/agent-retrieval/in/v1/kn/query_object_instance?kn_id=kn_medical&ot_id=drug" \
  -H "Content-Type: application/json" \
  -H "x-account-id: <your-account-id>" \
  -H "x-account-type: user" \
  -d '{
    "limit": 10,
    "condition": {
      "operation": "and",
      "sub_conditions": [
        { "field": "name", "operation": "like", "value_from": "const", "value": "阿司匹林" }
      ]
    }
  }'
```

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | Query 参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `ot_id` | 模型生成 | Query 参数（对象类 ID） | `模型生成` |
| `include_logic_params` | 固定值 | Query 参数（可选） | `false` |
| `limit` | 固定值 | 请求体参数 | `10` |
| `condition` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `sort` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `need_total` | 固定值 | 请求体参数（可选） | `false` |
| `properties` | 模型生成 | 请求体参数（可选） | `模型生成` |

### 6.4 query_instance_subgraph（实例子图查询）

> 接口定义：[docs/apis/api_private/query_instance_subgraph.yaml](../apis/api_private/query_instance_subgraph.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/query_instance_subgraph`
- Query 参数：
  - `kn_id`（必填）：业务知识网络 ID
  - `include_logic_params`（可选）：是否返回逻辑属性计算参数（默认 false）
- 作用：基于关系路径查询对象子图；支持多条路径，每条路径返回独立子图

使用要点：

- 请求体必须提供 `relation_type_paths`（以接口定义为准），用于描述关系路径模板
- `relation_type_paths[].object_types` 与 `relation_type_paths[].relation_types` 的数组顺序必须严格对应；若为 n 跳路径，则 `object_types` 长度应为 n+1、`relation_types` 长度应为 n
- `relation_type_paths[].object_types[].condition` 为可选，但如传入则 `condition.operation` 必填
- Condition 结构与 query_object_instance 一致（同样需要 `value_from` + `value` 配对，且 `value_from` 当前仅支持 `const`）

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | Query 参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `include_logic_params` | 固定值 | Query 参数（可选） | `false` |
| `relation_type_paths` | 模型生成 | 请求体参数 | `模型生成` |

### 6.5 find_skills（Skill 候选发现）

> 接口定义：[docs/apis/api_private/find_skills.yaml](../apis/api_private/find_skills.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/find_skills`
- 作用：在指定知识网络和业务上下文边界内发现 Skill 候选，返回最小化 Skill 元数据列表

<a id="find-skills-quick-start"></a>

#### 6.5.1 快速开始

先确认能不能用：

- `find_skills` 不是升级 Context Loader 后自动可用的能力
- 在调用 `find_skills` 前，知识网络中必须已存在固定的 `skills` ObjectType
- 该 ObjectType 的运行时识别键必须是 `object_type_id = "skills"`
- `skills` ObjectType 不能由任意自定义对象类替代，它是 Skill 的固定承接面
- `skills` ObjectType 必须至少定义 `skill_id`、`name` 两个数据属性；`description` 可选
- `skills` ObjectType 下必须已有可见的 Skill 元数据实例
- 业务对象与 `skills` 之间必须已配置绑定关系

> 如果以上条件未满足，`find_skills` 在入口会直接返回错误，而不是继续执行召回。

最短使用路径：

- `kn_id` 必填
- `object_type_id` 必填，且必须存在于当前知识网络中
- 已定位到实例时，再传 `instance_identities`
- `skill_query` 只用于当前范围内过滤，不替代上下文定位

进一步阅读：

- 前提和限制：见[启用前提](#find-skills-prerequisites)
- 返回空结果时如何排查：见[空结果排查](#find-skills-empty-results)
- 为什么必须按固定方式配置：见[背景说明](#find-skills-background)

<a id="find-skills-prerequisites"></a>

#### 6.5.2 启用前提

`find_skills` 只负责运行时候选发现，不负责创建 Skill，也不负责维护业务绑定关系。Skill recall 是否真正可用，不只取决于 Context Loader 版本，还取决于建模侧是否已准备完成。

接入前至少确认以下三件事：

- 已存在固定的 `skills` ObjectType，且运行时识别键为 `object_type_id = "skills"`
- `skills` ObjectType 的数据属性定义至少包含 `skill_id`、`name`；`description` 可选
- `skills` ObjectType 下已有可见的 Skill 元数据实例，至少能返回 `skill_id`、`name`
- 业务对象与 `skills` 之间已配置绑定关系；否则对象类级和实例级召回会天然为空

<a id="find-skills-empty-results"></a>

#### 6.5.3 空结果排查

`find_skills` 返回空结果，不一定代表接口异常，更常见的是当前范围内没有可召回 Skill，或启用前提尚未满足。

| 场景 | 常见含义 | 建议动作 |
| :--- | :--- | :--- |
| 传 `object_type_id` 返回空结果 | 该对象类通常没有绑定 Skill | 确认该对象类与 `skills` 是否已配置绑定关系 |
| 传 `object_type_id + instance_identities` 返回空结果 | 该实例通常没有命中 Skill，或不存在实例级绑定 | 可先回退到对象类级查看候选 Skill 是否存在 |
| 传 `skill_query` 后返回空结果 | 当前过滤条件过严，或当前边界内本就没有匹配的 Skill | 放宽或去掉 `skill_query` 后重试 |
| 传了不存在的 `object_type_id` | 当前知识网络中不存在该对象类 | 先通过 `search_schema` 确认合法对象类，再重试 |
| 当前知识网络中缺少 `skills` ObjectType，或缺少 `skill_id` / `name` 属性 | `find_skills` 的基础契约未满足 | 先补齐固定的 `skills` ObjectType 及必要数据属性后，再调用 `find_skills` |

#### 6.5.4 参数与调用方式

请求体（关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `object_type_id` | 是 | 业务对象类型 ID；当前版本必须提供，且必须存在于当前知识网络中 |
| `instance_identities` | 否 | 对象实例标识列表；传入时缩小到实例级召回，且必须同时提供 `object_type_id` |
| `skill_query` | 否 | Skill 过滤词；仅在当前边界内做文本过滤和排序 |
| `top_k` | 否 | 返回的最大 Skill 数量，默认 10，最大 20 |

调用要点：

- 当前版本暂不开放网络级召回
- 传 `kn_id + object_type_id`：对象类级召回
- 传 `kn_id + object_type_id + instance_identities`：实例级召回
- `skill_query` 不替代 `search_schema`；若调用方尚未明确对象类或实例，应先使用 `search_*` / `query_*`

返回要点：

- `entries`：Skill 候选列表
- 每个条目只返回 `skill_id`、`name`、`description`
- 无匹配时返回空数组
- `message`：空结果说明信息；仅当 `entries` 为空且接口返回成功时出现，用于解释当前为什么没有结果以及下一步建议

<a id="find-skills-background"></a>

#### 6.5.5 背景说明

下面这些不是请求参数要求，而是 Skill recall 能否真正生效的底层约束。

| 背景项 | 用户需要知道什么 | 对使用 `find_skills` 的影响 |
| :--- | :--- | :--- |
| 固定模板 / 固定承接面 | Skill 不是任意对象类都能被召回，系统约定通过固定的 `skills` ObjectType 承接 | 如果没有这个固定承接面，`find_skills` 无法稳定识别 Skill 数据 |
| 基础契约字段 | `find_skills` 运行时依赖 `skills` ObjectType 至少定义 `skill_id`、`name` 两个数据属性；`description` 只是可选补充信息 | 如果缺少 `skill_id` 或 `name`，`find_skills` 会在入口直接报错，而不是继续召回 |
| 共享只读视图 | Skill 元数据由上游统一管理，BKN 默认承接的是只读视图，而不是每个知识网络各自维护一份副本 | 新增或变更 Skill 后，召回是否可见取决于该视图是否已同步可用 |
| 运行时识别键 | 当前运行时固定通过 `object_type_id = "skills"` 识别 Skill ObjectType | 不是“同名即可”，必须满足固定识别键，否则 `find_skills` 不会把它当作 Skill 承接面 |

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 请求体参数 | `"kn_legal"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `object_type_id` | 模型生成 | 请求体参数（必填） | `模型生成` |
| `instance_identities` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `skill_query` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `top_k` | 固定值 | 请求体参数（可选） | `5` 或 `10` |

### 6.6 get_logic_properties_values（逻辑属性解析）

> 接口定义：[docs/apis/api_private/get_logic_properties_values.yaml](../apis/api_private/get_logic_properties_values.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/logic-property-resolver`
- 作用：针对某对象类的一个或多个实例，批量计算/查询逻辑属性（metric/operator）

请求体（关键字段）：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `ot_id` | 是 | 对象类 ID |
| `query` | 是 | 用户原始问题（用于生成 dynamic_params） |
| `_instance_identities` | 是 | 实例标识数组（支持批量）；应从上游实例结果的 `_instance_identity` 字段提取后按顺序组装 |
| `properties` | 是 | 逻辑属性名列表（metric/operator） |
| `additional_context` | 否 | 推荐传结构化 JSON 字符串，补充时间/对象上下文等 |
| `options` | 否 | 高级选项；当前主要支持 `return_debug` |

返回形态：

- 成功：返回 `datas`
- 缺参：返回 `error_code=MISSING_INPUT_PARAMS` 与 `missing` 清单，按 hint 补充 query 或 additional_context 后重试

缺参示例（节选）：

```json
{
  "error_code": "MISSING_INPUT_PARAMS",
  "message": "dynamic_params 缺少必需的 input 参数",
  "missing": [
    {
      "property": "approved_drug_count",
      "params": [
        { "name": "start", "hint": "在 additional_context 中补充时间范围，或在 query 中明确时间信息" }
      ]
    }
  ],
  "trace_id": "3f5d6c1c-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 请求体参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `ot_id` | 模型生成 | 请求体参数 | `模型生成` |
| `query` | 模型生成 | 请求体参数 | `模型生成` |
| `_instance_identities` | 模型生成 | 请求体参数 | `模型生成` |
| `properties` | 模型生成 | 请求体参数 | `模型生成` |
| `additional_context` | 模型生成 | 请求体参数（可选） | `模型生成` |
| `options.return_debug` | 固定值/模型生成 | 请求体参数（可选） | `false` |

### 6.7 get_action_info（行动信息召回 / 动态工具发现）

> 接口定义：[docs/apis/api_private/get_action_info.yaml](../apis/api_private/get_action_info.yaml)

- API：`POST /api/agent-retrieval/in/v1/kn/get_action_info`
- 作用：针对对象实例，召回可执行行动，并转换为 OpenAI Function Call 规范的工具定义列表

请求体：

| 字段 | 必填 | 说明 |
| :--- | :--- | :--- |
| `kn_id` | 是 | 业务知识网络 ID |
| `at_id` | 是 | 行动类型 ID |
| `_instance_identities` | 否 | 对象实例标识列表；每个元素为主键键值对，应从上游实例结果的 `_instance_identity` 字段提取，不可臆造 |

返回要点：

- `_dynamic_tools`：动态工具列表（每个工具包含 name/description/parameters/api_url/fixed_params 等）

当前版本限制：

- 仅支持 type=tool 的行动源（MCP 下版本支持）
- 仅处理 actions[0]
- 不处理 dynamic_params（由 LLM 侧生成）

Data Agent 配置（建议）：

| 配置项 | 推荐类型 | 说明 | 示例 |
| :--- | :--- | :--- | :--- |
| `x-account-id` | 应用变量 | Header 参数 | `header.x-account-id` |
| `x-account-type` | 固定值/应用变量 | Header 参数 | `user` 或 `header.x-account-type` |
| `kn_id` | 固定值/应用变量 | 请求体参数 | `"kn_medical"` 或 `self_config.data_source.knowledge_network[0].knowledge_network_id` |
| `at_id` | 模型生成 | 请求体参数 | `模型生成` |
| `_instance_identities` | 模型生成 | 请求体参数（可选） | `模型生成` |

## 7. 集成场景与最佳实践

### 7.1 场景：从问题到可审计事实链

1）探索概念：用 `search_schema` 确认对象类/关系类
2）精确定位实例：用 `query_object_instance`（单类过滤）或 `query_instance_subgraph`（跨关系/多跳）获取入口实例与事实  
3）发现候选 Skill：用 `find_skills` 获取当前知识网络和对象上下文下可考虑挂载的 Skill  
4）补充指标：用 `get_logic_properties_values` 获取逻辑属性值（必要时补 additional_context）  
5）动态动作：用 `get_action_info` 获取与实例关联的可执行行动

## 8. 附录

### 8.1 本目录 OpenAPI 定义文件

- `search_schema.yaml`
- `kn_search.yaml`
- `kn_schema_search.yaml`
- `query_object_instance.yaml`
- `query_instance_subgraph.yaml`
- `find_skills.yaml`
- `get_logic_properties_values.yaml`
- `get_action_info.yaml`
