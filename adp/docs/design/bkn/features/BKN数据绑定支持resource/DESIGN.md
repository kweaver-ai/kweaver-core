# BKN 数据绑定支持 Vega Resource 技术设计文档

> **状态**：草案  
> **版本**：0.2.0  
> **日期**：2026-04-01  
> **相关 Ticket**：#336

---

## 1. 背景与目标 (Context & Goals)

BKN 中 **对象类（ObjectType）** 与 **关系类（RelationType，间接关联）** 当前通过绑定 **数据视图（data view）** 完成字段映射；**索引构建任务**从视图拉取数据后写入 OpenSearch；**ontology-query** 通过 UniQuery 查询视图或经既有路径访问已索引数据。

业务数据在 **vega-backend** 中已通过 **Resource** 统一建模（表、逻辑视图、dataset 等多类 `category`）。为减少「必须先落视图再绑 BKN」的链路，需要在 BKN 侧支持将数据源直接绑定为 vega **Resource**，并在在线查询中与现有 **data_view** 能力在「读源」语义上对齐。

**目标**：

- 对象类、关系类（间接关联）的 `data_source` / `backing_data_source` 在 **`data_view` 之外** 支持类型 **`resource`**，仅依赖 **resource id**（不单独持久化 `catalog_id`）。
- **bkn-backend**：校验、详情补全在数据源为 `resource` 时，通过 vega-backend **查询资源元数据 / 资源数据**接口完成；**不对 resource 源对象类生成索引构建任务**（见下文「关键决策」与 §3.3）。
- **ontology-query**：新建 **vega-backend 客户端**；请求/响应类型与 **bkn-backend** 保持一致；当对象类数据源为 **`resource`** 时，**仅**通过 vega **Query Resource Data** 查询，**不提供**基于 OpenSearch 的索引查询路径（见 §3.4、§3.5）。

**非目标**：

- **不**在本文档范围内展开 vega-backend **ResourceData** 查询的内部实现（连接器、下推、分页细节等），以 vega 已暴露的 HTTP 契约为准。
- **不**替换既有 `data_view` 行为。
- **不**为 `data_source.type == resource` 的对象类提供索引构建或与索引查询等价的能力。

---

## 2. 方案概览 (High-Level Design)

### 2.1 核心思路

在 BKN 元数据与执行层采用统一 **数据源分流** 模型：

- **`ResourceInfo`**：`type` 为 `data_view` 或 **`resource`**，`id` 为视图 id 或 vega resource id；**仅需存储 `id`**，**不增加** `catalog_id` 字段（vega 侧仅凭 resource id 可解析资源）。
- **`type == resource`**：**建模与校验** 通过 `GetResourceByID` 拉取 schema，与视图分支对称；**在线查询批量取数** 统一走 vega **Query Resource Data** HTTP API；**不参与** OpenSearch 索引构建与索引查询。
- **ontology-query** 独立实现 **vega 访问适配器**，DTO 与 bkn-backend **逐字段对齐**，避免双端 JSON 契约漂移。

**已定稿方案**：采用 **BKN 统一分流** `data_view` | `resource`——`data_view` 延续既有 **UniQuery / 索引任务 / OpenSearch 查询** 全链路；`resource` 仅走 **vega 元数据 + Query Resource Data**，**不建索引、不查索引**，职责与 vega Resource 产品模型一致；ontology-query 需新增 vega 适配与配置，与 bkn-backend **同构 DTO**（可后续抽公共模块）。

### 2.2 总体架构

```text
                    ┌─────────────────────────────────────────┐
                    │            vega-backend                  │
                    │  Resource 元数据 / Query Resource Data   │
                    └─────────────────────────────────────────┘
                           ▲                    ▲
                           │                    │
            Resource 路径   │                    │ Resource 路径（仅直连 vega）
                           │                    │
        ┌──────────────────┴──────┐    ┌───────┴──────────────────┐
        │       bkn-backend         │    │     ontology-query       │
        │  校验 / 详情 / 审计        │    │  条件查询 / 子图间接关联   │
        │  索引 Worker（仅 data_view）│    │  resource：仅 QueryResourceData │
        └──────────────┬────────────┘    │  data_view：既有 OS / 视图路径 │
                       │                  └────────────┬─────────────┘
                       │                                 │
                       └─────────────┬───────────────────┘
                                     │
                            OpenSearch（仅 data_view 索引命中场景）
```

