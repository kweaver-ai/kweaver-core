# 🧩 PRD: ContextLoader 工具上下文结构优化

> 状态: Draft  
> 负责人: Cheng.cao  
> 更新时间: 2026-03-23  

---

## 📌 1. 背景（Background）

- 当前现状：
  - ContextLoader 当前向 Agent 暴露 6 个工具：`kn_schema_search`、`kn_search`、`query_object_instance`、`query_instance_subgraph`、`get_logic_properties_values`、`get_action_info`。
  - 当前 MCP 返回统一走 `BuildMCPToolResult`，即使 `response_format=toon`，`structuredContent` 仍会保留原始对象。
  - 六个工具中，真正影响上下文体积的高成本信息主要集中在执行装配信息、召回解释信息和调试信息。

- 存在问题：
  - 默认返回结果中包含大量不直接参与事实回答或下一步调用的字段，增加 Agent 推理负担和 Token 消耗。
  - 部分字段属于协议层、调试层或算法调参层，不应默认暴露给 LLM 决策。
  - 部分工具的公共契约承担了超出其职责边界的返回内容，例如实例查询同时回传 schema 信息。

- 触发原因 / 业务背景：
  - 本需求来源于 issue #301 `ContextLoader 上下文结构优化`。
  - 本轮目标是先优化 Agent / LLM 可见字段，而不是删除底层 service / DTO 实现。
  - 根据当前调研结论，第一阶段只覆盖 4 个有明确改动项的工具：`get_action_info`、`kn_schema_search`、`get_logic_properties_values`、`query_object_instance`。

---

## 🎯 2. 目标（Objectives）

- 业务目标：
  - 在 `0.5.0` 版本完成本轮 4 个工具的上下文结构收敛。
  - 让 ContextLoader 返回给 Agent 的上下文更聚焦于事实获取和下一步调用。
  - Token 降幅的量化基线、目标值与统计口径：待确认。

- 产品目标：
  - P0、P1 范围内已确认的字段裁剪、隐藏和公共契约收敛项全部落地。
  - 默认 MCP 输出不再暴露本需求明确列出的协议字段、调试字段、召回解释字段和冗余装配字段。
  - MCP schema、默认 MCP 输出与 HTTP 公共文档的对外行为保持一致。
  - 上线后不引入关键调用链回归；具体验证样本与判定口径：待确认。

---

## 👤 3. 用户与场景（Users & Scenarios）

### 3.1 用户角色

| 角色 | 描述 |
|------|------|
| 终端用户 | 间接受到 ContextLoader 返回上下文质量和推理效率的影响。 |
| 开发者 | 直接使用 MCP 客户端、Agent 编排链路或 HTTP 接口调用 ContextLoader 工具。 |
| 管理员 | 负责版本发布、兼容性验证和问题排查。 |

---

### 3.2 用户故事（User Story）

- 作为 Agent 开发者，我希望工具默认返回的上下文只保留必要字段，从而减少无效推理成本。
- 作为接口维护者，我希望 MCP schema、默认输出和 HTTP 公共文档边界清晰，从而降低兼容风险。

---

### 3.3 使用场景

- 场景1：Agent 通过 MCP 调用工具时，只接收对事实回答或后续调用有价值的字段。
- 场景2：研发和测试在 `0.5.0` 发版前，需要逐工具验证哪些字段被裁剪、隐藏或保留。

---

## 📦 4. 需求范围（Scope）

本节使用以下动作定义：

- `默认 MCP 输出裁剪`：仅在默认 MCP 返回视图中不暴露该字段，底层实现暂时保留。
- `MCP schema 中隐藏`：不在 MCP tool 的 `input_schema` / `output_schema` 中对 LLM 暴露。
- `MCP schema 中调整`：对 MCP tool 的输入字段可选性或暴露范围进行调整。
- `HTTP 公共契约/文档收敛`：对外 HTTP 接口文档与公共入口行为同步收敛。
- `固定`：服务端按约定值处理，调用方不再参与该字段决策。
- `保留`：继续对外暴露，行为不变。
- `底层保留`：底层共享结构或兼容逻辑继续保留，不在本次需求中删除。

### ✅ In Scope

