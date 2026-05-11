# Part 5: Operations & Reliability (运维与高可用)

## 15. 运维与监控 (Operations & Observability)

### 15.1. 监控指标 (Metrics)

| 指标类别 | 关键指标 | 告警阈值 |
| :--- | :--- | :--- |
| **查询性能** | `query_latency_p99`, `query_error_rate` | p99 > 5s, error_rate > 1% |
| **同步任务** | `sync_lag_seconds`, `sync_failure_count` | lag > 300s, failures > 3 |
| **资源使用** | `cpu_usage`, `memory_usage`, `disk_usage` | cpu > 80%, mem > 85%, disk > 90% |
| **连接池** | `connection_pool_active`, `connection_wait_time` | wait > 1s |
| **缓存** | `cache_hit_rate`, `cache_eviction_rate` | hit_rate < 70% |

### 15.2. 关键仪表板 (Dashboards)

```
┌─────────────────────────────────────────────────────────────────┐
│                    System Overview Dashboard                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [QPS Trend]     [Latency Distribution]    [Error Rate]         │
│  ┌──────────┐    ┌──────────────────┐      ┌──────────┐         │
│  │ ▄▄▄▄█▄▄▄ │    │ p50: 120ms       │      │ 0.05%    │         │
│  │          │    │ p95: 450ms       │      │ ──────── │         │
│  └──────────┘    │ p99: 1.2s        │      └──────────┘         │
│                  └──────────────────┘                           │
│                                                                  │
│  [Active Sync Jobs]  [Top Slow Queries]   [Resource Usage]      │
│  ┌──────────────┐    ┌────────────────┐   ┌──────────────┐      │
│  │ Running: 3   │    │ 1. sales.* 2.1s│   │ CPU:  65%    │      │
│  │ Pending: 12  │    │ 2. logs.* 1.8s │   │ MEM:  72%    │      │
│  │ Failed:  1   │    │ 3. user.* 1.5s │   │ Disk: 45%    │      │
│  └──────────────┘    └────────────────┘   └──────────────┘      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 15.3. 日志与追踪 (Logging & Tracing)

**结构化日志格式**:
```json
{
  "datetime": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "service": "query-engine",
  "trace_id": "abc123def456",
  "span_id": "span789",
  "user_id": "user_001",
  "message": "Query executed successfully",
  "context": {
    "asset": "sales.orders",
    "engine": "duckdb",
    "duration_ms": 145,
    "rows_returned": 1000
  }
}
```

**分布式追踪**: 使用 OpenTelemetry 标准，支持 Jaeger/Zipkin 后端，实现跨服务调用链追踪。

### 15.4. 运维操作手册

| 场景 | 操作 | 命令/API |
| :--- | :--- | :--- |
| **紧急熔断** | 禁用某数据源 | `PUT /api/v1/connections/{id}/disable` |
| **查询限流** | 设置 QPS 上限 | `PUT /api/v1/rate-limit` |
| **强制同步** | 触发全量同步 | `POST /api/v1/assets/{id}/sync?mode=full` |
| **清理缓存** | 清除查询缓存 | `DELETE /api/v1/cache` |
| **回滚版本** | Asset 版本回滚 | `POST /api/v1/assets/{id}/rollback?version=v3` |

---

## 16. 灾备与恢复 (Disaster Recovery)

### 16.1. 备份架构
```
┌─────────────────────────────────────────────────────────────────┐
│                    Backup Architecture                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                    Primary Region                        │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐              │    │
│  │  │ Metadata │  │ OpenSearch │  │ Cache    │              │    │
│  │  │ (PG)     │  │ Cluster    │  │ (Redis)  │              │    │
│  │  └────┬─────┘  └────┬─────┘  └──────────┘              │    │
│  └───────┼─────────────┼────────────────────────────────────┘    │
│          │             │                                         │
│          ▼             ▼                                         │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                 Backup Storage (S3/MinIO)                │    │
│  │  ├── /metadata/daily/2024-01-15.sql.gz                  │    │
│  │  ├── /opensearch/snapshots/2024-01-15/                  │    │
│  │  └── /opensearch/logs/                                  │    │
│  └─────────────────────────────────────────────────────────┘    │
│          │                                                       │
│          ▼ (异步复制)                                            │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │                   DR Region (Standby)                    │    │
│  │  ├── Metadata Replica (PG Streaming Replication)        │    │
│  │  └── OpenSearch Cross-Cluster Replication               │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 16.2. 备份策略
| 数据类型 | 备份方式 | 频率 | 保留期 | RTO | RPO |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **元数据 (PG)** | 逻辑备份 (pg_dump) | 每日 | 30天 | 1h | 24h |
| **元数据 (PG)** | WAL 归档 | 实时 | 7天 | 15min | 5min |
| **OpenSearch** | 快照 (Snapshot) | 每日 | 14天 | 2h | 24h |
| **OpenSearch** | 增量备份 (Delta) | 每小时 | 3天 | 30min | 1h |
| **配置文件** | Git 版本控制 | 每次变更 | 永久 | 5min | 0 |