**data_view** 路径保持现有：**UniQuery / DataViewAccess**、索引构建与「有索引则查索引」的在线查询逻辑不变。

### 2.3 关键决策

| 序号 | 决策 | 说明 |
|------|------|------|
| 1 | `data_source.type` 新增枚举 **`resource`** | 与 `data_view` 并列，由 BKN 全链路识别。 |
| 2 | **仅依赖 resource `id`** | `ResourceInfo` 不增加 `catalog_id`；前提为 vega 侧 resource id 全局可定位且权限闭合。 |
| 3 | **ontology-query 新建 vega 客户端** | 不引用 bkn-backend 二进制依赖；**请求/响应结构**与 bkn-backend `VegaBackendAccess` 侧定义 **保持一致**（可复制 struct 或后续抽公共模块）。 |
| 4 | **resource 不建索引、不查索引** | **索引任务调度**：创建/生成索引构建任务时，若对象类 `data_source.type == resource`，**不生成**对应 task。**ontology-query**：对象类数据源为 **resource** 时，**无** OpenSearch 索引查询选项，**仅** **Query Resource Data**。`data_view` 仍保持既有索引构建与「有索引优先 OS」策略。 |
| 5 | **审计与现网对齐** | 对象类/关系类绑定、变更的审计模型与现有 **data_view** 绑定 **同一套** 字段粒度与对象类型，扩展记录 **resource id** 与 **type**。 |
| 6 | **错误与退化语义** | vega **不可用** → **报错**；schema 变更导致 **映射字段缺失** → **报错**；映射字段在响应中 **均存在** → **允许继续**（可忽略 vega 新增列）；**分页/上限** 以 **vega 返回错误**为准，BKN **透传或包装**为业务错误。 |

---

## 3. 详细设计 (Detailed Design)

### 3.1 数据模型与枚举

#### 3.1.1 ResourceInfo（延续现有结构）

与现网 `ResourceInfo` 一致，**不新增** `catalog_id`：

| 字段 | 说明 |
|------|------|
| `type` | `data_view` \| **`resource`** |
| `id` | 视图 id 或 vega resource id |
| `name` | 可选；运行时可通过 `GetResourceByID` / 视图接口回填展示名 |

#### 3.1.2 对象类 ObjectType

- **`data_source`**：`type` 为 `resource` 时，`id` 必须为有效 vega resource id。
- **属性映射**：`mapped_field` 与 vega 返回 **entries** 行内字段名对齐；校验逻辑与 data_view 分支对称（字段存在性、类型与操作符等以产品/现网规则为准）。
- **索引**：**不支持**为 resource 源对象类构建 OpenSearch 索引；索引相关任务与 UI/接口选项对 resource 源 **不适用**。

#### 3.1.3 关系类 RelationType（间接关联）

- **`InDirectMapping.backing_data_source`**：`type` 可为 **`resource`**，`id` 为 backing 资源 id。
- **行为目标**：与 **backing 为 data_view** 时一致——批量拉取中间态数据、按映射规则关联两端对象；实现上共享「行数据为 `[]map[string]any`」之后的逻辑，**仅数据源适配层**分支。

### 3.2 vega-backend 接口依赖（契约摘要）

本特性 **依赖** vega-backend 已提供的 **按资源 id 查询数据**能力（实现细节见 vega 代码，BKN 只对接 HTTP）。

