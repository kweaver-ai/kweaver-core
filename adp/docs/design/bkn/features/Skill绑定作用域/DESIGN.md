# BKN Skill 绑定作用域声明 技术设计文档

> **状态**：草案  
> **版本**：0.1.0  
> **日期**：2026-04-13  
> **相关 Ticket**：待定（关联 context-loader `find_skills` / Issue #109 演进）

---

## 1. 背景与目标 (Context & Goals)

Skill 元数据通过数据源与 resource 映射，在 BKN 中以固定对象类 `skills`（`object_type_id = "skills"`）的实例形式存在；智能体通过 **Skill 与业务知识网络中对象/关系** 的绑定信息，在运行时召回适用技能。当前存在两类张力：

- **建模张力**：`skills` 与业务对象类之间无法像普通业务概念一样用「自然」的跨域继承表达联系，实际往往通过 **分侧过滤全连接（`filtered_cross_join`）** 等关系类型落地。
- **语义张力**：`context-loader` 中 `find_skills` 对 **网络级** Skill 采用 **临时规则**——仅当 `skills` 与任意其他 Object Type **均不存在** RelationType 时，才允许网络级召回（`QueryObjectInstances`）；一旦存在任意绑定，网络级路径即失效，**无法在同一知识网络内同时存在「全网 Skill」与「按对象类/实例绑定 Skill」**。

**目标**：

- 在 **BKN 绑定侧** 显式声明 Skill 的作用域层级：**知识网络级（KN）**、**对象类级**、**实例级**，而不要求在 Skill 主数据（如 `SKILL.md`）上预填 scope 枚举。
- 与现有召回分层（`matched_scope`: `network` / `object_type` / `object_selector`）语义对齐，便于上游 Agent 理解与合并结果。
- 为后续废弃「零关联即全网」的推断规则提供可迁移路径。

**非目标**：

- 不定义 Skill 的创建/删除/更新主流程（归属 execution-factory）。
- 不展开 `find_skills` 的 HTTP 契约与实现细节（见 context-loader 专项设计）。
- 不替代现有分侧过滤全连接在「对象类 ↔ skills」上的建模能力。

---

## 2. 方案概览 (High-Level Design)

### 2.1 核心思路

采用 **内置「网络锚点」对象类（KN Scope Anchor）+ 关系拓扑表达作用域**：

1. **知识网络级**：引入系统内置 Object Type（建议 `object_type_id = "kn_scope"`），每个业务知识网络 **恰好一个** 锚点实例；通过 **RelationType `kn_scope` ↔ `skills`** 将需要「整网可用」的 Skill 实例连到该锚点。**配置者只需「把 Skill 挂到网络根」**，无需在 Skill 侧填写 scope。
2. **对象类级**：沿用 **业务 Object Type ↔ `skills`** 的 RelationType（含分侧过滤全连接）；召回时沿 **类级路径** 进入子图，`matched_scope = object_type`。
3. **实例级**：仍使用 **业务 OT ↔ `skills`** 的同一套 RelationType，在实例层建立 **具体业务实例 → Skill 实例** 的边；请求携带实例锚点时走实例级子图，`matched_scope = object_selector`。

**Scope 的声明位置**：落在 **边与关系类型所连接的两端类型** 上，而不是 Skill 内容资产上的必填字段。

### 2.2 总体架构

```text
                    上游 Agent
                        |
              context-loader / find_skills
                        |
        +---------------+---------------+
        |  BKN 模式层   |  实例子图 / 实例查询  |
        | (RelationType)|  (ontology-query)    |
        +-------+-------+---------+------------+
                |                 |
    kn_scope OT |                 | 业务 OT
    (单例/每 KN) |                 | (contract, ...)
         \      |                 /      \
          \     |                /        \
       RelationType          RelationType  ...
       kn_scope↔skills       business_ot↔skills
                \                 /
                 \               /
                  Skill 实例 (skills OT)
```

- **全网 Skill**：仅出现在 **`kn_scope` 实例与 Skill 实例之间的边** 上。
- **按类 / 按实例 Skill**：出现在 **`业务 OT` 与 `skills` 关系** 下的子图中；实例级与对象类级的区分由 **查询起点（类查询 vs 实例锚点）** 决定，与现有三层召回模式一致。

### 2.3 关键决策

