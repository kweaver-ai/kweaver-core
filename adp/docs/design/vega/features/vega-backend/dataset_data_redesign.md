# Dataset Documents API 收敛到 `/resources/{id}/data` 技术设计文档

> **状态**：草案
> **负责人**：@待补
> **日期**：2026-05-07
> **相关 Ticket**：待补

---

## 1. 背景与目标

### 背景

vega-backend 当前 dataset documents 的 HTTP 端点设计有以下问题：

- **路径前缀冗余且形态错乱**：`/resources/dataset/{id}/docs/...` 把 "dataset" 硬塞进 path，但 `{id}` 本身已经是 resource id（且是 `Resource.category=dataset` 的 resource）。这一层 `/dataset/` 是双标记，无信息量。
- **缺 list / get 端点**：service 已实现 `ListDocuments` / `GetDocument`，但 HTTP 不暴露。客户端无法通过 vega-backend 列文档或取单文档。
- **删除按 query 走 method-override 黑魔法**：`POST /docs/query` + `X-HTTP-Method-Override: DELETE` 头，靠 HTTP 头切换语义。OpenAPI 难表达，客户端易踩坑。
- **方法-语义错配**：handler 名 `UpdateDatasetDocuments`，service 实际调 `UpsertDocuments`，命名误导。
- **错误码套了 `Resource.InternalError`**：dataset documents 操作的失败错误码用 Resource 系列。本次保留这一现状（见 §3.5），但需要在响应语义上明确。

### 目标

对外端点收敛到统一前缀 `/resources/{id}/data`，废弃 `/resources/dataset/{id}/docs/...`：

1. **collection 级走 POST + `X-HTTP-Method-Override`**：单一端点 `POST /resources/{id}/data`，通过 override 头分发到 list/create/delete-by-filter。
2. **PUT 直接表达批量 upsert**，无 override。
3. **批量按 ids 删走独立 DELETE 端点**，path 形态 `/data/{doc_ids}` 与 `/build-tasks/{ids}` `/discover-tasks/{ids}` 对齐。
4. **新增单条端点**：`GET` / `PUT` / 单条 DELETE 由 `DELETE /data/{doc_ids}` 长度 1 的退化情形覆盖。
5. **handler 层 category 校验**：写/删/单条 GET 强制 `resource.category=dataset`；override=GET 的查询保留通用语义（任意 category）。

### 非目标

- **不顶层化 dataset**：dataset 在系统中只是 `Resource.category=dataset` 的特化形态，没有独立 ID 体系，CRUD 走 `/resources/{id}`。
- **不新增 dataset 专属错误码**：复用 Resource 错误码（详见 §3.2）。
- **不改 service 接口**：`DatasetService` 现有方法（`CreateDocuments` / `UpsertDocuments` / `DeleteDocuments` / `DeleteDocumentsByQuery` / `ListDocuments` / `GetDocument`）保留不动；handler 层完成 HTTP ↔ service 的映射。
- **不改 `POST /resources/query` SQL 查询端点**（与 dataset 无关）。

## 2. 方案概览

### 2.1 端点变化总览

#### 新增 / 调整（外/内同形态）

```
POST   /api/vega-backend/v1/resources/{id}/data                  # 必须带 X-HTTP-Method-Override
       + Override: GET     → list/query（body: filter+分页）
       + Override: POST    → 批量创建（body: documents 数组）
       + Override: DELETE  → 按 filter 删（body: { filter: {...} }）

PUT    /api/vega-backend/v1/resources/{id}/data                  # 批量 upsert（body: documents，每条带 id）

GET    /api/vega-backend/v1/resources/{id}/data/{doc_id}         # 单条详情（新增）
PUT    /api/vega-backend/v1/resources/{id}/data/{doc_id}         # 单条 update（新增；id 走 path）
DELETE /api/vega-backend/v1/resources/{id}/data/{doc_ids}        # 按 ids 删，逗号分隔，单条即长度 1
```

