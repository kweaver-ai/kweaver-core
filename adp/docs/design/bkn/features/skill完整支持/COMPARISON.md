# BKN 外部资源支持 · 五方案横向对比

> **状态**：草案（五方案并列评估，未收敛；不含推荐结论）
> **版本**：0.7.0
> **日期**：2026-04-22
> **相关 Ticket**：#433
> **对应设计**：
> - [DESIGN.md](./DESIGN.md)（SkillBinding 方案）
> - [DESIGN2.md](./DESIGN2.md)（ResourceType + scope_binding 方案）
> - [DESIGN3.md](./DESIGN3.md)（ObjectType + 数据驱动 scope_binding 方案）
> - [DESIGN4.md](./DESIGN4.md)（rid logic property + 执行工厂 Vega 数据源方案）
> - [DESIGN5.md](./DESIGN5.md)（metadata-only ObjectType role 方案）
>
> **v0.7.0 重写说明**：
> - 五方案平权呈现；所有方案同为候选，不在本文档做排除或推荐
> - 反映各 DESIGN 本轮逻辑漏洞收敛后的当前状态（DESIGN4 的 A1/A2 叠加互补、DESIGN5 的 Vega Resource 单资源粒度前提等）
> - 移除历史决策演进叙述（v0.6.0 及之前的推荐 / 撤回 / 排除历程），只呈现当前平视对比
> - 不含"综合推荐"/"采纳"结论——本文档为讨论态文档，决策由后续评审触发

---

## 1. 对比目的

为 Ticket #433（BKN 完整支持 Skill 等外部资源）提供五方案横向对比，支持后续评审的选型判断。本文档只陈述五方案在共同维度上的差异与各自前提，不做终决推荐。

**当前状态声明：** 五方案均处于讨论态，均未进入实施。读者可据本文档矩阵、前提清单、开放决策项、场景适配参考自行权衡；收敛判断由单独的选型讨论承担。

---

## 2. 五方案核心思路速览

### DESIGN.md（SkillBinding 方案）

**定位：** Skill 不进入 BKN 本体；在 BKN 外部维护独立 SkillBinding 配置表，存"skill_id + 作用域"。

- 与 BKN 平级的 SkillBinding 独立配置对象；`skill_id` 仅作引用，不复制元数据
- 支持五种作用域层级：`global` / `network` / `object_type` / `relation_type` / `action_type` / `instance_selector`
- 语义召回在 bkn-backend 内调执行工厂的服务端内部 API 完成，context-loader 不感知召回内部机制
- Skill 主数据与语义索引均在执行工厂，BKN 侧不持 Skill 内容
- **一句话定位：** 不动本体，用独立配置对象承载 Skill 作用域

### DESIGN2.md（ResourceType + scope_binding 方案）

**定位：** BKN 内部统一框架，把外部资源抽象为新顶层类型；ObjectType+filter 作为 scope_binding 的第二种 source 形式。

- 引入 **ResourceType** 作为 BKN 第六种顶层类型（与 ObjectType / RelationType / ActionType / RiskType 并列）；**1 ResourceType = 1 具体资源**，非类抽象
- schema 从 Vega Resource metadata 自动推导
- 引入 `scope_binding` 作为 RelationType 的新 mapping type，**source 可为 ResourceType 或 ObjectType(+可选 filter)**；target_rules 在 schema 中声明
- 业务 ObjectType 数据绑定支持 `filter: CondCfg`，双用途（数据加载层过滤 + scope_binding source 侧粒度声明）；target 侧支持 ObjectType filter fallback
- 粒度层级（source 侧）：类型级 → 子集级 → 单实例级（`filter=id=xxx`）
- 所有粒度差异均通过 schema 层声明完成，无运行时数据化管理
- **一句话定位：** 以"第六种类型"统一外部资源；双 source 形式覆盖从类型级到单实例级的所有粒度

### DESIGN3.md（ObjectType + 数据驱动 scope_binding 方案）

**定位：** 不扩顶层类型；资源作为现有 ObjectType 承载，作用域绑定从 schema 级下沉到实例级（数据行）。

