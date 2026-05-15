# 行动类（ActionType）重构行动类结构 技术设计文档

> **状态**：草案  
> **版本**：0.5.1  
> **日期**：2026-05-11  
> **相关 Ticket**：#288

---

## 1. 背景与目标 (Context & Goals)

平台核心定位：**轻路由网关 + 重工具/MCP 自治执行**。行动类应从「暗示平台执行 CRUD」演进为「**声明路由契约、边界与意图**」，并与现有 BKN 文件、`bkn-backend` 元数据建模及 `strict_mode` 校验对齐。

### 1.1 现状痛点（旧架构）


| 痛点维度      | 具体表现                                                                              |
| --------- | --------------------------------------------------------------------------------- |
| **职责越界**  | `action_type: add/modify/delete` 易被解读为平台直接执行 CRUD，与 Tool/MCP 自治逻辑冲突，**事务边界**容易模糊。 |
| **作用域僵化** | 仅支持绑定**单对象类**；参数取值、影响声明难以在元数据层表达跨对象导航，复杂场景依赖硬编码或工具侧二次查询。                          |
| **隐式魔法**  | 如「库里无实例则自动 INSERT」若发生在平台层，**审计血缘**不透明，工具侧难以稳定感知上下文状态。                             |
| **治理薄弱**  | 影响面多靠注释或隐式推导，缺少可随模型分发的**结构化副作用声明**，难以做一致的前端风险提示，也难以为审计结构化存储预留标准字段。                |


### 1.2 重构目标


| 目标维度         | 说明                                                                                                                            |
| ------------ | ----------------------------------------------------------------------------------------------------------------------------- |
| **边界清晰化**    | 平台侧流程收敛为：**上下文加载 → 条件拦截 → 参数绑定 → 路由调用 → 审计记录**；平台自身**零数据变更**。                                                                 |
| **作用域图谱化**   | 从「单对象锚点」升级为**对象类 / 业务子图（Subgraph）**，提供安全、可预测的导航边界。                                                                            |
| **契约声明化**    | 影响面以**纯元数据契约**表达，**仅用于静态记录行动预期的副作用范围。平台不解析其内容以生成执行计划，也不进行运行时校验或拦截。其核心价值在于为前端提供操作风险提示（如“此操作将影响部门编制”），并为未来可能的审计日志结构化存储预留标准字段。** |
| **MCP 生态兼容** | 参数与意图标准化，平台不绑定特定 ORM 或执行引擎；**工具侧执行 100% 自治**。                                                                                 |


### 1.3 `action_intent`：平台侧数据加载指令（非仅业务标签）

业务上看似仍是 CRUD 时，**仍需要显式 `action_intent`**：在「轻平台、重工具」架构下，平台不写库，但必须向工具提供**一致的只读上下文**；不同意图对应**完全不同的读库/构图策略**。若无明确意图，平台无法判定应如何准备数据，工具也无法可靠理解「当前上下文是什么」。

下表归纳典型取值与**只读**平台行为（具体策略名与实现以运行时为准）：


| `action_intent` 取值（示例） | 平台侧行为（只读）                                                | 为何要区分                                    |
| ---------------------- | -------------------------------------------------------- | ---------------------------------------- |
| **CREATE**             | **不**按主键去库中拉取完整实体，或仅构造带默认值的**空壳/内存实例**，供参数绑定与下游识别「新建」语义。 | 数据尚不存在；若强查 `id=…` 会得到空，易引发后续参数解析与路由错误。   |
| **UPDATE**             | 按需加载**完整快照**（等价于为工具提供当前态，便于 before/after 或依赖现有字段的业务逻辑）。  | 工具常需基于**已有状态**做比较或分支；缺快照则无法保证语义正确。       |
| **DELETE**             | 按需预载待删对象及**关联依赖**（外键、级联涉及对象等），只读。                        | 删除前常需做约束检查或把依赖信息带给工具做安全拦截；缺依赖信息易导致工具侧盲删。 |