| 工具 | HTTP 接口输入 | HTTP 接口输出 | MCP 工具输入 | MCP 工具输出 |
|------|---------------|---------------|--------------|--------------|
| `get_action_info` | `保留` | 裁剪 `_dynamic_tools[].original_schema` | `保留` | `默认 MCP 输出裁剪` `_dynamic_tools[].original_schema` |
| `kn_schema_search` | `保留` | `保留` | `MCP schema 中隐藏` `response_format`、`rerank_action` | `默认 MCP 输出裁剪` `query_understanding`、`hits_total`、score、`samples` |
| `get_logic_properties_values` | `HTTP 公共契约/文档收敛` `options` 中仅保留 `return_debug` | `保留`，兼容 `return_debug=true` 的调试返回 | `MCP schema 中隐藏` `options`、`response_format`，`固定` 默认参数 | `保留`，由默认参数控制是否返回 `debug` |
| `query_object_instance` | `HTTP 公共契约/文档收敛`：`limit` 改可选、移除 `include_type_info`、`固定` `include_type_info=false` | `保留` | `MCP schema 中调整`：`limit` 改可选、移除 `include_type_info` | `默认 MCP 输出裁剪` `object_type` |

### ❌ Out of Scope

- `kn_search`：本轮不修改 MCP schema、默认 MCP 输出和内部实现。
- `query_instance_subgraph`：本轮不修改 MCP schema、默认 MCP 输出和内部实现。
- 本轮不做底层 service / DTO 的彻底删除。
- 本轮不新增独立 debug 模式或新的 action 执行协议。

---

## ⚙️ 5. 功能需求（Functional Requirements）

### 5.1 功能结构

    ContextLoader 工具上下文结构优化
    ├── HTTP 接口
    │   ├── 输入收敛
    │   ├── 输出保留 / 兼容
    │   └── 公共入口固定
    └── MCP 工具
        ├── 输入 schema 收敛
        ├── 默认参数固定
        └── 默认输出裁剪

---

### 5.2 详细功能

#### 【FR-1】`get_action_info` 输出裁剪 `original_schema`

**描述：**  
本轮对 `get_action_info` 的 HTTP 接口输出和 MCP 工具默认输出统一裁剪 `_dynamic_tools[].original_schema` 字段，输入参数不变。

**用户价值：**  
减少动态工具结果中的重复 schema 信息，降低无效上下文成本。

**交互流程：**
1. 调用方通过 HTTP 接口或 MCP 工具调用 `get_action_info`。
2. 服务端完成动作召回并生成动态工具定义。
3. HTTP 接口输出裁剪 `_dynamic_tools[].original_schema` 后返回。
4. MCP 默认输出裁剪 `_dynamic_tools[].original_schema` 后返回。

**业务规则：**
- `HTTP 接口`
  - 输入：`保留`，本轮不调整 HTTP 输入参数。
  - 输出：裁剪 `_dynamic_tools[].original_schema`。
  - 输出：`保留` `headers`、`_dynamic_tools[].name`、`description`、`parameters`、`api_url`、`fixed_params`、`api_call_strategy`。
- `MCP 工具`
  - 输入：`保留`，本轮不调整 MCP 输入 schema。
  - 输出：`默认 MCP 输出裁剪` `_dynamic_tools[].original_schema`。
  - 输出：`保留` `headers`、`_dynamic_tools[].name`、`description`、`parameters`、`api_url`、`fixed_params`、`api_call_strategy`。

**边界条件：**
- 本轮不涉及其他字段调整。
- 底层组装逻辑：`底层保留`。

**异常处理：**
- 若现有调用链仍依赖 `original_schema` 判断参数装配位置，则本项不能直接进入正式发布范围。

---

#### 【FR-2】`kn_schema_search` 对外契约调整

**描述：**  
本轮不调整 `kn_schema_search` 的 HTTP 输入；MCP 工具输入继续收敛；HTTP 接口和 MCP 工具的输出统一裁剪召回解释字段。

**用户价值：**  
让 HTTP 和 MCP 两个入口都返回更聚焦的概念结果，减少解释性负担，同时保持必要输入兼容。

**交互流程：**
1. 调用方通过 HTTP 接口或 MCP 工具调用 `kn_schema_search`。
2. HTTP 接口沿用既有输入参数。
3. MCP schema 不再暴露部分输入字段。
4. 服务端完成召回。
5. HTTP 接口和 MCP 工具的输出统一裁剪召回解释字段后返回。

**业务规则：**
- `HTTP 接口`
  - 输入：`保留`，继续保留 `response_format`、`rerank_action` 的兼容能力。
  - 输出：裁剪 `query_understanding`、`hits_total`、`concepts[].intent_score`、`concepts[].match_score`、`concepts[].rerank_score`、`concepts[].samples`。
  - 输出：`保留` `concepts[].concept_type`、`concepts[].concept_id`、`concepts[].concept_name`、压缩后的 `concept_detail`。
- `MCP 工具`
  - 输入：`MCP schema 中隐藏` `response_format`、`rerank_action`。
  - 输入：`保留` `query`、`search_scope`、`max_concepts`。
  - 输出：`默认 MCP 输出裁剪` `query_understanding`、`hits_total`、`concepts[].intent_score`、`concepts[].match_score`、`concepts[].rerank_score`、`concepts[].samples`。
  - 输出：`保留` `concepts[].concept_type`、`concepts[].concept_id`、`concepts[].concept_name`、压缩后的 `concept_detail`。