- Skill / App 等资源复用现有 ObjectType；ObjectType 数据绑定支持 `filter` 条件
- 新增 `scope_binding` 关系类：RelationType 新 mapping type，schema 只声明 source，不声明具体 target（可选 `default_target` 兜底）
- **关系实例数据化**：每条 scope_binding 实例为 `(source_instance_identity, target_rule)`，数据存 vega dataset，通过 CRUD API 管理
- 每条 skill 实例独立绑定作用域，实例级粒度原生支持
- **一句话定位：** 不加顶层类型，作用域配置数据化，粒度原生到单实例

### DESIGN4.md（rid logic property + 执行工厂 Vega 数据源方案）

**定位：** 不扩顶层类型；资源引用通过 ObjectType 上的 `rid` logic property 表达；资源 metadata 不进 BKN，语义能力归属 Vega 层。

- 引入 **`rid` logic property**（跨系统资源引用，借鉴 Palantir 约定）作为 BKN property schema 体系的新 property type
- **rid 的两种应用位置（叠加互补，非互斥，完整方案通常同时涉及两者）：**
  - **A1 位置**（资源载体型 ObjectType，rid 作主键）：ObjectType 绑定执行工厂 Vega 数据集，BKN 侧获得资源级 metadata 索引、作用域参与、跨系统深链能力
  - **A2 位置**（业务 ObjectType 挂 rid，非主键可空）：业务对象显式引用外部资源；context-loader 可按该 rid 反查取详情
- **rid 的核心价值：** context-loader 按 rid 调 Vega 数据集反查获取资源完整 metadata（description / 参数 schema / 执行配置），BKN 侧仅持 rid 与轻元数据
- 资源 metadata 不进 BKN；BKN 只持 rid + 必要索引字段（name / category）
- **语义召回由 Vega 层承担**：每个资源数据集注册阶段声明语义实现（Vega 本地索引 / 透传上游）；降级链为 向量 → 全文 → 关键字 → 拒绝
- `scope_binding.source` 允许任意 ObjectType(+filter)；多 rid property 场景需显式声明 `rid_field` 指定抽取哪个
- **一句话定位：** 不扩类型体系；rid property 表达跨系统引用，是 context-loader 深链取详情的入口；语义能力归属 Vega 数据集

### DESIGN5.md（metadata-only ObjectType role 方案）

**定位：** 不扩顶层类型；ObjectType 新增 `role` 字段区分业务对象与资源载体；其余机制完全沿用 DESIGN2。

- 不引入 ResourceType；ObjectType 新增 `role: business | metadata_only` 字段，默认 `business`
- **`role=metadata_only` ObjectType = 一个具体资源本身**（1 Type = 1 资源，改写 ObjectType 默认"类抽象"语义）
- **Vega Resource 粒度前提**：metadata_only ObjectType 绑定的 Vega Resource 必须为单一具体资源粒度（一条 metadata 记录）；每个 metadata_only ObjectType 与 Vega Resource 为 **1:1**
- **实例身份**：metadata_only ObjectType id 兼作该资源的实例身份
- metadata_only ObjectType 承担 DESIGN2 中 ResourceType 的角色：不参与业务子图遍历、schema 从 Vega Resource metadata 自动推导、`direct`/`data_view`/`filtered_cross_join` source/target 必须 `role=business`、`scope_binding` 允许任意 role
- 语义召回在 bkn-backend 本地（以 ObjectType 为单位，N 个 metadata_only ObjectType = N 条索引项）
- **一句话定位：** 保持 DESIGN2 所有机制，仅把 ResourceType 顶层类型降级为 ObjectType 上的一个 role 标记

---

## 3. 对比维度矩阵