1. **锚点对象类固定 ID**：`kn_scope`（或统一前缀的保留 ID，全平台一致），避免与业务 OT 命名冲突；在 BKN 产品文档中列为 **系统内置类型**。
2. **每 KN 单例**：每个 `kn_id` 下 `kn_scope` 实例数 = 1（或由平台在创建 KN 时自动创建），作为 **唯一「网络根」**。
3. **全网不再依赖「skills 与任何 OT 均无 RelationType」**：网络级召回改为 **显式沿 `kn_scope → skills` 路径**；旧规则可作为过渡期兼容，最终废弃。
4. **合并策略**：同一召回请求内，将 **网络锚点路径** 与 **业务 OT 路径** 的结果 **去重合并**，优先级建议：**实例级 > 对象类级 > 网络级**（与既有 `Priority` 设计一致时可对齐调整）。
5. **绑定侧职责**：配置「全网」= 在建模/绑定工具中把 Skill 实例关联到 **KN 锚点**；配置「某类」= 关联到 **业务关系**；配置「某实例」= 在实例层建边。

---

## 3. 详细设计 (Detailed Design)

### 3.1 概念与术语

| 术语 | 说明 |
|------|------|
| `skills` Object Type | BKN 中固定 ID 的 Skill 元数据对象类，实例即 Skill 条目 |
| `kn_scope` Object Type | 内置的「知识网络作用域锚点」对象类，实例语义为「本 KN 的根」 |
| 网络级绑定 | Skill 实例与某 KN 下唯一 `kn_scope` 实例通过约定 RelationType 相连 |
| 对象类级绑定 | 通过 `业务 OT ↔ skills` 的 RelationType，在不指定具体业务实例时类级召回 |
| 实例级绑定 | 同一 RelationType 下，边连接 **具体业务实例** 与 Skill 实例 |

### 3.2 内置对象类 `kn_scope` 约定

| 项目 | 约定 |
|------|------|
| `object_type_id` | `kn_scope`（最终 ID 以平台内置清单为准） |
| 实例基数 | 每个 `kn_id` 下一个锚点实例（建议 `id` 规则：`kn_scope@{kn_id}` 或平台统一生成） |
| 主要属性 | 至少包含与 KN 的关联键（如 `kn_id`），便于校验与展示；可选 `name` 固定为「本网络作用域」等 |
| 数据来源 | KN 创建时由 bkn-backend / 迁移脚本自动插入，**不对业务用户开放随意增删** |

### 3.3 RelationType 约定

#### 3.3.1 `kn_scope` ↔ `skills`

- **语义**：表示该 Skill 实例对本知识网络 **全局可见**（在「仅 KN 上下文」的召回中命中）。
- **类型**：`direct` 或 `filtered_cross_join` 均可，以团队现有建模习惯为准；**两端类型固定为 `kn_scope` 与 `skills`**。
- **方向**：建议支持 **双向查询**（与现有 `findRelationType` 对 `skills` 与业务 OT 的双向 OR 条件一致），即 `(source=kn_scope, target=skills)` 与 `(source=skills, target=kn_scope)` 择一建模并文档化。

#### 3.3.2 业务 Object Type ↔ `skills`

- **语义**：不变；继续承载 **对象类级 / 实例级** 绑定。
- **实现**：保留现有 **分侧过滤全连接** 等映射方式；不在此文档展开 `mapping_rules` 字段细节。

### 3.4 与 `find_skills` 召回模式的对应关系

| 召回模式（输入） | 绑定侧声明 | 查询路径（目标形态） | `matched_scope` |
|------------------|------------|----------------------|-----------------|
| 仅 `kn_id`（网络级） | Skill 与 **`kn_scope` 锚点** 有边 | `QueryInstanceSubgraph`：锚点 → `skills`（或等价：从锚点出发取关联 Skill 实例） | `network` |
| `kn_id` + `object_type_id` | `业务 OT` ↔ `skills` 存在 RelationType，且子图可达 | 现有子图逻辑 | `object_type` |
| `kn_id` + `object_type_id` + `instance_identities` | 同上，且实例层有边 | 现有实例子图逻辑 | `object_selector` |

**合并**：当请求带 `object_type_id` 时，结果集建议包含：

- 自 **业务路径** 命中；
- 以及可选策略：**始终并入** 网络锚点命中的全网 Skill（若产品希望「通用 Skill 永远可见」），或 **按规则过滤**（若仅在网络级 API 中返回全网 Skill）。该策略需在 PRD 中单点拍板，本设计推荐 **带业务上下文时仍合并全网候选并去重**，以符合「整网通用 + 场景专用」并存。

