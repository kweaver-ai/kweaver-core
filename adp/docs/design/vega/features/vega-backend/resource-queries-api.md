# 统一查询接口设计文档

## 接口描述

统一查询接口是系统提供的通用数据查询服务，支持多种数据源的查询操作。通过该接口，客户端可以查询任意resource的数据，支持VEGA-SQL查询和方言查询。

该接口设计灵活，支持以下查询类型：
- VEGA-SQL查询（默认）
- 方言查询（通过`resource_type`指定，如`resource_type=mysql`，`resource_type=opensearch`等）

接口支持流式查询通过query_type参数指定（只在SQL类查询中有效）。
- standard（默认标准查询，最多返回10000条数据，如需获取更多数据，请使用stream 流式查询）
- stream 流式查询（适用于大数据量）

## 逻辑设计

VEGA-SQL查询模式和方言查询模式流程基本相同，处理流程如下：

1. **resource_id提取与资源映射**
   - 从SQL语句中提取其中的`resource_id`
   - 通过资源元数据服务查询`resource_id`对应的实际数据源信息
   - 将SQL中的`resource_id`替换为对应的`schema.table`格式
   - 所有schema.table根据数据源类型不同，添加对应的转义字符（如VEGA-SQL的`"`，MySQL的`''`等）

2. **查询路由判断**
   - 分析SQL中涉及的所有表所属的数据源
   - 如果SQL涉及多个数据源（跨源查询），路由到ETrino引擎执行（方言查询模式则转换为VEGA-SQL语法）
   - 如果SQL只涉及单个数据源（单源查询），转换为对应数据源的方言SQL（方言查询模式不需要转换）

3. **查询执行**
   - **跨源查询**：将转换后的SQL提交给ETrino引擎执行，ETrino负责跨数据源的联邦查询
   - **单源查询**：
     - 根据数据源类型（MySQL、OpenSearch等）将SQL转换为对应的方言SQL（方言查询模式不需要转换）
     - 使用对应数据源的驱动直接执行查询
     - 如果执行失败，直接返回错误信息
   
4. **错误处理**
   - 所有查询模式下，如果执行失败，系统将直接返回错误信息，不进行重试或转换
   - 错误信息包含错误码、错误描述和可能的错误详情，帮助调用方定位问题

## SQL方言转换（sqlglot）
- sqlglot是一个python的SQL方言转换库，支持多种SQL方言之间的转换。在VEGA-SQL查询模式下，系统会将SQL转换为对应数据源的方言SQL，然后使用对应数据源的驱动直接执行查询。
- github地址：https://github.com/tobymao/sqlglot

## 请求信息

| 项目     | 内容                                        |
|--------|-------------------------------------------|
| 请求方法   | POST                                      |
| 外部请求路径 | `/api/vega-backend/v1/resources/query`    |
| 内部请求路径 | `/api/vega-backend/in/v1/resources/query` |
| 内容类型   | `application/json`                        |

### 请求头

| 参数名 | 类型 | 必填 | 说明           |
|--------|------|------|--------------|
| Authorization | string | 是 | Bearer token |

## 查询参数说明

该接口不使用URL查询参数，所有查询参数通过请求体传递。

## 请求体

| 字段名           | 类型     | 必填 | 说明                                                                                                   |
|---------------|--------|----|------------------------------------------------------------------------------------------------------|
| query         | any    | 否  | 查询语句。流式查询获取首页数据必填，获取后续数据不用填。query与query_id不可同时存在。                                                    |
| resource_type | string | 是  | 数据源类型，必选值：`opensearch`、`mysql`、`mariadb`、`postgresql`。                                               |
| resource_id   | string | 否  | 查询opensearch必填参数                                                                                     |
| query_type    | string | 否  | 查询类型，只在SQL类查询中有效，可选值：`standard`（标准查询，最多返回10000条数据，如需获取更多数据，请使用stream 流式查询）、`stream`（流式查询）。不填则默认为同步查询 |
| query_id      | string | 否  | SQL类流式查询必填参数，首次请求不需要，从第二次请求开始往后每次请求都必须填写。query与query_id不可同时存在。                                       |
| query_timeout | int    | 否  | 单位：秒。超时后可以选择是否继续等待，如果继续等待，根据响应中的query_id发起下一次请求。                                                     |
| stream_size   | int    | 否  | 流式查询指定每批次返回的记录数，默认值10000，取值范围100-10000。                                                              |