**边界条件：**
- 本轮不改变服务端默认召回策略。
- 本轮不允许模型动态决定 rerank 策略。

**异常处理：**
- 若后续明确存在展示召回解释链的需求，应通过 debug / internal 路径解决，不进入默认 Agent 上下文。

---

#### 【FR-3】`get_logic_properties_values` 对外契约调整

**描述：**  
将 `get_logic_properties_values` 的 HTTP 接口与 MCP 工具分别收敛：HTTP 保留调试入口；MCP 仅收敛输入参数，并使用内部默认参数处理。

**用户价值：**  
避免 MCP 调用方感知调试和执行调优参数，同时保留 HTTP 排障能力。

**交互流程：**
1. 调用方通过 HTTP 接口或 MCP 工具调用 `get_logic_properties_values`。
2. HTTP 接口保留调试能力入口，但收敛公共文档。
3. MCP schema 不再暴露调试和执行调优相关输入。
4. MCP 路径使用内部默认参数执行既有逻辑。
5. 输出仍按既有逻辑返回，由输入参数或默认参数决定是否带 `debug`。

**业务规则：**
- `HTTP 接口`
  - 输入：`HTTP 公共契约/文档收敛`，`options` 中仅保留 `return_debug`。
  - 输入：`保留` `ot_id`、`query`、`_instance_identities`、`properties`、`additional_context`。
  - 输出：`保留`，继续兼容 `return_debug=true` 时返回调试信息。
- `MCP 工具`
  - 输入：`MCP schema 中隐藏` `options`、`response_format`。
  - 输入：`保留` `ot_id`、`query`、`_instance_identities`、`properties`、`additional_context`。
  - 输入后处理：`固定` 默认参数 `return_debug=false`、`max_repair_rounds=1`、`max_concurrency=4`。
  - 输出：`保留`，由默认参数控制是否返回 `debug` 和 `datas`。
- 服务端默认值：`底层保留` `return_debug=false`、`max_repair_rounds=1`、`max_concurrency=4`。

**边界条件：**
- `additional_context` 属于真实业务上下文，本轮不做裁剪。
- 本轮不改变内部解析逻辑。
- 本轮不改变 HTTP 下 `return_debug=true` 的兼容语义。

**异常处理：**
- 若调用方需要调试信息，应通过 HTTP 显式传入 `return_debug=true` 获取，而不是通过 MCP 默认路径扩展输入。

---

#### 【FR-4】`query_object_instance` 对外契约调整

**描述：**  
将 `query_object_instance` 的 HTTP 接口与 MCP 工具同时收敛为“查实例”职责，不再默认回传重复的 schema 信息。

**用户价值：**  
减少实例查询结果中的冗余结构，明确实例查询工具边界。

**交互流程：**
1. 调用方调用 `query_object_instance`。
2. HTTP 接口和 MCP 工具都不再要求调用方控制 `include_type_info`。
3. 公共入口固定按 `include_type_info=false` 处理。
4. MCP 默认输出不再返回 `object_type`。

**业务规则：**
- `HTTP 接口`
  - 输入：`HTTP 公共契约/文档收敛`，`limit` 改为可选，默认值为 `10`。
  - 输入：`HTTP 公共契约/文档收敛`，移除 `include_type_info`。
  - 输入：`固定` `include_type_info=false`。
  - 输出：`保留`，本轮不调整 HTTP 输出结构。
- `MCP 工具`
  - 输入：`MCP schema 中调整`，`limit` 改为可选，移除 `include_type_info`。
  - 输入：`保留` `ot_id`、`condition`、`properties`。
  - 输出：`默认 MCP 输出裁剪` `object_type`。
  - 输出：`保留` `datas`。

**边界条件：**
- `include_logic_params`、`response_format` 不纳入本次变更范围。
- 底层 `DrivenOntologyQuery` 中的 `IncludeTypeInfo`：`底层保留`。

**异常处理：**
- 若外部调用方依赖 `object_type` 或 `include_type_info`，需在发版前完成兼容验证和影响识别。

---

## 🔄 6. 用户流程（User Flow）

    调用方
      ├── 走 HTTP 接口
      │   → 按收敛后的 HTTP 公共契约调用
      │   → 服务端执行既有检索 / 解析逻辑
      │   → 返回保留或兼容后的 HTTP 结果
      └── 走 MCP 工具
          → 按收敛后的 MCP schema 调用
          → 服务端执行既有检索 / 解析逻辑
          → 返回裁剪后的默认 MCP 输出

---