#### 弃用（一次性删除，BREAKING）

```
- POST   /resources/dataset/{id}/docs
- PUT    /resources/dataset/{id}/docs
- DELETE /resources/dataset/{id}/docs/{ids}
- POST   /resources/dataset/{id}/docs/query     # method-override 黑魔法
```

#### 保留不变

- `POST /resources/{id}/data` + `Override: GET`（已存在，承担"通用 resource 数据查询"职责，对所有 category 生效，本次扩展为支持更多 override）
- `POST /resources/query`（SQL 查询，与本次重构无关）

### 2.2 资源关系图

```mermaid
graph LR
    Resource[Resource<br/>category=dataset] -->|owns| Documents[Documents]

    User[用户] -->|POST /data + Override: GET 列表| Documents
    User -->|POST /data + Override: POST 批量创建| Documents
    User -->|POST /data + Override: DELETE 按 filter 删| Documents
    User -->|PUT /data 批量 upsert| Documents
    User -->|GET /data/{doc_id} 单条| Documents
    User -->|PUT /data/{doc_id} 单条 update| Documents
    User -->|DELETE /data/{doc_ids} 按 ids 删| Documents
```

Documents 没有独立全局 ID 体系（doc_id 是 dataset 内 unique），必须以 resource 为 owner 嵌套；这与 build-task / discover-task（顶层化）形成对比。本设计明确承认此差异，路径保留嵌套形态但去除冗余 `/dataset/` 段。

## 3. 详细设计

### 3.1 端点边界

#### POST `/resources/{id}/data` + `X-HTTP-Method-Override`

**强制要求 `X-HTTP-Method-Override` 头**，否则 400 `VegaBackend.InvalidParameter.OverrideMethod`。

| Override 值 | 行为 | category 限制 | body |
|---|---|---|---|
| `GET` | list/query 文档 | 任意 category（兼容现有通用查询语义） | `ResourceDataQueryParams`（filter + 分页） |
| `POST` | 批量创建 | **`dataset` 强制** | `[{...}, {...}]` 文档数组；每条 id 可选（无则后端生成） |
| `DELETE` | 按 filter 删 | **`dataset` 强制** | `{ filter: {...} }`；filter 为空整批拒绝（避免误删全表） |
| 其它 | 400 | — | — |

**响应：**

- override=GET → 200 + `{ entries: [...], total_count: N }`（与 build-tasks / discover-tasks 对齐）
- override=POST → 201 + `{ ids: [...] }`（创建成功的文档 id 数组）
- override=DELETE → 204

#### PUT `/resources/{id}/data`（批量 update）

- category 强制 `dataset`
- body：`[{ id: "...", ... }, ...]`，**每条必须带 id**；缺 id 返回 400
- service 调 `UpsertDocuments`（命名为 upsert，但对外契约是"强 update / 不存在则按 id 创建"）
- 响应：200 + `{ ids: [...] }`（upsert 成功的 id 数组）

#### GET `/resources/{id}/data/{doc_id}`

- category 强制 `dataset`
- service 调 `GetDocument`
- 200 + 文档对象
- 404：文档不存在 → `VegaBackend.Resource.NotFound`（复用，详见 §3.2 决策）

#### PUT `/resources/{id}/data/{doc_id}`（单条 update）

- category 强制 `dataset`
- doc_id **以 path 为准**；body 不需要 id 字段（即使带也忽略，path 覆盖）
- handler 内部组装为 `[{id: doc_id, ...body}]` 调 `UpsertDocuments`
- 响应：200 + 文档对象 / `{ id: "..." }`

#### DELETE `/resources/{id}/data/{doc_ids}`（按 ids 删）

- category 强制 `dataset`
- `doc_ids` 逗号分隔；单条即长度 1 的退化情形
- **best-effort 语义**：缺失 doc_id 静默跳过，不返回 4xx；只要请求格式合法且不发生底层错误即 204
- 不引入 `?ignore_missing` 选项（与 build-tasks/discover-tasks 不同——documents 场景的容忍度天然高，统一行为更简洁）
- 响应：204

