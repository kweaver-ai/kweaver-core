# BKN 系统设计

> Status: Draft | Version: 0.7.0 | Date: 2026-03-06

---

## 一、背景

### 现状

ADP（AI Data Platform）通过 **BKN Engine**（ontology-manager / ontology-query）管理业务知识网络。当前的管理方式是 JSON + RESTful API，Agent 通过 **Context Loader**（agent-retrieval）的六个工具（kn_search、query_object_instance、query_instance_subgraph 等）访问知识网络。

这套体系在运行时是完整的，但存在三个结构性缺口：

1. **知识定义不可流通。** 业务知识网络锁在 BKN Engine 内部，无法以文件形式独立存在、版本管理、跨环境复制。Agent 必须连接平台才能获取业务知识。
2. **知识管理是命令式的。** 通过 API 逐条创建/修改对象类、关系类、行动类，没有声明式的"目标状态"定义，变更不可审查、不可回放。
3. **缺少风险建模。** BKN Engine 当前支持三元组（ObjectType / RelationType / ActionType），没有 RiskType。行动的风险等级、审批策略、回滚方案无处声明。

### 目标

引入 **BKN 语言规范**（见 [SPEC.md](./SPEC.md)）和配套的 **bkn-lang** 工具，使业务知识网络可以：

- 以 `.bkn` 文件形式定义、版本管理、Git 追踪
- 通过声明式导入同步到 BKN Engine
- 从 BKN Engine 导出为 `.bkn` 文件（round-trip）
- 作为独立资产被 Agent 直接消费，无需依赖平台

### 非目标

- 不替换现有 RESTful API — BKN 语言是声明式补充，API 继续作为命令式入口
- 不改变 Context Loader 的运行时架构 — Context Loader 仍是 Agent 访问平台的主通道
- 不在本文档中定义 BKN 语言语法 — 语法规范见 [SPEC.md](./SPEC.md)

---

## 二、术语表

| 术语 | 含义 |
|------|------|
| BKN | Business Knowledge Network，业务知识网络 |
| BKN Engine | ontology-manager + ontology-query 的统称，构建和管理业务知识网络的运行时引擎 |
| 业务知识网络 | BKN Engine 内部构建的数据结构（概念模型 + Concept Index），不是独立模块 |
| Context Loader | agent-retrieval 服务，为 Agent 提供知识网络的运行时查询能力（Retrieval / Ranker） |
| BKN 语言 | 基于 Markdown 的业务知识网络建模语言，语法规范见 SPEC.md |
| .bkn 文件 | BKN 语言定义的 Markdown 文件，自带数据映射和工具配置，可独立流通 |
| SKILL.md | agentskills.io 标准的 Skill 入口文件，承担网络级编排（索引、拓扑、使用指南） |
| bkn-lang | BKN 语言的核心库 + CLI 工具，可独立运行也可被 BKN Engine 集成（CLI 可 alias 为 `bkn`） |
| type: network | BKN 文件类型之一，描述网络整体（身份、版本、定义清单），面向 Engine / bkn-lang |
| 四元组 | BKN 语言 v2.0.0 的四种基本类型：对象类(object) / 关系类(relation) / 行动类(action) / 风险类(risk) |
| IR | Intermediate Representation，bkn-lang core 内部的结构化中间表示 |

---

## 三、设计思路与折衷

### 3.1 一份代码，两种使用方式

**决策：** bkn-lang 的核心能力（解析、校验、diff、序列化、checksum）实现为一个独立的库（bkn-lang core）。这个库有两种使用方式：

1. **被 bkn-lang CLI 引用** — 作为独立的命令行工具，供人和 CI 离线使用
2. **被 BKN Engine 集成** — 作为内嵌库，使 Engine 具备 .bkn 解析和序列化能力

```
bkn-lang/core (库)
├── parser          .bkn text → IR
├── validator       IR → Diagnostic[]
├── differ          IR × IR → ChangeSet
├── serializer      IR → .bkn text
└── checksum        body → sha256

        ↗ bkn-lang CLI（独立离线工具）
       ↗ BKN Engine（内嵌集成）
```

**理由：**

