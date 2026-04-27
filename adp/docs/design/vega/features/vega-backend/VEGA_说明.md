# VEGA: 面向 AI 时代的数据虚拟化与物化服务
## Lightweight Data Virtualization & Materialization Service for the AI Era

---

## 1. 项目简介

### 1.1 产品定位
VEGA 是一个 AI Native 的数据供给层，解决 AI 应用对数据的三个需求：多源连接、快速探索、向量化加速。

### 1.2 核心价值
- **轻量级**：单二进制架构，无复杂依赖，秒级部署
- **虚拟化**：通过联邦机制实现连接即查询，无需数据搬运
- **物化加速**：通过 Sync 机制将数据物化到高性能存储引擎，支持向量检索

### 1.3 目标用户
- **AI 工程师**: 快速接入多源数据,构建 RAG 应用
- **数据分析师**: 探索和分析跨源数据,无需 ETL
- **平台架构师**: 构建统一的数据服务层

### 1.4 与其他方案的差异化
相比 Palantir Foundry 的"重型操作系统"定位，VEGA 定位为"轻量级 AI 加速器"。

| 特性 | VEGA | Palantir Foundry | Trino/Presto | Unity Catalog |
| :--- | :--- | :--- | :--- | :--- |
| **部署复杂度** | 低(单二进制) | 高(企业级) | 中(集群) | 高(云绑定) |
| **向量检索** | 原生支持 | 不支持 | 不支持 | 需集成 |
| **虚拟查询** | 内置 DuckDB/Trino | 需 ETL | 原生 | 需 Spark |
| **物化加速** | 多引擎(ES/OpenSearch/LanceDB) | Parquet | 无 | Delta Lake |
| **原生多模态支持** | **极高 (统一抽象表、文件、API、流)** | 中 (侧重对象模型，需 ETL) | 低 (侧重结构化 SQL) | 中 (侧重文件与表) |
| **目标用户** | AI 工程师 | 企业决策者 | 数据分析师 | 数据团队 |
---

## 2. 名词解释

### 2.1 Catalog (目录/联邦网关)
**定义**: 数据源连接与命名空间的管理单元, 是系统的一级命名空间。

**两类 Catalog**:

| 类型 | 说明 |
| :--- | :--- |
| **Physical Catalog** | 对应真实数据源连接 |
| **Logical Catalog** | 逻辑命名空间 |

**连接粒度**: Physical Catalog 支持两种连接粒度:
- **实例级连接**: 不指定 Database,连接到数据库实例,可发现和管理该实例下所有数据库的资源
- **库级连接**: 指定 Database,仅连接到特定数据库（向后兼容已有行为）

**设计原则**: Virtual First - 配置即联通,无需预先 ETL。

### 2.2 Data Resource (数据资源)
**定义**: 通过数据创造价值的统一实体, 是系统的二级命名空间(`catalog.resource`)。

**命名与标识**:
- **Name**: 用户可见标识。对于实例级 Catalog 下发现的资源,格式为 `db.table`（含数据库前缀）;库级 Catalog 下的资源则直接使用表名
- **Database**: 可选字段,记录资源所属的数据库名称,仅对有 database 概念的数据源（MySQL、PostgreSQL 等）有意义,可用于筛选和分组
- **SourceIdentifier**: 源端标识,存储原始表名/文件路径等,连接器实际使用
- **唯一约束**: `(catalog_id, name)`,因为实例级 Catalog 下 name 已包含 database 前缀,天然唯一

**九大数据资源类型**:

| 资源类型 | 数据源示例 | 语义 | Virtual 模式 | Local 模式 | 归属 Catalog |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Table** | MySQL, PostgreSQL | 结构化表 | JDBC 联邦查询 | CDC 实时同步 | Physical / Logical |
| **File** | Excel, Parquet, Lance, CSV, JSON | 单体结构化文件 | DuckDB SQL 直接查询 | 快速物化 (Bulk Load) | Physical / Logical |
| **Fileset** | S3, 飞书, AnyShare, Notion | 非结构化文件集 | 浏览/预览 | ETL Pipeline | Physical / Logical |
| **API** | REST, GraphQL | 应用接口 | 实时调用 | 定时轮询 | Physical / Logical |
| **Metric** | Prometheus, InfluxDB | 时序指标 | PromQL 下推 | 批量归档 | Physical / Logical |
| **Topic** | Kafka, Pulsar | 实时流 | 实时采样 | 微批写入 | Physical / Logical |
| **Index** | ElasticSearch, OpenSearch | 搜索引擎 | Search DSL 透传 | Reindex | Physical / Logical |
| **LogicView** | 其他数据资源衍生/复合 | 逻辑视图 | 实时计算 | 物化快照 | **仅 Logical** |
| **Dataset** | API 写入 | 原生可写数据集 | 不支持 | 直接存储 | **仅 Logical** |

