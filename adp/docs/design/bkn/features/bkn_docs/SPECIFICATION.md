# BKN 语言规范

版本: 1.0.0
spec_version: 1.0.0

## 概述

BKN (Business Knowledge Network) 是一种基于 Markdown 的声明式建模语言，用于定义业务知识网络中的对象、关系和行动。BKN 只负责描述模型结构与语义，不包含执行逻辑——校验引擎、数据管道、工作流等运行时能力由消费 BKN 模型的平台实现。

本文档定义了 BKN 的完整语法规范。

### 术语表（Glossary）

**核心概念**

| 术语 | 含义 |
|------|------|
| BKN | Business Knowledge Network，业务知识网络 |
| knowledge_network | 一个业务知识网络的整体集合 |
| object | 业务对象类型（例如 Pod/Node/Service） |
| relation | 连接两个 object 的关系类型（例如 belongs_to/routes_to） |
| action | 对 object 执行的操作定义（可绑定 tool 或 mcp） |
| risk | 风险类，对行动类和对象类的执行风险进行结构化建模 |

**对象结构**

| 术语 | 含义 |
|------|------|
| data_view | 数据视图，对象/关系可直接映射的数据来源 |
| data_properties | 对象的属性定义表，声明字段类型、主键、展示键等 |
| property_override | 属性覆盖，对继承属性的索引、约束等进行特殊配置 |
| logic_properties | 逻辑属性，基于其他数据源的派生字段（metric / operator） |
| primary_key | 主键字段，用于唯一定位实例（Data Properties 表中标记 YES） |
| display_key | 展示字段，用于 UI 显示和检索（Data Properties 表中标记 YES） |
| constraint | 属性值域约束，声明实例数据的合法范围（如 `>= 0`、`in(...)`） |
| metric | 逻辑属性类型：指标，从外部数据源获取的度量值 |
| operator | 逻辑属性类型：算子，基于输入参数的计算逻辑 |

**行动结构**

| 术语 | 含义 |
|------|------|
| trigger_condition | 触发条件，定义 action 自动执行的条件 |
| pre-conditions | 前置条件，执行前必须满足的数据检查（不满足则阻止执行） |
| tool | 行动绑定的外部工具 |
| mcp | Model Context Protocol，行动绑定的 MCP 工具 |
| schedule | 定时配置（FIX_RATE 或 CRON），用于周期性执行 |
| scope_of_impact | 影响范围，声明行动影响的对象 |

**文件组织**

| 术语 | 含义 |
|------|------|
| frontmatter | YAML 元数据区（`---` 包裹），每个 .bkn 文件的头部 |
| network | 文件类型 `type: network`，完整知识网络的顶层容器 |
| data | 文件类型 `type: data`，实例数据文件（建议使用 `.bknd` 扩展名） |
| namespace | 命名空间，用于大规模组织与避免 ID 冲突 |
| spec_version | 规范版本号，标识文件遵循的 BKN 规范版本 |

### 标准原语表 (Primitives)

Section 标题和表格列名的规范形式，建议使用英文。解析器应同时支持英文与中文以便兼容。

下表按 **统一标题层级** 组织，适用于所有文件类型（network / object_type / relation_type / action_type / risk_type / data）。

| Level | English (canonical) | Definition | 中文 | Syntax |
|:-----:|---------------------|------------|------|--------|
| `#` | Objects | Section: all object definitions in this file | 对象定义 | `# Objects` |
| `#` | Relations | Section: all relation definitions | 关系定义 | `# Relations` |
| `#` | Actions | Section: all action definitions | 行动定义 | `# Actions` |
| `#` | Risks | Section: all risk definitions | 风险定义 | `# Risks` |
| `##` | Object | Individual object definition | 对象 | `## Object: {id}` |
| `##` | Relation | Individual relation definition | 关系 | `## Relation: {id}` |
| `##` | Action | Individual action definition | 行动 | `## Action: {id}` |
| `##` | Risk | Individual risk definition | 风险 | `## Risk: {id}` |
| `###` | Data Source | The data view this object maps from | 数据来源 | `### Data Source` |
| `###` | Data Properties | Explicit list of fields (name, type, PK, index) | 数据属性 | `### Data Properties` |
| `###` | Property Override | Per-property overrides (e.g. index config) | 属性覆盖 | `### Property Override` |
| `###` | Logic Properties | Derived fields: metric, operator | 逻辑属性 | `### Logic Properties` |
| `###` | Business Semantics | Human-readable meaning of the object/relation | 业务语义 | `### Business Semantics` |
| `###` | Endpoints | Relation endpoints: source, target, type | 关联定义 | `### Endpoints` |
| `###` | Mapping Rules | How source/target properties map | 映射规则 | `### Mapping Rules` |
| `###` | Mapping View | For data_view relations: the join view | 映射视图 | `### Mapping View` |
| `###` | Source Mapping | Map source object props to view | 起点映射 | `### Source Mapping` |
| `###` | Target Mapping | Map view to target object props | 终点映射 | `### Target Mapping` |
| `###` | Bound Object | Object this action operates on | 绑定对象 | `### Bound Object` |
| `###` | Trigger Condition | When to run (YAML condition) | 触发条件 | `### Trigger Condition` |
| `###` | Pre-conditions | Data conditions required before action execution | 前置条件 | `### Pre-conditions` |
| `###` | Tool Configuration | tool or MCP binding | 工具配置 | `### Tool Configuration` |
| `###` | Parameter Binding | param name, source, binding | 参数绑定 | `### Parameter Binding` |
| `###` | Schedule | FIX_RATE or CRON | 调度配置 | `### Schedule` |
| `###` | Scope of Impact | What objects are affected | 影响范围 | `### Scope of Impact` |
| `####` | {property_name} | Individual logic property sub-section | — | `#### {name}` |
| — | Primary Key | Field that uniquely identifies an instance | 主键 | Data Properties table column |
| — | Display Key | Field used for UI label / search display | 显示属性 | Data Properties table column |
| — | Operation | add \| modify \| delete | 行动类型 | table column |

