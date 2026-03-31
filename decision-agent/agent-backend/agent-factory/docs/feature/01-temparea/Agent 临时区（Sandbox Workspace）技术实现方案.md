## 0. 目标与范围（再次明确）

### 0.1 核心目标
在 **Decision Agent Platform** 中实现 **Agent 临时区（Sandbox Workspace）**，用于：

1. 用户在指定 Agent + Conversation 下上传文件（通过 Sandbox Platform 接口）
2. 用户可查看文件列表、下载文件（通过 Sandbox Platform 接口）
3. Agent 在对话中基于文件路径生成代码并在 Sandbox Platform 内执行
4. Agent 与 Sandbox 的交互由 **算子平台注册的工具**完成
5. 整个过程采用 **LLM React loop**，支持流式输出

### 0.2 约束
+ 文件作用域：`(agent_id, conversation_id)`
+ 文件路径必须是 **真实沙箱路径**（Agent 生成的代码必须可执行）
+ 仅展示用户上传文件（Agent 生成文件不展示）
+ Conversation 删除立即清理文件，不可恢复
+ 跨 Conversation 文件引用不支持
+ 存储与沙箱内部对象存储不暴露给外部系统

---

# 1. C4 架构设计（概念级）

## 1.1 Context 图

```mermaid
graph TD
  User["User (Browser/SDK)"]
  DAP[Decision Agent Platform]
  OP[Operator Platform]
  SP[Sandbox Platform]

  User -->|Agent Conversation UI| DAP
  DAP -->|Invoke Tools| OP
  OP -->|Workspace/Execution APIs| SP

  SP -->|Internal Object Storage| SPStorage[Sandbox Internal Storage]
  SP -->|Internal Metadata DB| SPDB[Metadata DB]
```

**Decision Agent Platform（DAP）**

+ 负责 Agent 对话与 React Loop
+ 负责与 Operator Platform 交互
+ 管理 Sandbox Session 生命周期（直接调用 Sandbox Platform）
+ 在 Chat 流程中自动检查/创建 Sandbox Session

**Operator Platform（OP）**

+ 工具注册中心（Tool Registry）
+ 代理工具调用（Invoke）
+ 将沙箱能力封装成工具，供 DAP 调用

**Sandbox Platform（SP）**

+ 负责 Session、Execution、File Upload/List/Download（接口已存在）
+ 对外暴露的能力只通过接口，不暴露内部存储

---

## 1.2 Container 图

```mermaid
graph TD
  subgraph DAP [Decision Agent Platform]
    DAP_API[API Gateway / Frontend]
    DAP_Conv[Conversation Service]
    DAP_Agent["Agent Orchestrator (LLM React Loop)"]
  end

  subgraph OP [Operator Platform]
    OP_Reg[Tool Registry]
    OP_Invoke[Tool Invocation]
    OP_Auth[Tool Auth & Permission]
  end

  subgraph SP [Sandbox Platform]
    SP_Work["Workspace Service (upload/list/download)"]
    SP_Exec[Execution Service]
    SP_Internal["Internal Storage (hidden)"]
    SP_DB[Metadata DB]
  end

  User["User (Browser/SDK)"]

  %% User interactions
  User -->|Open Agent UI| DAP_API
  User -->|Upload/List/Download files| SP_Work

  %% DAP internal flows
  DAP_API --> DAP_Conv
  DAP_Conv --> DAP_Agent

  %% DAP -> OP tool calls
  DAP_Agent -->|"tool call (with sandbox_session_id)"| OP_Invoke
  OP_Invoke --> OP_Reg
  OP_Invoke -->|call SP execution API| SP_Exec

  %% SP internal dependencies
  SP_Work --> SP_DB
  SP_Work --> SP_Internal
  SP_Exec --> SP_DB
  SP_Exec --> SP_Internal
```

---

## 1.3 Chat 流程中 Sandbox Session 管理图

```mermaid
sequenceDiagram
    participant User as User (Browser)
    participant DAP as Decision Agent Platform
    participant SP as Sandbox Platform
    participant OP as Operator Platform

    User->>DAP: Chat Request (user_id, agent_id, conversation_id)
    
    Note over DAP: 生成 session_id = sb-session-{user_id}
    
    Note over DAP,SP: 1. 检测 Sandbox Session 状态
    DAP->>SP: GET /api/v1/sessions/{session_id}
    
    alt Session 存在且 running
        SP-->>DAP: Session Info (status=running)
        Note over DAP: 复用现有 Session
    else Session 不存在 (404)
        Note over DAP,SP: 2. 创建新 Session
        DAP->>SP: POST /api/v1/sessions (user_id, agent_id)
        SP-->>DAP: Create Response (session_id, status=pending)
        
        Note over DAP,SP: 3. 等待 Session 就绪
        loop Poll until running (max 30 times)
            DAP->>SP: GET /api/v1/sessions/{session_id}
            SP-->>DAP: Session Status
        end
    else Session 异常/非 running (error/stopped)
        SP-->>DAP: Session Info (status!=running)
        Note over DAP,SP: 4. 自动重新创建 Session
        DAP->>SP: POST /api/v1/sessions (user_id, agent_id)
        SP-->>DAP: Create Response (new session_id, status=pending)
        
        loop Poll until running (max 30 times)
            DAP->>SP: GET /api/v1/sessions/{new_session_id}
            SP-->>DAP: Session Status
        end
    end

    Note over DAP,OP: 5. 调用 Agent Executor
    DAP->>OP: Call Agent Executor (with sandbox_session_id)
    OP->>SP: Execute Code (with sandbox_session_id)
    SP-->>OP: Stream Result
    OP-->>DAP: Stream Result

    DAP-->>User: Stream Response (SSE)
```

---

# 2. 核心概念设计

## 2.1 Sandbox Session（用户沙箱会话）

### 定义

每个用户有且只有一个 Sandbox Session，该用户下的所有 Agent 和所有 Conversation 共享同一个 Session，Session 跟随用户生命周期。

### Session ID 生成规则

- **格式**：`sb-session-{user_id}`
- **示例**：`sb-session-user123`
- **固定性**：同一用户始终使用相同的 Session ID
- **共享性**：同一用户下的所有 Agent 和所有 Conversation 共享同一个 Sandbox Session
- **无需缓存**：每次 Chat 都基于 `user_id` 生成相同的 Session ID

### 生命周期管理

1. **检测时机**：用户发起 Chat 请求时
2. **检测方式**：调用 Sandbox Platform API 直接检测 Session 状态
3. **自动创建**：如果 Session 不存在或异常，自动创建新 Session
4. **等待就绪**：创建新 Session 后，轮询等待状态变为 `running`
5. **复用逻辑**：如果 Session 存在且状态为 `running`，直接复用

### 无需持久化存储

- 不使用数据库存储 Session 映射关系
- 不使用 Redis 缓存
- 不使用 sync.Map 内存缓存
- Session ID 直接由 `user_id` 生成，每次都相同

---

## 2.2 Workspace File（临时区文件）

### 文件路径结构（Sandbox Platform 内部）

```
/workspace/
└── {conversation_id}/
    └── uploads/
        └── temparea/              # 【新增】临时区上传的文件目录
            ├── data.csv
            ├── model.pkl
            └── config.json
```

### 路径说明

| 路径层级 | 说明 |
|---------|------|
| `{conversation_id}` | Conversation ID |
| `uploads` | Sandbox Platform 固定目录，用于存放上传文件 |
| `temparea` | **固定目录名**，用于区分用户上传文件与其他类型文件 |
| `*.csv, *.pkl, ...` | 用户上传的文件 |

### Agent 代码引用示例

```python
# Agent 生成的代码中引用用户上传的文件
import pandas as pd

# Agent 直接使用物理路径，无需 mount 映射
df = pd.read_csv('/workspace/{conversation_id}/uploads/temparea/data.csv')
```

---

# 3. 核心流程设计

## 3.1 Chat 流程中集成 Sandbox Session 管理

### 关键代码位置

在 `src/domain/service/agentrunsvc/chat.go` 的 `Chat` 函数中，第 122-126 行（HandleGetInfoOrCreate 之后）添加：

