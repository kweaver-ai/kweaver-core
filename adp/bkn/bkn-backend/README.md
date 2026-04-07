# 本体管理模块 (BKN Backend)

## 模块概述

本体管理模块是本体引擎的核心组件，负责本体模型的创建、编辑、管理和维护。该模块提供了完整的本体生命周期管理功能，支持对象类、关系类、行动类的定义和管理，以及知识网络的构建和维护。

## 核心功能

### 1. 知识网络管理 (Knowledge Network Management)
- **网络创建**: 创建领域特定的知识网络
- **网络拓扑**: 管理和可视化网络拓扑结构
- **语义层**: 构建多层次的语义网络
- **网络分析**: 网络结构分析和优化

### 2. 对象类管理 (Object Type Management)
- **对象类定义**: 创建和管理知识网络中的对象类
- **属性定义**: 为对象类定义丰富的属性字段
- **图标和样式**: 自定义对象类的可视化展示

### 3. 关系类管理 (Relation Type Management)
- **关系类定义**: 创建和管理知识网络中的关系类
- **方向性**: 支持有向和无向关系
- **属性约束**: 关系类的属性约束
- **多重性**: 支持一对一、一对多、多对多关系

### 4. 行动类管理 (Action Type Management)
- **行动类定义**: 创建和管理知识网络中的行动类
- **参数配置**: 行动类的输入输出参数定义
- **执行策略**: 行动类的执行策略和规则

### 5. 概念分组管理 (Concept Group Management)
- **概念组定义**: 创建和管理知识网络中的概念分组
- **成员管理**: 管理概念分组成员

### 6. 数据源集成 (Data Source Integration)
- **多源集成**: 基于VEGA实现多种数据源的集成

### 6. 任务调度 (Job Scheduling)
- **后台任务**: 异步任务处理

## 技术架构

### 技术栈
- **编程语言**: Go 1.24
- **Web框架**: Gin 1.11
- **数据库**: MariaDB/DM8, OpenSearch
- **容器化**: Docker, Kubernetes
- **监控**: OpenTelemetry

### 架构设计
```
server/
├── common/              # 公共配置和工具
├── config/              # 配置文件
├── drivenadapters/      # 数据访问层
│   ├── action_type/     # 动作类数据访问
│   ├── business_system/ # 业务域数据访问
│   ├── concept_group/   # 概念分组数据访问
│   ├── data_model/      # 数据模型数据访问
│   ├── data_view/       # 数据视图数据访问
│   ├── job/             # 任务数据访问
│   ├── knowledge_network/ # 知识网络数据访问
│   ├── model_factory/   # 模型工厂数据访问
│   ├── object_type/     # 对象类数据访问
│   ├── opensearch/      # OpenSearch数据访问
│   ├── permission/      # 权限数据访问
│   ├── relation_type/   # 关系类数据访问
│   └── user_mgmt/       # 用户管理数据访问
├── driveradapters/      # 接口适配层
├── errors/              # 错误定义
├── interfaces/          # 接口定义
├── locale/              # 国际化支持
├── logics/              # 业务逻辑层
├── version/             # 版本信息
└── worker/              # 后台任务处理
    ├── concept_syncer/  # 概念同步任务
    └── job_executor/    # 任务执行器
```

## API 接口

### 知识网络API
- **获取知识网络列表**: `GET /api/bkn-backend/v1/knowledge-networks`
- **创建知识网络**: `POST /api/bkn-backend/v1/knowledge-networks`
- **获取网络详情**: `GET /api/bkn-backend/v1/knowledge-networks/{id}`
- **更新网络配置**: `PUT /api/bkn-backend/v1/knowledge-networks/{id}`

### 对象类API
- **获取对象类列表**: `GET /api/bkn-backend/v1/knowledge-networks/{kn_id}/object-types`
- **创建对象类**: `POST /api/bkn-backend/v1/knowledge-networks/{kn_id}/object-types`
- **更新对象类**: `PUT /api/bkn-backend/v1/knowledge-networks/{kn_id}/object-types/{id}`
- **删除对象类**: `DELETE /api/bkn-backend/v1/knowledge-networks/{kn_id}/object-types/{id}`

### 关系类API
- **获取关系类列表**: `GET /api/bkn-backend/v1/knowledge-networks/{kn_id}/relation-types`
- **创建关系类**: `POST /api/bkn-backend/v1/knowledge-networks/{kn_id}/relation-types`
- **更新关系类**: `PUT /api/bkn-backend/v1/knowledge-networks/{kn_id}/relation-types/{id}`
- **删除关系类**: `DELETE /api/bkn-backend/v1/knowledge-networks/{kn_id}/relation-types/{id}`