### 3.2 错误码（复用 Resource，**不新增** Dataset 专属）

| HTTP | errcode | 触发场景 |
|---|---|---|
| 400 | `VegaBackend.InvalidParameter.OverrideMethod` | POST 请求未带 override 头或值非 GET/POST/DELETE |
| 400 | `VegaBackend.InvalidParameter.RequestBody` | body 不合法（如 PUT 缺 id、override=DELETE 缺 filter） |
| 400 | `VegaBackend.Resource.InternalError.InvalidCategory` | 写/删/单条 GET 操作针对非 dataset 的 resource |
| 404 | `VegaBackend.Resource.NotFound` | resource 不存在 / 单条 GET 文档不存在 |
| 500 | `VegaBackend.Resource.InternalError` | 底层 OpenSearch / 索引错误 |

**注意**：单条 GET 文档不存在与 resource 不存在共用同一个 errcode，仅靠 `error_details` 区分（`error_details: "document {doc_id} not found"`）。这是复用 Resource 错误码的代价，可接受（dataset 不顶层化的代价）。

### 3.3 请求体 schema

#### `POST /data` + Override: GET

```json
{
  "filter": { ... },           // 可选；无则全表
  "offset": 0,
  "limit": 20,
  "sort": "create_time",
  "direction": "desc"
}
```

#### `POST /data` + Override: POST

```json
[
  { "id": "abc", "title": "...", "content": "..." },   // id 可选
  { "title": "...", "content": "..." }                  // 无 id 后端生成
]
```

#### `POST /data` + Override: DELETE

```json
{
  "filter": { ... }   // 必填；空 filter 返回 400 避免误删
}
```

#### `PUT /data`

```json
[
  { "id": "abc", "title": "..." },   // 每条强制带 id
  { "id": "def", "title": "..." }
]
```

#### `PUT /data/{doc_id}`

```json
{ "title": "...", "content": "..." }   // 不需要 id 字段；path 的 doc_id 为准
```

### 3.4 数据模型变更

**无表 / schema 变更。** 仅扩展 handler 层路由 + override 分发逻辑。

`DatasetService` 接口不动；现有方法清单：

```go
type DatasetService interface {
    Create(ctx, *Resource) error                // dataset 本身的索引创建（resource 生命周期，本次不动）
    Update(ctx, *Resource) error                // 同上
    Delete(ctx, id string) error                // 同上
    CheckExist(ctx, id string) (bool, error)
    
    ListDocuments(ctx, indexName string, *Resource, *ResourceDataQueryParams) ([]map[string]any, int64, error)  // 新暴露
    GetDocument(ctx, id, docID string) (map[string]any, error)                                                   // 新暴露
    
    CreateDocuments(ctx, id string, docs []map[string]any) ([]string, error)
    UpsertDocuments(ctx, id string, docs []map[string]any) ([]string, error)
    DeleteDocument(ctx, id, docID string) error                                                                   // 内部用（暂不直接暴露）
    DeleteDocuments(ctx, id, docIDs string) error
    DeleteDocumentsByQuery(ctx, indexName string, *Resource, *ResourceDataQueryParams) error
}
```

**单条 update 的 service 调用**：handler 把 `PUT /data/{doc_id}` 收到的 body 注入 path 的 doc_id，组装为 `[{id: doc_id, ...body}]`，调 `UpsertDocuments`。不新增 `UpsertDocument(ctx, id, docID, doc)` 方法（避免 service 接口膨胀，handler 层一行 wrap 即可）。

### 3.5 handler 层 category 校验流程

每个写/删/单条 GET 请求进入 handler 后：

