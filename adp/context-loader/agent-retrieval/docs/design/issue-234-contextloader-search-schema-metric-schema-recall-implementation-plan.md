# Issue 234 `search_schema` Metric Schema Recall Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改变 `search_schema` 既有产品边界的前提下，为 ContextLoader 增加 `metric_types` 召回能力，并完成 BKN adapter、`search_schema` 主逻辑、HTTP/MCP 契约与发布物更新。

**Architecture:** 复用现有 `SearchSchema -> KnSearch -> Filter` 主链路，在 `SearchSchema` 包装层新增仅作用于 `metric_types` 的 mixed recall。BKN 侧新增 `SearchMetricTypes` 适配；内部 object 候选只作为 `metric_types` 的 expansion 线索，不反向影响其他资源桶。

**Tech Stack:** Go, Gin, MCP JSON Schema, OpenAPI YAML, BKN Backend HTTP API

---

**Status:** Draft  
**Owner:** 待确认  
**Updated:** 2026-04-20

## Checklist

### Acceptance Checklist

- [ ] `search_schema` 新增顶层 `metric_types`
- [ ] `search_scope` 新增 `include_metric_types`，默认 `true`
- [ ] `metric_types` 采用两条 mixed recall 路径：
  - query direct recall
  - 基于内部 object 候选的 metric expansion
- [ ] mixed recall 仅增强 `metric_types`，不引入 `metric -> object` 反向补齐
- [ ] `metric_types` 输出字段采用裁剪后的 `MetricDefinition`
- [ ] `schema_brief=true/false` 在 V1 返回同一字段集
- [ ] HTTP `kn_search` 兼容契约、MCP `search_schema` schema、release/toolset 文档同步更新

### Failure Checklist

- [ ] 不将 `scope_ref` 映射为 `object_type_id`
- [ ] 不新增 `object_type_name` 等派生字段
- [ ] 不让 `search_scope` 阻断内部 object 候选参与 metric expansion
- [ ] 不在本期引入 `metric -> object` 反向补齐
- [ ] 不让 mixed recall 反向影响 `object_types`、`relation_types`、`action_types`
- [ ] 单批执行修改不超过 3 个文件

### Edge-Case Checklist

- [ ] `include_metric_types=false` 时响应不返回 `metric_types`
- [ ] `include_object_types=false` 且 `include_metric_types=true` 时仍可做 metric expansion
- [ ] direct recall 与 expansion recall 命中同一 metric 时按 `id` 去重
- [ ] direct recall 为空但 expansion recall 非空时仍返回 `metric_types`
- [ ] 四个 scope 开关同时为 `false` 时返回 400
- [ ] `schema_brief=true` 时 `metric_types` 字段集与 `false` 保持一致

## File Map

### Core Files

- `server/interfaces/driven_bkn_backend.go`
  - 新增 `MetricType`、`MetricTypeConcepts` 和 `SearchMetricTypes`
- `server/drivenadapters/bkn_backend.go`
  - 对接 BKN `/metrics` 检索接口
- `server/drivenadapters/bkn_backend_test.go`
  - 覆盖 `SearchMetricTypes` 成功、404、HTTP 错误、字段解析
- `server/interfaces/search_schema.go`
  - 扩展 `SearchSchemaScope` 和 `SearchSchemaResp`
- `server/logics/knsearch/search_schema.go`
  - 在现有包装层编排 metric mixed recall、去重、截断
- `server/logics/knsearch/search_schema_test.go`
  - 覆盖默认行为、scope 行为、mixed recall 合并逻辑

### Interface Fallout Files

- `server/logics/knsearch/mock_test.go`
  - 为扩展后的 `BknBackendAccess` 接口补空实现
- `server/logics/knfindskills/mock_test.go`
  - 为扩展后的 `BknBackendAccess` 接口补空实现

### External Contract Files

- `server/driveradapters/knsearch/index_test.go`
  - 覆盖 HTTP `search_schema`/`kn_search` 对 `include_metric_types` 的透传与错误处理
- `docs/apis/api_private/kn_search.yaml`
  - 更新兼容 HTTP 契约示例与响应结构
- `server/driveradapters/mcp/schemas/search_schema.json`
  - 新增 `include_metric_types` 与 `metric_types`
- `server/driveradapters/mcp/schemas/tools_meta.json`
  - 更新 `search_schema` 描述，纳入 metric schema
