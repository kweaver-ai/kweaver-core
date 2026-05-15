# Part 2: Core Data Mechanics (核心机制)

## 5. API 规范 (API Specification)


### 5.1. RESTful API 概览

> **设计决策**: 采用 **扁平化 (Flat)** 这一主路由模式，辅以 **层级化 (Hierarchical)** 浏览视图。
> *   Asset ID 全局唯一 (UUID)，因此 `/api/v1/assets/{id}` 为一等公民入口。
> *   同时提供 `/api/v1/catalogs/{id}/assets` 用于文件树式的层级浏览。

| 模块 | 端点 | 方法 | 描述 |
| :--- | :--- | :--- | :--- |
| **Catalog** | `/api/v1/catalogs` | GET, POST | 目录管理 (含连接配置) |
| **Catalog Detail** | `/api/v1/catalogs/{id}` | GET, PUT, DELETE | 目录详情操作 |
| **Catalog Status** | `/api/v1/catalogs/{id}/status` | GET | 获取连接状态 |
| **Catalog Status Summary** | `/api/v1/catalogs/status/summary` | GET | 所有连接状态概览 |
| **Catalog Connection** | `/api/v1/catalogs/{id}/test-connection` | POST | 测试连接 |
| **Catalog Disable/Enable** | `/api/v1/catalogs/{id}/disable` | POST | 禁用连接 |
| | `/api/v1/catalogs/{id}/enable` | POST | 启用连接 |
| **Catalog Inventory** | `/api/v1/catalogs/{id}/inventory/scan` | POST | 触发 Asset 发现扫描 |
| **Catalog Assets** | `/api/v1/catalogs/{id}/assets` | GET | **层级视图**: 列出指定 Catalog 下的 Assets |
| **Asset (Primary)** | `/api/v1/assets` | GET, POST | 资产管理 (扁平视图) |
| **Asset (Detail)** | `/api/v1/assets/{id}` | GET, PUT, DELETE | 资产详情操作 |
| **Asset Status** | `/api/v1/assets/{id}/disable` | POST | 禁用 Asset |
| | `/api/v1/assets/{id}/enable` | POST | 启用 Asset |
| **Asset Schema** | `/api/v1/assets/{id}/schema` | GET, PATCH | 获取/变更 Schema |
| **Query** | `/api/v1/query` | POST | 执行 DSL 查询 |
| **Sync** | `/api/v1/assets/{id}/sync` | POST | 触发物化同步 |
| **Dataset Records** | `/api/v1/assets/{id}/records` | POST | 批量写入记录 (Dataset) |
| | `/api/v1/assets/{id}/records/{rid}` | GET, PUT, DELETE | 单条记录操作 |
| | `/api/v1/assets/{id}/records/delete` | POST | 批量删除记录 |

### 5.2. 核心 API 详细定义

#### 5.2.1 执行查询
```http
POST /api/v1/query
Content-Type: application/json
Authorization: Bearer <token>

{
  "asset": "sales.orders",
  "operation": "select",
  "fields": ["order_id", "amount"],
  "filter": {"field": "status", "op": "eq", "value": "completed"},
  "limit": 100,
  "options": {
    "timeout": "30s",
    "prefer_local": true,
    "include_metadata": true
  }
}
```

**响应**:
```json
{
  "request_id": "req_abc123",
  "status": "success",
  "data": [
    {"order_id": "ORD001", "amount": 1500.00},
    {"order_id": "ORD002", "amount": 2300.00}
  ],
  "metadata": {
    "engine": "duckdb",
    "execution_time_ms": 45,
    "rows_scanned": 10000,
    "rows_returned": 100,
    "data_source": "opensearch_index",
    "freshness": "2024-01-15T10:30:00Z"
  }
}
```