### 2.3 两种查询方式 (Dual Query Mode)
每个 Data Resource 支持两种查询方式:

1. **Virtual 查询**: 实时查询源端数据(联邦查询),无需物化,适合即时查询和数据探索
2. **Local 查询**: 查询本地存储引擎中的物化数据,适合 AI 分析、向量检索和高性能查询

**核心原则**:
- 物化是 Data Resource 的一种**能力**,而非特殊类型
- 同一 Data Resource 可以同时支持 Virtual 和 Local 两种查询方式
- 系统根据查询需求自动路由到最优查询方式

### 2.4 Sync (同步/物化)
将源端数据物化到本地存储引擎的过程,根据数据资源类型采用不同策略:
- **Table**: CDC(Change Data Capture)实时同步
- **Topic**: 微批消费
- **API**: 定时轮询
- **LogicView**: 执行计算逻辑,物化为快照

### 2.5 统一类型系统
**VEGA 类型系统**: 将异构数据源的类型映射为统一的标准类型,确保跨源数据的类型一致性。

**核心类型**:
- **数值类型**: `integer`, `unsigned_integer`, `float`, `decimal`
- **字符串类型**: `string`(精确匹配), `text`(全文搜索)
- **时间类型**: `date`, `datetime`, `time`
- **特殊类型**: `boolean`, `binary`, `json`, `vector`

### 2.6 字段特征 (Field Features)
为字段赋予超越基础类型的能力:
- **keyword**: 精确匹配、排序、聚合
- **fulltext**: 全文检索、中文分词
- **vector**: 向量语义搜索

无需修改源端 Schema，通过引用机制扩展能力。

### 2.7 LogicView (逻辑视图)
**定义**: LogicView 是基于其他数据资源的衍生或复合,通过逻辑定义创建新的数据视图,不直接对应物理数据源。

**两种形态**:

1. **衍生 (Derived)**: 基于单个数据资源的转换和加工
   - 示例: 对 Table 进行过滤、聚合、字段映射
   - 示例: 对 Fileset 进行格式转换、内容提取

2. **复合 (Composite)**: 多个数据资源的合并和关联
   - 示例: 多个 Table 的 Union 或 Join
   - 示例: Table + Fileset + API 的跨类型融合

**核心特性**:
- **无物理存储**: LogicView 本身不存储数据,数据来源于其依赖的数据资源
- **逻辑定义**: 通过声明式配置定义数据转换和合并规则
- **依赖关系**: 明确记录与源数据资源的血缘关系
- **两种查询方式**:
  - Virtual 查询: 实时计算,从源数据资源读取并执行逻辑
  - Local 查询: 物化快照,将计算结果缓存到本地存储
- **级联更新**: 源数据资源变化时,可触发 LogicView 的刷新

**定义方式**:
- **SQL 表达式**: 适用于结构化数据的关系运算
- **声明式映射**: 适用于异构数据的字段映射和类型转换
- **自定义脚本**: 适用于复杂的业务逻辑

**典型场景**:
- **数据过滤**: 从大表中筛选特定条件的数据子集
- **字段派生**: 基于现有字段计算新字段(如全名 = 姓 + 名)
- **多表关联**: 用户表 JOIN 订单表 → 用户订单全景
- **跨类型融合**: Table(结构化) + Fileset(文档) + API(外部数据) → 统一知识库
- **数据降维**: 从明细数据聚合为汇总数据

**与其他数据资源的区别**:

| 维度 | LogicView | 物理数据资源 (Table/Fileset 等) |
| :--- | :--- | :--- |
| **数据来源** | 衍生自其他数据资源 | 直接对应物理数据源 |
| **存储** | 无物理存储(可物化缓存) | 物理存储在源端 |
| **依赖关系** | 依赖源数据资源 | 无依赖 |
| **变更传播** | 源变化可触发刷新 | 自身变化 |
| **归属位置** | 必须归属 Logical Catalog | 归属 Physical/Logical Catalog |

