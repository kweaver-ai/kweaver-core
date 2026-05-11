# BKN 外部资源支持（方案四）：DESIGN2 + DESIGN3 合并 + Vega 语义外移 + rid

> **状态**：已采纳
> **版本**：1.0.0
> **日期**：2026-05-11
> **相关 Ticket**：#433
>
> **版本变更历史：**
> - **v1.0.0**（2026-05-11）：状态从「草案」升级为「已采纳」；新增落地前置条件与开放决策项章节；完善风险矩阵；更新参考文档链接
> - **v0.4.0**（2026-05-06）：完成 DESIGN2 + DESIGN3 合并逻辑；新增 rid logic property 设计；明确语义召回外移到 Vega 数据集
> - **v0.3.0**（2026-04-30）：确定核心架构；完成 scope_binding source 双路径设计
> - **v0.2.0**（2026-04-25）：初步方案构思；确定不扩展顶层类型的方向
> - **v0.1.0**（2026-04-20）：方案启动，分析 DESIGN2/DESIGN3 取舍点
>
> **承袭线（本方案不是从单一前序方案扩展而来，而是 DESIGN2 与 DESIGN3 的合并体加两项增量）：**
> - **资源承载承袭 [DESIGN3](./DESIGN3.md)**：不引入新顶层类型；资源走现有 ObjectType；ObjectType 数据绑定支持 `filter` 收窄
> - **scope_binding 形态承袭 [DESIGN2](./DESIGN2.md)**：scope_binding 是 RelationType 下的 mapping type；target_rules 在 RelationType **schema 级**声明（**不**采用 DESIGN3 的关系实例数据化）；TargetRule 结构、filter × condition 四组合语义、source/target 粒度由 filter 与 condition 双用途承担——均沿用 DESIGN2
> - **增量 1：执行工厂注册为 Vega 数据源 + 资源 metadata 不进 BKN**——执行工厂把 skill / tool / operator / agent 清单以**数据集**形式通过 Vega 暴露；BKN 资源载体 ObjectType 绑定这些数据集，实例字段 = dataset 主键 + 少量轻元数据；资源重元数据（description / 参数 schema / 执行配置）留在 Vega 数据集本地索引，context-loader 按行键直读 Vega
> - **增量 2：rid logic property**——在 BKN property schema 体系新增一种 property type `rid`（借鉴 Palantir 平台 `ri.<kind>.<...>` 约定）；rid **仅活在业务 ObjectType**，作为跨数据集引用外部资源的字段（典型：`contract.preferred_skill_rid`）；资源载体 ObjectType 自身**不**持 rid（dataset 主键已是该资源的标识）
>
> **共享内容引用（不重复描述）：**
> - `scope_binding` mapping type 定义、TargetRule 结构、作用域层级语义、filter × condition 四组合 → [DESIGN2 §3.2 / §3.2.1 / §3.3](./DESIGN2.md)
> - ObjectType `data_source.filter` 用于数据加载收窄 → [DESIGN3 §3.1](./DESIGN3.md)
> - ObjectType `data_source.filter` 双用途（数据加载 + scope source/target 粒度载体 + target fallback）→ [DESIGN2 §3.2 / §3.2.1 / §3.2.2](./DESIGN2.md)

---

## 1. 背景与核心思路

问题域与 [DESIGN2 §1](./DESIGN2.md) / [DESIGN3 §1](./DESIGN3.md) 一致：Skill / App / Agent / 数据流 / 大模型 等外部资源需要与 BKN 建立"作用域关联"，且机制需统一。

DESIGN2、DESIGN3 各自做了不同的取舍：

| 取舍维度 | DESIGN2 | DESIGN3 |
|---|---|---|
| 是否引入 ResourceType 顶层类型 | 是 | 否（走 ObjectType + filter） |
| target_rules 形态 | schema 级声明（统一治理、可审计） | 关系实例数据化（实例级独立作用域） |
| 资源 metadata 落点 | BKN 侧缓存 schema 与索引 | BKN 侧缓存被 filter 收窄后的实例数据 |

DESIGN4 选择把这两套各取其优合并，并叠加两项增量：

1. **资源承载走 DESIGN3 路径**——不扩 BKN 顶层类型，资源作为 ObjectType 承载，ObjectType `data_source.filter` 在数据加载阶段收窄；新资源接入框架成本仅"定义普通 ObjectType + 在 Vega 侧扩数据集"
2. **scope_binding 形态走 DESIGN2 路径**——target_rules schema 级声明，scope 配置可统一治理与审计；放弃 DESIGN3 实例级独立作用域的原生支持（承担机制见 §3.3）
3. **增量 执行工厂注册为 Vega 数据源 + 资源 metadata 不进 BKN**——bkn-backend 不为资源 name/description 建语义索引；语义召回外移到 Vega 数据集（本期统一 Vega 本地构建索引）；context-loader 取详情按 dataset 主键直读 Vega 行；数据主权清晰，BKN 不持资源重元数据
4. **增量 rid logic property**——为**业务 ObjectType**引入跨数据集资源引用能力（典型：`contract.preferred_skill_rid`），使业务对象能显式关联外部资源、context-loader 可顺该引用取资源详情。**rid 仅活在业务 ObjectType；资源载体 ObjectType 自身不持 rid**——这一点与"叠加 rid 主键"的直觉相反，详见 §2.1 / §3.2

**核心动机**（解释"为什么不直接走 DESIGN2 或 DESIGN3"）：
- BKN 类型体系零扩展（取自 DESIGN3）——避免顶层类型膨胀，新资源类别接入成本低
- 作用域配置可统一治理与审计（取自 DESIGN2）——避免实例级数据散落难以审计
- 数据主权清晰（增量 3）——资源 metadata 不在 BKN / Vega 缓存，执行工厂独立维护，变更不需协同同步
- 业务-资源跨数据集引用能力（增量 4）——业务对象需要显式关联外部资源时，rid 提供唯一可识别的引用载体

---

## 2. 方案概览

### 2.1 rid property 与 ObjectType 的关系

> **关键定位**：rid 是**逻辑标注**（logic property），不是 ObjectType 上的独立物理列。它是 `(kind, field)` 结构对同 ObjectType 上某个已有字段的类型化包装；id 值仍然存储在被引用的物理 field 字段里，rid 只声明"这个字段值代表 kind 类别下的资源标识"。