1. 取 `resource_id` from path
2. `r.rs.GetByID(ctx, id)` 拿 resource；不存在 → 404 `Resource.NotFound`
3. **不**是 `dataset` category → 400 `Resource.InternalError.InvalidCategory`
4. 通过校验后调 service

`POST /data + Override: GET` 不做 category 校验（保留对任意 category 的通用查询能力——这是当前 `QueryResourceDataByEx` 的现有行为）。

### 3.6 与现有 `QueryResourceDataByEx` 的整合

当前 `resource_data_handler.go` 已有 `QueryResourceDataByEx` 处理 `POST /resources/{id}/data` + `Override: GET`。本次重构在**同一个 handler 函数**内扩展 override 分发：

```go
func (r *restHandler) PostResourceDataByEx(c *gin.Context) {
    method := c.GetHeader(METHOD_OVERRIDE)
    switch strings.ToUpper(method) {
    case http.MethodGet:    r.queryResourceData(c, ...)       // 现有
    case http.MethodPost:   r.createResourceData(c, ...)      // 新增
    case http.MethodDelete: r.deleteResourceDataByQuery(c, ...) // 新增
    default:                r.replyOverrideError(c, ...)
    }
}
```

旧 `dataset_handler.go` 整体废弃，逻辑迁到 `resource_data_handler.go` 或新建 `resource_data_dataset_handler.go`（实施时决定）。

## 4. 边界情况与风险

| 类型 | 描述 | 应对 |
|---|---|---|
| Override 头大小写 | `X-HTTP-Method-Override: get` vs `GET` | handler 用 `strings.ToUpper` 规范化 |
| 空 filter 删除 | override=DELETE body `{ filter: {} }` | 拒绝 400；filter 字段必须 non-empty |
| PUT batch 中部分缺 id | body 有 5 条，其中 2 条缺 id | 整批拒绝 400，列出违规 index |
| 单条 GET 与 collection POST 路径冲突 | `GET /data` vs `GET /data/abc` | gin 不冲突（前者 hit collection，后者 hit single） |
| 单条 PUT 路径有 doc_id, body 也带 id 字段 | path=abc, body.id=xyz | path 覆盖，记 warn log；不报错（容忍） |
| DELETE 单条 doc_id 含逗号 | `DELETE /data/abc,def` | split 后 ["abc", "def"] 当批量删除两条 |
| category 校验竞态 | resource 在校验和 service 调用之间被改 category | 极低概率；service 层会再次 fail，返回 500（可接受） |
| 兼容性 | 弃用旧 `/resources/dataset/{id}/docs/...` 破坏现有客户端 | 一次性 BREAKING；CHANGELOG 标注，发布通知 |
| 性能 | 现有 list/get 服务方法已支持，无新增 SQL/索引压力 | — |
| 审计 | 写 / 删都应审计 | 沿用 `audit.NewWarnLog` / `NewInfoLog` |

## 5. 替代方案

### 方案 A：dataset 顶层化为 `/datasets/{id}/docs/...`

形态最干净，但与"dataset 是 Resource.category 之一"的事实相违——会导致 `/datasets/{id}` 与 `/resources/{id}` 指向同一实体的两套端点。

**结论**：放弃。dataset 没有独立 ID 体系，顶层化会引入实体身份混乱。

### 方案 B：每种操作独立端点（不用 method-override）

```
POST   /resources/{id}/data:query      # 列表
POST   /resources/{id}/data            # 创建
POST   /resources/{id}/data:delete     # 按 filter 删
PUT    /resources/{id}/data
GET    /resources/{id}/data/{doc_id}
PUT    /resources/{id}/data/{doc_id}
DELETE /resources/{id}/data/{doc_ids}
```

**优点**：路径表达力强，OpenAPI 干净。
**缺点**：与现有 `POST /resources/{id}/data` + `Override: GET` 的查询语义不兼容；需要把所有现有调用方改路径。

**结论**：放弃。维护现有 `POST /data` + `Override: GET` 路径是兼容性硬约束。