**结论**：若缺少 `action_intent`，平台无法统一回答「**该如何为本次行动准备上下文**」——**CREATE** 上无谓查库会浪费资源甚至报错；**DELETE** 上不预载依赖则工具可能缺少做安全判断所需的信息。因此 `**action_intent` 宜定义为「数据加载指令 / 加载策略枚举」**，而非笼统的业务口号。

---

**非目标**：不展开 context-loader / 执行工厂的运行时逻辑；不规定 UI 形态；不替代各服务内既有权限模型细节。

---

## 2. 核心设计哲学


| 原则          | 内涵                                  | 平台侧落地行为                                                                                                      |
| ----------- | ----------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| **轻路由，重自治** | 平台是「交通警察」，工具是「驾驶员」。                 | 平台**不写库**、**不控事务**、**不执行业务校验**。                                                                              |
| **声明即契约**   | 配置静态记录「预期副作用范围」，执行完全交给工具。           | **`impact_contract[]`**（数组）及过渡期单行 `affect` 均为**纯元数据契约**：**不**解析其语义以生成执行计划，**不**据其实施运行时校验或拦截；用于前端风险提示及预留审计字段；**不生成 SQL**。 |
| **上下文边界化**  | 收敛平台可查范围，避免 N+1 与越界访问。              | `navigation_boundary` **显式声明**允许的路径，对越界访问**拦截**。                                                             |
| **意图驱动初始化** | 由业务意图决定**如何加载只读上下文**，而非由平台编写业务执行逻辑。 | `action_intent` 表达**数据加载策略**（见 §1.3）；与 `action_type`（历史三值）语义隔离，避免「平台会代为 INSERT」的心理模型；**不替代**工具内业务分支。         |


---

## 3. 新旧架构对比与核心优势


| 维度            | 旧设计（平台执行型）                         | 新设计（轻路由网关型）                                                                | 核心优势                       |
| ------------- | ---------------------------------- | -------------------------------------------------------------------------- | -------------------------- |
| **职责边界**      | 平台负责条件校验、参数绑定与数据 CRUD。             | 平台只做拦截、绑定、路由与日志；**CRUD 100% 归工具**。                                         | 职责清晰，符合 MCP / 微服务架构习惯。     |
| **作用域**       | 单对象绑定；跨对象常需硬编码或工具再查。               | **子图上下文** + **显式导航路径**；平台按需物化只读视图。                                         | 查询范围可控、性能可预期，抑制隐式笛卡尔积。     |
| **参数 / 取值路径** | 多局限在本对象字段；扩展常要改平台内核。               | 声明层统一 `**dataPath` 语法**，在导航边界内安全跨对象取值与过滤。                                  | 配置直观、解析器可复用                |
| **缺失实例处理**    | 典型问题：`create` 时平台自动 INSERT，事务边界断裂。 | `**missing_root_policy`** 与 `**action_intent`** 联防；平台**打标透传**，不代写库。        | 创建与事务语义由工具自治，血缘可解释。        |
| **影响面管理**     | 隐式或单行绑定（单影响对象类）                     | **`impact_contract[]`：影响对象类为数组**，可逐条声明多类对象的预期副作用；前端按条展示风险提示并预留审计字段。                         | 一次行动可声明**多个**影响对象类；不增加平台对执行的推断职责。       |


---

## 4. 方案概览 (High-Level Design)

### 4.1 核心思路

- **职责定格**：`bkn-backend` 仅维护行动类**元数据**（路由目标、前置条件结构、参数声明、**`impact_contract[]`形式的副作用范围声明**等）；**零业务数据写入**。其中 **`impact_contract` 不触发平台侧执行或拦截逻辑**（定位见 §4.3）。
- **语义分层**：保留存量字段（如 `action_type`、`object_type_id`、**单行** `affect`）以兼容历史 BKN；通过新增 `action_intent`、`context_scope`、**`impact_contract`** 等承载网关语义，逐步弱化「CRUD 枚举」的心理模型。
- **规范同源**：所有进入 `.bkn` / `BKNRawContent` 的字段以 `github.com/kweaver-ai/bkn-specification` 为最终契约；Go DTO 与 `ToBKNActionType` / `ToADPActionType` 与之双向对齐。