| 维度 | 普通 ObjectType property | **rid logic property（本方案新增）** |
|------|--------------------------|-----------------------------------|
| 形态 | 独立物理列 | **基于已有 field 字段的类型化标注**（不新增物理列） |
| 值语义 | 业务字段 | 跨数据集资源引用；按 (kind, field 字段值) 路由 + 反查 |
| 值来源 | 数据绑定直接从数据源字段读 | 框架按 (kind, instance.field 字段值) 解释为 rid；输出形式可按约定渲染为 `ri.<kind>.<ns>.<field 字段值>` |
| 值唯一性 | 视字段定义 | **不要求 ObjectType 实例集内唯一**——field 字段值多行共享是常态（典型：可观测性日志多行共享同一 producer agent id）；唯一性约束转移到被引用方（kind 对应 Vega 数据集的唯一键列） |
| 是否允许空 | 视字段定义 | 由 field 字段是否允许空决定；rid property 自身的 `required` 是业务声明的"必须存在该引用"约束 |
| 解析动作 | 无 | 框架按 `kind` 查 Vega 数据集路由表定位被引用资源数据集，按 field 字段值在该数据集唯一键列上取行 |

**rid 应用位置：业务 ObjectType 挂 rid property（仅一种）**

> rid 只活在"业务 ObjectType 跨系统引用外部资源"这一个场景。资源本身的承载（资源载体型 ObjectType 绑定执行工厂注册的 Vega 数据集）**不需要** rid——Vega 数据集自身的主键即是该资源在 BKN 内的标识，再叠一个 rid logic property 是冗余。

- 在已有业务 ObjectType 上增加一个 rid 类型字段，表达"这个业务对象关联某外部资源"
- 典型：`contract` ObjectType 增加 `preferred_skill_rid` 字段——某合同指定偏好的审查 Skill
- 用途：
  - **业务展示与跳转**：管理界面按 rid 跳转执行工厂 / Vega 数据集中对应资源详情页
  - **跨系统取详情**：context-loader 处理业务对象时按 rid 反查 Vega 数据集（执行工厂注册）取资源完整 metadata，组装跨业务-资源的复合上下文
  - **作用域召回 source（次要）**：允许业务 ObjectType 作为 scope_binding source，声明 `rid_property` 解读业务对象上的 rid 形成作用域资源标识集；相对"资源载体型 ObjectType 直接作 source"较少使用（§3.3 详述）

**资源载体型 ObjectType（不持 rid）：**

资源承载走 ObjectType 绑定执行工厂注册的 Vega 数据集，主键 = 数据集业务主键（如 `skill_id`），schema 不含 rid property。该位置的资源 metadata 取详情走 Vega 数据集主键直接读行，不经过 rid 机制。详见 §3.2.1。

### 2.2 总体架构

```text
┌───────────────────────────────────────────────┐
│              execution-factory                 │
│  ┌─────────────────────────────────────────┐  │
│  │ Skill / Tool / Operator / Agent 主数据   │  │
│  └────────────────┬────────────────────────┘  │
│                   │ 注册自身为 Vega 数据源       │
│                   ▼                            │
│  ┌─────────────────────────────────────────┐  │
│  │  Vega Dataset                            │  │
│  │  - skill_dataset     (skill_id, name, …) │  │
│  │  - tool_dataset      (tool_id,  name, …) │  │
│  │  - operator_dataset  (op_id,    name, …) │  │
│  │  - agent_dataset     (agent_id, name, …) │  │
│  └────────────────┬────────────────────────┘  │
└───────────────────┼────────────────────────────┘
                    │ 数据集绑定
┌───────────────────▼────────────────────────────┐
│                 bkn-backend                    │
│                                                │
│  ObjectType（两类）                              │
│  ├── 资源载体型：skill / tool / … ObjectType    │
│  │   data_source → 对应 Vega 数据集             │
│  │   主键 = dataset 业务主键（不持 rid）         │
│  │   可配 data_source.filter                    │
│  └── 业务型（现有）：contract / customer …      │
│      可挂 rid logic property 引用外部资源        │
│      （rid 唯一应用位置）                        │
│                                                │
│  RelationType（type=scope_binding）             │
│  source: ObjectType(+可选 filter)               │
│       （与 DESIGN2 v0.5.0 source 的第二种形式等价） │
│  target_rules: [TargetRule]                     │
│       scope: kn/object_type/relation_type/…     │
│                                                │
│  作用域召回：查询 scope_binding 命中实例集 →    │
│  载体 source: 直接输出实例集（即 Vega 行）        │
│  业务 source: 按 rid_property 解读 (kind, val)  │
│                                                │
│  语义召回：**不在 BKN 内建资源语义索引**          │
│           调 Vega 数据集语义接口                  │
│           （Vega 本地构建索引）                    │
└───────────────────┬────────────────────────────┘
                    │ 资源标识列表 (scope + 语义合并)
┌───────────────────▼────────────────────────────┐
│              context-loader                    │
│   按行键调 Vega 数据集取 metadata               │
│   (载体 source: dataset 主键; 业务 source: rid) │
│   → 组装上下文返回给 Agent                       │
└────────────────────────────────────────────────┘
```

### 2.3 关键决策

| 决策 | 原因 |
|------|------|
| 不引入 ResourceType 顶层类型 | BKN 类型体系零扩展；资源走普通 ObjectType（承袭 DESIGN3），框架改动面小于 DESIGN2 |
| rid 仅活在业务 ObjectType（不作资源载体主键） | 资源载体 ObjectType 自身已绑定 Vega 数据集，dataset 主键即资源标识；再叠 rid 是冗余、违反单一数据源原则。rid 唯一价值在"业务对象跨数据集引用资源"——业务 ObjectType 不绑定该数据集，需要显式引用载体 |
| rid 作为独立 logic property 类型，不复用 string | rid 是 `(kind, field)` 结构对同 ObjectType 已有字段的类型化标注，承载结构化元信息（kind、field 引用、解析路由规则）；普通 string property 无法承载——独立实现是必要的，不是收益取舍 |
| 资源重元数据不进 BKN（仅持轻元数据） | 避免 BKN 缓存执行工厂的领域数据；资源 metadata 变更不触发 BKN 数据同步；单一数据源原则（Palantir 风格） |
| 执行工厂注册为 Vega 数据源 | 复用 BKN 现有 Vega Resource / 数据集绑定能力；执行工厂无需为 BKN 单独开接口暴露资源清单 |
| `scope_binding.source` 仅允许 ObjectType(+filter) | 去掉 DESIGN2 的 ResourceType 分支；source 侧从类型级到单实例级的粒度仍由 ObjectType+filter 覆盖（极端条件 `filter=<dataset 主键>=xxx` 或 `filter=<rid property 的 field 字段>=xxx` 退化为单实例） |
| 语义召回外移到 Vega 数据集，不在 BKN 建索引 | BKN 不持资源 description 等 metadata，无从建索引；召回时按标识候选集 + 查询词调 Vega 数据集语义接口，数据主权清晰 |