#### 5.2.2 触发同步
```http
POST /api/v1/assets/ast_123/sync
Content-Type: application/json

{
  "mode": "incremental",        // full | incremental
  "priority": "normal",         // low | normal | high
  "options": {
    "generate_embeddings": true,
    "partition_by": "created_date"
  }
}
```

#### 5.2.3 为物理表开启 Sync (1:1 加速)
无需创建新 Asset，直接对现有的 Physical Asset 配置 Sync 策略：

```http
POST /api/v1/assets/mysql_prod.orders/config
{
  "sync_config": {
    "enabled": true,
    "mode": "cdc",          // 实时同步
    "retention": "30d"
  }
}
```

#### 5.2.4 创建逻辑视图并物化 (复杂加工)
先定义 VIEW，再开启 Sync：

```http
POST /api/v1/assets
Content-Type: application/json

{
  "name": "customer_360",
  "type": "view",
  "catalog": "production",
  "definition": {
    "operation": "join",
    "sources": ["mysql.customers", "s3.transactions"],
    "logic": "..."
  },
  "sync_config": {
    "enabled": true,
    "schedule": "0 */6 * * *",
    "mode": "incremental"
  },
  "embedding_config": {
    "enabled": true,
    "source_fields": ["description"],
    "model": "text-embedding-3-small"
  }
}
```

### 5.3. Dataset API (原生可写数据集)

Dataset 是系统第 8 种 Asset 类型，支持通过 API 直接创建和写入数据。

#### 5.3.1 创建 Dataset
```http
POST /api/v1/assets
Content-Type: application/json

{
  "name": "user_feedback",
  "type": "dataset",
  "catalog": "app_data",
  "schema": {
    "fields": [
      { "name": "id", "type": "string", "nullable": false, "primary_key": true },
      { "name": "user_id", "type": "string", "nullable": false },
      { "name": "content", "type": "text", "nullable": false },
      { "name": "rating", "type": "integer", "nullable": true },
      { "name": "tags", "type": "json", "nullable": true },
      { "name": "created_at", "type": "datetime", "default": "now()" }
    ]
  },
  "settings": {
    "shards": 3,
    "replicas": 1
  },
  "embedding_config": {
    "enabled": true,
    "source_fields": ["content"],
    "model": "text-embedding-3-small"
  }
}
```

**响应**:
```json
{
  "id": "ast_dataset_001",
  "name": "user_feedback",
  "type": "dataset",
  "catalog": "app_data",
  "status": "active",
  "created_at": "2024-01-15T10:00:00Z",
  "storage": {
    "index_name": "dataset_app_data_user_feedback",
    "doc_count": 0
  }
}
```

#### 5.3.2 批量写入记录
```http
POST /api/v1/assets/{id}/records
Content-Type: application/json

{
  "mode": "upsert",                    // insert | upsert | replace
  "records": [
    {
      "id": "fb_001",
      "user_id": "u123",
      "content": "产品体验很好，界面简洁",
      "rating": 5,
      "tags": ["positive", "ux"]
    },
    {
      "id": "fb_002",
      "user_id": "u456",
      "content": "希望增加数据导出功能",
      "rating": 3,
      "tags": ["feature_request"]
    }
  ],
  "options": {
    "validate": true,                  // 是否校验 Schema
    "generate_embeddings": true,       // 是否生成向量
    "on_error": "skip_row"             // skip_row | fail | log_only
  }
}
```

**响应**:
```json
{
  "status": "success",
  "summary": {
    "total": 2,
    "inserted": 1,
    "updated": 1,
    "failed": 0
  },
  "failed_records": []
}
```

**写入模式说明**:

| 模式 | 行为 | 适用场景 |
| :--- | :--- | :--- |
| `insert` | 仅插入，主键冲突时失败 | 确保数据不重复 |
| `upsert` | 插入或更新，按主键去重 | 增量同步、幂等写入 |
| `replace` | 先按条件删除，再插入 | 全量覆盖 |