### 3.5 迁移与兼容

| 阶段 | 行为 |
|------|------|
| 过渡期 | 若某 KN **尚未** 部署 `kn_scope` 或未建立 `kn_scope↔skills` 边，可继续沿用 **「skills 与任意 OT 均无 RelationType → 网络级 `QueryObjectInstances`」** 的旧规则 |
| 目标态 | 全网 Skill **仅** 通过 `kn_scope` 锚点声明；旧规则下线，避免双语义 |

迁移检查清单建议：

- [ ] 为存量 KN 批量创建 `kn_scope` 实例  
- [ ] 将原「零关联」网络级 Skill 改为锚点边  
- [ ] 为存在业务绑定的 KN 补齐锚点边，以恢复「网络级候选」能力  

### 3.6 配置体验（产品侧）

- **声明全网**：选择「绑定到知识网络根 / 全网络」→ 写入 **锚点实例 ↔ Skill** 边。  
- **声明某类**：选择业务对象类与 Skill → 写入 **业务 OT ↔ skills** 关系下的绑定（与现有一致）。  
- **声明某实例**：在实例上绑定 Skill → 写入 **实例级边**。

配置者 **不需要** 选择 `network` / `object_type` / `instance` 枚举；界面根据操作对象（根 / 类 / 实例）自动对应。

---

## 4. 风险与边界 (Risks & Edge Cases)

- **内置类型与业务 OT 冲突**：`kn_scope` 必须使用保留命名空间，并在创建业务 OT 时拒绝同名 ID。  
- **锚点实例重复或缺失**：若同一 KN 出现多个 `kn_scope` 实例，需定义唯一性约束（创建时幂等、清理任务）。缺失则网络级召回为空或回退旧规则（过渡期）。  
- **性能**：网络级从「全表扫 skills 实例」变为「从锚点出发子图」，一般更可预期；需评估锚点度较大时的 TopK 与索引。  
- **跨分支 / 多版本 KN**：若 KN 存在 branch，锚点实例与边的 branch 策略须与现有概念一致，避免全网 Skill 绑定在错误分支。  
- **权限与审计**：锚点绑定变更应记入审计日志，避免「静默扩大」Skill 作用域。

---

## 5. 替代方案 (Alternatives Considered)

### 方案 A：仅在 RelationType 上使用 `tags` 标记 `skill_scope:network`

- **优点**：不新增 Object Type，改动面小。  
- **缺点**：仍依赖额外机制说明「哪些 Skill 实例」属于全网；标签落在类型层，**实例级归属** 不直观，易产生歧义。

### 方案 B：在 Skill 实例上增加必填属性 `scope`

- **优点**：查询路径简单。  
- **缺点**：与「配置 Skill 时无法预知绑定层级」冲突；Skill 复用到多 KN 时难以维护（见前文讨论）。

### 方案 C：维持「零 RelationType 即全网」

- **优点**：实现已存在，无新增类型。  
- **缺点**：与「同一 KN 内存在任意业务绑定」互斥，**无法混合全网与按类绑定**。

### 最终选择

采用 **方案：`kn_scope` 锚点 Object Type + `kn_scope` ↔ `skills` 关系**，在图结构与产品语义上统一「网络根」，并支持 **与业务路径并存、可合并召回**。

---

## 6. 任务拆分 (Milestones)

- [ ] 评审并冻结 `kn_scope` 的 `object_type_id`、锚点实例 ID 规则与创建时机（KN 创建流水线 / 迁移脚本）  
- [ ] 在 BKN 内置类型清单与文档中登记 `kn_scope` 与 `kn_scope↔skills` RelationType 模板  
- [ ] context-loader：`find_skills` 网络级路径改为基于锚点子图；实现与业务路径的去重合并策略  
- [ ] 绑定/建模工具：支持「绑定到网络根」操作，写入锚点边  
- [ ] 存量数据迁移与兼容开关；观测指标与告警（锚点缺失、多实例）  
- [ ] 废弃「零关联即全网」临时规则（目标态）  

---

## 参考

- `find_skills` 现有三层召回与 `matched_scope` 语义：实现见 `adp/context-loader/agent-retrieval/server/logics/knfindskills/`（如 `recall_coordinator.go`）；若后续补充专项设计文档，可在此追加链接。  
- [BKN 业务知识网络建模语言 技术设计文档](../../bkn_docs/DESIGN.md)（文档结构与 BKN 分层表述）  
