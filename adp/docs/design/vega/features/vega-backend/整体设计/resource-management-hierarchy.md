# VEGA 资源管理层级关系

## 概述

本文档详细说明 VEGA Backend 项目中的资源管理层级关系，包括核心实体、它们之间的关系以及实际应用场景。

---

## 核心实体层级关系

```
┌─────────────────────────────────────────────────────────────┐
│                    VEGA 资源管理层级                          │
└─────────────────────────────────────────────────────────────┘

第 1 层：ConnectorType (连接器类型注册)
    ↓ 定义数据源类型和配置规范
    ├─ mysql (table 类别)
    ├─ opensearch (index 类别)
    ├─ kafka (topic 类别)
    └─ ... (可扩展)

第 2 层：Catalog (数据目录/数据源)
    ↓ 具体的数据源连接
    ├─ Physical Catalog (物理目录)
    │   ├─ MySQL 生产库
    │   ├─ OpenSearch 集群
    │   └─ Kafka 集群
    └─ Logical Catalog (逻辑目录)
        └─ 业务视图集合

第 3 层：Resource (数据资源)
    ↓ 具体的数据资产
    ├─ Table 资源: users, orders, products...
    ├─ Index 资源: logs_2024, metrics...
    └─ Topic 资源: user_events, orders_stream...
```

---

## 详细关系图

```
┌────────────────────────────────────────────────────────────────┐
│                    ConnectorType (连接器类型)                   │
│  ├─ Type: "local-table-mysql"                                  │
│  ├─ Category: "table"                                           │
│  ├─ Mode: "local"                                              │
│  └─ FieldConfig: {host, port, username, password, database}    │
└────────────────────────────────────────────────────────────────┘
                              │
                              │ 注册/引用
                              ↓
┌────────────────────────────────────────────────────────────────┐
│                    Catalog (数据目录)                           │
│  ├─ Name: "生产 MySQL 实例"                                     │
│  ├─ Type: "physical"                                           │
│  ├─ ConnectorType: "local-table-mysql"                         │
│  ├─ ConnectorConfig: {                                         │
│  │    host: "10.0.0.1",                                        │
│  │    port: 3306,                                              │
│  │    username: "ENC:xxx",  // 加密存储                         │
│  │    password: "ENC:yyy"                                      │
│  │  }                                                          │
│  └─ HealthStatus: "healthy"                                    │
└────────────────────────────────────────────────────────────────┘
                              │
                              │ 包含 (1:N)
                              ↓
┌────────────────────────────────────────────────────────────────┐
│                    Resource (数据资源)                          │
│  ├─ CatalogID: "catalog-123"                                   │
│  ├─ Name: "users"                                              │
│  ├─ Category: "table"                                          │
│  ├─ Database: "app_db"                                         │
│  ├─ SourceIdentifier: "app_db.users"                           │
│  └─ SchemaDefinition: [                                        │
│      {name: "id", type: "integer"},                            │
│      {name: "name", type: "string"},                           │
│      {name: "email", type: "string"}                           │
│    ]                                                            │
└────────────────────────────────────────────────────────────────┘
```

---

## 实际场景示例

### 场景 1：电商平台数据资产管理

```
ConnectorType (连接器类型库)
│
├─ local-table-mysql      (MySQL 表连接器)
├─ local-index-opensearch (OpenSearch 索引连接器)
└─ remote-topic-kafka     (Kafka Topic 连接器)

Catalog (数据源连接)
│
├─ "MySQL 主库" [Physical]
│  ├─ ConnectorType: local-table-mysql
│  ├─ ConnectorConfig: {host: "db-master", ...}
│  │
│  └─ 包含的 Resources:
│     ├─ Table: users
│     ├─ Table: orders
│     ├─ Table: products
│     └─ Table: transactions
│
├─ "OpenSearch 日志集群" [Physical]
│  ├─ ConnectorType: local-index-opensearch
│  ├─ ConnectorConfig: {endpoint: "http://os-cluster:9200"}
│  │
│  └─ 包含的 Resources:
│     ├─ Index: access_logs_2024
│     ├─ Index: error_logs_2024
│     └─ Index: metrics_2024
│
└─ "用户行为分析" [Logical]
   ├─ Type: logical (无 Connector)
   │
   └─ 包含的 Resources:
      ├─ LogicView: user_profile_view
      └─ Dataset: user_behavior_summary
```

---

## 关键关系说明

### 1. ConnectorType → Catalog (1:N)