---

### 2.8 新老版本概念对应
| 老版本 | 新版本 |
| :---| :--- |
| DataConnection 数据连接 | Catalog |
| Table 库表 | Table(Data Resource) |
| LogicView 逻辑视图 | Data Resource |
| AtomicView 原子视图 | Data Resource（Table、Index、Fileset） |
| CustomView 自定义视图 | LogicView(Data Resource) |
| Excel/CSV/Json 数据源 | File(Data Resource) |


## 3. 设计理念

### 3.1 Virtual First
**理念**: 先连接,后物化。

- **零 ETL 探索**：配置数据源后立即可查询，无需预先加载数据
- **按需物化**：需要性能加速时才触发 Sync
- **分钟级接入**：新数据源接入只需几分钟

### 3.2 统一实体模型
**理念**: 用 Data Resource 抽象所有数据类型,提供一致的操作界面。

- **统一命名**: 所有数据都是 `catalog.resource`
- **统一接口**: 查询、同步、权限管理等操作对所有数据资源类型一致
- **统一存储**: 物化后都收敛为标准化的索引结构

### 3.3 适应性路由
**理念**: 系统根据查询类型和数据规模自动选择最优执行路径。

```
查询请求
  ↓
物化数据可用? → 是 → Local 查询 (毫秒级)
  ↓ 否
向量/全文搜索? → 是 → 建议物化
  ↓ 否
单源查询? → 是 → Virtual 查询
  ↓ 否
Virtual 查询 (跨源 Join)
  ↓
数据量 < 1GB? → 是 → DuckDB (嵌入式)
  ↓ 否
Trino (分布式)
```

### 3.4 多引擎物化存储
**理念**: 根据场景选择最适合的物化存储引擎,支持 ElasticSearch、OpenSearch、LanceDB 等。

**存储引擎选型对比**:

| 维度 | ElasticSearch/OpenSearch | LanceDB | pgvector |
| :--- | :--- | :--- | :--- |
| **向量规模** | 十亿级 | 亿级+ | 千万级 |
| **向量性能** | 高 (GPU 加速) | 极高 (零拷贝) | 中 |
| **全文搜索** | 原生强项 | 基础 | 支持 |
| **混合检索** | 原生支持 | 原生支持 | 需手动 |
| **结构化查询** | DSL | SQL-like | SQL |
| **聚合分析** | 强 | 基础 | 强 |
| **Trino 集成** | 官方 Connector | 需自研 | 官方 Connector |
| **云托管** | 完善 | LanceDB Cloud | 各云厂商 |
| **适用场景** | 混合检索、全文搜索 | 纯向量检索、大规模 | 小规模、关系型为主 |

**引擎选型策略**:

| 场景 | 推荐引擎 | 理由 |
| :--- | :--- | :--- |
| **RAG 混合检索** | ElasticSearch/OpenSearch | 向量 + 全文搜索原生支持 |
| **大规模向量检索** | LanceDB | 零拷贝架构,向量性能极致 |
| **以关系查询为主** | pgvector | SQL 原生,与现有数据库集成 |
| **企业级全文搜索** | ElasticSearch/OpenSearch | 倒排索引成熟,生态完善 |
| **轻量级部署** | LanceDB | 嵌入式友好,资源占用小 |

架构支持：
- 通过适配器模式统一接口，可动态切换存储引擎
- 不同数据资源可选择不同存储引擎
- 支持物化数据在不同引擎间迁移

### 3.5 双引擎策略
**理念**: 根据数据规模和查询复杂度,智能选择计算引擎。

| 引擎 | 场景 | 优势 |
| :--- | :--- | :--- |
| **DuckDB** | 数据探索、小文件分析(<10GB) | 进程内零延迟,部署简单 |
| **Trino** | 跨库大数据量 Join | 分布式内存计算,成熟稳定 |

---


## 4. 总体架构

### 4.1 系统分层

