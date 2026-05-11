# 查询逻辑与流程验证

本文档对照 VEGA_Part7_QueryJoinApiAndImplementation.md 验证实现逻辑与流程。

## 1. 整体流程对照

| 文档步骤 | 实现位置 | 验证状态 |
|----------|----------|----------|
| 校验 query_id 必填 | `query_service.validateRequest` | ✅ |
| 校验 tables、joins、limit | `query_service.validateRequest` | ✅ |
| ResourceService.GetByIDs | `query_service.Execute` L63 | ✅ |
| 从 resources 取 catalog_id 集合 | `query_service.Execute` L80-94 | ✅ |
| 若 len(catalog_id) > 1 → 501 | `query_service.Execute` L85-88 | ✅ |
| 校验 joins 中 alias 均在 tables 中 | `query_service.Execute` L96-111 | ✅ |
| CatalogService.GetByID | `query_service.Execute` L113-122 | ✅ |
| 从 SessionStore 取 (query_id, offset - limit) 游标 | `query_service.Execute` L156-166 | ✅ |
| 若命中：keyset SQL | `mariadb_join_query.ExecuteJoinQuery` L147-155 | ✅ |
| 若未命中：OFFSET/LIMIT | `mariadb_join_query.ExecuteJoinQuery` L211-214 | ✅ |
| need_total：COUNT | `mariadb_join_query.ExecuteJoinQuery` L176-208 | ✅ |
| 写回 (query_id, offset) 游标 | `query_service.Execute` L211-238 | ✅ |
| 刷新 session TTL | `query_service.Execute` L166, session_store.Touch | ✅ |
| 返回 entries, total_count, next_offset, has_more | `query_service.Execute` L240-249 | ✅ |

## 2. 游标逻辑验证

### 2.1 上一页 offset 计算

```
prevOffset = offset - limit
若 prevOffset < 0，则 prevOffset = -1（表示首页，无上一页）
```

- offset=0, limit=100 → prevOffset=-1 → 不查游标，用 OFFSET
- offset=100, limit=100 → prevOffset=0 → 查 (query_id, 0) 游标
- offset=200, limit=100 → prevOffset=100 → 查 (query_id, 100) 游标

### 2.2 Keyset 条件

- 游标编码：`base64(json.Marshal([sort_col1_val, sort_col2_val, ...]))`
- SQL 条件：`WHERE (col1, col2, ...) > (?, ?, ...) ORDER BY ... LIMIT limit`

### 2.3 游标写入

- 仅当 `len(result.Rows) > 0 && len(req.Sort) > 0` 时写入
- 取最后一行在 sort 字段上的值，编码后写入 `(query_id, offset)`

## 3. Session 管理验证

| 能力 | 实现 | 验证 |
|------|------|------|
| GetCursor | memorySessionStore.GetCursor | ✅ |
| SetCursor | memorySessionStore.SetCursor | ✅ |
| Touch 刷新 TTL | memorySessionStore.Touch | ✅ |
| TTL 30 分钟 | SessionTTL 常量 | ✅ |
| 定时清理 | cleanupLoop 每 5 分钟 | ✅ |

## 4. 错误码验证

| 错误码 | 触发场景 | 实现 |
|--------|----------|------|
| QueryIDRequired | query_id 为空 | validateRequest | ✅ |
| InvalidParameter | tables 为空 | validateRequest | ✅ |
| LimitExceeded | limit > 10000 | validateRequest | ✅ |
| JoinTableNotInTables | joins 中 alias 不在 tables | Execute | ✅ |
| MultiCatalogNotSupported | 多 catalog | Execute | ✅ |
| CatalogNotFound | catalog 不存在 | Execute | ✅ |
| ResourceNotFound | resource 不存在 | Execute | ✅ |
| ExecuteFailed | 执行失败 | Execute | ✅ |

## 5. 单表 / 多表 JOIN 验证

### 5.1 单表

- tables 仅 1 个，joins 为空
- ExecuteJoinQuery 走单表 FROM 逻辑
- 支持 filter、sort、output_fields、分页

### 5.2 多表 JOIN

- 支持 INNER/LEFT/RIGHT JOIN
- ON 条件支持多字段
- output_fields 支持 alias.field 格式

## 6. 单元测试覆盖

| 测试 | 覆盖 |
|------|------|
| TestValidateRequest_QueryIDRequired | query_id 必填 |
| TestValidateRequest_EmptyTables | tables 非空 |
| TestValidateRequest_LimitExceeded | limit 上限 |
| TestValidateRequest_DefaultLimitAndOffset | 默认值 |
| TestValidateRequest_ValidRequest | 合法请求 |
| TestCursorPrevOffsetLogic | 游标 prevOffset 计算 |
| TestMockSessionStore | SessionStore 行为 |
| TestMemorySessionStore_* | 内存 SessionStore |

## 7. 运行测试

```bash
cd server
$env:I18N_MODE_UT="true"; go test ./logics/query/... -v -count=1
```