**关系**：一个 ConnectorType 可以被多个 Catalog 引用

```go
type Catalog struct {
    ConnectorType   string         `json:"connector_type"`   // 引用 ConnectorType.Type
    ConnectorConfig map[string]any `json:"connector_config"` // 按 FieldConfig 填充
}
```

**示例**：
- `local-table-mysql` 可以连接到：
  - 开发库 (Catalog: "MySQL 开发环境")
  - 测试库 (Catalog: "MySQL 测试环境")
  - 生产库 (Catalog: "MySQL 生产环境")

### 2. Catalog → Resource (1:N)

**关系**：一个 Catalog 可以包含多个 Resource

```go
type Resource struct {
    CatalogID        string         `json:"catalog_id"`         // 所属 Catalog
    Category         string         `json:"category"`           // 资源类别
    Database         string         `json:"database,omitempty"` // 所属数据库
    SourceIdentifier string         `json:"source_identifier"`  // 源端标识
}
```

**示例**：
- MySQL 生产库 Catalog 可以包含：
  - `app_db.users` (Table)
  - `app_db.orders` (Table)
  - `log_db.access_logs` (Table)

### 3. Resource → Category (类型映射)

Physical Catalog 根据 ConnectorType.Category 创建对应 Category 的 Resource:

| ConnectorType | 可创建的 Resource Category |
|---------------|---------------------------|
| table-mysql | table |
| index-opensearch | index |
| topic-kafka | topic |
| file-s3 | file, fileset |
| metric-prometheus | metric |

---

## 数据库表关系

```sql
-- t_connector_type (连接器类型注册表)
CREATE TABLE t_connector_type (
    f_type VARCHAR(100) PRIMARY KEY,  -- local-table-mysql
    f_category VARCHAR(50),            -- table
    f_mode VARCHAR(20),                -- local | remote
    ...
);

-- t_catalog (数据目录表)
CREATE TABLE t_catalog (
    f_id VARCHAR(50) PRIMARY KEY,
    f_name VARCHAR(200),
    f_type VARCHAR(20),                -- physical | logical
    f_connector_type VARCHAR(100),    -- 外键引用 t_connector_type
    f_connector_config JSON,           -- 按 ConnectorType.FieldConfig 填充
    ...
    FOREIGN KEY (f_connector_type) REFERENCES t_connector_type(f_type)
);

-- t_resource (数据资源表)
CREATE TABLE t_resource (
    f_id VARCHAR(50) PRIMARY KEY,
    f_catalog_id VARCHAR(50),          -- 外键引用 t_catalog
    f_name VARCHAR(200),
    f_category VARCHAR(50),            -- table | index | topic ...
    f_database VARCHAR(100),           -- 所属数据库（可选）
    f_source_identifier VARCHAR(500),  -- 源端标识
    f_schema_definition JSON,          -- Schema 定义
    ...
    FOREIGN KEY (f_catalog_id) REFERENCES t_catalog(f_id)
);
```

---

## API 层级关系

```
GET /connector-types           # 1. 查询可用的连接器类型
                                ↓
POST /catalogs                # 2. 创建 Catalog（选择 ConnectorType）
{                              │
  "name": "生产 MySQL",        │
  "connector_type": "local-table-mysql",  │
  "connector_config": {...}    │
}                              ↓
GET /catalogs/:id/resources   # 3. 查看 Catalog 下的所有 Resource
                                ↓
POST /resources               # 4. 手动创建 Resource
{                              │
  "catalog_id": "xxx",         │
  "category": "table",         │
  "source_identifier": "db.users" │
}                              ↓
POST /catalogs/:id/discover   # 5. 自动发现 Resource（推荐）
                                # 根据 ConnectorType 扫描数据源
```

---

## 核心实体说明

### ConnectorType (连接器类型)

**定义位置**: `interfaces/connector_type.go`

**职责**：定义数据源连接器的类型和配置规范

**关键字段**：
- `Type`: 唯一标识 (如 `local-table-mysql`)
- `Category`: 资源类别 (table/index/topic/file...)
- `Mode`: 运行模式 (local/remote)
- `FieldConfig`: 配置字段定义 (host, port, username...)

**支持的模式**：
- **Local**: 连接器内置在 vega-backend 进程内
- **Remote**: 连接器作为独立服务运行，通过 HTTP 调用

### Catalog (数据目录)

**定义位置**: `interfaces/catalog.go`

**职责**：管理数据源连接和逻辑分组

