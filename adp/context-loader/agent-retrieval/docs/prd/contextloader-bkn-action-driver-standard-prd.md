# ContextLoader 行动工具适配 BKN 行动驱动

关联 Issue: #304  
负责人: @criller  
状态: Draft  

---

## 1. 背景（Why）

- 当前问题是什么？
  - 当前 ContextLoader 的 `get_action_info` 召回到行动后，返回的 `_dynamic_tools.api_url` 仍指向执行工厂工具 proxy 或 MCP proxy，调用方实际上仍在直接面向底层执行源编排调用。
  - 在 BKN 行动驱动逐步成为统一行动接入与执行入口后，现有模式会导致行动能力治理分散，且 `get_action_info` 返回的动态工具不保证一定能通过行动驱动完成执行。
  - 当前动态工具的执行语义仍偏向“直接调用底层工具并获取结果”，与行动驱动的“提交一次行动执行任务”语义不一致，容易造成调用方对返回契约的误解。
- 为什么要做？
  - 需要让 ContextLoader 暴露的动态工具与产品约定的统一执行契约保持一致，避免继续把执行工厂/MCP proxy 暴露给上层调用方。
  - 需要为后续行动能力的统一管理、执行、观测和演进提供基础能力。
- 来源（客户 / 内部 / 老板 / GitHub）
  - 来源于 GitHub Issue `#304`，需求名称为“ContextLoader 行动工具适配 BKN 行动驱动”。

---

## 2. 目标（Goal）

- 本次要解决什么问题？
  - 在 BKN 行动驱动可用的环境下，保证 ContextLoader 暴露的 `get_action_info` 所返回的 `_dynamic_tools` 统一指向行动驱动执行接口，而非执行工厂工具 proxy 或 MCP proxy。
  - 调整 `get_action_info` 的召回与组装逻辑，使其优先返回能够通过 BKN 行动驱动执行的动态工具信息，并以行动驱动契约暴露给调用方。
  - 在 `0.5.0` 版本完成上线，并临时兼容一期现有调用方式。
- 成功的标准是什么？
  - Given 测试知识网络下某行动已按产品约定接入 BKN 行动驱动，When 调用 `get_action_info` 且入参正确，Then 返回的 `_dynamic_tools` 非空，且其 `api_url` 指向行动驱动执行接口。
  - Given 上述 `_dynamic_tools` 已返回，When 调用方按返回契约发起一次执行验证，Then 能成功提交一次行动执行任务，并获得有效的 `execution_id`。
  - `get_action_info` 对外请求参数保持兼容，调用方无需因本次改造同步切换调用入口。
  - 不改造 MCP 协议本身，仅完成 ContextLoader 与 BKN 行动驱动之间的接入与召回逻辑调整。

---

## 3. 用户/使用场景（Who & When）

- 谁会用？
  - 直接用户是 Agent 开发者，他们依赖 `get_action_info` 返回的 `_dynamic_tools` 完成行动发现与后续执行。
- 在什么场景使用？
  - Agent 在面向某对象实例和某行动类型生成执行计划时，会调用 `get_action_info` 获取动态工具。
  - 当知识网络中的某行动已经按产品约定接入 BKN 行动驱动时，Agent 开发者希望返回的动态工具直接提交到统一行动执行接口，而不再关心底层是 tool 还是 MCP。
  - Agent 在提交行动执行任务后，需要基于返回的 `execution_id` 继续查询执行状态或消费后续执行结果。
  - 在兼容期内，已有调用方仍沿用当前 `get_action_info` 的调用方式，但返回的动态工具底层执行入口逐步切换到 BKN 行动驱动。

---

## 4. 功能需求（What）

### 4.1 动态工具执行入口统一切换到行动驱动接口

- 描述：
  - ContextLoader 在处理 `get_action_info` 时，需要将原先直连执行工厂/MCP 的行动工具执行入口替换为 BKN 行动驱动统一执行接口。
- 输入：
  - `get_action_info` 当前既有入参，包括但不限于行动类型标识、对象实例标识和知识网络上下文。
  - 行动驱动执行接口契约，包括路径参数、请求体结构以及异步提交返回结构。
- 输出：
  - 返回 `_dynamic_tools`，其中 `api_url` 指向行动驱动执行接口，动态工具整体契约满足 BKN 行动驱动执行约定，能够被上层调用方继续消费。
- 规则：
  - 在 BKN 行动驱动可用且目标行动已按约定接入时，返回结果必须可通过行动驱动完成执行。
  - `api_url` 应统一指向行动驱动执行接口，不再返回执行工厂工具 proxy 或 MCP proxy 地址。
  - 返回结构需体现“提交行动执行任务”的语义，而非“直接调用底层工具并同步拿结果”的语义。
  - 若产品存在统一网关或统一契约，返回结构需与该契约保持一致。

### 4.2 `get_action_info` 动态工具契约改造

- 描述：
  - 改造 `get_action_info` 的召回与动态工具组装逻辑，使其在候选行动中优先识别并返回可通过 BKN 行动驱动执行的动态工具，并按行动驱动接口契约组织 `api_url`、参数与固定上下文。
- 输入：
  - 当前知识网络中与目标对象实例、行动类型相关的行动集合及其接入信息。
  - 行动驱动执行接口的请求结构，包括 `branch`、`_instance_identities`、`dynamic_params`、`trigger_type` 等字段约束。