- **解释权唯一** — 只有一份解析/序列化实现，不存在 bkn-lang 和 Engine 各自解读 .bkn 格式导致不一致的问题。
- **离线可用** — bkn-lang CLI 不依赖 Engine 即可做格式校验、checksum 计算、解析。
- **Engine 原生感知 .bkn** — Engine 内嵌 bkn-lang core 后，可以直接接收 .bkn 文本、存储 .bkn 文本、在结构化数据变更时重新生成 .bkn 文本，保证一致性。
- **写入路径不收口** — 现有结构化 API 继续可写入。无论哪条路径修改了定义，Engine 都能用内嵌的 serializer 重新生成 .bkn 文本。

**折衷：** Engine 需要引入 bkn-lang core 作为依赖。如果 bkn-lang core 用 Go 实现（与 Engine 同语言），这只是一个普通的 Go module 依赖，代价很小。

### 3.2 .bkn 文件是自包含的可流通资产

**决策：** .bkn 文件不仅是知识的描述文档，还自带数据来源映射（data_view）、工具配置（tool/MCP）、参数绑定等连接信息。

**含义：** Agent 拿到一份 .bkn 文件，无需连接任何平台，就可以按其中的映射信息直连数据源、直调工具。这是 .bkn 文件作为"资产"的核心价值。

一个 .bkn 文件自包含了 Agent 直连所需的全部信息：

```markdown
## Object: inventory               <- 知道查什么
### 数据来源
| 类型 | ID |
|------|-----|
| data_view | inventory_view |     <- 知道数据在哪
> **主键**: `seq_no`                <- 知道怎么定位
> **显示属性**: `material_code`     <- 知道怎么展示

## Action: check_inventory          <- 知道能做什么
### 工具配置
| 类型 | 工具ID |
|------|--------|
| tool | query_object_instance |   <- 知道调什么工具
### 参数绑定
| 参数 | 来源 | 绑定 |
| material_code | input | - |      <- 知道传什么参数
```

**折衷：** 直连模式下 Agent 无法使用 Context Loader 提供的增强能力（Retrieval / Ranker / 逻辑属性计算 / 子图遍历）。这是有意为之 — 轻量可移植和深度能力之间的取舍，由 Agent 的部署环境决定。

### 3.3 三元组升级为四元组

**决策：** 在 BKN 语言 v1.0.0 的三元组（object / relation / action）基础上，v2.0.0 新增 risk（风险类）作为第四种基本类型。

**理由：** 当 Agent 开始执行真实业务操作（修改库存、发起采购），风险治理是 Day 1 需求。风险类独立于行动类，一个风险类可管控多个行动类，实现统一的风险策略。Action 的 `risk_level` 声明"多危险"，Risk 声明"如何管控"。

**对 BKN Engine 的影响：** 需要新增 RiskType 的存储和 API 支持。

### 3.4 业务知识网络是 BKN Engine 的内部状态

**决策：** 业务知识网络（概念模型 + Concept Index）是 BKN Engine 构建和维护的内部数据结构，不是独立模块。

**含义：** 架构中不存在一个单独的"业务知识网络服务"。Context Loader 通过调用 BKN Engine 的 API 获取知识网络信息，bkn-lang 通过同一组 API 读写知识网络。业务知识网络的生命周期完全由 BKN Engine 管理。

### 3.5 `type: network` 与 SKILL.md 各司其职

**决策：** 保留 `type: network` 作为 BKN 语言的文件类型，与 SKILL.md 并存，各自面向不同受众。

| | `type: network`（network.bkn） | SKILL.md |
|---|---|---|
| **面向谁** | Engine / bkn-lang（机器消费） | Agent（Agent 消费） |
| **标准** | BKN 语言规范 | agentskills.io 标准 |
| **内容** | 网络身份（id, version, owner）、定义清单、导入配置 | 网络拓扑、索引表、使用指南、风险判断指引 |
| **用途** | `bkn apply` 时识别网络身份和导入范围 | Agent 运行时的导航和决策参考 |

**理由：** `type: network` 是网络的**自描述**（"我是谁、我包含什么"），SKILL.md 是网络的**使用手册**（"Agent 怎么用我"）。两者不重叠：

