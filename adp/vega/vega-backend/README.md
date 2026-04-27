# VEGA Manager

面向 AI 时代的轻量级数据虚拟化与物化服务。

## 功能特性

- **Virtual First**: 零 ETL 数据探索，分钟级接入
- **OpenSearch Native**: 向量检索 + 随机访问，毫秒级响应
- **Dataset API**: 原生可写数据集，AI 产出直接落地
- **Unified DSL**: AI Agent 友好，一套语言查所有数据
- **Lightweight**: 单二进制部署，无 JVM 依赖

## 目录结构

```
vega-backend/
├── docker/                 # Docker 配置
├── helm/                   # Kubernetes Helm charts
├── migrations/             # 数据库迁移脚本
├── plan/                   # 设计文档
└── server/                 # 主服务代码
    ├── common/             # 公共工具库
    ├── config/             # 配置定义
    ├── errors/             # 错误码定义
    ├── locale/             # 国际化
    ├── version/            # 版本信息
    ├── interfaces/         # 接口定义层 (Ports)
    ├── driveradapters/     # Primary Adapters (HTTP Handlers)
    ├── drivenadapters/     # Secondary Adapters (外部服务)
    │   ├── connectors/     # 数据源连接器 (按 Asset 类型组织)
    │   │   ├── table/      # Table 类型 (MySQL, PG, 达梦等)
    │   │   ├── fileset/    # Fileset 类型 (S3, 飞书等)
    │   │   ├── topic/      # Topic 类型 (Kafka, Pulsar)
    │   │   ├── metric/     # Metric 类型 (Prometheus, InfluxDB)
    │   │   ├── index/      # Index 类型 (OpenSearch, ES)
    │   │   └── api/        # API 类型 (REST, GraphQL)
    │   ├── metadata/       # 元数据存储
    │   ├── opensearch/     # OpenSearch 客户端
    │   ├── embedding/      # Embedding 服务
    │   └── cache/          # 缓存层
    └── logics/             # 业务逻辑层
        ├── catalog/        # Catalog 管理
        ├── asset/          # Asset 管理
        ├── query/          # 查询引擎
        ├── sync/           # 同步服务
        ├── dataset/        # Dataset 服务
        ├── embedding/      # 向量化服务
        ├── lineage/        # 数据血缘
        └── quality/        # 数据质量
```

## 快速开始

```bash
cd server
go mod tidy
go run main.go
```

## 文档

详细设计文档请参阅 `plan/` 目录。
