# 查询与 JOIN 能力：完整 API 与实现逻辑（基于 query_id + 游标）

## 1. 概述

本文档定义**统一查询接口**：单表 / 同数据源多表 JOIN（直连）、大数据量分页（前端 limit+offset，后端游标取数）、多数据源 JOIN（预留 501）。**同一轮查询由 query_id 标识：首次查询可不传，后端生成并在响应中返回；后续分页请求必须带回该 query_id，用于后端游标 session 管理。**

| 能力 | 说明 | 实现方式 |
|------|------|----------|
| 同数据源多表 JOIN | 同一 Catalog 下多表 JOIN | 直连 MariaDB/MySQL 等 TableConnector |
| 大数据量分页 | 前端 offset+limit，后端游标取数 | query_id 关联 session，keyset 分页 |
| 多数据源 JOIN | 跨多个 Catalog | 预留，返回 501 |

**与现有架构**：沿用 Catalog、Resource、TableConnector、filter/sort 等；新增 `POST /api/vega-backend/v1/query/execute`、Join 请求体、TableConnector.ExecuteJoinQuery、基于 query_id 的游标 session。

---

## 2. 统一查询 API

### 2.1 端点与请求头

| 方法 | 路径 |
|------|------|
| POST | `/api/vega-backend/v1/query/execute` |

- `Content-Type: application/json`
- 鉴权与现有接口一致

### 2.2 请求体（Request Body）

**约定**：

- 单源/跨源不传模式，由服务端根据 `tables` 中 resource_id 解析出的 Catalog 集合识别。
- 表通过 **resource_id** 指定（非 resource name）。
- **分页**：前端仅传 **offset**、**limit**；不传 cursor。后端内部用游标取数时，依赖 **query_id** 关联"同一轮查询"的 session。
- **query_id**：**首页可不传**。用户执行该次查询的**首次请求**可以不携带 query_id，由后端生成并在响应中返回；从第二页开始（offset>0）必须携带该 query_id，用于后端游标 session 与缓存。

**示例**：

```json
{
  "tables": [
    { "resource_id": "res-orders-id", "alias": "o" },
    { "resource_id": "res-users-id", "alias": "u" }
  ],
  "joins": [
    {
      "type": "inner",
      "left_table_alias": "o",
      "right_table_alias": "u",
      "on": [
        { "left_field": "o.user_id", "right_field": "u.id" }
      ]
    }
  ],
  "output_fields": ["o.id", "o.amount", "u.name"],
  "filter_condition": { ... },
  "sort": [{ "field": "o.id", "direction": "asc" }],
  "offset": 0,
  "limit": 100,
  "need_total": true
}
```

**字段说明**：

| 字段 | 必填 | 说明 |
|------|------|------|
| **query_id** | 否（首页可不传） | 查询轮次标识；首次请求可不传，后端生成并回传；从第二页开始必须携带，用于后端游标 session（见第 4 节） |
| tables | 是 | 表列表，每项 resource_id + 可选 alias；单表时仅一个元素 |
| joins | 否 | JOIN 定义；单表时为空或不传 |
| output_fields | 否 | 输出列，可带表别名（如 o.id、u.name） |
| filter_condition | 否 | 过滤条件 |
| sort | 否 | 排序，field 可带表别名；游标分页时建议含唯一列（如主键） |
| offset | 否 | 分页偏移，首页为 0 |
| limit | 否 | 每页条数，最大 10000 |
| need_total | 否 | 是否返回总条数 |

### 2.3 响应体（200）

```json
{
  "query_id": "550e8400-e29b-41d4-a716-446655440000",
  "entries": [
    { "o.id": 1, "o.amount": 99.0, "u.name": "Alice" },
    { "o.id": 2, "o.amount": 199.0, "u.name": "Bob" }
  ],
  "total_count": 1000,
  "next_offset": 100,
  "has_more": true
}
```

| 字段 | 说明 |
|------|------|
| entries | 当前页数据，每行 key 与 output_fields 一致 |
| total_count | need_total=true 时返回，否则可省略或 null |
| next_offset | 下一页 offset 建议值（offset + limit） |
| has_more | 是否还有下一页 |

响应中**不包含** cursor/next_cursor，仅 offset/limit 与 next_offset/has_more。

### 2.4 “第一次查询 / 第二次查询”接口能力与示例

#### 能力说明

- **第一次查询（首页）**：调用 `POST /api/vega-backend/v1/query/execute`，`offset=0`，`query_id` **可不传**。后端会生成 `query_id` 并在响应体返回。
- **第二次查询（下一页/分页）**：仍然调用同一个接口，`offset=offset+limit`，并且 **必须**带上第一次响应返回的 `query_id`。后端会尝试命中 `(query_id, 上一页 offset)` 的游标缓存以走 keyset 分页；未命中则回退 OFFSET/LIMIT。