```go
// NOTE: 确保 Sandbox Session 存在并就绪
sessionID := fmt.Sprintf("sb-session-%s", req.UserID)
sandboxSessionID, err := agentSvc.EnsureSandboxSession(ctx, sessionID, req)
if err != nil {
    o11y.Error(newCtx, fmt.Sprintf("[chat] ensure sandbox session failed: %v", err))
    return nil, err
}

// 将 sandbox_session_id 传递给 Agent Executor
req.SandboxSessionID = sandboxSessionID
```

---

## 3.2 EnsureSandboxSession 函数实现

### 完整实现

```go
// 在 src/domain/service/agentrunsvc/ensure_sandbox_session.go 中新增

package agentsvc

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
    agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
    o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
    "github.com/pkg/errors"
)

// EnsureSandboxSession 确保 Sandbox Session 存在并就绪
// 完全移除 sync.Map 缓存，每次直接调用 Sandbox Platform 检测
func (s *agentSvc) EnsureSandboxSession(ctx context.Context, sessionID string, req *agentreq.ChatReq) (string, error) {
    ctx, _ = o11y.StartInternalSpan(ctx)
    defer o11y.EndSpan(ctx, nil)

    o11y.SetAttributes(ctx,
        o11y.String("session_id", sessionID),
        o11y.String("user_id", req.UserID),
        o11y.String("agent_id", req.AgentID),
    )

    // 1. 检测 Session 状态
    sessionInfo, err := s.sandboxPlatform.GetSession(ctx, sessionID)
    if err != nil {
        // 404 错误表示 Session 不存在，继续创建
        if s.isSessionNotFoundError(err) {
            o11y.SetAttributes(ctx, o11y.String("action", "create_new"))
            return s.createNewSession(ctx, sessionID, req)
        }
        
        // 其他错误：尝试创建新 Session
        o11y.SetAttributes(ctx, o11y.String("action", "recover_from_error"))
        s.logger.Warnf("[EnsureSandboxSession] get session failed: %v, will create new session", err)
        return s.createNewSession(ctx, sessionID, req)
    }

    // 2. 检查 Session 状态
    if sessionInfo.Status == "running" {
        o11y.SetAttributes(ctx, o11y.String("action", "reuse_existing"))
        s.logger.Infof("[EnsureSandboxSession] reuse existing session: %s", sessionID)
        return sessionID, nil
    }

    // 3. Session 状态非 running，自动重新创建
    o11y.SetAttributes(ctx, o11y.String("action", "recreate"))
    s.logger.Warnf("[EnsureSandboxSession] session status is %s, will recreate: %s", sessionInfo.Status, sessionID)
    return s.createNewSession(ctx, sessionID, req)
}

// createNewSession 创建新的 Sandbox Session
func (s *agentSvc) createNewSession(ctx context.Context, sessionID string, req *agentreq.ChatReq) (string, error) {
    createReq := sandboxplatformdto.CreateSessionReq{
        UserID:           req.UserID,
        AgentID:          req.AgentID,
        BusinessDomainID: req.XBusinessDomainID,
        Config: map[string]interface{}{
            "session_id": sessionID, // 使用预生成的 session_id
            "file_upload_config": map[string]interface{}{
                "max_file_size":      s.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileSize,
                "max_file_size_unit": s.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileSizeUnit,
                "max_file_count":     s.sandboxPlatformConf.DefaultFileUploadConfig.MaxFileCount,
                "allowed_file_types": s.sandboxPlatformConf.DefaultFileUploadConfig.AllowedFileTypes,
            },
        },
    }

    createResp, err := s.sandboxPlatform.CreateSession(ctx, createReq)
    if err != nil {
        // 检查是否为 "已存在" 错误（并发场景下其他请求已创建）
        if s.isSessionAlreadyExistsError(err) {
            s.logger.Infof("[createNewSession] session already exists: %s, will wait for ready", sessionID)
            // 直接等待现有 Session 就绪
            return s.waitForSessionReady(ctx, sessionID)
        }

        s.logger.Errorf("[createNewSession] create failed: %v", err)
        return "", errors.Wrap(err, "create sandbox session failed")
    }

    // 使用返回的 session_id（可能与请求的预生成 ID 不同）
    actualSessionID := createResp.SessionID
    if createResp.SessionID == "" {
        actualSessionID = sessionID
    }

    // 等待 Session 就绪
    return s.waitForSessionReady(ctx, actualSessionID)
}

// waitForSessionReady 等待 Session 就绪
func (s *agentSvc) waitForSessionReady(ctx context.Context, sessionID string) (string, error) {
    maxRetries := s.sandboxPlatformConf.MaxRetries
    retryInterval := s.sandboxPlatformConf.RetryInterval

    for i := 0; i < maxRetries; i++ {
        sessionInfo, err := s.sandboxPlatform.GetSession(ctx, sessionID)
        if err != nil {
            s.logger.Errorf("[waitForSessionReady] get session status failed (attempt %d): %v", i+1, err)
            time.Sleep(retryInterval)
            continue
        }

        if sessionInfo.Status == "running" {
            s.logger.Infof("[waitForSessionReady] session ready: %s (attempts: %d)", sessionID, i+1)
            return sessionID, nil
        }

        // 如果状态是 error/stopped，直接失败
        if sessionInfo.Status == "error" || sessionInfo.Status == "stopped" {
            return "", errors.Errorf("session in invalid state: %s", sessionInfo.Status)
        }

        time.Sleep(retryInterval)
    }

    return "", errors.New("timeout waiting for session ready")
}

// isSessionNotFoundError 判断是否为 Session 不存在的错误
func (s *agentSvc) isSessionNotFoundError(err error) bool {
    // 检查是否为 rest.HTTPError 类型
    var httpErr *rest.HTTPError
    if errors.As(err, &httpErr) {
        return httpErr.StatusCode == http.StatusNotFound
    }
    return false
}

// isSessionAlreadyExistsError 判断是否为 Session 已存在的错误
func (s *agentSvc) isSessionAlreadyExistsError(err error) bool {
    var httpErr *rest.HTTPError
    if errors.As(err, &httpErr) {
        return httpErr.StatusCode == http.StatusConflict // 409 Conflict
    }
    return strings.Contains(err.Error(), "already exists")
}
```

---

# 4. DDD 层次设计（简化版）

## 4.1 目录结构

```
src/
├── drivenadapter/
│   └── httpaccess/
│       └── sandboxplatformhttp/
│           ├── define.go
│           ├── create_session.go
│           ├── get_session.go
│           └── delete_session.go
└── domain/service/
    └── agentrunsvc/
        ├── chat.go
        ├── ensure_sandbox_session.go  # 【新增】
        ├── common.go
        └── define.go
```

---

## 4.2 Sandbox Platform HTTP 客户端

### 接口定义

**port/driven/ihttpaccess/isandboxplatformhttp/sandbox_platform.go**：

```go
package isandboxplatformhttp

import (
    "context"
    sandboxdto "github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
)

// ISandboxPlatform Sandbox Platform 接口
type ISandboxPlatform interface {
    // CreateSession 创建 Sandbox Session
    CreateSession(ctx context.Context, req sandboxdto.CreateSessionReq) (*sandboxdto.CreateSessionResp, error)
    // GetSession 获取 Sandbox Session 信息
    GetSession(ctx context.Context, sessionID string) (*sandboxdto.GetSessionResp, error)
    // DeleteSession 删除 Sandbox Session
    DeleteSession(ctx context.Context, sessionID string) error
    // DeleteConversationFiles 删除指定 Conversation 的文件
    DeleteConversationFiles(ctx context.Context, sessionID, conversationID string) error
    // ListFiles 列出指定目录下的文件
    ListFiles(ctx context.Context, sessionID, conversationID, subdir string) ([]string, error)
}
```

---

### DTO 定义

**drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto/req.go**：

```go
package sandboxplatformdto

// CreateSessionReq 创建 Session 请求
type CreateSessionReq struct {
    UserID           string                 `json:"user_id"`
    AgentID          string                 `json:"agent_id"`
    BusinessDomainID string                 `json:"business_domain_id"`
    Config           map[string]interface{} `json:"config,omitempty"`
}
```

**drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto/resp.go**：

```go
package sandboxplatformdto

// CreateSessionResp 创建 Session 响应
type CreateSessionResp struct {
    SessionID string                 `json:"session_id"`
    Status    string                 `json:"status"`
    CreatedAt int64                  `json:"created_at"`
    TTL       int64                  `json:"ttl"`
    Info      map[string]interface{} `json:"info,omitempty"`
}

// GetSessionResp 获取 Session 响应
type GetSessionResp struct {
    SessionID string `json:"session_id"`
    Status    string `json:"status"` // pending/running/stopped
    CreatedAt int64  `json:"created_at"`
    TTL       int64  `json:"ttl"`
}
```

---

### HTTP 客户端实现

**drivenadapter/httpaccess/sandboxplatformhttp/create_session.go**：

```go
package sandboxplatformhttp

import (
    "context"
    "fmt"

    sandboxdto "github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
    "github.com/kweaver-ai/decision-agent/agent-factory/conf"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpclient"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
    o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
    "github.com/pkg/errors"
)

func (svc *sandboxPlatformHttpAcc) CreateSession(ctx context.Context, req sandboxdto.CreateSessionReq) (*sandboxdto.CreateSessionResp, error) {
    ctx, _ = o11y.StartInternalSpan(ctx)
    defer o11y.EndSpan(ctx, nil)

    url := fmt.Sprintf("%s/api/v1/sessions",
        cutil.GetHTTPAccess(svc.sandboxPlatformConf.Svc.Host, svc.sandboxPlatformConf.Svc.Port, svc.sandboxPlatformConf.Svc.Protocol))

    var resp sandboxdto.CreateSessionResp
    err := svc.httpClient.Post(ctx, url, req, &resp)
    if err != nil {
        return nil, errors.Wrap(err, "create sandbox session failed")
    }

    return &resp, nil
}
```

**drivenadapter/httpaccess/sandboxplatformhttp/get_session.go**：

```go
package sandboxplatformhttp

import (
    "context"
    "fmt"

    sandboxdto "github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpclient"
    o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
    "github.com/pkg/errors"
)

func (svc *sandboxPlatformHttpAcc) GetSession(ctx context.Context, sessionID string) (*sandboxdto.GetSessionResp, error) {
    ctx, _ = o11y.StartInternalSpan(ctx)
    defer o11y.EndSpan(ctx, nil)

    url := fmt.Sprintf("%s/api/v1/sessions/%s",
        svc.sandboxPlatformConf.BaseURL, sessionID)

    var resp sandboxdto.GetSessionResp
    err := svc.httpClient.Get(ctx, url, nil, &resp)
    if err != nil {
        return nil, errors.Wrap(err, "get sandbox session failed")
    }

    return &resp, nil
}
```

**drivenadapter/httpaccess/sandboxplatformhttp/define.go**：

```go
package sandboxplatformhttp

import (
    "sync"

    "github.com/kweaver-ai/decision-agent/agent-factory/conf"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/httpclient"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
)

type sandboxPlatformHttpAcc struct {
    logger                icmp.Logger
    httpClient            httpclient.HTTPClient
    sandboxPlatformConf   *conf.SandboxPlatformConf
    baseURL              string
}

var (
    sandboxPlatformOnce sync.Once
    sandboxPlatformImpl ISandboxPlatform
)

func NewSandboxPlatformHttpAcc(sandboxPlatformConf *conf.SandboxPlatformConf, httpClient httpclient.HTTPClient, logger icmp.Logger) ISandboxPlatform {
    sandboxPlatformOnce.Do(func() {
        sandboxPlatformImpl = &sandboxPlatformHttpAcc{
            logger:              logger,
            httpClient:          httpClient,
            sandboxPlatformConf:  sandboxPlatformConf,
            baseURL:             cutil.GetHTTPAccess(sandboxPlatformConf.Svc.Host, sandboxPlatformConf.Svc.Port, sandboxPlatformConf.Svc.Protocol),
        }
    })
    return sandboxPlatformImpl
}
```

---

## 4.3 Conversation 删除时的文件清理

### 删除流程

在 `src/drivenadapter/dbaccess/conversationdbacc/delete.go` 的删除逻辑中增加 Hook：

```go
// 在 src/drivenadapter/dbaccess/conversationdbacc/delete.go 中
// 需要导入: "context", "fmt", "time"

func (acc *conversationDBAcc) Delete(ctx context.Context, req conversationdto.DeleteReq) error {
    // 1. 执行数据库软删除
    err := acc.db.Delete(...)
    if err != nil {
        return err
    }

    // 2. 异步清理 Sandbox Workspace 文件（带超时和重试）
    go func() {
        // 带超时的 context，防止清理操作无限期执行
        cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        sessionID := fmt.Sprintf("sb-session-%s", req.UserID)

        // 重试机制（最多 3 次，指数退避）
        var lastErr error
        for i := 0; i < 3; i++ {
            if err := acc.sandboxPlatform.DeleteConversationFiles(cleanupCtx, sessionID, req.ConversationID); err != nil {
                lastErr = err
                logger.Warnf("[DeleteConversation] cleanup attempt %d failed: %v", i+1, err)
                time.Sleep(time.Second * time.Duration(i+1)) // 指数退避: 1s, 2s, 3s
                continue
            }
            logger.Infof("[DeleteConversation] cleanup success: %s/%s", sessionID, req.ConversationID)
            return
        }

        // 所有重试失败，记录到持久化存储以便后续处理
        logger.Errorf("[DeleteConversation] cleanup failed after retries: %v", lastErr)

        // TODO: 将清理任务写入队列/数据库，由定时任务重试
        // 例如：acc.cleanupQueue.Enqueue(CleanupTask{SessionID: sessionID, ConversationID: req.ConversationID})
    }()

    return nil
}
```

### 设计考虑
- **异步执行**：文件清理不阻塞 Conversation 删除操作
- **超时控制**：使用 `context.WithTimeout` 防止清理操作无限期执行（30秒超时）
- **重试机制**：最多重试 3 次，采用指数退避策略（1s, 2s, 3s）
- **幂等性**：多次调用 `DeleteConversationFiles` 应该是安全的
- **持久化重试**：所有重试失败后，应将清理任务写入队列/数据库，由定时任务处理
- **应用重启容错**：即使应用重启，未完成的清理任务仍可由定时任务重试

---

# 5. 配置设计

## 5.1 Sandbox Platform 配置

**conf/sandbox_platform.go**：

```go
package conf

import "github.com/kweaver-ai/decision-agent/agent-factory/cconf"

// SandboxPlatformConf Sandbox Platform 配置
type SandboxPlatformConf struct {
    SvcConf                      cconf.SvcConf      `yaml:"svc"`
    DefaultTTL                   int64               `yaml:"default_ttl"`     // 默认 Session TTL（秒）
    MaxRetries                   int                 `yaml:"max_retries"`     // 等待 Session 就绪的最大重试次数
    RetryInterval                string              `yaml:"retry_interval"`  // 重试间隔（如 "500ms"）
    DefaultFileUploadConfig      FileUploadConfig    `yaml:"file_upload_config"`
}

// FileUploadConfig 文件上传配置
type FileUploadConfig struct {
    MaxFileSize      int64    `yaml:"max_file_size"`      // 最大文件大小（数值）
    MaxFileSizeUnit  string   `yaml:"max_file_size_unit"` // 单位：KB/MB/GB
    MaxFileCount     int      `yaml:"max_file_count"`     // 最大文件数量
    AllowedFileTypes []string `yaml:"allowed_file_types"` // 允许的文件类型
}
```

---

## 5.2 配置文件示例

**agent-factory.yaml**：

```yaml
sandbox_platform:
  svc:
    host: "sandbox-platform.kweaver.svc.cluster.local"
    port: 9100
    protocol: "http"
  default_ttl: 7200              # 默认 2 小时
  max_retries: 30                 # 最多重试 30 次
  retry_interval: "500ms"         # 每次间隔 500ms
  file_upload_config:
    max_file_size: 100
    max_file_size_unit: "MB"
    max_file_count: 50
    allowed_file_types:
      - "csv"
      - "xlsx"
      - "json"
      - "txt"
      - "pkl"
      - "png"
      - "jpg"
      - "pdf"
```