---

## 3. 详细设计

### 3.1 rid logic property

**定义：**`rid` 是 BKN property schema 体系的一种新 property type，与 `string` / `number` / `date` / `ref` 等并列。**rid 是逻辑标注**，不新增物理列——它声明"同 ObjectType 上某个已有字段的值代表 kind 类别下的资源标识"。

**结构约定：**

```
property:
  name: xxx                        # rid property 自身名称
  type: rid                        # 新 property type
  kind: "skill" | "tool" | "operator" | "agent" | "log" | ...
  field: <同 ObjectType 上已有字段名>   # 必填，承载 id 值的物理字段
  required: true | false
```

- **`kind`**：声明该 rid 指向的资源类别；框架据此路由到对应 Vega 资源数据集（路由表见下文）
- **`field`**：声明从同 ObjectType 哪个**物理字段**读取 id 值；同一物理字段允许被多个 rid property（不同 kind）标注，由调用方按 rid property 名选择
- **`required`**：业务声明"该引用是否必须存在"；具体约束实现取决于 field 字段是否允许空
- **值解释**：框架按 `(kind, instance.field_value)` 解释为 rid；输出形式按约定渲染为 `ri.<kind>.<namespace>.<field_value>`（namespace 由 Vega 数据集注册规范定义）
- **值唯一性**：rid property 在 ObjectType 实例集**不要求唯一**；field 字段值多行共享（如可观测性日志多行共享同一 producer agent id）是基础事实
- **被引用方唯一性**：kind 对应的 Vega 资源数据集"唯一键"列必须能由 field_value 定位到一行——这是反查能成功的前提
- **rid 不可变性**：业务侧约定 field 字段值一旦写入应稳定（除非确为更换引用）；BKN 视角下 field 字段值变化等同于"引用了另一个资源"，不感知"语义同一资源的标识变更"

**框架赋予 rid property 的特殊行为：**

1. **业务对象跨系统引用外部资源（rid 的核心价值）**：业务 ObjectType（如 `contract`）通过 rid property 引用执行工厂内的资源（如 preferred Skill）——业务 ObjectType 自身不绑定执行工厂数据集，但需要表达"这个业务对象关联哪个外部资源"。rid 是这种跨数据集引用的唯一载体；context-loader 处理业务对象时按 (kind, field 字段值) 反查 Vega 数据集（执行工厂注册）取资源完整 metadata，组装跨业务-资源的复合上下文。这是 DESIGN4 相对 DESIGN3 在"业务-资源关联表达"维度上的增量
2. **管理界面**：rid property 默认渲染为可点击链接（链接的 id 部分取自 field 字段值），点击后跳转 kind 路由表对应的资源详情页
3. **scope_binding source 语义（次要）**：source 若为"含 rid property 的业务 ObjectType"，scope_binding 的作用范围即该业务 ObjectType 实例按 source 声明的 `rid_property` 解读得到的 (kind, field 字段值) 集合；filter 在该集合上做收窄（抽取规则见下文）。资源载体型 ObjectType 作 source 时直接以实例集（即 Vega 行）参与召回，不经过 rid_property 解读
4. **跨系统一致性校验**：由 Vega 数据集层周期同步执行工厂数据承担——资源被删除时 Vega 数据集移除对应行，资源载体型 ObjectType 数据加载随之失效；业务对象上 rid 引用悬空（field 字段值指向已删除资源）由业务侧自行处理。bkn-backend **不**主动反查上游；仅在调用 Vega 出错时向上透明传递错误。详见 §3.5 调用方分工矩阵 + §5 风险表"资源标识悬空"行

> **关于"资源载体 vs 业务 ObjectType 的识别"**：BKN 不在框架层做隐式分流——资源载体型与业务型 ObjectType 在 BKN 视角下统一为"普通 ObjectType"，是否进入业务子图遍历、是否在 ontology-query 中被特殊对待等，由调用方（ontology-query / kn_schema_search 等）按业务需要在调用层决定。这与"BKN 类型体系零扩展"的整体理念一致——BKN 不引入"载体 vs 业务"的隐式 role 概念。

**rid property 数量约束：**

一个业务 ObjectType 可声明**多个** rid property（典型：business ObjectType 同时挂 `preferred_skill_rid` + `backup_skill_rid`，分别 field 指向不同物理字段）。同一物理字段也可被多个 rid property（不同 kind）标注——例如 `producer_id` 可同时作为 `producer_skill_rid (kind=skill)` 与 `producer_agent_rid (kind=agent)` 的 field，由业务/调用方选择按哪种 rid property 解读。资源载体型 ObjectType 不持 rid，无此约束。

**scope_binding source 是业务 ObjectType 时的 rid 解读规则：**

> 仅在 source 为业务 ObjectType 时适用；source 为资源载体型 ObjectType 时直接以实例集参与召回，不解读 rid。

- source 业务 ObjectType 仅含 1 个 rid property：默认按该 rid property 解读，`rid_property` 可省略
- source 业务 ObjectType 含多个 rid property：`scope_binding.source` 必须显式声明 `rid_property: <property_name>` 指定按哪个 rid property 解读；未指定时配置校验拒绝
- **filter 与 rid_property 的执行顺序**：source 侧 filter **先**对业务 ObjectType 实例集做行级过滤，**再**从过滤后行集按 `rid_property` 解读得到 (kind, field 字段值) 形成资源标识集合。filter 的语义主体始终是"实例行"，与 rid_property 解读正交——filter 可写在任意物理字段上（含被 rid property 引用的 field 字段），写在 field 字段上时仍按"行级过滤"语义
- **校验时机**：scope_binding 配置创建/更新时校验 filter 字段引用是否存在于 source ObjectType schema；rid_property 校验时机同（必须为 source 业务 ObjectType 已声明的 rid 类型 property 名）；多 rid 未指定 rid_property 直接拒绝创建