```
┌─────────────────────────────────────────────────────────────┐
│                    Application Layer                         │
│                  (AI Apps, BI Tools, APIs)                   │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                     VEGA Service Layer                       │
│  ┌─────────────┐  ┌─────────────┐  ┌────────────────────┐  │
│  │  Query API  │  │  Sync API   │  │  Metadata API      │  │
│  └─────────────┘  └─────────────┘  └────────────────────┘  │
│                              │                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │              Unified DSL Parser & Router              │  │
│  └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
          ↓ Virtual 查询              ↓ Local 查询
┌──────────────────────┐    ┌──────────────────────────────┐
│  Federation Layer    │    │    Materialization Layer     │
│                      │    │                              │
│  ┌────────────────┐  │    │  ┌────────────────────────┐ │
│  │ DuckDB Engine  │  │    │  │  Storage Engines       │ │
│  └────────────────┘  │    │  │  - ES/OpenSearch       │ │
│  ┌────────────────┐  │    │  │  - LanceDB             │ │
│  │ Trino Cluster  │  │    │  │  - pgvector (可选)    │ │
│  └────────────────┘  │    │  └────────────────────────┘ │
└──────────────────────┘    │  ┌────────────────────────┐ │
          ↓                 │  │  Sync Workers          │ │
┌──────────────────────┐    │  │  - CDC Consumer        │ │
│   Data Sources       │    │  │  - ETL Pipeline        │ │
│  - MySQL/PostgreSQL  │    │  │  - Polling Scheduler   │ │
│  - S3/HDFS          │    │  └────────────────────────┘ │
│  - Kafka/Pulsar     │    └──────────────────────────────┘
│  - Prometheus       │                  ↓
│  - 飞书/Notion      │    ┌──────────────────────────────┐
└──────────────────────┘    │      Metadata Store          │
                            │    (Resource Registry)       │
                            └──────────────────────────────┘
```

### 4.2 核心组件

#### 4.2.1 Control Plane (控制平面)
- **API Gateway**: 对外统一接口,处理查询、同步、元数据管理请求
- **DSL Parser**: 解析统一 DSL,转译为目标方言(SQL/PromQL/ES DSL)
- **Query Router**: 根据查询类型和数据状态路由到最优执行引擎
- **Catalog Manager**: 管理数据源连接和命名空间
- **Metadata Store**: 存储 Catalog、Data Resource、Schema 等元数据

#### 4.2.2 Data Plane (数据平面)
- **Virtual Engine**: DuckDB(嵌入式) + Trino(分布式),负责联邦查询
- **Local Engine**: 可插拔的物化存储引擎
  - ElasticSearch/OpenSearch: 混合检索场景
  - LanceDB: 大规模向量检索场景
  - pgvector: 小规模或关系型为主场景
- **Sync Workers**: 负责不同数据资源类型的数据同步
  - CDC Consumer: 实时捕获数据库变更
  - ETL Pipeline: 解析和转换文件数据
  - Polling Scheduler: 定时轮询 API 数据
  - Stream Consumer: 消费消息队列数据

#### 4.2.3 Discover & Monitoring (发现与监控)
- **Inventory Worker**: 自动发现数据源中的数据资源
- **Health Checker**: 监控数据源连接状态
- **Metrics Collector**: 收集查询性能和同步进度指标

### 4.3 数据流

#### 查询流程
```
Client Query
  ↓
API Gateway (DSL)
  ↓
Query Router (根据查询类型和数据状态选择查询方式)
  ├─→ Local 可用? → Local 查询(物化存储) → Result
  ├─→ 单源查询? → Virtual 查询(源端直连) → Result
  └─→ 跨源 Join? → Virtual 查询(DuckDB/Trino) → Result
```

#### 同步流程
```
User Trigger Sync
  ↓
Sync API
  ↓
Sync Worker (by Resource Type)
  ├─→ Table → CDC Consumer → Binlog → Transform → Local Engine
  ├─→ File → Bulk Load → Direct IO → Local Engine
  ├─→ Fileset → ETL Pipeline → Parse → Extract → Local Engine
  ├─→ API → Polling → Flatten → Local Engine
  └─→ Topic → Stream Consumer → Micro-batch → Local Engine
```

#### 发现流程
```
Inventory Worker (Scheduled)
  ↓
Catalog 连接粒度?
  ├─→ 实例级 → ListDatabases() → 遍历每个 Database → ListTables()
  │     └─→ Resource Name = "db.table", Database = db
  └─→ 库级 → ListTables()
        └─→ Resource Name = table
  ↓
Auto Discover
  ├─→ New Resource → Register as Virtual
  ├─→ Deleted Resource → Mark as Stale
  └─→ Schema Change → Update Metadata
```

