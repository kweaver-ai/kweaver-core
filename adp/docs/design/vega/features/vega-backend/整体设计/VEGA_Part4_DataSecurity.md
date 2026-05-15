# Part 4: Data Security (数据安全)

## 12. 安全与权限设计 (Security & Access Control)

### 12.1. 认证机制 (Authentication)
```
┌─────────────────────────────────────────────────────────────────┐
│                    Authentication Flow                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  支持多种认证方式:                                                │
│  ├── API Key: 适用于服务间调用、CI/CD 流水线                       │
│  ├── JWT Token: 适用于用户登录、前端应用                          │
│  ├── OAuth 2.0 / OIDC: 集成企业 SSO (Okta, Azure AD)             │
│  └── mTLS: 适用于高安全要求的内部服务通信                          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 12.2. 授权模型 (Authorization - RBAC)
采用 **资源-操作-角色** 三层授权模型：

| 层级 | 资源 (Resource) | 操作 (Action) | 示例角色 |
| :--- | :--- | :--- | :--- |
| **Catalog** | 整个目录 | admin, read | Catalog Admin |
| **Connection** | 数据源连接 | create, read, update, delete | Connection Manager |
| **Asset** | 单个资产 | read, write, sync, delete | Data Analyst, Data Engineer |
| **Row/Column** | 行/列级别 | read (with filter) | Restricted Viewer |

**权限配置示例**:
```yaml
roles:
  - name: data_analyst
    permissions:
      - resource: "catalog:production/*"
        actions: ["read", "query"]
      - resource: "catalog:production/sales.*"
        actions: ["read", "query", "export"]

  - name: data_engineer
    permissions:
      - resource: "catalog:*"
        actions: ["read", "query", "sync", "create_view"]
      - resource: "connection:*"
        actions: ["read"]
```

### 12.3. 数据安全 (Data Security)

| 安全能力 | 实现方式 | 适用场景 |
| :--- | :--- | :--- |
| **列级脱敏** | 动态替换 (Query Rewrite) | PII 字段 (手机号、身份证) |
| **行级过滤** | 自动注入 WHERE 条件 | 多租户数据隔离 |
| **传输加密** | TLS 1.3 | 所有 API 通信 |
| **存储加密** | AES-256 (OpenSearch 服务端加密) | 敏感 Asset |
| **审计日志** | 全量查询日志 + 敏感操作告警 | 合规审计 |

**脱敏规则示例**:
```yaml
masking_rules:
  - field_pattern: "*.phone"
    strategy: partial_mask     # 138****5678
  - field_pattern: "*.id_card"
    strategy: hash             # SHA256 后取前16位
  - field_pattern: "*.salary"
    strategy: range_bucket     # 10000-20000
    roles_exempt: ["hr_admin"] # 豁免角色
```

---


## 13. 资源隔离与调度 (Resource Isolation & Scheduling)

### 13.1. 资源隔离模型
```
┌─────────────────────────────────────────────────────────────────┐
│                   Resource Isolation Model                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Resource Pool                         │    │
│  │                                                          │    │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐    │    │
│  │  │ Pool: Default│ │ Pool: Premium│ │ Pool: Batch  │    │    │
│  │  │ CPU: 4 cores │ │ CPU: 16 cores│ │ CPU: 8 cores │    │    │
│  │  │ Mem: 8GB     │ │ Mem: 32GB    │ │ Mem: 16GB    │    │    │
│  │  │ QPS: 100     │ │ QPS: 1000    │ │ QPS: 50      │    │    │
│  │  │ Priority: 5  │ │ Priority: 10 │ │ Priority: 1  │    │    │
│  │  └──────────────┘ └──────────────┘ └──────────────┘    │    │
│  │        ▲                ▲                ▲              │    │
│  │        │                │                │              │    │
│  │  ┌─────┴────┐    ┌─────┴────┐    ┌─────┴────┐         │    │
│  │  │ Tenant A │    │ Tenant B │    │ ETL Jobs │         │    │
│  │  └──────────┘    └──────────┘    └──────────┘         │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 13.2. 查询队列管理
```yaml
query_queues:
  # 交互式查询队列
  interactive:
    max_concurrent: 50
    max_queue_size: 200
    timeout: 30s
    priority: high
    resource_pool: premium

  # 批量查询队列
  batch:
    max_concurrent: 10
    max_queue_size: 1000
    timeout: 10m
    priority: low
    resource_pool: batch
    scheduling: fair  # fifo | fair | priority

  # 同步任务队列
  sync:
    max_concurrent: 5
    max_queue_size: 100
    timeout: 1h
    priority: medium
    resource_pool: batch
```

### 13.3. 公平调度算法
```
┌─────────────────────────────────────────────────────────────────┐
│                    Fair Scheduler Algorithm                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  调度公式: Score = (Priority × Weight) / (Running + 1)          │
│                                                                  │
│  示例:                                                           │
│  ├── Tenant A: Priority=10, Running=5 → Score = 10×1/(5+1) = 1.67│
│  ├── Tenant B: Priority=10, Running=2 → Score = 10×1/(2+1) = 3.33│
│  └── 结果: Tenant B 的下一个查询优先执行                          │
│                                                                  │
│  防饥饿机制:                                                      │
│  ├── 最大等待时间: 60s 后提升优先级                               │
│  ├── 保底配额: 每租户至少 1 个并发槽位                            │
│  └── 突发容忍: 短时允许超配额 20%                                 │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 13.4. 租户配额配置
```yaml
tenant_quotas:
  - tenant_id: "tenant_a"
    quotas:
      max_qps: 500
      max_concurrent_queries: 20
      max_scan_bytes_per_query: 10GB
      max_result_rows: 100000
      daily_query_limit: 10000
      storage_quota: 100GB
    resource_pool: premium

  - tenant_id: "tenant_b"
    quotas:
      max_qps: 100
      max_concurrent_queries: 5
      max_scan_bytes_per_query: 1GB
      max_result_rows: 10000
      daily_query_limit: 1000
      storage_quota: 10GB
    resource_pool: default