表格列名（canonical）：Type, ID, Name, Property, Display Name, Type, Constraint, Primary Key, Display Key, Index, Index Config, Description; Source, Target, Required, Min, Max; Source Property, Target Property; Parameter, Type, Source, Binding, Description; Bound Object, Operation; Object, Check, Condition, Message; Object, Impact Description。解析器同时接受中文列名。

## 文件格式

### 文件扩展名

- `.bkn` - BKN 定义文件（schema），推荐
- `.bknd` - BKN 数据文件（instance data），推荐
- `.md` - 兼容载体，运行时支持；内容必须满足 BKN frontmatter/type/结构约束

**`.md` 兼容模式**：可用 `.md` 保存 BKN 内容，便于跨平台文档化与协作。运行时加载时，`.md` 与 `.bkn/.bknd` 走同一解析与校验路径；若缺少 frontmatter、`type` 或结构不符合要求，将直接报错。推荐实践：schema 优先 `.bkn`，data 优先 `.bknd`，`.md` 用于需与通用 Markdown 工具共存的场景。

### 文件编码

- UTF-8

### 基本结构

每个 BKN 文件由两部分组成：

1. **YAML Frontmatter**: 文件元数据
2. **Markdown Body**: 知识网络定义内容

```markdown
---
type: network
id: example-network
name: Example Network
version: 1.0.0
---

# Network Title

Network description...

## Object: object_id

Object definition...

## Relation: relation_id

Relation definition...

## Action: action_id

Action definition...
```

---

## Frontmatter 规范

### 工程可控性字段（推荐）

为支持规模化协作、审批与审计，建议在定义文件中使用以下字段：

| 字段 | 适用 type | 说明 |
|------|----------|------|
| `spec_version` | all | 该文件使用的规范版本（默认继承文档 spec_version） |
| `namespace` | object_type/relation_type/action_type | 命名空间/包名，用于大规模组织与避免冲突（例如 `platform.k8s`） |
| `owner` | object_type/relation_type/action_type | 负责人/团队（用于审计与审批路由） |
| `enabled` | action_type | 是否启用（建议默认 `false`，导入不等于启用） |
| `risk_level` | action_type | 风险等级（`low|medium|high`，用于审批与发布策略） |
| `requires_approval` | action_type | 是否需要审批才能启用/执行 |

### 文件类型 (type)

| type | 说明 | 用途 |
|------|------|------|
| `network` | 完整知识网络 | 包含多个定义的网络文件 |
| `object_type` | 单个对象定义 | 独立的对象文件，可直接导入 |
| `relation_type` | 单个关系定义 | 独立的关系文件，可直接导入 |
| `action_type` | 单个行动定义 | 独立的行动文件，可直接导入 |
| `risk_type` | 单个风险定义 | 独立的风险类文件，可直接导入 |
| `data` | 数据文件 | 承载 object/relation 的实例数据行（建议 `.bknd`） |

### 网络文件 (type: network)

```yaml
---
type: network                    # 完整知识网络
id: string                       # 网络ID，唯一标识
name: string                     # 网络显示名称
version: string                  # 版本号 (semver)
tags: [string]                   # 可选，标签列表
description: string              # 可选，网络描述
includes: [string]               # 可选，引用的其他文件
---
```

### 单对象文件 (type: object_type)

```yaml
---
type: object_type                # 单个对象定义
id: string                       # 对象ID，唯一标识
name: string                     # 对象显示名称
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
tags: [string]                   # 可选，标签列表
---
```

### 单关系文件 (type: relation_type)

```yaml
---
type: relation_type              # 单个关系定义
id: string                       # 关系ID，唯一标识
name: string                     # 关系显示名称
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
---
```

### 单行动文件 (type: action_type)

```yaml
---
type: action_type                # 单个行动定义
id: string                       # 行动ID，唯一标识
name: string                     # 行动显示名称
operation: add | modify | delete     # 行动类型（add/modify/delete）
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填，保证导入确定性）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
enabled: boolean                 # 可选，是否启用（建议默认 false）
risk_level: low | medium | high  # 可选，静态风险等级
requires_approval: boolean       # 可选，是否需要审批
---
```

