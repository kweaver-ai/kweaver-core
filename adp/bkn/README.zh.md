# 本体引擎 (Ontology Engine)

[中文](README.zh.md) | [English](README.md)

## 项目简介

本体引擎是一个基于Go语言开发的分布式业务知识网络管理系统，提供本体建模、数据管理和智能查询功能。该系统采用微服务架构，分为本体管理模块和本体查询模块，支持大规模知识网络的构建、存储和查询。

本体引擎是KWeaver AI平台的核心组件，专注于构建企业级业务知识网络，实现业务知识的建模、存储、查询和应用。系统采用清洁架构设计，遵循SOLID原则，具有良好的可扩展性和可维护性。

### 核心特性

#### 本体建模与管理
- **多维度本体定义**: 支持对象类、关系类、行动类的完整定义和管理
- **分支管理**: 支持本体模型的分支开发和合并
- **可视化配置**: 提供直观的本体模型可视化配置界面

#### 知识网络构建
- **多领域知识网络**: 支持构建跨领域的复杂知识网络
- **语义关系管理**: 支持定义和管理复杂的语义关系
- **拓扑分析**: 提供知识网络的拓扑结构分析和优化

#### 智能查询引擎
- **复杂关系查询**: 支持多跳关系路径查询和子图查询
- **语义搜索**: 基于向量相似度的智能语义搜索
- **性能优化**: 基于OpenSearch的高性能查询引擎

#### 数据集成与应用
- **VEGA虚拟化**: 通过VEGA虚拟化引擎集成多种数据源
- **数据同步**: 支持本体数据与业务数据的实时同步
- **任务调度**: 支持复杂的后台任务调度和执行
- **权限管理**: 基于角色的细粒度权限控制

#### 系统特性
- **微服务架构**: 基于微服务设计，支持水平扩展
- **高可用性**: 支持分布式部署和故障恢复
- **监控集成**: 集成OpenTelemetry实现全链路监控
- **国际化支持**: 多语言支持和本地化配置

## 系统架构

### 模块组成

```text
adp/
└── bkn/
    ├── bkn-backend/     # 本体管理模块
    ├── ontology-query/       # 本体查询模块
    ├── README.md             # 项目说明文档
    └── README.zh.md          # 中文说明文档
```

### BKN管理模块 (bkn-backend)

负责BKN模型的创建、编辑和管理，主要功能包括：

- **知识网络管理**: 构建和管理业务知识网络
- **对象类管理**: 定义和管理知识网络中的对象类
- **关系类管理**: 定义和管理知识网络中的关系类
- **行动类管理**: 定义可执行的操作和行动
- **任务调度**: 后台任务和作业管理

### 本体查询模块 (ontology-query)

提供高效的知识图谱查询服务，主要功能包括：

- **模型查询**: 本体模型的查询和浏览
- **图谱查询**: 复杂的关系路径查询
- **语义搜索**: 基于语义的智能搜索
- **数据检索**: 多维度数据过滤和检索

## 快速开始

### 环境要求

- **Go**: 1.24.0 或更高版本
- **数据库**: MariaDB 11.4+ 或 DM8（用于数据存储）
- **搜索引擎**: OpenSearch 2.x（用于搜索和索引）
- **依赖服务**: 需要KWeaver平台的其他服务支持
- **Docker**: 可选，用于容器化部署
- **Kubernetes**: 可选，用于集群部署

### 本地开发

#### 1. 克隆代码库

```bash
git clone https://github.com/kweaver-ai/adp.git
cd adp/ontology
```

#### 2. 配置环境

每个模块都有独立的配置文件，需要根据实际环境进行配置：

```yaml
# BKN管理模块配置
bkn-backend/server/config/bkn-backend-config.yaml

# 本体查询模块配置
ontology-query/server/config/ontology-query-config.yaml
```

**关键配置项**：
- 数据库连接信息（host、port、user、password）
- OpenSearch连接信息
- 依赖服务地址（如user-management、data-model等）
- 服务端口配置

#### 3. 初始化数据库

```bash
# 执行数据库初始化脚本
# MariaDB
mysql -u root -p < bkn-backend/migrations/mariadb/6.x.x/pre/init.sql

# DM8
disql SYSDBA/SYSDBA@localhost:5236 < bkn-backend/migrations/dm8/6.x.x/pre/init.sql
```

#### 4. 运行BKN管理模块

```bash
cd bkn-backend/server
go mod download
go run main.go
```

服务将在 `http://localhost:13014` 启动

**健康检查**：
```bash
curl http://localhost:13014/health
```

#### 5. 运行本体查询模块

```bash
cd ../ontology-query/server
go mod download
go run main.go
```

服务将在 `http://localhost:13018` 启动

**健康检查**：
```bash
curl http://localhost:13018/health
```

### Docker 部署

#### 构建镜像

