# BKN 外部资源支持（方案三）：扩展 ObjectType + 数据驱动的 scope_binding

> **状态**：草案
> **版本**：0.1.0
> **日期**：2026-04-21
> **相关 Ticket**：#433

---

## 1. 背景与核心思路

本文档是 [DESIGN.md](./DESIGN.md)（SkillBinding 方案）和 [DESIGN2.md](./DESIGN2.md)（ResourceType 方案）之外的第三种方案，基于以下认识：

- DESIGN.md 在 BKN 外部维护独立配置表，扩展到其他资源类型（App、Agent 等）时需每种单独建表和 API，维护成本持续累积。
- DESIGN2.md 引入 ResourceType 作为 BKN 第六种类型，框架改动大；且 scope_binding 的作用域配置是 schema 级，不支持实例级的个性化绑定。

DESIGN3 的核心思路：**不引入新顶层类型（即不在 ObjectType / RelationType / ActionType / RiskType 之外新增），完全复用现有的 ObjectType + RelationType，但将作用域绑定从 schema 级下沉到实例级（数据行）。**

具体做法：

1. **ObjectType 数据绑定支持 `filter` 过滤条件**：Skill、App 等资源作为现有 ObjectType 承载；绑定数据源（Vega Resource 或数据视图）时支持配置过滤条件，只导入需要的数据子集。
2. **新增 `scope_binding` 关系类**：RelationType 下的新 mapping type，schema 仅声明 source ObjectType，**不声明具体 target**（允许可选的 `default_target` 作为兜底）。
3. **关系实例数据化**：每条 scope_binding 的关系实例为 `(source_instance_identity, target_rule)`，数据存储在 vega dataset 中，通过 CRUD API 管理。

这样：资源管理复用 ObjectType 基础能力；作用域绑定是运行时数据而非 schema 配置，单个资源实例可独立配置自己的作用域范围；框架改动最小。

---

## 2. 方案概览

### 2.1 与 DESIGN.md / DESIGN2.md 的核心区别

| 维度 | DESIGN.md（SkillBinding） | DESIGN2.md（ResourceType） | DESIGN3.md（本方案） |
|------|--------------------------|---------------------------|--------------------|
| 资源在 BKN 中的位置 | BKN 外部独立配置表 | BKN 内部新类型 ResourceType | BKN 内部现有 ObjectType |
| 作用域配置层级 | 数据行（每条独立） | schema 级（RelationType 定义） | 数据行（关系实例） |
| 作用域粒度 | 实例级 | 类型级（ResourceType 整体） | 实例级（单个 ObjectType 实例） |
| 每资源独立作用域 | 支持 | 需定义多个 ResourceType | 原生支持 |
| BKN 引擎改动 | 小（新增独立表和 API） | 大（新类型 + 新 mapping type） | 中（扩展 ObjectType 过滤 + 新 mapping type + 关系实例存储） |

### 2.2 总体架构

```text
┌─────────────────────────────────────────────┐
│              execution-factory               │
│   Skill 主数据（t_skill_repository）          │
│   通过 Vega Resource 暴露数据                 │
└──────────────┬──────────────────────────────┘
               │ Vega Resource 绑定（可配 filter）
┌──────────────▼──────────────────────────────┐
│                bkn-backend                   │
│                                              │
│  ObjectType（现有类型，扩展 filter 能力）      │
│  ├── 数据绑定支持 filter: CondCfg            │
│  └── 实例数据按 filter 过滤后导入             │
│                                              │
│  RelationType（type=scope_binding）           │
│  ├── schema: 声明 source ObjectType          │
│  │         + 可选 default_target（兜底）      │
│  └── 关系实例存 vega dataset：               │
│       (source_instance_id, target_rule)     │
│                                              │
│  scope_binding CRUD API                     │
│  ├── 创建 / 查询 / 更新 / 删除关系实例        │
│  └── 支持按 source 或 target 检索             │
│                                              │
│  概念检索：ObjectType description/name/tag    │
│           语义匹配（复用 kn_schema_search）   │
│  作用域召回（两路合并）：                      │
│    ① 关系实例路：查询 scope_binding 实例       │
│    ② default_target 回退路：无实例数据时兜底   │
│  语义召回：命中 ObjectType 范围内实例语义检索  │
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
| scope_binding 设为 RelationType 的 mapping type，而非顶层新类型 | 避免引入第六种顶层类型，最小化框架改动；与现有 direct / data_view / filtered_cross_join 并列，类型体系完整性不受影响 |
| 资源复用 ObjectType 承载 | 复用 ObjectType 已有的 schema、数据访问、CRUD 与查询接口，无需为 Skill / App 等资源类型各自重做一套 |
| ObjectType 绑定支持 `filter` | 支持「这个 BKN 只需要某类 Skill」的个性化数据选择，无需为每个 BKN 单独建表 |
| scope_binding schema 只声明 source 和可选的 default_target | 具体 target 放到实例数据层，单个资源实例可独立配置作用域，避免 DESIGN2 的「类型级一刀切」问题；default_target 提供兜底 |
| 关系实例存 vega dataset | 复用 vega 的数据存储、查询、CRUD 基础设施，不为 scope_binding 单独建表 |
| TargetRule 结构复用 DESIGN2 | 统一表达 kn / object_type / relation_type / action_type / risk_type 等作用域层级 |

---

## 3. 详细设计

### 3.1 ObjectType 数据绑定扩展 `filter`

现有 ObjectType 绑定 Vega Resource 或数据视图时直接加载全量数据。本方案扩展绑定配置，增加 `filter?: CondCfg` 字段：

```
ObjectType: skill
  data_source:
    type: resource
    resource_id: execution_factory_skill_resource
    filter?: CondCfg       # 可选，数据加载时按条件过滤