> **动态 risk 属性**：Action 的运行时属性 `risk`（取值 `allow` | `not_allow` | `unknown`）由内置或用户提供的风险评估函数根据当前场景与带 `__risk__` tag 的知识计算，不在此 frontmatter 中声明。无规则或无匹配时返回 `unknown`，由执行侧按业务策略处理。

### 单风险文件 (type: risk_type)

```yaml
---
type: risk_type                  # 单个风险定义
id: string                       # 风险类ID，唯一标识
name: string                     # 风险类显示名称
version: string                  # 可选，版本号
network: string                  # 所属网络ID（建议必填）
namespace: string                # 可选，命名空间/包名
owner: string                    # 可选，负责人/团队
---
```

---

## 数据文件规范（.bknd / type: data）

`.bknd` 文件使用与 `.bkn` 相同的 Markdown 语法，但正文承载的是实例数据表，而非对象/关系/行动定义。

### Frontmatter

```yaml
---
type: data
network: recoverable-network
object: scenario            # 与 relation 二选一
source: PFMEA模板.xlsx      # 可选，数据来源
---
```

- `type` 必须为 `data`
- `object` 或 `relation` 二选一，用于指向对应 `.bkn` 中定义的 ID
- `network` 建议填写，保持与 schema 文件一致
- `source` 为可选字段，用于记录数据来源

### Body

正文由一个标题（`#` 或 `##`）+ 一个 GFM 表格组成。表头应与目标 object 的 Data Properties 列名（或 relation 的映射字段）一致。

```markdown
# scenario

| scenario_id | name | category | primary_object | description |
|-------------|------|----------|----------------|-------------|
| ops-rm-rf | rm -rf 删除备份存储 | integrity | backup_system | 直接销毁备份数据 |
```

### 约束

- 列名应与 schema 定义保持一致，避免隐式字段
- 每个 `type: data` 文件建议只包含一个数据表，便于版本化和审计

### Data Source 与可编辑性

对象的 Data Source 决定其数据是否可写入 `.bknd`：

| Data Source Type | 数据来源 | `.bknd` 是否允许 |
|------------------|----------|------------------|
| `data_view` | 外部系统（ERP、数据库、API） | **否**，数据由外部系统维护，`.bknd` 不适用 |
| `bknd` | BKN 知识原生数据 | **是**，`.bknd` 为数据源，可读写 |
| 无 Data Source | 平台默认 | 视平台实现而定 |

- 当 Object 的 Data Source 为 `data_view` 时，不应为该对象创建 `.bknd` 文件，数据由外部系统提供
- 当 Object 的 Data Source 为 `bknd` 时，数据存储在 `.bknd` 文件中，可编辑、可版本化

---

## 对象定义规范

### 语法

```markdown
## Object: {object_id}

**{display_name}** - {brief_description}

- **Tags**: {tag1}, {tag2}     (可选，定义级标签)
- **Owner**: {owner}          (可选，负责人/团队)

### Data Source

| Type | ID | Name |
|------|-----|------|
| data_view | {view_id} | {view_name} |
| bknd | {object_id} | {display_name} |

- `data_view`：数据来自外部系统，不可用 `.bknd` 维护
- `bknd`：知识原生数据，由 `.bknd` 文件承载，可编辑

### Data Properties

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| {prop} | {name} | {type} | | {desc} | YES | | YES |
| {prop} | {name} | {type} | | {desc} | | YES | |

- `Primary Key`：标记为 `YES` 的属性用于唯一定位实例，至少一个
- `Display Key`：标记为 `YES` 的属性用于 UI 展示和检索显示，至少一个
- `Constraint` 列为可选，声明该属性在实例数据层面的合法值范围；留空表示无约束。语法见下文"Constraint 列语法"

### Property Override

(optional) Declare only properties needing special configuration

| Property | Display Name | Index Config | Constraint | Description |
|----------|--------------|--------------|------------|-------------|
| ... | ... | ... | ... | ... |

#### Index Config 语法

`Index Config` 列支持组合式语法，多个索引类型用 ` + ` 连接。可在括号内传递可选参数：

| 类型 | 语法 | 说明 |
|------|------|------|
| keyword | `keyword` | 基础关键字索引 |
| keyword | `keyword(max_len)` | 关键字索引，`max_len` 为 ignore_above_len |
| fulltext | `fulltext` | 全文索引，默认分析器 |
| fulltext | `fulltext(analyzer)` | 全文索引，指定分析器（如 standard、ik_max_word） |
| vector | `vector` | 向量索引，默认模型 |
| vector | `vector(model_id)` | 向量索引，指定 embedding 模型 ID |

示例：`keyword(1024) + fulltext(standard) + vector(1951511856216674304)`

### Logic Properties

#### {property_name}

- **Type**: metric | operator
- **Source**: {source_id} ({source_type})
- **Description**: {description}

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| ... | string | property | {property_name} | 从对象属性绑定 |
| ... | array | input | - | 运行时用户输入 |
| ... | string | const | {value} | 常量值 |
```

