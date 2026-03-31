package sandbox

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	ierrors "github.com/kweaver-ai/adp/autoflow/flow-automation/errors"
	commonLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	sandboxutil "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/sandbox"
)

type SandboxHandler interface {
	Execute(ctx context.Context, req *SandboxExecuteRequest) (*SandboxExecuteResult, error)
}

var (
	sOnce sync.Once
	sh    SandboxHandler
)

type sandbox struct {
	log commonLog.Logger
}

type SandboxExecuteRequest struct {
	Code          string                               `json:"code"`
	Language      string                               `json:"language"`
	Event         map[string]interface{}               `json:"event,omitempty"`
	Timeout       int                                  `json:"timeout,omitempty"`
	SandboxConfig *drivenadapters.SandboxSessionConfig `json:"sandbox_config,omitempty"`
}

type SandboxExecuteResult struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	Status        string                 `json:"status"`
	Code          string                 `json:"code"`
	Language      string                 `json:"language"`
	Timeout       int                    `json:"timeout"`
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

func NewSandbox() SandboxHandler {
	sOnce.Do(func() {
		sh = &sandbox{
			log: commonLog.NewLogger(),
		}
	})
	return sh
}

func (s *sandbox) Execute(ctx context.Context, req *SandboxExecuteRequest) (*SandboxExecuteResult, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	log := traceLog.WithContext(newCtx)
	globalConfig := common.NewConfig().Sandbox

	sessionConfig := convertToSessionConfig(req.SandboxConfig)
	mergedConfig := sandboxutil.MergeSessionConfig(sessionConfig, globalConfig)
	if mergedConfig.TemplateID == "" {
		return nil, fmt.Errorf("sandbox template_id is required")
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = mergedConfig.Timeout
	}

	sessionID, err := sandboxutil.GetOrCreateSession(newCtx, mergedConfig)
	if err != nil {
		log.Warnf("[Sandbox.Execute] getOrCreateSession err: %s", err.Error())
		return nil, fmt.Errorf("get or create session failed: %w", err)
	}

	sandboxClient := drivenadapters.NewSandbox()
	execReq := &drivenadapters.ExecuteRequest{
		Code:     req.Code,
		Language: req.Language,
		Event:    req.Event,
		Timeout:  timeout,
	}

	execution, err := sandboxClient.ExecuteCode(newCtx, sessionID, execReq)
	if err != nil {
		log.Warnf("[Sandbox.Execute] ExecuteCode err: %s", err.Error())
		return nil, fmt.Errorf("execute code failed: %w", err)
	}

	log.Infof("[Sandbox.Execute] Execution started, execution_id: %s, session_id: %s", execution.ExecutionID, sessionID)

	interval := time.Duration(sandboxutil.DefaultIntervalSec) * time.Second
	timeoutDuration := time.Duration(timeout) * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) >= timeoutDuration {
			return nil, fmt.Errorf("execution timeout after %d seconds", timeout)
		}

		execStatus, err := sandboxClient.GetExecutionStatus(newCtx, execution.ExecutionID)
		if err != nil {
			log.Warnf("[Sandbox.Execute] GetExecutionStatus err: %s", err.Error())
		} else {
			switch execStatus.Status {
			case "completed":
				result, serr := sandboxClient.GetExecutionResult(newCtx, execution.ExecutionID)
				if serr != nil {
					log.Warnf("[Sandbox.Execute] GetExecutionResult err: %s", serr.Error())
					return nil, ierrors.NewIError(ierrors.ErrorDepencyService, "", serr)
				}
				return &SandboxExecuteResult{
					ID:            result.ID,
					SessionID:     result.SessionID,
					Status:        result.Status,
					Code:          result.Code,
					Language:      result.Language,
					Timeout:       result.Timeout,
					ExitCode:      result.ExitCode,
					ErrorMessage:  result.ErrorMessage,
					ExecutionTime: result.ExecutionTime,
					Stdout:        result.Stdout,
					Stderr:        result.Stderr,
					Artifacts:     result.Artifacts,
					RetryCount:    result.RetryCount,
					CreatedAt:     result.CreatedAt,
					StartedAt:     result.StartedAt,
					CompletedAt:   result.CompletedAt,
					ReturnValue:   result.ReturnValue,
					Metrics:       result.Metrics,
				}, nil
			case "failed", "timeout":
				return nil, ierrors.NewIError(ierrors.ErrorDepencyService, "", execStatus)
			default:
			}
		}

		select {
		case <-newCtx.Done():
			log.Warnf("[Sandbox.Execute] Context cancelled or timeout")
			return nil, newCtx.Err()
		case <-time.After(interval):
		}
	}
}

func convertToSessionConfig(config *drivenadapters.SandboxSessionConfig) *drivenadapters.SandboxSessionConfig {
	return config
}
