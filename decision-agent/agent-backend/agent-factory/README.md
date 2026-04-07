# Agent-factory

## OpenAPI 文档维护快速入口

如果你当前关注的是 OpenAPI 文档生成和维护，建议先看下面两个精简版：

- `cmd/openapi-docs` 精简版：查看 [cmd/openapi-docs/README.simple.md](./cmd/openapi-docs/README.simple.md)
- `internal/openapidoc` 精简版：查看 [internal/openapidoc/README.simple.md](./internal/openapidoc/README.simple.md)

如果需要完整说明，再看完整版：

- `cmd/openapi-docs` 完整版：查看 [cmd/openapi-docs/README.md](./cmd/openapi-docs/README.md)
- `internal/openapidoc` 完整版：查看 [internal/openapidoc/README.md](./internal/openapidoc/README.md)

## 项目简介

Agent-Factory 是一个基于 Go 语言开发的智能体（Agent）配置管理服务，采用领域驱动设计（DDD）架构模式，主要负责智能体的创建、配置管理、运行调试、发布管理等核心功能。

### 核心业务领域
- **智能体配置管理**：Agent 的创建、更新、删除、发布
- **智能体模板管理**：基于配置的模板化管理
- **智能体应用广场**：Agent App 的发布、分类、搜索
- **空间管理**：空间隔离和资源管理
- **产品集成**：与 AnyShare、DIP 等产品的集成

## 技术架构

### 架构模式
项目采用 **六边形架构（Hexagonal Architecture）** 结合 **DDD（领域驱动设计）**：

```
┌─────────────────────────────────────────────────────────────┐
│                    Driver Adapters                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   API       │  │     MQ      │  │    Task     │         │
│  │ (HTTP/REST) │  │ (Message    │  │ (Scheduled  │         │
│  │             │  │  Queue)     │  │   Tasks)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Entity    │  │   Service   │  │ Value Object│         │
│  │             │  │             │  │             │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
│  ┌─────────────┐  ┌─────────────┐                          │
│  │ Aggregate   │  │    E2P/P2E  │                          │
│  │             │  │             │                          │
│  └─────────────┘  └─────────────┘                          │
└─────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────────────────────────────────────┐
│                   Driven Adapters                          │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  DB Access  │  │ HTTP Access │  │   Others    │         │
│  │ (Database)  │  │ (External   │  │             │         │
│  │             │  │  Services)  │  │             │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### 目录结构

```
src/
├── boot/                    # 系统启动和初始化
├── domain/                  # 领域层（DDD核心）
│   ├── entity/             # 实体对象
│   ├── service/            # 领域服务
│   ├── aggregate/          # 聚合
│   ├── valueobject/        # 值对象
│   ├── e2p/               # 实体到持久化转换
│   ├── p2e/               # 持久化到实体转换
│   ├── enum/              # 枚举定义
│   └── types/             # 类型定义
├── port/                   # 端口层（接口定义）
│   ├── driven/            # 被驱动端口（由领域定义，适配器实现）
│   └── driver/            # 驱动端口（由领域实现，适配器调用）
├── drivenadapter/         # 被驱动适配器
│   ├── dbaccess/          # 数据库访问
│   └── httpaccess/        # HTTP外部服务访问
├── driveradapter/         # 驱动适配器
│   ├── api/               # HTTP API
│   ├── mq/                # 消息队列
│   └── task/              # 定时任务
└── infra/                 # 基础设施层
    ├── persistence/       # 持久化相关
    ├── common/           # 通用工具
    └── server/           # 服务器配置
```

## 核心组件

### 领域模型

#### 主要实体（Entity）
- **DataAgent**：智能体配置实体
- **DataAgentTpl**：智能体模板实体
- **Space**：空间实体
- **SpaceMember**：空间成员实体
- **SpaceResource**：空间资源实体
- **Release**：发布实体
- **Category**：分类实体

#### 聚合设计
- **Agent聚合**：以DataAgent为聚合根，包含配置、模板等
- **Space聚合**：以Space为聚合根，包含成员、资源管理
- **Release聚合**：以Release为聚合根，包含发布历史、权限等

### 数据持久化

#### 数据库设计
主要数据表包括：

```sql
-- 智能体相关
t_data_agent_config      -- 智能体配置表
t_data_agent_config_tpl  -- 智能体模板表

