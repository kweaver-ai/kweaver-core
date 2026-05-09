# Design Note: Issue 109 Skill Recall Release 前调整

> 状态: Draft
> 分支: `fix/find-skills-empty-result-hint`
> 基线设计: `docs/design/issue-109-contextloader-skill-recall-design.md`
> 背景: Issue 109 已合并到 `main`，但当前版本尚未拉取 release；在 release 前发现回归和体验问题，需要先完成一次小范围修正。

---

## 1. 本文档目的

本文件用于在实现前固化本次 release 前调整的边界，避免遗漏需求或在同一修复分支中继续扩大范围。

本次调整遵循以下原则：

- 单分支处理：使用一个 `fix/*` 分支完成本次修正
- 单 PR 交付：行为修复和配套 release 文档一起提交
- 小步推进：实现阶段按两个小任务顺序处理，而不是混在一起改
- 不扩功能：仅修正空结果终止语义与使用说明，不扩展 skill recall 能力边界

---

## 2. 本次确认要处理的问题

### 2.1 问题一：`find_skills` 空结果会诱发 Agent 重复调用

当前 `find_skills` 成功响应只有 `entries` 字段。当接口返回：

```json
{
  "entries": []
}
```

调用方只能知道“没有结果”，但无法知道：

- 为什么没有结果
- 当前是否应停止继续调用
- 下一步应该缩小什么边界或补什么上下文

内部反馈表明，这种返回形态可能诱发 Agent 持续重复调用 `find_skills`。

### 2.2 问题二：release 文档缺少 skill recall 使用指南

虽然本版增加了 skill recall 能力，但用户侧并不知道：

- `find_skills` 只负责运行时发现 Skill 候选，不负责创建 Skill
- Skill recall 的前提条件是什么
- skill 元数据、对象类、绑定关系、模板规范分别由谁负责
- 调用方应该在什么时机使用 `find_skills`
- 空结果意味着什么、下一步应该怎么处理

当前 release 文档对“怎么用”描述不足，容易导致接入方误以为只要升级 Context Loader 就能直接用好 skill recall。

---

## 3. 方案决策

### 3.1 分支策略

本次采用：

- 一个分支：`fix/find-skills-empty-result-hint`
- 一个 PR：包含行为修复和配套 release 文档
- 两个小任务：
  - 任务 A：修复 `find_skills` 空结果终止语义
  - 任务 B：补充 skill recall release 使用指南

这样做的原因：

- 这两项都属于同一发布单元，适合一起进入 release
- 当前由同一维护者在 fork 分支中处理，拆多个 PR 的收益不高
- 通过“两任务、两次提交”的方式，仍然可以保持 review 边界清晰

### 3.2 问题一的推荐修正

保留 `entries` 的纯数据语义，不向 `entries` 中塞入伪造的 Skill 条目。

推荐在成功响应顶层新增一个文本字段，命名为 `message`，用于在空结果时向 Agent 明确表达：

- 本次查询已正常结束
- 当前边界下没有匹配 Skill
- 如果需要继续探索，下一步建议是什么

目标形态示意：

```json
{
  "entries": [],
  "message": "No skills found in the current scope. If network-level recall returned empty, try narrowing with object_type_id or instance_identities."
}
```

约束如下：

- `entries` 继续只包含真实 Skill 元数据
- `message` 只作为辅助解释，不承载结构化业务数据
- 空结果时必须可读、可执行，不写成纯日志口吻
- 非空结果时保持最小兼容改动，默认不返回 `message`
- 4xx / 5xx 错误场景继续走错误响应体，不复用成功响应中的 `message`

#### 3.2.1 `message` 的职责边界

`message` 的职责是表达本次成功调用的结果说明，用于告诉调用方：

- 这次查询已经正常结束
- 当前为什么没有结果
- 下一步建议从哪个方向继续收缩或放宽边界

`message` 不承担以下职责：

- 不替代错误响应中的 `description`
- 不返回结构化状态码
- 不描述内部服务异常、重试退避等系统层行为

本次实现采用以下返回规则：