| 维度 | DESIGN | DESIGN2 | DESIGN3 | DESIGN4 | DESIGN5 |
|------|--------|---------|---------|---------|---------|
| **类型体系扩展** | 不扩展（BKN 外新增独立对象 SkillBinding） | 扩展：新增 ResourceType 顶层类型 | 不扩展（新增 `scope_binding` mapping type） | 不扩展（新增 `rid` property type + `scope_binding` mapping type） | 不扩展（ObjectType 新增 `role` 字段 + `scope_binding` mapping type） |
| **资源 metadata 进 BKN 范围** | 不进（仅 skill_id 引用） | 进 schema（Vega Resource metadata 推导） | 进 ObjectType（受 filter 过滤后的数据子集） | 不进重元数据（仅 rid + 轻元数据 name/category） | 进 schema（同 DESIGN2，单资源粒度 Vega Resource） |
| **语义召回路径归属** | bkn-backend 调执行工厂内部服务端 API | bkn-backend 本地（ResourceType 语义索引） | bkn-backend 本地（ObjectType 实例语义检索，复用 kn_schema_search） | **Vega 数据集**（本地索引 or 透传上游，由数据集声明） | bkn-backend 本地（metadata_only ObjectType 索引项，复用 kn_schema_search） |
| **召回链路跳数（context-loader 视角）** | 2（bkn-backend → execution-factory 取元数据） | 1（bkn-backend） | 1（bkn-backend） | 1（Vega 层内统一，BKN 不跨系统）；context-loader 按 rid 取详情为后续独立步骤 | 1（bkn-backend） |
| **实例粒度假设** | SkillBinding 每条独立绑定（实例级） | ResourceType = 1 资源；ObjectType = 多实例 | ObjectType 多实例 + 关系实例数据行（实例级） | A1 ObjectType = 多实例资源载体；A2 业务 ObjectType 挂 rid | metadata_only ObjectType = 1 资源；business ObjectType = 多实例 |
| **资源接入成本** | 新增 1 条 SkillBinding 记录（数据级，最低） | 新增 1 个 ResourceType 定义（schema 级） | 新增 1 个 ObjectType 定义 + N 条 scope_binding 关系实例 | 新增 1 个 ObjectType 定义 + 执行工厂注册 Vega 数据集 + rid 字段配置 | 新增 1 个 metadata_only ObjectType + 1 个单资源粒度 Vega Resource（N 资源 = N ObjectType + N Vega Resource） |
| **Vega Resource 粒度要求** | 不依赖 Vega Resource（直接调执行工厂） | 单资源 Vega Resource（1 ResourceType = 1 具体资源） | 业务数据源 Vega Resource（可多行） | 资源 metadata 数据集（多行，rid 作主键） | **单资源 Vega Resource（1:1 绑定）** |
| **scope_binding source 形态** | 不适用（SkillBinding 自身承担作用域，无 source/target 概念） | ResourceType **或** ObjectType(+可选 filter) | ObjectType（仅 source 声明；target 在关系实例数据行） | ObjectType(+可选 filter)，多 rid 时显式声明 `rid_field` | ObjectType(任意 role)+可选 filter |
| **隔离强度** | 最强：Skill 不进本体，物理隔离 | 强：ResourceType 与 ObjectType 类型层硬隔离 | 弱：Skill 作为 ObjectType 进业务本体，仅业务语义上区分 | 弱：资源载体 ObjectType 通过"数据源类型"默认识别；业务参与性由 ObjectType 自声明 | 弱：role 字段软隔离；框架校验入口统一拦截 |
| **框架改动面** | 小（独立表 + CRUD API + 召回 API） | 大（顶层类型 + mapping type + schema 推导） | 中（ObjectType filter 扩展 + mapping type + 关系实例数据化存储） | 中（新 property type + mapping type + 执行工厂注册 Vega 数据源） | 小（ObjectType 加 role 字段 + mapping type） |
| **数据主权** | 执行工厂独占 Skill 主数据；BKN 只存 skill_id | BKN 持资源 schema（含 metadata 推导结果），业务 metadata 不持 | BKN 持资源 ObjectType 数据（filter 后的子集） | **BKN 不持任何资源内容**，只持 rid + 轻元数据；重元数据留 Vega 数据集 | BKN 持资源 schema + 本地语义索引；重业务 metadata 留上游 |

---

## 4. 各方案的前提与落地前置条件

每方案自身的外部依赖与必须满足的落地前置条件。这些不是方案的缺点，而是方案能否成立的外部假设——选型时需确认这些前提在目标环境下是否成立。

### DESIGN（SkillBinding）

- 执行工厂必须对 bkn-backend 暴露 **服务端内部** 的 Skill 语义搜索 API（`POST /api/v1/internal/skills:search`），允许 BKN 在不感知用户上下文的情况下按自然语言查询 Skill
- 执行工厂必须为每个 Skill 维护 `skill_id` 的全局唯一且稳定的标识
- SkillBinding 独立存储层（独立表 / 独立数据服务）需要搭建
- 当扩展到非 Skill 资源（App、Agent 等）时，每种资源需单独复制 SkillBinding 结构（独立表 + 独立 CRUD API + 独立召回路径）

