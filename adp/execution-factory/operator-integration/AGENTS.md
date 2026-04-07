# Operator Integration 开发指南 (AGENTS.md)

## 1. 快速上手 (Quick Start)

### 环境要求
- **Go**: 1.23+
- **Docker**: 用于构建镜像
- **Helm**: (可选) 用于部署模板渲染
- **ktctl**: (可选) 用于远程连接开发环境

### 常用命令
项目根目录下的 `project.sh` 是主要的操作入口：

- **构建并运行 (Build & Run)**
  ```bash
  ./project.sh -b
  ```
  > **注意**: 此命令会自动将 `server/infra/config` 下的配置文件覆盖到 `/sysvol/config` 和 `/sysvol/secret`，并启动服务。

- **运行测试 (Test)**
  ```bash
  ./project.sh -t
  ```
  运行所有单元测试（自动排除 `server/tests` 和 `server/mocks`）。

- **生成 Mock (Generate Mocks)**
  ```bash
  ./project.sh -m
  ```
  基于 `go generate` 生成 Mock 对象，用于单元测试。

- **代码检查 (Lint)**
  ```bash
  ./project.sh -l
  ```
  运行 `golangci-lint`，提交前必须通过。

- **API 文档预览**
  ```bash
  ./project.sh -p
  ```

## 2. 架构概览 (Architecture)

本项目采用 **六边形架构 (Hexagonal Architecture / Ports and Adapters)**，核心在于分离业务逻辑与基础设施。

### 目录结构映射
| 目录 | 职责 (Role) | 说明 |
|------|-------------|------|
| `server/interfaces` | **Ports** | 定义入站和出站接口，声明核心业务契约。 |
| `server/logics` | **Core** | 核心业务逻辑实现，纯净的 Go 代码，不依赖具体外部技术。 |
| `server/driveradapters` | **Inbound Adapters** | 驱动层，如 HTTP Handler (Gin)、MQ Consumer。 |
| `server/drivenadapters` | **Outbound Adapters** | 被驱动层，如 DB Client、External API Client。 |
| `server/dbaccess` | **DA Layer** | 数据库访问层具体实现。 |
| `server/infra` | **Infrastructure** | 基础设施，包括配置 (`config`)、日志、可观测性 (`telemetry`)。 |
| `server/main.go` | **Entry Point** | 程序入口，负责依赖注入和启动。 |

## 3. 开发流程 (Workflow)

### 分支管理
- **功能开发**: `feature/your-feature-name`
- **Bug 修复**: `fix/bug-description`
- **文档修改**: `docs/update-description`

### 提交信息规范
遵循 [Conventional Commits](https://www.conventionalcommits.org/)：
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档变更
- `style`: 代码格式（不影响逻辑）
- `refactor`: 重构（无功能变更）
- `test`: 测试用例
- `chore`: 构建/工具链变动

**示例**: `feat:(execution-factory)add operator execution logic`

### 代码质量
- **Lint**: 提交前请运行 `./project.sh -l` 确保无 Lint 错误。
- **Test**: 新增逻辑必须编写单元测试，运行 `./project.sh -t` 验证。

### 调试技巧
- **开启 Debug 日志**: 修改 `server/infra/config/agent-operator-integration.yaml` 中的 `project.debug` 为 `true`，然后重新运行 `./project.sh -b`。

## 4. 部署与运维 (Deployment)

### Docker 构建
```bash
docker build -t agent-operator-integration:latest .
```

### Helm 部署
使用 Helm Chart 部署到 Kubernetes：
```bash
helm install agent-operator-integration ./helm/agent-operator-integration
```
本地渲染模板检查：
```bash
./project.sh -d
```

## 5. 安全规范 (Security)

- **敏感配置**:
  - **严禁**将 API Key、Password 等敏感信息提交到 Git。
  - 本地开发请使用 `/sysvol/secret` 下的配置文件或环境变量。
  - 生产环境敏感配置文件权限应设置为 `600`。
- **API 认证**:
  - 内部微服务调用必须携带 `x-account-id` 和 `x-account-type` Header。
  - 外部接口需遵循统一认证网关规范。
- **日志脱敏**: 禁止在日志中打印 Token、Password 或用户敏感数据。