### 4.4 技术栈选型

| 组件 | 技术选型 | 理由 |
| :--- | :--- | :--- |
| **服务框架** | Golang | 高性能并发,云原生生态成熟 |
| **元数据存储** | MySQL/PostgreSQL | 关系型数据存储,事务支持 |
| **物化存储** | ES/OpenSearch/LanceDB | 可插拔引擎,按场景选择 |
| **嵌入式引擎** | DuckDB | 进程内 OLAP,支持 Parquet/JSON |
| **分布式引擎** | Trino | 标准 Data Virtualization 引擎 |
| **CDC 引擎** | Debezium/Canal | 成熟的 CDC 解决方案 |

---

## 5. 功能模块

### 5.1 Catalog 管理模块（入口）

#### 功能描述
管理数据源连接和命名空间,提供 Physical Catalog 和 Logical Catalog 两种类型。

#### 核心能力
- **连接管理**: 配置和测试数据源连接(JDBC/S3/Kafka/API 等)
- **状态监控**: 实时监控连接健康状态(`healthy`/`degraded`/`unhealthy`)
- **命名空间**: 为数据资源提供逻辑分组和访问控制边界

#### 设计要点
- **Physical Catalog**: 与真实数据源一对一映射,连接信息加密存储,支持实例级和库级两种连接粒度
- **实例级连接**: ConnectorConfig.Database 为可选字段,不指定时连接到实例级别,可发现该实例下所有用户数据库的资源
- **库级连接**: 指定 Database 时仅管理该数据库下的资源,与已有行为兼容
- **Logical Catalog**: 纯逻辑命名空间,用于管理 LogicView 和 Dataset
- **级联效应**: Catalog 禁用时,其下所有数据资源均不可访问

---

### 5.2 数据资源注册与发现模块（数据自动进入）

#### 功能描述
自动发现数据源中的数据资源(表/文件/索引等),注册并维护元数据。

#### 核心能力
- **自动发现**: Inventory Worker 定期发现数据源中的新数据资源
- **多库发现**: 对于实例级 Catalog,连接器通过 `ListDatabases()` 获取所有用户数据库,再遍历每个库发现资源;资源 Name 格式为 `db.table`,Database 字段记录所属库名
- **Schema 同步**: 自动提取和更新数据资源的 Schema 信息
- **变更处理**: 处理数据资源的新增、删除、Schema 变更

#### 发现策略
| 触发模式 | 说明 | 适用场景 |
| :--- | :--- | :--- |
| **manual** | 手动触发 | 变更频率低的环境 |
| **scheduled** | 定时触发(cron) | 常规生产环境 |
| **event_driven** | 事件驱动 | 支持 DDL 事件的数据源 |

#### 变更处理策略
| 事件 | 策略 | 行为 |
| :--- | :--- | :--- |
| **新增数据资源** | auto_register / pending_review / ignore | 自动注册 / 待审核 / 忽略 |
| **删除数据资源** | auto_remove / mark_stale / ignore | 自动删除 / 标记过期 / 忽略 |
| **Schema 变更** | auto_update / pending_review / ignore | 自动更新 / 待审核 / 忽略 |
| **文件内容变更** | pending_review | 对于 File 数据资源，若底层文件 Schema 发生变化，系统应触发 pending_review 策略，以防止上层 LogicView 的 SQL 报错。

---

#### 5.2.1 File 数据资源连接机制
连接逻辑
- 连接共享：File 数据资源本身不存储连接凭据，而是从其所属的 Physical Catalog（如 S3 或本地文件系统）中获取认证信息。
- URI 绑定：在注册 File 数据资源时，系统记录文件在存储系统中的相对路径或绝对 URI。
- Schema 推导：连接建立后，VEGA 调用 DuckDB 引擎读取文件头（Metadata），自动将文件原始类型映射为 VEGA 统一类型系统。

执行流程
- Catalog 授权：通过 Physical Catalog 打通与 S3/HDFS 的存储链路。
- 按需加载：查询时根据元数据信息，仅连接并下载该文件所需的特定列或数据块。
- 引擎处理：利用 DuckDB 的嵌入式能力进行零拷贝计算，实现文件数据的即时 SQL 化。