---

# 6. 前端集成指南

## 6.1 前端如何获取 Sandbox Session ID

### Session ID 生成规则

Session ID 基于 `user_id` 生成，前后端使用相同的规则即可：

```typescript
// 前端生成 Session ID
function generateSessionId(userId: string): string {
    return `sb-session-${userId}`;
}

// 示例
const sessionId = generateSessionId('user123');
// 结果: "sb-session-user123"
```

### 使用方式

前端在上传文件时，使用相同的 Session ID 生成规则即可：

```typescript
// 上传文件
async function uploadFile(userId: string, conversationId: string, file: File) {
    const sessionId = generateSessionId(userId);

    const formData = new FormData();
    formData.append('file', file);
    formData.append('conversation_id', conversationId);
    formData.append('subdir', 'temparea');

    const response = await fetch(
        `${SANDBOX_API_URL}/api/v1/sessions/${sessionId}/files/upload`,
        {
            method: 'POST',
            body: formData,
        }
    );

    return response.json();
}
```

### 关键要点

- **无需从 Decision Agent 获取 Session ID**
- **前后端使用相同的生成规则**
- **格式**：`sb-session-{user_id}`
- **固定性**：同一用户始终使用相同的 Session ID
- **共享性**：同一用户下的所有 Agent 和 Conversation 共享同一个 Session

---

## 6.2 上传文件到 Sandbox

```typescript
interface UploadFileOptions {
  sandboxSessionId: string;
  conversationId: string;
  file: File;
}

async function uploadFile(options: UploadFileOptions): Promise<any> {
  const { sandboxSessionId, conversationId, file } = options;
  const formData = new FormData();
  formData.append('file', file);
  formData.append('conversation_id', conversationId);
  formData.append('subdir', 'temparea'); // 指定子目录为 temparea

  const response = await fetch(
    `${SANDBOX_API_URL}/api/v1/sessions/${sandboxSessionId}/files/upload`,
    {
      method: 'POST',
      body: formData,
    }
  );

  if (!response.ok) {
    throw new Error(`Upload failed: ${response.statusText}`);
  }

  return response.json();
}

// 使用示例
const sandboxSessionId = await chatWithSandboxSession('agent123', '上传一个文件分析');

await uploadFile({
  sandboxSessionId: sandboxSessionId,
  conversationId: 'conv456',
  file: selectedFile,
});
```

---

## 6.3 列出临时区文件

```typescript
interface ListFilesOptions {
  sandboxSessionId: string;
  conversationId: string;
}

async function listTempAreaFiles(options: ListFilesOptions): Promise<string[]> {
  const { sandboxSessionId, conversationId } = options;

  const response = await fetch(
    `${SANDBOX_API_URL}/api/v1/sessions/${sandboxSessionId}/files`,
    {
      method: 'GET',
      headers: {
        'conversation_id': conversationId,
        'subdir': 'temparea', // 只列出 temparea 目录下的文件
      },
    }
  );

  if (!response.ok) {
    throw new Error(`List files failed: ${response.statusText}`);
  }

  const data = await response.json();
  return data.files; // ["data.csv", "model.pkl", ...]
}

// 使用示例
const files = await listTempAreaFiles({
  sandboxSessionId: sandboxSessionId,
  conversationId: 'conv456',
});

console.log('Uploaded files:', files);
```

---

## 6.4 下载文件

```typescript
interface DownloadFileOptions {
  sandboxSessionId: string;
  fileName: string;
}

async function downloadFile(options: DownloadFileOptions): Promise<void> {
  const { sandboxSessionId, fileName } = options;

  const response = await fetch(
    `${SANDBOX_API_URL}/api/v1/sessions/${sandboxSessionId}/files/${fileName}`,
    {
      method: 'GET',
    }
  );

  if (!response.ok) {
    throw new Error(`Download failed: ${response.statusText}`);
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = fileName;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.URL.revokeObjectURL(url);
}

// 使用示例
await downloadFile({
  sandboxSessionId: sandboxSessionId,
  fileName: 'data.csv',
});
```

---

# 7. Agent 代码执行集成

## 7.1 Query 注入：将文件信息传递给 Agent Executor

### 设计原则

本方案采用 **agent-factory 端注入** 的方式：
1. 前端传递用户选择的文件列表（只包含文件名）
2. agent-factory 接收文件列表，直接构建完整路径
3. agent-factory 将文件信息注入到用户问题（query）中
4. agent-executor 只接收注入后的完整 query，不感知文件选择逻辑

### 数据流

```
前端选择文件 → SelectedFiles (仅文件名)
    ↓
ChatReq (SelectedFiles, Query)
    ↓
GenerateAgentCallReq
    ├─ buildUserQuery(Query + SelectedFiles) → finalQuery
    ↓
agentCallReq.Input["query"] = finalQuery
    ↓
agent-executor (只接收 finalQuery，不感知文件信息)
    ↓
LLM (在 prompt 中看到文件路径信息)
```

---

## 7.2 修改 ChatReq DTO

### 定义 SelectedFile 结构体

**driveradapter/api/rdto/agent/req/chat_req.go**：

```go
package req

// SelectedFile 用户选择的临时区文件
type SelectedFile struct {
    FileName string `json:"file_name" validate:"required"` // 文件名
    // 注：完整路径为 /workspace/{conversation_id}/uploads/temparea/{file_name}
}

type ChatReq struct {
    // ... 现有字段 ...

    // SelectedFiles 用户选择的临时区文件（新增）
    // 用户上传文件后，可以在对话时选择哪些文件参与本次对话
    SelectedFiles []SelectedFile `json:"selected_files,omitempty"`
}
```

**字段说明**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `selected_files` | array | 否 | 用户选择的临时区文件列表 |

**SelectedFile 结构**：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `file_name` | string | 是 | 文件名（不含路径） |

**使用场景**：

1. 用户先通过 Sandbox Platform API 上传文件到临时区
2. 用户发起对话时，可以选择已上传的文件参与本次对话
3. 前端将选中的文件名通过 `selected_files` 参数传递给后端
4. 后端将文件信息注入到 query 中，传递给 agent-executor

### 删除旧字段

**删除**以下字段（破坏性变更）：

```go
type ChatReq struct {
    // ... 其他字段 ...

    // 以下字段已删除
    // TemporaryAreaID string `json:"temporary_area_id"`  // 已删除
    // TempFiles      []valueobject.TempFile `json:"temp_files"`  // 已删除
}
```

---

## 7.3 实现 buildWorkspaceContextMessage 函数

### 新增文件

**domain/service/util/workspace_util.go**（新建）：

```go
package util

import (
    "fmt"
    "path"
    "strings"

    agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

// BuildWorkspaceContextMessage 生成独立的工作区上下文消息
// 作为单独的 "user" 角色消息插入，与用户问题分离
// selectedFiles.FileName 包含完整路径，例如：/workspace/conv-123/uploads/temparea/data.csv
func BuildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
    if len(selectedFiles) == 0 {
        return ""
    }

    var fileList strings.Builder
    for _, file := range selectedFiles {
        // file.FileName 现在是完整路径，例如：/workspace/conv-123/uploads/temparea/data.csv
        fullPath := file.FileName
        fileName := path.Base(fullPath)
        fileList.WriteString(fmt.Sprintf("- %s (%s)\n", fileName, fullPath))
    }

    rootPath := fmt.Sprintf("/workspace/%s/uploads/temparea/", conversationID)
    sandboxSessionID := fmt.Sprintf("sess-%s", userID)
    contextMsg := fmt.Sprintf(`【System auto-generated context - not user query】

Current workspace path: %s
Sandbox Session ID: %s (IMPORTANT: This ID MUST be passed as the 'session_id' parameter when calling code execution tools...)

Available files:
%s`,
        rootPath,
        sandboxSessionID,
        fileList.String(),
    )

    return contextMsg
}
```

### 设计说明