#### 5.3.3 单条记录操作
```http
# 获取单条记录
GET /api/v1/assets/{id}/records/{record_id}

# 更新单条记录
PUT /api/v1/assets/{id}/records/{record_id}
Content-Type: application/json

{
  "content": "更新后的反馈内容",
  "rating": 4
}

# 删除单条记录
DELETE /api/v1/assets/{id}/records/{record_id}
```

#### 5.3.4 批量删除记录
```http
POST /api/v1/assets/{id}/records/delete
Content-Type: application/json

# 按 ID 列表删除
{
  "ids": ["fb_001", "fb_002", "fb_003"]
}

# 按条件删除
{
  "filter": {
    "and": [
      { "field": "created_at", "op": "lt", "value": "2024-01-01" },
      { "field": "rating", "op": "lte", "value": 2 }
    ]
  }
}
```

**响应**:
```json
{
  "status": "success",
  "deleted_count": 150
}
```

#### 5.3.5 Schema 变更
```http
PATCH /api/v1/assets/{id}/schema
Content-Type: application/json

{
  "add_fields": [
    { "name": "source", "type": "string", "nullable": true },
    { "name": "processed", "type": "boolean", "default": false }
  ]
}
```

> **注意**: MVP 阶段仅支持新增字段，不支持删除或修改现有字段类型。

#### 5.3.6 查询 Dataset (复用统一 DSL)

Dataset 的查询使用与其他 Asset 相同的 DSL，无需特殊语法：

```http
POST /api/v1/query
Content-Type: application/json

# 结构化查询
{
  "asset": "app_data.user_feedback",
  "operation": "select",
  "fields": ["id", "content", "rating", "created_at"],
  "filter": {
    "and": [
      { "field": "rating", "op": "gte", "value": 4 },
      { "field": "tags", "op": "contains", "value": "ux" }
    ]
  },
  "sort": [{ "field": "created_at", "order": "desc" }],
  "limit": 50
}

# 向量检索
{
  "asset": "app_data.user_feedback",
  "operation": "vector_search",
  "vector_field": "content_embedding",
  "query_text": "界面设计建议",
  "top_k": 10,
  "filter": { "field": "rating", "op": "gte", "value": 3 }
}
```

#### 5.3.7 Dataset 统计信息
```http
GET /api/v1/assets/{id}/stats

# 响应
{
  "doc_count": 125000,
  "storage_size_bytes": 52428800,
  "last_write_at": "2024-01-15T14:30:00Z",
  "schema_version": 2,
  "index_health": "green"
}
```

---

### 5.4. gRPC API (高性能场景)
对于大数据量流式传输，提供 gRPC 接口：

```protobuf
service QueryService {
  // 流式返回查询结果 (Arrow RecordBatch)
  rpc StreamQuery(QueryRequest) returns (stream ArrowRecordBatch);

  // 向量检索
  rpc VectorSearch(VectorSearchRequest) returns (VectorSearchResponse);
}

service SyncService {
  // 监听同步任务状态
  rpc WatchSyncJob(WatchRequest) returns (stream SyncStatus);
}
```

---

## 6. 缓存架构 (Caching Architecture)

### 6.1. 多级缓存策略
```
┌─────────────────────────────────────────────────────────────────┐
│                    Multi-Level Cache Architecture                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  L1: 进程内缓存 (Local Cache)                                    │
│  ├── 实现: Go sync.Map / BigCache                               │
│  ├── 容量: 每节点 512MB                                          │
│  ├── TTL: 60s                                                    │
│  └── 用途: Schema 缓存、热点查询结果、DSL 解析结果                  │
│                                                                  │
│  L2: 分布式缓存 (Redis Cluster)                                  │
│  ├── 容量: 可配置 (默认 8GB)                                      │
│  ├── TTL: 可配置 (默认 5min)                                      │
│  └── 用途: 查询结果缓存、Session 状态、分布式锁                     │
│                                                                  │
│  L3: 物化层缓存 (OpenSearch Index)                               │
│  ├── 特性: 持久化、支持向量检索                                    │
│  └── 用途: 热点数据物化、预计算聚合结果                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 6.2. 缓存键设计
```yaml
cache_keys:
  # 查询结果缓存
  query_result: "qr:{tenant}:{asset_hash}:{query_hash}:{version}"

  # Schema 缓存
  schema: "schema:{catalog}:{asset}:{version}"

  # 连接元数据
  catalog_conn: "catalog:{catalog_id}:conn_meta"

  # 统计信息
  stats: "stats:{asset}:{stat_type}"
