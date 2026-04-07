package actions

import (
	"encoding/json"
	"fmt"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/vm"
)

type Define struct {
	Type  string `json:"type"` // string number object array
	Value any    `json:"value"`
}

func (a *Define) Name() string {
	return common.InternalDefineOpt
}

func (a *Define) ParameterNew() interface{} {
	return &Define{}
}

func convertValue(typ string, value any) (any, error) {
	switch typ {
	case "string":
		if s, ok := value.(string); ok {
			return s, nil
		}
		bs, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal value to string: %w", err)
		}
		return string(bs), nil
	case "number":
		switch v := value.(type) {
		case float64, float32, int, int64, int32, uint, uint32, uint64:
			return v, nil
		case string:
			var f float64
			_, err := fmt.Sscanf(v, "%f", &f)
			if err == nil {
				return f, nil
			}
			if n, err2 := json.Number(v).Float64(); err2 == nil {
				return n, nil
			}
			return nil, fmt.Errorf("failed to convert string to number: %v", v)
		default:
			return nil, fmt.Errorf("unsupported value type for number: %T", value)
		}
	case "array":
		switch v := value.(type) {
		case []any, []string:
			return v, nil
		case string:
			var arr []any
			if err := json.Unmarshal([]byte(v), &arr); err == nil {
				return arr, nil
			}
			return []any{v}, nil
		default:
			return []any{v}, nil
		}
	case "object":
		switch v := value.(type) {
		case map[string]any:
			return v, nil
		case string:
			var m map[string]any
			if err := json.Unmarshal([]byte(v), &m); err == nil {
				return m, nil
			}
			return nil, fmt.Errorf("string json parse to object failed: %v", v)
		default:
			return nil, fmt.Errorf("unsupported value type for object: %T", value)
		}
	default:
		return value, nil
	}
}

func (a *Define) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*Define)

	converted, err := convertValue(input.Type, input.Value)
	if err != nil {
		return nil, err
	}

	result := map[string]any{
		"value": converted,
	}
	ctx.ShareData().Set(ctx.GetTaskID(), result)
	return result, nil
}

type Assign struct {
	Target string
	Value  any `json:"value"`
}

func (a *Assign) Name() string {
	return common.InternalAssignOpt
}

func (a *Assign) ParameterNew() interface{} {
	return &Assign{}
}

func (a *Assign) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*Assign)

	tokens := vm.Parse(input.Target)

	if len(tokens) != 1 || tokens[0].Type != vm.TokenVariable {
		return nil, fmt.Errorf("target must be variable")
	}

	target := tokens[0].Value.(string)

	if len(tokens[0].AccessList) == 0 {
		ctx.ShareData().Set(target, input.Value)
		return nil, nil
	}

	value, ok := ctx.ShareData().Get(target)

	if !ok {
		return nil, fmt.Errorf("variable %s not exists", target)
	}

	valueBytes, _ := json.Marshal(value)
	obj := new(interface{})
	_ = json.Unmarshal(valueBytes, obj)

	var path []any

	for _, token := range tokens[0].AccessList {
		path = append(path, token.Value)
	}

	newObj := setJson(obj, path, input.Value)

	ctx.ShareData().Set(target, newObj)

	return nil, nil
}

var _ entity.Action = (*Assign)(nil)
var _ entity.Action = (*Define)(nil)