### 4.2 架构与职责边界

```text
                    Agent / 编排层
                            |
                   Context Loader / 查询
                            |
              +-----------------------------+
              |        BKN Engine          |
              |      (bkn-backend)         |
              |  ActionType CRUD / 导入    |
              |  校验、序列化、索引         |
              +--------------+------------+
                             |
              .bkn / API 写入行动类元数据
                             |
              +-----------------------------+
              |   外部执行体 (Tool / MCP)   |
              |   事务与数据变更自治         |
              +-----------------------------+
```

- **建模与存储**：Engine 写入 JSON/DB + `.bkn` 文本字段；严格模式校验工具与对象类引用。
- **运行时**：加载上下文、求值条件、绑定参数、调用 MCP、写审计——由执行链路与工具侧完成，本设计不展开。

### 4.3 关键决策

1. **`impact_contract[]`（契约声明化）定位**：**影响对象类以数组表达**；数组中每一项是一条「预期副作用」声明（目标对象类、预期操作、作用域、受影响字段等）。整体**仅用于静态记录「行动预期的副作用范围」**。平台**不**解析其内容以生成执行计划，**不**据此进行运行时校验或拦截。核心价值：**（1）** 为前端提供操作风险提示；**（2）** 为未来审计日志结构化存储**预留标准字段**。存量**单行** `affect` 在过渡期可并行，迁移时可折合为 **长度为 1** 的 `impact_contract` 数组，语义同上。
  - **与 `strict_mode` 的关系**：导入/建模阶段可对字段形态、引用对象类 ID 是否存在等做**与记录完整性相关的**校验（类比元数据外键）；**不**展开为对「声明副作用是否与实际一致」的运行时核验。
2. **边界声明与遍历分离**：`navigation_boundary` 等仅在元数据层约束「允许的路径」；具体子图物化、防 N+1 由查询/加载层实现。
3. `**strict_mode` 分层**：语法与枚举由 `validate_action_type` 承担；引用存在性由 `ActionTypeService` + `ObjectTypeService` / AgentOperator 承担；图谱路径合法性可分阶段引入。

---

## 5. 详细设计 (Detailed Design)

### 5.1 bkn-backend 现行模型

核心结构见 `interfaces.ActionType` / `ActionTypeWithKeyField`：


| 概念分组   | JSON / 字段                             | 说明                                                               |
| ------ | ------------------------------------- | ---------------------------------------------------------------- |
| 标识     | `id`、`name`、`kn_id`、`branch`          | KN 分支内主键与展示名                                                     |
| 历史语义分类 | `action_type`：`add`、`modify`、`delete` | 当前必填枚举；网关语义下建议逐步视为兼容字段或契约提示，**非**平台执行指令                          |
| 单对象锚点  | `object_type_id`                      | 绑定单个对象类；未绑定时 `strict_mode` 禁止 `cond` 且参数不可 `value_from=property` |
| 前置条件   | `cond` → `ActionCondCfg`              | 树形前置条件（**本轮不改变**其与校验实现；仅随存量模型继续使用）                               |
| 影响（弱）  | `affect` → `ActionAffect`             | 单对象类 ID + 注释；无操作类型、作用域、字段列表                                      |
| 调用目标   | `action_source`                       | TOOL：`box_id` + `tool_id`；MCP：`mcp_id` + `tool_name`             |
| 参数     | `parameters` → `Parameter`            | `value_from`：`input` / `property` / `const` / `param` 等          |
| 调度     | `schedule`                            | 与网关正交，保留                                                         |
| 落盘     | `BKNRawContent`                       | `ToBKNActionType` + `bknsdk.SerializeActionType`                 |


**外部概念命名对照**：`code` ↔ `id`；`description` ↔ `comment`（BKN `description`）；`toolInputs` ↔ `parameters`；`invocationTarget` ↔ `action_source`（暂无 timeout/retry）。