```

### 6.3. 缓存失效策略
| 触发事件 | 失效范围 | 策略 |
| :--- | :--- | :--- |
| **Schema 变更** | 该 Asset 所有缓存 | 主动失效 + 广播通知 |
| **数据同步完成** | 该 Asset 查询缓存 | 版本号递增，旧缓存自动失效 |
| **手动刷新** | 指定范围 | API 触发 `DELETE /api/v1/cache` |
| **TTL 过期** | 单条缓存 | 被动失效 |
| **内存压力** | LRU 淘汰 | 按访问频率淘汰 |

### 6.4. 缓存配置示例
```yaml
cache:
  l1:
    enabled: true
    max_size: 512MB
    default_ttl: 60s
  l2:
    enabled: true
    redis:
      cluster:
        - redis-node-1:6379
        - redis-node-2:6379
        - redis-node-3:6379
      password: ${REDIS_PASSWORD}
    default_ttl: 5m
    max_memory: 8GB
  policies:
    # 不缓存的查询模式
    bypass_patterns:
      - "operation: vector_search"  # 向量搜索结果不缓存
      - "filter.*.op: gt"          # 范围查询不缓存
    # 强制缓存的资产
    force_cache_assets:
      - "catalog.dim.*"            # 维度表强制缓存
```

---

## 7. 事务与一致性 (Transaction & Consistency)

### 7.1. 一致性模型
```
┌─────────────────────────────────────────────────────────────────┐
│                    Consistency Model                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  本系统采用 "最终一致性 + 可选强一致性" 的混合模型:                  │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Virtual Mode (联邦查询)                                  │    │
│  │ ├── 单源查询: 继承源端隔离级别 (通常 Read Committed)      │    │
│  │ ├── 跨源 Join: Read Committed (各源独立快照，非全局快照)  │    │
│  │ └── 限制: 无跨源事务保证                                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │ Local Mode (OpenSearch)                                  │    │
│  │ ├── 读取: Snapshot Isolation (MVCC)                      │    │
│  │ ├── 写入: Serializable (单写入者)                        │    │
│  │ └── 时间旅行: 支持读取历史版本                            │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 7.2. 跨源 Join 一致性说明
```yaml
cross_source_join:
  consistency: "read_committed_per_source"

  # 用户须知
  caveats:
    - "各数据源在 Join 执行时刻各自取快照"
    - "不保证全局时间点一致性"
    - "如需强一致，请先物化到 Local View"

  # 一致性增强选项
  options:
    # 选项1: 使用物化快照
    materialize_before_join:
      enabled: true
      description: "Join 前自动触发增量同步，确保 OpenSearch 数据新鲜"
      max_staleness: 5m

    # 选项2: 时间戳过滤
    timestamp_filter:
      enabled: true
      description: "自动注入时间戳条件，限定数据范围"
      field_convention: ["updated_at", "modified_time", "_ts"]
```

### 7.3. 写入冲突处理
```yaml
write_conflict_resolution:
  # CDC 同步冲突
  cdc_conflicts:
    strategy: "last_writer_wins"
    conflict_key: ["source", "pk", "version"]

  # 并发同步任务
  concurrent_sync:
    prevention: "distributed_lock"
    lock_provider: "redis"
    lock_timeout: 30m

  # 手动写入 vs CDC
  manual_vs_cdc:
    priority: "cdc_wins"
    conflict_log: true
    alert_on_conflict: true
```