**类型**：
- **Physical**: 物理数据源连接，需要 ConnectorType 和 ConnectorConfig
- **Logical**: 逻辑目录，用于组织跨数据源的虚拟视图

**关键字段**：
- `Type`: physical / logical
- `ConnectorType`: 引用的连接器类型
- `ConnectorConfig`: 加密存储的连接配置
- `HealthStatus`: 健康检查状态

**健康状态**：
- `healthy`: 连接正常
- `degraded`: 性能下降
- `unhealthy`: 连接失败
- `offline`: 不可用
- `disabled`: 已禁用

### Resource (数据资源)

**定义位置**: `interfaces/resource.go`

**职责**：管理具体的数据资产实体

**支持的类别**：
- **物理资源**: table, file, fileset, api, metric, topic, index
- **逻辑资源**: logicview, dataset

**关键字段**：
- `CatalogID`: 所属目录
- `Category`: 资源类别
- `Database`: 所属数据库（实例级 Catalog 时）
- `SourceIdentifier`: 源端标识（如 `db.table`, `path/to/file`）
- `SchemaDefinition`: Schema 定义（字段列表）
- `SourceMetadata`: 源端配置元数据

**状态**：
- `active`: 活跃
- `disabled`: 已禁用
- `deprecated`: 已弃用
- `stale`: 已过期（需要同步）

---

## 层级设计的优缺点

### ✅ 优点

1. **清晰的分层**
   - ConnectorType（类型定义）
   - Catalog（连接实例）
   - Resource（数据资产）

2. **连接器可扩展**
   - 新增数据源只需注册 ConnectorType
   - 支持 local/remote 两种模式

3. **配置复用**
   - 同一 ConnectorType 可创建多个 Catalog
   - 不同环境使用相同连接器类型

4. **逻辑物理分离**
   - Physical Catalog 管理真实数据源
   - Logical Catalog 组织跨数据源视图

### ⚠️ 潜在问题

1. **Catalog 概念过载**
   - Physical Catalog = 数据源连接
   - Logical Catalog = 命名空间/文件夹
   - 两者语义差异大，强行统一导致概念混淆

2. **Resource 的 CatalogID 依赖语义不清**
   - 物理资源：CatalogID 确实是数据源
   - 逻辑资源：CatalogID 只是逻辑分组
   - 关系不够清晰

3. **缺少独立的 Schema 层**
   - Resource 的 `SchemaDefinition` 只是 JSON 字段
   - 没有独立的 Schema 实体
   - Schema 变更历史在 `t_resource_schema_history` 表，但不是一等公民

4. **Database 字段语义模糊**
   - 仅在实例级 Catalog 时有意义
   - 其他情况下为空

---

## 建议的层级优化

### 方案 1: 引入 DataSource 和 Schema

```
当前结构:
ConnectorType → Catalog → Resource

优化建议:
ConnectorType → DataSource → Schema → Resource
     │              │          │         │
     │              │          │         └─ 数据资产实例
     │              │          └─ Schema 定义和版本
     │              └─ 连接配置和健康状态
     └─ 连接器类型注册
```

**改进点**：
- DataSource 重命名 Physical Catalog，语义更清晰
- Schema 独立成为实体，支持版本管理
- Resource 专注于数据资产实例

### 方案 2: 引入 Namespace 概念

```
ConnectorType
    ↓
DataSource (重命名 Physical Catalog)
    ↓
Asset (重命名 Resource)
    │
    └─ Namespace (Logical Catalog 改为纯分组概念)
```

**改进点**：
- DataSource 更准确表达"数据源"
- Asset 比 Resource 范围更广
- Namespace 作为纯分组，不混淆物理连接概念

---

## 实体关系总结表

| 实体 | 职责 | 基数关系 | 示例 |
|------|------|----------|------|
| **ConnectorType** | 连接器类型注册 | 1 → N Catalog | mysql, opensearch, kafka |
| **Catalog** | 数据源连接/逻辑分组 | 1 → N Resource | 生产库, 日志集群, 业务视图 |
| **Resource** | 具体数据资产 | N → 1 Catalog | users 表, logs 索引 |

**核心层级**: **类型定义** → **连接实例** → **数据资产**

---

## 相关文档

- [CLAUDE.md](../CLAUDE.md) - 项目架构和开发指南
- [VEGA 设计文档](../vega_plan/) - 详细的设计规划
- [API 文档](../README.md#api) - REST API 接口说明

---

**文档版本**: 1.0
**最后更新**: 2025-02-05
**维护者**: VEGA Team