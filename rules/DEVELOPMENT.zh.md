---
name: development
title: KWeaver 研发规范（DEVELOPMENT）
version: 0.1.0
scope: KWeaver-Core 及关联产品（Decision Agent、ISF、TraceAI 等）的所有服务
authors: [freeman.xu]
created: 2026-03-20
status: draft
related:
  - ARCHITECTURE.zh.md
  - TESTING.zh.md
  - WORKFLOW.zh.md
  - CONTRIBUTING.zh.md
  - RELEASE.zh.md
---

# KWeaver 研发规范（DEVELOPMENT）

中文 | [English](DEVELOPMENT.md)

本文件定义 KWeaver 服务的研发规范，覆盖 API 设计、HTTP 语义、请求与响应约定、认证、可观测性等。与 [ARCHITECTURE](ARCHITECTURE.zh.md)（系统怎么拆）、[TESTING](TESTING.zh.md)（怎么测）互补。

## 术语约定

本文档中的关键词按 [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119) 解释：

| 关键词 | 含义 |
|--------|------|
| **必须**（MUST） | 绝对要求，不可违反 |
| **禁止**（MUST NOT） | 绝对禁止 |
| **应该**（SHOULD） | 正常情况下遵守，例外需在设计文档中说明理由 |
| **不应**（SHOULD NOT） | 正常情况下不做，例外需说明理由 |
| **可以**（MAY） | 可选，按需采用 |

---

## 1. 错误处理

### 1.1 错误响应结构

所有服务**必须**使用统一的错误响应结构：

```json
{
  "error_code": "INVALID_PARAMETER",
  "message": "Field 'name' is required",
  "trace_id": "req-a1b2c3d4"
}
```

| 字段 | 类型 | 要求 | 说明 |
|------|------|------|------|
| `error_code` | string | 必须 | 机器可读错误码，`UPPER_SNAKE_CASE`，英文，全平台唯一语义 |
| `message` | string | 必须 | 面向开发者的可读描述；**可以**国际化，但 `error_code` 不变 |
| `trace_id` | string | 必须 | 请求追踪 ID，与 header `x-trace-id` 一致 |

补充字段（按需）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `details` | array | 多个错误时，每项包含独立的 `error_code` + `message` |
| `existing_id` | string | 资源冲突（409）时，返回已存在资源的 ID |

规则：

- 同一错误**必须**始终返回相同的 `error_code`，不同错误**禁止**复用同一 `error_code`。
- `error_code` **必须**为英文；`message` 的语言**可以**根据请求 header `Accept-Language` 变化。
- 字段名在所有服务中**必须**一致——**禁止**出现 `ErrorCode`、`Code`、`Description` 等变体。
- JSON-RPC 服务的错误**应该**在 Gateway 层展平为上述格式，保持调用方体验一致。

### 1.2 标准错误码

以下错误码为全平台保留，所有服务**必须**按语义使用：

| error_code | HTTP 状态码 | 语义 |
|-----------|------------|------|
| `INVALID_PARAMETER` | 400 | 请求参数不合法 |
| `UNAUTHORIZED` | 401 | 未认证或 token 无效 |
| `FORBIDDEN` | 403 | 认证有效但权限不足 |
| `RESOURCE_NOT_FOUND` | 404 | 资源不存在 |
| `RESOURCE_EXISTED` | 409 | 资源已存在（冲突） |
| `INTERNAL_ERROR` | 500 | 未预期的服务端错误 |

服务**可以**定义业务专属错误码（如 `BUILD_TIMEOUT`），但**必须**在对应的 OpenAPI spec 中声明。

---

## 2. 集合与分页

### 2.1 集合响应结构

返回集合的端点**必须**使用统一信封：

```json
{
  "entries": [
    {"id": "kn-1", "name": "供应链"},
    {"id": "kn-2", "name": "客户评分"}
  ],
  "total": 42
}
```

| 字段 | 类型 | 要求 | 说明 |
|------|------|------|------|
| `entries` | array | 必须 | 数据列表 |
| `total` | integer | 应该 | 符合过滤条件的总数（不是当前页数量） |

规则：

