# BKN 外部资源支持（方案二）：ResourceType + 作用域关系类

> **状态**：草案（候选方案之一，与 [DESIGN.md](./DESIGN.md) / [DESIGN3.md](./DESIGN3.md) / [DESIGN4.md](./DESIGN4.md) / [DESIGN5.md](./DESIGN5.md) 并列评估，尚未选定）
> **版本**：0.6.0
> **日期**：2026-04-21
> **相关 Ticket**：#433
>
> **v0.6.0 变更**：状态从"已采纳"退回"草稿"。在 DESIGN4（rid logic property）与 DESIGN5（metadata-only ObjectType）两个更轻量方向被提出后，需重新横向评估是否 ResourceType 新顶层类型仍是最优路径，不再作为默认选定方案。COMPARISON.md 中的"已采纳"判定同步失效，待四方案对比重新收敛后再定。
>
> **v0.5.0 变更**：
> - `scope_binding` source 由「仅 ResourceType」放宽为「ResourceType **或** ObjectType（+可选 filter）」——基于「ResourceType 与 ObjectType+filter 在作用域语义上等价，极端条件 `filter=id=xxx` 退化为单实例」的认识（§2.3 / §3.3）；
> - ObjectType `filter` 能力由「仅作 target 侧 fallback」升级为「**双用途**」：数据加载层过滤（v0.2.0 原意）+ scope_binding source 侧粒度声明（§3.2.2 新增）；
> - 明确不引入运行时动态管理——所有 source 侧粒度差异均通过 schema 层声明完成。
>
> **v0.3.0 变更**：
> - 状态由「草案」升级为「已采纳」；
> - 新增 §4「落地节奏 / 任务拆分」（M1–M5 里程碑）；
> - 原 §4「与 DESIGN.md 对比」删除，替换为指向 COMPARISON.md 的单行指针，避免与 COMPARISON.md 对比内容重复漂移。
>
> **v0.2.0 变更**：ObjectType 数据绑定新增可选 `filter: CondCfg` 字段（借鉴 DESIGN3 §3.1），作为 `scope_binding` 在 **target 侧**的 fallback——当 `TargetRule.scope=object_type` 未显式配 `condition` 时，沿用该 ObjectType 数据绑定 filter 所筛选出的实例集合作为作用域默认范围（§3.2.1）。

---

## 1. 背景与核心思路

本文档是 [DESIGN.md](./DESIGN.md)（SkillBinding 方案）的替代方案，基于以下新认识提出：

需要与 BKN 建立关联的外部资源不只是 Skill，还包括 **App、数据流、Agent、大模型**等多种类型。如果为每种资源单独设计绑定机制（如 SkillBinding、AppBinding…），会持续累积维护成本。更合理的做法是在 BKN 内部引入一种统一的机制，让所有外部资源类型都能以一致的方式接入。核心思路是：

1. **引入新类型 ResourceType（资源类）**：与 ObjectType 平级，但语义不同——ObjectType 表示一类业务数据（多行实例），ResourceType 表示一个**独立资源对象**（一个 Skill 服务、一个 App、一条数据流、一个 Agent、一个大模型实例）。其 schema 通过绑定 Vega Resource 自动获取，不需要手动定义属性。

2. **引入作用域关系类（`scope_binding` 关系类）**：扩展现有关系类，新增一种关系类型，专门描述「某个资源（类或子集）对某个业务类型/实例的作用范围」。`scope_binding` 的 source 允许两种等价形式：**ResourceType**（整类资源静态声明）或 **ObjectType+可选 filter**（业务对象子集动态筛选），覆盖从类型级到单实例级的所有粒度。

3. **ObjectType 数据绑定支持可选 `filter`**（双用途）：
   - **数据加载层过滤**：让业务 ObjectType 自身即可声明"该类型下参与本 BKN 的实例子集"，实现"同一数据源在不同 BKN 承载不同子集"
   - **作用域表达载体**：作为 scope_binding 的 target 侧 fallback（§3.2.1），并作为 source 侧粒度声明机制（§3.2.2）
   - 借鉴 DESIGN3 §3.1，但作用对象仅限业务 ObjectType，与 ResourceType 无关

