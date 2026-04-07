package policy

import (
	"context"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
)

const RetryPolicyName = "RetryPolicy"

type RetryPolicy struct {
	Max       int
	Delay     int
	RetryIf   func(err error) bool
	collector *ResultCollector
}

func NewRetryPolicy(max, delay int, fn func(err error) bool) *RetryPolicy {
	return &RetryPolicy{Max: max, Delay: delay, RetryIf: fn}
}

func (r *RetryPolicy) Name() string                            { return RetryPolicyName }
func (r *RetryPolicy) SetCollector(collector *ResultCollector) { r.collector = collector }

func (r *RetryPolicy) Init() {
	result := NewResult(r.Name(), &RetryData{
		Attempts: 0,
		Max:      r.Max,
	})
	// 初始化防止空指针
	r.collector.Add(result)
}

func (r *RetryPolicy) Do(ctx context.Context, fn func(context.Context) error) error {
	var err error
	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	result, ok := GetAs[*RetryData](r.collector, r.Name())
	if !ok {
		r.Init()
	}

	defer func() {
		if r.collector != nil {
			r.collector.Add(result)
		}
	}()

	for i := 0; i <= r.Max; i++ {
		result.Data.Attempts = i
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err = fn(ctx)
		if err == nil {
			return nil
		}

		if r.RetryIf != nil && r.RetryIf(err) {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(r.Delay) * time.Second):
				continue
			}
		}
		return err
	}
	return err
}
