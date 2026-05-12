# 🧩 PRD: Agent绑定Skill元数据注入

> 状态: Draft  
> 负责人: 待确认  
> 更新时间: 2026-05-12  

---

## 📌 1. 背景（Background）

- 当前现状：
  - AgentFactory 已支持在创建、编辑 Agent 时绑定 agent skill，并通过 Agent 配置中的 `skills.skills[].skill_id` 表达关联关系。
  - Agent Executor 已支持运行时注册 skill 相关内置工具，包括 `SkillLoadTool`、`SkillReadTool`、`SkillExecuteTool`，可通过 operator-integration 的 skill API 加载、读取和执行 skill。
  - Agent Executor 当前系统提示词中已有通用 skill 工具使用规则，但未把当前 Agent 绑定的 skill 元数据动态注入给模型。

- 存在问题：
  - 模型运行时不知道当前 Agent 绑定了哪些 skill，只能依赖用户输入或上下文偶然提供 `skill_id`。
  - 已绑定 skill 与运行时工具能力之间缺少提示词层面的连接，导致模型难以主动选择并调用 `builtin_skill_load(skill_id)`。
  - 如果直接注入完整 `SKILL.md`，会增加 token 成本和首包延迟，不适合作为首版默认方案。

- 触发原因 / 业务背景：
  - 需要打通“Agent 配置绑定 skill”到“Agent 运行时可感知并使用 skill”的闭环。
  - 首版聚焦后端运行时能力补齐，使绑定 skill 能以轻量 metadata 形式进入系统提示词，为模型按需加载完整 skill 内容提供入口。

---

## 🎯 2. 目标（Objectives）

- 业务目标：
  - 在首版发布后，已绑定 skill 的 Agent 运行请求中，Executor 能完成绑定 skill 元数据注入，功能正确性通过验收用例覆盖率达到 100%。
  - 不改变 AgentFactory 现有创建、编辑、保存链路，避免影响已有 Agent 配置管理流程。

- 产品目标：
  - 当 Agent 配置存在 `skills.skills[].skill_id` 且对应 skill 状态为 `published` 时，系统提示词中必须包含该 skill 的 `skill_id`、`name`、`description`。
  - skill metadata 拉取失败、无权限、404、非 `published` 状态时，本次 Agent 运行不中断，失败项不注入并记录日志。
  - 系统提示词不得注入完整 `SKILL.md`、文件清单或脚本清单，保持首版注入内容轻量可控。

---

## 👤 3. 用户与场景（Users & Scenarios）

### 3.1 用户角色

| 角色 | 描述 |
|------|------|
| 终端用户 | 使用已配置 Agent 发起对话或任务，希望 Agent 能利用已绑定 skill 完成回答或处理。 |
| Agent 配置者 | 在 AgentFactory 中为 Agent 绑定 skill 的产品运营、管理员或开发者。 |
| 后端开发者 | 维护 Agent Executor、AgentFactory、operator-integration 的工程人员。 |

---

### 3.2 用户故事（User Story）

- 作为终端用户，我希望已绑定 skill 的 Agent 能识别自身可用 skill，从而在需要时调用对应能力完成任务。
- 作为 Agent 配置者，我希望在配置 Agent 绑定 skill 后，运行时无需额外手写 `skill_id` 到系统提示词，从而降低配置成本和出错率。
- 作为后端开发者，我希望 skill metadata 注入逻辑有明确边界、异常降级和测试覆盖，从而降低对现有 Agent 运行链路的回归风险。

---

### 3.3 使用场景

- 场景1：标准 Agent 模式下，配置者为 Agent 绑定一个已发布 skill，用户发起对话后，模型在系统提示词中看到该 skill 的 `skill_id/name/description`，并可按需调用 `builtin_skill_load(skill_id)`。
- 场景2：自定义 Dolphin 模式下，Agent 的 dolphin 语句中包含 `/explore`，Executor 在运行前将绑定 skill metadata 注入该 explore 的 `system_prompt`。
- 场景3：Agent 绑定多个 skill，其中部分 skill 不存在、无权限或未发布，Executor 跳过不可用项，保留可用项注入并继续运行。

---

## 📦 4. 需求范围（Scope）

### ✅ In Scope

- Executor 解析 Agent 配置中的 `skills.skills[].skill_id`。
- Executor 通过 `AgentOperatorIntegrationService` 调用 operator-integration 私有接口获取 skill 详情。
- 注入内容限定为 `skill_id`、`name`、`description`。
- 仅注入状态为 `published` 的 skill。
- 单次 Agent 运行内按 `skill_id` 去重，保持配置顺序。
- 支持标准模式系统提示词注入。
- 支持自定义 Dolphin 模式 `/explore` 的 `system_prompt` 注入。
- metadata 获取失败时跳过失败项并记录日志，不阻断 Agent 运行。
- 补充必要单元测试和回归测试。

