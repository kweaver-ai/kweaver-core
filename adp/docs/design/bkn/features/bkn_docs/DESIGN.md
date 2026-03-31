# BKN 业务知识网络建模语言 技术设计文档

> **状态**：草案
> **版本**：0.9.0
> **日期**：2026-03-06
> **相关 Ticket**：#256

---

## 1. 背景与目标 (Context & Goals)

当前 ADP 通过 **BKN Engine** 管理业务知识网络，主要依赖 JSON + RESTful API 完成定义维护与落地。现有方式能够支撑运行，但在设计协作与 Agent 消费上仍存在明显不足：

- **知识定义不可流通**：知识网络被锁在平台内部，难以独立存储、版本管理、跨环境复制。
- **管理方式偏命令式**：通过 API 逐条变更对象类、关系类、行动类，缺少声明式的目标状态描述。
- **业务与 Agent 可读性不足**：JSON 对业务人员和 LLM 都不够友好，评审和维护成本较高。
- **风险建模缺失**：现有模型难以显式表达风险等级、审批策略、回滚方案与审计要求。

**目标**：引入基于 Markdown 的 BKN 建模语言，以及由 `bkn-sdk` 与 `bkn-tools` 组成的工具体系，使知识网络支持文件化建模、Git 管理、增量导入、导出回放和 Agent 直接消费。

**非目标**：不替换现有 RESTful API，不展开 Parser/Validator 的实现细节，不覆盖 UI 设计。

---

## 2. 方案概览 (High-Level Design)

### 2.1 核心思路

采用 **YAML frontmatter + Markdown body** 的 BKN 文件格式：

- Frontmatter 承载身份、版本、依赖、能力标签、治理字段与 checksums。
- Body 承载 `Network` / `ObjectType` / `RelationType` / `ActionType` / `RiskType` 的正文定义。
- `bkn-sdk` 负责解析、校验、diff、序列化、checksum 等基础能力。
- `bkn-tools` 负责基于 `bkn-sdk` 提供校验、导入、导出、diff 等工具链能力。
- BKN Engine 继续作为运行时最终承载，Importer 通过现有 API 完成落地。

### 2.2 总体架构

```text
                          Agent
                            |
                      Context Loader
                     (Retrieval / Ranker)
                            |
        +-------------------------------------------+
        |              BKN Engine                    |
        |         (ontology-manager / query)         |
        |                                            |
        |    +--- bkn-sdk (内嵌) ----------------+   |
        |    |  parser / validator / serializer  |   |
        |    |  differ / checksum                |   |
        |    +----------------------------------+   |
        |                                            |
        |    业务知识网络 (内部状态)                     |
        |    (.bkn 文本字段)                           |
        +------+---------------------+--------------+
               |                     |
          结构化 API (现有)      .bkn 文件系统
               |                     |
        +------+---------------------+------+
        |          bkn-tools CLI            |
        |          bkn-sdk (库)             |
        +----------------------------------+
                       |
                    人 / CI
```

- **BKN Engine** 是运行时核心，内嵌 `bkn-sdk` 完成解析、校验、序列化与 diff。
- **bkn-tools CLI** 面向开发者和 CI，通过 `bkn-sdk`（库）操作 `.bkn` 文件，通过 REST API 与 Engine 交互。
- **Context Loader** 从 Engine 检索知识网络内容，供 Agent 消费。
- 写入路径有两条：文件导入路径（`.bkn` → bkn-tools → API → Engine）和结构化 API 路径（API → Engine → `.bkn` 回写）。

### 2.3 关键决策

1. **基础能力与工具链分层**：`bkn-sdk` 作为共享基础库供 Engine 和工具复用，`bkn-tools` 作为面向开发者和 CI 的工具链提供离线能力，保证语义一致。
2. **BKN 是自包含资产**：文件中同时包含定义、元数据、工具配置和治理信息，可被 Agent 直接消费。
3. **BKN 不替代 API**：BKN 是声明式建模入口，RESTful API 仍是命令式执行与系统集成入口。
4. **模型升级为四元组**：语言层支持 `ObjectType`、`RelationType`、`ActionType`、`RiskType` 四类核心类型，并由 `Network` 负责网络级组织与承载。
5. **校验分层**：格式校验（语法、必填字段、表格结构）由 `bkn-sdk` 离线完成，不需要连接 Engine；业务校验（引用完整性、风险关联、与线上状态一致性）由 BKN Engine 在线完成。

---

## 3. 详细设计 (Detailed Design)

### 3.1 文件模型与类型

