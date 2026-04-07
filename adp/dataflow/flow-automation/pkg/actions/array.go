package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

type Filter string

const (
	FilterStringEq          Filter = "string-eq"
	FilterStringNeq         Filter = "string-neq"
	FilterStringContains    Filter = "string-contains"
	FilterStringNotContains Filter = "string-not-contains"
	FilterStringStartWith   Filter = "string-start-with"
	FilterStringEndWith     Filter = "string-end-with"
	FilterNumberEq          Filter = "number-eq"
	FilterNumberNeq         Filter = "number-neq"
	FilterNumberGte         Filter = "number-gte"
	FilterNumberGt          Filter = "number-gt"
	FilterNumberLte         Filter = "number-lte"
	FilterNumberLt          Filter = "number-lt"
	FilterDateEq            Filter = "date-eq"
	FilterDateNeq           Filter = "date-neq"
	FilterDateEarlierThan   Filter = "date-earlier-than"
	FilterDateLaterThan     Filter = "date-later-than"
	FilterIsNull            Filter = "is-null"
	FilterNotNull           Filter = "not-null"
)

type ArrayFilterCondition struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Filter Filter `json:"filter"`
	Value  any    `json:"value"`
}

type ArrayFilter struct {
	Data       any `json:"data"`
	Conditions [][]*ArrayFilterCondition
}

func (a *ArrayFilter) Name() string {
	return common.OpArrayFilter
}

func (a *ArrayFilter) ParameterNew() interface{} {
	return &ArrayFilter{}
}

func (a *ArrayFilter) compare(filter Filter, lhs any, rhs any) bool {

	switch filter {
	case FilterIsNull:
		return lhs == nil
	case FilterNotNull:
		return lhs != nil
	case FilterStringEq:
		vls, vrs, ok := a.ParseString(lhs, rhs)
		if !ok {
			return false
		}
		return vrs == vls
	case FilterStringNeq:
		vls, vrs, ok := a.ParseString(lhs, rhs)
		if !ok {
			return false
		}
		return vrs != vls
	case FilterStringContains:
		vls, vrs, ok := a.ParseString(lhs, rhs)
		if !ok {
			return false
		}
		return strings.Contains(vls, vrs)
	case FilterStringNotContains:
		vls, vrs, ok := a.ParseString(lhs, rhs)
		if !ok {
			return false
		}
		return !strings.Contains(vls, vrs)
	case FilterStringStartWith:
		vls, vrs, ok := a.ParseString(lhs, rhs)
		if !ok {
			return false
		}
		return strings.HasPrefix(vls, vrs)
	case FilterStringEndWith:
		vls, vrs, ok := a.ParseString(lhs, rhs)
		if !ok {
			return false
		}
		return strings.HasSuffix(vls, vrs)
	case FilterNumberEq:
		vls, vrs, ok := a.ParseNumber(lhs, rhs)
		if !ok {
			return false
		}
		return vrs == vls
	case FilterNumberNeq:
		vls, vrs, ok := a.ParseNumber(lhs, rhs)
		if !ok {
			return false
		}
		return vrs != vls
	case FilterNumberGte:
		vls, vrs, ok := a.ParseNumber(lhs, rhs)
		if !ok {
			return false
		}
		return vls >= vrs
	case FilterNumberGt:
		vls, vrs, ok := a.ParseNumber(lhs, rhs)
		if !ok {
			return false
		}
		return vls > vrs
	case FilterNumberLte:
		vls, vrs, ok := a.ParseNumber(lhs, rhs)
		if !ok {
			return false
		}
		return vls <= vrs
	case FilterNumberLt:
		vls, vrs, ok := a.ParseNumber(lhs, rhs)
		if !ok {
			return false
		}
		return vls < vrs
	case FilterDateEq:
		vls, vrs, ok := a.ParseDate(lhs, rhs)
		if !ok {
			return false
		}
		return vrs == vls
	case FilterDateNeq:
		vls, vrs, ok := a.ParseDate(lhs, rhs)
		if !ok {
			return false
		}
		return vrs != vls
	case FilterDateEarlierThan:
		vls, vrs, ok := a.ParseDate(lhs, rhs)
		if !ok {
			return false
		}
		return vls < vrs
	case FilterDateLaterThan:
		vls, vrs, ok := a.ParseDate(lhs, rhs)
		if !ok {
			return false
		}
		return vls > vrs
	}

	return false
}

// ParseString 变量转换为string类型
func (a *ArrayFilter) ParseString(vl, vr interface{}) (vls, vrs string, ok bool) {
	vls, ok = vl.(string)
	if !ok {
		return
	}
	vrs, ok = vr.(string)
	if !ok {
		return
	}

	return
}

// ParseNumber 转换数字
func (a *ArrayFilter) ParseNumber(vl, vr interface{}) (vls, vrs float64, ok bool) {
	switch tl := vl.(type) {
	case string:
		floatvar, err := strconv.ParseFloat(tl, 64)
		if err != nil {
			vls = 0
			break
		}
		vls = floatvar
	case float64:
		vls = tl
	case int64:
		vls = float64(tl)
	}

	switch tr := vr.(type) {
	case string:
		floatvar, err := strconv.ParseFloat(tr, 64)
		if err != nil {
			vls = 0
			break
		}
		vrs = floatvar
	case float64:
		vrs = tr
	case int64:
		vrs = float64(tr)
	}
	return vls, vrs, true
}

// ParseDate 转换日期
func (a *ArrayFilter) ParseDate(vl, vr interface{}) (vls, vrs int64, ok bool) {
	switch tl := vl.(type) {
	case string:
		v, err := time.Parse(time.RFC3339, tl)
		if err != nil {
			var newNum float64
			_, err := fmt.Sscanf(tl, "%e", &newNum)
			if err != nil {
				vls = 0
				break
			}
			vls = int64(newNum)
			break
		}
		vls = v.UnixNano() / 1e9 // 秒级
	case int64:
		vls = tl
	}

	switch tr := vr.(type) {
	case string:
		v, err := time.Parse(time.RFC3339, tr)
		if err != nil {
			var newNum float64
			_, err := fmt.Sscanf(tr, "%e", &newNum)
			if err != nil {
				vrs = 0
				break
			}
			vrs = int64(newNum)
			break
		}
		vrs = v.UnixNano() / 1e9 // 秒级
	case int64:
		vrs = tr
	}
	return vls, vrs, true
}

func (a *ArrayFilter) match(data any, conditions [][]*ArrayFilterCondition) bool {

	for _, group := range conditions {
		andResult := true

		for _, cond := range group {
			path := keyToPath(cond.Key)
			value := lookupJson(data, path, nil)
			andResult = andResult && a.compare(cond.Filter, value, cond.Value)
			if !andResult {
				break
			}
		}

		if andResult {
			return true
		}
	}

	return false
}

func (a *ArrayFilter) Run(ctx entity.ExecuteContext, params interface{}, token *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)
	input := params.(*ArrayFilter)
	var inputData, outputData []any

	switch value := input.Data.(type) {
	case string:
		var data any
		jsonerr := json.Unmarshal([]byte(value), &data)
		if jsonerr != nil {
			inputData = []any{value}
		} else if dataArr, ok := data.([]any); ok {
			inputData = dataArr
		} else {
			inputData = []any{data}
		}
	case []any:
		inputData = value
	default:
		inputData = []any{value}
	}

	for _, value := range inputData {
		if a.match(value, input.Conditions) {
			outputData = append(outputData, value)
		}
	}

	result := map[string]any{
		"data": outputData,
	}

	return result, nil
}

var _ entity.Action = (*ArrayFilter)(nil)
