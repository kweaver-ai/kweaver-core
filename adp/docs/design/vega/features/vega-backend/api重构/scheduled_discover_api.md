# 定时发现资源 API 文档

本文档描述了与定时发现资源相关的API接口。

## 1. 查询定时发现任务列表

查询定时发现任务列表，支持按目录ID和任务状态过滤。

### 请求

- **方法**: `GET`
- **路径**: `/api/vega-backend/v1/catalogs/scheduled-discover`

### 查询参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | string | 否 | 目录ID，如果为空则查询所有目录的定时发现任务 |
| enabled | string | 否 | 任务状态过滤，"true"表示只返回已启用的任务，"false"表示只返回已禁用的任务 |
| offset | int | 否 | 分页偏移量，默认值由系统定义 |
| limit | int | 否 | 每页返回数量，默认值由系统定义 |
| sort | string | 否 | 排序字段，默认为"create_time" |
| direction | string | 否 | 排序方向，默认为"desc" |

### 请求示例

```
GET /api/vega-backend/v1/catalogs/scheduled-discover?id=catalog123&enabled=true&offset=0&limit=10
```

### 响应

成功时返回HTTP状态码200和以下JSON响应体：

```json
{
  "entries": [
    {
      "id": "string",
      "catalog_id": "string",
      "cron_expr": "string",
      "start_time": 1680000000000,
      "end_time": 1700000000000,
      "enabled": true,
      "strategies": ["insert", "update"],
      "last_run": 1680000000000,
      "next_run": 1680086400000,
      "creator": {
        "id": "string",
        "type": "string"
      },
      "create_time": 1679900000000,
      "updater": {
        "id": "string",
        "type": "string"
      },
      "update_time": 1679900000000
    }
  ],
  "total_count": 100
}
```

### 错误响应

| HTTP状态码 | 描述 |
|------------|------|
| 400 | 请求参数无效 |
| 404 | 目录不存在（当提供目录ID时） |
| 500 | 服务器内部错误 |

---

## 2. 创建定时发现任务

创建一个新的定时发现任务，用于定期发现指定目录中的资源。

### 请求

- **方法**: `POST`
- **路径**: `/api/vega-backend/v1/catalogs/:id/scheduled-discover`
- **Content-Type**: `application/json`

### 路径参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | string | 是 | 目录ID |

### 请求体

| 字段名 | 类型 | 必填 | 描述                                                   |
|--------|------|------|------------------------------------------------------|
| cron_expr | string | 是 | Cron表达式（五位），用于定义任务执行的时间计划                            |
| start_time | int64 | 否 | 定时任务的开始时间（Unix时间戳，毫秒）                                |
| end_time | int64 | 否 | 定时任务的结束时间（Unix时间戳，毫秒）                                |
| strategies | []string | 否 | 发现策略，可以是["insert", "delete", "update"]中的一个或多个，或为空表示全部 |

### 请求示例

```json
{
  "cron_expr": "0 2 * * *",
  "start_time": 1680000000000,
  "end_time": 1700000000000,
  "strategies": ["insert", "update"]
}
```

### 响应

成功时返回HTTP状态码200和以下JSON响应体：

```json
{
  "task_id": "string",
  "message": "Scheduled discover task created successfully"
}
```

### 错误响应

| HTTP状态码 | 描述 |
|------------|------|
| 400 | 请求参数无效，如cron表达式格式错误 |
| 404 | 目录不存在 |
| 500 | 服务器内部错误 |

---

## 3. 启动定时发现任务

启动一个已存在的定时发现任务。

### 请求

- **方法**: `POST`
- **路径**: `/api/vega-backend/v1/catalogs/:id/scheduled-discover/:task_id/start`

### 路径参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | string | 是 | 目录ID |
| task_id | string | 是 | 定时发现任务ID |

### 响应

成功时返回HTTP状态码200和以下JSON响应体：

```json
{
  "task_id": "string",
  "message": "Scheduled discover task started successfully"
}
```

### 错误响应

| HTTP状态码 | 描述 |
|------------|------|
| 400 | 请求参数无效，如任务已启用或任务不属于指定目录 |
| 404 | 目录或任务不存在 |
| 500 | 服务器内部错误 |

---

## 4. 停止定时发现任务

停止一个正在运行的定时发现任务。

### 请求

- **方法**: `POST`
- **路径**: `/api/vega-backend/v1/catalogs/:id/scheduled-discover/:task_id/stop`

### 路径参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | string | 是 | 目录ID |
| task_id | string | 是 | 定时发现任务ID |

### 响应

成功时返回HTTP状态码200和以下JSON响应体：

```json
{
  "task_id": "string",
  "message": "Scheduled discover task stopped successfully"
}
```

### 错误响应

| HTTP状态码 | 描述 |
|------------|------|
| 400 | 请求参数无效，如任务不属于指定目录 |
| 404 | 目录或任务不存在 |
| 500 | 服务器内部错误 |

---

## 5. 更新定时发现任务

更新一个已存在的定时发现任务的配置。

### 请求

- **方法**: `PUT`
- **路径**: `/api/vega-backend/v1/catalogs/:id/scheduled-discover/:task_id`
- **Content-Type**: `application/json`

### 路径参数

| 参数名 | 类型 | 必填 | 描述 |
|--------|------|------|------|
| id | string | 是 | 目录ID |
| task_id | string | 是 | 定时发现任务ID |

### 请求体

| 字段名 | 类型 | 必填 | 描述                                                    |
|--------|------|------|-------------------------------------------------------|
| cron_expr | string | 否 | Cron表达式（五位），用于定义任务执行的时间计划                             |
| start_time | int64 | 否 | 定时任务的开始时间（Unix时间戳，毫秒）                                 |
| end_time | int64 | 否 | 定时任务的结束时间（Unix时间戳，毫秒）                                 |
| strategies | []string | 否 | 发现策略，可以是["insert", "delete", "update"]中的一个或多个，或为空表示全部 |

### 请求示例

```json
{
  "cron_expr": "0 3 * * *",
  "strategies": ["delete"]
}
```

### 响应

成功时返回HTTP状态码200和以下JSON响应体：

```json
{
  "task_id": "string",
  "message": "Scheduled discover task updated successfully"
}
```

### 错误响应

| HTTP状态码 | 描述 |
|------------|------|
| 400 | 请求参数无效，如cron表达式格式错误或任务不属于指定目录 |
| 404 | 目录或任务不存在 |
| 500 | 服务器内部错误 |

---

## 注意事项

1. 所有接口都需要有效的OAuth认证令牌。
2. Cron表达式格式应符合标准Cron表达式规范：必须是五位。
3. 更新任务时，如果任务当前是启用状态，系统会自动重新调度任务。
4. 时间戳使用Unix时间戳，单位为毫秒。