**rid 与 Vega 资源数据集的承载关系：**

> 本节明确 rid 在数据模型层的对应物，避免 rid（标注）与"具体取行入口"之间的歧义。

1. **被引用方（Vega 资源数据集）的"唯一键"**：每个 Vega 资源数据集的行携带一个**唯一键列**作为该行 rid 的来源——rid 形式 `ri.<kind>.<ns>.<value>` 中的 value 即取自该列
2. **唯一键与 dataset 主键的关系**：两者**可合一**（典型推荐：资源载体型 ObjectType 的 dataset 主键 = 该数据集的 rid 唯一键，三者一体；如 `skill_dataset` 的 `skill_id` 既是 dataset 主键又是唯一键来源）也**可独立**（dataset 主键由 Vega 数据集层管理，唯一键由执行工厂业务标识维度独立声明）；具体由 Vega 数据集注册阶段声明，BKN 侧不感知
3. **kind → 数据集路由表**：每个 Vega 资源数据集在注册阶段声明自身承载的 kind（全局唯一），形成 `kind → dataset_id` 路由表（由 Vega 数据源注册规范维护）。context-loader / bkn-backend 拿到 (kind, field_value) 后按 kind 查表定位数据集，按 field_value 在数据集唯一键列上取行
4. **取行接口**：Vega 数据集对外暴露"按行键批量取行"的统一接口；行键既可以是 dataset 主键（资源载体 source 路径直读），也可以是 rid 唯一键（业务 source 反查路径）——两种入口在数据集层透明合并
5. **BKN 侧消费契约**：BKN 不感知 Vega 内部承载形态（唯一键与 dataset 主键是否合一、路由表如何存储等），仅消费两个外部契约——(a) `kind → dataset_id` 路由可用 (b) 按行键批量取行可用

> **外部依赖**：rid 唯一键的具体承载形态与 `kind → dataset_id` 路由表的具体存储由 Vega 数据源注册规范定义（不属本方案决策范围）。建议路由表全局集中维护，避免分散在各资源数据集元数据中导致 kind 冲突。

**与 ref-to-objecttype-instance 的区别：**

BKN 现有 `ref` property 表达的是"指向另一个 BKN 内实例"的引用，解析在 BKN 内部完成；rid 表达的是"指向外部系统资源"的引用，解析跨系统。两者正交，分别覆盖"本体内引用"与"跨系统引用"。

### 3.2 ObjectType 承载方式 + rid 应用位置

DESIGN4 涉及两类 ObjectType：**资源载体型**（承载执行工厂注册的资源 metadata）与**业务型**（承载业务对象，可挂 rid 引用外部资源）。rid 仅活在业务型 ObjectType 上。

#### 3.2.1 资源载体型 ObjectType（不持 rid，绑定执行工厂 Vega 数据集）

资源走 ObjectType 承载，绑定执行工厂注册的 Vega 数据集；schema 主键为数据集业务主键，**不**含 rid property。

**示例：**

```
ObjectType: skill
  data_source:
    type: vega_dataset
    dataset_id: execution_factory_skill_dataset   # 执行工厂注册到 Vega 后的数据集
    filter?: CondCfg                               # 可选（见 §3.3）
  properties:
    - name: skill_id       # 业务主键（来自数据集自身主键，非 rid）
      type: string
      is_primary: true
    - name: name           # 轻元数据：列表展示与基础检索
      type: string
    - name: category       # 轻元数据：参与 filter / scope 判定
      type: string
```

- `skill_id` 是 Vega 数据集业务主键，作为该资源在 BKN 内的标识——不需要再叠 rid logic property
- `name` / `category` 等是执行工厂在数据集中主动暴露的"轻元数据"，供 BKN 侧做基础筛选、管理界面展示
- **不包含** description、参数 schema、执行配置等重元数据——那些留在 Vega 数据集本地索引（来源于执行工厂主数据周期同步）；context-loader 取详情按 dataset 主键直接读 Vega 行（不经过 rid 反查）

tool / operator / agent 等其他资源载体 ObjectType 结构完全对称。

> **为什么不在资源载体 ObjectType 上叠 rid？** rid 是"跨数据集引用外部资源"的标识；资源载体 ObjectType 自身就绑定该资源数据集，再叠 rid 是重复——dataset 主键已经唯一标识资源行，rid 只在业务 ObjectType（不绑定该数据集）需要引用资源时才有价值。

#### 3.2.2 业务 ObjectType 挂 rid property（rid 唯一应用位置）

在现有业务 ObjectType 上增加一个 rid 类型字段，表达"这个业务对象关联某外部资源"。

**示例：**

```
ObjectType: contract                   # 业务对象类
  data_source:
    type: resource
    resource_id: contract_resource
  properties:
    - name: contract_id
      type: string
      is_primary: true
    - name: contract_type
      type: string
    - name: preferred_skill_id          # 物理字段，承载被引用 skill 的唯一键值
      type: string
      required: false
    - name: preferred_skill_rid         # rid logic property，类型化标注
      type: rid
      kind: skill
      field: preferred_skill_id         # 引用同 ObjectType 上的物理字段
      required: false
```

- rid property 是基于已有 `preferred_skill_id` 物理字段的逻辑标注——id 值仍存储在 `preferred_skill_id` 列，rid 只声明"该字段值代表 kind=skill 的资源标识"
- 管理界面看到 rid property 时，按 (kind=skill, field 字段值) 渲染为可点击链接，跳转 skill 详情
- **BKN 侧契约**：业务 ObjectType 实例查询返回时，rid property 携带类型元数据（type=rid + kind + field 引用），调用方可据此识别为"可反查标识"。**BKN 不规定**何时触发反查——具体触发模型（如 Agent 请求合同详情时是否同步展开 preferred_skill_rid）属 context-loader 行为，由 context-loader 按业务场景决定（实时展开 / 延迟展开 / 不展开），按 §3.1 行为 1 的反查机制执行
- 允许作为 scope_binding source（见 §3.3 source 约束），相对资源载体型 ObjectType 直接作 source 较少使用