- `bkn apply ./supply-chain/` 时，bkn-lang 读 `network.bkn` 知道目标网络 ID、要导入哪些定义
- Agent 读 SKILL.md 知道如何按需加载、如何判断风险、何时需要审批

**对 SPEC.md 的影响：** SPEC.md v2.0.0 废弃了 `type: network`（改由 SKILL.md 承担）。需要恢复 `type: network`，但职责限定为网络身份和定义清单，不承担 Agent 编排。

### 3.6 .bkn 文本作为 Engine 存储字段，一致性由 Engine 保证

**决策：** BKN Engine 在存储每个定义的结构化数据的同时，存储对应的 .bkn 文本作为字段。当结构化数据通过任何路径（bkn-lang apply 或结构化 API）发生变更时，Engine 使用内嵌的 bkn-lang core serializer 重新生成 .bkn 文本，保证两者一致。

**理由：** .bkn 文本是知识定义的人类可读 / Agent 可读表示，具有独立价值。将其作为 Engine 的存储字段意味着：

- 任何时候都可以从 Engine 直接获取最新的 .bkn 表示，无需 bkn-lang export
- Context Loader 可以将 .bkn 文本直接返回给 Agent
- 不存在"结构化数据更新了但 .bkn 文本没跟上"的不一致问题

**折衷：** 每次定义变更都需要额外执行一次序列化。但 serializer 是纯内存操作，单个定义的序列化开销可忽略。

---

## 四、总体逻辑分层架构

```
                              Agent
                            /       \
                   /----------         ----------\
            读 .bkn 文件                    Context Loader
            按映射直连资源                 (Retrieval / Ranker)
                 |                               |
                 |                               |
                 |         +-----------------------------------------+
                 |         |          BKN Engine                     |
                 |         |      (ontology-manager / query)         |
                 |         |                                         |
                 |         |    +--- bkn-lang core (内嵌) ----------+  |
                 |         |    |  parser / validator / serializer |  |
                 |         |    |  differ / checksum              |  |
                 |         |    +---------------------------------+  |
                 |         |                                         |
                 |         |    +------------------------------+     |
                 |         |    |  业务知识网络 (内部状态)        |     |
                 |         |    |  概念模型: 数据 / 逻辑         |     |
                 |         |    |            风险 / 行动         |     |
                 |         |    |  Concept / Index              |     |
                 |         |    |  (.bkn 文本字段)               |     |
                 |         |    +------------------------------+     |
                 |         |          |              |               |
                 |         |     VEGA Engine    Exec Factory         |
                 |         +----+-----+--------------+--------------+
                 |              |     |              |
                 |              |     v              v
                 |              |  +-----------------------------+
                 |              |  | 数据库 / 工具 / MCP / Dataflow |
                 v              |  +-----------------------------+
            数据库 / 工具       |
            (直连)             |  结构化 API        .bkn API
                               |  (现有)            (新增，可选)
                               |       \           /
                               |        \         /
                               |   +-----+-------+-----+
                               |   |     bkn-lang CLI     |
                               |   |                    |
                               |   |  bkn-lang core (库)  |
                               |   +--------------------+
                               |            ^
                               |            |
                               |         人 / CI
                               |
                               |  import bkn-lang core
                               |
                      +--------+--------+
                      |  Agent 框架      |
                      |  (L2 结构化解析)  |
                      +-----------------+
```

### 四个实体

| 实体 | 性质 | 状态 |
|------|------|------|
| **network.bkn + .bkn 文件 + SKILL.md** | 静态资产，可独立流通 | 由人编写或 bkn-lang export 生成 |
| **bkn-lang** | 核心库 + CLI，可独立运行也可被 Engine 集成 | **待建** |
| **BKN Engine** | 运行时引擎，内嵌 bkn-lang core（Go 微服务） | 已有，需集成 bkn-lang core + 补 RiskType |
| **Context Loader** | Agent 运行时查询入口（Go 微服务） | 已有 |

### 两条定义写入路径

无论哪条路径写入，Engine 都保证结构化数据和 .bkn 文本的一致性。