### resource_type 说明（后续会按计划支持其他方言查询）

| 值          | 说明           | query格式 |
|------------|--------------|-----------|
| 不填         | VEGA-查询      | 标准SQL语句 |
| mysql      | MySQL查询      | MySQL语法SQL语句 |
| mariadb    | MariaDB查询    | MariaDB语法SQL语句 |
| postgresql | PostgreSQL查询 | PostgreSQL语法SQL语句 |
| opensearch | OpenSearch查询 | OpenSearch DSL语句 |

## 响应体

| 字段名                | 类型       | 说明                               |
|--------------------|----------|----------------------------------|
| columns            | array    | 结果集列定义                           |
| columns[].name     | string   | 列名                               |
| columns[].type     | string   | 列数据类型（如：integer、string、boolean等） |
| entries            | array    | 查询结果数据行                          |
| entries[]          | object   | 单行数据，键为列名，值为对应数据                 |
| total_count        | integer  | 记录数                              |
| stats              | object   | 查询状态信息                           |
| stats.has_more     | boolearn | 标识SQL类查询是否还有更多数据                 |
| stats.is_timeout   | boolearn | 标识SQL类查询是否超时                     |
| stats.query_id     | string   | 仅SQL类流式查询有此字段                    |
| stats.search_after | array    | 仅OpenSearch流式查询有此字段              |
| stats.offset       | integer  | 标识当前游标位置                         |

## 请求示例

### 示例1：VEGA-SQL同步查询

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "mariadb",
    "query": "select f_id,f_name from {{d74tbh60r29jmv6a6l70}} where f_id = 'd74tbh60r29jmv6a6l7g'"
  }'
```

### 响应示例

```json
{
  "columns": [
    {
      "name": "f_id",
      "type": "string"
    },
    {
      "name": "f_name",
      "type": "string"
    }
  ],
  "entries": [
    {
      "f_id": "d74tbh60r29jmv6a6l7g",
      "f_name": "adp.t_resource_schema_history"
    }
  ],
  "stats": {
    "is_timeout": false,
    "has_more": false,
    "offset": 0
  },
  "total_count": 1
}
```

### 示例2：VEGA-SQL流式查询

#### 首次请求

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "mariadb",
    "query_type": "stream",
    "stream_size": 5,
    "query": "select f_id,f_name from {{d74tbh60r29jmv6a6l70}}"
  }'
```

### 响应示例

```json
{
  "columns": [
    {
      "name": "f_id",
      "type": "string"
    },
    {
      "name": "f_name",
      "type": "string"
    }
  ],
  "entries": [
    {
      "f_id": "d74tbh60r29jmv6a6l60",
      "f_name": "adp.t_catalog"
    },
    {
      "f_id": "d74tbh60r29jmv6a6l6g",
      "f_name": "adp.t_catalog_discover_policy"
    },
    {
      "f_id": "d74tbh60r29jmv6a6l80",
      "f_name": "adp.t_connector_type"
    },
    {
      "f_id": "d74tbh60r29jmv6a6l8g",
      "f_name": "adp.t_discover_task"
    },
    {
      "f_id": "d74tbh60r29jmv6a6l70",
      "f_name": "adp.t_resource"
    }
  ],
  "stats": {
    "is_timeout": false,
    "query_id": "1a85b1eb-6d71-49a4-90c6-679eef0b5993",
    "has_more": true,
    "offset": 0
  },
  "total_count": 5
}
```

#### 第二次及后续请求

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "mariadb",
    "query_type": "stream",
    "stream_size": 5,
    "query_id": "1a85b1eb-6d71-49a4-90c6-679eef0b5993"
  }'
