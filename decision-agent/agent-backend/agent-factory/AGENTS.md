# AGENTS.md - Agent Factory Development Guide

## Project Overview

Agent Factory is a Go-based microservice for managing intelligent agent configurations, using Domain-Driven Design (DDD) with Hexagonal Architecture.

## Development Commands

### Code Quality
```bash
make fmt               # Format code with gofumpt and goimports
make wsl               # Run whitespace linter (wsl)
make all               # Run all formatting (wsl + fmt)
make ciLint            # Run golangci-lint checks
make ciLintFix         # Run golangci-lint with auto-fix
```

### Testing
```bash
make goTest            # Run all tests with verbose output
go test -v ./...       # Run all tests
go test ./path/to/package -run TestSpecificFunction         # Single test in package
go test -v ./... -run TestSpecificFunction                 # Single test across all packages
go test ./... -run TestSpecificFunction/SubTest             # Run specific subtest
make ut                # Run unit tests with coverage report (generates coverage_report/)
```

### Building & Running
```bash
make localRun          # Build and run locally with dev config
make reRunLocal        # Kill and restart local instance
make killLocal         # Kill local instance on port 13020
make goGenerate        # Run go generate ./... for mocks and codegen
```

## Code Style Guidelines

### Import Organization
Order: standard library, third-party, local. Use goimports.

```go
import (
	"context"
	"database/sql"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/entity"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess"
)
```

### Naming Conventions
- **Packages**: `snake_case` with functional suffixes: `dbacc`, `svc`, `repo`, `handler`, `valobj`
- **Exported**: `PascalCase` for types, functions, constants
- **Unexported**: `camelCase` for private fields and methods
- **Files**: `snake_case.go`, descriptive names like `conversation_repo.go`, `define.go`
- **Acronyms**: Full uppercase for type names (e.g., `IConversationRepo`), camelCase for variables (e.g., `id`)

### DDD Directory Structure
```
src/
├── boot/              # Initialization (config, DB, Redis, logging)
├── domain/            # Domain layer (business logic)
│   ├── entity/        # Entity objects (suffix: eo)
│   ├── service/       # Domain services (suffix: svc)
│   ├── valueobject/   # Value objects (suffix: vo)
│   ├── aggregate/     # Aggregates
│   ├── enum/          # Enumerations
│   ├── e2p/           # Entity to persistence conversion
│   └── p2e/           # Persistence to entity conversion
├── port/              # Interface definitions (dependency inversion)
│   ├── driven/        # Interfaces for adapters to implement (idbaccess, ihttpaccess)
│   └── driver/        # Interfaces for domain to implement (iportdriver, iv3portdriver)
├── drivenadapter/     # Infrastructure adapters
│   ├── dbaccess/      # Database access implementations (suffix: dbacc)
│   ├── httpaccess/    # HTTP client implementations
│   └── redisaccess/   # Redis access implementations
├── driveradapter/     # API layer
│   ├── api/           # HTTP API handlers (suffix: handler)
│   └── rdto/          # Request/response DTOs
└── infra/             # Common infrastructure (utils, global, helpers)
```

### Repository Pattern
Define interface in `port/driven/idbaccess/`, implement in `drivenadapter/dbaccess/`, use singleton with `sync.Once`. Embed `IDBAccBaseRepo` for base methods.

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

### Service Pattern
Domain services embed `service.SvcBase`, use DTO pattern for dependency injection, created via `NewXxxService(dto)`.

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

### Error Handling
Use `github.com/pkg/errors` for wrapping: `return errors.Wrap(err, "context")`.

### HTTP Handler Pattern
Handlers embed service interfaces, register routes via `RegPubRouter(router)` for public or `RegPriRouter(router)` for internal, use `sync.Once` singleton pattern. Use middleware for permissions and auth checks.

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

## Testing Patterns

### Table-Driven Tests
Use struct slices with name, input, expected fields, iterate with `t.Run()` for subtests. Test both success and error scenarios.

### Testing Framework
- Standard `go test` with testify assertions
- Mock generation: `go generate ./...` (uses mockgen with `//go:generate` directives)
- Coverage: `make ut` generates reports in `coverage_report/`

### Local Development Mode Testing Policy
**Do NOT test local development mode code paths in unit tests.**

**Rationale:**
- Local dev mode uses `cenvhelper.IsLocalDev()` which checks environment variables
- Environment variable modifications in tests cause race conditions even without `t.Parallel()`
- Go's test framework runs packages in parallel internally, causing unpredictable test failures

**Best Practice:**
- Only write tests for non-local-dev mode (production code paths)
- Configure `TestMain` to NOT set `AGENT_FACTORY_LOCAL_DEV=true`
- Use real mock objects (e.g., `mockUmHttp`) instead of `nil` parameters
- This ensures stable, reliable tests without race conditions

**Example:**
```go
// ✅ GOOD: Configure TestMain to run in non-local-dev mode
func TestMain(m *testing.M) {
    os.Setenv("SERVICE_NAME", "AGENT_FACTORY")
    // Note: Do NOT set AGENT_FACTORY_LOCAL_DEV=true
    // We only test non-local-dev (production) mode to avoid environment variable race conditions
    os.Setenv("I18N_MODE_UT", "true")
    
    cenvhelper.InitEnvForTest()
    locale.Register()
    
    code := m.Run()
    os.Exit(code)
}

// ✅ GOOD: Testing production mode with mock
func TestFunction_NonLocalDevMode(t *testing.T) {
    ctx := context.WithValue(context.Background(), cenum.VisitLangCtxKey.String(), rest.SimplifiedChinese)
    
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    mockUmHttp := httpaccmock.NewMockUmHttpAcc(ctrl)
    
    // Setup expectations
    mockUmHttp.EXPECT().GetOsnNames(ctx, gomock.Any()).Return(userMap, nil)
    
    result := SomeFunction(ctx, data, mockUmHttp)
    // Test production behavior
}

```

## Important Notes

- Always run `make all` before committing
- Use singleton pattern with `sync.Once` for service/repo instances
- Follow hexagonal architecture: port interfaces separate domain from infrastructure
- Use context propagation for request tracing
- OpenTelemetry tracing: `o11y.StartInternalSpan(ctx)` and `o11y.EndSpan(ctx, err)`
- Local dev detection: `cenvhelper.IsLocalDev()`
- Port number: 13020
