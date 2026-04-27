# AGENTS.md - Agent Factory 开发指南

## 项目概述

Agent Factory 是一个基于 Go 的微服务，用于管理智能代理配置，采用领域驱动设计（DDD）和六边形架构。

## 开发命令

### 代码质量
```bash
make fmt               # 使用 gofumpt 和 goimports 格式化代码
make wsl               # 运行空白字符检查器（wsl）
make all               # 运行所有格式化（wsl + fmt）
make ciLint            # 运行 golangci-lint 检查
make ciLintFix         # 运行 golangci-lint 自动修复
```

### 测试
```bash
make goTest            # 运行所有测试并输出详细信息
go test -v ./...       # 运行所有测试
go test ./path/to/package -run TestSpecificFunction         # 运行包中的单个测试
go test -v ./... -run TestSpecificFunction                 # 跨所有包运行单个测试
go test ./... -run TestSpecificFunction/SubTest             # 运行特定的子测试
make ut                # 运行单元测试并生成覆盖率报告（生成 coverage_report/）
```

### 构建和运行
```bash
make localRun          # 使用开发配置在本地构建和运行
make reRunLocal        # 终止并重启本地实例
make killLocal         # 终止端口 13020 上的本地实例
make goGenerate        # 运行 go generate ./... 生成 mock 和代码
```

## 代码风格指南

### 导入组织
顺序：标准库、第三方库、本地库。使用 goimports。

```go
import (
	"context"
	"database/sql"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/entity"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
)
```

### 命名规范
- **包名**: `snake_case` 带功能后缀：`dbacc`、`svc`、`repo`、`handler`、`valobj`
- **导出**: 类型、函数、常量使用 `PascalCase`
- **非导出**: 私有字段和方法使用 `camelCase`
- **文件**: `snake_case.go`，描述性名称如 `conversation_repo.go`、`define.go`
- **缩写**: 类型名全大写（如 `IConversationRepo`），变量名使用驼峰（如 `id`）

### DDD 目录结构
```
src/
├── boot/              # 初始化（配置、数据库、Redis、日志）
├── domain/            # 领域层（业务逻辑）
│   ├── entity/        # 实体对象（后缀：eo）
│   ├── service/       # 领域服务（后缀：svc）
│   ├── valueobject/   # 值对象（后缀：vo）
│   ├── aggregate/     # 聚合
│   ├── enum/          # 枚举
│   ├── e2p/           # 实体到持久化转换
│   └── p2e/           # 持久化到实体转换
├── port/              # 接口定义（依赖倒置）
│   ├── driven/        # 适配器实现的接口（idbaccess、ihttpaccess）
│   └── driver/        # 领域实现的接口（iportdriver、iv3portdriver）
├── drivenadapter/     # 基础设施适配器
│   ├── dbaccess/      # 数据库访问实现（后缀：dbacc）
│   ├── httpaccess/    # HTTP 客户端实现
│   └── redisaccess/   # Redis 访问实现
├── driveradapter/     # API 层
│   ├── api/           # HTTP API 处理器（后缀：handler）
│   └── rdto/          # 请求/响应 DTO
└── infra/             # 通用基础设施（工具、全局变量、辅助类）
```

### Repository 模式
在 `port/driven/idbaccess/` 中定义接口，在 `drivenadapter/dbaccess/` 中实现，使用 `sync.Once` 单例模式。嵌入 `IDBAccBaseRepo` 以获得基础方法。

```go
type IConversationRepo interface {
	IDBAccBaseRepo
	Create(ctx context.Context, po *dapo.ConversationPO) (*dapo.ConversationPO, error)
	GetByID(ctx context.Context, id string) (*dapo.ConversationPO, error)
}

type ConversationRepo struct {
	idbaccess.IDBAccBaseRepo
	db *sqlx.DB
	logger icmp.Logger
}

var repoOnce sync.Once
var repoImpl idbaccess.IConversationRepo

func NewConversationRepo() idbaccess.IConversationRepo {
	repoOnce.Do(func() {
		repoImpl = &ConversationRepo{db: global.GDB, logger: logger.GetLogger()}
	})
	return repoImpl
}
```