```
路径 1（通过 bkn-lang CLI）:
  人/CI → bkn-lang CLI → bkn-lang core 解析 .bkn → 结构化 API → Engine 存储
                                                              ↓
                                                Engine 用内嵌 serializer 重新生成 .bkn 文本并存储

路径 2（通过 .bkn API，可选）:
  人/CI → .bkn 文本 → Engine 内嵌 bkn-lang core 解析 → Engine 存储（结构化 + .bkn 文本）

路径 3（通过现有结构化 API）:
  UI/脚本 → 结构化数据 → Engine 存储
                           ↓
            Engine 用内嵌 serializer 重新生成 .bkn 文本并存储
```

### Agent 的两条消费路径

| 维度 | 路径 A：文件直连 | 路径 B：平台托管 |
|------|----------------|-----------------|
| Agent 读什么 | .bkn 文件（Markdown） | Context Loader API |
| 数据怎么来 | 按 .bkn 中的 data_view 映射直连数据源 | VEGA Engine 代理查询 |
| 工具怎么调 | 按 .bkn 中的工具配置直调 tool/MCP | Execution Factory 代理执行 |
| 依赖平台 | 不需要 | 需要 |
| 可移植 | .bkn 文件拿走就能用 | 绑定 ADP 平台 |
| 逻辑属性 | 不支持 | 支持（算子/指标计算） |
| 子图遍历 | 不支持 | 支持（多跳关系路径） |
| Retrieval/Ranker | 不支持 | 支持（语义搜索 + 排序） |
| 风险引擎 | 仅读声明（Agent 自行判断） | 引擎执行（审批/模拟/审计） |

两条路径不互斥。同一份 .bkn 文件既可以被路径 A 的 Agent 直接读取，也可以通过 `bkn apply` 导入 BKN Engine 后由路径 B 的 Agent 通过 Context Loader 使用。

### 依赖方向

```
bkn-lang CLI ------> BKN Engine <------- Context Loader <------- Agent
  |                    |                                        |
  |              内嵌 bkn-lang core                               |
  |                                                             |
  +------> .bkn 文件 <--------------------------------------------+
```

- bkn-lang CLI 依赖 BKN Engine（通过 REST API）和 .bkn 文件（通过文件系统）
- BKN Engine 依赖 bkn-lang core（作为内嵌库，编译期依赖）
- Context Loader 依赖 BKN Engine（通过内部 API）
- Agent 依赖 Context Loader（路径 B）或 .bkn 文件（路径 A）
- **没有环形依赖** — bkn-lang CLI 依赖 Engine 的 API，Engine 依赖 bkn-lang core 库（不是 CLI）

---

## 五、核心流程

### 5.1 导入：.bkn 文件 -> BKN Engine（通过 bkn-lang CLI）

```
人编写 .bkn 文件
       |
       v
bkn validate ./              <- bkn-lang core 解析 + 格式校验（离线，不连 Engine）
       |
       v
bkn diff ./ --target url     <- bkn-lang core 解析本地文件
       |                           调 Engine API 读线上状态
       |                           bkn-lang core diff -> 输出变更集
       v
bkn apply ./ --target url    <- validate -> diff -> 用户确认
       |                           调 Engine API 逐条写入
       |                           Engine 执行业务校验 + 存储
       v                           Engine 用内嵌 serializer 生成 .bkn 文本并存储
完成
```

### 5.2 导入：.bkn 文件 -> BKN Engine（通过 .bkn API）

```
人/CI 提交 .bkn 文本到 Engine 的 .bkn API
       |
       v
Engine 使用内嵌 bkn-lang core 解析 .bkn 文本
       |
       v
Engine 执行业务校验（引用完整性、风险关联）
       |
       v
Engine 存储结构化数据 + .bkn 文本
```

### 5.3 导出：BKN Engine -> .bkn 文件

```
bkn export --source url --output ./skills/supply-chain/
       |
       v
调 BKN Engine API 读取定义（Engine 直接返回 .bkn 文本字段）
       |
       v
bkn-lang 将 .bkn 文本写入文件
       |
       v
生成 SKILL.md 脚手架（索引表、拓扑图）
       |
       v
bkn checksum --generate    <- bkn-lang core 计算 SHA-256，写入 CHECKSUM 文件
```