#### 第一次查询示例（不传 query_id）

```json
{
  "tables": [
    { "resource_id": "res-orders-id", "alias": "o" }
  ],
  "output_fields": ["o.id", "o.amount"],
  "sort": [{ "field": "o.id", "direction": "asc" }],
  "offset": 0,
  "limit": 100,
  "need_total": true
}
```

响应示例（后端返回 query_id）：

```json
{
  "query_id": "d3p9kq0qg4m6m4s9g6d0",
  "entries": [
    { "o.id": 1, "o.amount": 99.0 }
  ],
  "total_count": 1000,
  "next_offset": 100,
  "has_more": true
}
```

#### 第二次查询示例（必须带 query_id）

```json
{
  "query_id": "d3p9kq0qg4m6m4s9g6d0",
  "tables": [
    { "resource_id": "res-orders-id", "alias": "o" }
  ],
  "output_fields": ["o.id", "o.amount"],
  "sort": [{ "field": "o.id", "direction": "asc" }],
  "offset": 100,
  "limit": 100,
  "need_total": false
}
```

### 2.5 跨源（多数据源）响应（501）

当解析出多个不同 catalog_id 时：

- HTTP **501 Not Implemented**
- Body：`{ "code": "VegaBackend.Query.MultiCatalogNotSupported", "message": "暂不支持多数据源 JOIN，计划使用 Trino/DuckDB 实现。" }`

### 2.6 错误码（与 query_id / session 相关）

| 错误码 | HTTP | 说明 |
|--------|------|------|
| VegaBackend.Query.InvalidParameter | 400 | 请求参数非法（如 tables 为空等） |
| VegaBackend.Query.InvalidParameter.QueryIDRequired | 400 | 非首页请求未传 query_id（offset>0 时必填） |
| VegaBackend.Query.InvalidParameter.LimitExceeded | 400 | limit > 10000 |
| VegaBackend.Query.InvalidParameter.JoinTableNotInTables | 400 | joins 中 alias 不在 tables 中 |
| VegaBackend.Query.SessionExpired | 410 | 该 query_id 对应 session 已过期（可选，便于前端提示重新查询） |
| VegaBackend.Query.MultiCatalogNotSupported | 501 | 多数据源暂不支持 |
| VegaBackend.Query.CatalogNotFound | 404 | Catalog 不存在 |
| VegaBackend.Query.ResourceNotFound | 404 | Resource 不存在或不属于该 Catalog |
| VegaBackend.Query.ExecuteFailed | 500 | 执行查询失败 |

---

## 3. 基于 query_id 的 Session 管理

### 3.1 概念

- **Query Session**：与一个 **query_id** 绑定的一段服务端状态，用于在该轮查询的多次分页请求之间共享游标缓存，从而用 keyset 方式执行"下一页"请求。
- **生命周期**：从后端首次收到该 query_id 的请求开始，到 session 过期或被清理为止。

### 3.2 Session 的创建与识别

- **创建**：首次请求若未传 query_id，后端会生成一个 query_id 并在响应中返回；后续请求带回该 query_id 后，后端为该 query_id 创建/维护 session（可懒创建：仅当需要缓存游标时才写入存储）。
- **识别**：后续请求只要带上**同一 query_id**，即视为同一 query session，后端用 `(query_id, 上一页 offset)` 查找该 session 下缓存的游标。
- **不传 query_id**：
  - **首页（offset=0）**：允许不传，后端生成并回传 query_id。
  - **非首页（offset>0）**：必须传 query_id；否则返回 400（QueryIDRequired）。

### 3.3 Session 内存储内容（建议）

- **游标缓存**：key = `(query_id, offset)`，value = 该 offset 对应页的"最后一行"在 **sort** 上的键值（编码后，如 JSON+Base64）。用于下一页请求 offset' = offset + limit 时，执行 `WHERE (sort_cols) > cursor ORDER BY ... LIMIT limit`。
- **可选**：缓存该 query_id 对应的**查询参数**（tables、joins、filter、sort、limit 等）的签名或副本，用于校验同一 query_id 的多次请求参数是否一致；若不一致可返回 400 或废弃旧 session 按新参数处理（实现自定）。

### 3.4 Session 过期与失效

