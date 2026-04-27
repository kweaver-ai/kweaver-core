# 🧩 PRD: ContextLoader `search_schema` 支持 Metric Schema 召回

> 状态: In Review
> 负责人: 待确认
> 更新时间: 2026-04-20

---

## 📌 1. 背景（Background）

- 当前现状：
  - `search_schema` 已作为 ContextLoader 的统一 Schema 探索入口，支持 `object_types`、`relation_types`、`action_types` 三类资源召回。
  - BKN 侧已引入原生指标能力，指标成为需要被 Agent 探索的新一类 schema 资源。

- 存在问题：
  - Agent 在第一跳 schema 探索阶段无法通过 `search_schema` 发现 metric 相关资源。
  - 调用方若想引入 metric，只能绕开统一入口，破坏 `search_schema` 的产品定位。
  - 当前接口和交付物均未定义 metric schema 的默认行为、scope 语义和验收口径。

- 触发原因 / 业务背景：
  - ContextLoader 需要在现有统一入口上扩展 metric schema 召回能力，使 Agent / MCP / HTTP 调用方能在通用 schema 探索场景下发现 metric 资源。

---

## 🎯 2. 目标（Objectives）

- 业务目标：
  - 让 ContextLoader 在统一 `search_schema` 入口下支持第四类 schema 资源 `metric_types`，避免调用方为 metric 另行寻找探索入口。

- 产品目标：
  - `search_schema` 支持 `metric_types` 的能力型交付，并保持现有三类资源行为不退化。
  - `search_scope` 新增 `include_metric_types`，且默认行为与现有三类资源一致。
  - MCP、HTTP、发布文档和 toolset 对 metric schema 召回形成一致交付。

---

## 👤 3. 用户与场景（Users & Scenarios）

### 3.1 用户角色

| 角色 | 描述 |
|------|------|
| Agent | 在第一跳探索时，需要同时发现对象、关系、动作、指标四类 schema 候选。 |
| MCP / HTTP 调用方 | 需要通过统一接口完成 schema 探索，而不是为 metric 单独接入新工具。 |
| ContextLoader 能力接入方 | 需要稳定的接口契约、scope 语义和发布说明，便于持续集成。 |

---

### 3.2 用户故事（User Story）

- 作为 Agent，我希望在调用 `search_schema` 时能够同时探索 metric schema，从而为后续工具调用提供更完整的 schema 上下文。
- 作为 MCP / HTTP 调用方，我希望继续使用统一的 `search_schema` 入口接入 metric schema，而不是新增独立搜索接口，从而降低接入复杂度。
- 作为 ContextLoader 能力接入方，我希望 metric schema 的默认行为、scope 语义和交付边界是明确的，从而可以稳定验收和发布。

---

### 3.3 使用场景

- 场景1：Agent 基于用户 query 做第一跳 schema 探索，希望同时发现 object / relation / action / metric 四类候选资源。
- 场景2：调用方希望只返回部分资源类型，但仍允许系统在内部使用其他资源线索辅助召回。
- 场景3：发布新版 MCP / HTTP 契约后，调用方需要在不新增独立 metric 搜索工具的前提下接收 metric schema 能力。

---

## 📦 4. 需求范围（Scope）

### ✅ In Scope

- 在 `search_schema` 中新增第四类资源 `metric_types`
- 在 `search_scope` 中新增 `include_metric_types`
- 保持返回结构按资源类型独立返回，新增顶层 `metric_types`
- `schema_brief` 对 `metric_types` 保持参数语义兼容；V1 阶段 `schema_brief=true` 与 `schema_brief=false` 暂返回同一字段集
- 支持两条 metric 召回路径：
  - 基于 query 的直接召回
  - 基于系统内部 object 候选的关联扩展召回
- 对两条路径命中的重复 metric 进行去重后统一返回
- mixed recall 仅作用于 `metric_types`，本期不反向影响 `object_types`、`relation_types`、`action_types`
- 明确 `search_scope` 的默认行为、关闭语义和报错规则
- 同步更新 MCP 工具契约、HTTP 接口契约、发布文档和 toolset

### ❌ Out of Scope

- metric 数据查询
- metric dry-run / evaluate
- BKN metric CRUD 能力接入
- 新增独立的 metric 搜索工具
- 在本 PRD 中定稿 `metric_types` 的详细字段结构
- 在本 PRD 中承诺 metric 召回排序质量、命中率或效果指标

---

## ⚙️ 5. 功能需求（Functional Requirements）

### 5.1 功能结构

    search_schema
    ├── object_types
    ├── relation_types
    ├── action_types
    └── metric_types

---

### 5.2 详细功能

#### 【FR-1】新增 Metric Schema 资源类型

**描述：**  
`search_schema` 在现有三类 schema 资源基础上，新增第四类资源 `metric_types`。

**用户价值：**  
调用方可以继续通过统一入口完成 schema 探索，不需要为 metric 单独接入新搜索工具。

**交互流程：**
1. 调用方基于 query 调用 `search_schema`
2. 系统执行现有 schema 探索流程，并补充 metric schema 召回
3. 响应按资源类型独立返回，其中新增 `metric_types`

