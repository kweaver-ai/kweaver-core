package entity

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoopDo(t *testing.T) {
	cnt := 0
	tests := []struct {
		giveDo       func() error
		giveInterval time.Duration
		timeout      time.Duration
		cancelAfter  time.Duration
		wantErr      error
		wantTraces   []string
	}{
		{
			giveDo: func() error {
				cnt++
				if cnt == 3 {
					return EndLoop
				}
				return nil
			},
			wantErr: nil,
		},
		{
			giveInterval: time.Millisecond,
			giveDo: func() error {
				cnt++
				if cnt == 3 {
					return fmt.Errorf("three failed")
				}
				return nil
			},
			wantErr: fmt.Errorf("three failed"),
		},
		{
			giveDo: func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			wantErr:      context.DeadlineExceeded,
			giveInterval: time.Millisecond,
			timeout:      time.Millisecond,
		},
		{
			giveDo: func() error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			wantErr: context.Canceled,
			wantTraces: []string{
				"task is canceled",
			},
			giveInterval: time.Millisecond,
			cancelAfter:  time.Millisecond,
		},
	}

	for _, tc := range tests {
		cnt = 0
		pc, cancel := context.WithCancel(context.TODO())
		if tc.timeout > 0 {
			pc, cancel = context.WithTimeout(pc, tc.timeout)
		}
		ctx := &DefExecuteContext{
			ctx: pc,
		}
		var traces []string
		ctx.trace = func(ctx context.Context, msg string, opt ...TraceOp) {
			traces = append(traces, msg)
		}
		if tc.cancelAfter > 0 {
			go func() {
				time.Sleep(tc.cancelAfter)
				cancel()
			}()
		}
		err := LoopDo(ctx, tc.giveDo, LoopInterval(tc.giveInterval))
		assert.Equal(t, tc.wantErr, err)
		assert.Equal(t, tc.wantTraces, traces)
	}
}