```

- `filter` 在数据加载阶段生效，只有满足条件的实例进入 BKN
- 不同 BKN 可以绑定同一 Vega Resource 但配置不同 `filter`，实现按需导入
- 未配置 `filter` 时行为与现有逻辑一致，向后兼容

**示例**：某个专注合同审查的 BKN，仅需要合同类 Skill：

```
ObjectType: skill
  data_source:
    type: resource
    resource_id: execution_factory_skill_resource
    filter: { field: skill_category, op: eq, value: "contract" }
```

### 3.2 `scope_binding` 关系类

`scope_binding` 是 RelationType 下的新 mapping type，与 `direct`、`data_view`、`filtered_cross_join` 并列。

**与 DESIGN2 的关键区别**：

- DESIGN2：target_rules 在 RelationType schema 中声明（schema 级配置）
- DESIGN3：**schema 不声明具体 target，仅允许配置可选的 default_target 作为兜底**；具体 target 存在关系实例数据中

**Schema 结构**：

```
RelationType: skill_scope (type=scope_binding)
  source: "skill"                   # ObjectType id
  default_target?: TargetRule       # 可选，未配置实例数据时的默认作用域
```

`scope_binding` 作为特殊 mapping type，schema 不要求显式的 target 类型声明——target 由关系实例数据动态提供。`default_target` 是可选的兜底配置：当某个 source 实例没有任何关系实例数据时，使用 `default_target` 作为其默认作用域。

**关系实例结构**（存储于 vega dataset）：

```
ScopeBindingInstance:
  relation_type_id: string        # 所属 scope_binding 关系类
  source_instance_id: string      # source ObjectType 的实例 identity（如 skill_id）
  target_rule: TargetRule         # 作用域规则
```

> 上述为应用层逻辑结构；物理存储字段布局见 §3.3（`target_rule` 实际在存储层拆分为独立字段）。

每条关系实例表示「某个具体的 Skill 实例绑定到某个 TargetRule 所描述的作用域」。同一个 Skill 可对应多条关系实例，表达多个作用域目标。

**TargetRule 结构**（完全复用 DESIGN2）：

```
TargetRule:
  scope: "kn" | "object_type" | "relation_type" | "action_type" | "risk_type" | ...
  id?: string          # scope 为 "kn" 时不需要；其余 scope 必填
  condition?: CondCfg  # 仅 scope="object_type" 时可用，过滤满足条件的实例