**业务规则：**
- `metric_types` 是 `search_schema` 的平级资源类型。
- `metric_types` 虽独立返回，但其产品语义不脱离绑定的 `object_type`。
- `metric_types` 的详细字段结构在后续设计中确认，本 PRD 只定义能力与边界。

**边界条件：**
- 当知识网络中不存在相关 metric schema 时，允许 `metric_types` 为空。
- 本需求不要求在 `search_schema` 中返回 metric 查询结果。

**异常处理：**
- 若 metric 召回无结果，不影响其他已开启资源类型的返回。
- 若下游暂时无法提供 metric schema 数据，接口应保持现有错误处理和响应规范。

---

#### 【FR-2】扩展 `search_scope` 语义

**描述：**  
在 `search_scope` 中新增 `include_metric_types`，并与现有三类资源保持一致的默认和显式关闭语义。

**用户价值：**  
调用方可以像控制其他 schema 资源一样控制 metric schema 的返回范围。

**交互流程：**
1. 调用方传入或省略 `search_scope`
2. 系统解析四类资源开关
3. 根据开关决定响应中返回哪些资源类型

**业务规则：**
- `search_scope` 不传时，四类资源默认均视为开启。
- `include_metric_types=false` 时，响应中不返回 `metric_types`。
- 任一资源类型开关显式为 `false` 时，响应中不返回对应资源类型数据。
- 四个资源开关不能同时为 `false`；若同时为 `false`，接口报错。

**边界条件：**
- `search_scope` 是输出约束，不等价于完全禁止系统内部使用相关线索。
- 关闭某类资源仅代表该类型不出现在响应中。

**异常处理：**
- 当四类开关均为 `false` 时，返回明确错误，提示至少开启一种资源类型。

---

#### 【FR-3】支持双路径召回并统一返回

**描述：**  
`metric_types` 的召回同时支持 query 直接召回与基于系统内部 object 候选的关联扩展召回，最终统一去重后返回。

**用户价值：**  
调用方既能直接探索 metric，也能在 object 线索下补全相关 metric 候选，提高探索完整度。

**交互流程：**
1. 系统基于 query 执行 metric 直接召回
2. 系统基于内部 object 候选执行关联 metric 扩展召回
3. 系统对两条路径的结果进行合并与去重
4. 系统将结果统一输出到顶层 `metric_types`

**业务规则：**
- 两条召回路径同时有效，不要求调用方选择模式。
- 两条路径命中的相同 metric 不重复返回。
- object 扩展路径使用系统内部 object 候选，不以 `object_types` 是否最终输出为前提。
- 当 `include_metric_types=true` 且 `include_object_types=false` 时，系统仍允许内部借助 object 候选辅助 metric 召回，但响应中不返回 `object_types`。
- mixed recall 本期仅用于增强 `metric_types` 召回，不要求反向补齐 `object_types`，也不改变 `relation_types`、`action_types` 的返回逻辑。

**边界条件：**
- 本需求不要求对两条路径的排序优先级给出产品承诺。
- 本需求不要求区分每个 metric 命中来源的返回结构。
- 本需求不要求因返回 `metric_types` 而同次补齐其绑定 `object_types`。

**异常处理：**
- 某一路径无结果时，不影响另一条路径产出的 metric 返回。

---

#### 【FR-4】交付物同步更新

**描述：**  
`metric schema` 召回能力不仅体现在服务逻辑上，也必须体现在接口契约和交付物上。

**用户价值：**  
调用方可以从工具定义、HTTP 契约和发布文档中获得一致认知，避免“能力存在但文档不可用”。

**交互流程：**
1. 产品能力定义完成
2. MCP 和 HTTP 契约同步更新
3. 发布文档和 toolset 同步发布

**业务规则：**
- MCP 工具契约需要体现 `metric_types` 和 `include_metric_types`
- HTTP 接口契约需要体现 `metric_types` 和 `include_metric_types`
- 发布文档和 toolset 需要反映 `search_schema` 已支持 metric schema 召回

**边界条件：**
- 本需求不要求新增独立 metric 工具或单独文档入口

**异常处理：**
- 若任一交付物未同步更新，则视为本需求未完整交付

---

### 5.3 接口整体形态示意

以下示意仅用于表达 `search_schema` 的产品形态，不作为最终 OpenAPI 或字段冻结版本。

**请求示例：**

```json
{
  "query": "查看 pod cpu 使用率相关指标",
  "search_scope": {
    "include_object_types": true,
    "include_relation_types": false,
    "include_action_types": false,
    "include_metric_types": true
  },
  "max_concepts": 10,
  "schema_brief": true,
  "enable_rerank": true
}
```

**响应示例：**

```json
{
  "object_types": [],
  "relation_types": [],
  "action_types": [],
  "metric_types": [
    {
      "...": "待设计阶段确认"
    }
  ]
}
```

**`metric_types` 字段设计原则：**