- `server/driveradapters/mcp/tools_field_reduction_test.go`
  - 覆盖 MCP 入参与结构化输出
- `docs/release/tool-usage-guide.md`
  - 更新 `search_schema` 能力说明
- `docs/release/toolset/context_loader_toolset.adp`
  - 更新 toolset 输出
- `docs/release/overview.md`
  - 更新 release 概述

## Todo List

### Task 1: BKN Metric Adapter Contract

**Files:**
- Modify: `server/interfaces/driven_bkn_backend.go`
- Modify: `server/drivenadapters/bkn_backend.go`
- Test: `server/drivenadapters/bkn_backend_test.go`

- [ ] **Step 1: Write the failing adapter tests**

```go
func TestSearchMetricTypes_Success(t *testing.T) {
	convey.Convey("SearchMetricTypes returns parsed metric entries", t, func() {
		// expect POST /metrics with x-http-method-override=GET
		// expect response body {"entries":[{"id":"m_1","name":"cpu_usage","metric_type":"atomic","scope_type":"object_type","scope_ref":"pod","calculation_formula":{"op":"avg"}}]}
	})
}

func TestSearchMetricTypes_HTTPError(t *testing.T) {
	convey.Convey("SearchMetricTypes wraps HTTP client errors", t, func() {})
}

func TestSearchMetricTypes_NotFound(t *testing.T) {
	convey.Convey("SearchMetricTypes wraps 404 as HTTP error", t, func() {})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./server/drivenadapters -run 'TestSearchMetricTypes'`
Expected: FAIL with missing `SearchMetricTypes`

- [ ] **Step 3: Add metric contract in the adapter interface**

```go
type MetricType struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Comment            string `json:"comment,omitempty"`
	UnitType           string `json:"unit_type,omitempty"`
	Unit               string `json:"unit,omitempty"`
	MetricType         string `json:"metric_type"`
	ScopeType          string `json:"scope_type"`
	ScopeRef           string `json:"scope_ref"`
	TimeDimension      any    `json:"time_dimension,omitempty"`
	CalculationFormula any    `json:"calculation_formula"`
	AnalysisDimensions any    `json:"analysis_dimensions,omitempty"`
}

type MetricTypeConcepts struct {
	Entries    []*MetricType `json:"entries"`
	TotalCount int64         `json:"total_count,omitempty"`
}
```

- [ ] **Step 4: Write the minimal adapter implementation**

```go
func (b *bknBackendAccess) SearchMetricTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/metrics", b.baseURL, query.KnID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"

	respCode, respBody, err := b.httpClient.PostNoUnmarshal(ctx, src, header, query)
	metricTypes := &interfaces.MetricTypeConcepts{}
	if err != nil {
		return metricTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] SearchMetricTypes request failed, err: %v", err))
	}
	if err := sonic.Unmarshal(respBody, metricTypes); err != nil {
		return nil, err
	}
	return metricTypes, nil
}
```

- [ ] **Step 5: Run adapter tests**

Run: `go test ./server/drivenadapters -run 'TestSearchMetricTypes'`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/interfaces/driven_bkn_backend.go server/drivenadapters/bkn_backend.go server/drivenadapters/bkn_backend_test.go
git commit -m "feat: add bkn metric type adapter"
```

---

### Task 2: Interface Fallout in Test Doubles

**Files:**
- Modify: `server/logics/knsearch/mock_test.go`
- Modify: `server/logics/knfindskills/mock_test.go`

- [ ] **Step 1: Add empty metric search stubs**

```go
func (m *mockBknBackend) SearchMetricTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
	return nil, nil
}
```

```go
func (m *testBknBackend) SearchMetricTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
	return nil, nil
}
```

- [ ] **Step 2: Run compile-focused tests**

Run: `go test ./server/logics/knsearch ./server/logics/knfindskills -run 'TestDoesNotExist'`
Expected: package compile succeeds, no interface implementation error

- [ ] **Step 3: Commit**

```bash
git add server/logics/knsearch/mock_test.go server/logics/knfindskills/mock_test.go
git commit -m "test: update bkn backend test doubles for metric search"
```

---

### Task 3: `search_schema` Mixed Recall for `metric_types`

**Files:**
- Modify: `server/interfaces/search_schema.go`
- Modify: `server/logics/knsearch/search_schema.go`
- Test: `server/logics/knsearch/search_schema_test.go`

- [ ] **Step 1: Write failing service tests**

```go
func TestSearchSchema_DefaultsIncludeMetricTypes(t *testing.T) {}
func TestSearchSchema_ExcludeMetricTypesFromResponse(t *testing.T) {}
func TestSearchSchema_MergesDirectAndExpansionMetrics(t *testing.T) {}
func TestSearchSchema_AllScopeDisabled_ReturnsBadRequest(t *testing.T) {}
```

- [ ] **Step 2: Run service tests to verify failure**

Run: `go test ./server/logics/knsearch -run 'TestSearchSchema_(DefaultsIncludeMetricTypes|ExcludeMetricTypesFromResponse|MergesDirectAndExpansionMetrics|AllScopeDisabled_ReturnsBadRequest)'`
Expected: FAIL because request/response types and service logic do not include `metric_types`

- [ ] **Step 3: Extend request and response types**

```go
type SearchSchemaScope struct {
	IncludeObjectTypes   *bool `json:"include_object_types,omitempty" default:"true"`
	IncludeRelationTypes *bool `json:"include_relation_types,omitempty" default:"true"`
	IncludeActionTypes   *bool `json:"include_action_types,omitempty" default:"true"`
	IncludeMetricTypes   *bool `json:"include_metric_types,omitempty" default:"true"`
}