---

### 5.3 统一查询接口(DSL)

#### 功能描述
提供统一的查询 DSL,屏蔽底层引擎差异,支持 SQL 和 JSON DSL 两种查询方式。

#### 核心能力
- **DSL 解析**: 将统一 DSL 解析为 AST(抽象语法树)
- **方言转译**: 将 AST 转译为目标方言(MySQL SQL/Trino SQL/ES DSL)
- **查询优化**: 谓词下推、投影裁剪、Join 重排序

#### DSL 设计
```json
{
  "from": "catalog.resource",
  "select": ["field1", "field2"],
  "where": {
    "and": [
      {"field": "age", "op": "gt", "value": 18},
      {"field": "city", "op": "eq", "value": "Beijing"}
    ]
  },
  "orderBy": [{"field": "created_at", "dir": "desc"}],
  "limit": 10
}
```

#### 向量检索扩展
```json
{
  "from": "catalog.resource",
  "vectorSearch": {
    "field": "embedding",
    "vector": [0.1, 0.2, ...],
    "k": 10,
    "metric": "cosine"
  },
  "hybridSearch": {
    "textQuery": "机器学习",
    "textWeight": 0.3,
    "vectorWeight": 0.7
  }
}
```

---

### 5.4 字段特征管理(Features)

#### 功能描述
为数据资源字段赋予超越基础类型的能力,支持向量检索、全文搜索、精确匹配等高级特性。

#### 核心能力
- **特征定义**: 为字段配置 keyword/fulltext/vector 特征
- **引用机制**: Virtual 模式下通过 RefField 引用其他字段
- **热切换**: 支持为一个字段配置多个特征,通过 Enabled 状态切换
- **物化增强**: Virtual 模式下定义的特征,物化时自动转换为目标引擎的原生能力

#### 特征类型
| 类型 | 用途 | 配置示例 |
| :--- | :--- | :--- |
| **keyword** | 精确匹配、排序、聚合 | `ignore_above_len: 2048` |
| **fulltext** | 全文检索、中文分词 | `analyzer: "ik_max_word"` |
| **vector** | 向量语义搜索 | `dimension: 768, space_type: "cosinesimil"` |

设计要点：
- 无需修改源端 Schema
- `IsNative` 区分系统同步和人工扩展
- `RefField` 允许字段借用其他列的能力

---

### 5.5 虚拟查询模块 (Virtual 查询)

#### 功能描述
提供 Virtual 查询能力,直连数据源实时查询,无需数据搬运。

#### 核心能力
- **多源连接**: 支持 8 大类数据资源的虚拟查询
- **双引擎路由**: 根据数据规模选择 DuckDB 或 Trino
- **查询下推**: 将过滤、聚合等操作下推到数据源执行

#### 查询路由策略
```
SQL Query
  ↓
单表查询? → 是 → 源端直连 (MySQL/PG/ClickHouse)
  ↓ 否
同源 Join? → 是 → 下推至源端执行
  ↓ 否
跨源 Join
  ↓
数据量 < 1GB? → 是 → DuckDB (嵌入式)
  ↓ 否
Trino (分布式集群)
```

#### 数据资源类型特定行为
| 资源类型 | Virtual 模式行为 |
| :--- | :--- |
| **Table** | JDBC 联邦查询,支持谓词下推 |
| **File** | 自动推导 Schema，利用 DuckDB 像查询数据库表一样直接执行 SQL，支持 WHERE 过滤和列裁剪 |
| **Fileset** | 浏览文件列表,预览文件内容 |
| **API** | 实时调用 API,返回前 N 条作为预览 |
| **Metric** | 透传 PromQL/InfluxQL 查询 |
| **Topic** | 采样最新 100 条消息 |
| **Index** | Search DSL 透传到源端 |
| **LogicView** | 实时计算,从源数据资源读取并执行逻辑 |

---

### 5.6 物化加速模块 (Local 查询)

#### 功能描述
提供 Local 查询能力,通过物化数据到本地存储引擎,解锁向量检索和高性能查询。