### 最终方案：POST + override 多分发，PUT 单独表达 upsert，DELETE 按 ids 单独走

详见第 2、3 节。

## 6. 任务拆分

按 [adp/CLAUDE.md](../../../../../CLAUDE.md) 规则 5 拆分。每批前按规则 1 描述细节方案 + 验收清单 + 失败条件待批准。

- [ ] **批 1：handler 层重构 — POST 多分发**
  - `driveradapters/resource_data_handler.go`：扩展 `POST /resources/{id}/data` 为 override 多分发，新增 create / delete-by-query 两条分支；保留现有 query 语义不变
  - `driveradapters/dataset_handler.go`：保留旧 4 个 handler 函数，但**不**注册新路由（待批 4 删除）
  - 不动 router；不删旧路由
  - 影响范围最小，可独立验证

- [ ] **批 2：handler 层新增 PUT / GET single / DELETE by ids**
  - `driveradapters/resource_data_handler.go`：新增 `PutResourceDataByEx/ByIn`（批量 upsert）、`GetResourceDataDocByEx/ByIn`（单条）、`PutResourceDataDocByEx/ByIn`（单条 update）、`DeleteResourceDataByEx/ByIn`（按 ids 删）
  - 每个 handler 内部做 category 校验（dataset 强制）
  - 单条 PUT 调用 `UpsertDocuments` 时注入 path 的 doc_id

- [ ] **批 3：router 挂新路由**
  - `driveradapters/router.go`：在外/内 `resources` group 下新增：
    - `PUT /:id/data`、`GET /:id/data/:doc_id`、`PUT /:id/data/:doc_id`、`DELETE /:id/data/:doc_ids`
  - 旧 `/resources/dataset/...` 路由暂留
  - 单元测试覆盖 path 匹配、override 多分发

- [ ] **批 4：弃用旧路由**
  - `driveradapters/router.go`：删除 4 条旧路由
  - `driveradapters/dataset_handler.go`：删除整个文件
  - CHANGELOG 标注 BREAKING

- [ ] **批 5：OpenAPI yaml**
  - `adp/docs/api/vega/vega-backend-api/resource-data.yaml`（或合并进既有 `resource.yaml`，实施时定）

## 7. 已决定事项

1. **dataset 不顶层化**：仅作为 `Resource.category=dataset` 存在；CRUD 走 `/resources`，仅 documents 子资源独立 API。

2. **`POST + Override` 三分发**：override=GET（list，任意 category）/ POST（create，dataset 强制）/ DELETE（按 filter 删，dataset 强制）。无 override 头 → 400。

3. **批量按 ids 删走独立 DELETE 端点**：`DELETE /data/{doc_ids}`，path 形态对齐 `/build-tasks/{ids}` `/discover-tasks/{ids}`。**best-effort 语义**：缺失 id 静默跳过，无 `?ignore_missing` 选项。

4. **PUT 直接表达批量 upsert**：保留 `PUT /data`（批量）+ `PUT /data/{doc_id}`（单条），不走 override。

5. **POST(create) vs PUT(update) 语义对外契约**：
   - POST：id 可选，无则后端生成
   - PUT：id 强制（batch 在 body，single 在 path）

6. **handler 层 category 校验**：所有写/删/单条 GET 强制 `resource.category=dataset`；override=GET 的查询保留对任意 category 的兼容。

7. **错误码复用 Resource，不新增 Dataset.\***：单条文档不存在与 resource 不存在共用 `Resource.NotFound`，靠 `error_details` 区分。

8. **list 响应字段对齐**：`{ entries, total_count }`，与 build-tasks / discover-tasks 一致。

9. **旧路径一次性删除**：`/resources/dataset/{id}/docs/...` 4 条 BREAKING 删除，无并存窗口。

10. **不新增 service 接口**：`DatasetService` 已有方法清单覆盖所有 HTTP 操作，handler 层完成 wrap。