-- 空间相关
t_space                  -- 空间主表
t_space_member          -- 空间成员关联表
t_space_resource        -- 空间资源关联表

-- 产品
t_product               -- 产品表
```

#### Repository模式
每个聚合都有对应的Repository接口和实现：
- `IDataAgentConfigRepo` / `DAConfigRepo`
- `IDataAgentTplRepo` / `DAConfigTplRepo`
- `ISpaceRepo` / `SpaceRepo`
- `ISpaceMemberRepo` / `SpaceMemberRepo`
- `ISpaceResourceRepo` / `SpaceResourceRepo`

### API设计

#### 接口规范
- **外部接口**：`/agent-factory/v3/{resource}`
- **内部接口**：`/agent-factory/internal/v3/{resource}`

#### 主要API模块
- **Agent配置管理**：`v3agentconfighandler`
- **产品管理**：`producthandler`
- **分类管理**：`categoryhandler`
- **发布管理**：`releasehandler`
- **应用广场**：`squarehandler`

## 技术栈

### 核心依赖
```go
// Web框架
github.com/gin-gonic/gin v1.10.0

// 数据库
devops.aishu.cn/AISHUDevOps/ONE-Architecture/_git/proton-rds-sdk-go

// Redis
github.com/go-redis/redis/v8 v8.11.5

// OpenAI集成
github.com/sashabaranov/go-openai v1.17.9

// 测试框架
go.uber.org/mock v0.5.1

// 错误处理
github.com/pkg/errors v0.9.1