BKN 文件统一采用 UTF-8 编码，结构如下：

1. **YAML Frontmatter**：元数据与治理信息
2. **Markdown Body**：定义正文

文件类型约束如下：

- `type: network`：网络级文件
- `type: object_type`：单对象类型定义
- `type: relation_type`：单关系类型定义
- `type: action_type`：单行动类型定义
- `type: risk_type`：单风险类型定义

组织模式固定为 **一定义一文件**，并要求整个知识网络目录遵循统一结构：

```text
supply-chain/
├── network.bkn                        # type: network — 网络身份 + 定义清单（Engine / bkn-sdk 用）
├── SKILL.md                           # Agent 使用指南（索引、拓扑、风险判断指引）
├── CHECKSUM                           # 网络级一致性校验（bkn-tools checksum 生成）
├── object_types/
│   ├── material.bkn                   # 对象类型：物料
│   └── inventory.bkn                  # 对象类型：库存
├── relation_types/
│   └── material_to_inventory.bkn      # 关系类型：物料 -> 库存
├── action_types/
│   ├── check_inventory.bkn            # 行动类型：查询库存（low risk）
│   └── adjust_inventory.bkn           # 行动类型：调整库存（high risk）
└── risk_types/
    └── inventory_adjustment_risk.bkn  # 风险类型：库存调整管控
```

目录结构与根文件职责约束如下：

- 一个知识网络目录只能有一个 `network.bkn`
- 每个定义文件只承载一个定义
- `object_types/`、`relation_types/`、`action_types/`、`risk_types/` 分别只存放对应类型文件

| 文件 | 职责 |
|------|------|
| `network.bkn` | 网络级定义入口，承载网络身份、版本、owner、capabilities、定义清单与 checksums，供 BKN Engine 与 `bkn-sdk` 使用 |
| `SKILL.md` | 面向 Agent 的首读入口，承载索引、拓扑理解、风险判断与使用建议；Agent 应先读 SKILL.md 获取全局概览，再按需读取具体 `.bkn` 定义文件。SKILL.md 不承载正式定义正文，也不参与定义解析和导入 |
| `CHECKSUM` | 网络级一致性校验文件，记录整个网络目录的校验结果，由 `bkn-tools` 生成和校验，不作为人工编辑的主文件 |

`network.bkn` 与 `SKILL.md` 的职责对比：

| 维度 | network.bkn | SKILL.md |
|------|-------------|----------|
| 面向谁 | Engine / bkn-sdk（机器消费） | Agent（Agent 消费） |
| 内容 | 网络身份（id, version, owner）、定义清单 | 网络拓扑、索引表、使用指南、风险判断指引 |
| 用途 | 导入时识别网络身份和导入范围 | Agent 运行时的导航和决策参考 |

补充原则：

- `network.bkn` 是网络级元数据入口，不替代各类型定义文件
- 目录结构用于约束定义归属与导入边界

### 3.2 五类核心定义

#### Network

描述知识网络本身的身份、范围、版本、能力边界及定义组织方式，典型内容包括：名称与说明、version / branch / status。

#### ObjectType

描述业务对象类型及其数据来源、属性与推导逻辑，典型内容包括：名称与说明、属性定义、数据源。

#### RelationType

描述对象类型之间的映射关系与业务语义，典型内容包括：名称与说明、映射规则。

#### ActionType

描述可执行动作类型及触发条件、工具配置、参数绑定与治理约束，典型内容包括：名称与说明、触发条件、影响的对象类、风险等级。

#### RiskType

描述高风险操作或场景的治理规则，典型内容包括：名称与说明、管控范围、管控策略、前置检查、回滚方案、审计要求。

> 各类型的完整字段定义、语法规范与示例详见 [SPECIFICATION.md](./SPECIFICATION.md)。

### 3.3 Skill-like Metadata

为增强可发现性、可治理性和可编排性，在 Frontmatter 中引入推荐字段：

- 标识类：`id`、`name`、`description`、`version`、`branch`
- 组织类：`author`、`status`
- 能力类：`capabilities`、`tags`
- 系统类：`created_at`、`updated_at`

这些字段既服务于文档自描述，也为审计、检索和后续治理扩展提供结构化基础。

### 3.4 定义级 Checksum

Checksum 的目标是识别”哪个定义变了”，而不是”哪个文件变了”。