### 3.3 ObjectType `data_source.filter` 与 `scope_binding`

**承袭与差异说明：**

- **承袭 [DESIGN2](./DESIGN2.md)**：scope_binding 是 RelationType 下的 mapping type；target_rules 在 RelationType **schema 级**声明；TargetRule 结构（scope/id/condition）；filter × condition 四组合作用域语义；`data_source.filter` 双用途（数据加载收窄 + scope source/target 粒度载体 + target fallback）。
- **承袭 [DESIGN3](./DESIGN3.md)**：资源走 ObjectType 承载（不扩 ResourceType 顶层类型）；`data_source.filter` 在数据加载阶段收窄实例子集；source 形式仅 ObjectType(+filter)。
- **差异（DESIGN4 增量）**：source schema 新增可选字段 `rid_property`（声明按哪个 rid property 解读，多 rid property 时必填，见 §3.1）。

**实例级独立作用域的承担机制（trade-off 说明）：**

DESIGN4 选择 schema 级 target_rules（DESIGN2 路径），放弃 DESIGN3 关系实例数据化原生支持的"单实例独立作用域"。当业务确需为单个 Skill 实例配置独立作用域时，DESIGN4 通过以下机制承担：

- **机制**：source 侧 filter 收窄到目标实例，新建一个 scope_binding RelationType 表达该单实例的作用域规则
  - 资源载体 source 路径：`filter: {field: <dataset 主键 e.g. skill_id>, op: eq, value: <唯一键值>}` 退化为单实例
  - 业务 source 路径：`filter: {field: <rid property 的 field 字段名 e.g. preferred_skill_id>, op: eq, value: <唯一键值>}` 退化为单实例
- **代价**：需要独立作用域的实例越多，RelationType 数量线性增长（每个独立作用域实例对应一个 RelationType）
- **何时接受**：单实例独立作用域是低频诉求（绝大多数 Skill 共享同类作用域规则）的场景下，schema 级配置的统一治理与审计优势超过其代价。当独立作用域实例数突破治理边界时，应重新评估是否需要降级为 DESIGN3 风格的关系实例数据化。

**与 DESIGN2 形态对照（仅列差异）：**

| 点 | DESIGN2 | **DESIGN4** |
|----|---------|-------------|
| `scope_binding.source` 可选形式 | ResourceType **或** ObjectType(+filter) | **仅** ObjectType(+filter) |
| `scope_binding.source` schema 字段 | 无 rid 解读概念（ResourceType 本身代表资源、ObjectType 源按业务 id） | **新增可选字段 `rid_property`**：仅 source 为业务 ObjectType 且含多个 rid property 时必填；source 为资源载体 ObjectType 时不适用 |
| filter 双用途 | 数据加载层 + scope_binding source/target 粒度载体 | 同 DESIGN2（无差异） |
| TargetRule 结构 | scope/id/condition | 同 DESIGN2（无差异） |
| filter × condition 四种组合作用域 | 见 DESIGN2 §3.2.1 表 | 同 DESIGN2（无差异） |

**source 约束与召回输出形式（DESIGN4）：**

`scope_binding.source` 允许指向**任意 ObjectType**，不做类别或结构层面的约束。但召回输出形式因 source ObjectType 是否为资源载体型分为两条路径：

- **source 是资源载体型 ObjectType**（典型）：filter 在 source 上做实例集收窄；scope 命中输出 = **资源载体 ObjectType 实例集**（即 Vega 数据集行集合，已含轻元数据）；context-loader 取详情按 dataset 主键直读 Vega 行——**不抽 rid，不经过 rid 反查**
- **source 是业务 ObjectType**（次要）：filter 在 source 上做业务实例集收窄；按 source 声明的 `rid_property` 解读得到 (kind, field 字段值) 形成 **资源标识集合**；context-loader 按该标识反查 Vega 数据集取资源 metadata
- target_rules 的语义不因 source 形态变化，由各 TargetRule 自身的 `scope/id/condition` 独立决定

**scope_binding 示例（DESIGN4 风格）：**

```
RelationType: skill_kn_scope (type=scope_binding)
  source:
    type: object_type
    id: skill                            # 资源载体型 ObjectType（绑定执行工厂 skill_dataset）
    # 可选 filter（与 DESIGN2 §3.2.2 同义）
    filter: { field: category, op: eq, value: "contract_audit" }
  target_rules:
    - scope: kn
    - scope: object_type, id: contract
    - scope: object_type, id: contract, condition: { field: contract_type, op: eq, value: "对赌协议" }
    - scope: risk_type, id: contract_risk
```

### 3.4 context-loader 召回机制

DESIGN4 召回流程围绕"**资源标识**"组织——具体指代视 source 形态而定：

| source 形态 | 作用域召回输出 | 资源标识形态 | 取详情走法 |
|---|---|---|---|
| 资源载体型 ObjectType（典型） | ObjectType 实例集（即 Vega 行） | **数据集主键**（如 `skill_id`） | context-loader 调 Vega 数据集按主键直读行（不经 rid 反查） |
| 业务 ObjectType（次要） | 业务实例集 → 按 `rid_property` 解读 | **(kind, field 字段值)** | context-loader 按 kind 路由数据集 + field 字段值取行 |

两条路径在 Vega 数据集对外接口层**统一**——Vega 数据集暴露"按行键批量取行"的统一接口，不区分调用方传入的是数据集主键还是 rid（rid 在 Vega 数据集行上即作为可索引列存在）。

#### 3.4.1 作用域召回

1. 给定 `kn_id + object_type_id + instance_identities`，通过查询 `scope_binding` 关系类找到命中的 source ObjectType 实例集；source 约束见 §3.3
2. **若 source 为资源载体型 ObjectType**：实例集**即**资源行集合，资源标识 = 各行的 dataset 主键
3. **若 source 为业务 ObjectType**：从过滤后实例集按 source 声明的 `rid_property` 解读得到 (kind, field 字段值) 形成资源标识集合（多 rid 时 rid_property 必填，见 §3.1）
4. TargetRule `condition` / ObjectType `filter` 的语义与 DESIGN2 §3.2.1 完全一致