1. **独立消息而非拼接**：上下文作为独立的 LLMMessage 插入，不与用户查询拼接
2. **使用 "user" 角色**：文件是短生命周期状态，不同于长期缓存的 system prompt
3. **包含 Sandbox Session ID**：明确告知 LLM 代码执行时需要传递的 session_id 参数
4. **完整路径处理**：从 `SelectedFiles.FileName`（完整路径）中提取文件名用于显示

---

## 7.4 修改 GenerateAgentCallReq 函数

### 修改位置

**domain/service/agentrunsvc/chat_req.go**：

```go
func (agentSvc *agentSvc) GenerateAgentCallReq(
    ctx context.Context,
    req *agentreq.ChatReq,
    contexts []*comvalobj.LLMMessage,
    agent agentfactorydto.Agent,
) (*agentexecutordto.AgentCallReq, error) {
    // ... 现有代码 ...

    if contexts == nil {
        contexts = []*comvalobj.LLMMessage{}
    }
    // NOTE: 如果req.History不为空，应该直接使用req.History
    if len(req.History) > 0 {
        contexts = req.History
    }

    // 新增：将当前消息的上下文添加到 contexts
    // 如果有选中的文件，先插入工作区上下文消息
    // 注意：当前用户查询不需要添加到 contexts，因为它已经通过 Input["query"] 单独传递
    if len(req.SelectedFiles) > 0 {
        contextMsg := &comvalobj.LLMMessage{
            Role:    "user",
            Content: util.BuildWorkspaceContextMessage(req.ConversationID, req.UserID, req.SelectedFiles),
        }
        contexts = append(contexts, contextMsg)
    }

    agentCallReq := &agentexecutordto.AgentCallReq{
        ID:           req.AgentID,
        AgentVersion: req.AgentVersion,
        Config:       AgentConfig2AgentCallConfig(ctx, &agent.Config, req),
        Input: map[string]interface{}{
            "query":   req.Query,  // 使用干净的原始 query
            "history": contexts,    // 包含历史 + 上下文消息（如果有文件）
            // ... 其他字段 ...
        },
        // ... 其他字段保持不变 ...
    }

    return agentCallReq, nil
}
```

### 代码对比

**修改前**（旧实现 - 将文件信息拼接进 query）：
```go
// 旧方式：拼接文件信息到 query
finalQuery := buildUserQuery(req.Query, req.ConversationID, req.SelectedFiles)

agentCallReq := &agentexecutordto.AgentCallReq{
    Input: map[string]interface{}{
        "query": finalQuery,  // 包含文件信息的拼接字符串
        "history": contexts,
    },
}
```

**修改后**（新实现 - 独立上下文消息）：
```go
// 新方式：上下文作为独立消息插入
if len(req.SelectedFiles) > 0 {
    contextMsg := &comvalobj.LLMMessage{
        Role:    "user",
        Content: util.BuildWorkspaceContextMessage(req.ConversationID, req.UserID, req.SelectedFiles),
    }
    contexts = append(contexts, contextMsg)
}

agentCallReq := &agentexecutordto.AgentCallReq{
    Input: map[string]interface{}{
        "query":   req.Query,  // 干净的原始 query
        "history": contexts,    // history 中包含上下文消息
    },
}
```

### 历史消息重建

**domain/service/conversationsvc/get_history.go**：

```go
// 在 GetHistory 函数中，为有 SelectedFiles 的用户消息重建上下文
for _, msg := range conversation.Messages {
    if msg.Role == cdaenum.MsgRoleUser {
        userContent := conversationmsgvo.UserContent{}
        if msg.Content != nil && *msg.Content != "" {
            err := sonic.Unmarshal([]byte(*msg.Content), &userContent)
            // ...
        }

        // NEW: 如果用户选中了文件，先插入工作区上下文消息
        if len(userContent.SelectedFiles) > 0 {
            contextMsg := &comvalobj.LLMMessage{
                Role:    "user",
                Content: util.BuildWorkspaceContextMessage(msg.ConversationID, conversation.CreateBy, userContent.SelectedFiles),
            }
            history = append(history, contextMsg)
        }

        // 然后添加实际的用户查询
        history = append(history, &comvalobj.LLMMessage{
            Role:    string(msg.Role),
            Content: userContent.Text,
        })
    }
}
```

---

## 7.5 agent-executor 端无需修改

由于文件信息已经作为独立的上下文消息注入到 `history` 中，agent-executor 端**无需任何修改**：

- agent-executor 接收干净的 `query` 和包含上下文的 `history`
- LLM 在 messages 中会看到类似以下结构：
  ```json
  [
    { "role": "system", "content": "..." },
    { "role": "user", "content": "【System auto-generated context - not user query】\n\nCurrent workspace path: /workspace/conv-123/uploads/temparea/\nSandbox Session ID: sess-user-123\n\nAvailable files:\n- data.csv (/workspace/conv-123/uploads/temparea/data.csv)" },
    { "role": "user", "content": "分析一下 data.csv 中的销售趋势" }
  ]
  ```
- Agent 生成的代码可以直接使用路径和 session_id

### Agent 代码示例

```python
# Agent 生成的代码中引用用户上传的文件
import pandas as pd

# 读取用户上传的文件（路径已在 context 中提供）
df = pd.read_csv('/workspace/conv-123/uploads/temparea/data.csv')
print(df.head())

# 处理数据
result = df.describe()

# Agent 生成结果文件（可选，不在 temparea 目录）
with open('/workspace/uploads/result.json', 'w') as f:
    f.write(result.to_json())

# 注意：如果需要执行代码，必须传递 session_id
# execute_code(session_id="sess-user-123", code="...")
```

---

## 7.6 完整请求示例

### 前端请求

```json
{
  "query": "分析一下 data.csv 中的销售趋势",
  "conversation_id": "conv-123",
  "user_id": "user-123",
  "stream": true,
  "selected_files": [
    { "file_name": "/workspace/conv-123/uploads/temparea/data.csv" },
    { "file_name": "/workspace/conv-123/uploads/temparea/config.json" }
  ]
}
```

### 传递给 agent-executor 的 Input

```json
{
  "query": "分析一下 data.csv 中的销售趋势",
  "history": [
    {
      "role": "system",
      "content": "You are a helpful assistant..."
    },
    {
      "role": "user",
      "content": "【System auto-generated context - not user query】\n\nCurrent workspace path: /workspace/conv-123/uploads/temparea/\nSandbox Session ID: sess-user-123 (IMPORTANT: This ID MUST be passed as the 'session_id' parameter when calling code execution tools...)\n\nAvailable files:\n- data.csv (/workspace/conv-123/uploads/temparea/data.csv)\n- config.json (/workspace/conv-123/uploads/temparea/config.json)"
    },
    {
      "role": "user",
      "content": "分析一下 data.csv 中的销售趋势"
    }
  ]
}
```

### 关键变化说明

1. **query 保持干净**：`query` 字段只包含用户原始问题
2. **上下文独立消息**：文件信息和 session_id 作为独立的 user 角色消息插入 history
3. **session_id 明确传递**：上下文消息中明确告知 LLM 需要使用的 session_id
4. **持久化保证**：`SelectedFiles` 已持久化到数据库，`GetHistory` 自动重建上下文

---

## 7.7 关键设计决策

### 为什么选择独立上下文消息而非拼接？

| 对比项 | 独立上下文消息 | 拼接到 query |
|--------|---------------|-------------|
| **语义清晰度** | 高（系统上下文 vs 用户问题明确分离） | 低（LLM 可能混淆） |
| **模型行为** | 不会重复路径/文件名 | 倾向于重复提示信息 |
| **持久化** | 从 `SelectedFiles` 重建，支持重入 | 需要额外处理拼接逻辑 |
| **可维护性** | 上下文逻辑集中管理 | 拼接逻辑分散在各处 |

### 为什么使用 "user" 角色而非 "system"？

- **文件是短生命周期状态**：每次对话可能变化
- **system prompt 是长期缓存**：相对稳定不变
- **语义一致性**：文件属于用户会话上下文，而非系统级配置

### 为什么不需要 Sandbox Platform API？

- 前端传递的 `SelectedFiles` 已经包含完整路径
- 直接使用 `SelectedFiles.FileName` 作为完整路径
- 避免额外的 API 调用，提升性能
- 简化错误处理逻辑