- `Type`：参数数据类型，如 string、number、boolean、array
- `Source`：值来源，`property`（对象属性）/ `input`（用户输入）/ `const`（常量）
- `Binding`：当 Source 为 property 时为属性名，为 const 时为常量值，为 input 时为 `-`

### 定义级元数据

在 `## Object:` 或 `## Relation:` 的定义头部（在 `### Data Source` 或 `### Endpoints` 之前），可使用可选的 inline 元数据行：

- **Tags**：该定义的标签列表（逗号分隔），用于分类、筛选和审计
- **Owner**：负责人或团队，用于审批路由和审计

在 network 文件中，多个对象或关系可各自拥有不同的 tags 和 owner。

### 风险相关定义（内置 tag `__risk__`）

- **保留标签**：**`__risk__`** 为规范内置保留 tag，仅用于参与「内置风险评估」的对象与关系；**用户不得将 `__risk__` 用于自定义用途**，以避免与内置逻辑冲突。
- 在需要参与**内置**风险评估的对象、关系定义头部增加 `- **Tags**: __risk__`。AI 应用与内置评估模块通过该 tag 识别风险相关定义。
- Action 拥有**运行时/计算属性** `risk`（见「行动定义规范」），取值 `allow` | `not_allow` | `unknown`，可由内置或用户提供的风险评估函数根据当前场景与带 `__risk__` tag 的数据计算得出，**不写入 BKN 文件**。无规则或无匹配时返回 `unknown`。

**开放性**：用户可按自身需求定义**自己的风险类**（使用**非保留** tag，如 `compliance`、`audit` 等）及**自己的风险评估函数**；内置的 `__risk__` 与默认评估逻辑仅为一种可选实现，不排斥用户扩展或替换。

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {object_id} | YES | 对象唯一标识，小写字母、数字、下划线 |
| {display_name} | YES | 人类可读名称 |
| Data Source | NO | 映射的数据视图，未设定时由平台自动管理 |
| Data Properties | YES | 属性定义，须标记 Primary Key 和 Display Key |
| Property Override | NO | 需要特殊配置的属性（索引、约束等） |
| Logic Properties | NO | 指标、算子等扩展属性 |

### 数据类型

Data Properties 表的 `Type` 列使用以下标准类型。类型名称大小写不敏感，推荐使用下表中的规范形式。

| 类型 | 说明 | JSON 对应 | SQL 对应 |
|------|------|-----------|----------|
| int32 | 32 位有符号整数 | number | INT / INTEGER |
| int64 | 64 位有符号整数 | number | BIGINT |
| integer | 泛型整数（精度未指定） | number | 平台相关（通常 int64） |
| float32 | 32 位浮点数 | number | FLOAT / REAL |
| float64 | 64 位浮点数 | number | DOUBLE / DOUBLE PRECISION |
| float | 泛型浮点数（精度未指定） | number | 平台相关（通常 float64） |
| decimal(p,s) | 精确十进制数，p 为精度，s 为小数位 | string / number | DECIMAL(p,s) / NUMERIC(p,s) |
| decimal | 泛型精确十进制（精度未指定） | string / number | 平台相关 |
| bool | 布尔值 | boolean | BOOLEAN |
| VARCHAR | 变长字符串 | string | VARCHAR / TEXT |
| TEXT | 长文本 | string | TEXT / CLOB |
| DATE | 日期（无时间） | string (ISO 8601) | DATE |
| TIME | 时间（无日期） | string (ISO 8601) | TIME |
| TIMESTAMP | 日期时间（含时区） | string (ISO 8601) | TIMESTAMP |
| JSON | JSON 结构数据 | object / array | JSON / JSONB |
| BINARY | 二进制数据 | string (base64) | BLOB / BYTEA |

> 当数据源使用的类型不在上表中时，可直接使用数据源原生类型名称（如 `ARRAY<VARCHAR>`），解析器应透传不识别的类型。

### 配置模式

#### 模式一：映射 + 最小属性声明

映射视图，仅声明主键和展示键：

```markdown
## Object: node

**Node**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_123 |

### Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| node_name | | YES |
```

#### 模式二：映射 + 属性覆盖

映射视图，声明键并配置需要特殊处理的属性：

```markdown
## Object: pod

**Pod Instance**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_456 |

### Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| pod_name | | YES |

### Property Override

| Property | Index Config | Constraint | Description |
|----------|--------------|------------|-------------|
| pod_status | fulltext(standard) + vector | in(Running,Pending,Failed,Unknown) | 支持全文和语义搜索 |
```

#### 模式三：完整定义

完整声明所有属性（含类型、约束、索引）：

```markdown
## Object: service

**Service**

### Data Source

| Type | ID |
|------|-----|
| data_view | view_789 |

### Data Properties

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| id | ID | int64 | | Primary key | YES | | YES |
| service_name | Name | VARCHAR | not_null | Service name | | YES | YES |
| service_type | Service Type | VARCHAR | in(ClusterIP,NodePort,LoadBalancer) | Service type | | | |
```

