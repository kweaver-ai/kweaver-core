package policy

import (
	"context"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/trace"
)

const TimeOutPolicyName = "TimeOutPolicy"

type TimeoutPolicy struct {
	Delay int
}

func NewTimeOutPolicy(duration int) *TimeoutPolicy {
	return &TimeoutPolicy{Delay: duration}
}

func (t *TimeoutPolicy) Name() string { return TimeOutPolicyName }

func (t *TimeoutPolicy) Do(ctx context.Context, fn func(context.Context) error) error {
	var err error

	ctx, span := trace.StartInternalSpan(ctx)
	defer func() { trace.TelemetrySpanEnd(span, err) }()

	cctx, cancle := context.WithTimeout(ctx, time.Duration(t.Delay)*time.Second)
	defer cancle()

	done := make(chan error, 1)

	// 用于确保 channel 只被关闭一次
	var closeOnce sync.Once

	// 安全关闭 channel
	safeClose := func() {
		closeOnce.Do(func() {
			close(done)
		})
	}

	go func() {
		defer safeClose()
		done <- fn(cctx)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-cctx.Done():
		return cctx.Err()
	case err := <-done:
		return err
	}
}
