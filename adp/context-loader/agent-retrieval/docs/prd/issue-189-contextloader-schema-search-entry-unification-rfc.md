# RFC: ContextLoader Schema Search 入口统一与升级兼容

> 状态: Frozen  
> 关联 Issue: [#189](https://github.com/kweaver-ai/kweaver-core/issues/189)  
> 负责人: Cheng.cao  
> 更新时间: 2026-04-15  
> 文档定位: 产品/接口层 RFC，明确 Schema Search 入口统一方案、功能收敛方向、兼容策略与交付文档规则。当前版本结论已冻结，后续实现落地以设计文档为准。

---

## 1. 背景与问题

当前 ContextLoader 在 Schema 探索场景同时存在两个入口：

- `kn_schema_search`
- `kn_search`

两者都承担“根据 query 探索对象类、关系类、动作类”的职责，但在命名、输入结构、输出结构和推荐使用方式上并不一致，带来以下问题：

- Agent / 调用方需要先判断该选哪个入口，增加接入成本。
- MCP 工具层缺少统一命名规则；当前工具集已有 `find_*`、`query_*`、`get_*`，而 search 相关仍保留历史 `kn_*` 风格。
- 发布文档、工具集导出和 HTTP OpenAPI 的主叙事不一致，用户难以判断“规范入口”和“兼容入口”。
- `kn_search` 还混入了实例检索相关语义，使 Schema 探索边界不够稳定。

因此，issue #189 要解决的不是单纯改名，而是：

- 统一 Schema Search 的 MCP / Agent 入口
- 收敛 `search_schema` 应交付的功能边界
- 明确用户如何从旧版工具集升级到新版工具集

---

## 2. 目标与边界

### 2.1 目标

- 对 MCP / Agent 暴露单一 Schema Search 规范入口。
- 降低新接入方的工具选择成本。
- 保持旧版 HTTP 接口和旧版工具集在过渡期内可继续工作。
- 统一发布文档、工具集导出和 OpenAPI 的主叙事。
- 将 `search_schema` 收敛为稳定的 Schema 探索工具，而不是继续沿用历史复合语义。

### 2.2 非目标

- 本轮不强制要求 `kn_schema_search` 与 `kn_search` 在底层实现层完成真实收敛。
- 本轮不修改现有 legacy HTTP 路径的入参与返回契约。
- 本轮不重构 `query_*`、`find_*`、`get_*` 的能力边界。
- 本轮不处理 Schema Search 之外的工具命名重构。

### 2.3 前提约束

- MCP 工具层不考虑旧工具名兼容，直接切换为新规范名。
- HTTP 接口保留兼容，避免破坏已有调用链。
- 旧版工具集在新版本环境中应继续可用。
- 发布文档必须优先解决“用户如何升级”的问题。

---

## 3. 用户、场景与产品定位

### 3.1 用户角色

| 角色 | 关注点 |
|------|--------|
| Agent / 应用编排者 | 需要一个稳定、单一的 Schema 探索入口，用于决定下一步调用哪个下游工具 |
| 工具集交付使用者 | 需要理解新版工具集与旧版工具集的关系，判断是否需要升级 |
| HTTP 集成调用方 | 更关注接口是否兼容，不希望因为 MCP 工具名变化被迫迁移 HTTP 路径 |

### 3.2 典型场景

#### 场景 1：首次探索 Schema

调用方尚不清楚问题落在哪个对象类、关系类或动作类上。  
此时需要先通过 `search_schema` 返回候选 Schema 概念，再决定是否进入下游精确查询。

#### 场景 2：为后续查询提供上下文

调用方已知需要继续查询，但尚未确定对象类或关系路径。  
此时 `search_schema` 需要提供足够清晰的 Schema 结果，帮助后续调用 `query_object_instance` 或 `query_instance_subgraph`。

#### 场景 3：为技能发现和行动发现提供入口

调用方需要继续调用 `find_skills` 或 `get_action_info`，但缺少合法的对象类或动作类上下文。  
此时 `search_schema` 需要返回结构化 Schema 结果，供后续工具直接消费。

### 3.3 产品定位

`search_schema` 是 ContextLoader 在 MCP / Agent 层暴露的唯一 Schema 探索入口。

它的职责不是回答业务事实，也不是返回实例数据，而是：

- 识别当前问题相关的 Schema 概念
- 为后续精确查询、技能发现、行动发现提供合法上下文
- 以最小但足够的方式把 Schema 信息交付给 Agent

从工具语义分层看：

| 工具类别 | 解决的问题 |
|---------|------------|
| `search_schema` | 有哪些对象类、关系类、动作类值得继续探索 |
| `query_*` | 已知 Schema 后，如何精确查询实例与事实 |
| `find_*` | 在业务边界内发现候选资源 |
| `get_*` | 获取动作信息或逻辑属性值 |

因此，`search_schema` 必须是纯粹的 Schema 探索工具，而不是综合搜索工具。

---

## 4. 统一方案

### 4.1 合并结论

本 RFC 采用“统一 MCP 规范名，新增标准 HTTP 契约，保留兼容与 legacy HTTP 路由”的方案。

#### MCP / Agent 规范名

Schema Search 在 MCP / Agent 层的唯一规范工具名定义为：

- `search_schema`

以下名称不再作为新版 MCP / Agent 工具名继续暴露：

- `kn_schema_search`
- `kn_search`

#### HTTP 契约边界

本轮在保留旧路径的前提下，新增标准 HTTP route，并明确三类角色：

| 类型 | 路径 | 角色 |
|------|------|------|
| 标准 HTTP 契约 | `POST /api/agent-retrieval/in/v1/kn/search_schema` | `search_schema` 的标准 HTTP route |
| 兼容 HTTP 契约 | `POST /api/agent-retrieval/in/v1/kn/kn_search` | 兼容旧调用路径，但与 `search_schema` 共用收敛后的 Schema Search logic |
| Legacy HTTP 契约 | `POST /api/agent-retrieval/in/v1/kn/semantic-search` | 保持现状的历史接口，不进入本次 shared logic 收敛范围 |

同时明确：

- `search_schema` 同时作为 MCP / Agent 规范工具名和新版标准 HTTP 接口描述存在
- `kn_search` 继续保留 HTTP 路径，但其共享能力边界与 `search_schema` 一致
- `kn_schema_search` 继续保留 legacy HTTP 路径与现有语义，不进入本次 shared logic 收敛范围

#### 为什么采用该方案

- 如果只在文档层统一，Agent 仍需面对两个重叠入口，问题没有真正解决。
- 如果 MCP 继续保留旧工具名兼容，产品层“统一入口”目标不成立。
- 仅让 `search_schema` 和 `kn_search` 共用收敛后的 logic，可以在统一新标准接口的同时平滑承接现有 `kn_search` 调用链。
- 保留 `kn_schema_search` 原状不动，可以避免本期同时触碰两条历史链路，降低兼容风险。

### 4.2 功能收敛结论

这次不仅是工具合并，也是一次功能边界调整。

`search_schema` 应交付的能力：

- 围绕自然语言问题返回相关的 `object_types`、`relation_types`、`action_types`
- 支持按概念类型控制探索范围
- 默认返回适合 Agent 消费的精简 Schema
- 输出结果可直接为 `query_*`、`find_*`、`get_*` 服务

`search_schema` 不应继续承担的能力：

- 不返回实例数据
- 不承担实例检索
- 不暴露底层算法和调参结构
- 不替代 `query_object_instance`、`query_instance_subgraph`、`find_skills`、`get_action_info`

上述能力收敛对 `search_schema` 和 `kn_search` 共用的 shared logic 同时生效。  
因此，`search_schema` 不是把 `kn_search` 原样更名，而是将 `kn_schema_search` 的简单产品语义与 `kn_search` 的有效 Schema 输出形态做统一收口；同时 `kn_search` 兼容入口也不再恢复实例检索能力。  
`kn_schema_search` 继续保持现状，不进入本次收敛范围。

### 4.3 V1 产品参数

结合现有 `kn_schema_search` 与 `kn_search` 的能力边界，`search_schema` 的 V1 建议采用精简参数集，不直接继承历史工具的全部输入结构。

#### 建议保留的参数

| 参数名 | 是否必填 | 产品语义 |
|------|----------|----------|
| `query` | 是 | 用户问题或关键词，用于触发 Schema 探索 |
| `search_scope` | 否 | 控制本次是否探索对象类、关系类、动作类，默认全开 |
| `max_concepts` | 否 | 控制本次返回的 Schema 候选规模上限，默认 `10` |
| `schema_brief` | 否 | 控制返回精简 Schema 还是相对完整的 Schema 详情，默认 `false` |
| `enable_rerank` | 否 | 控制是否启用重排能力，属于高级可选参数，默认 `true` |

#### V1 默认行为

当前版本的默认行为定义如下：

- `search_scope` 默认同时包含对象类、关系类、动作类
- `max_concepts` 默认值为 `10`
- `schema_brief` 默认值为 `false`
- `enable_rerank` 默认值为 `true`

#### 关于数量参数命名的结论

V1 建议采用：

- `max_concepts`

不建议采用：

- `top_k`
- `limit`

原因如下：

- `top_k` 更接近底层检索实现语义，容易把 `search_schema` 再次做成“检索配置接口”。
- `limit` 语义过于泛化，不足以清晰表达“Schema 候选上限”的产品含义。
- `max_concepts` 与 `search_schema` 的产品定位更一致，更利于 Agent 和交付用户理解。
- 当前 `max_concepts` 的默认值与内部 `top_k=10` 对齐，但对外规范名保持 `max_concepts`，不将 `top_k` 暴露为产品参数。

#### 不保留的历史参数形态

以下参数或结构不建议继续进入 `search_schema` 的 V1 产品契约：

- `retrieval_config`
- `concept_retrieval`
- `only_schema`
- `rerank_action`

原因如下：

- `retrieval_config` 与 `concept_retrieval` 属于底层能力组织方式，不适合作为 Agent 面向的产品契约。
- `only_schema` 在 `search_schema` 语义下已是默认前提，不应继续要求调用方显式理解。
- `rerank_action` 更接近内部策略选择，不适合作为产品主路径参数暴露。

#### V1 暂不纳入的能力参数

以下能力虽在历史工具中存在价值，但当前版本建议暂不纳入 `search_schema` V1：

- `include_sample_data`

原因如下：

- 样例数据会明显增加返回体重量。
- 对部分 Agent 尤其是小参数量模型而言，容易把样例内容误判为最终业务事实。
- 当前版本的首要目标是统一 Schema 探索入口，而不是把所有增强能力一次性收口。

### 4.4 V1 输出结构

`search_schema` 的输出建议采用分组结构，而不是历史 `concepts[]` 扁平结构：

- `object_types`
- `relation_types`
- `action_types`

原因如下：

- 更符合 Agent 后续进入 `query_*`、`find_*`、`get_*` 工具时的消费方式。
- 减少调用方从扁平结果中再次做概念分类的成本。
- 与 `search_schema` 的产品定位更一致，即“先识别 Schema，再进入下游专用工具”。

---

## 5. 兼容、升级与文档交付

### 5.1 用户升级路径

| 用户类型 | 是否必须升级 | 升级动作 |
|---------|--------------|----------|
| MCP / Agent 用户 | 是 | 将 `kn_schema_search`、`kn_search` 替换为 `search_schema`，并更新 prompt、tool allowlist、脚本与路由配置 |
| 旧版工具集用户 | 否 | 可继续运行，但应在文档中标记为“兼容保留，不再推荐新增接入” |
| 直接 HTTP 旧用户 | 否 | 可继续调用 `/kn/kn_search` 或 `/kn/semantic-search`；其中 `/kn/kn_search` 的共享能力将收敛为 Schema-only |
| 新接入用户 | 是 | 直接使用 `search_schema`，HTTP 场景优先使用 `/kn/search_schema` |

### 5.2 `docs/release`

`docs/release` 面向产品交付和工具使用者，只讲规范入口，不并列讲历史入口。

需要做以下调整：

- `overview.md`
  - 明确 `search_schema` 是唯一 Schema Search 规范工具名
  - 明确 `kn_schema_search` / `kn_search` 不再作为 MCP / Agent 工具名暴露
  - 明确标准 HTTP route 为 `/kn/search_schema`
  - 明确 `/kn/kn_search` 为兼容 route，`/kn/semantic-search` 为 legacy route
- `tool-usage-guide.md`
  - 将 `kn_schema_search` 与 `kn_search` 的主叙事收敛为 `search_schema`
  - 明确标准 HTTP、兼容 HTTP、legacy HTTP 三者关系
  - 增加升级说明
- `toolset/context_loader_toolset.adp`
  - 新版发布工具集只保留 `search_schema`

### 5.3 `docs/apis`

`docs/apis` 面向 HTTP 契约使用者，准确反映三条真实 HTTP 路径及其角色。

建议如下：

- 新增 `search_schema.yaml`
  - 标记为 `search_schema` 的标准 HTTP 契约
- 保留 `kn_search.yaml`
  - 标记为兼容 HTTP 契约
  - 明确其与 `search_schema` 共用收敛后的 shared logic
- 保留 `kn_schema_search.yaml`
  - 标记为 legacy endpoint
  - 保持现有契约说明，不进入本次 shared logic 收敛范围

三个 YAML 的 `summary` / `description` 需要明确：

- `search_schema.yaml` 对应的 MCP / Agent 规范工具名为 `search_schema`
- `/kn/kn_search` 为兼容入口，与 `search_schema` 共享收敛后的 logic
- `/kn/semantic-search` 为 legacy 入口，当前保持现状

### 5.4 影响范围

本 RFC 直接影响：

- MCP 工具注册与 schema
- `docs/release/overview.md`
- `docs/release/tool-usage-guide.md`
- `docs/release/toolset/context_loader_toolset.adp`
- `docs/apis/api_private/search_schema.yaml`
- `docs/apis/api_private/kn_search.yaml`
- `docs/apis/api_private/kn_schema_search.yaml`

本 RFC 不直接影响：

- `query_*`
- `find_*`
- `get_*`
- `kn_schema_search` 的现有逻辑与输出结构
- `kn_schema_search` / `kn_search` 在底层实现层的真实合并

---

## 6. 风险与验收

### 6.1 风险与应对

| 风险 | 应对 |
|------|------|
| 用户误以为所有 HTTP 路径都必须迁移 | 在发布文档中明确区分标准 HTTP、兼容 HTTP 和 legacy HTTP 三类角色 |
| 旧版工具集用户不清楚是否需要升级 | 在 `overview.md` 与 `tool-usage-guide.md` 中加入升级矩阵 |
| 文档继续把主入口与兼容入口平级展示 | `docs/release` 只保留一个主叙事：`search_schema` |
| OpenAPI 文件名与 MCP 名不一致引发困惑 | 在 OpenAPI 描述中明确“规范工具名”与“HTTP 路径角色” |

### 6.2 验收清单

- MCP / Agent 最终只暴露一个 Schema Search 工具名：`search_schema`
- 新版发布工具集只包含 `search_schema`
- `docs/release` 不再并列介绍 `kn_schema_search` 与 `kn_search`
- `docs/apis` 中明确：
  - `search_schema.yaml` 为标准 HTTP 契约
  - `kn_search.yaml` 为兼容 HTTP 契约
  - `kn_schema_search.yaml` 为 legacy HTTP 契约
- `search_schema` 的 V1 参数集明确收敛为：
  - `query`
  - `search_scope`
  - `max_concepts`
  - `schema_brief`
  - `enable_rerank`
- `search_schema` 的 V1 输出结构明确收敛为：
  - `object_types`
  - `relation_types`
  - `action_types`
- 用户能够从交付文档中明确判断自己是否需要升级，以及需要升级 MCP、工具集还是 HTTP

### 6.3 失败条件

- MCP 仍继续暴露多个 Schema Search 工具名
- `docs/release` 仍把新旧入口作为平级主入口描述
- OpenAPI 未标识主契约与兼容契约
- 文档未明确 `search_schema` 与 `kn_search` 共用 shared logic，而 `kn_schema_search` 保持现状
- 升级文档没有明确说明旧版工具集在新版本环境中的兼容关系
- `search_schema` 的产品契约仍混入实例检索或底层调参语义

---

## 7. 后续演进

本 RFC 解决的是“产品/接口层统一入口与迁移表达”问题，不直接解决“底层实现是否真实合并”问题。

后续可单独评估：

- 是否将 `kn_schema_search` 与 `kn_search` 在底层实现层完全收敛
- 是否在后续版本下线 `/kn/semantic-search`
- 是否在后续版本下线 `/kn/kn_search` 兼容 route
- 是否为所有 ContextLoader 工具建立统一的“规范工具名 -> HTTP 契约”索引规范