type SearchSchemaResp struct {
	ObjectTypes   []any `json:"object_types"`
	RelationTypes []any `json:"relation_types"`
	ActionTypes   []any `json:"action_types"`
	MetricTypes   []any `json:"metric_types"`
}
```

- [ ] **Step 4: Add metric mixed recall in `SearchSchema`**

```go
func (s *knSearchService) SearchSchema(ctx context.Context, req *interfaces.SearchSchemaReq) (*interfaces.SearchSchemaResp, error) {
	knReq, scope, err := NormalizeSearchSchemaReq(req)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}

	resp, err := s.KnSearch(ctx, knReq)
	if err != nil {
		return nil, err
	}

	metricTypes := []any{}
	if scope.IncludeMetricTypes {
		directMetrics, err := s.searchMetricTypesDirect(ctx, req)
		if err != nil {
			return nil, err
		}
		expansionMetrics, expansionErr := s.searchMetricTypesByObjectCandidates(ctx, req, resp.ObjectTypes)
		if expansionErr != nil {
			s.Logger.WithContext(ctx).Warnf("metric expansion failed, fallback to direct recall: %v", expansionErr)
		}
		metricTypes = mergeMetricTypesByID(directMetrics, expansionMetrics, *req.MaxConcepts)
	}

	return FilterSearchSchemaResp(resp, metricTypes, scope, *req.MaxConcepts), nil
}
```

- [ ] **Step 5: Preserve V1 field strategy**

```go
// Keep BKN field names unchanged:
// id, name, comment, unit_type, unit, metric_type,
// scope_type, scope_ref, time_dimension,
// calculation_formula, analysis_dimensions
```

- [ ] **Step 6: Run service tests**

Run: `go test ./server/logics/knsearch -run 'TestSearchSchema_(DefaultsIncludeMetricTypes|ExcludeMetricTypesFromResponse|MergesDirectAndExpansionMetrics|AllScopeDisabled_ReturnsBadRequest|AppliesMaxConceptsPerResourceType|LimitsObjectTypesWhenRelationTypesExcluded)'`
Expected: PASS

- [ ] **Step 7: Commit**

```bash
git add server/interfaces/search_schema.go server/logics/knsearch/search_schema.go server/logics/knsearch/search_schema_test.go
git commit -m "feat: add metric types to search schema"
```

---

### Task 4: HTTP Compatibility Contract

**Files:**
- Test: `server/driveradapters/knsearch/index_test.go`
- Modify: `docs/apis/api_private/kn_search.yaml`

- [ ] **Step 1: Add failing handler tests**

```go
func TestSearchSchema_IncludeMetricTypesPassedThrough(t *testing.T) {}
func TestSearchSchema_AllFourScopeDisabled_ReturnsBadRequest(t *testing.T) {}
```

- [ ] **Step 2: Run handler tests**

Run: `go test ./server/driveradapters/knsearch -run 'TestSearchSchema_(IncludeMetricTypesPassedThrough|AllFourScopeDisabled_ReturnsBadRequest)'`
Expected: FAIL before request binding and docs are aligned

- [ ] **Step 3: Update the HTTP contract**

```yaml
search_scope:
  type: object
  properties:
    include_metric_types:
      type: boolean
      default: true
