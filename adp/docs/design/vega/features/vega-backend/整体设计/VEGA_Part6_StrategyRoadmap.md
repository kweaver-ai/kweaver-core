# Part 6: Strategy & Roadmap (战略与规划)

## 18. 技术风险与应对 (Technical Risks & Mitigations)

| 风险 | 影响 | 可能性 | 应对措施 |
| :--- | :--- | :--- | :--- |
| **OpenSearch 资源开销** | 内存与存储成本较高 | 中 | 支持冷热分离架构；非核心数据仅存 S3 不建索引 |
| **OpenSearch 成本** | 托管服务费用较高 | 中 | 支持自建集群；使用冷热分离架构降低成本 |
| **DuckDB 内存溢出** | 大查询导致 OOM | 中 | 设置查询内存上限；超限自动路由到 Trino |
| **CDC 延迟过高** | 数据一致性问题 | 低 | 监控 lag 指标；支持手动触发全量同步 |
| **跨源 Join 性能差** | 用户体验差 | 中 | 引导用户物化热点数据；查询优化提示 |
| **Embedding 成本高** | API 调用费用 | 中 | 支持本地模型；增量更新策略 |
| **单点故障** | 服务不可用 | 低 | 核心组件 HA 部署；优雅降级策略 |

### 18.1. 降级策略 (Graceful Degradation)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Degradation Hierarchy                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Level 0 (正常): 完整功能                                         │
│      ↓ Trino 不可用                                              │
│  Level 1: 禁用跨源 Join，提示用户使用单源查询                       │
│      ↓ Embedding Service 不可用                                   │
│  Level 2: 禁用向量检索，回退到关键词搜索                            │
│      ↓ CDC 延迟过高                                               │
│  Level 3: Local View 标记为 stale，查询强制走 Virtual              │
│      ↓ 源数据库不可用                                              │
│  Level 4: 返回 Local View 缓存数据 + 过期警告                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 19. 竞品深度对标 (Competitive Landscape)
**核心定位**: 相比于 Palantir Foundry 的"重型操作系统"，我们做的是"轻量级 AI 加速器"。

### 19.1. Data Connection vs. Catalog
*   **Foundry**: **Ingestion Agent**。必须把数据搬运进来才能用。
*   **本设计**: **Federation Gateway**。Virtual Mode 允许"连接即查询"，极大地降低了数据探索的门槛。

### 19.2. Dataset (Parquet) vs. Asset (OpenSearch)
这是**Spark 时代**与**AI 时代**的分水岭。

| 维度 | Palantir Foundry Dataset | 本设计 Asset (Local Mode) |
| :--- | :--- | :--- |
| **底层格式** | Parquet + Transaction Log | **Lucene Index** (OpenSearch) |
| **设计目标** | **大规模批处理 (Batch Scan)** | **AI 随机访问 (Search + Vector)** |
| **数据处理** | Java / Spark (Magritte) | **Golang** (Concurrent Workers) |
| **点查延迟** | 分钟级 (需启动 Spark Job) | **毫秒级** (倒排索引查找) |
| **生态位** | 数据仓库 / 决策系统 | **AI 知识库 / RAG 后端 / 训练数据源** |

### 19.3. 与 Apache Gravitino 对比
Apache Gravitino 是新兴的统一元数据管理平台，与本设计有相似的定位但侧重点不同：

| 维度 | 本设计 | Apache Gravitino |
| :--- | :--- | :--- |
| **核心定位** | AI 数据供给层 (Catalog + Compute + Vector) | 统一元数据管理 (Catalog Only) |
| **查询能力** | 内置 DuckDB/Trino 路由，直接执行查询 | 仅元数据，需外接 Trino/Spark |
| **向量检索** | OpenSearch 原生支持 | 不支持 |
| **物化能力** | 内置 Sync (源 → OpenSearch) | 无，纯元数据层 |
| **数据源支持** | 7 类 Asset (Table/Fileset/API/Metric/Topic/Index/View) | Hive/Iceberg/JDBC/Kafka (元数据注册) |
| **部署模式** | 单二进制，开箱即用 | Java 服务，需配合 Trino 等 |
| **适用场景** | AI 应用快速接入数据 | 企业数据资产统一治理 |

**总结**: Gravitino 专注于"元数据联邦"，是 Hive Metastore 的现代化替代；本设计则是"元数据 + 计算 + 向量化"的一体化方案，更适合 AI 场景的端到端数据供给。

### 19.4. 与其他方案对比