### 5.2 BKN SDK 映射（`logics/bkn_convert.go`）


| ADP (`interfaces`) | BKN SDK (`BknActionType`) |
| ------------------ | ------------------------- |
| `ObjectTypeID`     | `BoundObject`             |
| `Comment`          | `Description` / `Summary` |
| `Condition`        | `TriggerCondition`        |
| `Affect`           | `AffectObject`            |
| `ActionSource`     | `ActionSource`            |
| `Parameters`       | `Parameters`              |
| `Schedule`         | `Schedule`                |


新增字段必须同步：**interfaces ↔ 转换函数 ↔ bkn-specification**。

### 5.3 校验与 `strict_mode`

- **模块类型**：`module_type` 须为 `action_type`（导入场景）。
- `**action_type`**：枚举三选一。
- **无对象类 + strict**：禁止 `cond`；禁止 `property` 参数。
- `**action_source` + strict**：`tool` 与 `mcp` 字段互斥；可调用 `AgentOperatorAccess` 校验工具/MCP 是否存在。
- `**cond` + strict**：沿用既有规则（操作符、`SubConds` 上限、叶子字段与取值类型等）。
- **Service + strict**：`object_type_id` 与 `affect.object_type_id` 在 KN 内存在。

### 5.4 关键字段：`action_intent`、`navigation_boundary`、`missing_root_policy`、`impact_contract[]`

以下字段在本方案中分工明确：`action_intent` 管**加载策略**；`context_scope` 内子字段管**范围与缺根**；**`impact_contract` 为数组（`[]`）**，逐元素声明影响的**对象类及预期副作用**，管**静态副作用声明**（见 §4.3）。

#### 5.4.0 `action_intent`：数据加载指令与为何不落格在 `create`/`update`/`delete` 字面值

**1）语义隔离（相对存量 `action_type`）**


| 维度   | 旧名 `action_type`（add/modify/delete） | `action_intent`                              |
| ---- | ----------------------------------- | -------------------------------------------- |
| 心理模型 | 易联想到「平台会执行对应 DML」。                  | 明示「平台选择的是 **LOAD_STRATEGY（加载策略）**」，与「谁写库」解耦。 |
| 目标   | 打破「平台自动落库」的预期，对齐轻网关定位。              | 名称指向**只读准备上下文**，而非代为写入。                      |


**2）未来兼容（即使当前只用少数取值）**

仅用 `create`/`update`/`delete` 三种字面量会难以描述**既不是单纯新建、也不是单纯更新**的行动（例如「以外部为准的同步」：可能既补建又覆写）。可为 `action_intent` 增加 `**SYNC`** 等枚举值，对应平台侧 **minimal load / no snapshot** 等加载策略，而无需把业务硬塞进 C/U/D 框里。

**3）与 `STATE_TRANSITION` 等扩展枚举**

例如在 `CREATE` / `UPDATE` / `DELETE` 之外，再增加 **`STATE_TRANSITION`** 这类名字，表示「这是一次状态机上的跳转」。  
平台**仍然只根据这个字符串做两件事**（与 §1.3 一致）：**需要加载哪些只读数据**、**根实例不存在时按规则是否允许虚拟根**——**不会**在平台里实现完整状态机逻辑。  
具体「`STATE_TRANSITION` 对应查库多深、是否走虚拟根」由**路由与加载组件**在各自版本里写清楚即可，与业务状态图怎么画、两边可独立演进。

#### 5.4.1 `navigation_boundary`（嵌于 `context_scope`）


| 维度         | 说明                                                                                                               |
| ---------- | ---------------------------------------------------------------------------------------------------------------- |
| **设计原因**   | 旧模型下缺少对跨对象读取路径的显式白名单，开放导航易导致 N+1、笛卡尔膨胀与非授权关联访问；网关层需要可判定、可拦截的「允许路径集合」。                                            |
| **平台行为边界** | 元数据声明**允许的**关系跳转方向与范围；运行时由加载/查询子系统在物化只读上下文时套用。若请求的取值路径越过声明边界，平台侧应**直接拦截**（如作用域违约类错误），而非静默放大查询。**不**等价于自动生成业务写操作。 |
| **优势**     | 查询范围可预测、防越界与安全审计更易做；与子图物料化策略对齐，可把「无限图遍历」收口为声明式 whitelist。                                                        |