#### 3.4.2 语义召回

语义能力归属 **Vega 数据集**（不在 BKN 内建资源语义索引）：

1. BKN 侧不维护资源 name/description 语义索引（资源 metadata 不在 BKN）
2. **Vega 数据集在数据同步入库后本地构建向量索引**——本期统一采用 Vega 本地索引模式，kn 可见性由 Vega 内部受控；不引入"透传上游"模式以保持降级链路简单与 SLA 可控
3. bkn-backend 拿到作用域召回的资源标识集合后，**按标识范围 + 查询词调 Vega 数据集的语义检索能力**（不直接调上游外部系统）
4. Vega 返回按语义得分排序的资源标识列表（与作用域召回输出同形态）
5. 可选：不传范围时，语义检索在该数据集全量空间内执行；跨资源类型检索需显式指定多数据集

**语义能力归属原则：**

- 语义能力挂载在**数据集**上，不挂载在 ObjectType 上
- 同一 Vega 数据集被多个 ObjectType 绑定时（例如 `skill` 资源载体 ObjectType 直接绑定 skill 数据集，业务 ObjectType 通过 rid_property 反查同一 skill 数据集），语义检索能力由数据集统一提供，各调用方共享；bkn-backend 在调用时以 dataset_id 为路由键，不以 ObjectType 为键
- 同数据集被多 ObjectType 绑定一般不常见；若出现属于业务显式需要（例如同一资源数据集需要以资源载体视角与业务引用视角两种方式被使用），各 ObjectType 按自身声明独立参与业务子图遍历 / 作用域召回，身份识别不因共享数据集而强制统一（参见 §3.1 末"关于资源载体 vs 业务 ObjectType 的识别"引导块——BKN 不在框架层做隐式分流）

#### 3.4.3 结果合并与取详情

- 作用域召回与语义召回的资源标识列表合并去重
- 合并排序规则沿用 DESIGN2 §3.4 策略（作用域权重 + 语义 score 次序）
- **按资源标识取详情**：context-loader 用合并后的标识列表调 **Vega 数据集**按行键批量取完整 metadata 行（description / 参数 schema / 执行配置等重元数据），组装完整上下文返回给 Agent
- BKN 侧仅持轻元数据（载体型 ObjectType schema 上的 name / category 等）+ 业务 ObjectType 上的 rid 引用，完整资源内容的落盘仍在 Vega 数据集，单一数据源原则由此保持

> 与 DESIGN2 / DESIGN3 在召回链路调用方分工层面的对比见 §4.1 trade-off 表；DESIGN4 内部 4 角色（bkn-backend / Vega 数据集层 / context-loader / 执行工厂）的完整分工见 §3.5 调用方分工矩阵。

### 3.5 调用方分工矩阵

DESIGN4 涉及四个角色（bkn-backend / Vega 数据集层 / context-loader / 执行工厂），各阶段分工如下：

| 阶段 | bkn-backend | Vega 数据集层 | context-loader | 执行工厂 |
|------|------------|--------------|---------------|---------|
| ObjectType 数据加载 | 配置 data_source + filter | 提供数据集查询能力 | — | 作为 Vega 数据源后端，提供原始 metadata |
| 作用域召回（scope_binding） | 主导：内部查 schema 级 target_rules，输出资源标识集（资源载体 source → dataset 主键集；业务 source → 按 rid_property 解读得 (kind, field 字段值) 集） | — | 调 bkn-backend 取标识集 | — |
| 语义召回 | 主导：拿到作用域标识集后，调 Vega 数据集语义接口（按标识范围 + 查询词），返回排序后的标识列表 | 提供数据集级语义检索（本地构建向量索引） | 调 bkn-backend 取语义命中标识列表 | 不直接参与召回；仅通过周期同步向 Vega 数据集供给 metadata |
| 结果合并与排序 | 主导：合并作用域标识 + 语义标识，去重排序后返回 | — | 接收合并后的标识列表 | — |
| metadata 取详情 | — | 提供按行键批量取行接口（行键可为 dataset 主键或 rid 唯一键，依调用方传入） | 主导：按标识批量调 Vega 数据集取完整 metadata，组装上下文 | 不直接参与；数据已在 Vega 数据集本地索引 |
| 一致性校验 | 不主动反查；仅当 Vega 调用错误时透明传递 | 通过周期同步从执行工厂拉取最新数据集；删除资源时移除对应行 | — | 资源删除事件由 Vega 同步周期感知（见 §5） |

**关键分工原则：**

- **bkn-backend 是 Vega 数据集语义接口的调用方**；**context-loader 是 Vega 数据集行查询接口的调用方**。同一 Vega 数据集对外暴露两类接口（语义检索 + 按行键取行），调用方按职责分工。
- **结果合并由 bkn-backend 完成**：作用域资源标识集与语义资源标识集在 bkn-backend 内部合并，context-loader 只接收最终资源标识列表（不感知合并过程）。
- **一致性维护由 Vega 数据集层承担**（周期同步 + 资源删除事件）；bkn-backend **不主动**做反查，只在调用 Vega 出错时向上透明传递错误（不静默吞掉）。这与 §3.1 行为 4 / §5 协同——异步告警通道的具体载体属 Vega 数据集层职责，BKN 不重复实现。

### 3.6 作用域覆盖范围说明

与 [DESIGN2 §3.5](./DESIGN2.md) 一致：`scope: "kn"` 表达网络级；跨 BKN global 本期不支持。

---

## 4. 落地节奏 / 任务拆分