```

### 13.5. 资源监控与限流
```yaml
rate_limiting:
  global:
    max_qps: 5000
    max_concurrent: 200

  per_tenant:
    algorithm: token_bucket
    bucket_size: 100
    refill_rate: 50/s

  per_asset:
    algorithm: sliding_window
    window_size: 1m
    max_requests: 1000

  circuit_breaker:
    failure_threshold: 50%
    window_size: 10s
    recovery_timeout: 30s
    half_open_requests: 5
```

---


## 14. 多租户隔离 (Multi-Tenancy Isolation)

### 14.1. 隔离级别
```
┌─────────────────────────────────────────────────────────────────┐
│                    Multi-Tenancy Isolation Levels                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Level 1: 逻辑隔离 (Logical Isolation) - 默认                    │
│  ├── 同一数据库，tenant_id 字段区分                              │
│  ├── 查询自动注入 WHERE tenant_id = ?                            │
│  └── 适用: 中小租户，成本敏感                                     │
│                                                                  │
│  Level 2: Schema 隔离 (Schema Isolation)                         │
│  ├── 同一数据库，不同 Schema/Namespace                           │
│  ├── 元数据: vega.tenant_a.*, vega.tenant_b.*                   │
│  └── 适用: 需要更强隔离的租户                                     │
│                                                                  │
│  Level 3: 存储隔离 (Storage Isolation)                           │
│  ├── 独立 OpenSearch Index 前缀                                  │
│  ├── 路径: index_tenant_a_*, index_tenant_b_*                   │
│  └── 适用: 数据敏感、合规要求高的租户                              │
│                                                                  │
│  Level 4: 完全隔离 (Full Isolation)                              │
│  ├── 独立计算资源池 + 独立存储                                    │
│  ├── 可部署独立实例                                               │
│  └── 适用: 企业大客户、政府/金融客户                               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 14.2. 租户配置
```yaml
tenants:
  - id: "tenant_a"
    name: "Acme Corp"
    isolation_level: "logical"

    # 存储配置
    storage:
      prefix: "tenant_a_"

    # 资源配额
    quotas:
      max_assets: 100
      max_storage_gb: 50
      max_qps: 200
      max_concurrent_queries: 10
      max_sync_jobs: 3

    # 网络隔离 (可选)
    network:
      allowed_source_ips: ["10.0.0.0/8"]

  - id: "tenant_b"
    name: "Enterprise Inc"
    isolation_level: "storage"

    storage:
      # 独立 OpenSearch 索引前缀
      prefix: "tenant_b_"                 # 独立前缀
      encryption_key_id: "key_tenant_b"   # 独立加密密钥

    quotas:
      max_assets: 1000
      max_storage_gb: 500
      max_qps: 2000
      max_concurrent_queries: 50

    resource_pool: "premium"  # 独立资源池
```

### 14.3. 数据访问隔离
```yaml
data_isolation:
  # 自动注入租户过滤
  auto_tenant_filter:
    enabled: true
    tenant_field: "tenant_id"
    inject_on: ["select", "update", "delete"]

  # 跨租户访问控制
  cross_tenant_access:
    enabled: false          # 默认禁止
    allowed_scenarios:
      - type: "admin_audit"
        roles: ["super_admin"]
      - type: "shared_asset"
        requires: "explicit_grant"

  # 租户数据泄露防护
  leak_prevention:
    # 查询结果检查
    validate_result_tenant: true
    # 日志脱敏
    mask_cross_tenant_data_in_logs: true
    # 告警
    alert_on_cross_tenant_attempt: true
```

### 14.4. 计算资源隔离
```yaml
compute_isolation:
  # DuckDB 实例隔离
  duckdb:
    mode: "per_tenant_instance"  # shared | per_tenant_instance
    memory_limit_per_tenant: 2GB

  # 连接池隔离
  connection_pools:
    mode: "per_tenant"
    max_connections_per_tenant: 20

  # Worker 隔离
  workers:
    mode: "shared_with_priority"  # shared | dedicated
    tenant_priorities:
      tenant_a: 5
      tenant_b: 10
```

### 14.5. 审计与合规
```yaml
tenant_audit:
  # 操作审计
  log_all_operations: true
  log_retention_days: 365

  # 合规报告
  compliance_reports:
    - type: "data_access_report"
      schedule: "monthly"
      include: ["queries", "exports", "access_patterns"]

    - type: "data_residency_report"
      schedule: "weekly"
      verify: "all_data_in_region"

  # 数据删除 (GDPR)
  data_deletion:
    tenant_offboarding:
      grace_period: 30d
      backup_retention: 90d
      complete_deletion: true
```

---


---