### DESIGN2（ResourceType）

- Vega Resource 必须能以**单资源粒度**暴露 metadata（1 ResourceType = 1 Vega Resource）；若上游系统只能提供多行数据源，需在 Vega 层先拆分
- schema 推导机制（从 Vega Resource metadata 到 ResourceType property schema）需在 BKN 侧实现
- BKN 需要承担额外顶层类型（ResourceType）的存储、API、查询、管理界面的扩展——这是**设计主动选择**，非外部依赖，但需要接受其框架改动成本

### DESIGN3（ObjectType + 数据驱动）

- BKN 需实现关系实例数据化的存储层（每条 scope_binding 关系作为数据行存于 vega dataset）
- 管理端必须提供 scope_binding CRUD API 供用户 / 运维 / 自动化流程配置每条关系实例
- 当作用域以类型级或子集级为主（不需要单实例级独立配置）时，该方案的"实例级数据化"能力不被使用，收益不兑现——需确认业务场景确有单实例动态配置需求

### DESIGN4（rid + Vega 数据源）

- 执行工厂必须注册为 Vega 数据源，并将 skill / tool / operator / agent 等暴露为 Vega 数据集（每数据集 = 资源 metadata 表，rid 作主键）
- 每个资源数据集必须在 Vega 注册阶段**声明语义检索实现**（本地索引 / 透传上游），由此承担语义召回的服务端能力
- 当透传上游时，上游外部系统必须支持 Vega 所声明的语义检索级别（或按降级链退化到关键字匹配）
- `rid` 全局唯一且不可变；执行工厂负责 rid 分配；rid 指向的资源若被删除，需配合 Vega 数据集同步

### DESIGN5（metadata-only ObjectType role）

- Vega Resource 必须能以**单资源粒度**暴露 metadata（1 metadata_only ObjectType = 1 Vega Resource，1:1）；若上游系统只能提供多行数据源，需在 Vega 层先拆分
- 每个资源的 Vega Resource metadata 必须包含可供语义检索的内容字段（典型为 name / description 或等价字段）；否则对应 ObjectType 退化为仅作用域召回
- schema 推导机制（从 Vega Resource metadata 到 ObjectType property schema）需在 BKN 侧实现（与 DESIGN2 共享此依赖）
- 框架层必须实现 role 维度的统一校验入口（ontology-query 过滤、关系类 source/target role 校验、管理界面分组）

---

## 5. 各方案的开放决策项

以下为各方案自身文档 §6 中的开放待决策项（非共享）。在方案选型之后，对应方案的开放项仍需单独收敛。

### DESIGN（SkillBinding）

- SkillBinding `skill_id` 是否允许指向非 Skill 资源（泛化为通用资源 ID），或仍限定 Skill
- 作用域层级中 `instance_selector` 的表达能力上限（简单 kv 匹配 / CondCfg / 完整查询语言）
- SkillBinding 多版本管理与生命周期策略

### DESIGN2（ResourceType）

- ResourceType 与现有 schema_search / ontology-query 的接口适配范围
- `scope_binding` 的 target fallback 策略中，未声明 `condition` 时的默认语义（是否 fallback 到 ObjectType `filter`，还是要求强制显式）
- ResourceType 在管理界面中的展示层级（与 ObjectType 同列 / 独立分组）

### DESIGN3（ObjectType + 数据驱动）

- 关系实例存储的规模上限（实例级粒度下条数可能显著大于 schema 级）
- scope_binding CRUD API 的权限模型（谁能配置、谁能批量导入）
- `default_target` 兜底与关系实例的优先级合并规则

### DESIGN4（rid + Vega）

- rid property type 是否作为独立 property type 实现（vs `string + 约束 + 元数据`）
- 资源载体型 ObjectType 的默认排除规则（管理界面覆盖机制）
- 执行工厂 Vega 数据集的字段契约（是否定义"资源数据集最小字段集"）
- Vega 数据集语义检索的 API 契约（统一接口定义）
- rid 值与 ObjectType 主键的关系约束备注（A1 主键唯一 / A2 可重复）
- Vega 数据集语义供给路径（本地索引 / 透传上游，在 Vega 数据集注册阶段声明）
- ObjectType 业务子图参与性的自声明载体（默认路由 vs 显式字段 / 管理界面覆盖）

