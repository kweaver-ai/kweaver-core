# Part 3: Data Quality (数据质量)

## 10. Schema 演进管理 (Schema Evolution)

### 10.1. Schema Registry 架构
```
┌─────────────────────────────────────────────────────────────────┐
│                      Schema Registry                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐       │
│  │ Schema Store │◄───│ Change       │◄───│ Source       │       │
│  │ (PostgreSQL) │    │ Detector     │    │ Connectors   │       │
│  └──────────────┘    └──────────────┘    └──────────────┘       │
│         │                   │                                    │
│         ▼                   ▼                                    │
│  ┌──────────────┐    ┌──────────────┐                           │
│  │ Version      │    │ Compatibility│                           │
│  │ History      │    │ Checker      │                           │
│  └──────────────┘    └──────────────┘                           │
│         │                   │                                    │
│         ▼                   ▼                                    │
│  ┌──────────────────────────────────────────────────────┐       │
│  │              Migration Executor                       │       │
│  │  ├── Auto Migration (兼容变更)                        │       │
│  │  ├── Manual Review (破坏性变更)                       │       │
│  │  └── Rollback Handler (回滚)                          │       │
│  └──────────────────────────────────────────────────────┘       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 10.2. Schema 变更检测
```yaml
schema_detection:
  # 检测频率
  schedule: "*/5 * * * *"  # 每5分钟

  # 检测方式
  methods:
    mysql: "SHOW CREATE TABLE + 哈希比对"
    postgresql: "information_schema 查询"
    s3_parquet: "文件 Schema 采样"

  # 变更通知
  notifications:
    webhook: "https://hooks.example.com/schema-change"
    slack: "#data-platform-alerts"
```

### 10.3. 兼容性规则
| 变更类型 | 兼容性 | 自动处理 | 示例 |
| :--- | :--- | :--- | :--- |
| **新增列 (nullable)** | ✅ 向后兼容 | 自动同步 | `ALTER TABLE ADD COLUMN` |
| **新增列 (with default)** | ✅ 向后兼容 | 自动同步 | `ADD COLUMN x DEFAULT 0` |
| **删除列** | ⚠️ 向前兼容 | 标记废弃，延迟删除 | 保留30天后清理 |
| **列类型扩展** | ✅ 兼容 | 自动迁移 | `INT → BIGINT` |
| **列类型收缩** | ❌ 不兼容 | 人工审核 | `VARCHAR(100) → VARCHAR(50)` |
| **列重命名** | ❌ 不兼容 | 人工映射 | 需配置别名 |
| **主键变更** | ❌ 不兼容 | 触发全量重建 | 需人工确认 |

### 10.4. Schema 版本管理
```json
{
  "asset": "sales.orders",
  "schema_versions": [
    {
      "version": 3,
      "created_at": "2024-01-15T10:00:00Z",
      "fields": [
        {"name": "order_id", "type": "string", "nullable": false},
        {"name": "amount", "type": "decimal(10,2)", "nullable": false},
        {"name": "discount", "type": "decimal(5,2)", "nullable": true, "added_in": 3}
      ],
      "change_summary": "Added discount column",
      "compatible_with": [1, 2]
    }
  ],
  "active_version": 3,
  "deprecated_fields": [
    {"name": "old_status", "deprecated_in": 2, "remove_after": "2024-03-01"}
  ]
}
```

---


## 11. 数据质量管理 (Data Quality)

### 11.1. 质量检查框架
```
┌─────────────────────────────────────────────────────────────────┐
│                    Data Quality Framework                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │
│  │ Profiling   │  │ Validation  │  │ Monitoring  │              │
│  │ 数据画像     │  │ 规则校验     │  │ 持续监控    │              │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘              │
│         │                │                │                      │
│         ▼                ▼                ▼                      │
│  ┌──────────────────────────────────────────────────────┐       │
│  │                  Quality Score Engine                 │       │
│  │  综合评分 = Σ(维度权重 × 维度得分)                     │       │
│  └──────────────────────────────────────────────────────┘       │
│                          │                                       │
│         ┌────────────────┼────────────────┐                     │
│         ▼                ▼                ▼                     │
│  ┌───────────┐    ┌───────────┐    ┌───────────┐               │
│  │ Dashboard │    │ Alerts    │    │ Lineage   │               │
│  │ 质量报表   │    │ 异常告警   │    │ 血缘追踪   │               │
│  └───────────┘    └───────────┘    └───────────┘               │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 11.2. 质量维度与指标
| 维度 | 指标 | 计算方式 | 告警阈值 |
| :--- | :--- | :--- | :--- |
| **完整性** | null_rate | 空值行数 / 总行数 | > 5% |
| **完整性** | missing_rate | 缺失必填字段行数 / 总行数 | > 0% |
| **准确性** | format_valid_rate | 格式正确行数 / 总行数 | < 99% |
| **准确性** | range_valid_rate | 值域内行数 / 总行数 | < 99% |
| **一致性** | referential_integrity | 外键匹配行数 / 总行数 | < 100% |
| **一致性** | cross_source_match | 跨源数据匹配率 | < 99% |
| **时效性** | freshness_lag | 当前时间 - 最新数据时间 | > 1h |
| **唯一性** | duplicate_rate | 重复行数 / 总行数 | > 0% |

### 11.3. 校验规则配置
```yaml
quality_rules:
  asset: "catalog.sales.orders"
  rules:
    # 完整性检查
    - name: "order_id_not_null"
      type: not_null
      field: order_id
      severity: critical

    # 格式检查
    - name: "email_format"
      type: regex
      field: customer_email
      pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
      severity: warning

    # 值域检查
    - name: "amount_positive"
      type: range
      field: amount
      min: 0
      max: 10000000
      severity: error

    # 引用完整性
    - name: "customer_exists"
      type: foreign_key
      field: customer_id
      reference: "catalog.crm.customers.id"
      severity: error

    # 自定义 SQL 检查
    - name: "order_date_logic"
      type: sql
      query: "SELECT COUNT(*) FROM orders WHERE ship_date < order_date"
      expected: 0
      severity: critical

    # 聚合检查
    - name: "daily_order_count"
      type: anomaly_detection
      metric: "COUNT(*) GROUP BY DATE(created_at)"
      method: "zscore"
      threshold: 3
      severity: warning

  # 执行配置
  execution:
    schedule: "0 * * * *"  # 每小时
    on_sync: true          # 同步后执行
    sample_rate: 0.1       # 10% 采样 (大表)
```

### 11.4. 数据血缘追踪
```
┌─────────────────────────────────────────────────────────────────┐
│                      Data Lineage Graph                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [MySQL.users] ──┬──> [View.user_360] ──> [OpenSearch.user_vec] │
│                  │          ▲                                    │
│  [S3.events] ────┴──────────┘                                   │
│                                                                  │
│  血缘元数据:                                                      │
│  ├── 上游依赖 (Upstream): 数据来源                               │
│  ├── 下游影响 (Downstream): 影响范围                             │
│  ├── 转换逻辑 (Transformation): SQL / ETL 逻辑                  │
│  └── 质量传播 (Quality Propagation): 上游质量问题影响下游         │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

**血缘 API**:
```http
GET /api/v1/assets/{id}/lineage?direction=both&depth=3
```

---

---


---