- 输出：
  - `_dynamic_tools` 中仅包含满足当前产品约定、可通过 BKN 行动驱动执行的动态工具信息。
  - 动态工具的参数定义与固定参数分工满足行动驱动接口调用要求，调用方无需再感知底层 tool/MCP 差异。
- 规则：
  - 返回结果应过滤掉无法按行动驱动执行的候选行动。
  - 动态工具应默认带入当前召回对象实例对应的执行上下文，避免调用方再次面向底层执行源重组对象定位信息；具体字段落位在实现设计阶段明确。
  - 动态工具应暴露行动驱动接口所需的用户可填参数，并隐藏底层 tool/MCP 特有的执行入口细节。
  - 当入参正确且测试知识网络存在符合约定的行动时，`_dynamic_tools` 不得为空。
  - 具体排序、过滤、失败返回策略在实现设计阶段补充，未确认部分在本文档中标注为 `待确认`。

### 4.3 临时兼容一期

- 描述：
  - 在过渡期内保留一期兼容能力，降低现有调用方迁移成本。
- 输入：
  - 当前已接入 `get_action_info` 的 Agent 或编排调用方。
- 输出：
  - 调用方仍可沿用现有 `get_action_info` 请求方式，并获得可用于统一执行链路的动态工具结果。
- 规则：
  - 本期兼容范围仅限 `get_action_info` 的对外请求方式与返回字段骨架，不代表长期保留旧执行路径。
  - 在兼容期内，即使调用方不修改 `get_action_info` 调用方式，返回的 `_dynamic_tools.api_url` 也应切换为行动驱动执行接口。
  - 兼容期结束时间、开关策略、灰度方案为 `待确认`。
  - 若 BKN 行动驱动不可用时是否返回错误或空结果，处理策略为 `待确认`。

---

## 5. 非功能需求（Non-Functional）

- 性能（QPS / 延迟）
  - `get_action_info` 在接入 BKN 行动驱动后不应出现明显性能退化；具体延迟阈值与压测口径为 `待确认`。
- 可用性（SLA）
  - 在 BKN 行动驱动可用环境下，`get_action_info` 需要稳定返回可执行动态工具；当下游异常时应具备明确、可诊断的失败表现。
  - 行动驱动执行接口应能稳定返回异步提交结果；若执行提交失败，错误信息需可被调用方定位。
- 安全性
  - ContextLoader 与 BKN 行动驱动之间的调用需遵循平台既有鉴权与访问控制规范；账号身份、业务域等上下文透传要求需与行动驱动真实实现保持一致，具体细节为 `待确认`。
- 可扩展性
  - 本次改造不修改 MCP 协议本身，但返回结构和接入方式需支持后续更多行动按统一驱动方式接入。
  - 动态工具契约应避免继续绑定某一种底层执行源，支持后续新增执行源仍复用同一行动驱动入口。

---

## 6. 验收标准（DoD - Definition of Done）✅

- Given 测试知识网络下某行动已按产品约定接入 BKN 行动驱动，When 调用 `get_action_info` 且入参正确，Then 返回的 `_dynamic_tools` 非空，且 `api_url` 指向行动驱动执行接口。
- Given 上述 `_dynamic_tools` 已返回，When 按行动驱动约定发起一次执行验证，Then 接口返回 `202 Accepted`，且响应中包含有效的 `execution_id`。
- Given 上述 `execution_id` 已返回，When 调用方继续查询执行状态，Then 可以通过行动驱动提供的状态查询接口获取执行进度或结果。
- Given 调用方仍沿用现有 `get_action_info` 请求方式，When 升级到 `0.5.0` 版本，Then 无需修改 MCP 协议即可继续完成调用。
- Given 某候选行动未按产品约定接入 BKN 行动驱动，When 调用 `get_action_info`，Then 返回行为符合产品约定的过滤或失败策略，具体策略为 `待确认`。
- Given 本次需求开发完成，When 交付验收，Then 相关测试、文档和发布说明已补齐。

---

## 7. 不做范围（Out of Scope）

- 不改造 MCP 协议本身。
- 不定义或重构 BKN 行动驱动服务自身的产品能力。
- 不扩展与 `get_action_info` 无关的其他 ContextLoader 工具能力。
- 不在本期承诺长期保留旧执行路径；仅明确“临时兼容一期”。

---

## 8. 风险 & 依赖

- 外部依赖：
  - BKN 行动驱动服务可用，且已有明确的接入契约或统一网关约定。
  - `ontology-query` 行动执行接口的真实实现与对外文档能够达成一致，或已在设计阶段明确以代码实现为准。
  - 存在可用于验收的测试知识网络与执行验证环境。
- 潜在风险：
  - 若行动驱动契约在开发中变更，可能导致 `get_action_info` 返回结构、参数映射和执行验证逻辑返工。
  - 若行动执行接口文档与真实代码实现持续不一致，可能导致 PRD、实现设计和联调口径出现偏差。
  - 若兼容策略、开关策略和失败策略未提前明确，可能影响 `0.5.0` 发版节奏。
  - 若测试环境缺少已接入行动驱动的标准样例，验收结果将难以稳定复现。

---

## 9. 参考资料

- GitHub Issue: `https://github.com/kweaver-ai/adp/issues/304`
- BKN Ontology Query API: `docs/api/bkn/ontology-query-ai/ontology-query.yaml`
- BKN Ontology Query action execution routes and handlers: `bkn/ontology-query/server/driveradapters/routers.go`、`bkn/ontology-query/server/driveradapters/action_execution_handler.go`
- BKN Ontology Query action execution request model: `bkn/ontology-query/server/interfaces/action_execution.go`