这三个概念组合，可以在 BKN 内部统一表达所有外部资源的作用域语义，复用 BKN 现有的查询和图遍历能力，新增资源类型无需修改框架。

---

## 2. 方案概览

### 2.1 ResourceType 与 ObjectType 的区别

| 维度 | ObjectType（对象类） | ResourceType（资源类） |
|------|---------------------|----------------------|
| 表示的是 | 一类业务数据，有多个行级实例 | 一个独立资源对象（服务/能力/文件）——**1 Type = 1 具体资源，非类抽象，详见 [§3.1 语义约定](#31-resourcetype资源类)** |
| 数据来源 | 绑定数据视图或 Vega Resource，行数据 | 绑定 Vega Resource，metadata schema |
| 实例语义 | 业务对象的一条记录（如一份合同） | 资源本身（如「合同审查 Skill」这个服务） |
| 参与图遍历 | 是（业务子图核心节点） | 不参与业务子图遍历 |
| 典型用途 | 合同、客户、设备… | Skill、File、外部 API… |
| 本方案新增能力 | 数据绑定支持可选 `filter: CondCfg`（§3.2）——双用途：数据加载过滤 + scope_binding source/target 侧粒度载体 | ResourceType 本身为新增顶层类型（§3.1） |

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
│  ResourceType（资源类）                       │
│  ├── schema 来自 Vega Resource metadata      │
│  └── 实例数据通过 Vega Resource 访问          │
│                                              │
│  ObjectType（业务类，现有类型，扩展 filter）   │
│  ├── data_source.filter?: CondCfg            │
│  └── 作 scope_binding source（+filter）或     │
│      target（condition 缺省时 fallback）      │
│                                              │
│  RelationType（type=scope_binding）           │
│  ├── source: ResourceType 或 ObjectType(+filter) │
│  └── target_rules: [TargetRule]             │
│       scope: kn/object_type/relation_type/  │
│              action_type/risk_type/...       │
│                                              │
│  作用域召回：查询 scope_binding 关系类        │
│  语义召回：对 ResourceType schema 做检索      │
│                                              │
│  RelationType / ActionType /                 │
│  RiskType（其余现有类型，不变）               │
└──────────────┬──────────────────────────────┘
               │ skill_ids
┌──────────────▼──────────────────────────────┐
│            context-loader                    │
│   调用 bkn-backend 获取召回结果               │
└─────────────────────────────────────────────┘
```

### 2.3 关键决策

| 决策 | 原因 |
|------|------|
| ResourceType 作为 BKN 第六种类型 | 统一框架，Skill/App/Agent 等外部资源用同一机制接入；复用 BKN 现有 schema 查询和图遍历能力 |
| Schema 从 Vega Resource 自动推导 | 不需要在 BKN 中手动维护资源属性定义；资源 schema 变更自动同步 |
| `scope_binding` 作为 RelationType 下的新 mapping type | 与 `direct`、`data_view`、`filtered_cross_join` 并列，复用 RelationType 基础设施（存储、API、schema），无需引入新的顶层类型 |
| `scope_binding` source 可为 ResourceType 或 ObjectType（+可选 filter） | ResourceType 表达「明确的资源类别」，ObjectType+filter 表达「有共性的资源子集」；极端条件（filter=id=xxx）下两者等价。由此覆盖从类型级到单实例级的所有 source 侧粒度，无需运行时数据化管理 |
| 作用域关系类 TargetRule 支持 `condition` | 表达「资源 X 作用于满足条件 C 的 ObjectType Y 实例」，覆盖目标侧实例级作用域 |
| ObjectType 数据绑定新增可选 `filter: CondCfg` | 双用途：①数据加载层过滤，支持"本 BKN 只承载实例子集"；②作为 scope_binding 在 target 侧的 fallback（`TargetRule.scope=object_type` 未显式配 `condition` 时沿用）和 source 侧粒度声明载体（§3.2） |

---

## 3. 详细设计

### 3.1 ResourceType（资源类）

ResourceType 是 BKN 中的新类型，与 ObjectType、RelationType、ActionType、RiskType 并列：

- **定义方式**：在 BKN 中声明一个 ResourceType，绑定一个 Vega Resource，schema 自动从资源 metadata 推导
- **实例语义**：ResourceType 的每个实例代表一个独立资源（如一个具体的 Skill）
- **数据访问**：实例数据通过 Vega Resource 访问，与 `data_source.type = resource` 的对象类路径一致（见 [Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)）
- **不参与业务本体**：ResourceType 不出现在业务对象图中，只通过 `scope_binding` 关系类 与业务类型建立作用域关联

> **语义约定（防止误读）：**
> ResourceType 表示的是**一个具体资源**，而不是"一类资源的抽象"。
> - 1 个 ResourceType 定义 = 1 个具体资源本身（如一个具体的"合同审查 Skill"）
> - 新接入 1 个 skill = **新定义 1 个 ResourceType**（不是给已有 ResourceType 增加一行实例）
> - 业务域若有 N 个 skill，对应 N 个 ResourceType 定义
> - 下文示例中的 `skill`/`app`/`agent` 等为**具体资源的占位名**，不是"skill 大类"——实际场景中这些位置会是 `contract_audit_skill` / `app_claim_center` 这类单资源标识

ResourceType 定义示例（概念；名称为具体资源占位）：

```
ResourceType: skill
  vega_resource: execution_factory_skill_resource
  description: 执行工厂中的 Skill 资源

ResourceType: app
  vega_resource: platform_app_resource
  description: 平台应用

ResourceType: dataflow
  vega_resource: dataflow_pipeline_resource
  description: 数据流任务

ResourceType: agent
  vega_resource: agent_resource
  description: Agent 实例

ResourceType: llm
  vega_resource: model_resource
  description: 大模型实例
```

不同的外部资源类型各自定义一个 ResourceType，绑定对应的 Vega Resource，schema 自动推导。框架本身不需要为每种资源类型做特殊处理。

### 3.2 ObjectType 数据绑定扩展 `filter`（双用途）

现有业务 ObjectType 通过 `data_source` 绑定数据视图或 Vega Resource 时，会全量加载该数据源的实例。本方案扩展绑定配置，新增可选 `filter: CondCfg` 字段，表示"该 ObjectType 在 BKN 内只承载满足条件的实例子集"：

```
ObjectType: contract
  data_source:
    type: resource
    resource_id: contract_resource
    filter?: CondCfg       # 可选，数据加载时按条件过滤
```

**基础语义：**

- `filter` 属于 ObjectType schema 的一部分（`data_source` 子结构），在数据加载阶段生效，只有满足条件的实例进入该 BKN 的该 ObjectType
- 不同 BKN 可绑定同一数据源但配置不同 `filter`，实现"按 BKN 按需承载业务数据子集"
- 未配置 `filter` 时行为与现有逻辑一致，向后兼容
- `filter` 字段依赖源数据 schema，校验与 Vega Resource / 数据视图 schema 变更策略对齐（见 §5）

**示例**：某个专注合同审查的 BKN，合同 ObjectType 只承载非过期合同：

```
ObjectType: contract
  data_source:
    type: resource
    resource_id: contract_resource
    filter: { field: status, op: neq, value: "expired" }
```

**适用范围**：**仅业务 ObjectType**。ResourceType 不引入 filter——ResourceType 是"一个独立资源对象"的抽象，类别差异应通过定义多个 ResourceType 解决，不是通过过滤实例。

> **双用途概览：** ObjectType `filter` 在本方案中承担两种角色，由其出现位置决定语义：
> - **数据加载层过滤**（本节基础语义）+ **scope_binding target 侧 fallback**（§3.2.1）——filter 作为 ObjectType 自身定义的一部分
> - **scope_binding source 侧粒度声明**（§3.2.2）——filter 出现在 scope_binding source 引用处

#### 3.2.1 作为 scope_binding target 侧 fallback 的语义

`scope_binding.target_rules` 中当 `scope="object_type"` 时支持可选 `condition`（§3.3），用于在目标 ObjectType 的基础上再加一层实例过滤。实际使用中存在一类常见需求：**「资源 X 作用于 ObjectType Y 的所有合法实例」**——此时 scope_binding 作者不希望重复把 ObjectType Y 自身的业务筛选条件写进 `condition`，也没有能力保证它与 ObjectType Y 定义方持续一致。

ObjectType `filter` 天然成为这类需求的 fallback：

- 当 `TargetRule.scope=object_type, id=Y`，**未显式给 `condition`** 时：作用范围 = Y 数据绑定 filter 筛选后的实例集合（即 BKN 内 Y 的全部实例）
- 当 `TargetRule.scope=object_type, id=Y`，**显式给了 `condition=C`** 时：作用范围 = Y 的 filter 筛选后实例 ∩ 满足 C 的实例——`condition` 是在 filter 之上的**进一步收窄**，而非覆盖
- 由于 `filter` 是在数据加载阶段生效的，不显式写 `condition` 时召回逻辑不需要额外评估；fallback 的效果来源于"数据源本身已被 filter 限定"

**三方语义对齐表：**

| ObjectType `filter` | TargetRule `condition` | 有效作用域 |
|---------------------|------------------------|-----------|
| 无 | 无 | ObjectType Y 的全部业务数据 |
| 有 | 无 | filter 筛选后的 Y 实例集合（fallback 生效） |
| 无 | 有 | Y 全部数据中满足 condition 的子集 |
| 有 | 有 | filter ∩ condition（filter 已前置生效，condition 再收窄） |

**收益：**

- 个性化 BKN 数据选择能力：不同 BKN 可按需承载业务 ObjectType 的数据子集，不需要为此修改 scope_binding 或 ResourceType
- scope_binding 作者不必重复编写 ObjectType 业务筛选条件，降低双方漂移风险
- 向后兼容：未配置 filter 的 ObjectType 行为不变

**边界：**

- `filter` 只作用于业务 ObjectType 的数据加载，与 ResourceType 无关；ResourceType 不引入 filter 配置
- `filter` 一经配置，ObjectType 对 BKN 其他能力（kn_schema_search、ontology-query、业务子图遍历等）可见的也是 filter 后的实例集合，语义对齐无缝
- 不与 `TargetRule.scope=relation_type / action_type / risk_type` 联动；那些 scope 不参照 ObjectType filter

#### 3.2.2 作为 scope_binding source 侧粒度声明的语义（v0.5.0 新增）

当 ObjectType 作为 `scope_binding` 的 source（§3.3 允许的第二种 source 形式）时，可在 source 引用处**就地**声明一个 `filter: CondCfg`，直接指定该 source 所代表的业务实例集合——从类型级（无 filter）到子集级（条件过滤）到单实例级（`filter=id=xxx`）。

**位置区别（同一 `filter: CondCfg` 字段的两处出现）：**

| 出现位置 | 语义 | 作用阶段 |
|---------|------|---------|
| `ObjectType.data_source.filter` | 数据加载层过滤 + target 侧 fallback（§3.2.1） | 数据加载阶段生效 |
| `scope_binding.source.filter`（ObjectType 形式） | 声明此 scope_binding 所指向的 source 实例子集（§3.3） | 作用域声明阶段，召回时按此集合评估 |

两处 filter 概念统一、字段复用，但语义独立：data_source.filter 影响 ObjectType 整体可见数据；scope_binding.source.filter 仅影响**该条 scope_binding** 的作用 source 范围。

**示例：**

```
# 类型级 source：整个 contract 作为 source
RelationType: skill_for_contract (type=scope_binding)
  source:
    type: object_type
    id: contract

# 子集级 source：仅对赌协议合同
RelationType: skill_for_betting_contracts (type=scope_binding)
  source:
    type: object_type
    id: contract
    filter: { field: contract_type, op: eq, value: "对赌协议" }

# 单实例级 source（退化为单例）：特定合同专属 skill 绑定
RelationType: skill_for_contract_42 (type=scope_binding)
  source:
    type: object_type
    id: contract
    filter: { field: id, op: eq, value: "contract_42" }
```

**粒度选择指引（source 形式）：**

- 资源具有**独立 schema** 且类别稳定（Skill、App、Agent…）→ 用 ResourceType
- 资源本身就是业务对象的**子集或单例**（某批合同、某个客户）→ 用 ObjectType+filter
- 需要**类型级一刀切**时两者皆可，优先 ResourceType（语义更明确）

**与运行时动态管理的界线：** 由于 filter 走 schema 变更即可覆盖单实例场景（`filter=id=xxx`），本方案**不引入运行时数据化管理**——所有 source 侧粒度差异均通过 schema 层声明完成，避免 DESIGN3 中「关系实例数据承担 schema 职责」的语义混淆。若业务出现"作用域需由运营高频编辑"的刚需，参见 [COMPARISON.md](./COMPARISON.md) §8.6 反向切换条件。

### 3.3 `scope_binding` 关系类

`scope_binding` 是 RelationType 下的新 mapping type，与 `direct`、`data_view`、`filtered_cross_join` 并列，复用 RelationType 的存储、API 和 schema 结构。

**source 类型约束：**`scope_binding` 的 source 允许两种形式——**ResourceType**，或 **ObjectType（+可选 filter，即 §3.2.2）**。其他 mapping type（`direct` 等）的 source 仍限定为 ObjectType 且不带 filter。校验逻辑按 mapping type 分支处理，RelationType 基础结构无需改动。

| source 形式 | 语义 | 典型使用 |
|------------|------|---------|
| `ResourceType(id)` | 明确的、独立建模的资源类别 | 整个 Skill 库、整个 App 库——类别稳定、对外暴露为一级概念 |
| `ObjectType(id)` + 可选 `filter: CondCfg` | 一组具有共性的业务对象子集（极端情况下退化为单实例） | 合同域 Skill（filter=contract_type=对赌）、某一具体合同（filter=id=xxx） |

两种形式在**作用域语义**上等价：scope_binding 的作用范围即 source 所代表的实例集合。差异仅在**建模粒度与管理方式**：
- ResourceType 倾向**整类资源的静态声明**，schema 独立、与业务本体解耦
- ObjectType+filter 倾向**业务实例子集的 schema 化声明**，共用业务 schema，可表达到单实例级（`filter=id=xxx`）

**结构：**

```
scope_binding mapping_rules:
  target_rules: [TargetRule]       # 一组作用域规则，一条 scope_binding 可绑定多个目标
```

每条 `TargetRule` 结构统一：

```
TargetRule:
  scope: "kn" | "object_type" | "relation_type" | "action_type" | "risk_type" | ...
  id?: string      # scope 为 "kn" 时不需要；其余 scope 必填，为对应类型的 ID
  condition?: CondCfg  # 仅 scope="object_type" 时可用，在 ObjectType filter 之上再收窄实例
```

- `scope: "kn"`：资源作用于整个 BKN，无需指定具体类型
- `scope: "object_type"` + `id`：作用于指定对象类的所有实例（沿用该 ObjectType 的 `filter`，即 target 侧 fallback，见 §3.2.1）；加 `condition` 则在 filter 之上进一步收窄
- `scope: "relation_type"` / `"action_type"` / `"risk_type"` + `id`：作用于指定的关系类 / 行动类 / 风险类
- `scope` 字段对齐 BKN 类型体系，未来新增类型时直接扩展枚举值，无需修改框架

**示例：**

```
# skill ResourceType（所有 Skill）作用于：整个 KN + 合同对象类 + 对赌协议类型合同实例 + 合同风险类
RelationType: skill_kn_scope (type=scope_binding)
  source:
    type: resource_type
    id: skill
  target_rules:
    - scope: kn
    - scope: object_type, id: contract
      # 未写 condition → fallback 到 contract ObjectType 的 data_source.filter（如 status != expired）
    - scope: object_type, id: contract, condition: { field: contract_type, op: eq, value: "对赌协议" }
      # 在 contract.filter 之上进一步收窄为"对赌协议"合同
    - scope: risk_type, id: contract_risk
```

无键匹配语义，不表示业务关联，不参与业务子图遍历。

**与 filtered_cross_join 的区别：**

`filtered_cross_join` 是「分侧过滤全连接」——两侧各自过滤后做笛卡尔积，生成业务边，语义是「建立业务关联」。`scope_binding` 是「作用域声明」——声明资源对哪些业务实体生效，不生成业务边，不参与业务子图。两者语义和用途完全不同。

### 3.4 context-loader 召回机制

引入 ResourceType 后，context-loader 的 `find_skills` 可以复用 BKN 现有的查询能力：

**作用域召回：**
- 给定 `kn_id + object_type_id + instance_identities`，通过查询 `scope_binding` 关系类 找到与该上下文关联的资源（ResourceType 实例或 ObjectType+filter source 所代表的实例子集）
- 查询路径与现有 `QueryInstanceSubgraph` 类似，沿 `scope_binding` 关系类 从 ObjectType 侧反向查找关联的资源实例
- source 为 ObjectType+filter 时：按 filter 评估 source 侧是否命中当前上下文（例如 source=`contract + filter=id=X`，则仅当上下文实例为 X 时命中）
- 命中 `TargetRule.scope=object_type` 且未写 `condition` 时，ObjectType `filter` 已在数据加载阶段生效，无需在召回路径上额外评估（见 §3.2.1）
- 命中 `TargetRule.scope=object_type` 且写了 `condition` 时，按 `condition` 对 `instance_identities` 做校验，filter 已是前置条件
- 多个 scope 层级（kn、object_type、relation_type 等）的命中结果合并返回，不做排序权重区分
- 当前 context-loader 仅传入 object_type 上下文；`relation_type` / `action_type` 作用域的召回路径已预留，待 context-loader 补充对应上下文后自然生效

**语义召回：**
- 对 ResourceType 的 `name`/`description` 字段做语义检索（与现有 kn_schema_search 机制类似）
- `name`/`description` 来自绑定的 Vega Resource metadata，语义索引由 BKN 内部维护
- ResourceType 的 schema 已在 BKN 中注册，可直接利用现有索引能力

### 3.5 作用域覆盖范围说明

`target_rules` 中 `scope: "kn"` 已可表达「作用于整个 BKN 所有类型与实例」的网络级语义，无需在 ResourceType 定义上单独标记。

跨 BKN 的 global 作用域属于低频场景，**本期不支持**，所有资源绑定要求指定具体 BKN 内创建 `scope_binding`。后续若有需要，可在平台层引入「全局 ResourceType」机制，届时再评估。

---

## 4. 落地节奏 / 任务拆分

尽管方案一步到位，实施层按以下里程碑推进，以控制单次改动范围、便于阶段性验证：

| 里程碑 | 范围 | 依赖 | 交付物 |
|--------|------|------|--------|
| **M1** | ResourceType 顶层类型定义 + 基础存储 + CRUD API；Vega Resource 绑定与 schema 自动推导 | Vega Resource 接入就绪 | ResourceType 可定义、可查询、可绑定 Vega Resource |
| **M2** | `scope_binding` mapping type 实现：RelationType schema 扩展、source 类型校验（允许 ResourceType 或 ObjectType+filter）、target_rules 结构与校验 | M1 | scope_binding 关系类可创建、可配置 source（两种形式）与 target_rules |
| **M3** | ObjectType 数据绑定 `filter` 扩展；target 侧 fallback 召回语义实现（condition 缺省时沿用 ObjectType filter）；source 为 ObjectType+filter 时的 source 侧命中评估 | M2；对现有 ObjectType 数据加载路径的影响评审 | ObjectType 可配 filter；scope_binding 两种 source 形式召回语义与 §3.2 / §3.3 对齐 |
| **M4** | context-loader 接入：作用域召回（scope_binding 关系反查）+ 语义召回（ResourceType schema）+ 合并排序 | M3；执行工厂语义搜索接口 | context-loader `find_skills` 走新路径，端到端可用 |
| **M5** | 三类资源接入验证：Skill、Agent、数据流 分别定义 ResourceType 或 ObjectType+filter、绑定 scope_binding，在真实 BKN 上验证召回与个性化场景 | M4；各上游（执行工厂 / Agent 平台 / 数据流平台）的 Vega Resource 就绪 | 三类资源在 BKN 内可被 context-loader 召回 |

**阶段性验证策略：**

- M1 / M2 / M3 在内部灰度，不对 context-loader 暴露接口，避免半成品对上层产生行为变化
- M4 完成后与 DESIGN.md 的旧路径（如本期存在）并行运行一段时间，通过对比召回结果收敛差异
- M5 按资源类型逐一开通（建议顺序：Skill → Agent → 数据流），每类资源独立验证后再开通下一类

**非目标（本期不做）：**

- 跨 BKN 的 global 作用域（§3.5）
- 运行时数据化管理作用域（由运营角色直接 CRUD 作用域数据）——如出现此刚需，参见 [COMPARISON.md](./COMPARISON.md) §8.6 反向切换条件
- Skill 执行路由、授权控制（由执行工厂侧负责）

### 4.1 与其他备选方案的对比

三方案（DESIGN / DESIGN2 / DESIGN3）的横向对比、选型理由、不选 DESIGN / DESIGN3 的依据统一见 **[COMPARISON.md](./COMPARISON.md)**（v0.5.0）。本设计文档只维护本方案自身的技术细节，不重复对比内容以避免信息漂移。

---

## 5. 风险与边界

| 风险 | 解决方案 |
|------|----------|
| global 作用域无法原生表达 | 明确本期是否需要支持；若需要，先选定平台级 ResourceType 或约定标记方案再实现 |
| ResourceType 被误当 ObjectType 使用 | 在 BKN Engine 中强制区分；ResourceType 不出现在业务子图遍历路径中 |
| Vega Resource schema 变更导致 ResourceType schema 失效 | 与 Vega Resource DESIGN.md 中的 schema 变更处理策略对齐 |
| `scope_binding` 关系类查询性能 | 按需查询，不物化边；大量 scope_binding 规则时需关注查询展开的性能影响 |
| BKN 引擎改动范围较大 | 分阶段：先实现 ResourceType 基础能力，再实现 `scope_binding` 关系类，最后叠加 ObjectType filter |
| ObjectType `filter` 引用的字段在源数据 schema 中不存在或被重命名 | 创建/更新 ObjectType 或 scope_binding 时按源数据 schema 校验 filter 字段；源数据 schema 变更时触发依赖该字段的 ObjectType / scope_binding 告警 |
| ObjectType filter 对 scope_binding 作者不透明（target 侧 fallback） | 管理界面在展示 scope_binding target_rule 时同时展示关联 ObjectType 当前 filter，避免 fallback 行为"隐式"导致的认知盲区 |
| ObjectType filter 变更影响已有 scope_binding 的作用范围 | data_source.filter 变更视为 ObjectType schema 变更，纳入变更审计；收窄 filter 时提示该 ObjectType 上挂载的 scope_binding 数量，由管理员确认 |
| scope_binding.source.filter 与 ObjectType.data_source.filter 同字段的心智混淆 | 管理界面区分两处 filter 的展示上下文；文档明确「同一 CondCfg 字段、由位置决定语义」（§3.2 / §3.2.2） |
| source 为 ObjectType+filter 时 source 侧命中评估性能 | filter 下推到 ontology-query 层评估，避免在 context-loader 做大范围实例扫描；必要时对高频 source filter 建立索引 |

---

## 6. 待决策项

在推进实现前，以下问题需要明确：

1. **ResourceType 与现有 kn_search / ontology-query 的兼容性**：ResourceType 的 schema 不应出现在业务本体检索结果中（如 ontology-query 的类型列表、kn_schema_search 结果）。需在这些接口层按类型过滤，明确哪些接口需要排除 ResourceType。
2. **各资源类型的 Vega Resource 接入顺序**：Skill、App、数据流、Agent、大模型分别依赖各自的上游服务通过 Vega 暴露资源，需与各团队对齐接入计划。
3. **ObjectType `data_source.filter` 的变更策略**：filter 收窄后，被踢出 BKN 的实例是否需要联动通知已有 scope_binding（用于审计而非数据级联）；filter 放宽时新纳入实例是否自动纳入 scope_binding 作用域（默认：是，fallback 天然生效）。
4. **ResourceType vs ObjectType+filter 的团队选型规范**：两种 source 形式语义等价但建模粒度不同；需在开发/评审环节提供明确的选型指引（§3.2.2 已给出初版），否则容易出现"同一需求两种实现"的不一致。是否在管理界面内置选型提示？

---

## 参考

- [DESIGN.md（SkillBinding 方案）](./DESIGN.md)
- [DESIGN3.md（ObjectType + 数据驱动 scope_binding 方案）](./DESIGN3.md)
- [COMPARISON.md（三方案横向对比）](./COMPARISON.md)
- [filtered_cross_join DESIGN.md](../filtered_cross_join/DESIGN.md)
- [BKN 数据绑定支持 Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)
- [bkn_docs DESIGN.md](../bkn_docs/DESIGN.md)