---

## 关系定义规范

### 语法

```markdown
## Relation: {relation_id}

**{display_name}** - {brief_description}

- **Tags**: {tag1}, {tag2}     (可选，定义级标签)
- **Owner**: {owner}          (可选，负责人/团队)

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| {source} | {target} | direct \| data_view | YES \| NO | 0 | - |

- `Required`: YES/NO，是否必须存在至少一条关系（从 Source 侧看）
- `Min`: 最小关系数，默认 0
- `Max`: 最大关系数，`-` 表示无限制
- Required / Min / Max 均为可选列，省略时不做约束

### Mapping Rules

| Source Property | Target Property |
|------------------|-----------------|
| {source_prop} | {target_prop} |

### Business Semantics

(optional) Human-readable meaning of the relation...
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {relation_id} | YES | 关系唯一标识 |
| Source | YES | 起点对象 ID |
| Target | YES | 终点对象 ID |
| Type | YES | `direct` (直接映射) 或 `data_view` (视图映射) |
| Mapping Rules | YES | 属性映射关系 |
| Required | NO | 是否必须存在至少一条关系（从 Source 侧看） |
| Min | NO | 最小关系数 |
| Max | NO | 最大关系数，`-` 表示无限制 |

### 关系类型

#### 直接映射 (direct)

通过属性值匹配建立关联：

```markdown
## Relation: pod_belongs_node

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| pod | node | direct | YES | 1 | 1 |

每个 Pod 必须属于且仅属于一个 Node。

### Mapping Rules

| Source Property | Target Property |
|------------------|-----------------|
| pod_node_name | node_name |
```

#### 视图映射 (data_view)

通过中间视图建立关联：

```markdown
## Relation: user_likes_post

| Source | Target | Type | Required | Min | Max |
|--------|--------|------|----------|-----|-----|
| user | post | data_view | NO | 0 | - |

### Mapping View

| Type | ID |
|------|-----|
| data_view | user_post_likes_view |

### Source Mapping

| Source Property | View Property |
|-----------------|----------------|
| user_id | uid |

### Target Mapping

| View Property | Target Property |
|---------------|-----------------|
| pid | post_id |
```

---

## 行动定义规范

### 语法

```markdown
## Action: {action_id}

**{display_name}** - {brief_description}

| Bound Object | Operation |
|--------------|-----------|
| {object_id} | add | modify | delete |

### Trigger Condition

```yaml
field: {property_name}
operation: == | != | > | < | >= | <= | in | not_in | exist | not_exist
value: {value}
```

### Pre-conditions

(optional) 执行前的数据前置条件，不满足则阻止行动执行

| Object | Check | Condition | Message |
|--------|-------|-----------|---------|
| {object_id} | relation:{relation_id} | exist | 违反时的说明 |
| {object_id} | property:{property_name} | {op} {value} | 违反时的说明 |

- `Check`: `property:{name}` 或 `relation:{id}`，指明检查目标
- `Condition`: 复用 Trigger Condition 操作符语法
- Trigger 决定「何时触发」，Pre-conditions 决定「能否执行」

### Tool Configuration

| Type | Tool ID |
|------|---------|
| tool | {tool_id} |

or

| Type | MCP |
|------|-----|
| mcp | {mcp_id}/{tool_name} |

### Parameter Binding

| Parameter | Type | Source | Binding | Description |
|-----------|------|--------|---------|-------------|
| {param_name} | string | property | {property_name} | {说明} |
| {param_name} | string | input | - | {说明} |
| {param_name} | string | const | {value} | {说明} |

### Schedule

(optional)

| Type | Expression |
|------|------------|
| FIX_RATE | {interval} |
| CRON | {cron_expr} |

### Execution Description

