package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	sandboxutil "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/utils/sandbox"
)

const (
	sandboxExecutionPrefix       = "automation:sandbox_execution"
	defaultSandboxCacheExpireSec = int64(24 * 60 * 60)
)

type SandboxExecute struct {
	Code          string                               `json:"code"`
	Language      string                               `json:"language"`
	Event         map[string]interface{}               `json:"event,omitempty"`
	Timeout       int                                  `json:"timeout,omitempty"`
	CacheResult   *bool                                `json:"cache_result,omitempty"`
	IntervalSec   int                                  `json:"interval_sec,omitempty"`
	SandboxConfig *drivenadapters.SandboxSessionConfig `json:"sandbox_config,omitempty"`
}

func (s *SandboxExecute) Name() string {
	return common.OpSandboxExecute
}

func (s *SandboxExecute) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)
	globalConfig := common.NewConfig().Sandbox

	input := params.(*SandboxExecute)
	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return nil, fmt.Errorf("get taskinstance failed")
	}

	log := traceLog.WithContext(newCtx)

	cacheResult := false
	if input.CacheResult != nil {
		cacheResult = *input.CacheResult
	}

	intervalSec := input.IntervalSec
	if intervalSec <= 0 {
		intervalSec = sandboxutil.DefaultIntervalSec
	}

	sessionConfig := s.mergeConfig(input.SandboxConfig, globalConfig)
	if sessionConfig.TemplateID == "" {
		return nil, fmt.Errorf("sandbox template_id is required")
	}

	executor := &sandboxExecutor{
		action:        s,
		input:         input,
		sessionConfig: sessionConfig,
		intervalSec:   intervalSec,
		timeout:       sessionConfig.Timeout,
	}

	if taskIns.Settings != nil && taskIns.Settings.TimeOut != nil && taskIns.Settings.TimeOut.Delay > 0 {
		executor.timeout = taskIns.Settings.TimeOut.Delay
	}

	if cacheResult {
		manager := NewAsyncTaskManager(ctx.NewExecuteMethods()).
			WithLockPrefix(sandboxExecutionPrefix)
		return manager.Run(ctx, executor)
	}

	result, err := executor.Execute(newCtx)

	if err != nil {
		ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusFailed)
		log.Warnf("[SandboxExecute.Run] Execute err: %s, taskInsID: %s", err.Error(), taskIns.ID)
		return nil, err
	}

	ctx.ShareData().Set("__status_"+taskIns.ID, entity.TaskInstanceStatusSuccess)
	ctx.ShareData().Set(ctx.GetTaskID(), result)
	return result, nil
}

func (s *SandboxExecute) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	manager := NewAsyncTaskManager(ctx.NewExecuteMethods())
	return manager.RunAfter(ctx)
}

func (s *SandboxExecute) ParameterNew() interface{} {
	return &SandboxExecute{}
}

func (s *SandboxExecute) mergeConfig(nodeConfig *drivenadapters.SandboxSessionConfig, globalConfig common.Sandbox) *drivenadapters.SandboxSessionConfig {
	return sandboxutil.MergeSessionConfig(nodeConfig, globalConfig)
}

type sandboxExecutor struct {
	action        *SandboxExecute
	input         *SandboxExecute
	sessionConfig *drivenadapters.SandboxSessionConfig
	intervalSec   int
	timeout       int
}

func (e *sandboxExecutor) GetTaskType() string {
	return e.action.Name()
}

func (e *sandboxExecutor) GetHashContent() string {
	payload := struct {
		Action   string                 `json:"action"`
		Code     string                 `json:"code"`
		Language string                 `json:"language"`
		Event    map[string]interface{} `json:"event,omitempty"`
	}{
		Action:   e.action.Name(),
		Code:     e.input.Code,
		Language: e.input.Language,
		Event:    e.input.Event,
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Sprintf("%s:%s:%s:%v", e.action.Name(), e.input.Code, e.input.Language, e.input.Event)
	}
	return string(data)
}

func (e *sandboxExecutor) GetExpireSeconds() int64 {
	return defaultSandboxCacheExpireSec
}

func (e *sandboxExecutor) GetResultFileExt() string {
	return ".json"
}

func (e *sandboxExecutor) Execute(ctx context.Context) (map[string]any, error) {
	log := traceLog.WithContext(ctx)

	sessionID, err := sandboxutil.GetOrCreateSession(ctx, e.sessionConfig)
	if err != nil {
		return nil, fmt.Errorf("get or create session failed: %w", err)
	}

	sandboxClient := drivenadapters.NewSandbox()
	execReq := &drivenadapters.ExecuteRequest{
		Code:     e.input.Code,
		Language: e.input.Language,
		Event:    e.input.Event,
		Timeout:  e.timeout,
	}

	execution, err := sandboxClient.ExecuteCode(ctx, sessionID, execReq)
	if err != nil {
		return nil, fmt.Errorf("execute code failed: %w", err)
	}

	log.Infof("[SandboxExecute.Execute] Execution started, execution_id: %s, session_id: %s", execution.ExecutionID, sessionID)

	interval := time.Duration(e.intervalSec) * time.Second
	if interval <= 0 {
		interval = time.Duration(sandboxutil.DefaultIntervalSec) * time.Second
	}

	timeoutDuration := time.Duration(e.timeout) * time.Second
	startTime := time.Now()

	for {
		if time.Since(startTime) >= timeoutDuration {
			return nil, fmt.Errorf("execution timeout after %d seconds", e.timeout)
		}

		execStatus, err := sandboxClient.GetExecutionStatus(ctx, execution.ExecutionID)
		if err != nil {
			log.Warnf("[SandboxExecute.Execute] GetExecutionStatus err: %s", err.Error())
		} else {
			switch execStatus.Status {
			case "completed":
				result, err := sandboxClient.GetExecutionResult(ctx, execution.ExecutionID)
				if err != nil {
					return nil, fmt.Errorf("get execution result failed: %w", err)
				}
				return map[string]any{"return_value": result.ReturnValue}, nil
			case "failed", "timeout":
				errMsg := "execution failed"
				if execStatus.ErrorMessage != nil {
					errMsg = *execStatus.ErrorMessage
				}
				return nil, fmt.Errorf("%v", errMsg)
			default:
			}
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
	}
}