#### 5.4.2 `missing_root_policy`（嵌于 `context_scope`，与 `action_intent` 强绑定）

**扩展动因：消解「缺实例」场景下的逻辑分叉**  
仅配置 `missing_root_policy` 而不与 `**action_intent` 联动**，会出现「找不到根记录时，究竟是异常还是正常」无法判定的问题。`**action_intent` 是判断「查无此根」应视为错误还是可走虚拟根路径的首要依据**（与 §1.3 各意图下的加载语义一致）。


| 场景                     | 输入 / 状态                              | 期望行为                                                       | 若不做意图区分的典型错误                                  |
| ---------------------- | ------------------------------------ | ---------------------------------------------------------- | --------------------------------------------- |
| **A. 按「更新」语义，但库中无该员工** | `action_intent = UPDATE`（或等价），查库无根实例 | 应**立即失败**或返回明确错误：资源不存在（禁止静默当新建）。                           | 不区分意图时，可能生成**虚拟空对象**再交给工具，工具误执行「对空壳做更新」，语义错乱。 |
| **B. 按「新建」语义，库中尚无该主键** | `action_intent = CREATE`（或等价），查库无根实例 | **正常**：允许虚拟根（如 `{ id: null, isNew: true }` 等约定），继续绑定参数并路由。 | 不区分意图时，一律报「资源不存在」，**新建流程被误杀**。                |


**结论**：**「数据未找到」是异常还是可继续，首先由 `action_intent`（与 `missing_root_policy` 组合）定义**；平台只做加载与打标透传，不替代工具写库。


| 维度                        | 说明                                                                                                         |
| ------------------------- | ---------------------------------------------------------------------------------------------------------- |
| **设计原因**                  | 单独「是否允许无根实例」开关易被误配；必须与 `**action_intent` 强绑定**，避免 UPDATE 误走虚拟根、CREATE 误当异常。                                |
| **与 `action_intent` 的协同** | `missing_root_policy` 表达**策略空间**（如是否允许虚拟根）；`action_intent` 表达**当前行动在何种语义下解释缺根**（见上表）。二者组合决定加载路径，平台仍不落业务判断。 |
| **平台行为边界**                | 平台**不**在缺根时代为 INSERT；至多打标并把上下文与控制标志**透传**给 Tool/MCP。                                                       |
| **优势**                    | 消除「隐式魔法」；CREATE/UPDATE 在缺根条件下的分支清晰，工具契约稳定。                                                                 |


#### 5.4.4 `impact_contract[]`（影响对象类：**数组**）


**结构约定**：**影响对象类以数组（YAML 列表 / JSON 数组）承载**；每个元素是一条独立契约。一次行动可声明**多个**受影响对象类，区别于旧模型**仅一条** `affect`。

| 维度 | 说明 |
| :--- | :--- |
| **设计原因**       | 旧模型里影响面单靠单行 `affect` + 注释，无法机器可读地表达「预期影响**多个**对象类 / 字段」；需要**结构化、数组化**的静态声明，与真实写操作**解耦**。                            |
| **平台行为边界（核心）** | **仅静态记录**：数组中每条元素**独立**描述一条「元数据层面的预期副作用」。平台**不**解析其业务语义以**生成执行计划**；**不**在运行时据契约做**业务校验或拦截**（放行与否仍由 `cond`、工具与业务侧决定）。 |
| **核心价值**       | **（1）前端**：按数组**逐条**渲染风险提示（多条影响对象类）。**（2）演进**：为今后审计日志按对象类/字段维度**结构化落库**预留标准字段 / schema，而非当下由平台强写审计正文。 |
| **与执行的关系**     | 执行 100% 由 Tool/MCP 自治；契约数组**不作为** ORM/SQL 生成、编排脚本或平台侧 diff 依据，避免与真实写入不一致。                                          |


