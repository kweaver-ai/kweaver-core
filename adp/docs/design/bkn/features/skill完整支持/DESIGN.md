# BKN Skill 完整支持 技术设计文档

> **状态**：草案
> **版本**：0.2.0
> **日期**：2026-04-16
> **相关 Ticket**：#433

---

## 1. 背景与目标

Skill 由**执行工厂**统一管理，是 Decision Agent 的核心能力单元。现有系统缺少一种机制来描述「Skill X 作用于哪些 BKN 资源」，导致 Agent 在业务上下文中无法自动发现适用的 Skill。

历史方案试图将 Skill 映射为 BKN 内的对象类、再用 `filtered_cross_join` 关系类建立绑定，存在语义污染、作用域表达力不足、与 Skill 生命周期强耦合等问题（见 §5）。

**目标**：

1. 引入**与 BKN 平级**的新概念 **SkillBinding**，描述 Skill 与 BKN 资源的作用域关系，不将 Skill 引入 BKN 本体模型。
2. 支持五种作用域层级：`global`（全平台）、`network`（指定 BKN）、`object_type` / `relation_type` / `action_type`（指定类型）、`instance_selector`（指定对象类中满足条件的实例子集）。
3. 支持语义化召回：通过自然语言查询发现语义相关的 Skill，作为作用域召回的补充路径。
4. Skill 主数据由执行工厂持有，BKN 侧只存储绑定关系（Skill ID + 作用域），不复制元数据。

**非目标**：不改变执行工厂对 Skill 的管理方式；不修改现有 BKN 本体模型；不在本期实现 Skill 执行路由或授权控制。

---

## 2. 方案概览

### 2.1 核心思路

**SkillBinding 是与 BKN 平级的配置对象，而非 BKN 本体内的节点或边。**

| 维度 | BKN 本体 | SkillBinding |
|------|----------|--------------|
| 描述的是 | 业务概念与业务关系 | Skill 的作用域配置 |
| 存储层 | BKN 本体存储 | 独立配置表 |
| Skill 语义 | 不包含 | 引用执行工厂中的 skill_id |

### 2.2 总体架构

```text
┌─────────────────────────────────────────┐
│           execution-factory              │
│   Skill 主数据 + 语义搜索索引              │
└──────────┬──────────────────────────────┘
           │ ① skill_id 引用  ② 内部语义搜索调用（服务端）
┌──────────▼──────────────────────────────┐
│              bkn-backend                 │
│                                          │
│  SkillBinding（独立配置表）               │
│  skill_id + scope                        │
│  ├── CRUD API（管理端）                  │
│  └── FindSkills API（内部）              │
│       作用域召回 + 语义召回 → 合并 → ids  │
│                                          │
│  BKN 本体（ObjectType / RelationType…）  │
│  ← 不感知 Skill，不引入 skills ObjectType │
└──────────┬──────────────────────────────┘
           │ skill_ids（含权重/score）
┌──────────▼──────────────────────────────┐
│            context-loader                │
│   find_skills(上下文 + skill_query?)     │
│   → bkn-backend FindSkills（拿 ids）     │
│   → execution-factory（按 ids 拿元数据） │
│   → 返回给 Agent                         │
└─────────────────────────────────────────┘
```

### 2.3 关键决策

| 决策 | 原因 |
|------|------|
| SkillBinding 与 BKN 平级 | Skill 不是业务概念，不应进入本体；绑定配置应独立于本体演化 |
| 不引入 skills ObjectType | 消除「技术型对象类污染业务本体」；context-loader 不再依赖图遍历语义查 Skill |
| Skill 主数据由执行工厂持有 | BKN 侧只存 skill_id 引用，单一数据来源，无缓存同步问题 |
| 语义索引归属执行工厂 | 数据主权在执行工厂，context-loader 和 BKN 均不缓存 Skill 内容 |
| 语义召回在 bkn-backend 内部完成 | context-loader 只调用一次 bkn-backend（拿 ids）+ 一次 execution-factory（拿元数据），不感知召回内部机制；避免 context-loader 直接依赖执行工厂语义接口 |

---

## 3. 详细设计

### 3.1 SkillBinding 数据模型

SkillBinding 描述一个 Skill 对 BKN 资源的作用域配置，核心字段：

- `skill_id`：执行工厂中的 Skill 标识
- `scope`：作用域，包含 `type` 和按类型选填的定位字段

**作用域类型与字段约束：**

| scope.type | 必填字段 | 语义 |
|------------|----------|------|
| `global` | 无 | 作用于系统内全部 BKN |
| `network` | kn_id | 作用于指定 BKN 的所有类型与实例 |
| `object_type` | kn_id, object_type_id | 作用于指定对象类的所有实例 |
| `relation_type` | kn_id, relation_type_id | 作用于指定关系类 |
| `action_type` | kn_id, action_type_id | 作用于指定行动类 |
| `instance_selector` | kn_id, object_type_id, selector | 作用于指定对象类中满足 selector 条件的实例子集；selector 不可为空，为空时应使用 `object_type` 类型 |

`selector` 复用 BKN 现有的 `CondCfg` 条件配置结构。

### 3.2 作用域召回（Scope Recall）

bkn-backend 提供内部 Resolve 接口，接收 BKN 上下文（kn_id、object_type_id、instance_identities），返回适用的 skill_id 列表。

**解析逻辑（加法语义，所有层级结果合并，越细粒度权重越高）：**