### 16.3. 恢复流程
```yaml
recovery_procedures:
  # 场景1: 单 Asset 误删除
  asset_recovery:
    steps:
      - "1. 从备份列表选择恢复点: GET /api/v1/backups?asset={id}"
      - "2. 执行恢复: POST /api/v1/assets/{id}/restore?backup_id={bid}"
      - "3. 验证数据完整性: POST /api/v1/assets/{id}/validate"
    rto: 30min

  # 场景2: 元数据库故障
  metadata_recovery:
    steps:
      - "1. 启动备用 PG 实例"
      - "2. 恢复最近的基础备份"
      - "3. 应用 WAL 日志到故障点"
      - "4. 切换服务连接串"
    rto: 15min

  # 场景3: 区域级故障 (DR 切换)
  region_failover:
    steps:
      - "1. 确认主区域不可恢复"
      - "2. DNS 切换至 DR 区域"
      - "3. 提升 DR 区域为主区域"
      - "4. 通知用户数据截止时间"
    rto: 1h
    data_loss: "最多丢失1小时数据 (RPO)"
```

### 16.4. 备份 API
```http
# 创建手动备份
POST /api/v1/backups
{
  "scope": "asset",           # full | catalog | asset
  "asset_id": "ast_123",
  "description": "Before major migration"
}

# 列出可用备份
GET /api/v1/backups?asset_id=ast_123&from=2024-01-01

# 恢复备份
POST /api/v1/restore
{
  "backup_id": "bak_456",
  "target": "ast_123",
  "mode": "replace"           # replace | new_asset
}
```

---

## 17. 负载均衡与高可用 (Load Balancing & HA)

### 17.1. 整体架构
```
┌─────────────────────────────────────────────────────────────────┐
│                    High Availability Architecture                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│                      ┌──────────────┐                           │
│                      │   DNS/LB     │                           │
│                      │ (Cloud LB)   │                           │
│                      └──────┬───────┘                           │
│                             │                                    │
│              ┌──────────────┼──────────────┐                    │
│              ▼              ▼              ▼                    │
│       ┌──────────┐   ┌──────────┐   ┌──────────┐               │
│       │ Gateway  │   │ Gateway  │   │ Gateway  │               │
│       │ Node 1   │   │ Node 2   │   │ Node 3   │               │
│       └────┬─────┘   └────┬─────┘   └────┬─────┘               │
│            │              │              │                      │
│            └──────────────┼──────────────┘                      │
│                           │                                      │
│         ┌─────────────────┼─────────────────┐                   │
│         ▼                 ▼                 ▼                   │
│  ┌────────────┐    ┌────────────┐    ┌────────────┐            │
│  │ Worker     │    │ Worker     │    │ Worker     │            │
│  │ Pool A     │    │ Pool B     │    │ Pool C     │            │
│  │ (Query)    │    │ (Sync)     │    │ (CDC)      │            │
│  └────────────┘    └────────────┘    └────────────┘            │
│                                                                  │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │              Stateful Layer (HA Configured)              │    │
│  │  PostgreSQL    Redis Cluster    OpenSearch Cluster      │    │
│  │  (Primary +    (3+ nodes)       (3+ nodes)              │    │
│  │   Replica)                                               │    │
│  └─────────────────────────────────────────────────────────┘    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 17.2. Gateway 负载均衡
```yaml
load_balancer:
  algorithm: "least_connections"  # round_robin | least_connections | ip_hash

  health_check:
    path: "/health"
    interval: 5s
    timeout: 2s
    healthy_threshold: 2
    unhealthy_threshold: 3

  sticky_session:
    enabled: false  # 无状态设计，无需会话粘滞

  backends:
    - address: "gateway-1:8080"
      weight: 100
    - address: "gateway-2:8080"
      weight: 100
    - address: "gateway-3:8080"
      weight: 100
```

### 17.3. 服务发现与注册
```yaml
service_discover:
  provider: "consul"  # consul | etcd | kubernetes

  registration:
    service_name: "vega-gateway"
    health_check:
      http: "http://localhost:8080/health"
      interval: 10s
    tags:
      - "api"
      - "v2"

  discover:
    services:
      - name: "vega-worker"
        load_balance: "round_robin"
      - name: "embedding-service"
        load_balance: "least_connections"
```

### 17.4. 熔断器配置
```yaml
circuit_breaker:
  # 全局配置
  default:
    failure_rate_threshold: 50      # 失败率阈值
    slow_call_rate_threshold: 80    # 慢调用率阈值
    slow_call_duration: 5s          # 慢调用定义
    minimum_calls: 10               # 最小调用数
    sliding_window_size: 100        # 滑动窗口大小
    wait_duration_in_open: 30s      # 熔断等待时间
    permitted_calls_in_half_open: 5 # 半开状态允许调用数

  # 特定服务配置
  services:
    embedding_service:
      failure_rate_threshold: 30    # Embedding 服务更敏感
      fallback: "skip_embedding"    # 降级策略

    trino:
      failure_rate_threshold: 40
      fallback: "route_to_duckdb"   # 回退到 DuckDB
```

### 17.5. 故障转移策略
```
┌─────────────────────────────────────────────────────────────────┐
│                    Failover Strategies                           │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  组件故障          检测方式          故障转移动作                  │
│  ─────────────────────────────────────────────────────────────  │
│  Gateway 节点      健康检查失败      LB 自动摘除，流量分发到健康节点│
│  Worker 节点       心跳超时          任务重新分配到其他 Worker     │
│  PostgreSQL       复制延迟监控       自动提升 Replica 为 Primary  │
│  Redis 节点       Cluster 监控       Sentinel 自动主从切换        │
│  Trino           连接超时          熔断后回退到 DuckDB           │
│  Embedding API   错误率过高         切换到备用模型/本地模型        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

---