- **算法**：SHA-256
- **格式**：`sha256:<前16位hex>`
- **粒度**：每个 `Network` / `ObjectType` / `RelationType` / `ActionType` / `RiskType` 定义分别计算
- **规范化**：统一换行、去首尾空白、去行尾空格、压缩连续空行
- **存储位置**：统一写入网络根目录的 `CHECKSUM` 文件，由 `bkn-tools checksum` 生成，不内嵌到各定义文件的 Frontmatter 中

`CHECKSUM` 文件示例：

```text
network  sha256:1234567890abcdef
object_type:pod  sha256:a1b2c3d4e5f67890
relation_type:pod_belongs_node  sha256:1a2b3c4d5e6f7890
action_type:restart_pod  sha256:f1e2d3c4b5a69870
risk_type:restart_pod_high_risk  sha256:abcd1234ef567890
```

后端持久化 checksum 作为 diff 基准；`CHECKSUM` 文件用于离线校验与导出一致性。

### 3.5 导入、导出与一致性

系统存在两条定义写入路径：

1. **文件导入路径**：`.bkn` → `bkn-tools` / `bkn-sdk` → Importer → RESTful API → Storage
2. **结构化 API 路径**：RESTful API → Storage → Serializer → `.bkn` 回写

设计原则：**系统以数据库中的 JSON 结构化存储为主；正常情况下由 Engine 保证 JSON 与 `.bkn` 文本一致，若因异常变更导致二者不一致，则以 JSON 为准，并由导出/回写流程修复 `.bkn`。**

导入语义如下：

- 唯一键：`(kn_id, branch, definition_type, definition_id)`
- 默认更新：`replace`
- 首次导入：全部视为新增
- 增量导入：按定义级 checksum 决定 create / update / skip

**依赖方向**：

```text
bkn-tools CLI ------> BKN Engine <------- Context Loader <------- Agent
  |                      |
  |                内嵌 bkn-sdk
  |
  +------> .bkn 文件
```

- bkn-tools CLI 依赖 BKN Engine（通过 REST API）和 .bkn 文件（通过文件系统）
- BKN Engine 依赖 bkn-sdk（作为内嵌库，编译期依赖）
- Context Loader 依赖 BKN Engine（通过内部 API）
- 无环形依赖

### 3.6 Agent 消费模式

当前评审范围内，运行时消费模式统一为 **平台托管模式**：

- **平台托管模式**：Agent 通过 Context Loader 消费平台知识网络，适用于统一权限、审计和在线检索场景。

文件化目录与 `.bkn` 资产主要服务于定义管理、导入导出和版本协作；运行时消费仍以平台托管模式为主。

---

## 4. 风险与边界 (Risks & Edge Cases)

- **双写风险**：API 直写与 BKN 导入并存，可能造成 JSON 与 `.bkn` 状态漂移；数据库存储中以 JSON 为主，若因异常变更导致二者不一致，则以 JSON 为准，并通过导出或回写重新修复 `.bkn`。
- **Checksum 碰撞**：理论上存在，但概率极低，可接受。
- **格式伪变更**：通过 normalization 和统一序列化输出规避。
- **引用失效**：需在 Validator 阶段校验依赖、引用与作用域。
- **RiskType 落地不完整**：短期先落到语言与存储层，后续再接审批与自动回滚链路。

---

## 5. 替代方案 (Alternatives Considered)

### 方案 A：仅增强现有 RESTful API

优点是改造范围小；缺点是无法解决可读性、声明式管理和 Agent 友好问题，因此不采用。

### 方案 B：文件级 Checksum

优点是实现简单；缺点是粒度过粗，无法支撑精确的增量导入，因此不采用。

### 方案 C：外置 `checksums.json`

优点是不改 BKN 文件；缺点是增加维护复杂度且容易不同步，因此不采用。

### 最终选择

采用 **Markdown + Frontmatter + 定义级 Checksum + `bkn-sdk` 基础能力 + `bkn-tools` 工具链 + 现有 API 落地** 的组合方案，在兼容现有系统的同时，提供更适合设计协作和 Agent 消费的知识建模能力。

---

## 6. 任务拆分 (Milestones)

- [ ] 完成 `DESIGN.md` 与 `SPECIFICATION.md` 的字段、章节和示例对齐
- [ ] 补齐 `RiskType` 的规范与示例
- [ ] 实现 `bkn-sdk` 的 Parser / Validator / Differ / Serializer / Checksum
- [ ] 提供 `bkn-tools` 的 `validate` / `diff` / `import` / `export`
- [ ] 在 BKN Engine 中接入导入、导出与 BKN 文本持久化
- [ ] 打通 Context Loader 与平台托管消费模式

---

## 参考

- [SPECIFICATION.md](./SPECIFICATION.md)