(optional) Detailed execution flow...
```

### 治理要求（强烈建议）

行动定义连接执行面（tool/mcp），为了稳定性与安全性，建议在每个 Action 中**显式写清**以下四类信息，并在工程侧落地相应治理：

1. **触发**：何时触发（手动/定时/条件触发），触发条件是否可复现
2. **影响范围**：影响哪些对象、范围边界、预期副作用
3. **权限与前置条件**：谁可以导入/启用/执行，是否需要审批，依赖的权限/凭据
4. **回滚/失败策略**：失败处理、重试策略、熔断/限流、可撤销性

> 推荐实践：导入不等于启用；启用与执行需要独立的权限与审计日志，并能关联到对应的 BKN 定义版本。

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| {action_id} | YES | 行动唯一标识 |
| Bound Object | YES | 目标对象 ID |
| Operation | YES | `add` / `modify` / `delete` |
| Trigger Condition | NO | 自动触发的条件 |
| Pre-conditions | NO | 执行前的数据前置条件 |
| Tool Configuration | YES | 执行的工具或 MCP |
| Parameter Binding | YES | 参数来源配置 |
| Schedule | NO | 定时执行配置 |
| risk（计算属性） | - | 运行时属性，取值 `allow` \| `not_allow` \| `unknown`，由内置或用户提供的风险评估函数根据场景与带 `__risk__` tag 的对象/关系计算，不写入 BKN。无规则或无匹配时返回 `unknown` |

### 触发条件操作符

以下操作符适用于 Trigger Condition、Pre-conditions 以及 Data Properties / Property Override 表中的 Constraint 列：

| 操作符 | 说明 | 示例 |
|--------|------|------|
| == | 等于 | `value: Running` |
| != | 不等于 | `value: Running` |
| > | 大于 | `value: 100` |
| < | 小于 | `value: 100` |
| >= | 大于等于 | `value: 100` |
| <= | 小于等于 | `value: 100` |
| in | 包含于 | `value: [A, B, C]` |
| not_in | 不包含于 | `value: [A, B, C]` |
| exist | 存在 | (无需 value) |
| not_exist | 不存在 | (无需 value) |
| range | 范围内 | `value: [0, 100]` |
| not_null | 不为空 | (无需 value，约束专用) |
| regex | 正则匹配 | `value: "^[a-z]+$"`（约束专用） |

### Constraint 列语法

`Constraint` 列出现在 Object 的 **Data Properties** 和 **Property Override** 表格中，用于声明该属性在实例数据层面的合法值范围。该列为可选，留空表示无约束。

#### 语法格式

每条约束写在单个表格单元格内，格式为 **`operator`** 或 **`operator(args)`** 或 **`operator value`**。

| 类别 | 语法 | 含义 | 适用类型 | 示例 |
|------|------|------|----------|------|
| 比较 | `== value` | 等于固定值 | 数值、字符串 | `== 1` |
| 比较 | `!= value` | 不等于固定值 | 数值、字符串 | `!= 0` |
| 比较 | `> value` | 大于 | 数值 | `> 0` |
| 比较 | `< value` | 小于 | 数值 | `< 1000` |
| 比较 | `>= value` | 大于等于 | 数值 | `>= 0` |
| 比较 | `<= value` | 小于等于 | 数值 | `<= 100` |
| 范围 | `range(min,max)` | 闭区间 [min, max] | 数值 | `range(0,100)` |
| 枚举 | `in(v1,v2,…)` | 值必须为列表之一 | 字符串、数值 | `in(Running,Pending,Failed)` |
| 枚举 | `not_in(v1,v2,…)` | 值不能为列表之一 | 字符串、数值 | `not_in(Deleted,Archived)` |
| 存在性 | `not_null` | 值不能为空 | 任意 | `not_null` |
| 存在性 | `exist` | 属性必须存在 | 任意 | `exist` |
| 存在性 | `not_exist` | 属性不能存在 | 任意 | `not_exist` |
| 正则 | `regex:pattern` | 值须匹配正则表达式 | 字符串 | `regex:^[a-z0-9_]+$` |

#### 组合约束

当一个属性需要多条约束时，使用 `; ` （分号 + 空格）分隔：

```
not_null; >= 0
not_null; regex:^[a-z_]+$
>= 0; <= 100
not_null; in(ClusterIP,NodePort,LoadBalancer)
```

组合约束表示 **逻辑 AND**——所有约束必须同时满足。

#### 完整示例

| Property | Display Name | Type | Constraint | Description | Primary Key | Display Key | Index |
|----------|--------------|------|------------|-------------|:-----------:|:-----------:|:-----:|
| id | ID | int64 | not_null | 主键 | YES | | YES |
| name | 名称 | VARCHAR | not_null; regex:^[a-z0-9_]+$ | 唯一标识名 | | YES | YES |
| quantity | 数量 | int32 | >= 0 | 不允许负数 | | | |
| status | 状态 | VARCHAR | in(Active,Inactive,Archived) | 枚举值 | | | YES |
| score | 评分 | float64 | range(0,100) | 百分制 | | | |
| priority | 优先级 | int32 | not_null; range(1,5) | 1-5 级 | | | |

### 参数来源

| 来源 | 说明 |
|------|------|
| property | 从对象属性获取 |
| input | 运行时用户输入 |
| const | 常量值 |

---

## 风险类定义规范

风险类（Risk）是第四种基本类型，用于对行动类和对象类的执行风险进行结构化建模。风险类是独立类型，不是行动类的附属字段；Action 的 `risk_level` 声明「多危险」，Risk 声明「如何管控」。

### 语法

```markdown
## Risk: {risk_id}

**{显示名称}** - {简短描述}

### 管控范围

| 管控对象 | 管控行动 | 风险等级 |
|----------|----------|----------|
| {object_id} | {action_id} | low | medium | high |

### 管控策略

| 条件 | 策略 |
|------|------|
| {condition_description} | {strategy_description} |

### 前置检查

(可选)

| 检查项 | 类型 | 说明 |
|--------|------|------|
| {check_name} | permission | {description} |
| {check_name} | simulation | {description} |
| {check_name} | approval | {description} |
| {check_name} | precondition | {description} |

### 回滚方案