#### 核心能力
- **多引擎适配**: 支持 ES/OpenSearch/LanceDB 等存储引擎
- **多策略同步**: 根据数据资源类型采用不同同步策略
- **增量更新**: CDC/微批/增量轮询,最小化数据传输
- **类型映射**: 自动将源端类型映射为 VEGA 统一类型
- **向量化增强**: 为字段挂载 vector/fulltext/keyword 特征

#### 同步策略
| 资源类型 | 同步策略 | 说明 |
| :--- | :--- | :--- |
| **Table** | CDC 实时同步 | 监听 Binlog,实时捕获变更 |
| **File** | Bulk Load | 解析文件,提取结构化信息 |
| **Fileset** | ETL Pipeline | 解析文件,提取结构化信息 |
| **API** | 定时轮询 | 按配置周期调用 API,扁平化存储 |
| **Metric** | 批量归档 | 定时降采样,归档历史数据 |
| **Topic** | 微批消费 | 消费者组微批写入 |
| **Index** | Reindex | 从远程索引迁移到本地 |
| **LogicView** | 物化快照 | 执行计算逻辑,结果存储为索引 |
| **Dataset** | 直接写入 | API 驱动,直接写入存储引擎 |

#### 存储引擎适配
- **类型映射**: VEGA 类型 → 目标引擎类型
- **特征映射**: keyword/fulltext/vector 特征 → 引擎原生能力
- **索引策略**: 根据数据规模和引擎特性配置存储参数

---

### 5.7 LogicView 逻辑视图

#### 功能描述
提供基于其他数据资源的衍生和复合能力,通过逻辑定义创建新的数据视图。

#### 核心能力
- **衍生转换**: 对单个数据资源进行过滤、聚合、字段映射等转换
- **复合合并**: 支持多个数据资源的 Union、Join、Enrich 等操作
- **血缘追踪**: 自动记录与源数据资源的依赖关系
- **级联刷新**: 源数据资源变化时可触发 LogicView 更新
- **Schema 推导**: 自动推导衍生/复合后的 Schema

#### 定义方式
| 方式 | 适用场景 | 示例 |
| :--- | :--- | :--- |
| **SQL 表达式** | 结构化数据的关系运算 | `SELECT * FROM users WHERE age > 18` |
| **声明式映射** | 异构数据的字段映射 | `{source: "name", target: "user_name", transform: "upper"}` |
| **自定义脚本** | 复杂业务逻辑 | Python/JavaScript 脚本处理 |

#### 衍生场景
- **数据过滤**: 从用户表筛选活跃用户
- **字段计算**: 基于单价和数量计算总价
- **数据转换**: 将 JSON 格式转换为结构化字段

#### 复合场景
- **同类合并**: 多个分库的用户表 Union 为全局用户表
- **跨源关联**: 用户表(MySQL) JOIN 订单表(PostgreSQL)
- **跨类型融合**: 文档(Fileset) + 元数据(API) + 评论(Topic)

#### 刷新策略
| 模式 | 刷新策略 | 适用场景 |
| :--- | :--- | :--- |
| **Virtual 查询** | 实时计算 | 数据量小,实时性要求高,源数据资源在线 |
| **Local 查询 - 全量** | 定时全量刷新 | 数据量适中,可接受延迟 |
| **Local 查询 - 增量** | CDC/事件驱动增量 | 数据量大,需要准实时更新 |

#### 依赖管理
- **依赖声明**: 显式声明依赖的源数据资源
- **循环检测**: 防止 LogicView 间的循环依赖
- **影响分析**: 源数据资源变更时评估影响范围
- **版本管理**: 支持 LogicView 定义的版本控制

---

### 5.8 Dataset 原生存储模块

#### 功能描述
提供原生可写的数据集类型,数据通过 API 直接写入本地存储引擎,无需外部数据源。

#### 核心能力
- **Schema 定义**: 用户自定义字段结构、类型约束、主键
- **CRUD API**: 通过 REST API 进行创建、读取、更新、删除操作
- **向量化支持**: 原生支持 vector 字段,可直接进行向量检索

#### 典型使用场景
- **AI 产出存储**: RAG 提取的知识片段、LLM 生成的结构化数据
- **用户自定义数据**: 手动上传的数据、应用程序写入的事件
- **中间结果持久化**: 跨源 Join 结果、预计算的聚合数据
- **标注/反馈数据**: 人工标注的训练数据、用户反馈

