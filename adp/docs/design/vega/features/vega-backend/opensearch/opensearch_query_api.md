# OpenSearch查询接口文档

## 接口信息
- **接口路径**: `POST /api/vega-backend/v1/resources/:id/data`
- **请求方法**: POST
- **功能描述**: 查询OpenSearch类型资源的数据

## 请求参数

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | 资源ID |

### 请求体参数
```json
{
  "offset": 0,
  "limit": 10,
  "sort": [
    {
      "field": "timestamp",
      "direction": "desc"
    }
  ],
  "filter_condition": {
    "operation": "and",
    "sub_conditions": [
      {
        "field": "status",
        "operation": "==",
        "value": "active"
      }
    ]
  },
  "output_fields": ["id", "name", "status", "timestamp"],
  "need_total": true
}
```

### 参数说明
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| offset | integer | 否 | 偏移量，默认0 |
| limit | integer | 否 | 返回记录数，默认10 |
| sort | array | 否 | 排序条件 |
| sort[].field | string | 是 | 排序字段 |
| sort[].direction | string | 是 | 排序方向，asc或desc |
| filter_condition | object | 否 | 过滤条件 |
| filter_condition.operation | string | 是 | 条件操作符，如and、or等 |
| filter_condition.sub_conditions | array | 否 | 子条件列表 |
| filter_condition.field | string | 否 | 字段名 |
| filter_condition.value | any | 否 | 字段值 |
| output_fields | array | 否 | 指定返回的字段列表 |
| need_total | boolean | 否 | 是否返回总数，默认false |

## 响应参数

### 成功响应
```json
{
  "entries": [
    {
      "_id": "doc_id",
      "_score": 1.0,
      "field1": "value1",
      "field2": "value2"
    }
  ],
  "total_count": 100
}
```

### 参数说明
| 参数名 | 类型 | 说明 |
|--------|------|------|
| entries | array | 数据记录列表 |
| entries[]._id | string | 文档ID |
| entries[]._score | number | 相关性得分 |
| entries[].field | any | 文档字段 |
| total_count | integer | 总记录数（need_total为true时返回） |

## 请求示例

### 示例1: 基本查询
```json
{
  "offset": 0,
  "limit": 10
}
```

### 示例2: 带排序的查询
```json
{
  "offset": 0,
  "limit": 10,
  "sort": [
    {
      "field": "create_time",
      "direction": "desc"
    }
  ]
}
```

### 示例3: 带过滤条件的查询
```json
{
  "offset": 0,
  "limit": 10,
  "filter_condition": {
    "field": "status",
    "operation": "==",
    "value": "active"
  }
}
```

### 示例4: 组合条件查询
```json
{
  "offset": 0,
  "limit": 10,
  "filter_condition": {
    "operation": "and",
    "sub_conditions": [
      {
        "field": "status",
        "operation": "==",
        "value": "active"
      },
      {
        "field": "age",
        "operation": ">",
        "value": 18
      }
    ]
  }
}
```

### 示例5: 指定输出字段
```json
{
  "offset": 0,
  "limit": 10,
  "output_fields": ["id", "name", "status"],
  "need_total": true
}
```

## 错误响应

### 400 Bad Request
```json
{
  "error": "Invalid request parameter"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```