(可选) 失败或误操作时的恢复策略说明...

### 审计要求

(可选) 审计日志、告警通知等要求说明...
```

### 字段说明

| 字段 | 必须 | 说明 |
|------|:----:|------|
| risk_id | YES | 风险类唯一标识，小写字母、数字、下划线 |
| 显示名称 | YES | 人类可读名称 |
| 管控范围 | YES | 关联的对象类、行动类及风险等级 |
| 管控策略 | YES | 按条件分级的管控规则（至少一条） |
| 前置检查 | NO | 执行前检查项列表 |
| 回滚方案 | NO | 失败恢复策略 |
| 审计要求 | NO | 审计日志与告警配置 |

### 风险等级

| 等级 | 含义 | 默认管控 |
|------|------|---------|
| `low` | 只读查询或无副作用操作 | 无需审批，直接执行 |
| `medium` | 有副作用但可撤销的操作 | 建议确认，记录审计日志 |
| `high` | 不可逆或高影响操作 | 必须审批 + 行动模拟 + 完整审计 |

---

## 通用语法元素

### 表格格式

使用标准 Markdown 表格：

```markdown
| 列1 | 列2 | 列3 |
|-----|-----|-----|
| 值1 | 值2 | 值3 |
```

居中对齐（用于布尔值）：

```markdown
| 列1 | 列2 |
|-----|:---:|
| 值1 | YES |
```

### YAML 代码块

用于复杂结构（如条件表达式）：

```markdown
```yaml
condition:
  operation: and
  sub_conditions:
    - field: status
      operation: ==
      value: Failed
    - field: retry_count
      operation: <
      value: 3
`` `
```

### Mermaid 图表

用于可视化关系：

```markdown
```mermaid
graph LR
    A --> B
    B --> C
`` `
```

### 引用块

用于关键信息高亮：

```markdown
> **注意**: 该对象变更需要审批流程
```

### 标题层级

标题层级在所有文件类型中保持一致：

- `#` - 文档/分组标题（例如网络标题，或 `# Objects` / `# Relations` / `# Actions` / `# Risks`）
- `##` - 类型定义（`## Object:` / `## Relation:` / `## Action:` / `## Risk:`）
- `###` - 定义内 section（Data Source, Data Properties, Mapping Rules, Trigger Condition 等）
- `####` - 子项（例如逻辑属性名）

> 规则：不再区分“单定义文件层级上移”；所有定义统一使用上述层级。

---

## 文件组织

### 一定义一文件

每个对象/关系/行动/风险独立一个文件。支持两种编排入口：

**模式 A：SKILL.md 编排**（Agent Skill 标准）

```
{business_dir}/
├── SKILL.md                     # agentskills.io 标准入口，含网络拓扑、索引、使用指南
├── CHECKSUM                     # 可选，目录级一致性校验（bkn-tools checksum 生成）
├── object_types/
│   ├── material.bkn             # type: object_type
│   └── inventory.bkn           # type: object_type
├── relation_types/
│   └── material_to_inventory.bkn # type: relation_type
├── action_types/
│   ├── check_inventory.bkn      # type: action_type
│   └── adjust_inventory.bkn    # type: action_type
├── risk_types/
│   └── inventory_adjustment_risk.bkn  # type: risk_type
└── data/                        # 可选，.bknd 实例数据
    └── scenario.bknd
```

**模式 B：index.bkn 编排**（传统 index 模式）

```
{business_dir}/
├── index.bkn                    # type: network，作为网络入口
├── CHECKSUM                     # 可选，目录级一致性校验（bkn-tools checksum 生成）
├── object_types/
│   ├── pod.bkn                  # type: object_type
│   ├── node.bkn                 # type: object_type
│   └── service.bkn              # type: object_type
├── relation_types/
│   ├── pod_belongs_node.bkn     # type: relation_type
│   └── service_routes_pod.bkn   # type: relation_type
├── action_types/
│   ├── restart_pod.bkn          # type: action_type
│   └── cordon_node.bkn          # type: action_type
├── risk_types/
│   └── pod_restart_risk.bkn     # type: risk_type
└── data/                        # 可选，.bknd 实例数据
```

目录名（`object_types/`、`relation_types/`、`action_types/`、`risk_types/`、`data/`）为约定，文件的 `type` 字段为定义类型的权威声明。

**单对象文件示例** (`pod.bkn`):

```markdown
---
type: object_type
id: pod
name: Pod Instance
network: k8s-network
---

## Object: pod

**Pod Instance**

Minimal deployable unit in Kubernetes.

### Data Source

| Type | ID |
|------|-----|
| data_view | view_123 |

### Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| pod_name | | YES |
```

---

## 增量导入规范

BKN 支持将任何 `.bkn` 文件动态导入到已有的知识网络。

### 导入器能力要求（工程可控性 9+ 的前提）

建议实现一个 **BKN Importer**，将 BKN 文件转换为系统变更，并提供以下能力（缺一不可）：

