# Catalogs API

## 目录
- [列出Catalogs](#列出catalogs)
- [创建Catalog](#创建catalog)
- [获取Catalogs](#获取catalogs)
- [更新Catalog](#更新catalog)
- [删除Catalogs](#删除catalogs)
- [获取Catalog健康状态](#获取catalog健康状态)
- [测试Catalog连接](#测试catalog连接)
- [发现Catalog资源](#发现catalog资源)
- [列出Catalog资源](#列出catalog资源)

---

## 列出Catalogs

### 接口描述
获取Catalog列表，支持按类型和健康状态过滤，支持分页查询。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs`
- **方法**: `GET`

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 查询参数
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| offset | int | 否 | 0 | 偏移量 |
| limit | int | 否 | 20 | 每页数量(最大100) |
| type | string | 否 | - | Catalog类型(physical/logical) |
| health_check_status | string | 否 | - | 健康状态(healthy/degraded/unhealthy/offline/disabled) |

### 请求体
无

### 响应
#### 成功响应 (200 OK)
```json
{
  "entries": [
    {
      "id": "catalog-id",
      "name": "catalog-name",
      "tags": ["tag1", "tag2"],
      "description": "catalog description",
      "type": "physical",
      "enabled": true,
      "connector_type": "mysql",
      "connector_config": {
        "host": "localhost",
        "port": 3306,
        "username": "root"
      },
      "metadata": {},
      "health_check_enabled": true,
      "health_check_status": "healthy",
      "last_check_time": 1234567890,
      "health_check_result": "connection successful",
      "creator": {
        "id": "account-id",
        "type": "account-type",
        "name": "account-name"
      },
      "create_time": 1234567890,
      "updater": {
        "id": "account-id",
        "type": "account-type",
        "name": "account-name"
      },
      "update_time": 1234567890
    }
  ],
  "total_count": 100
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| entries | array | Catalog列表 |
| total_count | int | 总数量 |

#### entries字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | Catalog ID |
| name | string | Catalog名称 |
| tags | array | 标签列表 |
| description | string | 描述 |
| type | string | Catalog类型(physical/logical) |
| enabled | bool | 是否启用 |
| connector_type | string | 连接器类型 |
| connector_config | object | 连接器配置(敏感字段已隐藏) |
| metadata | object | 元数据 |
| health_check_enabled | bool | 是否启用健康检查 |
| health_check_status | string | 健康状态 |
| last_check_time | int64 | 最后检查时间(Unix时间戳) |
| health_check_result | string | 健康检查结果 |
| creator | object | 创建者信息 |
| create_time | int64 | 创建时间(Unix时间戳) |
| updater | object | 更新者信息 |
| update_time | int64 | 更新时间(Unix时间戳) |

#### 错误响应
- **500 Internal Server Error**: 服务器内部错误

### 说明
- Catalog列表支持分页查询
- 默认每页返回20条记录
- 最大每页可返回100条记录
- connector_config中的敏感字段(如密码)会被隐藏
- Catalog类型包括：physical(物理Catalog)、logical(逻辑Catalog)
- 健康状态包括：healthy(健康)、degraded(降级)、unhealthy(不健康)、offline(离线)、disabled(已禁用)

---

## 创建Catalog

### 接口描述
创建新的Catalog，创建时会自动测试连接是否成功。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs`
- **方法**: `POST`
- **Content-Type**: `application/json`

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
```json
{
  "name": "catalog-name",
  "tags": ["tag1", "tag2"],
  "description": "catalog description",
  "connector_type": "mysql",
  "connector_config": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "encrypted-password",
    "databases": ["db1", "db2"],
    "options": {
      "charset": "utf8mb4"
    }
  }
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 是 | Catalog名称(唯一) |
| tags | array | 否 | 标签列表 |
| description | string | 否 | 描述 |
| connector_type | string | 否 | 连接器类型(为空则创建逻辑Catalog) |
| connector_config | object | 否 | 连接器配置 |

### 响应
#### 成功响应 (201 Created)
```json
{
  "id": "catalog-id"
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | 创建的Catalog ID |

#### 错误响应
- **400 Bad Request**: 请求参数错误
- **409 Conflict**: Catalog名称已存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- Catalog名称必须唯一
- connector_type为空时创建逻辑Catalog
- connector_type不为空时创建物理Catalog，需要提供connector_config
- 创建物理Catalog时会自动测试连接，连接失败则创建失败
- 敏感字段(如密码)需要使用RSA加密后传输，格式为加密后的Base64字符串
- 不同数据源类型的connector_config格式不同，详见下文"数据源类型配置示例"

---

## 获取Catalogs

### 接口描述
根据ID获取一个或多个Catalog的详细信息。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:ids`
- **方法**: `GET`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | string | 是 | Catalog ID，多个ID用逗号分隔 |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
无

### 响应
#### 成功响应 (200 OK)
```json
{
  "entries": [
    {
      "id": "catalog-id",
      "name": "catalog-name",
      "tags": ["tag1", "tag2"],
      "description": "catalog description",
      "type": "physical",
      "enabled": true,
      "connector_type": "mysql",
      "connector_config": {
        "host": "localhost",
        "port": 3306,
        "username": "root"
      },
      "metadata": {},
      "health_check_enabled": true,
      "health_check_status": "healthy",
      "last_check_time": 1234567890,
      "health_check_result": "connection successful",
      "creator": {
        "id": "account-id",
        "type": "account-type",
        "name": "account-name"
      },
      "create_time": 1234567890,
      "updater": {
        "id": "account-id",
        "type": "account-type",
        "name": "account-name"
      },
      "update_time": 1234567890
    }
  ]
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| entries | array | Catalog列表 |

#### entries字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | Catalog ID |
| name | string | Catalog名称 |
| tags | array | 标签列表 |
| description | string | 描述 |
| type | string | Catalog类型(physical/logical) |
| enabled | bool | 是否启用 |
| connector_type | string | 连接器类型 |
| connector_config | object | 连接器配置(敏感字段已隐藏) |
| metadata | object | 元数据 |
| health_check_enabled | bool | 是否启用健康检查 |
| health_check_status | string | 健康状态 |
| last_check_time | int64 | 最后检查时间(Unix时间戳) |
| health_check_result | string | 健康检查结果 |
| creator | object | 创建者信息 |
| create_time | int64 | 创建时间(Unix时间戳) |
| updater | object | 更新者信息 |
| update_time | int64 | 更新时间(Unix时间戳) |

#### 错误响应
- **404 Not Found**: Catalog不存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 支持批量获取多个Catalog，使用逗号分隔ID
- connector_config中的敏感字段(如密码)会被隐藏

---

## 更新Catalog

### 接口描述
更新指定Catalog的信息。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:id`
- **方法**: `PUT`
- **Content-Type**: `application/json`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Catalog ID |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
```json
{
  "name": "new-catalog-name",
  "tags": ["new-tag1", "new-tag2"],
  "description": "new catalog description",
  "connector_type": "mysql",
  "connector_config": {
    "host": "new-host",
    "port": 3307,
    "username": "new-user",
    "password": "new-encrypted-password",
    "databases": ["db3", "db4"],
    "options": {
      "charset": "utf8mb4"
    }
  }
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| name | string | 否 | Catalog名称(唯一) |
| tags | array | 否 | 标签列表 |
| description | string | 否 | 描述 |
| connector_type | string | 否 | 连接器类型(不可修改) |
| connector_config | object | 否 | 连接器配置 |

### 响应
#### 成功响应 (204 No Content)
无响应体

#### 错误响应
- **400 Bad Request**: 请求参数错误
- **404 Not Found**: Catalog不存在
- **409 Conflict**: Catalog名称已存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 不允许修改connector_type
- 修改connector_config时会自动测试连接，连接失败则更新失败
- 敏感字段(如密码)需要使用RSA加密后传输，格式为加密后的Base64字符串
- 不同数据源类型的connector_config格式不同，详见下文"数据源类型配置示例"

---

## 删除Catalogs

### 接口描述
删除一个或多个Catalog。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:ids`
- **方法**: `DELETE`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | string | 是 | Catalog ID，多个ID用逗号分隔 |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
无

### 响应
#### 成功响应 (204 No Content)
无响应体

#### 错误响应
- **404 Not Found**: Catalog不存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 支持批量删除多个Catalog，使用逗号分隔ID
- 删除Catalog会同时删除其关联的资源

---

## 获取Catalog健康状态

### 接口描述
获取指定Catalog的健康状态信息。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:id/health-status`
- **方法**: `GET`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | string | 是 | Catalog ID |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
无

### 响应
#### 成功响应 (200 OK)
```json
{
  "id": "catalog-id",
  "health_check_status": "healthy",
  "last_check_time": 1234567890,
  "health_check_result": "connection successful"
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | Catalog ID |
| health_check_status | string | 健康状态 |
| last_check_time | int64 | 最后检查时间(Unix时间戳) |
| health_check_result | string | 健康检查结果 |

#### 错误响应
- **404 Not Found**: Catalog不存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 健康状态包括：healthy(健康)、degraded(降级)、unhealthy(不健康)、offline(离线)、disabled(已禁用)

---

## 测试Catalog连接

### 接口描述
测试指定Catalog的连接是否正常。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:id/test-connection`
- **方法**: `POST`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Catalog ID |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
无

### 响应
#### 成功响应 (200 OK)
```json
{
  "health_check_status": "healthy",
  "last_check_time": 1234567890,
  "health_check_result": "connection successful"
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| health_check_status | string | 健康状态 |
| last_check_time | int64 | 最后检查时间(Unix时间戳) |
| health_check_result | string | 健康检查结果 |

#### 错误响应
- **404 Not Found**: Catalog不存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 测试连接会更新Catalog的健康状态
- 测试连接会更新最后检查时间
- 健康状态包括：healthy(健康)、degraded(降级)、unhealthy(不健康)、offline(离线)、disabled(已禁用)

---

## 发现Catalog资源

### 接口描述
触发指定Catalog的资源发现任务，该接口会创建一个异步发现任务，返回任务ID。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:id/discover`
- **方法**: `POST`
- **Content-Type**: `application/json`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| id | string | 是 | Catalog ID |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 请求体
无

### 响应
#### 成功响应 (200 OK)
```json
{
  "id": "task-id"
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | 发现任务ID |

#### 错误响应
- **404 Not Found**: Catalog不存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 此接口为异步操作，立即返回任务ID
- 发现任务执行状态可通过任务ID查询
- 任务状态包括：pending、running、completed、failed
- 任务进度可通过进度字段(0-100)查看

---

## 列出Catalog资源

### 接口描述
获取指定Catalog下的资源列表，支持分页查询。

### 请求信息
- **URL**: `/api/vega-backend/v1/catalogs/:ids/resources`
- **方法**: `GET`

### 路径参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| ids | string | 是 | Catalog ID |

### 请求头
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| x-account-id | string | 是 | 账户ID |
| x-account-type | string | 是 | 账户类型 |

### 查询参数
| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| offset | int | 否 | 0 | 偏移量 |
| limit | int | 否 | 20 | 每页数量(最大100) |

### 请求体
无

### 响应
#### 成功响应 (200 OK)
```json
{
  "entries": [
    {
      "id": "resource-id",
      "catalog_id": "catalog-id",
      "name": "resource-name",
      "tags": ["tag1", "tag2"],
      "description": "resource description",
      "category": "table",
      "status": "active",
      "status_message": "",
      "database": "database-name",
      "source_identifier": "table-name",
      "source_metadata": {},
      "schema_definition": [
        {
          "name": "column-name",
          "type": "string",
          "display_name": "Column Name",
          "original_name": "column_name",
          "description": "column description"
        }
      ],
      "creator": {
        "id": "account-id",
        "type": "account-type",
        "name": "account-name"
      },
      "create_time": 1234567890,
      "updater": {
        "id": "account-id",
        "type": "account-type",
        "name": "account-name"
      },
      "update_time": 1234567890
    }
  ],
  "total_count": 100
}
```

| 字段名 | 类型 | 说明 |
|--------|------|------|
| entries | array | 资源列表 |
| total_count | int | 总数量 |

#### entries字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | string | 资源ID |
| catalog_id | string | 所属Catalog ID |
| name | string | 资源名称 |
| tags | array | 标签列表 |
| description | string | 描述 |
| category | string | 资源类别(table/file/fileset/api/metric/topic/index/logicview/dataset) |
| status | string | 状态(active/disabled/deprecated/stale) |
| status_message | string | 状态消息 |
| database | string | 所属数据库(实例级Catalog时填充) |
| source_identifier | string | 源端标识(原始表名/路径) |
| source_metadata | object | 源端配置(JSON) |
| schema_definition | array | Schema定义 |
| creator | object | 创建者信息 |
| create_time | int64 | 创建时间(Unix时间戳) |
| updater | object | 更新者信息 |
| update_time | int64 | 更新时间(Unix时间戳) |

#### 错误响应
- **404 Not Found**: Catalog不存在
- **500 Internal Server Error**: 服务器内部错误

### 说明
- 资源列表支持分页查询
- 默认每页返回20条记录
- 最大每页可返回100条记录
- 资源类别包括：table、file、fileset、api、metric、topic、index、logicview、dataset
- 资源状态包括：active、disabled、deprecated、stale

---

## 数据源类型配置示例

### MySQL

```json
{
  "name": "mysql-catalog",
  "connector_type": "mysql",
  "connector_config": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "encrypted-password",
    "databases": ["db1", "db2"],
    "options": {
      "charset": "utf8mb4",
      "timeout": "10s"
    }
  }
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| host | string | 是 | MySQL服务器主机地址 |
| port | int | 是 | MySQL服务器端口(1-65535) |
| username | string | 是 | 数据库用户名 |
| password | string | 是 | 数据库密码(需加密) |
| databases | array | 否 | 数据库名称列表(为空则连接实例级别) |
| options | object | 否 | 连接参数(如charset, timeout等) |

### MariaDB

```json
{
  "name": "mariadb-catalog",
  "connector_type": "mariadb",
  "connector_config": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "encrypted-password",
    "databases": ["db1", "db2"],
    "options": {
      "charset": "utf8mb4",
      "timeout": "10s"
    }
  }
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| host | string | 是 | MariaDB服务器主机地址 |
| port | int | 是 | MariaDB服务器端口(1-65535) |
| username | string | 是 | 数据库用户名 |
| password | string | 是 | 数据库密码(需加密) |
| databases | array | 否 | 数据库名称列表(为空则连接实例级别) |
| options | object | 否 | 连接参数(如charset, timeout等) |

### Oracle

```json
{
  "name": "oracle-catalog",
  "connector_type": "oracle",
  "connector_config": {
    "host": "localhost",
    "port": 1521,
    "service_name": "ORCL",
    "username": "system",
    "password": "encrypted-password",
    "schemas": ["SCHEMA1", "SCHEMA2"],
    "options": {
      "connect_timeout": "10s"
    }
  }
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| host | string | 是 | Oracle服务器主机地址 |
| port | int | 是 | Oracle服务器端口(1-65535) |
| service_name | string | 是 | Oracle服务名 |
| username | string | 是 | 数据库用户名 |
| password | string | 是 | 数据库密码(需加密) |
| schemas | array | 否 | 模式名称列表(为空则连接实例级别) |
| options | object | 否 | 连接参数 |

### OpenSearch

```json
{
  "name": "opensearch-catalog",
  "connector_type": "opensearch",
  "connector_config": {
    "host": "localhost",
    "port": 9200,
    "username": "admin",
    "password": "encrypted-password",
    "index_pattern": "log-*"
  }
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| host | string | 是 | OpenSearch服务器主机地址 |
| port | int | 是 | OpenSearch服务器端口 |
| username | string | 否 | 认证用户名 |
| password | string | 否 | 认证密码(需加密) |
| index_pattern | string | 否 | 索引匹配模式(如log-*) |

### 逻辑Catalog

```json
{
  "name": "logical-catalog",
  "connector_type": "",
  "connector_config": {}
}
```

| 字段名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| connector_type | string | 否 | 为空表示创建逻辑Catalog |
| connector_config | object | 否 | 逻辑Catalog不需要连接器配置 |

### 说明
- 所有数据源类型的password字段都需要使用RSA加密后传输
- 加密格式为Base64编码的RSA加密字符串
- 数据库/模式列表为可选字段，为空时连接到实例级别
- options字段用于传递额外的连接参数，具体参数取决于数据源类型
- 系统会自动排除系统数据库/模式(如information_schema、sys等)