- **路径（示例）**：`POST /api/vega-backend/v1/resources/:id/data`（内网部署可选用 `in` 前缀路径，以环境配置为准）。
- **BKN 客户端必须遵循**：对上述路径发起 **`POST`**，**`Content-Type: application/json`**，请求体为查询参数 JSON；同时必须携带请求头 **`x-http-method-override: GET`**（vega 常量 `interfaces.HTTP_HEADER_METHOD_OVERRIDE`）。handler 在校验该头不等于 `GET` 时返回 **400**，未带或与 vega 约定不一致将导致查询失败。
- **请求体**：与 vega `ResourceDataQueryParams` 对齐（如 `limit`、`offset`/`search_after`、`sort`、`filter_condition`、`output_fields`、`need_total` 等），**BKN 适配层 DTO 字段名与 JSON tag 须与 vega 一致**。
- **响应体**：包含 `entries`（行列表）、可选 `total_count`（当 `need_total` 为真时）等，**以 vega 实际响应为准**。
- **鉴权**：与现网 bkn-backend 访问 vega 方式一致（如 Account header）；ontology-query 使用 **相同 header 约定**，保证权限模型一致。

### 3.3 bkn-backend 改造要点

| 模块 | 内容 |
|------|------|
| **interfaces** | 在 `VegaBackendAccess`（或与 Dataset 查询对称的位置）增加 **QueryResourceData(ctx, resourceID, params)**（及请求/响应类型）；与 ontology-query **同构**。 |
| **drivenadapters/vega_backend** | 实现 HTTP 调用、错误映射、超时；与现有 `GetResourceByID` 共用 baseUrl/header 构建方式。 |
| **object_type 逻辑** | `processObjectTypeDetails` 等：`type == resource` 时 **`GetResourceByID`** + `SchemaDefinition` 补全名称、字段显示名与类型，支撑条件操作符等与 view 对称能力。 |
| **relation_type 逻辑** | 间接映射：`backing_data_source.type == resource` 时 **资源存在性与映射字段校验**（不再 `GetDataViewByID`）。 |
| **worker / 索引任务** | **仅**对 **`data_source.type == data_view`** 的对象类生成并执行索引构建任务。当对象类数据源为 **`resource`** 时，在**创建索引构建任务**的入口（调度/入队逻辑）**跳过，不生成 task**；不得对 resource 源走「分页拉批写入 OpenSearch」路径。现有 **`data_view`** 索引管道与行为 **不变**。 |

### 3.4 ontology-query 改造要点

| 模块 | 内容 |
|------|------|
| **配置** | 增加 **vega-backend BaseURL**（及必要超时），与 bkn-backend 配置项命名风格一致。 |
| **drivenadapters** | **新建** `vega_backend` 包，实现与 bkn-backend **相同签名与 DTO** 的 QueryResourceData（及按需的 GetResourceByID，若详情接口需要）。 |
| **object_type 查询** | 当 `data_source.type == resource`：**始终**走 **QueryResourceData**；**不提供**「走 OpenSearch 索引」的查询分支或选项（与 `data_view` 已建索引可走 OS 的路径区分）。条件重写、分页参数与现有 `ViewQuery` **尽量对齐**，降低两套语义成本。 |
| **knowledge_network / 间接关联** | 拉取 `backing_data_source` 为 resource 的批量数据时走 **QueryResourceData**，后续 **复用** 基于 `viewData` 的匹配与关联逻辑。 |
| **data_view 路径** | 若对象类为 `data_view` 且已具备可用索引，**保持**既有「优先 OpenSearch」等行为；**不**将上述逻辑套用到 **resource** 源。 |

### 3.5 与 data_view 的行为对齐表

| 能力 | data_view | resource |
|------|-----------|----------|
| 绑定对象类/间接关系类 | 支持 | **支持** |
| 详情展示与字段元数据 | 视图字段 | **GetResourceByID schema** |
| 索引构建取数 | UniQuery / DataView | **不支持**（不生成索引任务、不写入 OpenSearch） |
| 在线查询（无索引 / 读源） | UniQuery 视图 | **vega QueryResourceData** |
| 在线查询（有索引 / 搜索索引） | OpenSearch（既有） | **不适用**（无索引查询选项，仅 vega） |
| 必经 BKN dataset | 视现有实现 | **否** |