| 能力 | 说明 | 目的 |
|------|------|------|
| `validate` | 结构/表格/YAML block 校验，引用完整性校验，参数绑定校验 | 阻止错误进入系统 |
| `diff` | 计算变更集（新增/更新/删除）与影响范围 | 让变更可解释、可审计 |
| `dry_run` | 在不落地的情况下执行 validate + diff | 上线前预演 |
| `apply` | 执行落地（按确定性语义与冲突策略） | 可控执行 |
| `export` | 将线上知识网络状态导出为 BKN（可 round-trip） | 防漂移、可回滚、可复现 |

> 要求：所有导入操作必须记录审计信息（操作者、时间、输入文件指纹、变更集、结果）。

### 导入的确定性（必须保证）

为保证多人协作与可回放性，导入语义必须是**确定性的（deterministic）**：

- 对同一组输入文件（不考虑文件系统顺序）导入结果一致
- 同一文件重复导入结果一致（幂等）
- 冲突可解释：要么明确失败（fail-fast），要么有明确规则（例如 last-wins），不得“隐式合并”

### 唯一键与作用域

每个定义的唯一键建议为：

- `key = (network_id, type, id)`

其中 `network_id` 取自：

- 优先使用 frontmatter `network`
- 若缺失，则由导入目标网络（导入命令参数或 `type: network` 的 `id`）补齐

### 更新语义（replace vs merge）

默认建议使用 **replace（整段覆盖）**：

- 当 `key` 已存在时，用导入文件中的定义整体替换旧定义
- **缺失字段不代表删除**：仅代表“该字段不在本次定义中”；删除元素应通过 SDK/CLI 的显式删除 API 执行，而非 BKN 文件

如确有需要，可支持受控的 **merge-by-section（按章节合并）**，但必须满足：

- 仅允许合并少数“附加型章节”（例如 `属性覆盖`、`逻辑属性`）
- 冲突必须可控：同名逻辑属性/同名字段配置冲突时 fail-fast 或 last-wins（需配置）
- 合并策略必须在导入器中显式配置并记录到导入审计日志

### 冲突与优先级

当同一个 `key` 在一次导入批次中被多个文件重复声明：

- 默认：**fail-fast**（推荐，保证稳定性）
- 可选：按显式优先级排序（例如命令行顺序或 `priority` 字段），否则不建议支持

### 导入行为

| 场景 | 行为 |
|------|------|
| ID 不存在 | 创建新定义 |
| ID 已存在 | 更新定义（覆盖） |
| 删除元素 | 通过 SDK/CLI 的 delete API 显式执行，不通过 BKN 文件 |

### 导入示例

**场景：向已有网络添加新对象**

创建 `deployment.bkn`:

```markdown
---
type: object_type
id: deployment
name: Deployment
network: k8s-network
---

## Object: deployment

**Deployment**

Kubernetes deployment controller.

## Data Source

| Type | ID |
|------|-----|
| data_view | deployment_view |

## Data Properties

| Property | Primary Key | Display Key |
|----------|:-----------:|:-----------:|
| id | YES | |
| deployment_name | | YES |
```

导入后，`k8s-network` 将包含新的 `deployment` 对象。

**场景：更新已有对象**

创建同 ID 的文件，导入后自动覆盖：

```markdown
---
type: object_type
id: pod
name: Pod实例（更新版）
network: k8s-network
---

## Object: pod

**Pod实例（更新版）**

更新后的定义...
```

---

## 无 patch 更新模型

BKN 采用**无 patch 的更新模型**：定义文件仅用于新增与修改，删除通过 SDK/CLI API 显式执行。

### 定义文件导入（add/modify）

- 单个 `.bkn` 或 `.bknd` 文件导入时，按 `(network, type, id)` 执行 **upsert**（新增或覆盖）
- 修改：直接编辑对应定义文件，重新导入即可覆盖
- 缺失字段不代表删除：仅表示该字段不在本次定义中

### 删除元素

- 删除应通过 **SDK/CLI 的 delete API** 显式执行，不通过 BKN 文件
- 删除操作要求：显式参数、可审计、支持 dry-run 与批量删除

### 编辑流程

1. **新增**：创建 `.bkn` 文件，导入
2. **修改**：编辑 `.bkn` 文件，重新导入
3. **删除**：调用 SDK/CLI 的 delete 接口

---

## 最佳实践

### 命名规范

- **ID**: 小写字母、数字、下划线，如 `pod_belongs_node`
- **显示名称**: 简洁明确，如 "Pod属于节点"
- **标签**: 使用统一的标签体系

### 文档结构

1. 网络描述放在文件开头
2. 使用 mermaid 图展示整体拓扑
3. 对象定义在前，关系和行动在后
4. 相关定义放在一起

### 简洁原则

- 优先使用完全映射模式
- 只在需要时声明属性覆盖
- 避免重复信息

### 可读性

- 使用表格呈现结构化数据
- 添加业务语义说明
- 必要时使用 mermaid 图

---

## 参考

- [架构设计](./ARCHITECTURE.md)
- 样例：
  - [每定义一文件](../examples/k8s-modular/) - 每个定义独立文件
