# Resource Data 查询：过滤 + 聚合 / 分组 / 排序 / Having（请求体 Schema 设计）

> **状态**：草案（供 OpenAPI / 实现评审）  
> **关联接口**：`POST /api/vega-backend/v1/resources/:id/data`（请求体 `ResourceDataQueryParams` 扩展）  
> **对齐**：与 BKN 原生指标设计中的 `calculation_formula`（`aggregation` / `group_by` / `order_by` / `having`）语义一致，便于 ontology-query 将 **统一中间表示** 映射到 vega。

---

## 1. 设计目标

在 **已支持** `filter_condition`（行级过滤，等价 SQL `WHERE` 作用于明细行）的基础上，增加：

| 能力 | 语义（概念 SQL） | 请求体字段 |
|------|------------------|------------|
| 聚合 | `SELECT aggr(property) …` | `aggregation` |
| 分组 | `GROUP BY …` | `group_by` |
| 聚合后排序 | `ORDER BY …`（分组/聚合结果集） | `order_by`（**新增**，见下） |
| 聚合后过滤 | `HAVING …` | `having` |

**保留**现有字段：`offset`、`limit`、`filter_condition`、`output_fields`、`need_total`、`query_type`、`search_after` 等；连接器在 **聚合模式** 下按需选用或忽略（例如 OpenSearch composite / SQL 引擎）。

---

## 2. 与现有 `sort` 的关系

当前 `sort`（`[]SortField`，`field` + `direction`）表示 **明细行** 排序（无分组时等价 `ORDER BY` 明细列）。

为避免语义重叠，约定：

- **明细查询**（默认）：仅使用 `filter_condition` + `sort` + 分页，行为与线上一致。
- **聚合查询**：当请求中出现 **`aggregation` 或 `group_by` 或 `having`** 时，进入 **聚合模式**：
  - **`sort`**：应为空或省略；若实现选择 **严格模式**，校验失败并提示改用 `order_by`。
  - **`order_by`**：对 **分组/聚合后的结果集** 排序（列名为维度 `property` 或度量别名 `aggregation.alias` / 约定名如 `__value`）。

若实现采用 **宽松模式**：可同时存在 `sort` 与 `order_by`，由引擎定义优先级（不推荐，建议文档只支持严格模式）。

---

## 3. 聚合模式判定

满足以下 **任一** 条件即视为聚合模式：

- `aggregation` 非空，或  
- `group_by` 非空且长度 ≥ 1，或  
- `having` 非空。

可选显式字段（便于网关与客户端）：

```json
"query_shape": "document"
```

或

```json
"query_shape": "aggregate"
```

若省略 `query_shape`，由服务端按上述规则 **推断**。

---

## 4. 字段说明（聚合相关）

### 4.1 `aggregation`（object，聚合模式核心）

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `property` | string | 是 | 被聚合的 **资源字段名**（与 `Resource.schema_definition[].name` 一致）。`count` 类可不对应物理列时由实现约定（如 `*` 或省略）。 |
| `aggr` | string | 是 | `count` \| `count_distinct` \| `sum` \| `max` \| `min` \| `avg` |
| `alias` | string | 否 | 结果列别名；缺省时实现可使用 `__value` 或与 `property`+`aggr` 组合生成唯一列名。 |

**单度量**：一个请求体只包含 **一个** `aggregation` 对象；若需多度量，可多次调用或未来扩展 `aggregations: []`（本稿不展开）。

### 4.2 `group_by`（array）

| 元素字段 | 类型 | 必填 | 说明 |
|----------|------|------|------|
| `property` | string | 是 | 分组维度，对应资源字段名 |
| `description` | string | 否 | 仅文档/调试，不参与执行 |

### 4.3 `order_by`（array）

| 元素字段 | 类型 | 必填 | 说明 |
|----------|------|------|------|
| `property` | string | 是 | 排序列：维度名或度量别名（如 `__value`） |
| `direction` | string | 是 | `asc` \| `desc` |

与 BKN `calculation_formula.order_by` 一致（使用 `property` 而非 `name`）。

### 4.4 `having`（object）

与 BKN `calculation_formula.having` 同构：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `field` | string | 是 | 固定为 **`__value`**（表示对聚合度量过滤） |
| `operation` | string | 是 | `==` \| `!=` \| `>` \| `>=` \| `<` \| `<=` \| `in` \| `not_in` \| `range` \| `out_range` |
| `value` | any | 是 | 与 `operation` 匹配 |

对维度做 HAVING 时，实现可扩展允许 `field` 为维度 `property`（本稿 **最小集** 仅 `__value`）。

---

## 5. JSON Schema（请求体扩展片段，draft 2020-12）