### 3.6 权限、审计、血缘

- **权限**：延续 BKN 与 vega 之间现有 **账号 / 租户** 透传方式；resource 路径不得绕过 vega 侧鉴权。
- **审计**：对象类/关系类 **创建/更新/删除** 绑定关系时，审计对象中体现 **`data_source.type` 与 `id`**；与 data_view **同一审计模型**扩展字段即可。
- **血缘**：若现网对视图绑定有血缘或依赖展示，resource 绑定应 **同等记录**（resource id、名称可选）。

---

## 4. 风险与边界 (Risks & Edge Cases)

| 风险/场景 | 说明与处理 |
|-----------|------------|
| **契约漂移** | bkn-backend 与 ontology-query 各自维护 vega DTO 时易不一致；**首期强制字段对齐 + 单测/契约注释**；后续可抽共享模块。 |
| **查询语义不一致** | 视图过滤表达式与 vega `filter_condition` 若不完全等价，需在 **条件重写**层明确支持矩阵与降级（不支持则返回明确错误）。 |
| **大结果集与限流** | 依赖 vega 分页与错误；若存在需批量扫数的场景（如报表），在 **产品/实现** 上限定分页与超时，与 vega 约定一致。 |
| **混合绑定** | 一端 data_view、一端 resource 的间接关系若允许，需 **集成测试**覆盖；若不允许，应在校验层拒绝。 |
| **性能与新鲜度** | resource 对象类 **仅**走 vega 实时（或 vega 语义下的）查询，**无**本地搜索索引加速；需在 **产品说明**中说明与 `data_view`+索引路径的差异。 |

### 4.1 建议覆盖的测试用例（摘要）

- `data_source.type == resource` 时：**详情补全**、**非法 id**、**映射字段在 schema 中缺失**。  
- **索引任务**：`resource` 源对象类 **创建索引任务** → **不产生 task**（或等价「无任务入队」断言）；`data_view` 行为回归。  
- ontology-query：**resource** 对象类 **仅** QueryResourceData，**无** OS 分支；**data_view** **有索引** 仍可走 OS；**间接关联 + resource backing**。  
- schema 变更：**映射列被删** → 错误；**仅新增列** → 查询仍成功。

---

## 5. 任务拆分 (Milestones)

- [ ] **契约对齐**：在 bkn-backend 定义 `QueryResourceData` 与 vega OpenAPI/实现 **字段级对照表**（可附于 PR 说明）。  
- [ ] **bkn-backend**：`VegaBackendAccess` 扩展与实现；object_type / relation_type 分支；索引任务调度 **排除** `data_source.type == resource`（**不生成** task）；单测与 mock。  
- [ ] **ontology-query**：配置、新建 vega 适配层（DTO 与 bkn-backend 一致）；object_type、子图/间接关联 **resource** 分支；确认 **无** resource 的索引查询路径；单测。  
- [ ] **审计**：绑定变更事件 **type + id** 记录与现网一致。  
- [ ] **文档**：对外 API/OpenAPI 若包含 `data_source.type` 枚举，**更新枚举说明**；运维手册补充 vega 不可用时的 **报错与重试**说明。

---

## 6. 参考

- 设计文档规范与结构参考：[bkn_docs/DESIGN.md](../bkn_docs/DESIGN.md)  
- vega Resource 实体定义（概念参考）：`vega/vega-backend/server/interfaces/resource.go`  
- vega 资源数据查询 Handler（路径与入参参考）：`vega/vega-backend/server/driveradapters/resource_data_handler.go`  
- vega 查询参数定义（字段参考）：`vega/vega-backend/server/interfaces/resource_data.go`  
- BKN 现有数据视图与索引任务（改造切入点）：`bkn/bkn-backend/server/worker/object_type_task.go`、`bkn/ontology-query/server/logics/object_type/object_type_service.go`
