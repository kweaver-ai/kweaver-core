package actions

import (
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/drivenadapters"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

type CallAgent struct {
	AgentID string                 `json:"agent_id"`
	Inputs  map[string]interface{} `json:"inputs"`
}

// Name implements entity.Action.
func (c *CallAgent) Name() string {
	return common.OpAnyDataCallAgent
}

func (c *CallAgent) ParameterNew() interface{} {
	return &CallAgent{}
}

// Run implements entity.Action.
func (c *CallAgent) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {

	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(newCtx, "run start", entity.TraceOpPersistAfterAction)

	input := params.(*CallAgent)
	agentApp := drivenadapters.NewAgentApp()

	req := &drivenadapters.ChatCompletionReq{
		AgentID:      input.AgentID,
		Stream:       false,
		Query:        "",
		CustomQuerys: map[string]any{},
		BizDomainID:  ctx.GetTaskInstance().RelatedDagInstance.BizDomainID,
	}

	for k, v := range input.Inputs {
		switch k {
		case "history", "tool", "header", "self_config":
			continue
		case "query":
			req.Query, _ = v.(string)
		default:
			req.CustomQuerys[k] = v
		}
	}

	answer, thinking, err := agentApp.ChatCompletion(ctx.Context(), input.AgentID, req, token.Token)

	if err != nil {
		ctx.Trace(newCtx, fmt.Sprintf("run error: %v", err))
		return nil, err
	}

	taskID := ctx.GetTaskID()

	result := map[string]interface{}{
		"answer": answer,
		"think":  thinking,
		"json":   ParseJsonValue(answer),
	}

	ctx.ShareData().Set(taskID, result)

	ctx.Trace(newCtx, "run end")
	return result, nil
}

var _ entity.Action = (*CallAgent)(nil)
