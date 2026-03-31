# libs/go

`libs/go` 是 dataflow 服务的通用 Go 语言工具库，封装了常用的基础设施组件和业务通用逻辑，旨在提高开发效率并规范代码风格。

## 功能模块

本库包含以下核心模块：

### 1. 基础设施 (Infrastructure)
- **db**: 数据库连接管理封装 (基于 GORM)。
- **mq**: 消息队列 SDK 封装 (支持 Proton MQ, Kafka 等)。
- **pools**: 协程池或连接池管理 (基于 ants)。
- **store**: 存储层抽象接口。
- **lock**: 分布式锁实现。

### 2. 网络与通信 (Network & Communication)
- **http**: HTTP 客户端封装。
  - 支持 OAuth2 认证 (`oauth2_client.go`)。
  - 集成 OpenTelemetry 链路追踪 (`otel_http_client.go`)。
- **rest**: REST API 服务端辅助工具。
  - 基于 Gin 框架的 Server 封装 (`server.go`)。
  - 请求参数校验 (`check_req.go`, `validation.go`)。
  - 标准化错误处理与响应 (`errors.go`, `common_errors.go`)。

### 3. 可观测性 (Observability)
- **telemetry**: 遥测数据采集 SDK 封装 (OpenTelemetry)。
- **log**: 日志库封装 (基于 logrus)。

### 4. 通用工具 (Utilities)
- **errors**: 通用错误定义与处理。
- **i18n**: 国际化支持 (`go-i18n`)。
- **utils**: 其他通用辅助函数。
- **mock**: 用于单元测试的 Mock 对象。

### 开发规范

- 新增通用功能时，请确保代码具有良好的可复用性和测试覆盖率。
- 遵循项目的错误处理和日志规范。
