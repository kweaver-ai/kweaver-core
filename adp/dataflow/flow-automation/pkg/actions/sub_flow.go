package actions

import (
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

// SubFlowCallWorkflow 工作流子流程调用参数
type SubFlowCallWorkflow struct {
	Data      map[string]any `json:"data"`
	DagID     string         `json:"dag_id"`
	VersionID string         `json:"version_id"`
}

func (*SubFlowCallWorkflow) Name() string {
	return common.SubFlowCallWorkflowOpt
}

func (*SubFlowCallWorkflow) ParameterNew() interface{} {
	return &SubFlowCallWorkflow{}
}

func (c *SubFlowCallWorkflow) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	taskIns := ctx.GetTaskInstance()
	input := params.(*SubFlowCallWorkflow)

	err = subFlowCall(ctx, input.DagID, input.VersionID, input.Data, token, taskIns, false)
	if err != nil {
		log.Warnf("[run.SubFlowCallWorkflow] subFlowCall failed, detail: %s", err.Error())
		return nil, err
	}

	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusBlocked)

	return map[string]any{}, nil
}

func (c *SubFlowCallWorkflow) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	status, ok := ctx.ShareData().Get(statusKey)
	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

// SubFlowCallDataflow 数据流子流程调用参数
type SubFlowCallDataflow struct {
	DagID     string `json:"dag_id"`
	VersionID string `json:"version_id"`
}

func (*SubFlowCallDataflow) Name() string {
	return common.SubFlowCallDataflowOpt
}

func (*SubFlowCallDataflow) ParameterNew() interface{} {
	return &SubFlowCallDataflow{}
}

func (c *SubFlowCallDataflow) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	log := traceLog.WithContext(ctx.Context())

	taskIns := ctx.GetTaskInstance()
	input := params.(*SubFlowCallDataflow)

	err = subFlowCall(ctx, input.DagID, input.VersionID, nil, token, taskIns, true)
	if err != nil {
		log.Warnf("[run.SubFlowCallDataflow] subFlowCall failed, detail: %s", err.Error())
		return nil, err
	}

	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	ctx.ShareData().Set(statusKey, entity.TaskInstanceStatusBlocked)

	return map[string]any{}, nil
}

func (c *SubFlowCallDataflow) RunAfter(ctx entity.ExecuteContext, _ interface{}) (entity.TaskInstanceStatus, error) {
	taskIns := ctx.GetTaskInstance()
	statusKey := fmt.Sprintf("__status_%s", taskIns.ID)
	status, ok := ctx.ShareData().Get(statusKey)
	if ok && status == entity.TaskInstanceStatusBlocked {
		return entity.TaskInstanceStatusBlocked, nil
	}

	return entity.TaskInstanceStatusSuccess, nil
}

func subFlowCall(ctx entity.ExecuteContext, dagID, versionID string, data map[string]any, token *entity.Token, taskIns *entity.TaskInstance, isDataflow bool) error {
	dag, err := ctx.NewExecuteMethods().GetDag(ctx.Context(), dagID, versionID)
	if err != nil {
		return err
	}

	if !isDataflow && dag.Steps[0].DataSource != nil {
		return fmt.Errorf("subflow execution supports only single instance triggering")
	}

	var url string
	config := common.NewConfig()

	if dag.Trigger == entity.TriggerManually {
		url = fmt.Sprintf("http://%s:%s/api/automation/v1/run-instance/%s?version_id=%s", config.ContentAutomation.PublicHost, config.ContentAutomation.PublicPort, dagID, versionID)
	} else if dag.Trigger == entity.TriggerForm {
		url = fmt.Sprintf("http://%s:%s/api/automation/v1/run-instance-form/%s?version_id=%s", config.ContentAutomation.PublicHost, config.ContentAutomation.PublicPort, dagID, versionID)
	}

	headers := map[string]string{
		"Authorization":         token.Token,
		"X-Parent-Execution-ID": taskIns.RelatedDagInstance.ID,
		"X-Callback-URL":        fmt.Sprintf("http://%s:%s/api/automation/v1/trigger/continue/%s", config.ContentAutomation.PrivateHost, config.ContentAutomation.PrivatePort, taskIns.ID),
	}

	client := drivenadapters.NewOtelHTTPClient()
	_, _, err = client.Post(ctx.Context(), url, headers, map[string]any{"data": data})
	if err != nil {
		return err
	}

	return nil
}