以下为 **在现有请求体上增加** 的字段定义，可与现有 `ResourceDataQueryParams` 合并为同一 `properties` 对象。

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://adp.kweaver.ai/schemas/vega/ResourceDataQueryParamsAnalyticsExtension.json",
  "title": "ResourceDataQueryParams（聚合扩展）",
  "type": "object",
  "properties": {
    "query_shape": {
      "type": "string",
      "description": "可选。document=明细；aggregate=聚合。省略时由服务端根据 aggregation/group_by/having 推断。",
      "enum": ["document", "aggregate"]
    },
    "filter_condition": {
      "description": "行级过滤（WHERE），与现网一致；结构同 FilterCondCfg（field / operation / sub_conditions / value）。",
      "type": "object"
    },
    "aggregation": {
      "type": "object",
      "description": "聚合度量；与 group_by 等同时构成聚合查询。",
      "required": ["property", "aggr"],
      "properties": {
        "property": { "type": "string", "minLength": 1 },
        "aggr": {
          "type": "string",
          "enum": ["count", "count_distinct", "sum", "max", "min", "avg"]
        },
        "alias": { "type": "string" }
      },
      "additionalProperties": false
    },
    "group_by": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["property"],
        "properties": {
          "property": { "type": "string", "minLength": 1 },
          "description": { "type": "string" }
        },
        "additionalProperties": false
      }
    },
    "order_by": {
      "type": "array",
      "description": "聚合结果排序；聚合模式下与 sort 二选一（推荐校验互斥）。",
      "items": {
        "type": "object",
        "required": ["property", "direction"],
        "properties": {
          "property": { "type": "string", "minLength": 1 },
          "direction": { "type": "string", "enum": ["asc", "desc"] }
        },
        "additionalProperties": false
      }
    },
    "having": {
      "type": "object",
      "description": "对聚合结果过滤（HAVING）。",
      "required": ["field", "operation", "value"],
      "properties": {
        "field": { "type": "string", "const": "__value" },
        "operation": {
          "type": "string",
          "enum": [
            "==", "!=", ">", ">=", "<", "<=",
            "in", "not_in", "range", "out_range"
          ]
        },
        "value": {}
      },
      "additionalProperties": false
    },
    "sort": {
      "type": "array",
      "description": "明细行排序；聚合模式下应为空。",
      "items": {
        "type": "object",
        "required": ["field", "direction"],
        "properties": {
          "field": { "type": "string" },
          "direction": { "type": "string", "enum": ["asc", "desc"] }
        }
      }
    }
  }
}
```

**合并说明**：完整请求体 = 现有字段（`offset`、`limit`、`output_fields`、`need_total`、`query_type`、`search_after` 等） + 上表扩展。`output_fields` 在聚合模式下可列出 **维度列 + 度量列**（名与 `alias` / 约定一致）。

---

## 6. 请求示例

### 6.1 明细（与现网一致）

```json
{
  "offset": 0,
  "limit": 10,
  "sort": [{ "field": "timestamp", "direction": "desc" }],
  "filter_condition": {
    "operation": "and",
    "sub_conditions": [{ "field": "status", "operation": "==", "value": "active" }]
  },
  "output_fields": ["id", "timestamp", "status"],
  "need_total": true
}
```

### 6.2 聚合：按 namespace 分组，avg(cpu)，带过滤与 HAVING、ORDER BY

```json
{
  "offset": 0,
  "limit": 100,
  "query_shape": "aggregate",
  "filter_condition": {
    "operation": "and",
    "sub_conditions": [
      { "field": "cluster", "operation": "==", "value": "c1" }
    ]
  },
  "group_by": [{ "property": "namespace" }],
  "aggregation": { "property": "cpu_usage", "aggr": "avg", "alias": "avg_cpu" },
  "having": {
    "field": "__value",
    "operation": ">=",
    "value": 0.5
  },
  "order_by": [
    { "property": "namespace", "direction": "asc" },
    { "property": "avg_cpu", "direction": "desc" }
  ],
  "output_fields": ["namespace", "avg_cpu"],
  "need_total": true
}
```

> **注**：第二行 `order_by` 的 `property` 为 `aggregation.alias`；若未设 `alias`，实现可使用 `__value` 作为单列度量名，则 `order_by` 写 `{ "property": "__value", "direction": "desc" }`。

---

## 7. 实现与 OpenAPI 落地建议

1. **interfaces**：在 `ResourceDataQueryParams` 中增加 `Aggregation`、`GroupByItem`、`OrderByItem`、`HavingClause` 等类型（或内嵌 struct），JSON tag 与上文一致。
2. **validate**：`ValidateResourceDataQueryParams` 中增加聚合模式校验（`sort` 与 `order_by` 互斥、`having` 依赖 `aggregation` 等）。
3. **connectors**：OpenSearch / SQL 等分别将 `filter_condition` → WHERE，`group_by` + `aggregation` → GROUP BY，`having` → HAVING，`order_by` → ORDER BY；不支持聚合的引擎返回明确错误码。
4. **文档**：在 `opensearch_query_api.md` 中增加指向本文的链接或附录。

---

## 8. 与 BKN 映射（简要）

| BKN `calculation_formula` | vega `ResourceDataQueryParams` |
|---------------------------|--------------------------------|
| `condition` | `filter_condition` |
| `aggregation` | `aggregation` |
| `group_by` | `group_by` |
| `order_by` / `having` | `order_by` / `having` |

时间窗、跨资源等仍由 **ontology-query** 上层组装；vega 侧以 **单 Resource** 为边界执行上述语义。