### 为什么独立上下文能解决重入问题？

- **`SelectedFiles` 已持久化**：存储在数据库的 `UserContent` 中
- **`GetHistory` 重建上下文**：读取历史消息时自动重建
- **无需额外存储**：不增加消息索引污染
- **自动同步**：文件变化时，新上下文自动反映当前状态

---

# 8. 关键实现注意点

## 8.1 并发控制

### 并发场景处理

当多个 Chat 请求同时到达时，可能出现以下情况：

```go
// 场景：用户同时发起 3 个 Chat 请求
// Request 1: GetSession(404) -> CreateSession -> Wait -> Success
// Request 2: GetSession(404) -> CreateSession -> Wait -> ?
// Request 3: GetSession(404) -> CreateSession -> Wait -> ?
```

**Request 2/3 可能遇到**：
1. CreateSession 返回 "already exists" 错误（409 Conflict）
2. CreateSession 成功但返回的是相同的 session_id
3. CreateSession 卡在 pending 状态

**处理策略**：

- **Session ID 固定**：同一用户始终使用相同的 Session ID
- **直接检测状态**：每次 Chat 都调用 Sandbox Platform API 检测
- **优雅处理冲突**：通过 `isSessionAlreadyExistsError` 检测 409 Conflict 错误
  - 如果检测到 "已存在" 错误，直接调用 `waitForSessionReady` 等待现有 Session 就绪
  - 避免因并发创建导致的失败
- **提取等待逻辑**：独立的 `waitForSessionReady` 函数供多种场景复用

### Sandbox Platform 需保证的幂等性

Sandbox Platform 的 `POST /api/v1/sessions` API 应该保证：
- 如果 `session_id` 已存在且状态正常，返回 409 Conflict 或现有 Session 信息
- 如果 `session_id` 不存在，创建新 Session
- 如果 `session_id` 存在但状态异常，可以选择：
  - 返回错误，由 Decision Agent 重新创建（使用新的 session_id）
  - 自动恢复 Session 到可用状态

---

## 8.2 错误处理

### Sandbox Platform 创建失败

```go
createResp, err := s.sandboxPlatform.CreateSession(ctx, createReq)
if err != nil {
    o11y.Error(ctx, fmt.Sprintf("[EnsureSandboxSession] create failed: %v", err))
    return "", rest.NewHTTPError(ctx, http.StatusInternalServerError,
        apierr.AgentAPP_SandboxSessionCreateFailed).
        WithErrorDetails("Failed to create sandbox session")
}
```

### Session 就绪超时

```go
if !sessionReady {
    o11y.Error(ctx, "[EnsureSandboxSession] session not ready after retries")
    return "", rest.NewHTTPError(ctx, http.StatusServiceUnavailable,
        apierr.AgentAPP_SandboxSessionNotReady).
        WithErrorDetails("Sandbox session initialization timeout")
}
```

### 轮询验证失败

```go
sessionInfo, err := s.sandboxPlatform.GetSession(ctx, sessionID)
if err != nil {
    o11y.Error(ctx, fmt.Sprintf("[EnsureSandboxSession] get session failed: %v", err))
    // 继续执行，尝试创建新 Session
}
```

---

## 8.3 日志与追踪

### OpenTelemetry 追踪

```go
func (s *agentSvc) EnsureSandboxSession(ctx context.Context, sessionID string, req *agentreq.ChatReq) (string, error) {
    ctx, _ = o11y.StartInternalSpan(ctx)
    defer o11y.EndSpan(ctx, nil)

    // 设置属性
    o11y.SetAttributes(ctx,
        o11y.String("session_id", sessionID),
        o11y.String("user_id", req.UserID),
        o11y.String("agent_id", req.AgentID),
        o11y.String("business_domain_id", req.XBusinessDomainID),
    )

    // ... 执行逻辑 ...

    return sessionID, nil
}
```

### 关键日志

```go
// 复用现有 Session
s.logger.Infof("[EnsureSandboxSession] reuse existing session: %s", sessionID)

// 创建新 Session
s.logger.Infof("[EnsureSandboxSession] create new sandbox session: %s", createResp.SessionID)

// Session 就绪
s.logger.Infof("[EnsureSandboxSession] sandbox session ready: %s (attempts: %d)", sessionID, attempts)
```

---

## 8.4 清理策略

### 无需主动清理

由于不使用任何缓存机制，无需主动清理 Session 映射：

- **无缓存**：不使用 sync.Map、Redis、数据库存储 Session 信息
- **无清理**：Session 清理由 Sandbox Platform 的 TTL 机制负责
- **无泄漏风险**：应用重启不影响，因为没有任何持久化状态

### Sandbox Platform 的清理责任

- Session TTL 过期后自动清理
- Session 达到资源限制时自动清理
- 支持手动删除 Session API（如果需要）

## 8.5 安全性设计

### 防止跨 Conversation 目录遍历

**威胁模型**：
恶意 Agent 或用户尝试访问其他 Conversation 的文件，例如：
```python
# 尝试访问其他对话的文件
df = pd.read_csv('/workspace/conv-123/uploads/temparea/../../conv-999/uploads/temparea/secret.csv')
```

**防护措施**：

1. **Sandbox Platform 端路径验证**
```go
// 在 Sandbox Platform 的文件操作中
func validatePath(sessionID, conversationID, requestedPath string) error {
    // 构建允许的前缀
    allowedPrefix := fmt.Sprintf("/workspace/%s/uploads/temparea/", conversationID)

    // 解析路径，防止 ../ 绕过
    cleanPath := filepath.Clean(requestedPath)
    if !strings.HasPrefix(cleanPath, allowedPrefix) {
        return errors.New("path traversal detected")
    }

    return nil
}
```

2. **容器内文件系统隔离**
- 每个独立的 Conversation 目录设置正确的权限
- 使用 Linux namespace 或 chroot 限制访问范围

3. **Agent 代码沙箱执行**
- Agent 生成的代码在受限环境中执行
- 禁止使用 `os.chdir()` 改变工作目录
- 禁止使用 `symlink()` 创建符号链接

### 路径注入安全

**防止 Prompt 注入攻击**：
- 文件名在注入前进行转义
- 使用结构化格式（JSON）而非字符串拼接
- 验证文件名只包含合法字符

```go
func sanitizeFileName(fileName string) string {
    // 移除路径分隔符
    fileName = strings.ReplaceAll(fileName, "/", "")
    fileName = strings.ReplaceAll(fileName, "\\", "")
    // 限制文件名长度
    if len(fileName) > 255 {
        fileName = fileName[:255]
    }
    return fileName
}
```

---

## 8.6 Agent Chat API 变更

### 背景

现有 Agent Chat API（`POST /api/agent-app/v1/app/{app_key}/chat/completion`）使用了旧的临时区实现方式：
- `temporary_area_id`：临时区域 ID
- `temp_files`：临时文件列表

### 新的 API 设计

采用 **query 注入**的方式，将用户选择的文件信息注入到用户问题中：
- 删除 `temporary_area_id` 和 `temp_files` 参数
- 新增 `selected_files` 参数，只包含文件名
- agent-factory 将文件信息注入到 query，agent-executor 不感知

### API 请求参数变更

**移除的参数**：

| 参数名 | 类型 | 说明 |
|--------|------|------|
| `temporary_area_id` | string | 临时区域 ID（已删除） |
| `temp_files` | array | 临时文件列表（已删除） |

**新增的参数**：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `selected_files` | array | 否 | 用户选择的临时区文件列表 |

**selected_files 参数结构**：

```json
{
  "selected_files": [
    {
      "file_name": "data.csv"
    },
    {
      "file_name": "config.json"
    }
  ]
}
```

**保持不变的参数**：

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `app_key` | string | 是 | Agent App Key |
| `conversation_id` | string | 是 | 会话 ID |
| `query` | string | 是 | 用户提问问题 |
| `stream` | boolean | 否 | 是否流式返回 |
| `inc_stream` | boolean | 否 | 是否增量流式返回 |
| `custom_querys` | object | 否 | 自定义输入变量 |

### 完整请求示例