### ❌ Out of Scope

- 不改 AgentFactory 创建、编辑、保存 Agent 配置的产品交互和 API。
- 不改前端 Agent 配置页面。
- 不注入完整 `SKILL.md`。
- 不注入 skill 文件清单、脚本清单、依赖信息或 `extend_info`。
- 不新增跨请求全局缓存、TTL 或缓存失效机制。
- 不新增调用转化率、模型选择效果等分析型指标。
- 不改变 `SkillLoadTool`、`SkillReadTool`、`SkillExecuteTool` 的既有调用契约。

---

## ⚙️ 5. 功能需求（Functional Requirements）

### 5.1 功能结构

    Agent Skill 元数据注入
    ├── 配置解析
    │   └── 支持 skills.skills[].skill_id
    ├── 元数据获取
    │   └── AgentOperatorIntegrationService.get_skill_info
    ├── 元数据过滤与格式化
    │   ├── 请求内去重
    │   ├── published 状态过滤
    │   └── prompt 片段生成
    ├── 系统提示词注入
    │   ├── 标准模式注入
    │   └── Dolphin 模式注入
    └── 降级与观测
        ├── 失败跳过
        └── 日志记录

---

### 5.2 详细功能

#### FR-1 配置解析支持绑定 skill

**描述：**  
Executor 需要识别 Agent 配置中的 `skills.skills` 字段，并将每个元素解析为包含 `skill_id` 的结构化对象。

**用户价值：**  
确保 AgentFactory 已保存的 skill 绑定关系能被 Executor 正确消费。

**交互流程：**
1. AgentFactory 调用 Executor 运行接口，透传 Agent 配置。
2. Executor 构造 `AgentConfigVo`。
3. `AgentConfigVo.skills` 将 `skills.skills[].skill_id` 解析为可遍历对象。

**业务规则：**
- `skills.skills` 为空时，不生成 skill metadata prompt。
- `skill_id` 为空字符串时，该项视为无效并跳过。
- 原有 `tools`、`agents`、`mcps` 字段解析行为保持兼容。

**边界条件：**
- `skills` 为 `null` 时，按空 skill 配置处理。
- `skills.skills` 为 `null` 时，按空列表处理。
- 重复 `skill_id` 在后续 metadata 构建阶段去重。

**异常处理：**
- 配置结构不符合预期时，不中断 Agent 运行，按空绑定 skill 处理并记录日志。

---

#### FR-2 获取 skill 元数据

**描述：**  
Executor 在运行时通过 `AgentOperatorIntegrationService.get_skill_info(skill_id)` 获取 skill 详情。

**用户价值：**  
模型能获得已绑定 skill 的可读名称和描述，而不是只看到无语义的 ID。

**交互流程：**
1. Executor 提取绑定的 `skill_id` 列表。
2. Executor 按配置顺序对 `skill_id` 去重。
3. Executor 对每个去重后的 `skill_id` 调用 operator-integration skill 详情接口。
4. Executor 从响应 `data` 中读取 `skill_id`、`name`、`description`、`status`。

**业务规则：**
- 接口路径为 `GET /api/agent-operator-integration/internal-v1/skills/market/{skill_id}`。
- 只使用 `skill_id`、`name`、`description`、`status` 四个字段。
- 仅当 `status == "published"` 时，该 skill 可注入系统提示词。
- 单次请求内去重，不做跨请求缓存。

**边界条件：**
- 接口返回成功但缺少 `skill_id` 时跳过该项。
- `name` 或 `description` 为空时，仍可注入已有字段，但不得伪造内容。
- 同一 Agent 绑定多个 skill 时，注入顺序与去重后的配置顺序一致。

**异常处理：**
- 非 200、404、403、网络异常、响应解析失败均跳过该项。
- 跳过原因需要记录 warn 日志，便于排查配置或权限问题。

---

#### FR-3 构建 skill metadata prompt

**描述：**  
Executor 将可用 skill 元数据格式化为系统提示词片段，指导模型先调用 `builtin_skill_load(skill_id)` 再使用具体 skill。

**用户价值：**  
模型可以理解当前 Agent 可用 skill 的用途，并具备按需加载完整 skill 内容的入口。

**交互流程：**
1. Executor 收集所有可注入 skill metadata。
2. Executor 生成固定标题和使用说明。
3. Executor 按列表格式输出每个 skill 的 `skill_id`、`name`、`description`。

**业务规则：**
- prompt 片段必须包含“使用 skill 前先调用 `builtin_skill_load(skill_id)`”的规则。
- prompt 片段不得包含完整 `SKILL.md`。
- prompt 片段不得包含文件清单、脚本清单、依赖信息。