1. 查询所有 `global` 绑定（基础权重：1）
2. 查询当前 `kn_id` 下的 `network` 绑定（权重：2）
3. 若传入 `object_type_id`，查询对应的 `object_type` 绑定（权重：3）
4. 若同时传入 `instance_identities`，对所有 `instance_selector` 绑定发起**一次联合查询**：将各条 selector 条件与实例标识合并为复合条件，通过一次 ontology-query 调用判断哪些绑定命中（权重：4）
5. 合并所有层级命中结果，按 skill_id 去重（同一 Skill 被多个层级命中时取最高权重），排序后返回给 context-loader

若未传入 `instance_identities`，则跳过 `instance_selector` 类型绑定（与 `object_type` 语义区分）。

> 约束：同一个 `(skill_id, scope)` 组合唯一，不存在重复绑定，无需在合并阶段处理同层级重复。

### 3.3 语义化召回（Semantic Recall）

`skill_query` 非空时，bkn-backend 在内部触发语义召回路径：

- 调用执行工厂的语义搜索接口（服务端调用，不经过 context-loader），对 Skill 的 name/description 做相似度检索
- 搜索范围由调用方指定：可传入 kn_id 将语义搜索收窄至该 BKN 相关的 Skill，不传则在全平台范围内搜索；两种情况均可返回未显式绑定但语义相关的 Skill（「仅语义命中」）
- 若作用域召回结果为空（当前上下文无任何绑定），语义召回仍可执行；若语义召回也无结果，返回空数组，不报错

bkn-backend 统一完成作用域召回与语义召回的协调，向 context-loader 只暴露一个 FindSkills 接口，返回合并后的 skill_id 列表。

### 3.4 合并排序策略

bkn-backend 内部合并两路结果，排序规则：

- **主排序**：作用域权重（instance_selector=4 > object_type / relation_type / action_type=3 > network=2 > global=1）
- **次排序**：语义 score 降序（有 skill_query 时）
- 仅语义命中的 Skill（作用域未绑定但语义相关）权重为 0，排在所有作用域命中之后

任意一路（作用域查询或语义搜索）出错时，降级为仅使用另一路结果，不对外报错。

合并后 bkn-backend 将 skill_ids（含权重/score）返回给 context-loader，context-loader 再调用执行工厂批量接口获取元数据（name、description）。

### 3.5 权限设计

| scope.type | 所需权限 |
|------------|----------|
| `global` | 独立权限操作 `skill_binding:write:global`，仅系统管理员持有 |
| 其余 scope | 对应 BKN 的写权限 |

Resolve 接口为内部接口，仅供 context-loader 调用，通过内部身份认证保护。

---

## 4. 风险与边界

| 风险 | 解决方案 |
|------|----------|
| 创建绑定时 skill_id 在执行工厂不存在 | 创建时同步校验 skill_id 有效性，失败则拒绝创建；resolve 时不再重复校验 |
| 执行工厂后续删除已绑定的 skill_id | 后台任务定期检测悬空绑定并告警，由管理员决定是否清理 |
| BKN 资源被删除后绑定不清理 | 删除 BKN 资源前前置校验是否有绑定；或级联删除对应绑定 |
| 执行工厂语义搜索接口未就绪 | 语义召回为可选路径，bkn-backend 内部语义搜索失败时降级为仅返回作用域召回结果，不对外报错 |
| relation_type / action_type 作用域当前无上下文传入 | resolve 接口预留对应入参字段，当前不填时跳过；待 context-loader 补充对应上下文后自然生效 |

---

## 5. 替代方案

### 方案 A（历史方案）：Skill 映射为 BKN ObjectType + filter_cross_join 关系类

将执行工厂的 Skill 表通过 Vega Resource 绑定为 BKN 中固定的 `skills` ObjectType，再为每个对象类创建 `filtered_cross_join` 关系类建立绑定。`filtered_cross_join` 本身作为通用关系类型继续保留，供其他业务场景使用，不随本方案废弃。

**不采用原因**：
- Skill 不是业务概念，引入本体污染业务知识图谱
- 无法自然表达 global / network 级作用域
- 每新增一个绑定就要创建一个关系类，管理成本高
- context-loader 依赖「skills ObjectType 无关联时视为全局」的脆弱临时规则

### 方案 B：context-loader 内部维护绑定配置

**不采用原因**：context-loader 是无状态召回层，绑定配置变更需要重新部署。

### 方案 C（本方案）：SkillBinding 作为与 BKN 平级的独立配置对象

本体模型纯净、作用域层级完整、管理成本低、与 Skill 生命周期解耦、可扩展。

---

## 6. 任务拆分

- [ ] **M1 - 数据模型与 CRUD**：`t_skill_binding` 表、bkn-backend CRUD API、字段校验
- [ ] **M2 - Resolve 接口**：bkn-backend 内部 Resolve API、各层级查询逻辑、instance_selector 联合查询
- [ ] **M3 - context-loader 接入（作用域召回）**：`FindSkillsService` 切换为调用 Resolve API，移除旧的 skills ObjectType 逻辑
- [ ] **M4 - 语义化技能召回**：与执行工厂对齐语义搜索接口，bkn-backend 在 FindSkills 接口内集成语义召回，内部完成两路合并排序

---

## 参考

- [filtered_cross_join DESIGN.md](../filtered_cross_join/DESIGN.md)
- [BKN 数据绑定支持 Vega Resource DESIGN.md](../BKN数据绑定支持resource/DESIGN.md)
- [ContextLoader Skill 召回 Design Doc](../../../../context-loader/agent-retrieval/docs/design/issue-109-contextloader-skill-recall-design.md)
