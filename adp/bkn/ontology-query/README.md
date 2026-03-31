# 本体查询模块 (Ontology Query)

## 模块概述

本体查询模块是本体引擎的查询服务组件，专注于提供高效、智能的知识网络查询能力。该模块基于OpenSearch和VEGA虚拟化查询技术，支持复杂的关系查询、语义搜索、多维度数据检索功能。

## 核心功能

### 1. 网络查询 (Network Query)
- **关系路径查询**: 支持复杂的多跳关系路径查询
- **子图查询**: 查询特定条件下的子图结构
- **模式匹配**: 基于图模式的智能匹配查询
- **最短路径**: 实体间的最短路径计算

### 2. 语义搜索 (Semantic Search)
- **向量搜索**: 基于向量相似度的语义搜索
- **关键词搜索**: 支持全文检索和模糊匹配
- **语义理解**: 自然语言查询的理解和转换
- **相关性排序**: 基于语义相关性的智能排序

### 3. 数据检索 (Data Retrieval)
- **多维度过滤**: 支持多条件组合过滤
- **分页查询**: 高效的分页查询支持
- **排序功能**: 多字段排序和自定义排序
- **字段选择**: 灵活的字段选择和投影
- **统计查询**: 数据的统计分析查询

### 4. 模型查询 (Model Query)
- **本体浏览**: 本体模型的浏览和导航
- **模型搜索**: 基于关键词和语义化的模型搜索
- **模型导出**: 模型的导出和序列化

## 技术架构

### 技术栈
- **编程语言**: Go 1.24
- **Web框架**: Gin 1.11
- **搜索引擎**: OpenSearch 2.x
- **查询语言**: 自定义查询语言
- **缓存**: 支持多种缓存策略
- **监控**: OpenTelemetry

### 架构设计
```
server/
├── common/              # 公共配置和工具
├── config/              # 配置文件
├── drivenadapters/      # 数据访问层
│   ├── agent_operator/  # AIAgent数据访问
│   ├── model_factory/   # 模型工厂数据访问
│   ├── ontology_manager/ # 本体管理数据访问
│   ├── opensearch/      # OpenSearch数据访问
│   └── uniquery/        # VEGA统一查询数据访问
├── driveradapters/      # 接口适配层
├── errors/              # 错误定义
├── interfaces/          # 接口定义
├── locale/              # 国际化支持
├── logics/              # 业务逻辑层
└── version/             # 版本信息
```

## API 接口

### 对象查询
- **对象数据**: `POST /api/ontology-query/v1/knowledge-networks/{kn_id}/object-types/{ot_id}`
- **对象属性**: `POST /api/ontology-query/v1/knowledge-networks/{kn_id}/object-types/{ot_id}/properties`
- **对象子图**: `POST /api/ontology-query/v1/knowledge-networks/{kn_id}/subgraph`

### 行动查询  
- **行动数据**: `POST /api/ontology-query/v1/knowledge-networks/{kn_id}/action-types/{at_id}`

### 内部接口
- 相同路径，前缀改为 `/api/ontology-query/in/v1/`
- 跳过OAuth认证

### 系统接口
- **健康检查**: `GET /health`

详细的API文档请参考: [API文档](./api_doc/)

## 快速开始

### 环境要求
- Go 1.24.0+
- OpenSearch 2.x
- 本体管理模块 (运行在13014端口)

### 本地开发

1. **克隆代码库**
```bash
git clone https://github.com/your-org/ontology-opensource.git
cd ontology-opensource/ontology-query
```

2. **配置依赖服务**
编辑 `server/config/ontology-query-config.yaml`，配置OpenSearch和本体管理模块的连接信息。

3. **安装依赖**
```bash
cd server
go mod download
```

4. **运行服务**
```bash
go run main.go
```

服务将在 http://localhost:13018 启动

### Docker 部署

1. **构建镜像**
```bash
docker build -t ontology-query:latest -f docker/Dockerfile .
```

2. **运行容器**
```bash
docker run -d -p 13018:13018 --name ontology-query ontology-query:latest
```

### Kubernetes 部署

使用提供的Helm chart进行部署：

```bash
helm3 install ontology-query helm/ontology-query/ 
```

## 配置说明

### 主要配置项

```yaml
# server/config/ontology-query-config.yaml
server:
  http_port: 13018
  read_timeout: 60
  write_timeout: 60
  language: zh-CN
  run_mode: debug
```

## 监控与运维

### 健康检查
- **健康检查端点**: `GET /health`
- **就绪检查端点**: `GET /ready`

### 日志配置
支持结构化日志输出，可配置日志级别和输出格式：

```yaml
log:
  logLevel: debug
  developMode: false
  maxAge: 100
  maxBackups: 20
  maxSize: 100
```

## 开发规范

### 代码规范
1. 遵循Go语言官方编码规范
3. 遵循清洁架构原则
4. 接口和实现分离
5. 完善的错误处理

### 查询开发
1. **查询构建**: 使用查询构建器构建复杂查询
2. **参数验证**: 严格的输入参数验证
3. **性能考虑**: 查询性能优化
4. **结果处理**: 统一的结果处理格式
5. **错误处理**: 详细的错误信息和处理

### 测试要求
1. 单元测试覆盖率 > 85%
2. 集成测试覆盖主要查询场景
3. 性能测试验证查询性能
4. 负载测试验证系统容量
5. 混沌测试验证系统稳定性

## 故障排查

### 常见问题

1. **OpenSearch连接失败**
   - 检查OpenSearch配置和连接参数
   - 验证网络连通性和防火墙设置
   - 确认OpenSearch集群健康状态
   - 检查认证和权限配置

2. **查询性能问题**
   - 检查索引配置和映射设置
   - 分析查询执行计划和复杂度
   - 优化查询语句和过滤条件
   - 调整缓存策略和参数

3. **内存使用过高**
   - 检查查询结果集大小
   - 优化聚合查询和分桶设置
   - 调整缓存大小和TTL
   - 监控垃圾回收情况

4. **查询结果不准确**
   - 检查数据同步状态
   - 验证索引数据完整性
   - 分析查询逻辑和条件
   - 检查分词器和分析器配置

### 调试工具
- **性能分析**: pprof性能分析
- **日志分析**: 运行日志分析
- **链路追踪**: 分布式链路追踪

## 版本历史

- **v0.1.0**: 初始版本，支持向量和语义搜索

## 支持与联系

- **技术支持**: AISHU ADP团队
- **文档更新**: 持续更新中
- **问题反馈**: 通过内部系统提交

---

**注意**: 这是一个高性能的查询引擎，需要根据实际业务场景进行性能调优和配置优化。