- **TTL（推荐）**：每个 query_id 的 session 设置**生存时间**，例如自**最后一次访问**该 query_id 起 30 分钟，或自**创建**起 1 小时，超时后删除该 query_id 下所有游标缓存。
- **最后一次访问**：任意携带该 query_id 的请求都会刷新"最后访问时间"（可选用滑动过期）。
- **过期后行为**：若请求的 query_id 对应 session 已过期，后端无法命中游标缓存，则**回退为 OFFSET/LIMIT** 执行；可选返回 **410 Gone** 或业务错误码（如 VegaBackend.Query.SessionExpired），便于前端提示"请重新执行查询"以获得更好性能。
- **前端行为**：用户点击"重新查询"或"搜索"时，前端**重新生成 query_id**，旧 query_id 对应的 session 将不再被访问，自然在 TTL 后回收。

### 3.5 同一 query_id 的请求顺序与一致性

- **顺序**：游标缓存按"(query_id, 上一页 offset)"存储，后端假设同一 query_id 的请求在逻辑上是**顺序分页**（offset=0 → offset=limit → offset=2*limit …）。若前端乱序请求（如先发 offset=200 再发 offset=100），后端可能无法命中 (query_id, 100) 的游标，则回退 OFFSET。
- **参数一致性**：同一 query_id 的多次请求，除 offset 外，tables、joins、filter、sort、limit 等应保持一致；否则游标语义可能错乱。实现上可校验并拒绝不一致请求，或按"新参数"覆盖该 query_id 的 session（并清空已有游标缓存）。

### 3.6 Session 存储与实现建议

- **存储**：内存（单机）、Redis 等，key 建议带前缀如 `vega:query_session:{query_id}`，子 key 或 field 为 offset；或单 key 存 JSON `{ "cursors": { "0": "...", "100": "..." }, "last_access": 1234567890 }`。
- **清理**：后台定时任务或访问时检查 TTL，删除过期 query_id 数据；或使用 Redis 的 expire。

---

## 4. 后端游标取数（与 query_id 结合）

### 4.1 约定

- **前端**：首页可不传 query_id（后端会在响应返回）；从第二页开始传 **offset**、**limit**、**query_id**；只收 entries、next_offset、has_more、total_count；不涉及 cursor。
- **后端**：用 **query_id** 定位 session；在 session 内用 `(query_id, 当前 offset - limit)` 即"上一页 offset"查游标缓存；命中则用 keyset SQL，否则用 OFFSET/LIMIT；返回前将本页最后一行排序键写入 `(query_id, 当前 offset)`，供下一页使用。

### 4.2 单次请求处理流程

1. 校验 tables、joins、limit 等；若 offset>0 则必须有 query_id（首页可不传）。
2. 解析 tables 得到 resources、catalog_id 集合；若多 catalog → 501。
3. 若单源：根据 **query_id** 取 session 中 key=`(query_id, offset - limit)` 的游标值（即"上一页"的游标）。
4. **若命中游标**：拼 SQL `WHERE (sort_cols) > (cursor_vals) ORDER BY sort_cols LIMIT :limit`，执行查询。
5. **若未命中**：拼 SQL `ORDER BY sort_cols LIMIT :limit OFFSET :offset`，执行查询。
6. 若 need_total：同条件 COUNT。
7. **写回 session**：将本页**最后一行**在 sort 上的键值编码后写入 `(query_id, offset)`，并刷新 session 最后访问时间（滑动 TTL）。
8. 返回 entries、total_count、next_offset、has_more。

### 4.3 Sort 与游标唯一性

- **sort** 中至少包含一个**唯一列**（如主键），否则"最后一行"可能不唯一，游标会歧义。若请求未指定，服务端可在 sort 末尾补主键（如第一张表的主键）。

### 4.4 小结

| 角色 | 行为 |
|------|------|
| 前端 | 首页可不传 query_id（后端回传）；从第二页开始携带 query_id（同一轮不变）；传 offset、limit；消费 query_id、next_offset、has_more、entries、total_count |
| 后端 | 用 query_id 定位 session；用 (query_id, 上一页 offset) 取游标；命中则 keyset SQL，否则 OFFSET；写回 (query_id, 当前 offset) 游标并刷新 TTL |

---

## 5. 实现逻辑与现有架构

### 5.1 整体流程

