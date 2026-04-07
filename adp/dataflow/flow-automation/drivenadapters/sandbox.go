package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/sandbox.go -destination ../tests/mock_drivenadapters/sandbox_mock.go

const (
	SessionStatusRunning   = "running"
	SessionStatusCreating  = "creating"
	SessionStatusFailed    = "failed"
	SessionStatusCompleted = "completed"
)

// Sandbox 沙箱服务接口
type Sandbox interface {
	// CreateSession 创建沙箱会话
	CreateSession(ctx context.Context, config *SandboxSessionConfig) (*Session, error)

	// GetSession 查询沙箱会话
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// ExecuteCode 在沙箱中执行代码
	ExecuteCode(ctx context.Context, sessionID string, req *ExecuteRequest) (*Execution, error)

	// GetExecutionStatus 查询执行状态
	GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error)

	// GetExecutionResult 获取执行结果
	GetExecutionResult(ctx context.Context, executionID string) (*ExecutionResult, error)
}

// SandboxSessionConfig 沙箱会话配置
type SandboxSessionConfig struct {
	// TemplateID 模板ID，必填
	TemplateID string `json:"template_id"`
	// CPU 资源限制
	CPU string `json:"cpu,omitempty"`
	// Memory 内存限制
	Memory string `json:"memory,omitempty"`
	// Disk 磁盘限制
	Disk string `json:"disk,omitempty"`
	// Timeout 会话超时时间（秒）
	Timeout int `json:"timeout,omitempty"`
	// Dependencies 依赖包列表
	Dependencies []SessionDependency `json:"dependencies,omitempty"`
}

// SessionDependency 沙箱依赖包
type SessionDependency struct {
	// Name 包名
	Name string `json:"name"`
	// Version 版本号，如 "=2.31.0"
	Version string `json:"version,omitempty"`
}

// Session 沙箱会话信息
type Session struct {
	ID             string                 `json:"id"`
	TemplateID     string                 `json:"template_id"`
	Status         string                 `json:"status"` // creating, running, completed, failed
	ResourceLimit  map[string]interface{} `json:"resource_limit,omitempty"`
	WorkspacePath  string                 `json:"workspace_path,omitempty"`
	RuntimeType    string                 `json:"runtime_type,omitempty"`
	ContainerID    string                 `json:"container_id,omitempty"`
	Timeout        int                    `json:"timeout,omitempty"`
	CreatedAt      string                 `json:"created_at,omitempty"`
	UpdatedAt      string                 `json:"updated_at,omitempty"`
	CompletedAt    *string                `json:"completed_at,omitempty"`
	LastActivityAt string                 `json:"last_activity_at,omitempty"`
}

// ExecuteRequest 执行代码请求
type ExecuteRequest struct {
	Code     string                 `json:"code"`
	Language string                 `json:"language"`
	Event    map[string]interface{} `json:"event,omitempty"`
	Timeout  int                    `json:"timeout,omitempty"`
}

// Execution 执行任务信息
type Execution struct {
	ExecutionID string `json:"execution_id"`
	SessionID   string `json:"session_id"`
	Status      string `json:"status"` // pending, running, completed, failed
	CreatedAt   string `json:"created_at,omitempty"`
}