**边界条件：**
- 无可用 skill 时，返回空字符串，不改变原系统提示词。
- 文本字段中包含换行时，输出应保持可读，不破坏 Dolphin prompt 构造。

**异常处理：**
- 格式化异常时返回空字符串并记录错误日志，不阻断运行。

---

#### FR-4 标准模式系统提示词注入

**描述：**  
在非 Dolphin 模式下，Executor 自动生成 `/explore` 的 `system_prompt` 时，将 skill metadata prompt 注入到系统提示词中。

**用户价值：**  
标准 Agent 无需手写 Dolphin 语句即可获得绑定 skill 感知能力。

**交互流程：**
1. Executor 生成通用 skill 工具使用规则。
2. Executor 追加绑定 skill metadata prompt。
3. Executor 追加用户配置的 `system_prompt`。
4. Executor 使用组合后的 system prompt 构造 `/explore`。

**业务规则：**
- 注入顺序为：通用 skill 工具使用规则、绑定 skill metadata、用户系统提示词。
- 计划模式下，保留现有 plan mode system prompt 逻辑，并在最终用户提示词之前注入 metadata。
- `features.skill_enabled` 为 `false` 时，不注入通用 skill 工具规则和绑定 skill metadata。

**边界条件：**
- 用户 `system_prompt` 为空时，仍可注入通用规则和绑定 skill metadata。
- 无可用 skill metadata 时，保持现有 prompt 行为。

**异常处理：**
- prompt 组合失败时降级为现有系统提示词并记录日志。

---

#### FR-5 Dolphin 模式系统提示词注入

**描述：**  
在自定义 Dolphin 模式下，Executor 对最终 dolphin prompt 中的 `/explore` 语句进行保守改写，将 skill metadata prompt 注入 `system_prompt`。

**用户价值：**  
使用自定义 Dolphin 编排的 Agent 也能获得绑定 skill 感知能力。

**交互流程：**
1. Executor 拼接 `pre_dolphin`、用户 `dolphin`、`post_dolphin`。
2. Executor 扫描 dolphin prompt 中的 `/explore/(...)` 语句。
3. 对已有 `system_prompt` 的 explore，prepend skill metadata prompt。
4. 对缺少 `system_prompt` 的 explore，补入 `system_prompt` 参数。
5. 无法安全识别的 explore 语句保持原样。

**业务规则：**
- 仅在 `features.skill_enabled` 为 `true` 且存在可用 skill metadata 时执行注入。
- 不改变 `/explore` 的其他参数含义。
- 不改变非 `/explore` Dolphin 语句。

**边界条件：**
- 同一 dolphin prompt 中存在多个 `/explore` 时，每个可安全识别的 explore 都应注入。
- 原 `system_prompt` 包含转义字符或复杂表达式时，无法安全处理则跳过该 explore。

**异常处理：**
- 单个 explore 注入失败不影响其他 explore。
- 所有 explore 均无法注入时，Agent 继续按原 dolphin prompt 运行并记录 warn 日志。

---

## 🔄 6. 用户流程（User Flow）

    Agent 配置者在 AgentFactory 绑定 skill
        ↓
    终端用户发起 Agent 对话
        ↓
    AgentFactory 将 Agent 配置透传给 Agent Executor
        ↓
    Agent Executor 解析 skills.skills[].skill_id
        ↓
    Agent Executor 查询 operator-integration 获取 skill metadata
        ↓
    Agent Executor 将可用 skill metadata 注入 system_prompt
        ↓
    Dolphin SDK 执行 Agent
        ↓
    模型按需调用 builtin_skill_load/read/execute
        ↓
    Agent 返回回答或执行结果

---

## 🎨 7. 交互与体验（UX/UI）

### 7.1 页面 / 模块
- AgentFactory Agent 创建与编辑页面：首版不改造。
- Agent 运行接口：首版无新增用户可见参数。
- Agent Executor 日志与链路追踪：用于研发和运维排查。

### 7.2 交互规则
- 点击行为：无新增前端点击行为。
- 状态变化：无新增前端状态。
- 提示文案：首版不新增用户可见提示文案；metadata 拉取失败仅记录服务端日志。

---

## 🚀 8. 非功能需求（Non-functional Requirements）

### 8.1 性能
- 单次 Agent 运行内对重复 `skill_id` 只查询一次。
- metadata 查询在 prompt 构造前完成，不引入跨请求缓存。
- 首版不设置强制性能指标；首包延迟影响需在测试环境观察并记录，目标为不造成可感知阻塞。该目标待确认。

### 8.2 可用性
- 任意单个 skill metadata 获取失败不得导致 Agent 运行失败。
- operator-integration 不可用时，Executor 跳过 metadata 注入并继续执行原有 Agent 流程。