### 5.4 Agent 文件直连（路径 A）

```
Agent 接收用户请求
       |
       v
读 SKILL.md -> 获取网络拓扑和索引表
       |
       v
按需读取相关 .bkn 文件
       |
       +--- 从 Object 定义中获取:
       |      data_view    -> 知道数据在哪
       |      primary_key  -> 知道怎么定位
       |      display_key  -> 知道怎么展示
       |
       +--- 从 Action 定义中获取:
       |      tool / MCP   -> 知道调什么
       |      参数绑定      -> 知道传什么
       |      risk_level   -> 知道风险等级
       |
       +--- 从 Risk 定义中获取:
       |      管控策略      -> 知道是否需要审批
       |      前置检查      -> 知道执行前要做什么
       |
       v
Agent 按映射信息直连数据源 / 直调工具
```

### 5.5 Agent 平台托管（路径 B）

```
Agent 接收用户请求
       |
       v
调用 Context Loader 工具:
  kn_search / kn_schema_search       <- 探索发现
  query_object_instance              <- 条件查询
  query_instance_subgraph            <- 子图遍历
  get_logic_properties_values        <- 逻辑属性计算
  get_action_info                    <- 获取动作配置
       |
       v
Context Loader 调用 BKN Engine
       |
       v
BKN Engine 通过映射访问:
  VEGA Engine        <- 数据查询
  Exec Factory       <- 工具 / 算子 / MCP 执行
  Dataflow           <- 流处理 / 批处理
```

---

## 六、核心模块

### 6.1 bkn-lang core（共享库）

#### 定位

BKN 语言的唯一解析/序列化实现。既被 bkn-lang CLI 引用，也被 BKN Engine 内嵌集成。确保无论从哪条路径读写 .bkn，都是同一份代码在执行。

#### 内部结构

```
bkn-lang/core/
├── parser/            <- .bkn text -> IR
│   ├── frontmatter    <- YAML frontmatter 解析
│   └── body           <- Markdown body 解析（Object/Relation/Action/Risk sections）
├── validator/         <- IR -> Diagnostic[]（格式层校验）
│   ├── structure      <- 必填字段、表格格式、YAML block 合法性
│   └── checksum       <- checksum 格式合法性 + 内容一致性
├── differ/            <- IR × IR -> ChangeSet
├── serializer/        <- IR -> .bkn text
└── checksum/          <- body text -> sha256 hash
```

#### 核心数据结构（IR）

```typescript
interface Definition {
  type: 'object' | 'relation' | 'action' | 'risk'
  id: string
  name: string
  version?: string
  network?: string
  namespace?: string
  checksum?: string
  owner?: string
  tags?: string[]
  body: ObjectBody | RelationBody | ActionBody | RiskBody
}

interface ObjectBody {
  dataSource: { type: string; id: string }
  primaryKey: string
  displayKey: string
  propertyOverrides?: PropertyOverride[]
  logicProperties?: LogicProperty[]
}

interface RelationBody {
  source: string              // 起点 object id
  target: string              // 终点 object id
  relationType: 'direct' | 'data_view'
  mappingRules: MappingRule[]
}

interface ActionBody {
  boundObject: string         // 绑定的 object id
  actionType: 'add' | 'modify' | 'delete'
  riskLevel?: 'low' | 'medium' | 'high'
  enabled?: boolean
  requiresApproval?: boolean
  toolConfig: { type: 'tool' | 'mcp'; id: string }
  paramBindings: ParamBinding[]
  trigger?: TriggerCondition
  schedule?: ScheduleConfig
}

interface RiskBody {
  scope: { objectId: string; actionId: string; riskLevel: string }[]
  policies: { condition: string; strategy: string }[]
  preChecks?: { name: string; type: string; description: string }[]
  rollbackPlan?: string
  auditRequirements?: string
}

interface ChangeSet {
  creates: Definition[]
  updates: { before: Definition; after: Definition }[]
  deletes: { type: string; id: string }[]
}
```

