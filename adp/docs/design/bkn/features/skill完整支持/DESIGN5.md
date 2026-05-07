# BKN 外部资源支持（方案五）：metadata-only ObjectType + 作用域关系类

> **状态**：草案（候选方案之一，与 [DESIGN.md](./DESIGN.md) / [DESIGN2.md](./DESIGN2.md) / [DESIGN3.md](./DESIGN3.md) / [DESIGN4.md](./DESIGN4.md) 并列评估）
> **版本**：0.1.0
> **日期**：2026-04-21
> **相关 Ticket**：#433
>
> **本方案核心差异（相对 DESIGN2）：**
> - 不引入 ResourceType 顶层类型；资源承载改为在现有 ObjectType 上增加一个 `role` 标记，区分 `business`（业务对象，默认）与 `metadata_only`（资源载体）两种角色
> - metadata-only ObjectType 沿用 DESIGN2 对 ResourceType 的所有行为约束：不参与业务子图遍历、schema 从 Vega Resource metadata 自动推导、仅通过 `scope_binding` 与业务类型建立作用域关联
> - 其余（Vega Resource 绑定、`scope_binding` mapping type、filter 双用途、召回机制、作用域层级）**完全复用 DESIGN2**
>
> **与 DESIGN2 的共享部分（不重复描述，直接引用）：**
> - Vega Resource 绑定与 schema 推导 → [DESIGN2 §3.1](./DESIGN2.md)
> - ObjectType `data_source.filter` 双用途（数据加载 + target fallback + source 粒度声明）→ [DESIGN2 §3.2 / §3.2.1 / §3.2.2](./DESIGN2.md)
> - `scope_binding` mapping type、TargetRule 结构、作用域层级 → [DESIGN2 §3.3](./DESIGN2.md)
> - context-loader 召回机制（作用域召回 + 语义召回 + 合并）→ [DESIGN2 §3.4](./DESIGN2.md)
>
> **与 DESIGN4 的分界：**
> - DESIGN4 走"资源 metadata 完全不进 BKN + rid logic property + 语义召回外调"的路径；
> - DESIGN5 保留"资源 schema 进 BKN（Vega Resource metadata 推导）+ 本地语义索引 + 单跳召回"——只在"不新增顶层类型"这一件事上与 DESIGN2 不同。

---

## 1. 背景与核心思路

问题域与 [DESIGN2 §1](./DESIGN2.md) 一致。DESIGN5 与 DESIGN2 的唯一分歧：**"资源载体"是否必须作为 BKN 顶层类型存在？**

DESIGN2 回答"是"，为此引入 ResourceType；DESIGN5 回答"否"——资源载体与业务对象在绝大多数机制上对称（schema、数据绑定、作用域关联、召回），真正差异只在两点：
- 不参与业务子图遍历
- 数据来自 Vega Resource metadata 而非业务数据视图

这两点作为 ObjectType 上的一个**角色标记**足以表达，不需要升格为顶层类型。

**核心思路：**

1. **扩展 ObjectType 定义**：新增字段 `role: "business" | "metadata_only"`，默认 `business`（向后兼容）。`metadata_only` 即承担 DESIGN2 中 ResourceType 的角色。

2. **metadata-only ObjectType 的约束由 `role` 标记派生**：
   - 不出现在业务子图遍历路径（ontology-query 按 role 过滤）
   - schema 从 Vega Resource metadata 自动推导（与 DESIGN2 ResourceType 一致）
   - 允许作为 `scope_binding.source`；与 `direct`/`data_view`/`filtered_cross_join` 的 source 语义不兼容（那些 mapping type 要求 source 是 business ObjectType）

3. **其他机制完全沿用 DESIGN2**：Vega Resource 绑定、`scope_binding` schema、filter 双用途、召回链路（作用域召回 + 本地语义召回）。

这三条组合后，DESIGN2 的语义完整性保留，但 BKN 类型体系扩展面从"新增顶层类型"降为"ObjectType 增加一个字段"。

---

## 2. 方案概览

### 2.1 metadata-only ObjectType 与业务 ObjectType 的区别