## 🎨 7. 交互与体验（UX/UI）

### 7.1 页面 / 模块
- MCP tool `input_schema`
- 默认 MCP `structuredContent`
- HTTP 公共接口文档

### 7.2 交互规则
- 点击行为：无页面交互，本需求以参数暴露和返回结构为核心交互对象。
- 状态变化：`success`、`error`、`debug-only`
- 提示文案：被隐藏的调试或协议字段不再通过默认 MCP 路径对 LLM 暴露；具体错误提示沿用现有实现。

---

## 🚀 8. 非功能需求（Non-functional Requirements）

### 8.1 性能
- 本轮优化不新增下游 RPC / HTTP 调用。
- 默认输出裁剪应优先在适配层或视图层完成。
- 响应时间与上下文体积的量化阈值：待确认。

### 8.2 可用性
- `0.5.0` 发布后，4 个目标工具的关键调用链不得出现明显回归。

### 8.3 安全
- 默认 MCP 路径不得继续暴露不必要的调试字段和执行装配字段。

### 8.4 可观测性
- 继续依赖现有 tracing、日志和监控能力定位字段裁剪与兼容问题。

---

## 📊 9. 埋点与分析（Analytics）

| 事件 | 目的 |
|------|------|
| 本期暂不新增专项埋点 | 以返回结构对比、联调结果和回归验证作为主要验收依据 |

---

## ⚠️ 10. 风险与依赖（Risks & Dependencies）

### 风险
- 当前 MCP 客户端若将 `structuredContent` 原样纳入模型上下文，则默认输出裁剪是本轮收益成立的前提。
- `get_action_info` 的调用方若仍依赖 `original_schema`，本次裁剪会引入兼容风险。
- `query_object_instance` 的外部调用方若依赖 `object_type` 或 `include_type_info`，公共契约收敛会影响既有行为。

### 依赖
- 外部系统：MCP 客户端、依赖现有工具返回结构的上游 Agent 编排链路
- 内部服务：MCP adapter、相关 interfaces / logic 模块、HTTP 接口文档

---

## 📅 11. 发布计划（Release Plan）

- 本需求随 `0.5.0` 版本节奏推进，具体评审、开发、测试和发布时间待确认。

---

## ✅ 12. 验收标准（Acceptance Criteria）

- Given 调用方通过 HTTP 接口调用 `get_action_info`
  When 工具正常返回动态动作定义
  Then `_dynamic_tools[].original_schema` 不出现在 HTTP 输出中，且其余必要字段仍可用于后续动作调用

- Given 调用方通过 MCP 工具调用 `get_action_info`
  When 工具正常返回动态动作定义
  Then `_dynamic_tools[].original_schema` 不出现在默认 `structuredContent` 中，且其余必要字段仍可用于后续动作调用

- Given 调用方通过 HTTP 接口调用 `kn_schema_search`
  When 查看返回结果
  Then `query_understanding`、`hits_total`、`concepts[].intent_score`、`concepts[].match_score`、`concepts[].rerank_score`、`concepts[].samples` 不再出现在输出中

- Given 调用方通过 MCP 工具调用 `kn_schema_search`
  When 查看 MCP schema 与默认输出
  Then `response_format`、`rerank_action` 不再对 LLM 暴露，且 `query_understanding`、`hits_total`、`concepts[].intent_score`、`concepts[].match_score`、`concepts[].rerank_score`、`concepts[].samples` 不出现在默认结果中

- Given 调用方通过 HTTP 接口调用 `get_logic_properties_values`
  When 查看接口文档并传入 `return_debug=true`
  Then 文档中 `options` 仅保留 `return_debug`，且服务仍按兼容语义返回调试信息

- Given 调用方通过 MCP 工具调用 `get_logic_properties_values`
  When 查看 MCP schema 并按默认参数调用
  Then `options`、`response_format` 不再对 LLM 暴露，且由于 `return_debug=false`，默认结果中不返回 `debug`

- Given 调用方通过 HTTP 接口调用 `query_object_instance`
  When 检查输入参数并未传入 `limit`
  Then `limit` 为可选且默认值为 `10`，`include_type_info` 不再作为公共输入暴露，公共入口按 `include_type_info=false` 处理

- Given 调用方通过 MCP 工具调用 `query_object_instance`
  When 查看 MCP schema 与默认输出
  Then `limit` 为可选，`include_type_info` 不再暴露，且默认输出中不包含 `object_type`

- Given 发版前联调发现关键调用链仍依赖被裁剪字段
  When 进入 `0.5.0` 发布评审
  Then 对应改动项不能直接进入正式发布范围

---

## 🔗 附录（Optional）

- 相关文档：
  - [context-loader 工具使用指南](../release/tool-usage-guide.md)

---