#### 校验分层

| 层次 | 归属 | 内容 | 需要 Engine |
|------|------|------|:-----------:|
| 格式校验 | bkn-lang core validator | 语法、必填字段、表格结构、YAML 合法性 | 否 |
| 业务校验 | BKN Engine | 引用完整性、风险关联、与线上状态一致性 | 是 |

### 6.2 bkn-lang CLI

#### 定位

面向人和 CI 的命令行工具。引用 bkn-lang core 做 .bkn 解析，通过 BKN Engine REST API 完成业务操作。

#### 内部结构

```
bkn-lang/
├── core/              <- 共享库（见 6.1）
│
├── providers/         <- I/O 适配层
│   ├── file/          <- 读写 .bkn / SKILL.md / CHECKSUM 文件
│   └── engine/        <- 调 BKN Engine REST API
│
└── cli/               <- 命令层（用户交互）
    ├── validate       <- file.read -> core.parse -> core.validate（离线）
    ├── diff           <- file.read + engine.read -> core.diff
    ├── apply          <- validate -> diff -> confirm -> engine.write
    │                     Engine 侧执行业务校验，失败则返回错误
    ├── export         <- engine.read（含 .bkn 文本）-> file.write
    ├── checksum       <- file.read -> core.checksum -> file.write / verify
    └── init           <- 生成 SKILL.md 脚手架 + 目录结构
```

#### CLI 命令

| 命令 | 数据流 | 需要 Engine |
|------|--------|:-----------:|
| `bkn validate <path>` | file -> core.parse -> core.validate -> 报告 | 否 |
| `bkn diff <path> --target <url>` | file -> core.parse + Engine API read -> core.diff -> 变更集 | 是 |
| `bkn apply <path> --target <url>` | validate -> diff -> confirm -> Engine API write | 是 |
| `bkn export --source <url> -o <path>` | Engine API read（含 .bkn 文本）-> file write | 是 |
| `bkn checksum --generate <path>` | file -> core.parse -> core.checksum -> 写 CHECKSUM | 否 |
| `bkn checksum --verify <path>` | file -> core.parse -> core.checksum -> 比对 CHECKSUM | 否 |
| `bkn init <name>` | 生成目录 + SKILL.md 脚手架 | 否 |

#### 作为库使用

bkn-lang core 可被 Agent 框架直接引用，实现 L2 级别的结构化解析：

```typescript
import { parse, validate } from 'bkn-lang'

const result = parse(bknFileContent)
const diagnostics = validate([result.definition])
```

### 6.3 BKN Engine 扩展

#### 集成 bkn-lang core

BKN Engine 将 bkn-lang core 作为 Go module 依赖引入，获得以下能力：

| 能力 | 用途 |
|------|------|
| parser | 接收 .bkn 文本，解析为结构化数据后存储（.bkn API） |
| serializer | 结构化数据变更时，重新生成 .bkn 文本并存储 |
| validator | 对 .bkn API 接收的文本做格式校验 |
| checksum | 导出时计算内容指纹 |

#### .bkn 文本存储

每个定义（ObjectType / RelationType / ActionType / RiskType）新增 `bkn_text` 字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `bkn_text` | TEXT | 该定义对应的 .bkn 文本表示 |
| `bkn_text_updated_at` | TIMESTAMP | .bkn 文本最后更新时间 |

一致性保证：

- **bkn-lang apply / .bkn API 写入：** 同时存储结构化数据和 .bkn 原文
- **结构化 API 写入：** Engine 在落地后使用内嵌 serializer 重新生成 .bkn 文本
- **读取：** API 返回结构化数据时可附带 `bkn_text` 字段

#### .bkn API（可选，新增）

BKN Engine 可暴露直接接收 .bkn 文本的端点，使用内嵌的 bkn-lang core 解析：

| API | 方法 | 说明 |
|-----|------|------|
| `/api/kn/{id}/import-bkn` | POST | 接收 .bkn 文本，解析 + 格式校验 + 业务校验 + 落地 |
| `/api/kn/{id}/validate-bkn` | POST | 接收 .bkn 文本，返回格式校验 + 业务校验结果 |