### DESIGN5（metadata-only role）

- role 字段的正式命名（`role` / `kind` / `nature`）
- kn_schema_search 对 metadata_only 的默认可见性
- role 变更的约束级别（允许但审计 / 强制约束）
- `scope_binding` target 指向 metadata_only ObjectType 的业务用途（是否保留开放 / 补用例 / 限制为 target=business）

---

## 6. 场景适配参考（方案 → 场景）

以下按 "方案 → 该方案的前提/约束与哪些场景适配" 描述，非推荐关系。读者结合本节与 §3 矩阵、§4 前提清单、§5 开放项自行判断选型。

### DESIGN（SkillBinding）适配的典型场景

- Skill 主数据严格由执行工厂持有，BKN 方不希望感知任何 Skill 语义或 schema
- 资源类型仅限 Skill（或后续资源种类少、可接受"每种资源单独复制 SkillBinding 结构"的成本）
- 作用域配置量小且变化不频繁（每条 SkillBinding 作为数据记录维护是可控的）
- 团队对 BKN 框架改动容忍度低

### DESIGN2（ResourceType）适配的典型场景

- 外部资源类型多（Skill / Tool / Agent / LLM / App…），希望用统一机制接入
- 愿意承担顶层类型扩展的框架改动成本（存储、API、管理界面、遍历）
- 资源 schema 希望进 BKN 便于本地索引、schema 自动推导、本地语义召回
- 作用域配置以 schema 声明为主，不需要运行时动态管理
- 资源粒度能统一到"1 ResourceType = 1 具体资源"，且 Vega Resource 层能按单资源粒度组织

### DESIGN3（ObjectType + 数据驱动）适配的典型场景

- 单个资源实例需要运行时高频动态调整作用域（实例级动态配置是刚需而非边际能力）
- 希望将资源管理完全复用 ObjectType 基础能力，不引入新类型也不引入新 property type
- 作用域配置量大、需要 CRUD API 管理而非 schema 声明
- 可接受"Skill 作为 ObjectType 进业务本体"的语义混合

### DESIGN4（rid + Vega）适配的典型场景

- 资源 metadata 的数据主权要严格留在上游（执行工厂等），BKN 不沉淀上游领域数据
- 业务对象需要直接表达"引用某外部资源"（A2 位置），rid 提供 schema 级跨系统引用能力
- 接受"资源内容 context-loader 侧深链取详情"的调用形态
- 上游系统愿意注册为 Vega 数据源并承担语义检索能力（本地索引或透传），或可接受降级链兜底
- 愿意引入新 property type（rid）并配套其跨系统解析、一致性校验、管理界面跳转机制

### DESIGN5（metadata-only role）适配的典型场景

- 希望保留 DESIGN2 的资源 schema 进 BKN + 本地语义索引 + 单跳召回机制
- **不希望**扩展顶层类型体系（用 ObjectType.role 软隔离替代 ResourceType 硬隔离）
- 上游 Vega Resource 能以单资源粒度组织（1:1 基数可满足）
- 接受 metadata_only ObjectType 的"1 Type = 1 资源"语义是对 ObjectType 默认语义的改写
- 框架改动容忍度中等——可承受新增 role 字段 + role 统一校验入口 + 管理界面分组展示的改动

---

## 7. 参考

- [DESIGN.md（SkillBinding 方案）](./DESIGN.md)
- [DESIGN2.md（ResourceType + scope_binding 方案）](./DESIGN2.md)
- [DESIGN3.md（ObjectType + 数据驱动 scope_binding 方案）](./DESIGN3.md)
- [DESIGN4.md（rid logic property + 执行工厂 Vega 数据源方案）](./DESIGN4.md)
- [DESIGN5.md（metadata-only ObjectType role 方案）](./DESIGN5.md)
- [filtered_cross_join DESIGN.md](../filtered_cross_join/DESIGN.md)
- [BKN 数据绑定支持 Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)