| 维度 | 本设计 | Trino/Presto | Databricks Unity Catalog | Palantir Foundry | Apache Gravitino |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **部署复杂度** | 低 (单二进制) | 中 (集群) | 高 (云绑定) | 高 (企业级) | 中 (Java 服务) |
| **查询执行** | 内置 (DuckDB/Trino) | 原生 | 需 Spark | 内置 | 需外接 |
| **向量检索** | 原生支持 | 不支持 | 需集成 | 不支持 | 不支持 |
| **实时数据** | CDC 原生 | 需外部 | 需外部 | Agent 方式 | 元数据级 |
| **物化加速** | OpenSearch 原生 | 无 | Delta Lake | Parquet | 无 |
| **成本** | 低 | 中 | 高 | 极高 | 低 |
| **目标用户** | AI 工程师 | 数据分析师 | 数据团队 | 企业决策者 | 平台架构师 |

---

## 20. MVP 与实施路线图 (MVP & Roadmap)

### 20.1. MVP 范围 (8-12 周)

**MVP 目标**: 验证核心价值 —— "连接即查询 + 一键向量化"

| 模块 | MVP 范围 | 不包含 |
| :--- | :--- | :--- |
| **Catalog** | 单 Catalog，支持 MySQL/PG/S3 | 多 Catalog、权限管理 |
| **Asset** | Table, Fileset, View, **Dataset** 四种类型 | API, Metric, Topic, Index |
| **Dataset** | 创建、批量写入、删除、Schema 定义 | 流式写入、版本控制、分区管理 |
| **Query** | 基础 DSL (select, filter, join) | 复杂聚合、窗口函数 |
| **Engine** | DuckDB only | Trino 集成 |
| **Sync** | 手动触发全量同步 | 增量 CDC、定时调度 |
| **Vector** | 单一 Embedding 模型 | 多模型、自定义模型 |
| **API** | REST API | gRPC、SDK |
| **Auth** | API Key 认证 | RBAC、SSO |

### 20.2. 实施阶段

```
┌─────────────────────────────────────────────────────────────────┐
│                      Implementation Phases                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Phase 1: Foundation (Week 1-4)                                  │
│  ├── 项目脚手架 (Go modules, CI/CD)                               │
│  ├── Catalog Service + 元数据存储                                 │
│  ├── Connection 管理 (MySQL/PG 连接池)                            │
│  └── 基础 REST API 框架                                           │
│                                                                  │
│  Phase 2: Query Engine (Week 5-8)                                │
│  ├── DSL Parser + AST                                            │
│  ├── DuckDB 集成 (联邦查询)                                       │
│  ├── Query Transpiler (DSL -> SQL)                               │
│  └── 查询结果流式返回                                             │
│                                                                  │
│  Phase 3: Materialization + Dataset (Week 9-12)                  │
│  ├── OpenSearch 集成 (读写)                                      │
│  ├── ETL Worker (Trino -> OpenSearch)                            │
│  ├── **Dataset API (创建/批量写入/删除)**                         │
│  ├── Embedding Pipeline                                          │
│  └── 向量检索 API                                                 │
│                                                                  │
│  Phase 4: Production Ready (Week 13-16)                          │
│  ├── Trino 集成                                                   │
│  ├── CDC Stream Worker                                           │
│  ├── RBAC + 审计日志                                              │
│  └── 监控告警 + 运维工具                                          │
│                                                                  │
│  Phase 5: Advanced Features (Week 17+)                           │
│  ├── 更多 Asset 类型 (API, Metric, Topic)                         │
│  ├── 数据血缘可视化                                               │
│  ├── Python SDK (Client only) + CLI                              │
│  └── 自托管 Embedding 模型                                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 20.3. 关键里程碑

| 里程碑 | 交付物 | 验收标准 |
| :--- | :--- | :--- |
| **M1: 可连接** | Catalog + Connection | 成功连接 MySQL 并浏览表结构 |
| **M2: 可查询** | Query Engine | DSL 查询返回正确结果 (< 1s) |
| **M3: 可物化** | Sync + OpenSearch | 100万行数据成功同步至 OpenSearch |
| **M3.5: 可写入** | Dataset API | 创建 Dataset + 批量写入 10 万条 (< 10s) |
| **M4: 可向量化** | Embedding + Search | 向量检索 Top-10 (< 100ms) |
| **M5: 可生产** | Auth + Observability | 通过安全审计 + 7x24 稳定运行 |

---

## 21. 总结

本设计通过 **"Asset Catalog + OpenSearch"** 的组合，创造了一个独特的价值点：它可以像 Catalog 一样轻量地管理多源数据，同时又能像专业向量数据库一样为 AI 应用提供极速的数据供给。它是企业从传统 BI 向 AI 转型过程中的理想中间件。

**核心差异化优势**:
1. **Virtual First**: 零 ETL 数据探索，分钟级接入
2. **OpenSearch Native**: 向量检索 + 随机访问，毫秒级响应
3. **Dataset API**: 原生可写数据集，AI 产出直接落地
4. **Unified DSL**: AI Agent 友好，一套语言查所有数据
5. **Lightweight**: 单二进制部署，无 JVM 依赖

---

*文档版本: v2.2*
*最后更新: 2025-01-14*