- `entries` 非空：默认不返回 `message`
- `entries` 为空且接口返回 `200`：返回 `message`
- 接口返回 `4xx / 5xx`：走现有错误响应体

#### 3.2.2 `message` 场景划分与中文文案

为避免不同空结果原因混用同一条提示，本次按触发优先级划分 `message`：

1. 优先判断结构性原因
2. 其次判断是否为 `skill_query` 过滤后无结果
3. 最后回落到通用“当前范围无结果”

推荐场景与文案如下：

| 场景 | 触发条件 | 推荐 `message` |
|------|----------|----------------|
| 网络级范围过宽 | 仅传 `kn_id`，且 `skills` 与其他对象存在关系 | 当前知识网络中的 Skill 不是全局生效，网络级范围未返回结果。请补充 `object_type_id`；如已定位到实例，再补充 `instance_identities` 后重试。 |
| 网络级下确实没有 Skill | 仅传 `kn_id`，允许网络级查询，但结果为空 | 当前知识网络下未找到可召回的 Skill。请确认 Skill 是否已创建并在当前知识网络中可见。 |
| 对象类未配置 Skill 绑定 | 传了 `object_type_id`，但对象类与 `skills` 不存在绑定关系 | 当前对象类未配置 Skill 绑定关系，无法在该范围内召回 Skill。请确认该对象类是否已绑定 Skill。 |
| 对象类范围无匹配 | 对象类级查询成功执行，但结果为空，且未传 `skill_query` | 当前对象类范围内未找到可召回的 Skill。可尝试缩小到具体实例范围，或确认该对象类下是否已绑定 Skill。 |
| 实例范围无匹配 | 实例级查询成功执行，但结果为空，且未传 `skill_query` | 当前实例范围内未找到可召回的 Skill。可回退到 `object_type_id` 级别查看该类对象的候选 Skill。 |
| 当前筛选条件无匹配 | 查询成功执行，但结果为空，且传了 `skill_query` | 当前范围和 `skill_query` 条件下未找到匹配的 Skill。可尝试放宽或去掉 `skill_query` 后重试。 |

文案约束：

- 每条 `message` 都应包含“结果说明 + 下一步建议”
- 文案要面向 Agent 和接入方，不写成日志口吻
- 不写“可能系统异常”“请稍后重试”之类会误导调用方的泛化表述

### 3.3 问题二的推荐修正

在 release 文档中增加一段“Skill Recall 使用指南”，但内容组织方式不应以“参数怎么传”为主，而应以“能否启用、为什么空结果”为主。

文档目标是降低接入方误解和研发答疑成本，因此优先回答：

- `find_skills` 为什么不是升级后自动可用
- 使用前需要满足哪些前提
- 不满足前提时为什么会出现空结果
- 最后才补充参数与调用方式

推荐在 `docs/release/tool-usage-guide.md` 中采用“两层结构”：

- 第一层：快速开始，服务新用户快速判断“能不能用、怎么最快用起来”
- 第二层：详细说明，服务接入排查和背景理解

#### 3.3.1 快速开始

快速开始必须先写“启用前提”，不能只写调用参数。建议包含以下内容：

**a. 先确认能不能用**

在调用 `find_skills` 前，必须先满足以下硬性要求：

- BKN 中已存在固定的 `skills` ObjectType
- 该 ObjectType 的运行时识别键必须是 `object_type_id = "skills"`
- 该 ObjectType 不是任意自定义对象类替代，而是 Skill 的固定承接面
- `skills` ObjectType 下已有可见的 Skill 元数据实例
- 业务对象与 `skills` 之间已配置绑定关系

建议增加一条强提醒：

> 如果以上条件未满足，`find_skills` 虽然可以调用，但很可能只返回空结果。

**b. 最短使用路径**

在快速开始中仅保留最短调用路径：

- 传 `kn_id`
- 有对象类时传 `object_type_id`
- 能定位实例时再传 `instance_identities`
- `skill_query` 只用于当前范围内过滤，不替代上下文定位

**c. 失败后的文内引导**

快速开始末尾增加文内引导：

