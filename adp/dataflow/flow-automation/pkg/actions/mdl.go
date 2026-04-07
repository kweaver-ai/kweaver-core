package actions

import (
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/rds"
)

type MDLDataViewTrigger struct{}

// Name implements entity.Action.
func (m *MDLDataViewTrigger) Name() string {
	return common.MDLDataViewTrigger
}

func (m *MDLDataViewTrigger) ParameterNew() interface{} {
	return &MDLDataViewTrigger{}
}

// Run implements entity.Action.
func (m *MDLDataViewTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	id := ctx.GetTaskID()
	taskIns := ctx.GetTaskInstance()
	dagIns := taskIns.RelatedDagInstance

	var originalResult any
	switch dagIns.EventPersistence {
	case entity.DagInstanceEventPersistenceOss:
		return nil, fmt.Errorf("DagInstance is already archived")
	case entity.DagInstanceEventPersistenceSql:
		events, err := dagIns.ListEvents(ctx.Context(), &rds.DagInstanceEventListOptions{
			Types: []rds.DagInstanceEventType{rds.DagInstanceEventTypeVariable},
			Names: []string{fmt.Sprintf("__%s", id)},
		})

		if err != nil {
			return nil, err
		}

		if len(events) > 0 {
			originalResult = events[len(events)-1].Data
		}
	default:
		if data, ok := ctx.ShareData().Get("__" + id); ok {
			originalResult = data
		}
	}

	result := make(map[string]any)
	result["_type"] = "dataview"

	resultMap, ok := originalResult.(map[string]any)
	if !ok {
		ctx.Trace(ctx.Context(), "run end")
		return result, nil
	}

	if status, ok := resultMap["status"]; ok && status == string(entity.TaskInstanceStatusFailed) {
		return nil, fmt.Errorf("trigger error")
	}

	data, ok := resultMap["data"]
	if !ok {
		ctx.Trace(ctx.Context(), "run end")
		return result, nil
	}

	dataArr, ok := data.([]any)
	if !ok {
		ctx.Trace(ctx.Context(), "run end")
		return result, nil
	}

	if len(dataArr) > 10 {
		result["summary"] = fmt.Sprintf("共 %d 条数据，仅显示前 10 条", len(dataArr))
		result["data"] = dataArr[0:10]
	} else {
		result["data"] = dataArr
	}

	ctx.Trace(ctx.Context(), "run end")
	return result, nil
}

var _ entity.Action = (*MDLDataViewTrigger)(nil)