```json
{
  "query": "分析一下 data.csv 文件中的销售数据",
  "conversation_id": "conv-123",
  "stream": true,
  "selected_files": [
    { "file_name": "data.csv" },
    { "file_name": "config.json" }
  ]
}
```

### 工作流程

```mermaid
sequenceDiagram
    participant User as 用户
    participant Frontend as 前端
    participant SP as Sandbox Platform
    participant DAP as Decision Agent Platform

    Note over User,SP: 步骤1: 用户上传文件
    User->>Frontend: 选择文件上传
    Frontend->>SP: POST /files/upload (session_id, conversation_id, file)
    SP-->>Frontend: 上传成功

    Note over User,DAP: 步骤2: 用户发起对话并选择文件
    User->>Frontend: 输入问题，选择已上传文件
    Frontend->>DAP: POST /chat/completion {query, selected_files}

    Note over DAP: buildUserQuery 注入文件信息
    DAP->>DAP: final_query = buildUserQuery(query, selected_files)

    Note over DAP: 调用 Agent Executor（只接收 final_query）
    DAP->>DAP: AgentExecutor.Call(final_query)

    DAP-->>Frontend: 返回 Chat 响应
```

### 前端集成变更

**文件上传**：

```typescript
// 1. 上传文件到 Sandbox Platform
async function uploadFile(userId: string, conversationId: string, file: File) {
    const sessionId = `sb-session-${userId}`;

    const formData = new FormData();
    formData.append('file', file);
    formData.append('conversation_id', conversationId);
    formData.append('subdir', 'temparea');

    const response = await fetch(
        `${SANDBOX_API_URL}/api/v1/sessions/${sessionId}/files/upload`,
        { method: 'POST', body: formData }
    );

    return response.json(); // { file_name: "data.csv", ... }
}
```

**发起对话（选择文件）**：

```typescript
// 2. 发起对话时选择已上传的文件
interface SelectedFile {
    file_name: string;
}

async function chatWithFiles(
    query: string,
    conversationId: string,
    selectedFiles: SelectedFile[]
) {
    const response = await fetch('/api/agent-app/v1/app/my-agent/chat/completion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            query: query,
            conversation_id: conversationId,
            stream: true,
            selected_files: selectedFiles  // 用户选择的文件
        })
    });

    return response.json();
}

// 使用示例
await chatWithFiles(
    '分析一下 data.csv 中的销售趋势',
    'conv-123',
    [
        { file_name: 'data.csv' },
        { file_name: 'config.json' }
    ]
);
```

**获取可用文件列表**：

```typescript
// 3. 获取已上传的文件列表供用户选择
async function listUploadedFiles(userId: string, conversationId: string) {
    const sessionId = `sb-session-${userId}`;

    const response = await fetch(
        `${SANDBOX_API_URL}/api/v1/sessions/${sessionId}/files`,
        {
            method: 'GET',
            headers: {
                'conversation_id': conversationId,
                'subdir': 'temparea'
            }
        }
    );

    const data = await response.json();
    return data.files; // ["data.csv", "config.json", ...]
}
```

### API 响应变更

响应结构**保持不变**。

---

# 9. 实施步骤

## 本次实施范围

本次实现专注于 **query 注入功能**，不涉及 Sandbox Session 管理。

- ✅ 实现 `selected_files` 参数接收
- ✅ 实现文件信息注入到 query
- ✅ 删除旧的 `temporary_area_id` 和 `temp_files` 字段
- ❌ 不涉及 Sandbox Session 管理（后续独立实施）

---

## 阶段一：DTO 修改

### 1.1 定义 SelectedFile 结构体

**文件**: `driveradapter/api/rdto/agent/req/chat_req.go`

```go
package req

// SelectedFile 用户选择的临时区文件
type SelectedFile struct {
    FileName string `json:"file_name" validate:"required"` // 文件名
    // 注：完整路径为 /workspace/{conversation_id}/uploads/temparea/{file_name}
}
```

### 1.2 修改 ChatReq

**文件**: `driveradapter/api/rdto/agent/req/chat_req.go`

```go
type ChatReq struct {
    // ... 现有字段 ...

    // SelectedFiles 用户选择的临时区文件（新增）
    SelectedFiles []SelectedFile `json:"selected_files,omitempty"`

    // 以下字段已删除
    // TemporaryAreaID string `json:"temporary_area_id"`  // 已删除
    // TempFiles      []valueobject.TempFile `json:"temp_files"`  // 已删除
}
```

### 1.3 同步修改 DebugReq

**文件**: `driveradapter/api/rdto/agent/req/debug_req.go`

删除 `TempFiles` 字段，保持与 ChatReq 一致。

---

## 阶段二：实现 buildWorkspaceContextMessage 函数

### 2.1 创建共享工具文件

**文件**: `domain/service/util/workspace_util.go`（新建）

```go
package util

import (
    "fmt"
    "path"
    "strings"

    agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

// BuildWorkspaceContextMessage 生成独立的工作区上下文消息
func BuildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
    if len(selectedFiles) == 0 {
        return ""
    }

    var fileList strings.Builder
    for _, file := range selectedFiles {
        fullPath := file.FileName
        fileName := path.Base(fullPath)
        fileList.WriteString(fmt.Sprintf("- %s (%s)\n", fileName, fullPath))
    }

    rootPath := fmt.Sprintf("/workspace/%s/uploads/temparea/", conversationID)
    sandboxSessionID := fmt.Sprintf("sess-%s", userID)
    contextMsg := fmt.Sprintf(`【System auto-generated context - not user query】

Current workspace path: %s
Sandbox Session ID: %s (IMPORTANT: This ID MUST be passed as the 'session_id' parameter when calling code execution tools...)

Available files:
%s`,
        rootPath,
        sandboxSessionID,
        fileList.String(),
    )

    return contextMsg
}
```

### 2.2 创建包装函数（可选）

**文件**: `domain/service/agentrunsvc/inject_workspace_context.go`（新建）

```go
package agentsvc

import (
    "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/util"
    agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

// buildWorkspaceContextMessage 生成独立的工作区上下文消息
// 实际逻辑已移至 util.BuildWorkspaceContextMessage
func buildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
    return util.BuildWorkspaceContextMessage(conversationID, userID, selectedFiles)
}
```

**文件**: `domain/service/conversationsvc/helpers.go`（更新）

```go
package conversationsvc

import (
    "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/util"
    agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
)

// buildWorkspaceContextMessage 生成独立的工作区上下文消息
func buildWorkspaceContextMessage(conversationID string, userID string, selectedFiles []agentreq.SelectedFile) string {
    return util.BuildWorkspaceContextMessage(conversationID, userID, selectedFiles)
}
```

---

## 阶段三：修改 GenerateAgentCallReq 函数

### 3.1 修改 ChatReq 处理逻辑

**文件**: `domain/service/agentrunsvc/chat_req.go`

**修改点**：
1. 将上下文消息插入到 `contexts`（而非修改 query）
2. 使用干净的原始 `req.Query`
3. 删除旧的 `TempFiles` 处理逻辑

**修改后**：
```go
// 新增：将当前消息的上下文添加到 contexts
// 如果有选中的文件，先插入工作区上下文消息
if len(req.SelectedFiles) > 0 {
    contextMsg := &comvalobj.LLMMessage{
        Role:    "user",
        Content: buildWorkspaceContextMessage(req.ConversationID, req.UserID, req.SelectedFiles),
    }
    contexts = append(contexts, contextMsg)
}

// 注意：不需要将当前用户查询添加到 contexts，因为它通过 Input["query"] 单独传递

agentCallReq := &agentexecutordto.AgentCallReq{
    Input: map[string]interface{}{
        "query":   req.Query,  // 使用干净的原始 query
        "history": contexts,    // 包含历史 + 上下文消息
        // ...
    },
    // ... 其他字段保持不变 ...
}
```

### 3.2 修改 GetHistory 函数

**文件**: `domain/service/conversationsvc/get_history.go`

**修改点**：为有 `SelectedFiles` 的用户消息重建上下文