| 维度 | 业务 ObjectType（`role=business`，默认） | **metadata-only ObjectType（`role=metadata_only`）** |
|------|-------------------------------------|----------------------------------------------------|
| 表示的是 | 一类业务数据，有多个行级实例 | 一个独立资源对象（服务/能力/模型实例）——**1 metadata_only ObjectType = 1 具体资源，改写 ObjectType 默认"类抽象"语义，详见 [§3.1 语义约定](#31-metadata-only-objecttype)** |
| 数据来源 | 绑定数据视图或 Vega Resource（业务行数据） | 绑定 Vega Resource（metadata schema） |
| 实例语义 | 业务对象的一条记录（如一份合同） | 资源本身（如「合同审查 Skill」这个服务） |
| 参与图遍历 | 是（业务子图核心节点） | **否**（由 role 标记排除） |
| 参与 kn_schema_search 业务语义 | 是 | 可选：默认参与，但属于"资源池"而非"业务本体"；搜索接口按 role 分组 |
| 作为 `scope_binding.source` | 是（+可选 filter，见 DESIGN2 §3.2.2） | 是（代表 DESIGN2 ResourceType 使用场景） |
| 作为 `direct`/`data_view`/`filtered_cross_join` 的 source/target | 是 | **否**（校验拦截） |
| 典型用途 | 合同、客户、设备… | Skill、Tool、Operator、Agent、大模型… |

### 2.2 总体架构

```text
┌─────────────────────────────────────────────┐
│              execution-factory               │
│   Skill 主数据（t_skill_repository）          │
│   通过 Vega Resource 暴露 metadata schema     │
└──────────────┬──────────────────────────────┘
               │ Vega Resource 绑定
┌──────────────▼──────────────────────────────┐
│                bkn-backend                   │
│                                              │
│  ObjectType（统一类型，字段 role 区分角色）    │
│  ├── role=metadata_only                     │
│  │   (DESIGN2 ResourceType 的替代位)         │
│  │   schema 来自 Vega Resource metadata      │
│  │   不参与业务子图遍历                       │
│  │                                           │
│  └── role=business（默认）                   │
│      业务对象类，可配 data_source.filter     │
│                                              │
│  RelationType（type=scope_binding）           │
│  source: ObjectType (business 或 metadata_only) │
│          + 可选 filter                        │
│  target_rules: [TargetRule]                   │
│                                              │
│  作用域召回：查 scope_binding                 │
│  语义召回：对 metadata-only ObjectType 的      │
│           name/description 本地语义索引       │
└──────────────┬──────────────────────────────┘
               │ skill_ids / object instance ids
┌──────────────▼──────────────────────────────┐
│            context-loader                    │
│   调用 bkn-backend 获取召回结果               │
└─────────────────────────────────────────────┘
```

### 2.3 关键决策

| 决策 | 原因 |
|------|------|
| 不引入顶层 ResourceType，用 ObjectType.role 标记替代 | 类型体系零扩展；强类型隔离需求在框架内部通过 role 校验即可满足，无需暴露为顶层概念 |
| metadata-only 的所有行为约束由 role 标记派生 | 避免 DESIGN2 中"ResourceType 是什么"这一额外心智负担；新加资源类型只是新加一个 `role=metadata_only` 的 ObjectType 定义 |
| Vega Resource 绑定方式与 schema 推导沿用 DESIGN2 | 资源 schema 进 BKN（相对 DESIGN4 保留本地索引能力）；执行工厂侧无需改造数据源形式 |
| `scope_binding.source` 可为任意 role 的 ObjectType(+filter) | 语义与 [DESIGN2 v0.5.0](./DESIGN2.md) 完全一致，仅 source 形态从"ResourceType 或 ObjectType+filter"简化为"ObjectType(role=任意)+可选 filter" |
| 校验 `direct`/`data_view`/`filtered_cross_join` source 必须为 `role=business` | 保持业务关系类语义纯净，metadata-only ObjectType 不进业务图 |
| 语义召回仍在 bkn-backend 本地 | metadata-only ObjectType schema 已在 BKN 注册，可直接利用现有索引能力；不额外引入外部依赖（与 DESIGN4 区别） |

---

## 3. 详细设计

### 3.1 metadata-only ObjectType

> **语义约定（防止误读）：`metadata_only` 改写 ObjectType 默认语义**
>
> - `role=business`（默认）ObjectType：沿用现有语义 = **一类业务数据的多个行级实例**（例：ObjectType `contract` 承载 N 条合同记录）
> - `role=metadata_only` ObjectType：继承 [DESIGN2 ResourceType 语义](./DESIGN2.md#31-resourcetype资源类) = **一个具体资源本身**（例：ObjectType `contract_audit_skill` 表示「合同审查 Skill」这**一个**服务）
> - 这是 ObjectType 默认"类抽象"语义在 `metadata_only` 角色下的**改写**——同一个 ObjectType 类型结构，但"实例"指向从"多行记录"变为"一个资源整体"
> - 新接入 1 个 skill = **定义 1 个 `role=metadata_only` 的 ObjectType**（不是给某 metadata_only ObjectType 增加一行实例）；N 个 skill 对应 N 个 metadata_only ObjectType 定义
> - 下文示例中的 ObjectType 名为**具体资源标识**（如 `contract_audit_skill` / `llm_gpt4_turbo`），不是"skill/tool 大类"
>
> **Vega Resource 粒度前提：** metadata_only ObjectType 绑定的 Vega Resource 必须为"单一具体资源的 metadata（一条记录）"粒度。若上游系统将多个资源组织为一张表（例如统一的 skill 清单表），需在 Vega 数据源层按单资源粒度**拆分**为多个 Vega Resource 后方可用于 metadata_only 绑定。该前提是"1 Type = 1 资源"语义的落地承载，不满足则本方案无法直接应用该资源类型（见 §5 落地前置条件）。
>
> **基数关系：** metadata_only ObjectType 与 Vega Resource 为 **1:1**——每个 metadata_only ObjectType 独立绑定一个具体资源的 Vega Resource，不共享；这是"1 Type = 1 资源"在绑定关系层的自然推论。
>
> **实例身份：** metadata_only ObjectType "1 Type = 1 资源"前提下，**ObjectType id 本身兼作该资源的实例身份**（例如 `contract_audit_skill` ObjectType id 即表达该具体 Skill 的身份）；Vega Resource metadata 提供的内容字段（name / description 等）承担展示与语义召回用途，不承担身份语义。作用域召回命中 metadata_only 实例时，返回的标识即 ObjectType id。

在 ObjectType 定义上增加 `role` 字段：

```
ObjectType: contract_audit_skill                                         # 一个具体 Skill
  role: metadata_only                                                    # 新字段，默认 "business"
  data_source:
    type: resource
    resource_id: execution_factory_contract_audit_skill_resource         # 单资源粒度 Vega Resource
    filter?: CondCfg                                                     # 与 DESIGN2 §3.2 语义一致
  # properties 从 Vega Resource metadata 自动推导，无需手写
```

**`role` 字段语义：**

- `business`（默认）：现有 ObjectType 行为，完全向后兼容
- `metadata_only`：承担 DESIGN2 中 ResourceType 的角色；框架对其施加以下行为：
  1. **业务子图遍历排除**：ontology-query 在查询业务子图时按 role 过滤，metadata-only ObjectType 及其关系不进入遍历结果
  2. **kn_schema_search 分组**：默认参与 schema 搜索，但返回结果按 `role` 分组展示（"业务类型" vs "资源类型"）
  3. **关系类 source/target 约束**：`direct`/`data_view`/`filtered_cross_join` 类型的关系类其 source/target 必须为 `role=business`；`scope_binding` 对 source 和 target_rules 中 `scope=object_type` 的 `id` **均允许指向 metadata_only ObjectType**（不做 role 约束）
  4. **管理界面样式**：metadata_only ObjectType 在类型列表中以独立分组展示（可选折叠），降低与业务类型的视觉混淆

**metadata-only ObjectType 示例**（每个 ObjectType 标识**一个具体资源**；每个 data_source 指向**单资源粒度**的 Vega Resource）：

```
ObjectType: contract_audit_skill
  role: metadata_only
  data_source:
    type: resource
    resource_id: execution_factory_contract_audit_skill_resource

ObjectType: pdf_extract_tool
  role: metadata_only
  data_source:
    type: resource
    resource_id: execution_factory_pdf_extract_tool_resource

ObjectType: risk_scoring_operator
  role: metadata_only
  data_source:
    type: resource
    resource_id: execution_factory_risk_scoring_operator_resource

ObjectType: legal_review_agent
  role: metadata_only
  data_source:
    type: resource
    resource_id: agent_platform_legal_review_agent_resource

ObjectType: llm_gpt4_turbo
  role: metadata_only
  data_source:
    type: resource
    resource_id: model_registry_llm_gpt4_turbo_resource
```

新**具体资源**接入成本：1 条 ObjectType 定义 + 1 个单资源粒度 Vega Resource 就绪（无顶层类型改动、无新 mapping type）。

### 3.2 ObjectType `data_source.filter` 与 `scope_binding`

定义与语义**与 [DESIGN2](./DESIGN2.md) 完全相同**，本方案无差异，不重复描述。

| 条目 | DESIGN2 对应章节 | DESIGN5 是否变化 |
|------|----------------|----------------|
| `data_source.filter` 基础语义（数据加载层过滤） | [§3.2](./DESIGN2.md) | 无变化 |
| `filter` 作为 scope_binding target 侧 fallback | [§3.2.1](./DESIGN2.md) | 无变化 |
| `filter` 作为 scope_binding source 侧粒度声明 | [§3.2.2](./DESIGN2.md) | 无变化（`role=metadata_only` 的 ObjectType 也可携带 filter） |
| 三方语义对齐表（filter × condition 四种组合） | [§3.2.1 表](./DESIGN2.md) | 无变化 |
| `scope_binding` mapping type 定义 | [§3.3](./DESIGN2.md) | 无变化 |
| `scope_binding.source` 允许形式 | ResourceType **或** ObjectType(+filter) | **ObjectType(任意 role)+可选 filter**（合并为单一表达） |
| TargetRule 结构 scope/id/condition | [§3.3](./DESIGN2.md) | 无变化 |

**scope_binding 示例（DESIGN5 风格）：**

```
RelationType: skill_kn_scope (type=scope_binding)
  source:
    type: object_type
    id: skill                            # role=metadata_only 的 ObjectType
    filter?: { field: category, op: eq, value: "contract_audit" }   # 可选，同 DESIGN2 §3.2.2
  target_rules:
    - scope: kn
    - scope: object_type, id: contract
    - scope: object_type, id: contract, condition: { field: contract_type, op: eq, value: "对赌协议" }
    - scope: risk_type, id: contract_risk
```

source 不区分 business 与 metadata_only——框架只在 mapping type 校验阶段拦截"非 scope_binding 的关系类把 metadata_only 当 source/target"。

### 3.3 context-loader 召回机制

**与 [DESIGN2 §3.4](./DESIGN2.md) 完全一致，不重复描述。**

- 作用域召回：沿 `scope_binding` 关系类反查命中实例
- 语义召回：对每个 metadata_only ObjectType 绑定的 Vega Resource metadata（单条记录）中 `name` / `description` 等内容字段做本地语义检索（bkn-backend 内部索引，沿用 kn_schema_search 能力）
  - **索引粒度**：以 ObjectType 为单位——N 个 metadata_only ObjectType = N 条索引项，每条索引项对应一个具体资源（与 §3.1 "1:1 基数"和"ObjectType id 即资源实例身份"一致）
  - **前置假设**：Vega Resource metadata 推导出的 properties 必须包含可供语义检索的内容字段（典型为 `name` 与 `description` 或等价字段）。该前置在 metadata_only ObjectType 创建时由框架校验；若资源 schema 无此类字段，则该 ObjectType 不具备本地语义召回能力，退化为仅作用域召回（管理界面显式提示）
- 合并排序：同 DESIGN2 §3.4

**召回链路跳数与 DESIGN2 一致（context-loader → bkn-backend 单跳），显著低于 DESIGN4 的两跳。**

### 3.4 作用域覆盖范围说明

与 [DESIGN2 §3.5](./DESIGN2.md) 一致：`scope: "kn"` 表达网络级；跨 BKN global 本期不支持。

---

## 4. 落地节奏 / 任务拆分

| 里程碑 | 范围 | 依赖 | 交付物 |
|--------|------|------|--------|
| **M1** | ObjectType 定义扩展：新增 `role` 字段（`business`/`metadata_only`），默认 `business`；向后兼容校验 | 对 ObjectType schema 存储模型的小幅扩展评审 | ObjectType 可声明 role；存量数据兼容无迁移 |
| **M2** | 框架级 role 行为规则：业务子图遍历按 role 过滤；关系类 mapping type 的 source/target role 校验 | M1 | `role=metadata_only` 不进业务子图；`direct`/`data_view`/`filtered_cross_join` source 必须 business |
| **M3** | `scope_binding` mapping type 实现：source=ObjectType(+filter)；target_rules 结构与校验；target 侧 fallback 语义 | M2 | scope_binding 可配置并召回，覆盖所有 role 的 source |
| **M4** | ObjectType `data_source.filter` 扩展与 scope_binding source 侧粒度复用（同 DESIGN2 §3.2） | M3 | filter 双用途落地，语义对齐 DESIGN2 §3.2.1 表 |
| **M5** | context-loader 接入：作用域召回 + 本地语义召回 + 合并排序；管理界面 role 分组 | M4；执行工厂 Vega Resource 就绪 | find_skills 走新路径；管理界面 metadata_only 分组展示 |
| **M6** | Skill / Tool / Operator / Agent ObjectType 接入验证；逐类开通 | M5；各上游 Vega Resource 就绪 | 4 类资源可被 context-loader 召回 |

**非目标：**
- 跨 BKN global 作用域（§3.4）
- 运行时数据化管理作用域（与 DESIGN2 一致，不支持）
- 将 role 扩展为更多角色值（本期仅 business / metadata_only 两档）

### 4.1 与其他方案对比

与 DESIGN2 / DESIGN3 / DESIGN4 的横向对比见 **[COMPARISON.md](./COMPARISON.md)**。DESIGN5 作为新候选方案尚未纳入；如进入正式对比，应与 DESIGN4 一并扩入对比矩阵。

与 DESIGN2 / DESIGN4 的核心差异速览：

| 维度 | DESIGN2 | **DESIGN5** | DESIGN4 |
|------|---------|-------------|---------|
| 类型体系扩展 | 新增 ResourceType 顶层类型 | **不扩展**（仅 ObjectType 加 role 字段） | 不扩展（仅新增 rid property type） |
| 资源 schema 是否进 BKN | 是（Vega Resource metadata 推导） | **是**（同 DESIGN2） | 否（仅 rid + 少量索引字段） |
| 资源语义索引归属 | bkn-backend 本地 | **bkn-backend 本地**（同 DESIGN2） | execution-factory（外调） |
| 召回链路 context-loader 视角跳数 | 1 | **1** | 2 |
| 强类型隔离 | 强（ResourceType 与 ObjectType 类型层隔离） | **弱**（role 是同类型内的软隔离） | 弱（rid property + 管理界面约束） |
| 新资源接入成本 | 定义新 ResourceType | **定义新 metadata_only ObjectType**（最轻） | 新增 rid property 配置 + Vega 数据集 + ObjectType |
| schema 自动推导 | 支持 | **支持**（同 DESIGN2） | 不支持（人工定义 ObjectType schema） |

**相对 DESIGN2 的净收益：** 取消顶层类型扩展，换取强类型隔离的弱化（role 软约束）。
**相对 DESIGN4 的净收益：** 保留本地语义索引（召回少一跳）与 schema 自动推导，换取资源 schema 仍需进 BKN。

---

## 5. 风险与边界

| 风险 | 解决方案 |
|------|----------|
| role 标记的软隔离比 ResourceType 的强类型隔离弱 | 框架层统一校验入口（ontology-query、关系类 source/target 校验）；代码审查 + 自动化测试覆盖 role 判定；管理界面对 role 变更做二次确认 |
| **Vega Resource 粒度前提不被上游满足**（落地前置条件，非运行时风险） | 若上游系统不支持按单资源粒度拆分 Vega Resource（只能暴露多行 metadata 表），该资源类型无法直接通过本方案接入——需与 Vega 数据源设计 / 上游系统协同，先在 Vega 层拆分为单资源粒度 Resource 后方可绑定 metadata_only ObjectType |
| role 字段被误改（业务 ObjectType 误设为 metadata_only 或反之） | role 变更视为 schema 变更，走审计流程。约束方向如下：<br/>• `business → metadata_only`：要求该 ObjectType 无 direct/data_view/filtered_cross_join 关系引用（这些 mapping type 不允许 metadata_only 作 source/target）；**同时必须将 data_source 切换到单资源粒度 Vega Resource**（满足 §3.1 粒度前提）——原 data_source（多行业务数据）不再匹配 metadata_only 语义，原实例数据由 data_source 切换自然清除，不留残留<br/>• `metadata_only → business`：data_source 需切换到多行业务数据源（否则新 role 下数据不足）；关系类引用方面无需额外清理（scope_binding 允许任意 role 作 source/target，§3.1 行为 3） |
| 业务 ObjectType 与 metadata_only ObjectType 在管理界面混杂 | 类型列表按 role 分组展示；metadata_only 组可折叠；列表接口支持按 role 过滤 |
| kn_schema_search 返回 metadata_only 结果混入业务检索 | 搜索接口按 role 分组输出，调用方按需过滤；默认不影响 context-loader 既有语义（context-loader 明确调用 role=metadata_only 路径） |
| Vega Resource schema 变更导致 metadata_only ObjectType schema 失效 | 与 [Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md) schema 变更策略对齐（同 DESIGN2 相关风险） |
| role 模型未来扩展（出现第三种角色） | role 字段采用枚举而非布尔；本期约束为两值，但结构允许后续扩展，不需要二次 schema 变更 |
| 与 DESIGN2 迁移/共存路径 | 本期不共存——若选 DESIGN5，直接采用 role 模型；若先上 DESIGN2 再切换，需一次性把 ResourceType 实体改写为 `role=metadata_only` ObjectType（数据迁移工具） |

---

## 6. 待决策项

1. **role 字段的正式命名**：`role` / `kind` / `nature` 等候选，需与现有 ObjectType 字段命名风格对齐（避免与 rid property 的 `kind` 字段混淆——后者如果 DESIGN4 不采纳则不是问题）。
2. **kn_schema_search 对 metadata_only 的默认可见性**：默认参与搜索并分组，还是默认排除？建议默认分组参与，避免隐式差异；但 context-loader 调用链必须明确是否需要 metadata_only 结果。
3. **role 变更的约束级别**：是允许但审计，还是强制约束（必须先解除所有引用）？建议强制约束——role 变更语义过重，运行时切换风险高。
4. **`scope_binding` target 指向 metadata_only ObjectType 的业务用途**（承接 §3.1 行为 3 的开放行为）：行为层面已允许 `target_rules.scope=object_type, id=<metadata_only ObjectType>`，语义上表达"作用域是某具体资源"；典型候选场景如"编排 Skill 的作用域是某子 Skill"（资源依赖类表达）。待决策：（a）保留开放并在 §3.1 行为 3 补一个用例；（b）补一个具体需求后再开放；（c）本期限制为 target 只能指向 role=business，避免开放语义未定的路径。建议（a），与 scope_binding 的灵活性定位一致。

---

## 参考

- [DESIGN.md（SkillBinding 方案）](./DESIGN.md)
- [DESIGN2.md（ResourceType + scope_binding 方案）](./DESIGN2.md)
- [DESIGN3.md（ObjectType + 数据驱动 scope_binding 方案）](./DESIGN3.md)
- [DESIGN4.md（rid logic property + 执行工厂 Vega 数据源方案）](./DESIGN4.md)
- [COMPARISON.md（三方案横向对比 v0.5.1）](./COMPARISON.md)（尚未纳入 DESIGN4 / DESIGN5）
- [filtered_cross_join DESIGN.md](../filtered_cross_join/DESIGN.md)
- [BKN 数据绑定支持 Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)