```

### 响应示例

```json
{
  "columns": [
    {
      "name": "f_id",
      "type": "string"
    },
    {
      "name": "f_name",
      "type": "string"
    }
  ],
  "entries": [
    {
      "f_name": "adp.t_resource_schema_history",
      "f_id": "d74tbh60r29jmv6a6l7g"
    },
    {
      "f_name": "opensearch_dashboards_sample_data_ecommerce",
      "f_id": "d750k3m0r29ju43hcgkg"
    },
    {
      "f_name": "security-auditlog-2026.03.17",
      "f_id": "d750k3e0r29ju43hcgi0"
    },
    {
      "f_id": "d750k3e0r29ju43hcgig",
      "f_name": "security-auditlog-2026.03.18"
    },
    {
      "f_id": "d750k3e0r29ju43hcgj0",
      "f_name": "security-auditlog-2026.03.19"
    }
  ],
  "stats": {
    "is_timeout": false,
    "query_id": "1a85b1eb-6d71-49a4-90c6-679eef0b5993",
    "has_more": true,
    "offset": 5
  },
  "total_count": 5
}
```

### 示例3：OpenSearch DSL同步查询

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "opensearch",
    "query": {
      "resource_id": "d750k3e0r29ju43hcgk0",
      "query": {
        "match_all": {}
      },
      "size": 10
    }
  }'
```

### 响应示例

```json
{
  "columns": [
    {
      "name": "id",
      "type": "string"
    },
    {
      "name": "name",
      "type": "string"
    },
    {
      "name": "age",
      "type": "string"
    }
  ],
  "entries": [
    {
      "name": "张三",
      "age": 25,
      "id": 1
    },
    {
      "id": 2,
      "name": "李四",
      "age": 30
    },
    {
      "id": 3,
      "name": "王五",
      "age": 28
    }
  ],
  "stats": {
    "is_timeout": false,
    "has_more": false
  },
  "total_count": 3
}
```

### 示例4：OpenSearch DSL流式查询

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "opensearch",
    "query": {
      "resource_id": "d750k3e0r29ju43hcgk0",
      "query": {
        "match_all": {}
      },
      "sort": [
        {"id": "asc"}
      ],
      "size": 2
    }
  }'
```

### 响应示例

```json
{
  "columns": [
    {
      "name": "id",
      "type": "string"
    },
    {
      "name": "name",
      "type": "string"
    },
    {
      "name": "age",
      "type": "string"
    }
  ],
  "entries": [
    {
      "id": 1,
      "name": "张三",
      "age": 25
    },
    {
      "id": 2,
      "name": "李四",
      "age": 30
    }
  ],
  "stats": {
    "is_timeout": false,
    "has_more": true,
    "search_after": [
      2
    ]
  },
  "total_count": 2
}
```

### 示例5：OpenSearch DSL流式查询 第二页及后续请求

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "opensearch",
    "query": {
      "resource_id": "d750k3e0r29ju43hcgk0",
      "query": {
        "match_all": {}
      },
      "sort": [
        {"id": "asc"}
      ],
      "search_after": [2],
      "size": 2
    }
  }'
```

### 响应示例

```json
{
  "columns": [
    {
      "name": "age",
      "type": "string"
    },
    {
      "name": "id",
      "type": "string"
    },
    {
      "name": "name",
      "type": "string"
    }
  ],
  "entries": [
    {
      "id": 3,
      "name": "王五",
      "age": 28
    }
  ],
  "stats": {
    "is_timeout": false,
    "has_more": false,
    "search_after": [
      3
    ]
  },
  "total_count": 1
}
```

### 示例6：SQL类查询超时示例

#### 首次请求

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "mariadb",
    "query_timeout": 5,
    "query": "select * from resource_id offset 0 limit 10"
}'
```

#### 响应示例

```json
{
  "columns": [],
  "entries": [],
  "stats": {
    "is_timeout": true,
    "query_id": "query_id_xxx",
    "has_more": true,
    "offset": 0
  },
  "total_count": 0
}
```

#### 若继续等待，则用首次请求的query_id进行后续流式查询。

```bash
curl -X POST "https://your-domain/api/vega-backend/v1/resources/query" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token-here" \
  -d '{
    "resource_type": "mariadb",
    "query_timeout": 5,
    "query_id": "query_id_xxx"
}'
```

## 错误码说明

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权访问 |
| 403 | 无权限访问该资源 |
| 500 | 服务器内部错误 |

## 注意事项

1. 查询语句建议添加适当的限制条件（如LIMIT），避免返回过多数据
2. OpenSearch DSL查询时，query字段需要传入完整的JSON对象
3. 不同数据源的数据类型映射可能存在差异，请参考columns中的type字段进行数据解析