// ExecutionStatus 执行状态信息
type ExecutionStatus struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	Status        string                 `json:"status"` // pending, running, completed, failed
	Code          string                 `json:"code,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Timeout       int                    `json:"timeout,omitempty"`
	ExitCode      *int                   `json:"exit_code,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	ExecutionTime *float64               `json:"execution_time,omitempty"`
	Stdout        string                 `json:"stdout,omitempty"`
	Stderr        string                 `json:"stderr,omitempty"`
	Artifacts     []interface{}          `json:"artifacts,omitempty"`
	RetryCount    int                    `json:"retry_count,omitempty"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	StartedAt     *string                `json:"started_at,omitempty"`
	CompletedAt   *string                `json:"completed_at,omitempty"`
	ReturnValue   map[string]interface{} `json:"return_value,omitempty"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	Status        string                 `json:"status"`
	Code          string                 `json:"code,omitempty"`
	Language      string                 `json:"language,omitempty"`
	Timeout       int                    `json:"timeout,omitempty"`
	ExitCode      *int                   `json:"exit_code,omitempty"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
	ExecutionTime *float64               `json:"execution_time,omitempty"`
	Stdout        string                 `json:"stdout,omitempty"`
	Stderr        string                 `json:"stderr,omitempty"`
	Artifacts     []interface{}          `json:"artifacts,omitempty"`
	RetryCount    int                    `json:"retry_count,omitempty"`
	CreatedAt     string                 `json:"created_at,omitempty"`
	StartedAt     *string                `json:"started_at,omitempty"`
	CompletedAt   *string                `json:"completed_at,omitempty"`
	ReturnValue   map[string]interface{} `json:"return_value,omitempty"`
	Metrics       map[string]interface{} `json:"metrics,omitempty"`
}

type sandbox struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

var (
	sandboxOnce sync.Once
	s           Sandbox
)

// NewSandbox 创建沙箱服务适配器
func NewSandbox() Sandbox {
	sandboxOnce.Do(func() {
		config := common.NewConfig()
		baseURL := fmt.Sprintf("http://%s:%v", config.Sandbox.Host, config.Sandbox.Port)
		s = &sandbox{
			baseURL:    baseURL,
			httpClient: NewOtelHTTPClient(),
		}
	})
	return s
}

// CreateSession 创建沙箱会话
func (s *sandbox) CreateSession(ctx context.Context, config *SandboxSessionConfig) (*Session, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/v1/sessions", s.baseURL)

	// 构建请求体
	body := map[string]interface{}{
		"template_id": config.TemplateID,
	}
	if config.CPU != "" {
		body["cpu"] = config.CPU
	}
	if config.Memory != "" {
		body["memory"] = config.Memory
	}
	if config.Disk != "" {
		body["disk"] = config.Disk
	}
	if config.Timeout > 0 {
		body["timeout"] = config.Timeout
	}
	if len(config.Dependencies) > 0 {
		body["dependencies"] = config.Dependencies
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, respParam, err := s.httpClient.Post(ctx, target, headers, body)
	if err != nil {
		log.Warnf("[Sandbox.CreateSession] Post failed: %v, url: %v", err, target)
		return nil, fmt.Errorf("create sandbox session failed: %w", err)
	}

	var session Session
	respBytes, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("[Sandbox.CreateSession] Marshal failed: %v", err)
		return nil, fmt.Errorf("marshal session response failed: %w", err)
	}

	if err := json.Unmarshal(respBytes, &session); err != nil {
		log.Warnf("[Sandbox.CreateSession] Unmarshal failed: %v", err)
		return nil, fmt.Errorf("unmarshal session response failed: %w", err)
	}

	return &session, nil
}

// GetSession 查询沙箱会话
func (s *sandbox) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/v1/sessions/%s", s.baseURL, sessionID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, respParam, err := s.httpClient.Get(ctx, target, headers)
	if err != nil {
		log.Warnf("[Sandbox.GetSession] Get failed: %v, url: %v", err, target)
		return nil, fmt.Errorf("get sandbox session failed: %w", err)
	}

	var session Session
	respBytes, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("[Sandbox.GetSession] Marshal failed: %v", err)
		return nil, fmt.Errorf("marshal session response failed: %w", err)
	}

	if err := json.Unmarshal(respBytes, &session); err != nil {
		log.Warnf("[Sandbox.GetSession] Unmarshal failed: %v", err)
		return nil, fmt.Errorf("unmarshal session response failed: %w", err)
	}

	return &session, nil
}

// ExecuteCode 在沙箱中执行代码
func (s *sandbox) ExecuteCode(ctx context.Context, sessionID string, req *ExecuteRequest) (*Execution, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/v1/executions/sessions/%s/execute", s.baseURL, sessionID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, respParam, err := s.httpClient.Post(ctx, target, headers, req)
	if err != nil {
		log.Warnf("[Sandbox.ExecuteCode] Post failed: %v, url: %v", err, target)
		return nil, fmt.Errorf("execute code failed: %w", err)
	}

	var execution Execution
	respBytes, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("[Sandbox.ExecuteCode] Marshal failed: %v", err)
		return nil, fmt.Errorf("marshal execution response failed: %w", err)
	}

	if err := json.Unmarshal(respBytes, &execution); err != nil {
		log.Warnf("[Sandbox.ExecuteCode] Unmarshal failed: %v", err)
		return nil, fmt.Errorf("unmarshal execution response failed: %w", err)
	}

	return &execution, nil
}

// GetExecutionStatus 查询执行状态
func (s *sandbox) GetExecutionStatus(ctx context.Context, executionID string) (*ExecutionStatus, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/v1/executions/%s/status", s.baseURL, executionID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, respParam, err := s.httpClient.Get(ctx, target, headers)
	if err != nil {
		log.Warnf("[Sandbox.GetExecutionStatus] Get failed: %v, url: %v", err, target)
		return nil, fmt.Errorf("get execution status failed: %w", err)
	}

	var status ExecutionStatus
	respBytes, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("[Sandbox.GetExecutionStatus] Marshal failed: %v", err)
		return nil, fmt.Errorf("marshal status response failed: %w", err)
	}

	if err := json.Unmarshal(respBytes, &status); err != nil {
		log.Warnf("[Sandbox.GetExecutionStatus] Unmarshal failed: %v", err)
		return nil, fmt.Errorf("unmarshal status response failed: %w", err)
	}

	return &status, nil
}

// GetExecutionResult 获取执行结果
func (s *sandbox) GetExecutionResult(ctx context.Context, executionID string) (*ExecutionResult, error) {
	log := traceLog.WithContext(ctx)
	target := fmt.Sprintf("%s/api/v1/executions/%s/result", s.baseURL, executionID)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	_, respParam, err := s.httpClient.Get(ctx, target, headers)
	if err != nil {
		log.Warnf("[Sandbox.GetExecutionResult] Get failed: %v, url: %v", err, target)
		return nil, fmt.Errorf("get execution result failed: %w", err)
	}

	var result ExecutionResult
	respBytes, err := json.Marshal(respParam)
	if err != nil {
		log.Warnf("[Sandbox.GetExecutionResult] Marshal failed: %v", err)
		return nil, fmt.Errorf("marshal result response failed: %w", err)
	}

	if err := json.Unmarshal(respBytes, &result); err != nil {
		log.Warnf("[Sandbox.GetExecutionResult] Unmarshal failed: %v", err)
		return nil, fmt.Errorf("unmarshal result response failed: %w", err)
	}

	return &result, nil
}