这些端点是可选的增强。即使没有它们，bkn-lang CLI 通过现有结构化 API 也能完成所有操作。

#### RiskType 支持

BKN Engine 当前支持三种类型：

| 类型 | API | 存储 |
|------|-----|------|
| ObjectType | `/api/kn/{id}/object-types` | MariaDB + OpenSearch |
| RelationType | `/api/kn/{id}/relation-types` | MariaDB + OpenSearch |
| ActionType | `/api/kn/{id}/action-types` | MariaDB + OpenSearch |

需要新增：

| 类型 | API | 说明 |
|------|-----|------|
| **RiskType** | `/api/kn/{id}/risk-types` | CRUD + 关联 ObjectType / ActionType |

RiskType 需要支持：

- **管控范围：** 关联哪些 ObjectType + ActionType，声明风险等级
- **管控策略：** 按条件分级的管控规则
- **前置检查：** permission / simulation / approval / precondition
- **回滚方案、审计要求**

#### 业务校验（Engine 侧）

当定义通过任何路径写入时，BKN Engine 执行业务校验：

| 校验项 | 说明 |
|--------|------|
| 引用完整性 | Relation 引用的 source/target ObjectType 必须存在 |
| Action 绑定 | Action 绑定的 ObjectType 必须存在 |
| Risk 关联 | Risk 管控范围中引用的 ObjectType / ActionType 必须存在 |
| 风险完整性 | risk_level: high 的 Action 建议关联 RiskType（可配置为 warning 或 error） |

校验失败时返回结构化错误，由调用方（bkn-lang CLI 或 .bkn API 客户端）展示。

#### 对 Context Loader 的影响

Context Loader 需要新增获取 RiskType 信息的能力，使 Agent 在路径 B 中可以查询到风险管控信息。具体可扩展 `kn_schema_search` 返回 risk_types，或新增 `get_risk_info` 工具。

### 6.4 .bkn 文件 + SKILL.md

#### 定位

可独立流通的知识资产。一份 .bkn 文件自包含了 Agent 直连数据源和工具所需的全部信息。

#### 文件结构

```
.claude/skills/supply-chain/
├── network.bkn                        # type: network — 网络身份 + 定义清单（Engine / bkn-lang 用）
├── SKILL.md                           # Agent 使用指南（索引、拓扑、风险判断指引）
├── CHECKSUM                           # 网络级一致性校验（bkn checksum 生成）
├── objects/
│   ├── material.bkn                   # 对象类：物料
│   └── inventory.bkn                  # 对象类：库存
├── relations/
│   └── material_to_inventory.bkn      # 关系类：物料 -> 库存
├── actions/
│   ├── check_inventory.bkn            # 行动类：查询库存（low risk）
│   └── adjust_inventory.bkn           # 行动类：调整库存（high risk）
└── risks/
    └── inventory_adjustment_risk.bkn  # 风险类：库存调整管控
```

#### .bkn 文件自包含的连接信息

| 信息 | 在 .bkn 中的位置 | Agent 用途 |
|------|-----------------|-----------|
| 数据在哪 | Object -> 数据来源表（data_view ID） | 直连数据源查询 |
| 怎么定位 | Object -> 主键 | 唯一标识实例 |
| 怎么展示 | Object -> 显示属性 | UI / 回复中的显示名 |
| 对象怎么连 | Relation -> 映射规则表 | 跨对象关联查询 |
| 调什么工具 | Action -> 工具配置表（tool/MCP ID） | 调用工具执行操作 |
| 传什么参数 | Action -> 参数绑定表 | 构造工具调用参数 |
| 风险多高 | Action frontmatter -> risk_level | 判断是否需要审慎执行 |
| 怎么管控 | Risk -> 管控策略表 | 决定审批 / 模拟 / 回滚策略 |

#### network.bkn 的角色

`type: network` 文件是网络的自描述，面向 Engine 和 bkn-lang：

- **网络身份：** id、version、owner、BKN spec version
- **定义清单：** 列出网络包含的所有定义及其文件路径
- **导入配置：** `bkn apply` 时的目标网络标识