- 应能够标识 metric 的身份
- 应能够表达 metric 的可读名称和基础说明
- 应能够体现 metric 与绑定 `object_type` 的关系
- 应能够兼容 `schema_brief` 的参数语义；V1 阶段可先复用同一字段集，后续再根据场景迭代精简
- 不应将 BKN `MetricDefinition` 的重字段原样透传为 `search_schema` 的最终产品输出

---

## 🔄 6. 用户流程（User Flow）

    调用方输入 query → 调用 search_schema → 系统进行四类 schema 探索 → 根据 search_scope 过滤输出 → 返回 object_types / relation_types / action_types / metric_types

---

## 🎨 7. 交互与体验（UX/UI）

### 7.1 页面 / 模块
- 本需求无新增页面
- 本需求无新增前端模块

### 7.2 交互规则
- 调用方式沿用现有 `search_schema`
- 交互重点为接口契约、工具定义和返回形态一致性
- metric schema 的产品体验遵循“统一入口、分桶返回、默认可用”

---

## 🚀 8. 非功能需求（Non-functional Requirements）

### 8.1 性能
- 本 PRD 不新增独立性能指标，保持 `search_schema` 现有性能基线

### 8.2 可用性
- metric schema 扩展不应导致现有三类资源能力失效或退化

### 8.3 安全
- 延续现有 `search_schema` 的鉴权与访问控制策略

### 8.4 可观测性
- metric schema 召回能力应纳入现有日志、错误处理和基础观测体系

---

## 📊 9. 埋点与分析（Analytics）

| 事件 | 目的 |
|------|------|
| 本期不新增独立埋点 | 本需求以能力型交付为主，不在 PRD 中定义质量分析埋点 |

---

## ⚠️ 10. 风险与依赖（Risks & Dependencies）

### 风险
- 若将 `metric` 设计成完全独立于 `object_type` 的自由概念，会偏离 BKN 约束
- 若将 `metric` 仅作为 object 附属结果返回，会破坏其在 `search_schema` 中的平级资源定位
- 若只修改服务逻辑、不更新契约和文档，会导致交付不完整

### 依赖
- 外部系统：BKN 原生指标能力与相关接口
- 内部服务：ContextLoader 现有 `search_schema` 能力、MCP 契约、HTTP 契约、发布 toolset
- 参考文档：
  - `adp/docs/design/bkn/features/bkn_native_metrics/DESIGN.md`
  - `adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml`
  - `docs/prd/issue-189-contextloader-schema-search-entry-unification-rfc.md`

---

## 📅 11. 发布计划（Release Plan）

| 阶段 | 时间 | 内容 |
|------|------|------|
| 需求评审 | 待确认 | 确认 PRD 结论与验收边界 |
| 设计评审 | 待确认 | 输出实现设计方案 |
| 开发与测试 | 待确认 | 按设计完成实现与验证 |
| 发布 | 待确认 | 同步服务能力、契约、文档与 toolset |

---

## ✅ 12. 验收标准（Acceptance Criteria）

- Given 调用方开启 `include_metric_types`，When 调用 `search_schema`，Then 接口具备返回 `metric_types` 的能力。
- Given 调用方不传 `search_scope`，When 调用 `search_schema`，Then 四类资源默认均参与响应。
- Given 调用方显式设置 `include_metric_types=false`，When 调用 `search_schema`，Then 响应中不返回 `metric_types`。
- Given 任一资源类型开关显式为 `false`，When 调用 `search_schema`，Then 响应中不返回对应资源类型数据。
- Given `include_object_types=false` 且 `include_metric_types=true`，When 调用 `search_schema`，Then 响应中不返回 `object_types`，但仍允许返回 `metric_types`。
- Given 四个资源开关均为 `false`，When 调用 `search_schema`，Then 返回错误，并提示至少开启一种资源类型。
- Given query 同时命中 metric 直接召回与 object 关联扩展召回，When 返回 `metric_types`，Then 重复 metric 不重复出现。
- Given `include_object_types=false` 且 `include_metric_types=true`，When 系统执行 metric mixed recall，Then 仍允许使用内部 object 候选参与 metric 扩展召回。
- Given 本期启用 metric mixed recall，When 返回结果，Then mixed recall 只增强 `metric_types`，不要求同次补齐 `object_types`，也不改变 `relation_types`、`action_types` 的返回逻辑。
- Given 调用方使用 MCP 工具调用 `search_schema`，When 查看工具契约，Then `include_metric_types` 与 `metric_types` 已被纳入定义。
- Given 调用方使用 HTTP 接口调用 `search_schema`，When 查看接口契约，Then `include_metric_types` 与 `metric_types` 已被纳入定义。
- Given 本需求交付完成，When 检查发布文档与 toolset，Then metric schema 召回能力已被同步说明。
- Given 本需求交付完成，When 回归现有三类 schema 探索场景，Then `object_types`、`relation_types`、`action_types` 的产品行为不退化。

---

## 🔗 附录（Optional）

- 相关文档：
  - `adp/context-loader/agent-retrieval/docs/design/issue-234-contextloader-search-schema-metric-schema-recall-design.md`（待实现设计阶段补充）

- 参考资料：
  - `adp/docs/design/bkn/features/bkn_native_metrics/DESIGN.md`
  - `adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml`