```
POST /api/vega-backend/v1/query/execute
  → 校验 tables、joins、limit 等；若 offset>0 则 query_id 必填（首页可不传）
  → ResourceService.GetByIDs(tables 的 resource_id) → resources[]
  → 从 resources 取 catalog_id 集合（去重）
  → 若 len(catalog_id) > 1 → 返回 501
  → 若单源：
       → 校验 joins 中 alias 均在 tables 中
       → CatalogService.GetByID(catalog_id)
       → 从 SessionStore 取 (query_id, offset - limit) 的游标
       → 若命中：拼 WHERE (sort_key) > cursor ORDER BY ... LIMIT limit，执行
       → 若未命中：拼 ORDER BY ... OFFSET offset LIMIT limit，执行
       → 若 need_total：COUNT
       → 将本页最后一行 sort 键写入 SessionStore (query_id, offset)，刷新 session TTL
       → 返回 entries, total_count, next_offset, has_more
```

### 5.2 分层与扩展

| 层次 | 新增/扩展 |
|------|-----------|
| driveradapters | query_handler：POST /query/execute，校验 query_id 等 |
| interfaces | QueryExecuteRequest（QueryID：首页可不传/后续必填）、TableInQuery、JoinSpec、SessionStore 抽象（可选） |
| logics | query.ExecuteSingleCatalogQuery；游标 session 的读/写与 TTL |
| connectors | TableConnector.ExecuteJoinQuery(ctx, catalog, resources, joins, params)；params 可带 cursor 供 keyset |
| mariadb | 多表 JOIN SQL；支持 keyset WHERE 或 OFFSET/LIMIT |

### 5.3 核心数据结构

```go
// 统一查询请求；query_id：首页可不传（后端生成并回传），非首页必填，用于游标 session
type QueryExecuteRequest struct {
    QueryID         string         `json:"query_id"`              // 首页可不传（后端生成并回传）；非首页必填
    Tables          []TableInQuery `json:"tables"`
    Joins           []JoinSpec     `json:"joins,omitempty"`
    OutputFields    []string       `json:"output_fields,omitempty"`
    FilterCondition any            `json:"filter_condition,omitempty"`
    Sort            []*SortField   `json:"sort,omitempty"`
    Offset          int            `json:"offset,omitempty"`
    Limit           int            `json:"limit,omitempty"`
    NeedTotal       bool           `json:"need_total,omitempty"`
}

type TableInQuery struct {
    ResourceID string `json:"resource_id"`
    Alias      string `json:"alias,omitempty"`
}

type JoinSpec struct {
    Type            string       `json:"type"`
    LeftTableAlias  string       `json:"left_table_alias"`
    RightTableAlias string       `json:"right_table_alias"`
    On              []JoinOnCond `json:"on"`
}

type JoinOnCond struct {
    LeftField  string `json:"left_field"`
    RightField string `json:"right_field"`
}
```

### 5.4 TableConnector 扩展

```go
ExecuteJoinQuery(ctx context.Context, catalog *interfaces.Catalog, resources []*interfaces.Resource,
    joins []*interfaces.JoinSpec, params *interfaces.ResourceDataQueryParams) (*interfaces.QueryResult, error)
```

- params 中可包含"本次游标值"（由上层从 session 取出并注入），实现 keyset 条件；若无则走 OFFSET/LIMIT。
- 单表时 resources 长度 1、joins 为空，可与现有 ExecuteQuery 等价或复用。

### 5.5 SessionStore 抽象（可选）

便于测试与替换存储：

```go
type QuerySessionStore interface {
    GetCursor(ctx context.Context, queryID string, offset int) (cursorEncoded string, ok bool)
    SetCursor(ctx context.Context, queryID string, offset int, cursorEncoded string) error
    Touch(ctx context.Context, queryID string) error // 刷新最后访问时间，用于 TTL
}
```

实现可为内存 map + 定时清理，或 Redis（key: vega:query_session:{query_id}, field: offset, value: cursor; 或单 key + JSON；expire 用滑动窗口）。

---

## 6. 与现有接口的关系

- **保留**：`POST /api/vega-backend/v1/resources/:id/data` 单表查询不变。
- **新能力**：单表也可走 `query/execute`（tables 仅 1 个，joins 为空）；同源多表仅通过 `query/execute`；**首页可不传 query_id（后端回传），后续分页必须带回**。
- **跨源**：同一请求体，服务端识别多 Catalog 后返回 501。

---

## 7. 附录：路由与错误码速查

| 场景 | 方法 | 路径 | 说明 |
|------|------|------|------|
| 单表查询（现有） | POST | `/api/vega-backend/v1/resources/:id/data` | 不变 |
| 统一查询（新） | POST | `/api/vega-backend/v1/query/execute` | 首页 query_id 可不传（后端回传）；后续分页必填；前端 offset+limit，后端游标 session |

错误码见 2.6；Session 相关见 3.4（如 SessionExpired 410）。

以上为基于 **query_id + 游标 session** 的完整方案，可直接作为开发与联调依据。