- 集合字段**必须**统一命名为 `entries`——**禁止**使用 `data`、`datas`、`items`、`list`、`messages` 等变体。
- 空集合**必须**返回 `{"entries": [], "total": 0}`，**禁止**返回 `null` 或省略 `entries` 字段。
- 获取单个资源（Get）**必须**直接返回裸对象，**禁止**包装在 `{"entries": [obj]}` 中。

### 2.2 分页

支持分页的端点**必须**实现游标分页（cursor-based）：

**请求参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `limit` | integer | 每页条数，服务端定义默认值和上限 |
| `cursor` | string | 不透明游标，从上一次响应的 `next_cursor` 获取 |

**响应字段：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `next_cursor` | string \| null | 下一页游标；`null` 表示无更多数据 |

```json
{
  "entries": [...],
  "total": 42,
  "next_cursor": "eyJpZCI6Imtu..."
}
```

规则：

- 游标对客户端**必须**不透明——客户端**禁止**构造或修改游标值。
- 偏移分页（`offset` + `limit`）**可以**保留用于向后兼容，但新接口**应该**优先使用游标分页。
- 集合**必须**有稳定的排序，确保分页遍历不会跳过或重复条目。

### 2.3 排序

排序参数**应该**使用统一格式：

```
GET /api/v1/knowledge-networks?sort=create_time&direction=desc
```

| 参数 | 类型 | 说明 |
|------|------|------|
| `sort` | string | 排序字段名 |
| `direction` | string | `asc`（升序，默认）或 `desc`（降序） |

### 2.4 过滤

过滤条件的操作符**必须**全平台统一。采用以下标准：

| 操作符 | 说明 | 值类型 |
|--------|------|--------|
| `eq` | 等于 | scalar |
| `neq` | 不等于 | scalar |
| `gt` | 大于 | number |
| `gte` | 大于等于 | number |
| `lt` | 小于 | number |
| `lte` | 小于等于 | number |
| `in` | 在列表中 | array |
| `not_in` | 不在列表中 | array |
| `like` | 模糊匹配 | string |
| `exist` | 字段存在 | — |
| `not_exist` | 字段不存在 | — |

规则：

- 所有接受过滤条件的端点**必须**使用上述操作符命名，**禁止**同一平台内混用不同风格（如 `==` 与 `eq` 并存）。

---

## 3. 标准方法

### 3.1 Create

```
POST /api/v1/{collection}
```

规则：

- 创建成功**应该**返回 `201 Created` 和新建资源的完整表示。
- 资源已存在**必须**返回 `409 Conflict`，响应体**必须**包含 `existing_id`，调用方可直接 `GET` 获取。
- 必填字段缺失**必须**在请求时立即返回 `400`——**禁止**接受不完整数据后在异步流程中才报错。
- 服务端能推断的字段（如根据引用资源自动填充）**应该**由服务端完成，不应要求客户端手工组装。

### 3.2 Get

```
GET /api/v1/{collection}/{id}
```

规则：

- **必须**直接返回资源对象，不包装在集合信封中。
- 资源不存在**必须**返回 `404`。

### 3.3 List

```
GET /api/v1/{collection}?limit=20&cursor=xxx
```

规则：

