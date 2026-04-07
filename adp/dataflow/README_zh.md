# Dataflow (数据流处理平台)

Dataflow 是一个综合性的数据处理平台，通过可视化流程设计、代码执行和数据转换功能，使用户能够构建、编排和执行自动化数据管道。

[English Documentation](README.md)

## 概述

Dataflow 为企业数据处理需求提供完整的解决方案。无论您需要处理大量数据、转换和分析信息，还是集成复杂的数据源，Dataflow 都能高效地提供实现目标所需的工具和服务。

## 核心能力

通过可视化流程设计、代码执行和数据转换功能构建和执行自动化数据管道。

**主要特性:**
- 数据管道的可视化流程设计器
- 沙箱化 Python 代码执行
- 数据转换和分析
- 文档处理 (Word, Excel, PDF)
- OCR 和文本提取
- 定时和事件驱动执行
- 实时数据流处理

**应用场景:**
- ETL (提取、转换、加载) 管道
- 数据质量验证和清洗
- 自动化报告生成
- 文档处理和分析
- 图像和文本识别流程

## 架构

Dataflow 采用微服务架构，包含以下组件：

```
┌─────────────────────────────────────────────────────────┐
│                    前端层                                │
│  - dia-flow-web: 数据流可视化设计器                      │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                 数据处理服务                             │
│  - flow-automation: 数据流编排                           │
│  - coderunner: 沙箱化代码执行                            │
│  - flow-stream-data-pipeline: 实时流处理                 │
│  - ecron: 定时任务管理                                   │
└─────────────────────────────────────────────────────────┘
                           ↕
┌─────────────────────────────────────────────────────────┐
│                   共享库                                 │
│  - ide-go-lib: 通用 Go 库                                │
└─────────────────────────────────────────────────────────┘
```

## 服务概览

### 数据处理服务

#### [flow-automation](flow-automation/)
核心数据流编排服务，管理数据管道执行的完整生命周期。
- **语言**: Go
- **框架**: Gin
- **主要功能**: DAG 管理、执行器管理、触发器系统、数据连接

#### [coderunner](coderunner/)
沙箱化 Python 代码执行服务，用于运行自定义数据处理逻辑。
- **语言**: Python 3.9
- **主要功能**: RestrictedPython 执行、包管理、文档处理、OCR

#### [flow-stream-data-pipeline](flow-stream-data-pipeline/)
实时数据流处理管道服务。
- **主要功能**: 流处理、实时数据转换

#### [ecron](ecron/)
分布式定时任务调度和执行服务。
- **语言**: Go
- **主要功能**: 基于 Cron 的调度、即时执行、任务监控、多节点支持

### 前端应用

#### [dia-flow-web](dia-flow-web/)
用于构建数据处理流的可视化设计器。
- **技术**: 现代 Web 框架
- **功能**: 拖放式流程设计、节点配置、执行监控

### 共享库

#### [ide-go-lib](ide-go-lib/)
跨 Go 服务共享的通用 Go 库。

## 技术栈

### 后端服务
- **Go**: flow-automation, ecron, ide-go-lib
- **Python**: coderunner, flow-stream-data-pipeline

### 框架与库
- **Go**: Gin, MongoDB, Redis, Kafka
- **Python**: Tornado, RestrictedPython, pandas, SQLAlchemy

### 基础设施
- **数据库**: MongoDB, MySQL/MariaDB, Redis
- **消息队列**: Kafka, NSQ
- **容器编排**: Kubernetes (Helm)
- **认证**: OAuth2

## 快速开始

### 前置要求

- Docker 和 Docker Compose
- Kubernetes 集群（用于生产部署）
- MongoDB, MySQL/MariaDB, Redis
- Kafka 或 NSQ 消息队列

### 使用 Docker Compose 快速启动

```bash
# 克隆仓库
git clone <repository-url>
cd dataflow

# 启动所有服务
docker-compose up -d

# 访问应用
# 数据流设计器: http://localhost:3000
```

### 单个服务设置

每个服务都可以独立运行。有关具体设置说明，请参阅每个服务目录中的 README：

- [flow-automation/README.md](flow-automation/README.md)
- [coderunner/README.md](coderunner/README.md)
- [ecron/README.md](ecron/README.md)

## 部署

### Kubernetes 部署

每个服务都包含用于 Kubernetes 部署的 Helm Charts：

```bash
# 部署 flow-automation
cd flow-automation/helm
helm install flow-automation . -f values.yaml

# 部署 coderunner
cd coderunner/helm
helm install coderunner . -f values.yaml

# 部署 ecron
cd ecron/helm
helm install ecron . -f values.yaml
```

### 配置

每个服务使用环境变量或配置文件。主要配置区域：

- **数据库连接**: MongoDB, MySQL, Redis
- **消息队列**: Kafka/NSQ 端点
- **认证**: OAuth2 服务端点
- **服务发现**: 内部服务 URL

## 集成

### 外部系统集成

- **OAuth2 认证**: 与外部身份提供商集成
- **消息队列**: 连接到 Kafka/NSQ 实现事件驱动架构
- **REST API**: 所有服务都公开 RESTful API 用于集成
- **Webhook**: 配置 Webhook 接收事件通知

## 开发

### 项目结构

```
dataflow/
├── flow-automation/       # 数据流编排 (Go)
├── coderunner/           # 代码执行服务 (Python)
├── ecron/                # 定时任务 (Go)
├── flow-stream-data-pipeline/  # 流处理管道 (Python)
├── dia-flow-web/         # 数据流界面
└── ide-go-lib/          # 共享 Go 库
```

### 贡献

1. 选择您想要贡献的服务
2. 遵循服务 README 中的开发指南
3. 为您的更改编写测试
4. 提交 Pull Request

### 代码风格

- **Go**: 遵循 Go 标准规范，使用 `golangci-lint`
- **Python**: 遵循 PEP 8，使用 `black` 和 `pylint`

## 文档

- [Flow Automation 文档](flow-automation/README.md)
- [CodeRunner 文档](coderunner/README.md)
- [ECron 文档](ecron/README.md)

## 使用案例示例

### 示例 1: 自动化数据处理管道

1. 在 dia-flow-web 中设计数据流
2. 配置数据源连接
3. 添加带有自定义 Python 代码的转换节点
4. 通过 ecron 设置定时执行
5. 在 flow-automation 仪表板中监控执行

### 示例 2: 文档处理流程

1. 创建用于文档摄取的数据流
2. 添加 OCR 和文本提取节点
3. 配置数据转换和验证
4. 设置自动化报告生成
5. 监控处理结果

### 示例 3: 实时数据流处理

1. 配置流式数据源
2. 设计转换管道
3. 设置实时处理规则
4. 监控流式数据流
5. 导出处理结果

## 监控和可观测性

- **健康检查**: 所有服务都公开健康检查端点
- **指标**: Prometheus 指标用于监控
- **日志**: 所有服务的结构化日志
- **追踪**: 分布式追踪支持

## 安全

- **认证**: 基于 OAuth2 的认证
- **授权**: 基于角色的访问控制
- **代码执行**: 沙箱化执行环境
- **数据隔离**: 多租户数据隔离
- **审计追踪**: 全面的审计日志

## 支持

如有问题和支持需求：
- 查看特定服务的 README 文件
- 查阅 API 文档
- 联系开发团队