### 5.5 目标数据模型（演进草案）

以下为逻辑视图，**YAML 键名以 bkn-specification 终稿为准**。

```yaml
id: employee_onboard
name: 新员工入职
type: action_type
tags: []
comment: ...

# 兼容期并存
action_type: modify
object_type_id: <ot_uuid>

# 网关增强（建议新增）
action_intent: STATE_TRANSITION

context_scope:
  type: SUBGRAPH
  root_object_type_id: <ot_uuid>
  missing_root_policy: ALLOW_VIRTUAL
  navigation_boundary:
    mode: EXPLICIT_PATHS
    paths:
      - "employee -> department"
    max_traversal_depth: 3

# cond / trigger_condition：本轮不改变，继续使用现有结构与校验

parameters:
  # 沿用现有 Parameter（不在本文展开）

# 影响对象类：数组；每项一条契约（可多条）
impact_contract:
  - target_object_type_id: <ot_uuid_employee>
    expected_operation: UPDATE
    scope: ROOT_CONTEXT
    description: ""
    affected_fields: ["status", "hire_date"]
  - target_object_type_id: <ot_uuid_department>
    expected_operation: UPDATE
    scope: RELATED
    description: "部门编制 +1"
    affected_fields: ["headcount"]

action_source:
  type: mcp
  mcp_id: hr_sync_service_v2
  tool_name: ...
```

### 5.7 `strict_mode` 扩展建议（增量）

1. `**action_type**`：strict 下仍可要求枚举；文档与错误信息引导使用 `action_intent`；远期允许在无 `action_type` 时以 `action_intent` 代替（需版本门控）。
2. `**context_scope**`：strict 要求根对象类存在；`**navigation_boundary` 的路径串**可先格式校验，图谱级合法性可分阶段接入。
3. `**missing_root_policy` + `action_intent`**：strict 下校验允许的意图—策略组合，禁止明显矛盾的配置入库。
4. **`impact_contract[]`**（与 §4.3 一致）：仅做**元数据记录完整性**相关约束（如可选地校验**各数组元素**中引用的 `target_object_type_id` 在 KN 内存在）；**不**把契约语义当作运行时拦截或副作用核验依据。

### 5.8 代码改动范围（ActionType 相关）


| 模块                                          | 工作项                                                                                                       |
| ------------------------------------------- | --------------------------------------------------------------------------------------------------------- |
| `interfaces/action_type.go`                 | 新 struct：`ActionIntent`、`ContextScope`（含 `NavigationBoundary`、`MissingRootPolicy`）、**`impact_contract []ImpactContractItem`（切片/数组）** 等 |
| `driveradapters/validate_action_type.go`    | `context_scope` 等新块校验；`impact_contract` 形态与引用完整性（与 §4.3 一致）；strict 联防规则；新旧字段默认推导                          |
| `logics/bkn_convert.go`                     | 双向转换与新 SDK 字段                                                                                             |
| `logics/action_type/action_type_service.go` | `impact_contract` 引用对象类存在性（若采用）；根对象类等存在性校验                                                                |
| `drivenadapters`                            | 若存储列与 JSON 分裂，按需迁移；以实际 `*_access` 为准                                                                      |


### 5.9 迁移与兼容性

- **读**：无 `context_scope` 时视为 `type=SINGLE`，`root_object_type_id` 取自 `object_type_id`。
- **写**：优先写 **`impact_contract` 数组**；可由旧单行 `affect` **折合为仅含一条元素的数组**（派生 `expected_operation` 须标注推断来源，避免伪精确）。
- **行动条件**：无迁移要求；继续使用现有 `cond` / `trigger_condition`。

---

## 6. 风险与边界 (Risks & Edge Cases)

