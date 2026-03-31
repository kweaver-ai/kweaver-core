package actions

import (
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/entity"
)

const (
	daysPerYear           = 365
	daysPerMonth          = 30
	hoursPerDay           = 24
	minutesPerHour        = 60
	secondsPerMinute      = 60
	microsecondsPerSecond = 1000000
)

// TimeNow 获取当前时间
type TimeNow struct {
}

// TimeNowParam 获取当前时间输入参数
type TimeNowParam struct {
}

// Name 操作名称
func (a *TimeNow) Name() string {
	return common.InternalTimeNow
}

// Run 操作方法
func (a *TimeNow) Run(ctx entity.ExecuteContext, _ interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)

	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	data := map[string]int64{
		"curtime": time.Now().UnixNano() / 1e3,
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *TimeNow) ParameterNew() interface{} {
	return &TimeNowParam{}
}

// TimeRelative 获取相对时间
type TimeRelative struct {
}

// TimeRelativeParam 获取相对时间输入参数
type TimeRelativeParam struct {
	OldTime       int64  `json:"old_time"`
	RelativeType  string `json:"relative_type"`
	RelativeValue int64  `json:"relative_value"`
	RelativeUnit  string `json:"relative_unit"`
}

// Name 操作名称
func (a *TimeRelative) Name() string {
	return common.InternalTimeRelative
}

// Run 操作方法
func (a *TimeRelative) Run(ctx entity.ExecuteContext, params interface{}, _ *entity.Token) (interface{}, error) {
	var err error
	newCtx, span := trace.StartInternalSpan(ctx.Context())
	defer func() { trace.TelemetrySpanEnd(span, err) }()
	ctx.SetContext(newCtx)
	ctx.Trace(ctx.Context(), "run start", entity.TraceOpPersistAfterAction)

	input := params.(*TimeRelativeParam)
	oldTime := input.OldTime
	newTime := input.OldTime
	relativeType := input.RelativeType
	relativeValue := input.RelativeValue
	relativeUnit := input.RelativeUnit
	var sep int64

	switch relativeUnit {
	case "year":
		sep = relativeValue * daysPerYear * hoursPerDay * minutesPerHour * secondsPerMinute * microsecondsPerSecond
	case "month":
		sep = relativeValue * daysPerMonth * hoursPerDay * minutesPerHour * secondsPerMinute * microsecondsPerSecond
	case "day":
		sep = relativeValue * hoursPerDay * minutesPerHour * secondsPerMinute * microsecondsPerSecond
	case "hour":
		sep = relativeValue * minutesPerHour * secondsPerMinute * microsecondsPerSecond
	case "minute":
		sep = relativeValue * secondsPerMinute * microsecondsPerSecond
	default:
		sep = relativeValue * microsecondsPerSecond
	}

	if relativeType == "add" {
		newTime = oldTime + sep
	} else if relativeType == "sub" {
		newTime = oldTime - sep
	}

	data := map[string]int64{
		"new_time": newTime,
	}
	id := ctx.GetTaskID()
	ctx.ShareData().Set(id, data)
	ctx.Trace(ctx.Context(), "run end")
	return data, nil
}

// ParameterNew 初始化参数
func (a *TimeRelative) ParameterNew() interface{} {
	return &TimeRelativeParam{}
}