| 里程碑 | 范围 | 依赖 | 交付物 |
|--------|------|------|--------|
| **M1** | rid logic property 定义：property type 注册、kind 路由表、格式校验、管理界面 renderer | 对 BKN property schema 模型的扩展评审 | 可在 ObjectType 上声明 `type=rid` 字段；管理界面可跳转 |
| **M2** | 执行工厂注册为 Vega 数据源：skill / tool / operator / agent 数据集暴露（字段 = dataset 主键 + 必要索引字段） | 执行工厂侧改造；Vega 数据源注册规范 | 4 个数据集可在 Vega 中发现、订阅 |
| **M3** | 资源载体型 ObjectType 接入：skill / tool / operator / agent ObjectType 定义与数据集绑定（不持 rid，主键 = dataset 业务主键）；ObjectType filter 支持 | M1, M2 | 4 类资源可作为 ObjectType 实例被查询 |
| **M4** | `scope_binding` mapping type 实现（仅支持 source=ObjectType(+filter)；source 业务 ObjectType 时支持 rid_property）；TargetRule 结构与校验；target 侧 fallback 语义 | M3 | scope_binding 可配置并召回（资源载体 source / 业务 source 两条路径） |
| **M5** | Vega 数据集语义检索能力（本地构建向量索引）；bkn-backend 语义召回调 Vega 路径；**bkn-backend 内部完成作用域标识集与语义标识集的合并、去重、排序，对外输出统一资源标识列表**（合并归 bkn-backend 不归 context-loader，见 §3.5 分工矩阵） | M4；Vega 语义能力就绪 | 语义召回端到端可用，bkn-backend 输出合并后的资源标识列表 |
| **M6** | context-loader 接入：调 bkn-backend 取合并后的资源标识列表 → 按行键调 Vega 数据集批量取 metadata 组装上下文；上线验证 | M5 | context-loader `find_skills` 走新路径 |

**非目标：**
- 跨 BKN global 作用域（§3.6）
- BKN 本地缓存资源 description（违反 §2.3 关键决策"资源重元数据不进 BKN"）
- 运行时数据化管理作用域（与 DESIGN2 一致，不支持）

### 4.1 与其他方案对比

五方案（DESIGN / DESIGN2 / DESIGN3 / DESIGN4 / DESIGN5）的横向对比见 **[COMPARISON.md](./COMPARISON.md)**。

DESIGN4 作为 DESIGN2 + DESIGN3 的合并体加两项增量，与最近邻 DESIGN2 / DESIGN3 的 trade-off 速览：

| 维度 | DESIGN2 | DESIGN3 | **DESIGN4（本方案）** |
|------|---------|---------|---------------------|
| 类型体系扩展 | 新增 ResourceType 顶层类型 | 不扩展（ObjectType + filter） | **不扩展**（承袭 DESIGN3） |
| 资源承载 | ResourceType | ObjectType + filter | **ObjectType + filter**（承袭 DESIGN3） |
| target_rules 形态 | schema 级声明 | 关系实例数据化（vega dataset） | **schema 级声明**（承袭 DESIGN2） |
| 实例级独立作用域 | 不原生支持 | 原生支持 | **不原生支持，filter 退化承担**（见 §3.3） |
| 作用域配置可审计性 | 高（schema 级集中） | 低（数据散落） | **高**（承袭 DESIGN2） |
| 资源 metadata 是否进 BKN | 是（schema，不含业务 metadata） | 是（filter 后实例数据） | **否**（仅资源载体 ObjectType 上 dataset 主键 + 轻元数据；DESIGN4 增量） |
| 语义召回路径 | bkn-backend 本地 ResourceType 索引 | bkn-backend 复用现有 kn_schema_search 实例级语义能力 | **bkn-backend 调 Vega 数据集**（DESIGN4 增量） |
| 资源语义索引维护方 | bkn-backend | bkn-backend | **Vega 数据集本地构建向量索引**（DESIGN4 增量） |
| 业务-资源跨数据集引用 | 无显式概念（需自建外键约定） | 无显式概念 | **rid logic property（业务 ObjectType 唯一应用）**（DESIGN4 增量） |
| 数据主权 | BKN 持资源 schema | BKN 持资源 filter 后实例 | **BKN 不持资源重元数据；资源内容唯一落点为 Vega 数据集** |
| 框架改动量 | 大（新顶层类型 + 新 mapping type + schema 推导） | 中（filter + 新 mapping type + 关系实例存储） | **中**（新 property type + 新 mapping type；复用 ObjectType + filter） |

---

## 5. 风险与边界

| 风险 | 解决方案 |
|------|----------|
| 语义召回依赖 Vega 数据集能力 | Vega 数据集本地构建向量索引；索引构建失败 / 不可用时 Vega 向 BKN 返回降级响应（空语义 + 错误码），bkn-backend 据此仅用作用域召回结果。Vega 数据集本地索引的构建周期与 SLA 由 Vega 数据源注册规范定义 |
| rid / 资源标识悬空（执行工厂删除资源） | **责任主体：Vega 数据集层**。两个维度分别承担：<br/>① **资源载体型 ObjectType 数据失效**：Vega 数据集周期同步执行工厂主数据，资源被删除时 Vega 数据集移除对应行；BKN 该 ObjectType 数据加载随之失效；scope_binding 关系自然不再命中。<br/>② **业务 ObjectType 上 rid 引用悬空**：业务对象上 rid property 所引用的 field 字段值（如 `preferred_skill_id`）指向已删除资源时，**field 字段值仍保留在业务数据源中**（rid 是逻辑标注，不是独立列，无独立"清理 rid"动作）；该字段是否清理/迁移/保留为历史快照由业务侧决定，BKN 不强制。<br/>bkn-backend 不重复实现反查任务，仅在管理界面/调用链路上透明感知失效（见 §3.5 分工矩阵） |
| rid 变更（极少发生但需准备） | 框架约定"资源唯一键永不变更"；若上游必须更换资源标识，由**执行工厂在主数据层**执行"删除旧资源 + 新建新资源"两步，Vega 数据集随同步周期反映为"旧行消失 + 新行出现"；BKN 侧仅被动观察该过程，不感知其为"语义同一资源的标识变更"——即 BKN 视角无变更概念，仅有出现/消失。引用旧资源的业务对象上 field 字段值需由业务侧自行迁移到新唯一键值 |
| rid 被误读为独立物理列（而非基于 field 字段的逻辑标注） | §2.1 / §3.1 明确：rid 是 `(kind, field)` 结构对同 ObjectType 已有字段的类型化标注，不新增物理列、不重复存储 id 值；id 值仍存储在 field 字段。误读会带来"在 ObjectType 上为 rid 单建一列存完整 rid 字符串"的设计倾向，与 logic property 定位相悖。框架在 schema 创建时校验 field 引用必须存在于同 ObjectType schema |
| rid 应用位置仅一种，常被误读为还包括资源载体型 ObjectType 主键 | §2.1 / §3.2 明确：rid 仅活在业务 ObjectType 作为跨数据集引用标注；资源载体型 ObjectType 主键由 Vega 数据集业务主键承担，**不**叠 rid。误读会带来在资源载体 ObjectType 上重复声明 rid 的设计倾向，与单一数据源原则相悖。 |
| 执行工厂作为 Vega 数据源的 schema 变更 | 与 [Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md) 对齐；索引字段（name/category）变更需协同 BKN ObjectType schema |
| kind 路由表中出现未知 kind | kind 枚举集中维护；未知 kind 的 rid property 创建时校验拦截 |
| 性能：作用域/语义命中 → 资源标识批量 → Vega 数据集批量取行 | Vega 数据集按行键批量取行接口设计；context-loader 一次批量不得超过 N（需与 Vega 约定上限） |
| **schema 级 target_rules 在大量实例独立作用域场景下导致 RelationType 数量膨胀** | DESIGN4 已知 trade-off（见 §3.3 承担机制）：每个独立作用域实例需新建一个 scope_binding RelationType + filter 收窄到该实例（资源载体 source 时 filter 写在 dataset 主键 / 轻元数据字段；业务 source 时 filter 写在 rid property 引用的 field 字段或其他业务字段）。缓解策略：业务侧限制独立作用域实例数量、按业务分类聚合作用域规则、监控 scope_binding 类型的 RelationType 总数。当膨胀突破治理边界时，重新评估是否引入 DESIGN3 风格的关系实例数据化作为补充承载 |
| rid logic property 对现有 kn_schema_search 的影响 | rid property 在语义搜索中默认不参与内容匹配（仅做类型标注）；但 kind 字段参与类型过滤 |
| Vega 数据集取行失败（资源已删除 / 上游不可达 / 数据集失联，含资源载体 dataset 主键路径与业务 source rid 反查路径） | 属 context-loader 侧错误处理范畴，BKN 不规定具体策略；BKN 侧仅保证 Vega 调用错误向上透明传递（不静默吞掉），并由 §3.1 行为 4 异步告警承接长期一致性 |