`bkn apply ./supply-chain/` 时，bkn-lang 首先读 `network.bkn`，确定目标网络 ID，再按清单加载各定义文件。

#### SKILL.md 的角色

SKILL.md 遵循 agentskills.io 标准，面向 Agent：

- **拓扑图：** Mermaid 图展示对象之间的关系网络
- **索引表：** 列出所有 object / relation / action / risk 及其文件路径
- **使用指南：** 告诉 Agent 如何按需加载、如何判断风险等级、何时需要审批

#### network.bkn 与 SKILL.md 的关系

两者并存，各司其职：

- **network.bkn** 回答"这个网络是什么"（身份 + 清单），**SKILL.md** 回答"Agent 怎么用"（导航 + 指南）
- `bkn init` 和 `bkn export` 同时生成两个文件
- 索引表在两者中都有，但 network.bkn 的清单是机器可解析的（供 bkn-lang 和 Engine 用），SKILL.md 的索引表是 Agent 可阅读的（带描述和使用建议）

---

## 七、总结

### 做了什么

本文档定义了 BKN 语言规范落地的架构方案。核心是一份共享库、四个实体、两条路径。

**一份共享库：**

bkn-lang core 是 BKN 语言的唯一解析/序列化实现，同时被 bkn-lang CLI 和 BKN Engine 使用，确保解释权唯一。

**四个实体：**

| 实体 | 角色 |
|------|------|
| network.bkn + .bkn 文件 + SKILL.md | 可流通的知识资产 |
| bkn-lang | 核心库 + CLI，可独立运行也可被 Engine 集成 |
| BKN Engine | 运行时引擎，内嵌 bkn-lang core，保证 .bkn 文本一致性 |
| Context Loader | Agent 运行时查询入口 |

**两条路径：**

- **路径 A（文件直连）：** Agent 读 .bkn 文件，按映射信息直连数据源 / 工具。轻量、可移植、不依赖平台。
- **路径 B（平台托管）：** Agent 通过 Context Loader 访问 BKN Engine。深度能力（Retrieval / Ranker / 逻辑属性 / 子图遍历 / 风险评估）。

### 待建工作

| 工作项 | 依赖 | 说明 |
|--------|------|------|
| bkn-lang core | SPEC.md v2.0.0 | 共享库：parser / validator / differ / serializer / checksum |
| bkn-lang CLI | bkn-lang core | 命令行工具：providers（file + engine）+ CLI 命令 |
| BKN Engine 集成 bkn-lang core | bkn-lang core | 内嵌库，.bkn 文本存储，一致性保证 |
| BKN Engine RiskType | 无 | 新增存储 + API + 业务校验逻辑 |
| BKN Engine .bkn API（可选） | bkn-lang core | 接收 .bkn 文本的端点 |
| Context Loader 风险支持 | BKN Engine RiskType | 扩展工具，暴露 RiskType 信息给 Agent |
| SPEC.md 恢复 `type: network` | 无 | 恢复 network 类型定义，职责限定为网络身份 + 定义清单 |

### 设计原则

1. **一份代码，两种使用** — bkn-lang core 是 .bkn 的唯一实现，CLI 和 Engine 共享同一份代码
2. **.bkn 文件是自包含资产** — 拿走就能用，不绑定平台
3. **Engine 存储 .bkn 文本并保证一致性** — 结构化数据变更时自动重新序列化
4. **业务知识网络是内部状态** — 不是独立模块，由 BKN Engine 构建和维护
5. **两条路径不互斥** — 同一份 .bkn 文件服务于文件直连和平台托管两种场景
6. **network.bkn 与 SKILL.md 各司其职** — network.bkn 面向 Engine（网络身份），SKILL.md 面向 Agent（使用指南）

---

## 参考

- [BKN 语言规范 v2.0.0](./SPEC.md) — 语法、字段、校验规则、增量导入语义
- [BKN as Enhanced Skill](./BKN_AS_ENHANCED_SKILL.md) — 设计动机与立论
- [BKN 语言规范 v1.0.0](./SPECIFICATION.md) — 历史版本