- 前提是否满足：见“启用前提”
- 返回空结果如何理解：见“空结果排查”
- 为什么必须按固定方式配置：见“背景说明”

#### 3.3.2 详细说明

详细说明用于承接快速开始之后的深入阅读，建议按以下顺序组织：

**a. 启用前提**

这里作为详细说明的主段，进一步解释：

- `find_skills` 不是升级 Context Loader 后自动可用的能力
- `find_skills` 只负责运行时候选发现，不负责创建 Skill
- Skill 是否可召回，不只取决于 Context Loader 版本，还取决于建模侧是否已准备完成

**b. 空结果排查**

这一段直接承接接入方最常见的问题，给出空结果解释：

- 网络级空结果：通常表示当前 Skill 不是全局生效，或知识网络下没有可见 Skill
- 对象类级空结果：通常表示该对象类没有绑定 Skill
- 实例级空结果：通常表示该实例没有命中 Skill
- 传 `skill_query` 仍为空：通常表示过滤条件过严，或当前边界内本就没有可匹配的 Skill

**c. 参数与调用方式**

这一段后置，保持简短即可：

- `kn_id` 必填
- 优先补 `object_type_id`
- 能定位实例时再补 `instance_identities`
- `skill_query` 只用于当前范围内过滤，不替代上下文定位

**d. 背景说明：为什么有这些前提**

将重要但偏底层的说明放在主流程之后，避免一开始阻塞阅读，但不能省略：

- 固定模板 / 固定承接面
- 共享只读视图
- 运行时识别键为 `object_type_id = "skills"`

建议这一段使用表格表达，每项都回答两个问题：

- 用户需要知道什么
- 这会如何影响 `find_skills` 的可用性

---

## 4. 本次调整范围

### 4.1 In Scope

- 为 `find_skills` 空结果补充面向 Agent 的文本提示
- 同步更新接口文档，明确新增响应字段语义
- 同步更新 release 文档，补齐 skill recall 使用指南和前提说明
- 验证 HTTP 和 MCP 默认输出路径都能看到该提示信息

### 4.2 Out of Scope

- 不新增新的召回模式
- 不改 Skill 创建、上传、模板生成流程
- 不改 execution-factory 对 Skill 包的管理方式
- 不改 BKN 侧对象类、关系类的落库逻辑
- 不改 Agent 侧重试策略本身，仅提供更明确的工具输出信号
- 不在本次顺手扩展其他 release 文档主题

---

## 5. 验收清单

- `find_skills` 返回空结果时，除 `entries=[]` 外，还包含明确的顶层文本提示
- 顶层文本提示能表达“当前已正常结束”以及“下一步建议”
- `entries` 中不混入伪造的 Skill 条目
- 非空结果仍保持原有主语义，不引入额外歧义
- API 文档已同步描述新增字段和空结果语义
- release 文档已明确 skill recall 的使用前提、推荐调用顺序、空结果含义
- 本次改动维持在一个 `fix/*` 分支和一个 PR 内完成

---

## 6. 失败条件

- 空结果仍然只有 `entries: []`
- 将提示文案塞进 `entries`，伪装成 Skill
- 只改实现，不同步 API 文档和 release 文档
- 在本次修复中继续扩展新的 skill recall 功能范围
- release 文档仍未说明 skill recall 的前置条件和使用边界

---

## 7. 实施顺序建议

实现时按以下顺序推进：

1. 先补能覆盖空结果行为的测试
2. 再调整 `find_skills` 响应结构与空结果文案
3. 验证空结果在 HTTP / MCP 路径下都能被调用方感知
4. 最后补 release 使用指南和接口文档

建议提交拆分为两个 commit：

1. `test+fix: add explicit empty-result hint for find_skills`
2. `docs: add skill recall usage prerequisites and guide`

---

## 8. 当前待最终确认项

当前暂无阻塞性的方案待确认项。

进入实现前，只需严格遵守本文档中已经确认的边界：

- 成功响应新增字段命名采用 `message`
- `message` 仅在空结果且成功响应时返回
- `message` 文案按已定义场景分类生成