### 动作类API
- **获取动作类列表**: `GET /api/bkn-backend/v1/knowledge-networks/{kn_id}/action-types`
- **创建动作类**: `POST /api/bkn-backend/v1/knowledge-networks/{kn_id}/action-types`
- **更新动作类**: `PUT /api/bkn-backend/v1/knowledge-networks/{kn_id}/action-types/{id}`
- **删除动作类**: `DELETE /api/bkn-backend/v1/knowledge-networks/{kn_id}/action-types/{id}`

### 概念分组API
- **获取概念组列表**: `GET /api/bkn-backend/v1/knowledge-networks/{kn_id}/concept-groups`
- **创建概念组**: `POST /api/bkn-backend/v1/knowledge-networks/{kn_id}/concept-groups`
- **更新概念组**: `PUT /api/bkn-backend/v1/knowledge-networks/{kn_id}/concept-groups/{id}`
- **删除概念组**: `DELETE /api/bkn-backend/v1/knowledge-networks/{kn_id}/concept-groups/{id}`
- **添加概念组成员**: `POST /api/bkn-backend/v1/knowledge-networks/{kn_id}/concept-groups/{id}/members`
- **删除概念组成员**: `DELETE /api/bkn-backend/v1/knowledge-networks/{kn_id}/concept-groups/{id}/members/{member_id}`

### 任务管理API
- **获取任务列表**: `GET /api/bkn-backend/v1/jobs`
- **创建任务**: `POST /api/bkn-backend/v1/jobs`
- **获取任务详情**: `GET /api/bkn-backend/v1/jobs/{id}`
- **删除任务**: `DELETE /api/bkn-backend/v1/jobs/{id}`
- **获取子任务列表**: `GET /api/bkn-backend/v1/jobs/{id}/tasks`

### 内部接口
- 相同路径，前缀改为 `/api/ontology-query/in/v1/`
- 跳过OAuth认证

### 系统接口
- **健康检查**: `GET /health`

详细的API文档请参考: [API文档](./api_doc/)

## 快速开始

### 环境要求
- Go 1.24.0+
- MariaDB 11.4+
- OpenSearch 2.x

### 本地开发

1. **克隆代码库**
```bash
git clone https://github.com/your-org/ontology-opensource.git
cd ontology-opensource/bkn-backend
```

2. **配置数据库连接**
编辑 `server/config/bkn-backend-config.yaml`，配置数据库和OpenSearch连接信息。

3. **安装依赖**
```bash
cd server
go mod download
```

4. **运行服务**
```bash
go run main.go
```

服务将在 http://localhost:13014 启动

### Docker 部署

1. **构建镜像**
```bash
docker build -t bkn-backend:latest -f docker/Dockerfile .
```

2. **运行容器**
```bash
docker run -d -p 13014:13014 --name bkn-backend bkn-backend:latest
```

### Kubernetes 部署

使用提供的Helm chart进行部署：

```bash
helm install bkn-backend helm/bkn-backend/
```

## 配置说明

### 主要配置项

```yaml
# server/config/bkn-backend-config.yaml
server:
  http_port: 13014
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

### 测试要求
1. 单元测试覆盖率 > 80%
2. 集成测试覆盖主要业务流程
3. 性能测试验证系统性能
4. 安全测试确保系统安全

### 运行单元测试

单元测试依赖 `I18N_MODE_UT=true` 环境变量，使测试跳过 locale 文件加载：

```bash
cd server
I18N_MODE_UT=true go test ./...
```

单个包：

```bash
I18N_MODE_UT=true go test ./logics/knowledge_network/... -v
```

带覆盖率：

```bash
I18N_MODE_UT=true go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 提交规范
1. 使用语义化提交信息
2. 每个提交对应一个功能点
3. 详细的提交描述
4. 关联相关的issue

## 故障排查

### 常见问题

1. **数据库连接失败**
   - 检查数据库配置
   - 验证网络连通性
   - 确认数据库服务状态

2. **OpenSearch连接失败**
   - 检查OpenSearch配置
   - 验证证书和认证信息
   - 确认OpenSearch集群状态

### 调试工具
- **性能分析**: pprof性能分析
- **日志分析**: 运行日志分析
- **链路追踪**: 分布式链路追踪

## 版本历史

- **v0.1.0**: 初始版本，支持核心本体管理功能

## 支持与联系

- **技术支持**: AISHU ADP团队
- **文档更新**: 持续更新中
- **问题反馈**: 通过内部系统提交

---

**注意**: 这是一个企业级内部项目，请根据实际业务需求进行配置和使用。