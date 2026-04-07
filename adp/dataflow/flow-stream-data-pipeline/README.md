# flow-stream-data-pipeline 流式数据管道

## 1. 项目概述

flow-stream-data-pipeline 是一个流数据处理管道系统，主要负责实时数据的传输、处理和存储。系统采用微服务架构，分为管道管理服务（pipeline-mgmt）和工作器服务（pipeline-worker）两个核心组件，通过 Kafka 进行消息传递，最终将数据写入 OpenSearch。

主要功能包括：
- 管道的创建、删除、更新和查询管理
- 实时数据消费、处理和写入
- 分布式任务调度和管理
- 异常处理和错误重试机制
- 系统监控和健康检查

## 2. 系统架构

### 2.1 整体架构

```
┌──────────────────┐      ┌───────────────┐      ┌──────────────────┐
│                  │      │               │      │                  │
│  pipeline-mgmt   │──────▶    Kafka     │◀─────▶  pipeline-worker  │
│                  │      │               │      │                  │
└────────┬─────────┘      └──────┬────────┘      └────────┬─────────┘
         │                       │                        │
         │                       │                        │
         ▼                       ▼                        ▼
┌──────────────────┐      ┌───────────────┐      ┌──────────────────┐
│  数据库存储      │      │ 消息存储      │      │  OpenSearch      │
│                  │      │               │      │                  │
└──────────────────┘      └───────────────┘      └──────────────────┘
```

### 2.2 模块划分

#### 2.2.1 pipeline-mgmt 模块

pipeline-mgmt 模块是管道管理服务，负责管道的生命周期管理。

```
server/pipeline-mgmt/
├── drivenadapters/    # 数据访问层
│   ├── pipeline_access.go      # 管道数据访问
│   ├── permission_service.go   # 权限服务
│   └── mq_adapter.go           # MQ适配器
├── driveradapters/    # 接口适配层
│   ├── pipeline_handler.go     # 管道处理器
│   └── routers.go              # 路由配置
├── interfaces/        # 接口定义层
│   ├── pipeline_service.go     # 管道服务接口
│   ├── pipeline_access.go      # 数据访问接口
│   └── mock/                   # Mock接口
├── logics/            # 业务逻辑层
│   ├── pipeline_service.go     # 管道服务实现
│   └── worker_service.go       # 工作器服务实现
└── main.go            # 服务入口
```

#### 2.2.2 pipeline-worker 模块

pipeline-worker 模块是工作器服务，负责数据的消费和处理。

```
server/pipeline-worker/
├── drivenadapters/    # 数据访问层
│   ├── bulk_indexer.go         # OpenSearch批量写入
│   └── mq_access.go            # MQ访问
├── interfaces/        # 接口定义层
│   ├── worker_service.go       # 工作器服务接口
│   ├── index_base_access.go    # 索引库访问接口
│   ├── mq_access.go            # MQ访问接口
│   └── mock/                   # Mock接口
├── logics/            # 业务逻辑层
│   ├── worker_service.go       # 工作器服务实现
│   └── task.go                 # 任务处理逻辑
└── main.go            # 服务入口
```

### 2.3 核心接口

#### 2.3.1 PipelineMgmtService 接口

定义了管道管理的核心功能：
- CreatePipeline：创建管道
- DeletePipeline：删除管道
- UpdatePipeline：更新管道
- GetPipeline：获取管道信息
- ListPipelines：列出管道
- UpdatePipelineStatus：更新管道状态

#### 2.3.2 Task 接口

定义了工作器任务的核心功能：
- Run：运行任务
- Stop：停止任务
- Execute：执行任务
- longPolling：长轮询获取消息
- processJSON：处理JSON数据

## 3. 核心功能详解

### 3.1 管道管理功能

#### 3.1.1 管道创建流程

1. 权限检查：验证用户是否有权限创建管道
2. 账户信息获取：获取创建者信息
3. 索引库校验：验证索引库是否存在
4. Kafka主题创建：创建相应的输入、输出和错误主题
5. 管道部署：部署管道资源
6. 资源策略注册：注册资源使用策略

#### 3.1.2 管道删除流程

1. 权限检查：验证用户是否有权限删除管道
2. 状态检查：检查管道状态
3. 主题删除：删除相应的Kafka主题
4. 管道信息删除：从数据库删除管道信息

### 3.2 数据处理功能

#### 3.2.1 数据消费处理流程

1. Kafka消费者创建：创建消费者连接
2. 消息拉取：使用长轮询方式拉取消息
3. 消息处理：处理消息内容，确保为JSON格式
4. 批量写入：将处理后的消息批量写入OpenSearch
5. 异常处理：处理消费和写入过程中的异常，将错误消息写入错误主题