// JSON处理
github.com/json-iterator/go v1.1.12
```

### 中间件和组件
- **认证中间件**：OAuth验证
- **日志中间件**：请求日志记录
- **恢复中间件**：Panic恢复
- **国际化**：多语言支持
- **链路追踪**：分布式追踪

## 部署和运维

### 容器化部署
- **基础镜像**：Ubuntu 22.04
- **端口**：13020

### Kubernetes部署
- **Helm Chart**：标准化部署
- **健康检查**：`/health/ready` 和 `/health/alive`
- **配置管理**：ConfigMap和Secret
- **服务发现**：K8s Service

### CI/CD流程
- **代码检查**：golangci-lint
- **单元测试**：go test with coverage
- **安全扫描**：Trivy镜像扫描
- **质量门禁**：SonarQube代码质量检查

## 代码质量

运行代码格式化和检查：

```bash
make lint
```

该命令会自动执行代码格式化和 golangci-lint 检查。

## 名词解释

| 术语名称                      | 定义                                                                            |
|---------------------------|-------------------------------------------------------------------------------|
| Agent（智能体）                | 一种借助大模型、知识库、记忆、工具、数据流等多种基础功能，实现智能化的交互、决策、任务执行的自主实体，通过 Agent 配置进行创建，发布后可以被用户使用 |
| Agent App Market（智能体应用市场） | 一个用于查找、使用 Agent App 的入口，支持按照空间、分类等维度显示 Agent App，也可以查看历史使用的 Agent App         |
| Agent Template（智能体模板）     | 基于 Agent 配置保存的配置模板，模板支持导入导出，可以基于模板快速创建 Agent App                              |
| Agent App（智能体应用）          | 基于一个或多个Agent 构建的完整的应用或系统，通常结合业务逻辑、多 Agent 协作提供特定产品能力的、解决实际场景的复杂问题             |
| Agent配置 - 基本信息            | 包括 Agent 头像、名称、简介信息                                                           |
| Agent配置 - 所属产品            | 产品用于控制 Agent 数据源范围、是否支持临时区等设置                                                 |
| Agent配置 - 人设&指令           | 描述AI助手的角色定位、专业领域、回复风格以及技能使用规则                                                 |
| Agent配置 - 开场白             | Agent对话开始时的首条自动回复                                                             |
| Agent配置 - 预设问题            | 提前设置的一系列问题，用于引导用户对话                                                           |
| Agent配置 - 输入配置            | 配置Agent的输入相关设置                                                                |
| Agent配置 - 知识库             | 支持AnyShare文档、Wiki、知识图谱、本体作为知识来源，Agent 可以通过知识库获取各种领域知识，实现准确推理、保证对话质量、提升效率和性能   |
| Agent配置 - 临时区             | 支持在对话过程中临时保存用户输入的临时数据、上下文、中间结果等信息，以便在生成回复时参考，临时区数据具有较短的生命周期，随着会话、任务结束而清理      |
| Agent配置 - 技能              | 支持将其他 Agent 或工具作为 Agent 的技能用于完成特定任务或实现特定功能                                    |

## DDD相关概念说明

| 英文                  | 缩写                | 中文    | 说明                                     |
|---------------------|-------------------|-------|----------------------------------------|
| persistence object  | po                | 持久化对象 | 与数据库表结构一一对应的对象，用于数据库表的增删改查操作，属于持久化层    |
| entity              | eo(entity object) | 实体    | 具有唯一标识的领域对象，即使属性相同，不同实体也被视为不同对象        |
| value object        | vo                | 值对象   | 没有唯一标识的领域对象，通过其属性值来识别，相同属性值的值对象被视为同一对象 |
| aggregate           | -                 | 聚合    | 由根实体、值对象和其他实体组成的一组相关对象的集合，作为一个整体被外界访问  |
| aggregate root      | -                 | 聚合根   | 聚合的根实体，是外部访问聚合内部对象的唯一入口                |
| repository          | repo              | 仓储    | 提供对聚合的持久化和查询能力，隐藏数据访问细节                |
| domain service      | -                 | 领域服务  | 表示领域中的一个操作，这个操作不属于任何实体或值对象             |
| application service | -                 | 应用服务  | 协调多个领域对象完成用户用例，是应用层与领域层的边界             |

## 缩写约定

| 缩写     | 全称                                        | 说明                                        |备注|
|--------|-------------------------------------------|-------------------------------------------|----|
| da     | data agent                                | 数据智能体                                     |
| um     | user_management                           | 用户管理服务相关                                  |
| cmp    | component                                 | 组件                                        |
| dto    | data transfer object                      | 数据传输对象                                    |在不同层或组件间传输数据的简单对象，只包含数据属性和简单的获取/设置方法，不包含业务逻辑|
| rdto   | request and response data transfer object | 请求和响应数据传输对象                               |
| d2e    | dto to entity                             | 数据传输对象到实体（当dto为rdto的req时，指将rdto转换为entity） |
| ds     | datasource                                | 数据源                                       |data agent的数据源|
| assoc  | association                               | 关联                                        |data agent和数据集的关联|
| obj    | object                                    | 对象                                        |数据源中的对象|
| obj_id | object id                                 | 对象id                                      |“文档”对象的唯一标识|
|pa     | published agent                           | 已发布  agent                                      | |
|ptpl   | published agent template                  | 已发布  agent 模板                                      | |
|sr     | space resource                            | 空间资源                                      | |
|cs     | custom space                              | 自定义空间                                      | |

## 接口定义规范

### 格式规范
接口路径采用三级分段结构， 格式为：
```
/服务名/版本/Restful资源对象
```
比如 `/agent-factory/v3/agent`

### 内外接口区分规范
1. 外部接口（前端/外部系统访问）: 
  - 用途： 提供给前端或第三方系统访问的公共接口， 通过 `Ingress` 暴漏服务
  - 路径格式： `/服务名/版本/Restful资源对象`
  - 访问方式： 通过 Ingress 配置的域名或者 IP 访问
2. 内部接口（后端微服务间调用）
  - 用途： 通过 K8S Service 暴漏服务，不对外暴漏
  - 路径格式： 在基础格式前增加 `internal`前缀， `/服务名/internal/版本/Restful资源对象`
  - 访问方式： 通过集群内 Service 名称或 DNS 访问