### 7.4. 数据版本与时间旅行
```http
# 查询历史版本
GET /api/v1/query
{
  "asset": "sales.orders",
  "operation": "select",
  "options": {
    "as_of": "2024-01-10T00:00:00Z",  # 时间点查询
    "version": 42                       # 或指定版本号
  }
}

# 查看版本历史
GET /api/v1/assets/{id}/versions?limit=10

# 版本对比
GET /api/v1/assets/{id}/diff?from_version=40&to_version=42
```

---

## 8. API 版本管理 (API Versioning)

### 8.1. 版本策略
```
┌─────────────────────────────────────────────────────────────────┐
│                    API Versioning Strategy                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  版本格式: /api/v{major}/...                                     │
│                                                                  │
│  版本生命周期:                                                    │
│  ┌──────────┬──────────┬──────────┬──────────┐                  │
│  │  Alpha   │   Beta   │   GA     │ Deprecated│                  │
│  │ (v2-alpha│ (v2-beta)│  (v2)    │  (v1)     │                  │
│  │  3个月)  │  (3个月) │ (无限期) │ (12个月)  │                  │
│  └──────────┴──────────┴──────────┴──────────┘                  │
│                                                                  │
│  兼容性承诺:                                                      │
│  ├── Major 版本: 可包含破坏性变更                                 │
│  ├── 同一 Major 内: 仅新增，不删除/修改                           │
│  └── 废弃字段: 标记 deprecated，保留至少 2 个 Major 版本          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 8.2. 版本协商
```yaml
version_negotiation:
  # 版本指定方式 (优先级从高到低)
  methods:
    - path: "/api/v2/query"           # URL 路径 (推荐)
    - header: "X-API-Version: 2"      # 请求头
    - query: "?api_version=2"         # 查询参数

  # 默认版本
  default: "v1"

  # 当前支持版本
  supported:
    - version: "v2"
      status: "ga"
      released: "2024-06-01"
    - version: "v1"
      status: "deprecated"
      released: "2024-01-01"
      sunset: "2025-01-01"
```

### 8.3. 废弃通知机制
```http
# 废弃版本的响应头
HTTP/1.1 200 OK
Deprecation: true
Sunset: Sat, 01 Jan 2025 00:00:00 GMT
Link: </api/v2/query>; rel="successor-version"
X-Deprecation-Notice: "API v1 将于 2025-01-01 下线，请迁移至 v2"
```

### 8.4. 变更日志
```yaml
changelog:
  v2:
    released: "2024-06-01"
    changes:
      breaking:
        - "移除 /api/v1/datasets 端点，使用 /api/v2/assets 替代"
        - "响应格式从 {data: []} 改为 {items: [], pagination: {}}"
      new:
        - "新增 /api/v2/lineage 血缘查询接口"
        - "查询结果支持 Arrow 格式返回"
      deprecated:
        - "filter.op='like' 废弃，使用 'contains' 替代"

  v1:
    released: "2024-01-01"
    sunset: "2025-01-01"
    migration_guide: "https://docs.example.com/migration/v1-to-v2"