```bash
# 构建BKN管理模块
cd bkn-backend
docker build -t bkn-backend:latest -f docker/Dockerfile .

# 构建本体查询模块  
cd ../ontology-query
docker build -t ontology-query:latest -f docker/Dockerfile .
```

#### 运行容器

```bash
# 运行BKN管理模块
docker run -d -p 13014:13014 --name bkn-backend bkn-backend:latest

# 运行本体查询模块
docker run -d -p 13018:13018 --name ontology-query ontology-query:latest
```

### Kubernetes 部署

项目提供了Helm charts用于Kubernetes部署：

```bash
# 部署BKN管理模块
helm3 install bkn-backend bkn-backend/helm/bkn-backend/

# 部署本体查询模块
helm3 install ontology-query ontology-query/helm/ontology-query/
```

## API 文档

系统提供完整的RESTful API文档，支持OpenAPI 3.0规范：

### BKN管理API

- **知识网络API**: [知识网络API](bkn-backend/api_doc/bkn-backend-network.html)
  - 支持知识网络的创建、查询、更新和删除

- **对象类API**: [对象类API](bkn-backend/api_doc/bkn-backend-object-type.html)
  - 支持对象类的定义和管理

- **关系类API**: [关系类API](bkn-backend/api_doc/bkn-backend-relation-type.html)
  - 支持关系类的定义和管理
  - 提供方向性和多重性配置

- **动作类API**: [动作类API](bkn-backend/api_doc/bkn-backend-action-type.html)
  - 支持动作类的定义和管理

- **任务管理API**: [任务管理API](bkn-backend/api_doc/bkn-backend-job-api.html)
  - 支持后台任务的创建和调度

### 本体查询API

- **查询服务API**: [查询服务API](ontology-query/api_doc/ontology-query.html)
  - 支持复杂的关系路径查询
  - 提供语义搜索和模式匹配
  - 支持多维度数据过滤和检索
  - 提供高性能的分页查询

### API访问方式

**本地开发环境**:
```
# BKN管理API
http://localhost:13014/api/bkn-backend/v1/

# 本体查询API
http://localhost:13018/api/ontology-query/v1/
```

**生产环境**:
```
# BKN管理API
https://your-domain.com/api/bkn-backend/v1/

# 本体查询API
https://your-domain.com/api/ontology-query/v1/
```

## 数据库支持

系统支持多种数据库：

- **MariaDB**: 主数据存储
- **DM8**: 达梦数据库支持
- **OpenSearch**: 搜索引擎和数据分析

数据库升级脚本位于：

- `bkn-backend/migrations/`
- `ontology-query/migrations/`

## 监控与日志

- **日志系统**: 集成结构化日志，支持多级别日志记录
- **链路追踪**: 基于OpenTelemetry的分布式链路追踪
- **健康检查**: 提供健康检查端点

## 开发指南

### 代码结构

项目采用清洁架构设计，遵循SOLID原则，代码结构清晰，易于维护和扩展：

```text
server/
├── common/              # 公共配置、工具函数和常量
├── config/              # 配置文件和配置加载
├── drivenadapters/      # 数据访问层（驱动适配器）
├── driveradapters/      # 接口适配层（驱动适配器）
├── errors/              # 错误定义和处理机制
├── interfaces/          # 接口定义和抽象
├── locale/              # 国际化支持和多语言配置
├── logics/              # 业务逻辑层
├── main.go              # 应用入口和启动逻辑
├── version/             # 版本信息和构建信息
└── worker/              # 后台任务和作业执行器
```

### 开发规范

1. **清洁架构**: 遵循清洁架构原则，分离关注点
2. **接口隔离**: 明确定义接口和实现，依赖抽象而非具体实现
3. **错误处理**: 统一的错误处理机制，使用自定义错误类型
4. **日志规范**: 结构化的日志记录，使用Zap日志库
5. **测试覆盖**: 单元测试和集成测试，追求高测试覆盖率
6. **代码风格**: 遵循Go官方代码风格，使用go fmt格式化代码
7. **注释规范**: 清晰的注释，包括函数注释和复杂逻辑注释
8. **版本控制**: 遵循Git Flow工作流，使用语义化版本

### 测试指南

```bash
# 运行单元测试
cd bkn-backend/server
go test ./... -v

# 运行集成测试
cd bkn-backend/server
go test ./... -v -tags=integration

# 生成测试覆盖率报告
cd bkn-backend/server
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 贡献指南

参考kweaver项目内容

## 版本历史

- **v0.1.0**: 当前版本，基于Go 1.24

## 许可证

本项目采用 Apache License 2.0 许可证。详情请参阅 [LICENSE](../../LICENSE.txt) 文件。

## 支持与联系

- **技术支持**: AISHU ADP研发团队
- **文档更新**: 持续更新中
- **问题反馈**: 通过内部系统提交

---

**注意**: 这是一个企业级内部项目，代码和文档可能包含特定的业务逻辑和配置。请根据实际环境进行相应的调整。