metric_types:
  type: array
  items:
    type: object
    additionalProperties: true
```

- [ ] **Step 4: Re-run handler tests**

Run: `go test ./server/driveradapters/knsearch -run 'TestSearchSchema_(IncludeMetricTypesPassedThrough|AllFourScopeDisabled_ReturnsBadRequest)'`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add server/driveradapters/knsearch/index_test.go docs/apis/api_private/kn_search.yaml
git commit -m "test: cover metric search schema http contract"
```

---

### Task 5: MCP Tool Contract

**Files:**
- Modify: `server/driveradapters/mcp/schemas/search_schema.json`
- Modify: `server/driveradapters/mcp/schemas/tools_meta.json`
- Test: `server/driveradapters/mcp/tools_field_reduction_test.go`

- [ ] **Step 1: Add failing MCP tests**

```go
func TestHandleSearchSchema_IncludeMetricTypesPassedThrough(t *testing.T) {}
func TestHandleSearchSchema_ReturnsMetricTypes(t *testing.T) {}
```

- [ ] **Step 2: Run MCP tests**

Run: `go test ./server/driveradapters/mcp -run 'TestHandleSearchSchema_(IncludeMetricTypesPassedThrough|ReturnsMetricTypes)'`
Expected: FAIL because tool schema and output do not include `metric_types`

- [ ] **Step 3: Update MCP input and output schema**

```json
"include_metric_types": {
  "type": "boolean",
  "default": true,
  "description": "是否返回指标类"
}
```

```json
"metric_types": {
  "type": "array",
  "description": "指标类型列表",
  "items": {
    "type": "object",
    "additionalProperties": true
  }
}
```

- [ ] **Step 4: Update the MCP tool description**

```json
"description": "统一的 Schema 探索入口。根据 query 返回相关 object_types、relation_types、action_types、metric_types，供 query_*、find_*、get_* 工具继续消费。"
```

- [ ] **Step 5: Re-run MCP tests**

Run: `go test ./server/driveradapters/mcp -run 'TestHandleSearchSchema_(IncludeMetricTypesPassedThrough|ReturnsMetricTypes)'`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add server/driveradapters/mcp/schemas/search_schema.json server/driveradapters/mcp/schemas/tools_meta.json server/driveradapters/mcp/tools_field_reduction_test.go
git commit -m "feat: expose metric types in mcp search schema"
```

---

### Task 6: Release Documentation

**Files:**
- Modify: `docs/release/tool-usage-guide.md`
- Modify: `docs/release/toolset/context_loader_toolset.adp`
- Modify: `docs/release/overview.md`

- [ ] **Step 1: Update release docs**

```md
- `search_schema` 新增 `metric_types`
- `search_scope` 新增 `include_metric_types`
- 默认行为为 object / relation / action / metric 四类全开
```

- [ ] **Step 2: Verify docs contain the new contract**

Run: `rg -n "metric_types|include_metric_types" docs/release`
Expected: all three release files contain the new terms

- [ ] **Step 3: Commit**

```bash
git add docs/release/tool-usage-guide.md docs/release/toolset/context_loader_toolset.adp docs/release/overview.md
git commit -m "docs: update release assets for metric schema recall"
```

## Self-Review Checklist

### Spec Coverage

- PRD 中关于 `metric_types`、`include_metric_types`、mixed recall、字段策略、默认行为、兼容契约的要求，均已映射到 Task 1-Task 6
- 未发现需要额外新增的实现任务

### Placeholder Scan

- 已移除单独的“TODO later”类措辞
- 每个任务都带有可执行步骤、命令和提交步骤
- 唯一保留的高层片段是测试函数名和最小实现骨架，用于指导落地，不承载未定义范围

### Type Consistency

- 全文统一使用 `metric_types`、`include_metric_types`、`SearchMetricTypes`、`MetricTypeConcepts`
- 未再引入 `object_type_id`、`object_type_name` 等与已确认字段策略冲突的命名

## Execution Handoff

Plan complete and saved to `docs/design/issue-234-contextloader-search-schema-metric-schema-recall-implementation-plan.md`. Two execution options:

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

Which approach?
