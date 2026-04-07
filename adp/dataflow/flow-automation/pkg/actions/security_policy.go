package actions

import (
	"encoding/json"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

type SecurityPolicyTrigger struct {
}

type SecurityPolicyTriggerParam struct {
}

func (a *SecurityPolicyTrigger) Name() string {
	return common.SecurityPolicyTrigger
}

// Run 操作方法 表单触发
func (a *SecurityPolicyTrigger) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	attr := make(map[string]interface{})
	fields := make(map[string]interface{})
	accessor := Accessor{}
	var source map[string]interface{}
	ctx.IterateVars(func(key, val string) (stop bool) {
		if key == "operator_id" {
			accessor.ID = val
		} else if key == "operator_name" {
			accessor.Name = val
		} else if key == "operator_type" {
			accessor.Type = val
		} else if key == "source" {
			json.Unmarshal([]byte(val), &source)
		} else {
			fields[key] = val
		}
		return false
	})

	accessorBytes, _ := json.Marshal(accessor)

	attr["fields"] = fields
	attr["accessor"] = string(accessorBytes)

	if objectID, ok := source["object_id"].(string); ok {
		metadata, err := ctx.NewASDoc().GetObjectMsg(ctx.Context(), objectID)
		if err != nil {
			return nil, err
		}

		var sourceType string

		if metadata.Size == -1 {
			sourceType = "folder"
		} else {
			sourceType = "file"
		}

		source = map[string]interface{}{
			"type": sourceType,
			"id":   metadata.DocID,
			"name": metadata.Name,
			"rev":  metadata.Rev,
			"size": metadata.Size,
			"path": metadata.Path,
		}
	}

	attr["source"] = source
	id := ctx.GetTaskID()
	ctx.ShareData().Set("source", source)
	ctx.ShareData().Set(id, attr)
	ctx.Trace(ctx.Context(), "run end")
	return attr, nil
}

// ParameterNew new parameter
func (a *SecurityPolicyTrigger) ParameterNew() interface{} {
	return &SecurityPolicyTriggerParam{}
}