### 8.3 安全
- Executor 沿用现有请求头透传和 operator-integration 权限控制。
- 不在系统提示词中注入下载地址、文件路径、脚本命令、依赖信息或敏感扩展字段。
- 非授权或不存在的 skill 不注入系统提示词。

### 8.4 可观测性
- 支持日志记录 metadata 获取失败、状态过滤、Dolphin 注入跳过等事件。
- 支持 tracing 标识 Agent ID、Agent Run ID、skill_id 查询阶段。具体 span 字段待实现设计确认。
- 首版不新增业务 metrics，后续可根据调用转化率分析需求扩展。

---

## 📊 9. 埋点与分析（Analytics）

| 事件 | 目的 |
|------|------|
| skill_metadata_fetch_success_log | 研发排查 metadata 获取是否成功。 |
| skill_metadata_fetch_skip_log | 研发排查 skill 被跳过的原因，包括状态不可用、无权限、404、接口异常。 |
| skill_metadata_prompt_injected_log | 研发确认本次运行是否完成 metadata prompt 注入。 |
| dolphin_skill_metadata_inject_skip_log | 研发排查 Dolphin 模式中无法安全注入的 explore 语句。 |

---

## ⚠️ 10. 风险与依赖（Risks & Dependencies）

### 风险
- Dolphin 模式 prompt 改写存在语法兼容风险，复杂表达式无法安全注入时将保持原语句并记录日志。
- 绑定 skill 数量较多时，会增加运行前 metadata 查询次数。
- skill 名称或描述质量不足时，模型对 skill 适用场景的判断准确性会下降。
- operator-integration 接口响应结构变化会影响 metadata 获取。

### 依赖
- 外部系统：无第三方外部依赖。
- 内部服务：AgentFactory、Agent Executor、Agent Operator Integration、Dolphin SDK。
- 接口依赖：`GET /api/agent-operator-integration/internal-v1/skills/market/{skill_id}`。
- 配置依赖：`features.skill_enabled` 继续作为 skill 能力总开关。

---

## 📅 11. 发布计划（Release Plan）

| 阶段 | 时间 | 内容 |
|------|------|------|
| 需求评审 | 待确认 | 确认注入字段、状态策略、失败降级、Dolphin 模式范围。 |
| 开发 | 待确认 | 完成 VO、operator-integration service、metadata builder、PromptBuilder 注入逻辑。 |
| 测试 | 待确认 | 完成单元测试、标准模式回归、Dolphin 模式回归、异常降级验证。 |
| 发布 | 待确认 | 随 Agent Executor 后端版本发布，发布后观察日志中 metadata 获取和注入情况。 |

---

## ✅ 12. 验收标准（Acceptance Criteria）

    Given Agent 配置包含一个 status 为 published 的绑定 skill
    When 用户发起标准模式 Agent 运行
    Then Executor 生成的 system_prompt 包含该 skill 的 skill_id、name、description

    Given Agent 配置包含重复 skill_id
    When Executor 构建 skill metadata
    Then 相同 skill_id 只查询一次并只注入一次

    Given Agent 配置包含 status 为 offline 或 unpublish 的 skill
    When Executor 获取到该 skill 详情
    Then 该 skill 不注入 system_prompt，Agent 运行继续

    Given operator-integration 返回 404、403、非 200 或请求异常
    When Executor 获取绑定 skill metadata
    Then Executor 跳过失败项、记录日志，并继续运行 Agent

    Given Agent 使用自定义 Dolphin 模式且 /explore 已配置 system_prompt
    When Executor 构建 dolphin prompt
    Then Executor 将绑定 skill metadata prepend 到原 system_prompt 中

    Given Agent 使用自定义 Dolphin 模式且 /explore 未配置 system_prompt
    When Executor 构建 dolphin prompt
    Then Executor 为可安全识别的 /explore 补入包含绑定 skill metadata 的 system_prompt

    Given Agent 配置绑定 skill
    When Executor 注入系统提示词
    Then 注入内容不包含完整 SKILL.md、文件清单、脚本清单或下载地址

    Given features.skill_enabled 为 false
    When Executor 构建 Agent 运行 prompt
    Then Executor 不注入 skill 工具使用规则和绑定 skill metadata

---

## 🔗 附录（Optional）

- 相关文档：
  - HLD链接：待确认
  - LLD链接：待确认

- 参考资料：
  - operator-integration skill API：`/Users/guochenguang/project/kweaver/adp/execution-factory/operator-integration/docs/apis/api_private/skill.yaml`
  - Agent Executor skill 工具实现：`agent-executor/app/common/tool_v2/skill_contract_tools.py`
  - Agent Executor prompt 构建逻辑：`agent-executor/app/logic/agent_core_logic_v2/prompt_builder.py`
  - Agent Executor operator-integration 客户端：`agent-executor/app/driven/dip/agent_operator_integration_service.py`

---
