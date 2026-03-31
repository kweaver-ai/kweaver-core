package actions

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const pyExe = "@internal/tool/py3"

const (
	PyExeModeAsync = "async"
	PyExeModeSync  = "sync"
)

// PyExe python代码执行
type PyExe struct {
	Mode         string           `json:"mode"`
	Code         string           `json:"code"`
	InputParams  []map[string]any `json:"input_params"`
	OutputParams []map[string]any `json:"output_params"`
}

// Name 操作名称
func (a *PyExe) Name() string {
	return pyExe
}

// RunBefore 操作方法
func (a *PyExe) RunBefore(_ entity.ExecuteContext, params interface{}) (entity.TaskInstanceStatus, error) {

	input := params.(*PyExe)

	if input.Mode == PyExeModeSync {
		return entity.TaskInstanceStatusRunning, nil
	}

	return entity.TaskInstanceStatusBlocked, nil
}

// Run 操作方法
func (a *PyExe) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*PyExe)
	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return nil, fmt.Errorf("get taskinstance failed")
	}
	codeRunnerAdapter := drivenadapters.NewCodeRunner()

	if input.Mode == PyExeModeSync {
		id := ctx.GetTaskID()
		codeRunnerAdapter := drivenadapters.NewCodeRunner()
		res, err := codeRunnerAdapter.RunPyCode(ctx.Context(), input.Code, input.InputParams, input.OutputParams)
		if err != nil {
			return nil, err
		}

		ctx.ShareData().Set(id, res)
		ctx.Trace(ctx.Context(), "run end")
		return res, err
	}

	retryTime, ok := ctx.ShareData().Get("__retrytimes_" + taskIns.ID)
	if ok {
		retryTimesInt := ParseInt(retryTime)
		if retryTimesInt > 2 {
			if res, iok := ctx.ShareData().Get("__" + taskIns.TaskID); iok {
				return nil, fmt.Errorf("%v", res)
			}
			return nil, fmt.Errorf("retry too many times")
		}
		ctx.ShareData().Set("retrytimes_"+taskIns.ID, retryTimesInt+1)
	} else {
		ctx.ShareData().Set("retrytimes_"+taskIns.ID, 1)
	}

	applyID := taskIns.ID
	if ctx.IsDebug() {
		applyID = fmt.Sprintf("DEBUG:%v", applyID)
	}

	config := common.NewConfig()
	callback := fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/continue/%s", config.ContentAutomation.PrivateHost, config.ContentAutomation.PrivatePort, applyID)
	_, err = codeRunnerAdapter.AsyncRunPyCode(ctx.Context(), input.Code, input.InputParams, input.OutputParams, callback)
	if err != nil {
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")
	return map[string]interface{}{}, nil
}

// ParameterNew 初始化参数
func (a *PyExe) ParameterNew() interface{} {
	return &PyExe{}
}
