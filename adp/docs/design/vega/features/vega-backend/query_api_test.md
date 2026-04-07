# Query API 调用示例

## 1. 启动服务

确保 MariaDB 已启动，并修改 `server/config/vega-backend-config.yaml` 中的数据库连接信息。

```powershell
cd d:\adp\adp\vega\vega-backend\server
$env:AUTH_ENABLED="false"
go run main.go
```

服务默认监听端口 **13014**。

## 2. 健康检查

```powershell
curl http://localhost:13014/health
```

## 3. 调用 Query Execute 接口

### 单表查询

```powershell
curl -X POST "http://localhost:13014/api/vega-backend/v1/query/execute" `
  -H "Content-Type: application/json" `
  -d '{
    "query_id": "550e8400-e29b-41d4-a716-446655440000",
    "tables": [
      { "resource_id": "YOUR_RESOURCE_ID", "alias": "t1" }
    ],
    "output_fields": ["*"],
    "offset": 0,
    "limit": 10,
    "need_total": true
  }'
```

将 `YOUR_RESOURCE_ID` 替换为实际的 resource ID（可从 `/api/vega-backend/v1/resources` 获取）。

### 多表 JOIN 查询

```powershell
curl -X POST "http://localhost:13014/api/vega-backend/v1/query/execute" `
  -H "Content-Type: application/json" `
  -d '{
    "query_id": "550e8400-e29b-41d4-a716-446655440001",
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
    "sort": [{ "field": "o.id", "direction": "asc" }],
    "offset": 0,
    "limit": 100,
    "need_total": true
  }'
```

### 分页（第二页）

同一轮查询使用相同 `query_id`，将 `offset` 设为 `next_offset`：

```powershell
curl -X POST "http://localhost:13014/api/vega-backend/v1/query/execute" `
  -H "Content-Type: application/json" `
  -d '{
    "query_id": "550e8400-e29b-41d4-a716-446655440000",
    "tables": [{ "resource_id": "YOUR_RESOURCE_ID", "alias": "t1" }],
    "output_fields": ["*"],
    "offset": 10,
    "limit": 10,
    "need_total": false
  }'
```

## 4. 启用鉴权时

若 `AUTH_ENABLED` 为 true，需在请求头中携带 token：

```powershell
curl -X POST "http://localhost:13014/api/vega-backend/v1/query/execute" `
  -H "Content-Type: application/json" `
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" `
  -d '{ ... }'
```