```go
// 在处理用户消息时
if msg.Role == cdaenum.MsgRoleUser {
    userContent := conversationmsgvo.UserContent{}
    if msg.Content != nil && *msg.Content != "" {
        err := sonic.Unmarshal([]byte(*msg.Content), &userContent)
        // ...
    }

    // NEW: 如果用户选中了文件，先插入工作区上下文消息
    if len(userContent.SelectedFiles) > 0 {
        contextMsg := &comvalobj.LLMMessage{
            Role:    "user",
            Content: buildWorkspaceContextMessage(msg.ConversationID, conversation.CreateBy, userContent.SelectedFiles),
        }
        history = append(history, contextMsg)
    }

    // 然后添加实际的用户查询
    history = append(history, &comvalobj.LLMMessage{
        Role:    string(msg.Role),
        Content: userContent.Text,
    })
}
```

---

## 阶段四：清理旧代码

### 4.1 删除 TempFile 相关代码

查找并删除以下代码：

1. **TempFile 结构体使用**：
   - `domain/service/agentrunsvc/chat_msg.go` 中的 `TempFiles` 字段赋值
   - `domain/service/agentrunsvc/chat_post_process.go` 中的 `TempFiles` 字段赋值
   - `src/domain/valueobject/temp_file.go`（如果不再使用）

2. **TempFileProcess 相关**（如果完全废弃）：
   - `domain/entity/dolphintpleo/temp_file_process_content.go`
   - `domain/enum/cdaenum/dolphin_tpl_key.go` 中的 `DolphinTplKeyTempFileProcess`
   - 相关的配置和引用

**注意**：如果 `TempFileProcess` 还被其他功能使用，请谨慎删除。

### 4.2 更新单元测试

查找并更新以下测试文件：
- 测试文件中 `TempFiles` 字段的使用
- 测试文件中 `TemporaryAreaID` 字段的使用
- 更新测试用例以使用 `SelectedFiles`

---

## 阶段五：前端集成

### 5.1 移除旧的前端代码

删除以下代码（如果存在）：
- 传递 `temporary_area_id` 的代码
- 传递 `temp_files` 的代码

### 5.2 实现文件选择功能

```typescript
// 1. 上传文件到 Sandbox Platform
async function uploadFile(userId: string, conversationId: string, file: File) {
    const sessionId = `sb-session-${userId}`;

    const formData = new FormData();
    formData.append('file', file);
    formData.append('conversation_id', conversationId);
    formData.append('subdir', 'temparea');

    const response = await fetch(
        `${SANDBOX_API_URL}/api/v1/sessions/${sessionId}/files/upload`,
        { method: 'POST', body: formData }
    );

    return response.json();
}

// 2. 发起对话时选择文件
interface SelectedFile {
    file_name: string;
}

async function chatWithFiles(
    query: string,
    conversationId: string,
    selectedFiles: SelectedFile[]
) {
    const response = await fetch('/api/agent-app/v1/app/my-agent/chat/completion', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            query: query,
            conversation_id: conversationId,
            stream: true,
            selected_files: selectedFiles
        })
    });

    return response.json();
}

// 3. 获取可用文件列表
async function listUploadedFiles(userId: string, conversationId: string) {
    const sessionId = `sb-session-${userId}`;

    const response = await fetch(
        `${SANDBOX_API_URL}/api/v1/sessions/${sessionId}/files`,
        {
            method: 'GET',
            headers: {
                'conversation_id': conversationId,
                'subdir': 'temparea'
            }
        }
    );

    const data = await response.json();
    return data.files;
}
```

---

## 阶段六：测试验证

### 6.1 单元测试

**测试文件**: `domain/service/agentrunsvc/inject_workspace_context_test.go`（新建）

```go
package agentsvc_test

import (
    "testing"

    agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
    "github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service/util"
    "github.com/stretchr/testify/assert"
)

func TestBuildWorkspaceContextMessage(t *testing.T) {
    tests := []struct {
        name            string
        conversationID  string
        userID          string
        selectedFiles   []agentreq.SelectedFile
        expectedPrefix  string
        expectedContent string
    }{
        {
            name:           "no files selected",
            conversationID: "conv-123",
            userID:         "user-123",
            selectedFiles:  []agentreq.SelectedFile{},
            expectedPrefix: "",
        },
        {
            name:           "single file selected",
            conversationID: "conv-123",
            userID:         "user-456",
            selectedFiles: []agentreq.SelectedFile{
                {FileName: "/workspace/conv-123/uploads/temparea/data.csv"},
            },
            expectedPrefix:  "【System auto-generated context - not user query】",
            expectedContent: "Sandbox Session ID: sess-user-456",
        },
        {
            name:           "multiple files selected",
            conversationID: "conv-456",
            userID:         "user-789",
            selectedFiles: []agentreq.SelectedFile{
                {FileName: "/workspace/conv-456/uploads/temparea/data1.csv"},
                {FileName: "/workspace/conv-456/uploads/temparea/data2.csv"},
            },
            expectedPrefix: "【System auto-generated context - not user query】",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := util.BuildWorkspaceContextMessage(tt.conversationID, tt.userID, tt.selectedFiles)
            if tt.expectedPrefix == "" {
                assert.Empty(t, result)
            } else {
                assert.Contains(t, result, tt.expectedPrefix)
            }
            if tt.expectedContent != "" {
                assert.Contains(t, result, tt.expectedContent)
            }
        })
    }
}
```

### 6.2 集成测试

1. 测试 `GenerateAgentCallReq` 函数
   - 验证上下文消息正确插入到 `contexts`
   - 验证 `query` 保持干净（不包含文件信息）
2. 测试 `GetHistory` 函数
   - 验证历史消息的上下文重建
   - 验证 `SelectedFiles` 正确解析

### 6.3 端到端测试

1. 文件上传 → 选择文件 → 对话 → 验证 LLM 收到正确的文件路径
2. 验证 Agent 生成的代码能正确访问文件
3. 验证退出重进后上下文正确恢复
4. 验证 session_id 正确传递给代码执行工具

---

## 阶段七：API 文档更新

### 7.1 更新 API 文档

1. 移除 `temporary_area_id` 和 `temp_files` 参数说明
2. 添加 `selected_files` 参数说明
3. 添加新的文件上传方式说明（直接调用 Sandbox Platform API）

### 7.2 更新使用示例

提供完整的前后端集成示例代码。

---

# 10. 总结

### 核心设计原则

1. **简化优先**：不依赖任何缓存机制（无 Redis、无 sync.Map、无数据库）
2. **无侵入性**：不新增 Decision Agent API 接口
3. **自动管理**：在 Chat 流程中自动检查/创建 Sandbox Session
4. **路径约定**：使用 `temparea` 子目录区分用户上传文件
5. **容错设计**：创建失败、就绪超时等场景均有处理

### 关键技术点

- **固定 Session ID**：基于 `user_id` 生成
- **直接检测状态**：每次调用 Sandbox Platform API 检测
- **自动重新创建**：Session 不存在或异常时自动创建
- **轮询等待就绪**：创建后轮询等待 Session 状态变为 `running`
- **Agent Executor 传递**：将 `sandbox_session_id` 传递给 Agent Executor
- **Sandbox Platform 路径映射**：根据 session_id 和 conversation_id 自动映射

### 扩展性考虑

- 可选：用户登出时清理 Sandbox Session（调用 Sandbox Platform 删除 API）
- 可选：支持 Session 配置自定义（通过 config 参数）
- 可选：添加 Session 状态监控和告警

### API 变更摘要

**删除的参数**：
- `temporary_area_id`：已删除
- `temp_files`：已删除

**新增的参数**：
- `selected_files`：用户选择的临时区文件列表
  ```json
  {
    "selected_files": [
      { "file_name": "data.csv" },
      { "file_name": "config.json" }
    ]
  }
  ```

**工作流程**：
1. 用户先通过 Sandbox Platform API 上传文件
2. 用户发起对话时，可以选择已上传的文件
3. 前端将选中的文件通过 `selected_files` 参数传递
4. 后端根据用户选择的文件填充 WorkspaceContext

**客户端变更**：
- **文件上传**：直接调用 Sandbox Platform API
- **发起对话**：通过 `selected_files` 参数传递用户选择的文件