---

## 6. 术语表

| 术语 | 定义 |
|------|------|
| **rid（Resource ID）** | 逻辑属性类型，用于表达业务 ObjectType 对外部资源的跨数据集引用；结构为 `(kind, field)`，基于已有物理字段的类型化标注，不新增独立列 |
| **资源载体型 ObjectType** | 绑定执行工厂 Vega 数据集的 ObjectType，承载外部资源的轻元数据；主键为数据集业务主键，不持 rid |
| **业务 ObjectType** | 承载业务数据的现有 ObjectType；可挂 rid property 引用外部资源 |
| **scope_binding** | RelationType 下的 mapping type，描述资源对业务对象的作用范围；source 可为任意 ObjectType(+filter) |
| **TargetRule** | 作用域规则，定义 scope、id、condition 三要素，在 RelationType schema 级声明 |
| **kind** | 资源类别标识（如 skill、tool、operator、agent），用于路由到对应 Vega 数据集 |
| **作用域召回** | 通过查询 scope_binding 关系类获取资源标识集合的过程 |
| **语义召回** | 通过语义匹配获取资源标识集合的过程，由 Vega 数据集层承担 |
| **context-loader** | 调用 bkn-backend 获取资源标识列表，再调 Vega 数据集取详情组装上下文的组件 |

---

## 7. 落地前置条件

本方案的落地依赖以下外部条件，需在实施前确认：

| 前置条件 | 说明 |
|----------|------|
| 执行工厂注册为 Vega 数据源 | 执行工厂必须将 skill / tool / operator / agent 等资源暴露为 Vega 数据集（每数据集 = 资源 metadata 表） |
| 资源数据集声明语义检索实现 | 每个资源数据集必须在 Vega 注册阶段声明语义检索实现（本地索引 / 透传上游），承担语义召回能力 |
| rid 全局唯一且不可变 | `rid` 值需全局唯一且稳定；执行工厂负责 rid 分配；资源被删除时需配合 Vega 数据集同步 |
| 上游语义检索能力 | 当采用透传上游模式时，上游外部系统必须支持 Vega 所声明的语义检索级别（或按降级链退化到关键字匹配） |

---

## 8. 开放决策项（已收敛）

以下为方案开放决策项的收敛结论：

| 决策项 | 收敛结论 |
|--------|----------|
| rid property type 实现方式 | **已确定**：rid 作为独立的 logic property type 实现，与 `string`/`number`/`date`/`ref` 等并列 |
| 资源载体型 ObjectType 默认排除规则 | **不考虑**：由调用方（ontology-query / kn_schema_search 等）在调用层决定，BKN 不在框架层做隐式分流 |
| 执行工厂 Vega 数据集字段契约 | **不需要**：BKN 仅消费 Vega 数据集的外部契约（`kind → dataset_id` 路由 + 按行键批量取行），不规定内部字段结构 |
| Vega 数据集语义检索 API 契约 | **不在 BKN 范围内**：由 Vega 数据源注册规范定义，BKN 仅调用其对外暴露的统一接口 |
| rid 值与 ObjectType 主键关系约束 | **已确定**：rid 值与 ObjectType 主键没有强关联性；rid 是基于已有字段的逻辑标注，不强制与主键绑定 |
| Vega 数据集语义供给路径 | **不在 BKN 设计范围内**：由 Vega 数据集注册阶段声明，BKN 不感知其内部实现（本地索引 / 透传上游） |
| ObjectType 业务子图参与性 | **不考虑**：BKN 不在框架层做隐式分流，由调用方按业务需要在调用层决定 |

---

## 参考

- [DESIGN.md（SkillBinding 方案）](./DESIGN.md)
- [DESIGN2.md（ResourceType + scope_binding 方案）](./DESIGN2.md)
- [DESIGN3.md（ObjectType + 数据驱动 scope_binding 方案）](./DESIGN3.md)
- [DESIGN5.md（metadata-only ObjectType role 方案）](./DESIGN5.md)
- [COMPARISON.md（五方案横向对比 v0.7.0）](./COMPARISON.md)
- [filtered_cross_join DESIGN.md](../filtered_cross_join/DESIGN.md)
- [BKN 数据绑定支持 Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)
- Palantir Foundry 平台 `rid` 字段约定（作为 logic property 设计借鉴）