### Service 模式
领域服务嵌入 `service.SvcBase`，使用 DTO 模式进行依赖注入，通过 `NewXxxService(dto)` 创建。

```go
type agentSvc struct {
	service.SvcBase
	repo   idbaccess.IAgentRepo
	logger icmp.Logger
}

type NewAgentSvcDto struct {
	SvcBase service.SvcBase
	Repo    idbaccess.IAgentRepo
}

func NewAgentService(dto *NewAgentSvcDto) *agentSvc {
	return &agentSvc{SvcBase: dto.SvcBase, Repo: dto.Repo}
}
```

### 错误处理
使用 `github.com/pkg/errors` 进行错误包装：`return errors.Wrap(err, "context")`。

### HTTP 处理器模式
处理器嵌入服务接口，通过 `RegPubRouter(router)` 注册公共路由或 `RegPriRouter(router)` 注册内部路由，使用 `sync.Once` 单例模式。使用中间件进行权限和认证检查。

```go
type observabilityHandler struct {
	svc    iportdriver.IObservability
	logger icmp.Logger
}

func (h *observabilityHandler) RegPubRouter(router *gin.RouterGroup) {
	permissionRouter := router.Group("", apimiddleware.CheckAgentUsePms())
	permissionRouter.POST("/observability/agent/:agent_id/detail", h.AgentDetail)
}
```

## 测试模式

### 表驱动测试
使用包含 name、input、expected 字段的结构体切片，使用 `t.Run()` 进行子测试迭代。测试成功和错误场景。

### 测试框架
- 标准 `go test` 配合 testify 断言
- Mock 生成：`go generate ./...`（使用 mockgen 和 `//go:generate` 指令）
- 覆盖率：`make ut` 在 `coverage_report/` 中生成报告

### 本地开发模式测试策略
**不要在单元测试中测试本地开发模式的代码路径。**

**原因：**
- 本地开发模式使用 `cenvhelper.IsLocalDev()` 检查环境变量
- 测试中修改环境变量会导致竞态条件，即使不使用 `t.Parallel()`
- Go 测试框架内部会并行运行测试包，导致不可预测的测试失败

**最佳实践：**
- 只编写非本地开发模式（生产代码路径）的测试
- 在 `TestMain` 中配置为不设置 `AGENT_FACTORY_LOCAL_DEV=true`
- 使用真实的 mock 对象（如 `mockUmHttp`）而不是 `nil` 参数
- 这确保了稳定、可靠的测试，没有竞态条件

**示例：**
```go
// ✅ 正确：配置 TestMain 在非本地开发模式下运行
func TestMain(m *testing.M) {
    os.Setenv("SERVICE_NAME", "AGENT_FACTORY")
    // 注意：不要设置 AGENT_FACTORY_LOCAL_DEV=true
    // 我们只测试非本地开发模式（生产模式）以避免环境变量竞态条件
    os.Setenv("I18N_MODE_UT", "true")
    
    cenvhelper.InitEnvForTest()
    locale.Register()
    
    code := m.Run()
    os.Exit(code)
}

// ✅ 正确：使用 mock 测试生产模式
func TestFunction_NonLocalDevMode(t *testing.T) {
    ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese)
    
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
    
    // 设置期望
    mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(userMap, nil)
    
    result := SomeFunction(ctx, data, mockUmHttp)
    // 测试生产行为
}

```

## 重要说明

- 提交前务必运行 `make all`
- 服务/仓库实例使用 `sync.Once` 单例模式
- 遵循六边形架构：端口接口分离领域和基础设施
- 使用上下文传播进行请求追踪
- OpenTelemetry 追踪：`o11y.StartInternalSpan(ctx)` 和 `o11y.EndSpan(ctx, err)`
- 本地开发检测：`cenvhelper.IsLocalDev()`
- 端口号：13020