```

---

## 9. 批量操作 (Bulk Operations)

### 9.1. 批量导入
```
┌─────────────────────────────────────────────────────────────────┐
│                    Bulk Import Pipeline                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  支持格式: CSV, JSON, Parquet, Arrow, NDJSON                     │
│                                                                  │
│  导入流程:                                                        │
│  ┌────────────┐    ┌────────────┐    ┌────────────┐            │
│  │ 1. Upload  │───>│ 2. Parse   │───>│ 3. Validate│            │
│  │ 分片上传    │    │ 流式解析    │    │ 质量检查   │            │
│  └────────────┘    └────────────┘    └────────────┘            │
│         │                                    │                   │
│         │         ┌────────────┐            │                   │
│         │         │ 4. Transform│<───────────┘                   │
│         │         │ 字段映射     │                               │
│         │         └─────┬──────┘                                │
│         │               │                                        │
│         ▼               ▼                                        │
│  ┌────────────────────────────────────────────────────┐         │
│  │ 5. Write to OpenSearch (Parallel + Batched)        │         │
│  │    ├── Batch Size: 10,000 rows                     │         │
│  │    ├── Parallelism: 4 writers                      │         │
│  │    └── Progress: Checkpoint every 100,000 rows     │         │
│  └────────────────────────────────────────────────────┘         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**批量导入 API**:
```http
# 步骤1: 初始化导入任务
POST /api/v1/assets/{id}/import
{
  "source": {
    "type": "s3",
    "path": "s3://bucket/data/large_file.parquet"
  },
  "options": {
    "format": "parquet",
    "mode": "append",           # append | overwrite | upsert
    "upsert_key": ["id"],       # upsert 模式的主键
    "batch_size": 10000,
    "parallelism": 4,
    "validate": true,
    "on_error": "skip_row"      # skip_row | fail | log_only
  },
  "mapping": {
    "field_renames": {"old_name": "new_name"},
    "type_casts": {"amount": "decimal(10,2)"}
  }
}

# 响应
{
  "job_id": "import_abc123",
  "status": "running",
  "progress_url": "/api/v1/jobs/import_abc123"
}

# 步骤2: 查询进度
GET /api/v1/jobs/import_abc123
{
  "status": "running",
  "progress": {
    "total_rows": 10000000,
    "processed_rows": 4500000,
    "failed_rows": 123,
    "percentage": 45,
    "eta_seconds": 300
  }
}
```

### 9.2. 批量导出
```http
# 导出到文件
POST /api/v1/assets/{id}/export
{
  "destination": {
    "type": "s3",
    "path": "s3://bucket/exports/",
    "filename_pattern": "orders_{date}_{part}.parquet"
  },
  "format": "parquet",          # csv | parquet | json | arrow
  "options": {
    "compression": "zstd",
    "partition_by": ["year", "month"],
    "max_file_size": "256MB",
    "include_schema": true
  },
  "filter": {
    "field": "created_at",
    "op": "gte",
    "value": "2024-01-01"
  }
}

# 流式导出 (大数据量)
GET /api/v1/assets/{id}/export/stream
Accept: application/vnd.apache.arrow.stream
X-Export-Format: arrow
X-Export-Filter: {"status": "completed"}
```

### 9.3. 批量更新/删除
```http
# 批量更新
POST /api/v1/assets/{id}/bulk-update
{
  "filter": {
    "field": "status",
    "op": "eq",
    "value": "pending"
  },
  "updates": {
    "status": "cancelled",
    "cancelled_at": "2024-01-15T10:00:00Z"
  },
  "options": {
    "limit": 100000,
    "dry_run": false
  }
}

# 批量删除
POST /api/v1/assets/{id}/bulk-delete
{
  "filter": {
    "and": [
      {"field": "created_at", "op": "lt", "value": "2023-01-01"},
      {"field": "status", "op": "eq", "value": "archived"}
    ]
  },
  "options": {
    "soft_delete": true,        # 软删除
    "archive_before_delete": true
  }
}
```

### 9.4. 批量操作限制与保护
```yaml
bulk_operations:
  limits:
    max_import_file_size: 50GB
    max_export_rows: 100000000
    max_bulk_update_rows: 1000000
    max_bulk_delete_rows: 10000000
    max_concurrent_imports: 5
    max_concurrent_exports: 10

  safety:
    require_filter_for_bulk_update: true
    require_filter_for_bulk_delete: true
    require_confirmation_for_overwrite: true
    auto_backup_before_bulk_delete: true

  throttling:
    import_rate_limit: "100MB/s"
    export_rate_limit: "200MB/s"
```

---

---

---

