# Catalog / Resource 自定义扩展（方案 B）技术设计文档

> **状态**：草案  
> **负责人**：@待补  
> **日期**：2026-05-08  
> **相关 Ticket**：[GitHub #382](https://github.com/kweaver-ai/kweaver-core/issues/382)  
> **总览选型**：[issue-382-catalog-resource-labels-design.md](./issue-382-catalog-resource-labels-design.md)（本文落地 **方案 B**：实体级扩展 **无主表 JSON 列**，仅关系副表 **`t_entity_extension`**；字段级为 **字段级树形概念**（见 **§3.4 / §6.11**：随 `schema_definition` JSON，展示不筛选）。**不引入**独立搜索中间件，库为 **MariaDB / DM8**）

---

## 目录

- [1. 背景与目标](#1-背景与目标)
- [2. 方案概览](#2-方案概览)
- [3. 数据模型](#3-数据模型)
- [4. 存储与一致性](#4-存储与一致性)
- [5. 查询模式](#5-查询模式)
- [6. 对外 API 设计（第一期）](#6-对外-api-设计第一期)
- [6.11 Resource：字段级 extensions（展示，字段级树形概念）](#611-resource字段级-extensions展示字段级树形概念)
- [7. 第二期：独立 extensions 子路径（规划）](#7-第二期独立-extensions-子路径规划)
- [8. 错误码与校验](#8-错误码与校验)
- [9. 迁移与兼容](#9-迁移与兼容)
- [10. 测试要点](#10-测试要点)
- [11. 里程碑](#11-里程碑)
- [12. 参考](#12-参考)

---

## 1. 背景与目标

### 1.1 背景

业务需在 **Catalog**、**Resource** 上挂载 DIP 等域外元数据（部门、信息系统等），形态为 **扁平 KV**，且列表需 **按 key/value 筛选**。  
另需在 **表 / 索引等资源的字段（Property）** 上挂载同类业务 KV，**仅用于详情与血缘等 UI 展示**，**不参与** Resource 列表的 SQL 筛选（**字段级树形概念**：以 Resource 为根、字段为子节点，扩展挂在「字段」叶子上；**不做**列表检索，与可检索实体级扩展解耦，避免列表 JOIN 爆炸）。

选定 **方案 B**：**不在** `t_catalog` / `t_resource` 上为「实体级」扩展增加大 JSON 列；实体级扩展**仅**存放在专用副表 **`t_entity_extension`** 中，由应用层组装为根对象上的 **`extensions`**。字段级扩展**不落** `t_entity_extension`，而是挂在 **`schema_definition` 各 Property 可选字段 `extensions`** 上，随 **`t_resource` 既有 schema JSON** 一并持久化。

### 1.2 目标

1. 使用 **一张副表**（**`t_entity_extension`**），以 **`(f_entity_id, f_key)`** 作为**复合主键**：`f_entity_id` 与 `t_catalog.f_id` / `t_resource.f_id` **共用同一套全局唯一 ID**（不在表内区分类型列）；**同一 `f_entity_id` 下每个 `f_key` 至多一行**（单值语义）。  
2. **不变量（须写进实现与评审）**：任意 Catalog、Resource 的 `f_id` 在**全系统不重复**；否则 `(f_entity_id, f_key)` 主键无法同时安全挂载两类实体。  
3. 列表筛选 **只 JOIN 该表**；详情/更新中的 **`extensions`** 由服务层 **读/写该表** 映射为 JSON 对象。  
4. **第一期**：**`extensions`** 的读写与筛选 **全部内嵌**在现有 Catalog / Resource 的 CRUD 与 List 接口中。  
5. **第二期**：再提供 **独立的 `.../extensions` 子资源接口**（见 §7），本期不实现、不阻塞第一期交付。  
6. **字段级**：在 **Resource** 的 **`schema_definition`** 上支持 **按 Property 可选 `extensions`**（展示用）；**不提供**按「某字段的某个 extension key/value」过滤 **`GET /resources` 列表** 的 query 参数，也不在列表 SQL 中解析 `schema_definition` JSON 做 WHERE。

### 1.3 非目标

- 不提供主表 JSON 快照列（区别于方案 A）。  
- 本期不提供 `GET/PUT /catalogs/{id}/extensions` 等独立子路径（第二期）。  
- 不引入 OpenSearch 等；不假设 JSON 路径索引能力（与总览文档一致）。  
- **字段级扩展**：不做列表筛选、不建独立副表索引路径；**不**要求 Catalog 对「列」级元数据对称能力（Catalog 无 `schema_definition` 时本期不涉及）。

---

## 2. 方案概览

### 2.1 为何单表且不区分 `f_scope`

- **前提**：`t_catalog.f_id` 与 `t_resource.f_id` **全局唯一、永不撞号**（同一发号空间或等价保证）。在此前提下，**主键 `(f_entity_id, f_key)`** 已唯一确定「挂在哪个实体上的哪一个扩展键」，**无需**再存 `catalog` / `resource` 类型列。  
- **好处**：列更少、主键更短；删除/替换扩展 KV 时条件仅为 `f_entity_id`；仓储 **一套** 读写逻辑，按 HTTP 上下文（catalog 或 resource 路由）写入正确的 `f_entity_id` 即可。  
- **风险与约束**：若未来 ID 规则变化导致 **catalog 与 resource 可能同号**，当前模型 **不适用**，需回退为带 `f_scope` 或拆表。请在架构评审中 **书面冻结**「全局唯一 ID」不变量。

### 2.2 数据流（第一期）

```mermaid
flowchart LR
  subgraph API[HTTP]
    L[List / Filter]
    D[Detail GET]
    W[Create / Update]
  end
  subgraph DB[(MariaDB / DM8)]
    TC[t_catalog]
    TR[t_resource]
    EX[t_entity_extension]
  end
  L -->|JOIN EX| EX
  L --> TR
  D --> TR
  D --> EX
  W -->|事务内写 EX| EX
  TC -.->|f_id| EX
  TR -.->|f_id| EX
```

---

## 3. 数据模型

### 3.1 表：`t_entity_extension`（单表，无 `f_scope`）

> **命名说明**：关系副表 **`t_entity_extension`** 存可检索扩展 KV；**HTTP JSON 仅使用 `extensions`** 字段名（与 Issue「扩展」一致）。

| 列名 | 类型（MariaDB 示例） | 约束 | 说明 |
|------|----------------------|------|------|
| `f_entity_id` | `VARCHAR(40) NOT NULL` | **主键 1** | 等于 `t_catalog.f_id` **或** `t_resource.f_id`（**全局唯一**） |
| `f_key` | `VARCHAR(128) NOT NULL` | **主键 2** | 建议 `dip:` 等业务前缀 |
| `f_value` | `VARCHAR(512) NOT NULL` | | 字符串；结构化可约定为 JSON 文本 |
| `f_create_time` | `BIGINT NOT NULL DEFAULT 0` | 可选 | |
| `f_update_time` | `BIGINT NOT NULL DEFAULT 0` | 可选 | |

**主键**：`PRIMARY KEY (f_entity_id, f_key)`  

**索引**（按查询习惯选配，可评审后定稿）：

- `KEY idx_entity (f_entity_id)`：按实体拉全量 KV（详情、整包替换前删除）。  
- `KEY idx_entity_key_value (f_entity_id, f_key, f_value(191))`：列表 JOIN 后按 key/value 筛选。  
- Resource 在 **某 catalog 下** 且带标签筛选条件时：**`JOIN t_resource r`**，再 **`JOIN t_entity_extension e ON e.f_entity_id = r.f_id`**（见 §5）。

**外键**：`f_entity_id` 无法同时 FK 到 `t_catalog` 与 `t_resource` 两表；**本期推荐不设 FK**，由 **应用层** 在删除 catalog / resource 时删除对应副表行。

### 3.2 DM8

与 `migrations/dm8` 现有脚本风格对齐；**复合主键 `(f_entity_id, f_key)`** 语义不变。

### 3.3 单值语义

本期 **`(f_entity_id, f_key)` 仅一行**；更新为 `INSERT ... ON DUPLICATE KEY UPDATE` / 先按 `f_entity_id` 删再插整包均可。

**多值**（同一 key 多个 value）若未来需要：需调整主键（如增加 `f_ordinal` 或改用代理主键 `f_id`）；**本期不展开**。

### 3.4 字段级扩展（Resource / `schema_definition`，展示用）

> **定位**：与 §3.1 **实体级** `t_entity_extension` **并列的产品能力**，但**存储与检索路径不同**——采用 **字段级树形概念**（实体为根、字段为子，扩展仅挂在字段节点）：**只做展示、不做列表筛选**。

| 维度 | 实体级 `extensions` | 字段级 `Property.extensions` |
|------|---------------------|------------------------------|
| **挂载点** | Catalog / Resource **根对象** | **`schema_definition[]` 每一项**（`Property`） |
| **持久化** | 表 **`t_entity_extension`** | **`t_resource` 上已有列**（如 `schema_definition` 的 JSON 文本；列名以现网为准） |
| **列表筛选** | 支持 `extension_key` / `extension_value` + JOIN 副表（§5） | **不支持**；**禁止**在 `GET /resources` 列表 SQL 中依赖 JSON 路径解析做 WHERE |
| **详情 / 写** | 读写副表行 ↔ 根 `extensions` | 读写 **同一条** `t_resource` 记录中的 schema JSON；由服务层序列化/反序列化 `Property.extensions` |

**JSON 形状（示意）**：

```json
{
  "name": "order_id",
  "type": "string",
  "display_name": "订单号",
  "extensions": {
    "dip:security_level": "L3",
    "dip:owner_team": "trade"
  }
}
```

**约定**：

1. **`extensions` 键名**与实体级一致：扁平 **`string` → `string`**；`key` / `value` 长度与条数上限建议**复用**实体级配额策略（可略收紧，如单字段 ≤16 条，便于评审）。  
2. **`attributes` 与 `extensions` 分工不变**：`attributes` 为 **connector opaque**；`extensions` 为 **业务域外 KV**（与 OpenAPI `Property` 对齐时见 `resource.yaml`）。  
3. **适用范围（第一期）**：以 **`category` 为 table / index 等且携带 `schema_definition`** 的 Resource 为主；`logicview` 的 `logic_definition` **本期不强制**对称扩展（若节点输出字段也要同类能力，另立变更）。  
4. **发现（Discover）与全量 PUT 合并**：连接器回写 `schema_definition` 时，应 **按 `name`（或约定的稳定字段键）合并保留** 客户端已写入的 **`extensions`**，避免发现任务把展示用标签整体冲掉（实现可约定「仅当源端无对等业务列时保留旧值」或「显式字段级版本号」；具体以评审结论写死一种）。

---

## 4. 存储与一致性

### 4.1 写入顺序（更新 Resource 且携带 `extensions`）

在同一数据库事务内：

1. 校验请求体中的 **`extensions`** 对象（key/value 类型、长度、条数上限）。  
2. 对解析得到的 **整包 object** 采用 **替换语义**：删除 `t_entity_extension` 中 `f_entity_id = ?`（当前 resource 的 `f_id`）的全部行，再批量插入新行（或 diff 后 UPSERT，效果等价于整包替换）。  
3. 提交事务。

**部分更新**：若产品需要「只改一个 key、不动其它」，第一期可约定 **仍传完整对象**（客户端先 GET 再合并后 PUT）；或在第一期支持 `PATCH` 增量语义（单独开小节实现时再定）。本文默认 **PUT 整包替换** 以降低歧义。**触发条件**：根对象出现 **`extensions` 键**（含 `{}`）即视为更新 KV。

### 4.2 Catalog 的 extensions

对 `t_entity_extension` 使用相同策略：`f_entity_id` 为当前 catalog 的 `f_id`（与 resource 的 id 不重复，故不会误删）。

### 4.3 删除

- 删除 **Resource**：`DELETE FROM t_entity_extension WHERE f_entity_id = ?`（与删 `t_resource` 同事务）。  
- 删除 **Catalog**：`DELETE FROM t_entity_extension WHERE f_entity_id = ?`（该 catalog 自身 extensions）；若业务 **级联删除** 其下所有 resource，需额外 `DELETE FROM t_entity_extension WHERE f_entity_id IN (...)`（待删 resource 的 id 列表）。

### 4.4 字段级 `extensions` 写入（Resource）

1. 客户端在 **`PUT /resources/{id}`**（或创建 **POST /resources**）的 **`schema_definition`** 中，对部分 `Property` 携带 **`extensions`**。  
2. 服务端校验每个 `Property.extensions` 为 object、扁平 string→string，满足配额与保留前缀规则（与 §8 实体级一致或子集）。  
3. **同一事务**内更新 `t_resource` 中持久化 schema 的列即可；**不**对 `t_entity_extension` 按「字段」增删行（避免 `(f_entity_id, f_key)` 与实体级 key 命名冲突及语义混淆）。  
4. 若请求体 **未携带** `schema_definition` 键：按现有资源部分更新语义处理——**不**单独解析「只改字段 extensions」的 PATCH 特例（除非产品另开 PATCH）；默认 **整段 schema 替换** 时须整体校验。

---

## 5. 查询模式

### 5.1 Resource 列表：按单个扩展 KV 条件筛选

```sql
SELECT r.*
FROM t_resource r
INNER JOIN t_entity_extension e
  ON e.f_entity_id = r.f_id
WHERE r.f_catalog_id = ?
  AND e.f_key = ?
  AND e.f_value = ?;
```

### 5.2 Resource 列表：多个扩展 KV 条件 AND

```sql
SELECT r.*
FROM t_resource r
WHERE r.f_catalog_id = ?
  AND EXISTS (
    SELECT 1 FROM t_entity_extension e1
    WHERE e1.f_entity_id = r.f_id
      AND e1.f_key = ? AND e1.f_value = ?
  )
  AND EXISTS (
    SELECT 1 FROM t_entity_extension e2
    WHERE e2.f_entity_id = r.f_id
      AND e2.f_key = ? AND e2.f_value = ?
  );
```

限制单次请求中 **扩展筛选条件对数上限**（建议可配置，如 ≤5）。

### 5.3 Catalog 列表：按 extensions 筛选

```sql
SELECT c.*
FROM t_catalog c
INNER JOIN t_entity_extension e
  ON e.f_entity_id = c.f_id
WHERE e.f_key = ? AND e.f_value = ?;
```

多条件 AND 同理使用多个 `EXISTS`。

### 5.4 详情：组装 `extensions`

**Catalog 与 Resource 相同形态**（由路由/上下文已知实体类型，仅 `f_entity_id` 不同）：

```sql
SELECT f_key, f_value
FROM t_entity_extension
WHERE f_entity_id = ?
ORDER BY f_key;
```

服务层转为 JSON 对象，填入响应体 **`extensions`** 字段。

### 5.5 字段级扩展与列表（明确排除）

- **`GET /resources` 列表**：**不得**增加「按某 `schema_definition[i].extensions` 的 key/value 过滤」的 query 参数；实现上 **不得** 对 `schema_definition` 做大字段 JSON 解析筛选。  
- **性能**：若列表响应包含精简版 `schema_definition`，字段级 `extensions` 是否随列表返回由 **`schema_definition` 是否在列表中返回** 决定；若列表不返回 schema，则 UI 在 **详情 `GET /resources/{id}`** 或批量详情中取字段级展示数据。

---

## 6. 对外 API 设计（第一期）

以下路径前缀与现有文档一致：**`/api/vega-backend/v1`**（参见 [catalogs-api.md](./catalogs-api.md)）。OpenAPI 合入时以仓库 YAML 为准；本文给出**契约级**约定。

### 6.1 通用约定

| 项 | 约定 |
|----|------|
| **体字段名** | 仅 **`extensions`**：扁平 object（`string`→`string`）。**请求**：根对象出现 **`extensions` 键**（含 `{}`）即对该实体 KV **整包替换**；未出现则不修改副表。**响应**：仅 **`extensions`**。 |
| 列表 query | **`extension_key` / `extension_value`** 成对出现（多对 **AND**）；数组 query 形式见 OpenAPI。 |
| 字符与配额 | `key` / `value` 长度上限与表列一致；单实体 **最大 KV 条数**（如 64，可配置）；保留 **`vega_` 前缀** 给平台内置 key，业务侧建议使用 **`dip:`** 等前缀。 |
| 空 object | 副表无行即等价 `{}`；`PUT` 传 `extensions: {}` 表示删除该实体全部扩展行。 |
| 认证与分页 | 与现有 Catalog / Resource 列表一致（如 `x-account-id`、`offset`/`limit`）。 |
| **字段级（Resource）** | 见 **§6.11**：`schema_definition[].extensions` 与根级 **同名**、**同形态**，但**不落** `t_entity_extension`、**不参与**列表筛选。 |

### 6.2 Catalog：创建

**`POST /catalogs`**

在现有请求体上 **增加可选字段**：

```json
{
  "name": "my-catalog",
  "type": "physical",
  "extensions": {
    "dip:department_id": "D001",
    "dip:information_system_id": "SYS-01"
  }
}
```

**行为**：创建 `t_catalog` 成功后，在同一事务内写入 `t_entity_extension` 多行（`f_entity_id` 为新 catalog 的 `f_id`，若有 `extensions`）。

**响应**：与现有一致；在响应体 **Catalog** 对象上 **增加 `extensions`** 字段（从副表组装，无行则为 `{}`）。

### 6.3 Catalog：更新

**`PUT /catalogs/{id}`**（或项目实际使用的更新动词）

请求体可选：

```json
{
  "name": "…",
  "extensions": {
    "dip:department_id": "D002"
  }
}
```

**语义**：

- **若请求体包含 `extensions` 键**（含空对象 `{}`）：对该 catalog 执行 **整包替换**（见 §4.1）。  
- **若未包含 `extensions` 键**：不修改副表（与其它字段部分更新策略一致）。

**响应**：返回更新后的 Catalog，带 **`extensions`**。

### 6.4 Catalog：详情

**`GET /catalogs/{id}`**

响应体在现有字段基础上增加：

```json
{
  "id": "catalog-id",
  "name": "…",
  "extensions": {
    "dip:department_id": "D001"
  }
}
```

数据来源：**仅** `t_entity_extension` 按 `f_entity_id = catalog id` 聚合，**不读**主表 JSON 列。

### 6.5 Catalog：列表与筛选

**`GET /catalogs`**

在现有 query 参数（如 `type`、`health_check_status`、`offset`、`limit`）基础上 **增加**：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `extension_key` | string | 否 | 与 `extension_value` **成对**；等值匹配 `f_key` |
| `extension_value` | string | 否 | 与 `extension_key` 成对；等值匹配 `f_value` |

**多对 AND**：重复多组 `extension_key` / `extension_value`（数组 query），服务端解析为条件列表，生成 §5.2 类 `EXISTS` SQL。

**示例**：

```http
GET /api/vega-backend/v1/catalogs?type=physical&extension_key=dip:department_id&extension_value=D001&extension_key=dip:information_system_id&extension_value=SYS-01
```

**列表是否返回 `extensions`**：

- **默认 `include_extensions=false`**：列表项 **不** 带 `extensions`，减轻负载与 N+1。  
- **`include_extensions=true`**：每条 Catalog **批量**查 `t_entity_extension` 组装后填充 `extensions`。  
- **`include_extension_keys`**（可选）：仅返回列出的 key（逗号分隔，与 OpenAPI 对齐）。

### 6.6 Resource：创建

**`POST /resources`**（路径以现网为准）

请求体可选：

```json
{
  "catalog_id": "catalog-id",
  "name": "order_table",
  "category": "table",
  "extensions": {
    "dip:owner": "alice",
    "env": "prod"
  }
}
```

事务内：插入 `t_resource` 后，写 `t_entity_extension`（`f_entity_id` 为新资源 `f_id`）。

### 6.7 Resource：更新

**`PUT /resources/{id}`**

与 Catalog 相同：**根对象出现 `extensions` 键**则整包替换；未出现则不删不改副表。

### 6.8 Resource：详情

**`GET /resources/{id}`**

响应增加 **`extensions`**：来自 `t_entity_extension`（`f_entity_id = resource id`）。

### 6.9 Resource：列表与筛选

**`GET /resources`**（或 `GET /catalogs/{id}/resources`，以现网为准）

在现有过滤参数基础上增加与 **§6.5 相同**的 `extension_key` / `extension_value` 成对参数及多对 AND 规则；SQL 见 §5.1 / §5.2。

**`include_extensions` / `include_extension_keys`**：语义同 Catalog 列表。

### 6.10 前缀匹配（可选，第一期可不做）

若需要 **前缀筛选**，可增加：

| 参数 | 说明 |
|------|------|
| `extension_key` | 同上 |
| `extension_value_prefix` | 与 `extension_key` 成对；SQL 使用 `e.f_value LIKE CONCAT(?, '%')` 且**必须**参数绑定；需评估索引（前缀索引或仅限低基场景）。 |

默认第一期仅支持 **等值**，降低产品与性能歧义。

### 6.11 Resource：字段级 `extensions`（展示，字段级树形概念）

**仅 Resource**，且落在 **`schema_definition`** 的 OpenAPI **`Property`** 上（与 [resource.yaml](../../../../../api/vega/vega-backend-api/resource.yaml) 合入时对齐）。

| 项 | 约定 |
|----|------|
| **字段名** | 每个 `Property` 上可选 **`extensions`**：`string`→`string`，与根级实体 `extensions` 形态一致。 |
| **读写** | 随 **`schema_definition`** 在创建 / 更新 Resource 时写入；在详情与（若响应含 schema 的）列表中读出。 |
| **筛选** | **不提供**列表 query；不参与 §5.1–§5.3 SQL。 |
| **存储** | **不**写入 `t_entity_extension`；见 §3.4、§4.4。 |

**请求 / 响应示例（片段）**

```json
{
  "catalog_id": "cat1",
  "name": "orders",
  "category": "table",
  "schema_definition": [
    {
      "name": "order_id",
      "type": "string",
      "extensions": { "dip:data_class": "internal" }
    },
    {
      "name": "amount",
      "type": "decimal",
      "extensions": {}
    }
  ]
}
```

**OpenAPI**：在 `Property` 的 `properties` 中增加可选 **`extensions`**，其 schema 可与根级 `EntityExtensions` 复用（`maxProperties` 等可单独收紧）。**Catalog** 无字段清单，**不适用**本节。

---

## 7. 第二期：独立 extensions 子路径（规划）

本期 **不实现**以下内容，仅作为后续迭代占位，避免第一期范围膨胀。

| 能力 | 建议路径（草案） | 说明 |
|------|------------------|------|
| 只读 | `GET /catalogs/{id}/extensions` | 与详情中 `extensions` 结构一致 |
| 只读 | `GET /resources/{id}/extensions` | 同上 |
| 整包替换 | `PUT /catalogs/{id}/extensions` | Body 为完整 object；与第一期 body 内嵌等价 |
| 整包替换 | `PUT /resources/{id}/extensions` | 同上 |

> 第一期已在主 CRUD body 中支持 **`extensions`** 时，第二期独立子路径为 **可选**；用于减少大 body 或权限细分场景。

第二期实现时：**复用**第一期仓储与校验逻辑，仅增加路由与权限策略；**数据仍落在** `t_entity_extension`。

---

## 8. 错误码与校验

| 场景 | HTTP | 说明（错误码命名以合入时枚举为准） |
|------|------|--------------------------------------|
| `extensions` 非 object | 400 | 非法 JSON 形状 |
| key 非 string / value 非 string | 400 | 第一期仅支持扁平 string→string |
| 超出条数或长度上限 | 400 | `QuotaExceeded` 类 |
| key 使用保留前缀 `vega_` | 400 | `ReservedKey` |
| `extension_key` 无对应 `extension_value`（或反之） | 400 | 查询参数成对校验 |
| 扩展筛选条件对数超过上限 | 400 | 防止恶意深嵌 `EXISTS` |
| `schema_definition[].extensions` 非法形状或超配额 | 400 | 与实体级规则对齐或为其子集（如 `VegaBackend.Extensions.*`） |

---

## 9. 迁移与兼容

- **新环境**：在 `migrations/mariadb`、`migrations/dm8` 中增加 **`t_entity_extension`** 及索引；**不改** `t_catalog` / `t_resource` 主表结构。  
- **老客户端**：可不传 `extensions`；响应新增字段须 **向后兼容**（客户端应忽略未知字段）。  
- **从方案 A 回退或迁移**：若曾存在主表 `f_extensions` 列，需一次性脚本 **解析 JSON → 插入 `t_entity_extension`**（`f_entity_id` 为主表 id）后，再按需删列（另立变更单）。

---

## 10. 测试要点

- 创建/更新 catalog、resource **带 / 不带** `extensions`；`{}` 清空。  
- 列表：单条件、多条件 AND、与原有 catalog_id/type 等组合；`include_extensions` 开/关。  
- 删除 resource / catalog 后副表行 **不存在**（FK 或应用级级联）。  
- 并发：两请求同时 PUT 同一 id 的 `extensions`（最后写入胜或乐观锁策略按产品定）。  
- **字段级**：`schema_definition` 带 / 不带 `Property.extensions`；Discover 回写后 **关键业务 tag 不丢**（见 §3.4）。  
- **字段级负例**：确认列表 **无** 按字段 extensions 筛选的 query，且 SQL 计划 **无** schema JSON 路径 WHERE。  
- DM8 + MariaDB 双方言迁移与核心路径。

---

## 11. 里程碑

| 阶段 | 内容 |
|------|------|
| **第一期** | **`t_entity_extension` 单表** DDL；仓储；Catalog/Resource 创建、更新、详情、列表的 **根级 `extensions`** 与 **`extension_*` 筛选**；**Resource `schema_definition[].extensions`（展示）** 与 OpenAPI `Property` 对齐；联调 |
| **第二期** | 独立 `GET/PUT …/extensions`；权限与缓存策略；不与第一期表结构冲突 |

---

## 12. 参考

- [issue-382-catalog-resource-labels-design.md](./issue-382-catalog-resource-labels-design.md)  
- [catalogs-api.md](./catalogs-api.md)  
- [catalog-resource-labels-scheme-a-design.md](./catalog-resource-labels-scheme-a-design.md)（方案 A，对照用）  
- 代码：`adp/vega/vega-backend`