#### 3.2.2 任务管理功能

1. 任务启动：启动指定管道的处理任务
2. 任务停止：停止正在运行的任务
3. 任务重启：重启失败的任务
4. 资源监控：监控任务的CPU和内存使用情况

## 4. 技术栈

- **开发语言**：Go
- **Web框架**：Gin
- **消息队列**：Kafka
- **搜索引擎**：OpenSearch
- **数据库**：MariaDB
- **单元测试**：GoConvey, GoMock
- **部署方式**：Docker, Kubernetes
- **CI/CD**：Azure Pipelines

## 5. 部署配置

### 5.1 环境配置

主要配置文件：`server/config/pipeline-config.yaml`

核心配置项：

```yaml
server:
  httpPort: 13012            # 服务端口
  flushMiB: 5               # 批量写入大小限制
  flushItems: 10000         # 批量写入条数限制
  flushIntervalSec: 3       # 批量写入间隔时间
  failureThreshold: 10      # 失败阈值
  watchWorkersIntervalMin: 5  # 自动恢复异常任务间隔

kafka:
  sessionTimeoutMs: 45000    # 会话超时时间
  socketTimeoutMs: 60000     # 套接字超时时间
  maxPollIntervalMs: 300000  # 最大轮询间隔
  autoOffsetReset: earliest  # 自动偏移量重置策略

mq:
  mqHost: kafka-headless.resource.svc.cluster.local.  # Kafka主机
  mqPort: 9097              # Kafka端口
  username: anyrobot        # 用户名
  password: xxxxxx   # 密码
```

### 5.2 Docker 构建

使用两阶段构建流程：
1. **builder阶段**：使用go.build镜像编译应用
2. **prod阶段**：使用ubuntu基础镜像运行应用

Dockerfile主要步骤：
- 设置构建参数（BUILD_IMAGE, BASE_IMAGE）
- 下载依赖
- 设置CGO编译配置
- 多服务构建（pipeline-mgmt, pipeline-worker）

### 5.3 CI/CD流程

Azure Pipelines配置（azure-pipelines.yaml）：

1. **变量初始化**：设置SDP_VERSION等变量
2. **代码检查**：
   - 单元测试
   - SonarQube代码分析
   - 质量门禁检查
3. **镜像构建**：
   - amd64/arm64双架构编译
   - Docker镜像构建和推送
4. **Helm Chart部署**：
   - 修改Chart配置
   - 打包并推送Chart包

## 6. 测试覆盖情况

### 6.1 单元测试

项目包含多个测试文件，覆盖了核心功能：

1. **pipeline-mgmt测试**：
   - pipeline_handler_test.go：测试管道处理器的HTTP接口
   - pipeline_access_test.go：测试管道数据访问层

2. **pipeline-worker测试**：
   - task_test.go：测试任务处理逻辑
   - bulk_indexer_test.go：测试批量索引功能

### 6.2 测试框架

- **测试框架**：GoConvey
- **Mock工具**：GoMock
- **打桩工具**：gomonkey

## 7. 监控与维护

### 7.1 日志配置

```yaml
log:
  logLevel: info          # 日志级别
  developMode: false      # 开发模式
  maxAge: 100             # 日志保留天数
  maxBackups: 20          # 最大备份数
  maxSize: 100            # 单文件最大大小(MB)
```

### 7.2 健康检查

系统提供了健康检查接口，返回服务信息：
- 服务名称
- 服务版本
- 语言版本
- Go版本和架构

### 7.3 常见问题排查

1. **Kafka连接问题**：检查mq配置和网络连接
2. **OpenSearch写入失败**：检查索引库配置和权限
3. **任务异常停止**：查看日志中的错误信息，检查failureThreshold配置
4. **资源使用过高**：调整部署配置中的CPU和内存限制

## 8. 后续优化建议

1. **完善文档**：补充详细的API文档和使用说明
2. **增加集成测试**：完善端到端测试用例
3. **优化性能**：
   - 调整批量写入参数，提高吞吐量
   - 优化错误重试机制，减少资源浪费
4. **增强监控**：
   - 增加更详细的指标监控
   - 添加告警机制
5. **安全加固**：
   - 加密敏感配置信息
   - 加强权限控制粒度

## 9. 总结

flow-stream-data-pipeline 是一个流数据处理系统，采用微服务架构，实现了管道管理和数据处理的核心功能。系统通过Docker容器化部署，支持CI/CD流程，具有良好的扩展性和可维护性。后续维护需要关注系统性能、资源使用和错误处理，确保数据处理的稳定性和可靠性。