#### 与其他数据资源的区别
| 维度 | Dataset | LogicView | 物理数据资源 |
| :--- | :--- | :--- | :--- |
| **数据来源** | API 直接写入 | 衍生自其他数据资源 | 物理数据源 |
| **写入方式** | API(可读写) | 只读(计算结果) | Sync(单向同步) |
| **依赖关系** | 无依赖 | 依赖源数据资源 | 无依赖 |
| **Virtual 模式** | 不支持 | 支持 | 支持 |

---

### 5.9 状态管理模块

#### 功能描述
管理 Catalog 和数据资源的状态,支持健康检查、状态变更、级联控制。

#### Catalog 状态
| 状态 | 说明 | 行为 |
| :--- | :--- | :--- |
| **healthy** | 连接正常 | 正常查询和同步 |
| **degraded** | 性能降级 | 可用但延迟高 |
| **unhealthy** | 连接异常 | 查询失败,同步暂停 |
| **offline** | 离线 | 不可达 |
| **disabled** | 已禁用 | 不进行健康检查 |

#### 数据资源状态
| 状态 | 说明 | 查询行为 | 同步行为 |
| :--- | :--- | :--- | :--- |
| **active** | 正常可用 | 正常执行 | 正常同步 |
| **disabled** | 已禁用 | 返回错误 | 暂停同步 |
| **deprecated** | 已废弃 | 返回警告 + 结果 | 继续同步 |
| **stale** | 数据过期 | 返回警告 + 结果 | 继续同步 |

#### 级联效应
- Catalog 状态优先于数据资源状态
- Catalog 禁用时,其下所有数据资源均不可访问
- Catalog 不健康时,数据资源查询可能失败但不会被禁用

---

### 5.10 降级与容错模块

#### 功能描述
当依赖组件不可用时,系统自动降级,保证核心功能可用。

#### 降级策略
```
Level 0 (正常): 完整功能
  ↓ Trino 不可用
Level 1: 禁用跨源 Join,提示使用单源查询
  ↓ Embedding Service 不可用
Level 2: 禁用向量检索,回退到关键词搜索
  ↓ CDC 延迟过高
Level 3: 物化数据标记为 stale,强制走 Virtual 查询
  ↓ 源数据库不可用
Level 4: 返回物化数据缓存 + 过期警告
```

#### 容错机制
- **查询超时**: 设置查询超时时间,避免长时间阻塞
- **连接池管理**: 连接池隔离,避免单个数据源故障影响全局
- **熔断机制**: 连续失败后自动熔断,避免雪崩效应
- **重试策略**: 指数退避重试,处理临时性故障

---

## 6. 总结

### 6.1 主要特点
1. **Virtual First**：零 ETL 数据探索，分钟级接入
2. **多引擎物化**：按场景选择存储引擎（ES/OpenSearch/LanceDB）
3. **LogicView 衍生**：基于数据资源的衍生和复合，内置血缘追踪
4. **Dataset API**：原生可写数据集，AI 产出直接落地
5. **Unified DSL**：AI Agent 友好，一套语言查所有数据
6. **Lightweight**：单二进制部署，无 JVM 依赖
7. **原生多模态**：通过 File 与 Fileset 统一处理结构化文件与非结构化文件集

### 6.2 适用场景
- **RAG 应用**: 多源知识接入 + 向量检索 + LogicView 融合
- **AI 训练数据准备**: 快速探索和抽取跨源数据,LogicView 加工
- **实时数据分析**: 联邦查询 + 物化加速 + 衍生指标
- **数据产品原型**: 快速验证数据可行性,灵活调整 LogicView

### 6.3 技术风险与应对
| 风险 | 应对措施 |
| :--- | :--- |
| **存储引擎资源开销** | 按场景选择轻量级引擎(LanceDB);支持冷热分离 |
| **DuckDB 内存溢出** | 设置内存上限;超限路由到 Trino |
| **CDC 延迟过高** | 监控 lag;支持手动全量同步 |
| **跨源 Join 性能差** | 引导用户物化热点数据;LogicView 预计算 |
| **LogicView 依赖复杂** | 依赖图可视化;循环依赖检测 |

### 6.4 后续计划
- 根据查询模式自动推荐物化策略
- 自动识别可下推的计算，选择合适的物化时机
- 支持流计算场景的持续物化
- 多租户隔离
- 自动冷热分离，降低存储成本