- **语义双轨**：`action_type` 与 `action_intent` 并存期，客户端易混淆；需文档、校验提示与渐近废弃策略。
- **规范分裂**：若仅改 DB 不改 `bkn-specification`，将导致 `.bkn` 导出与 SDK 解析不一致。
- **路径校验成本**：`navigation_boundary` 全量按图谱语义校验可能增加导入耗时；宜分阶段（格式 → 引用 → 拓扑）。
- `**timeout` / `retry` 与全局策略冲突**：若在 ActionType 与执行工厂重复定义，需约定优先级。
- **越界笛卡尔积**：子图上下文若由下游实现，需防无界 join；本设计仅约束元数据声明侧。

---

## 7. 本期交付范围与约束（#288 实施边界）

本章与 [TASK.md](./TASK.md) 验收项对应，用于区分**本期必交付**与**架构预留、后续迭代**；避免评审时把 §4–§5 的远期模型误当作当前迭代全部落地。

### 7.1 `action_intent` 与存量 `action_type`

- **枚举对齐**：本期冻结为与 `action_type` **完全一致**的三值：`add` / `modify` / `delete`（见 TASK P0-2）。SDK、校验与 OpenAPI `enum` 须同一来源，避免文档与代码漂移。
- **兼容与双写**：请求可同时携带 `action_type` 与 `action_intent` 时，**二者取值必须相同**，否则 400（见 TASK V2-2）。
- **缺省回填**：仅带其一时的回填策略三选一并在实现 PR / ADR 中写死（见 TASK V2-3），与存量 BKN/API 平滑兼容。

### 7.2 `impact_contracts` 与单行 `affect`

- **载荷形态**：本期以 **`impact_contracts`**（JSON 数组，元素形如 `ImpactContractItem`：`object_type_id`、`expected_operation`、`description`、`affected_fields`，及与设计一致的 `scope` 若本期纳入）为**可多目标**结构化声明的首选承载。
- **与单行 `affect`**：扩展 `affect` 上的 `expected_operation`、`affected_fields` 等与契约条目对齐；**互斥/并存/折合规则**以实现与 TASK V2-6、C3-3 为准（单行 `affect` ⇄ **长度为 1** 的 `impact_contracts` 数组等）。
- **strict**：在 strict 模式下，**每条**契约中的 `object_type_id` 须在当前 KN/分支存在（见 TASK V2-5）；**不**将契约当作运行时执行拦截依据（与 §4.3 一致）。

### 7.3 本期**不落地**字段（仅设计预留）

以下 §5 中已叙述的增强能力**本期不进入运行时实现**：导入/反序列化策略须在 P0-3 中明确（忽略未知字段 vs 拒绝），避免「半实现」歧义。

- `missing_root_policy`
- `context_scope`（含子图类型、根绑定等）
- `navigation_boundary`

### 7.4 `cond` 与 `parameters`

- **行为不变**：校验路径、存储与既有规则保持一致；不因本需求的 `action_intent` / `impact_contracts` 改动而改变（见 TASK V2-7）。

### 7.5 OpenAPI 与对外契约

- 对 **`action_type`**、`affect`（或等价内联对象）标注 **`deprecated: true`**，描述指向替代字段 **`action_intent`**、**`impact_contracts`**（见 TASK A5、失败条件 OpenAPI 条）。
- 若 ontology-query 等独立 OpenAPI 含 ActionType 投影，同步 deprecated 与新字段说明（TASK A5-4）。

---

## 8. 任务拆分 (Milestones)

- 与 `bkn-specification` 对齐 `BknActionType` 字段与 YAML 示例
- 扩展 `interfaces` 与 `validate_action_type`（含 strict 规则）
- 实现 `bkn_convert` 双向映射与单元测试
- `ActionTypeService` 中 `impact_contract` 引用完整性（可选）、根对象类等存在性校验
- 导入/导出与旧 BKN 回归；OpenSearch 索引字段评估
- 新旧 BKN 均可导入；无对象类绑定的纯路由场景用例
- `ToBKNActionType` 往返一致性（允许文档化默认填充）

---

## 参考

- [bkn_docs 技术设计](../bkn_docs/DESIGN.md)（文档结构与 BKN 语言总体框架）

