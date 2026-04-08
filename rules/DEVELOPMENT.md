---
name: development
title: KWeaver Development Specification (DEVELOPMENT)
version: 0.1.0
scope: All services in KWeaver-Core and related products (Decision Agent, ISF, TraceAI, etc.)
authors: [freeman.xu]
created: 2026-03-20
status: draft
related:
  - ARCHITECTURE.md
  - TESTING.md
  - WORKFLOW.md
  - CONTRIBUTING.md
  - RELEASE.md
---

# KWeaver Development Specification (DEVELOPMENT)

[中文](DEVELOPMENT.zh.md) | English

This document defines the development standards for KWeaver services, covering API design, HTTP semantics, request/response conventions, authentication, and observability. Complements [ARCHITECTURE](ARCHITECTURE.md) (how to decompose the system) and [TESTING](TESTING.md) (how to test).

## Terminology

Keywords in this document are interpreted per [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119):

| Keyword | Meaning |
|---------|---------|
| **MUST** | Absolute requirement, no exceptions |
| **MUST NOT** | Absolute prohibition |
| **SHOULD** | Expected in normal circumstances; exceptions require justification in the design doc |
| **SHOULD NOT** | Normally avoided; exceptions require justification |
| **MAY** | Optional, adopt as needed |

---

## 1. Error Handling

### 1.1 Error Response Structure

All services **MUST** use a unified error response structure:

```json
{
  "error_code": "INVALID_PARAMETER",
  "message": "Field 'name' is required",
  "trace_id": "req-a1b2c3d4"
}
```

| Field | Type | Requirement | Description |
|-------|------|-------------|-------------|
| `error_code` | string | MUST | Machine-readable error code, `UPPER_SNAKE_CASE`, English, semantically unique across the platform |
| `message` | string | MUST | Human-readable description for developers; **MAY** be internationalized, but `error_code` remains constant |
| `trace_id` | string | MUST | Request trace ID, consistent with header `x-trace-id` |

Additional fields (as needed):

| Field | Type | Description |
|-------|------|-------------|
| `details` | array | For multiple errors, each item contains its own `error_code` + `message` |
| `existing_id` | string | On resource conflict (409), returns the ID of the existing resource |

Rules:

- The same error **MUST** always return the same `error_code`; different errors **MUST NOT** reuse the same `error_code`.
- `error_code` **MUST** be in English; the language of `message` **MAY** vary based on the `Accept-Language` request header.
- Field names **MUST** be consistent across all services — variants such as `ErrorCode`, `Code`, `Description` are **forbidden**.
- Errors from JSON-RPC services **SHOULD** be flattened at the Gateway layer into the above format for a consistent caller experience.

### 1.2 Standard Error Codes

The following error codes are platform-reserved; all services **MUST** use them according to their semantics:

| error_code | HTTP Status | Semantics |
|------------|-------------|-----------|
| `INVALID_PARAMETER` | 400 | Invalid request parameter |
| `UNAUTHORIZED` | 401 | Not authenticated or token invalid |
| `FORBIDDEN` | 403 | Authenticated but insufficient permissions |
| `RESOURCE_NOT_FOUND` | 404 | Resource does not exist |
| `RESOURCE_EXISTED` | 409 | Resource already exists (conflict) |
| `INTERNAL_ERROR` | 500 | Unexpected server error |

Services **MAY** define business-specific error codes (e.g., `BUILD_TIMEOUT`), but **MUST** declare them in the corresponding OpenAPI spec.

---

## 2. Collections and Pagination

### 2.1 Collection Response Structure

Endpoints returning collections **MUST** use a unified envelope:

```json
{
  "entries": [
    {"id": "kn-1", "name": "Supply Chain"},
    {"id": "kn-2", "name": "Customer Scoring"}
  ],
  "total": 42
}
```

| Field | Type | Requirement | Description |
|-------|------|-------------|-------------|
| `entries` | array | MUST | Data list |
| `total` | integer | SHOULD | Total count matching filters (not the current page count) |

Rules:

- The collection field **MUST** be named `entries` — variants such as `data`, `datas`, `items`, `list`, `messages` are **forbidden**.
- An empty collection **MUST** return `{"entries": [], "total": 0}`; returning `null` or omitting the `entries` field is **forbidden**.
- Single resource retrieval (Get) **MUST** return the bare object; wrapping in `{"entries": [obj]}` is **forbidden**.

### 2.2 Pagination

Endpoints supporting pagination **MUST** implement cursor-based pagination:

**Request parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `limit` | integer | Items per page; server defines default and maximum |
| `cursor` | string | Opaque cursor from the previous response's `next_cursor` |

**Response fields:**

| Field | Type | Description |
|-------|------|-------------|
| `next_cursor` | string \| null | Next page cursor; `null` means no more data |

```json
{
  "entries": [...],
  "total": 42,
  "next_cursor": "eyJpZCI6Imtu..."
}
```

Rules:

- The cursor **MUST** be opaque to the client — clients **MUST NOT** construct or modify cursor values.
- Offset pagination (`offset` + `limit`) **MAY** be retained for backward compatibility, but new endpoints **SHOULD** prefer cursor pagination.
- Collections **MUST** have a stable sort order to ensure pagination traversal does not skip or duplicate entries.

### 2.3 Sorting

Sort parameters **SHOULD** use a unified format:

```
GET /api/v1/knowledge-networks?sort=create_time&direction=desc
```

| Parameter | Type | Description |
|-----------|------|-------------|
| `sort` | string | Sort field name |
| `direction` | string | `asc` (ascending, default) or `desc` (descending) |

### 2.4 Filtering

Filter operators **MUST** be unified across the platform:

| Operator | Description | Value type |
|----------|-------------|------------|
| `eq` | Equal | scalar |
| `neq` | Not equal | scalar |
| `gt` | Greater than | number |
| `gte` | Greater than or equal | number |
| `lt` | Less than | number |
| `lte` | Less than or equal | number |
| `in` | In list | array |
| `not_in` | Not in list | array |
| `like` | Fuzzy match | string |
| `exist` | Field exists | — |
| `not_exist` | Field does not exist | — |

Rules:

- All endpoints accepting filter conditions **MUST** use the operator names above; mixing different styles (e.g., `==` and `eq`) within the same platform is **forbidden**.

---

## 3. Standard Methods

### 3.1 Create

```
POST /api/v1/{collection}
```

Rules:

- Successful creation **SHOULD** return `201 Created` with the full representation of the new resource.
- If the resource already exists, **MUST** return `409 Conflict` with `existing_id` in the response body so the caller can `GET` the existing resource directly.
- Missing required fields **MUST** return `400` immediately at request time — accepting incomplete data and reporting errors later in async flows is **forbidden**.
- Fields the server can infer (e.g., auto-populating from referenced resources) **SHOULD** be handled server-side; clients should not be required to assemble them manually.

### 3.2 Get

```
GET /api/v1/{collection}/{id}
```

Rules:

- **MUST** return the resource object directly, without wrapping in a collection envelope.
- If the resource does not exist, **MUST** return `404`.

### 3.3 List

```
GET /api/v1/{collection}?limit=20&cursor=xxx
```

Rules:

- Response format: see [2.1 Collection Response Structure](#21-collection-response-structure).
- **MUST** support pagination (see [2.2 Pagination](#22-pagination)) even if current data volume is small. Adding pagination later is a breaking change, so it **MUST** be supported from the initial version.

### 3.4 Update

```
PUT /api/v1/{collection}/{id}        # Full replacement
PATCH /api/v1/{collection}/{id}      # Partial update
```

Rules:

- `PUT` **MUST** use full replacement semantics — fields not provided revert to defaults.
- `PATCH` **MUST** use partial update semantics — only provided fields are modified.
- If the resource does not exist, **MUST** return `404` — implicit creation (upsert) is **forbidden** unless explicitly declared in the API documentation.

### 3.5 Delete

```
DELETE /api/v1/{collection}/{id}
```

Rules:

- Successful deletion **SHOULD** return `204 No Content`.
- If the resource does not exist, **SHOULD** return `404`. Idempotent deletion (returning `204` on repeated deletes) **MAY** be supported as needed, but must be declared in the API documentation.

---

## 4. Authentication and Security

### 4.1 Token Validation

The API Gateway **MUST** equally accept tokens from all standard issuance paths:

- `access_token` issued via OAuth2 Authorization Code flow
- `access_token` issued via OAuth2 Client Credentials flow
- `access_token` from OAuth2 Refresh Token renewal
- Session token issued via browser login flow

Rules:

- Tokens issued by the same identity provider **MUST NOT** be treated differently based on their issuance path.
- Token validation **SHOULD** use a standard introspection endpoint (e.g., OAuth2 Token Introspection) and **SHOULD NOT** be tied to a specific frontend login session.

### 4.2 Authentication Header

All endpoints requiring authentication **MUST** accept the standard `Authorization` header:

```
Authorization: Bearer {access_token}
```

- Requiring clients to send non-standard headers (e.g., a custom `token` header) as a necessary authentication condition is **forbidden**.
- Business domain identifiers **SHOULD** be passed via the `x-business-domain` header, decoupled from authentication.

---

## 5. Observability

### 5.1 Request Tracing

All services **MUST** support distributed tracing:

- Every request **MUST** generate or propagate a `trace_id`.
- Response headers **MUST** include `x-trace-id` — regardless of success or failure.
- The `trace_id` in the error response body **MUST** match the `x-trace-id` in the header.

### 5.2 Request ID

- The server **SHOULD** return `x-request-id` in response headers, identifying the unique request processed by the current service.
- If the client provides `x-request-id` in the request, the server **SHOULD** correlate that value in its logs.

---

## 6. Compatibility

> For detailed versioning strategy and breaking change definitions, see [ARCHITECTURE Section 1.3](ARCHITECTURE.md). This section provides operational conventions.

### 6.1 Non-breaking Changes (Safe)

The following changes **MAY** be made within the same major version:

- Adding new fields to responses (clients **MUST** ignore unknown fields)
- Adding optional request parameters (with defaults)
- Adding new endpoints
- Adding new error codes (clients **SHOULD** handle unknown `error_code` values generically)
- Adding new enum values (clients **SHOULD** tolerate unknown enum values)

### 6.2 Breaking Changes (Forbidden within the Same Major)

The following changes **MUST** be released via a new major version (`/api/v2`):

- Removing or renaming response fields
- Removing or renaming endpoints
- Changing optional parameters to required
- Changing field types
- Changing semantics of existing error codes
- Changing the collection envelope structure

### 6.3 Content-Type Handling

- Endpoints accepting JSON body **MUST** correctly handle `Content-Type: application/json`.
- When the request body is JSON, parsing behavior **MUST** be consistent regardless of whether Content-Type is `application/json` or `text/plain`.
- Unsupported Content-Type **SHOULD** return `415 Unsupported Media Type`.

---

## 7. Server-Side Responsibility Boundaries

The following logic is the server's responsibility and **SHOULD NOT** be delegated to clients or SDKs:

| Responsibility | Description |
|----------------|-------------|
| Data type normalization | Mapping database-native types → platform standard types is done server-side |
| Reference field population | When a resource references other entities, inferable associated fields are auto-populated by the server |
| Error format flattening | Errors from internal protocols (gRPC, JSON-RPC) are converted to the unified HTTP error format at the Gateway layer |
| Default value injection | Default values for optional fields are injected server-side; client omission implies "use default" |

---

## 8. Checklist

Before submitting a PR that adds or modifies an API, verify each item:

- [ ] Error response format is unified (`error_code` + `message` + `trace_id`)
- [ ] Collection endpoints use the `{"entries": [...]}` envelope
- [ ] HTTP status codes are semantically correct (201 create / 409 conflict / 404 not found)
- [ ] Pagination is supported from the initial version
- [ ] Create conflict returns `409` + `existing_id`
- [ ] Required fields are validated at request time, not deferred to async flows
- [ ] Filter operators use platform-standard naming
- [ ] `Content-Type: application/json` works correctly
- [ ] Response headers include `x-trace-id`
- [ ] OpenAPI spec is updated
- [ ] No breaking changes (or released via a new major version)
