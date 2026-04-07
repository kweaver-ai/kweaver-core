package actions

import (
	"context"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/dependency"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

type CustomAction struct {
	ExecutorID uint64                 `json:"executor_id"`
	ActionID   uint64                 `json:"action_id"`
	Parameters map[string]interface{} `json:"parameters"`
}

type ActionInput struct {
	Key      *string `json:"key"`
	Name     *string `json:"name"`
	Type     *string `json:"type"`
	Required *bool   `json:"required"`
}

type ActionOutput struct {
	Key  *string `json:"key"`
	Name *string `json:"name"`
	Type *string `json:"type"`
}

type ActionConfig struct {
	Code *string `json:"code"`
}

func (*CustomAction) Name() string {
	return "@custom"
}

// RunBefore 操作方法
func (a *CustomAction) RunBefore(_ entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	return entity.TaskInstanceStatusBlocked, nil
}

func (*CustomAction) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	newCtx = context.WithValue(newCtx, common.Authorization, token.Token)
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*CustomAction)

	taskIns := ctx.GetTaskInstance()
	if taskIns == nil {
		return nil, fmt.Errorf("get taskinstance failed")
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

	customExecutor := dependency.NewCustomExecutor()
	action, err := customExecutor.GetAccessableAction(ctx.Context(), input.ActionID, input.ExecutorID, token.UserID)

	if err != nil {
		ctx.Trace(ctx.Context(), err.Error())
		return nil, err
	}

	if action.Config == nil {
		return nil, fmt.Errorf("invalid action config")
	}

	var actionConfig ActionConfig

	err = jsoniter.UnmarshalFromString(*action.Config, &actionConfig)

	if err != nil || actionConfig.Code == nil || *actionConfig.Code == "" {
		return nil, fmt.Errorf("invalid action config")
	}

	actionInputs := make([]ActionInput, 0)
	actionOutputs := make([]ActionOutput, 0)

	if action.Inputs != nil && len(*action.Inputs) > 0 {
		err := jsoniter.UnmarshalFromString(*action.Inputs, &actionInputs)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	}

	if (action.Outputs != nil) && len(*action.Outputs) > 0 {
		err := jsoniter.UnmarshalFromString(*action.Outputs, &actionOutputs)

		if err != nil {
			ctx.Trace(ctx.Context(), err.Error())
			return nil, err
		}
	}

	inputParams := make([]map[string]any, 0)
	outputParams := make([]map[string]any, 0)

	if len(actionInputs) > 0 {
		for _, actionInput := range actionInputs {

			var inputType string
			switch *actionInput.Type {
			case "number":
				inputType = "int"
			case "multipleFiles":
				inputType = "array"
			default:
				inputType = "string"
			}

			if value, ok := input.Parameters[*actionInput.Key]; ok {
				inputParams = append(inputParams, map[string]any{
					"type":  inputType,
					"value": value,
					"id":    *actionInput.Name,
					"key":   *actionInput.Key,
				})
			} else if *actionInput.Required {
				ctx.Trace(ctx.Context(), fmt.Sprintf("missing required action input %s", *actionInput.Name))
				return nil, fmt.Errorf("missing required action input %s", *actionInput.Name)
			} else {
				inputValue := ""
				switch inputType {
				case "int":
					inputValue = "0"
				case "array":
					inputValue = "[]"
				case "object":
					inputValue = "{}"
				default:
					inputValue = ""
				}

				inputParams = append(inputParams, map[string]any{
					"type":  inputType,
					"value": inputValue,
					"id":    *actionInput.Name,
					"key":   *actionInput.Key,
				})
			}
		}
	}

	if len(actionOutputs) > 0 {
		for _, actionOutput := range actionOutputs {
			var outputType string
			switch *actionOutput.Type {
			case "number":
				outputType = "int"
			case "multipleFiles":
				outputType = "array"
			default:
				outputType = "string"
			}

			outputParams = append(outputParams, map[string]any{
				"id":   *actionOutput.Name,
				"key":  *actionOutput.Key,
				"type": outputType,
			})
		}
	}

	applyID := taskIns.ID
	config := common.NewConfig()

	if ctx.IsDebug() {
		applyID = fmt.Sprintf("DEBUG:%v", applyID)
	}
	callback := fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/continue/%s", config.ContentAutomation.PrivateHost, config.ContentAutomation.PrivatePort, applyID)
	codeRunnerAdapter := drivenadapters.NewCodeRunner()
	_, err = codeRunnerAdapter.AsyncRunPyCode(ctx.Context(), *actionConfig.Code, inputParams, outputParams, callback)
	if err != nil {
		return nil, err
	}

	ctx.Trace(ctx.Context(), "run end")

	return map[string]interface{}{}, nil
}

func (*CustomAction) ParameterNew() interface{} {
	return &CustomAction{}
}

var (
	_ entity.Action = (*CustomAction)(nil)
)