- 响应格式见 [2.1 集合响应结构](#21-集合响应结构)。
- **必须**支持分页（见 [2.2 分页](#22-分页)），即使当前数据量小。添加分页是破坏性变更，因此**必须**从初始版本开始支持。

### 3.4 Update

```
PUT /api/v1/{collection}/{id}        # 全量替换
PATCH /api/v1/{collection}/{id}      # 部分更新
```

规则：

- `PUT` **必须**为全量替换语义——未提供的字段恢复默认值。
- `PATCH` **必须**为部分更新语义——只修改提供的字段。
- 资源不存在**必须**返回 `404`——**禁止**隐式创建（upsert），除非在 API 文档中显式声明。

### 3.5 Delete

```
DELETE /api/v1/{collection}/{id}
```

规则：

- 删除成功**应该**返回 `204 No Content`。
- 资源不存在**应该**返回 `404`。幂等删除（重复删除返回 `204`）**可以**按需支持，但需在 API 文档中声明。

---

## 4. 认证与安全

### 4.1 Token 校验

API Gateway **必须**对所有标准签发路径的 token 同等认可：

- OAuth2 Authorization Code 流程签发的 `access_token`
- OAuth2 Client Credentials 流程签发的 `access_token`
- OAuth2 Refresh Token 刷新后的 `access_token`
- 浏览器登录流程签发的 session token

规则：

- 同一身份提供者签发的 token，**禁止**因签发路径不同而被区别对待。
- Token 校验**应该**通过标准 introspection 端点（如 OAuth2 Token Introspection），**不应**绑定特定前端登录会话。

### 4.2 认证 Header

所有需要认证的端点**必须**接受标准 `Authorization` header：

```
Authorization: Bearer {access_token}
```

- **禁止**要求客户端额外发送非标准 header（如自定义 `token` header）作为认证的必要条件。
- 业务域标识**应该**通过 `x-business-domain` header 传递，与认证解耦。

---

## 5. 可观测性

### 5.1 请求追踪

所有服务**必须**支持分布式追踪：

- 每个请求**必须**生成或传播 `trace_id`。
- 响应 header **必须**包含 `x-trace-id`——无论成功或失败。
- 错误响应体中的 `trace_id` **必须**与 header 中的 `x-trace-id` 一致。

### 5.2 请求 ID

- 服务端**应该**在响应 header 中返回 `x-request-id`，标识当前服务处理的唯一请求。
- 如果客户端在请求中提供了 `x-request-id`，服务端**应该**在日志中关联该值。

---

## 6. 兼容性

> 详细的版本策略与破坏性变更定义见 [ARCHITECTURE 1.3 节](ARCHITECTURE.zh.md)。本节补充实操层面的约定。

### 6.1 非破坏性变更（安全）

以下变更**可以**在同一 major 版本内进行：

- 响应中新增字段（客户端**必须**忽略未知字段）
- 新增可选请求参数（带默认值）
- 新增端点
- 新增错误码（客户端**应该**对未知 `error_code` 做通用处理）
- 新增枚举值（客户端**应该**容忍未知枚举值）

### 6.2 破坏性变更（禁止在同一 major 内）

以下变更**必须**通过新 major 版本（`/api/v2`）发布：

- 删除或重命名响应字段
- 删除或重命名端点
- 将可选参数改为必填
- 修改字段类型
- 修改现有错误码语义
- 修改集合信封结构

### 6.3 Content-Type 处理

- 接受 JSON body 的端点**必须**正确处理 `Content-Type: application/json`。
- 请求体为 JSON 时，无论 Content-Type 为 `application/json` 还是 `text/plain`，解析行为**必须**一致。
- 不支持的 Content-Type **应该**返回 `415 Unsupported Media Type`。

---

## 7. 服务端职责边界

以下逻辑属于服务端职责，**不应**下放到客户端或 SDK：

| 职责 | 说明 |
|------|------|
| 数据类型归一化 | 数据库原始类型 → 平台标准类型的映射在服务端完成 |
| 引用字段填充 | 资源引用其他实体时，可推断的关联字段由服务端自动填充 |
| 错误格式展平 | 内部协议（gRPC、JSON-RPC）的错误在 Gateway 层转换为统一 HTTP 错误格式 |
| 默认值注入 | 可选字段的默认值由服务端注入，客户端省略即表示使用默认值 |

---

## 8. 检查清单

新增或修改 API 时，PR 提交前逐项确认：

- [ ] 错误响应格式统一（`error_code` + `message` + `trace_id`）
- [ ] 集合端点使用 `{"entries": [...]}` 信封
- [ ] HTTP 状态码语义正确（201 创建 / 409 冲突 / 404 不存在）
- [ ] 分页从初始版本开始支持
- [ ] Create 冲突返回 `409` + `existing_id`
- [ ] 必填字段在请求时校验，不延迟到异步流程
- [ ] 过滤操作符使用平台标准命名
- [ ] `Content-Type: application/json` 正常工作
- [ ] 响应 header 包含 `x-trace-id`
- [ ] OpenAPI spec 已同步更新
- [ ] 无破坏性变更（或已通过新 major 版本发布）