```

语义同 DESIGN2 §3.2，不重复。

**示例**：合同审查 Skill（id=`skill_contract_review`）绑定到多个目标：

```
# 关系实例数据（存于 vega dataset）
[
  { source: "skill_contract_review", target: { scope: "kn" } },
  { source: "skill_contract_review", target: { scope: "object_type", id: "contract" } },
  { source: "skill_contract_review", target: { scope: "object_type", id: "contract",
                                                condition: { field: contract_type, op: eq, value: "对赌协议" } } },
  { source: "skill_contract_review", target: { scope: "risk_type", id: "contract_risk" } },
  { source: "skill_risk_analysis",   target: { scope: "risk_type", id: "contract_risk" } }
]
```

`skill_contract_review` 和 `skill_risk_analysis` 各自独立配置作用域，互不影响。

**语义：**无键匹配，不生成业务边，不参与业务子图遍历——仅用于作用域声明与召回。

### 3.3 关系实例数据存储与 CRUD

**存储**：scope_binding 的关系实例存储于 vega dataset，该 dataset 由 scope_binding RelationType 在创建时自动建立，作为 RelationType 的内部存储，调用方不直接感知。

物理字段布局（由 RelationType 内部决定、外部以 TargetRule 组合对象暴露）：

```
relation_type_id:    string          # 所属 scope_binding 关系类
source_instance_id:  string          # source ObjectType 实例 identity
target_scope:        string          # target_rule.scope
target_id:           string?         # target_rule.id（scope=kn 时为空）
target_condition:    JSON?           # target_rule.condition（仅 scope=object_type 可用）
```

`target_rule` 拆为 `target_scope` / `target_id` / `target_condition` 三个独立字段，以支持 §5 所述的 `(relation_type_id, target_scope, target_id)` 复合索引查询。应用层 API 仍以 TargetRule 组合对象出入参，存储层拆分对调用方透明。

**CRUD API**（bkn-backend 提供）：

| 接口 | 说明 |
|------|------|
| Create | 创建一条 scope_binding 关系实例；入参 `(relation_type_id, source_instance_id, target_rule)`，创建时校验 `source_instance_id` 在 source ObjectType 中存在 |
| Query | 按 source（给定实例 id 查其所有作用域）、按 target（给定上下文查匹配的 source 实例）、按 relation_type 等维度检索 |
| Update | 修改指定关系实例的 `target_rule` |
| Delete | 删除指定关系实例 |

**级联清理**：当 source ObjectType 实例被删除时，系统自动清理所有引用该实例的 scope_binding 关系实例；该行为由 ObjectType 实例删除事件触发，不在 Delete API 的调用范围内。作为兜底，后台任务定期扫描悬空关系实例（见 §5）。

**唯一性与重复绑定策略**：不对 `target_rule` 施加唯一性约束。允许同一 `(relation_type_id, source_instance_id)` 下挂多条 target_rule 相同或重叠的关系实例；召回时对匹配结果取并集（按 `source_instance_id` 去重），因此重复不影响正确性，仅影响存储与查询开销。管理界面可按需在应用层去重展示。

**权限控制**：scope_binding 关系实例的 CRUD 沿用 BKN 现有管理后台权限模型，不新增独立角色。

### 3.4 context-loader 召回机制

context-loader 不感知具体 ObjectType 的业务定位，通过**概念检索 → 实例检索**两步来定位资源：

1. **概念检索**：以 context-loader 接收到的查询/意图文本作为 query，对 ObjectType 的 `description` / `name` / `tag` 等 schema 元数据做语义匹配，识别「资源类 ObjectType」（如 skill、app、agent）。此步骤复用现有的 `kn_schema_search` 能力（同 DESIGN2 §3.3），无需新增索引基础设施
2. **实例检索**：在命中的 ObjectType 范围内，结合作用域召回和语义召回定位具体实例

**作用域召回**采用两路查询后合并：

1. **关系实例路**：给定 `kn_id + object_type_id + instance_identities`（即当前上下文锁定的对象实例唯一标识集合，用于评估 `target_rule.condition` 是否命中），按 target 维度查询 scope_binding 关系实例表；匹配规则为 `target_rule.scope` 与上下文层级对齐，`id` / `condition` 进一步收窄
2. **default_target 回退路**：对于 source ObjectType 中**没有任何关系实例数据**的实例，若 scope_binding schema 的 `default_target` 匹配当前上下文，将这些实例一并纳入结果。兜底触发条件是该 source 实例**全局无任何关系实例数据**，而非本次查询未命中——这使得语义可预测、避免单实例的多查询结果交叉污染

两路结果合并后返回 `source_instance_id` 集合（即 Skill ID）。其他说明：

- 多个 scope 层级（kn、object_type、relation_type 等）的命中结果合并返回，不做排序权重区分
- 当前 context-loader 仅传入 object_type 上下文；`relation_type` / `action_type` 作用域召回路径已预留

**语义召回：**
- 在概念检索命中的 ObjectType 范围内，对其实例数据做语义检索，复用 BKN 现有的实例级语义检索能力（已存在，无需新增）
- 检索字段为实例的 name / description 等
- 不依赖新增的索引基础设施

### 3.5 作用域覆盖范围说明

target_rule 中 `scope: "kn"` 表达「作用于当前 BKN 所有类型与实例」的网络级语义。

跨 BKN 的 global 作用域属于低频场景，**本期不支持**。所有 scope_binding 关系实例归属于某个 BKN，跨 BKN 绑定需在各 BKN 内分别创建。

---

## 4. 与 DESIGN.md / DESIGN2.md 方案对比

| 维度 | DESIGN.md | DESIGN2.md | DESIGN3.md |
|------|-----------|-----------|-----------|
| 资源承载 | BKN 外部独立表 | BKN 内 ResourceType | BKN 内现有 ObjectType |
| 作用域配置 | 独立 t_skill_binding 表 | RelationType schema（target_rules） | RelationType 实例数据（vega dataset） |
| 作用域粒度 | 实例级 | 类型级 | 实例级 |
| 个性化数据选择 | N/A（数据不在 BKN） | 需多 ResourceType | ObjectType `filter` 配置 |
| 扩展到其他资源 | 每种单独建表 | 新建 ResourceType 定义 | 新建 ObjectType + 新建 scope_binding RelationType + 按需填关系实例数据 |
| BKN 引擎改动 | 小 | 大 | 中 |
| 框架一致性 | 低（外挂配置） | 中（新增第六类型） | 较高（类型体系复用，但引入新 mapping type 和关系实例存储机制） |
| 查询能力复用 | 需新 API | 部分复用 | 完全复用 ObjectType 能力 |

---

## 5. 风险与边界

| 风险 | 解决方案 |
|------|----------|
| Skill 作为 ObjectType 混入业务本体，污染业务语义 | context-loader 复用 `kn_schema_search` 对 ObjectType description / name / tag 做语义匹配识别资源类，依赖现有语义匹配质量；ontology-query 等接口按需过滤（见 §6） |
| `filter` 配置对同一 Vega Resource 多 BKN 引用时语义混乱 | 各 BKN 独立加载各自过滤后的实例子集；Vega Resource 层保持只读，不受 filter 影响 |
| scope_binding 关系实例量大导致查询性能下降 | vega dataset 层建立 `(relation_type_id, target_scope, target_id)` 索引；按需查询，不物化 |
| `default_target` 回退路径的性能成本 | 兜底候选集 = source ObjectType 全局无关系实例数据的实例集合，该集合可缓存/索引（关系实例表变更时失效），非每次查询重算；实例量大的 ObjectType 仍需评估是否启用 default_target，或限定 default_target 只支持较粗粒度的 scope（如 kn）以降低匹配开销 |
| source ObjectType 实例被删除后残留关系实例 | 删除 ObjectType 实例时级联清理；后台任务定期检测悬空关系实例 |
| Vega Resource schema 变更导致 filter 条件失效 | 与 Vega Resource DESIGN.md 的 schema 变更策略对齐；filter 字段变更时告警 |
| 同一 Skill 实例重复配置作用域 | 不做唯一性约束；召回取并集后按 `source_instance_id` 去重，重复实例不影响结果正确性，仅带来存储/查询开销（见 §3.3） |

---

## 6. 待决策项

1. **ontology-query 排除 scope_binding 的具体实现路径**：scope_binding 不参与业务子图遍历（§3.2 已定调），但具体实现路径待决策——是 ontology-query 默认排除所有 `type=scope_binding` 的 RelationType，还是要求调用方显式声明排除/包含？影响 API 契约设计。
2. **Skill 作为 ObjectType 的命名与展示策略**：管理界面如何区分「业务对象类」与「资源对象类」，避免用户混淆。context-loader 定位已复用 `kn_schema_search`（见 §5），管理界面是否需要更强的区分机制待定。
3. **ObjectType `filter` 的校验与失效处理**：filter 字段引用的字段若在源数据 schema 中不存在或被重命名，如何检测与告警。

---

## 参考

- [DESIGN.md（SkillBinding 方案）](./DESIGN.md)
- [DESIGN2.md（ResourceType 方案）](./DESIGN2.md)
- [filtered_cross_join DESIGN.md](../filtered_cross_join/DESIGN.md)
- [BKN 数据绑定支持 Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)
- [bkn_docs DESIGN.md](../bkn_docs/DESIGN.md)
